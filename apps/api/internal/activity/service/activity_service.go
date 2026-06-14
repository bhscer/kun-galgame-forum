package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"maps"
	"time"

	"kun-galgame-api/internal/activity/dto"
	"kun-galgame-api/internal/activity/repository"
	galgameClient "kun-galgame-api/internal/galgame/client"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/userclient"

	"github.com/redis/go-redis/v9"
)

// activityCacheTTL bounds how stale the "最新动态" feed can be. Short on purpose:
// it collapses the home page's per-render cost (the multi-source timeline query
// + wiki name-enrichment + OAuth identity round-trips) into one computation per
// window, while still surfacing new activity within seconds.
const activityCacheTTL = 30 * time.Second

const (
	// activityMaxRounds bounds serveKeyset's fetch-until-`limit`-survivors loop
	// so a pathological run of enrichment drops (NSFW-in-SFW, deleted-from-wiki)
	// can't fan out unboundedly. Each round is one keyset slice of `limit` rows.
	activityMaxRounds = 12
	// wikiBatchChunk caps ids per GetBatchPublic call — it sends every id in one
	// query param (no chunking). Matches the ≤100 batch convention.
	wikiBatchChunk = 100
)

type ActivityService struct {
	repo       *repository.ActivityRepository
	wikiGC     *galgameClient.GalgameClient
	userClient *userclient.Client
	rdb        *redis.Client
}

func NewActivityService(
	repo *repository.ActivityRepository,
	gc *galgameClient.GalgameClient,
	userClient *userclient.Client,
	rdb *redis.Client,
) *ActivityService {
	return &ActivityService{repo: repo, wikiGC: gc, userClient: userClient, rdb: rdb}
}

// Result is one page of activity plus the opaque keyset cursor for the next
// page ("" when there are no more rows behind it).
type Result struct {
	Items      []dto.ActivityItem `json:"items"`
	NextCursor string             `json:"nextCursor"`
}

// GetActivity returns a filtered activity feed. type "all" → the mixed
// timeline. isSFW is forwarded to wiki so NSFW galgame names never enter the
// public activity stream (docs/galgame_wiki/00-handbook §16).
func (s *ActivityService) GetActivity(ctx context.Context, typeStr, cursor string, limit int, isSFW, showNoResource bool) (*Result, *errors.AppError) {
	if typeStr == "all" {
		return s.GetTimeline(ctx, cursor, limit, isSFW, showNoResource)
	}
	src, ok := s.repo.GetSource(typeStr)
	if !ok {
		return &Result{Items: []dto.ActivityItem{}, NextCursor: ""}, nil
	}
	cacheKey := fmt.Sprintf("activity:v2:%s:%s:%d:%t:%t", typeStr, cursor, limit, isSFW, showNoResource)
	return s.cachedKeyset(ctx, cacheKey, []repository.ActivitySource{src}, cursor, limit, isSFW, showNoResource)
}

// getCachedResult returns a cached activity Result, and true on a hit.
// Fail-open: redis.Nil (miss) or ANY error (redis down, bad JSON) is treated as
// a miss so the request still serves fresh data.
func (s *ActivityService) getCachedResult(ctx context.Context, key string) (*Result, bool) {
	if s.rdb == nil {
		return nil, false
	}
	raw, err := s.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}
	var result Result
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, false
	}
	return &result, true
}

// cacheResult best-effort stores the enriched result with activityCacheTTL.
// A redis/marshal failure is ignored — caching must never break the response.
func (s *ActivityService) cacheResult(ctx context.Context, key string, result *Result) {
	if s.rdb == nil || result == nil {
		return
	}
	raw, err := json.Marshal(result)
	if err != nil {
		return
	}
	_ = s.rdb.Set(ctx, key, raw, activityCacheTTL).Err()
}

// GetTimeline returns the mixed activity timeline across all sources.
func (s *ActivityService) GetTimeline(ctx context.Context, cursor string, limit int, isSFW, showNoResource bool) (*Result, *errors.AppError) {
	cacheKey := fmt.Sprintf("activity:v2:all:%s:%d:%t:%t", cursor, limit, isSFW, showNoResource)
	return s.cachedKeyset(ctx, cacheKey, s.repo.AllSources(), cursor, limit, isSFW, showNoResource)
}

// cachedKeyset wraps serveKeyset with the activityCacheTTL cache-aside. Keyed by
// cursor (the page), isSFW (which galgame names the wiki returns) and
// showNoResource (whether resource-less GALGAME_CREATION rows are dropped) so
// viewers with different filters never share an entry.
func (s *ActivityService) cachedKeyset(ctx context.Context, cacheKey string, sources []repository.ActivitySource, cursor string, limit int, isSFW, showNoResource bool) (*Result, *errors.AppError) {
	if cached, ok := s.getCachedResult(ctx, cacheKey); ok {
		return cached, nil
	}
	result, appErr := s.serveKeyset(ctx, sources, cursor, limit, isSFW, showNoResource)
	if appErr != nil {
		return nil, appErr
	}
	s.cacheResult(ctx, cacheKey, result)
	return result, nil
}

// serveKeyset fills one page to `limit` survivors despite enrichment drops. It
// seeks the deterministic keyset (created, type_str, id) from `cursor`; because
// enrichment can drop rows (NSFW-in-SFW, deleted-from-wiki), it keeps fetching
// the next slice — advancing the cursor past each fully-consumed batch — until
// it has `limit` survivors or the feed is exhausted. nextCursor is the LAST
// survivor's keyset, so the next page resumes exactly where this one stopped:
// no offset drift, and the total-order tiebreaker means no dup/skip across pages.
func (s *ActivityService) serveKeyset(ctx context.Context, sources []repository.ActivitySource, cursor string, limit int, isSFW, showNoResource bool) (*Result, *errors.AppError) {
	cur := decodeCursor(cursor)
	collected := make([]dto.ActivityItem, 0, limit)
	exhausted := false

	for round := 0; len(collected) < limit && round < activityMaxRounds; round++ {
		rows, err := s.repo.FetchKeyset(sources, limit, cur, isSFW, showNoResource)
		if err != nil {
			return nil, errors.ErrInternal("查询活动数据失败")
		}
		if len(rows) == 0 {
			exhausted = true
			break
		}
		for _, it := range s.enrichAndHydrate(ctx, rows, isSFW) {
			collected = append(collected, it)
			if len(collected) == limit {
				break
			}
		}
		if len(collected) >= limit {
			break // page full — nextCursor is the last collected survivor
		}
		// Whole batch consumed without filling the page; advance past its last
		// RAW row (some were dropped by enrichment) and fetch the next slice.
		last := rows[len(rows)-1]
		cur = &repository.Cursor{Created: last.Created, TypeStr: last.TypeStr, ID: last.ID}
		if len(rows) < limit {
			exhausted = true
			break // short batch → nothing behind it
		}
	}

	next := ""
	switch {
	case len(collected) == limit:
		// Full page — resume after the last survivor (one extra empty fetch if
		// this happened to be the very tail; harmless).
		last := collected[len(collected)-1]
		next = encodeCursor(last.Timestamp, last.Type, last.ID)
	case !exhausted && cur != nil:
		// Hit activityMaxRounds before filling the page (heavy enrichment drops):
		// resume from where we stopped rather than signalling a false end-of-feed.
		next = encodeCursor(cur.Created, cur.TypeStr, cur.ID)
	}
	return &Result{Items: collected, NextCursor: next}, nil
}

// ──────────────────────────────────────────
// Keyset cursor codec
// ──────────────────────────────────────────

// cursorPayload is the JSON behind the opaque base64 cursor: the last row's
// (created, type_str, id). type_str is required because `id` is unique only
// within a source — the feed UNIONs many tables.
type cursorPayload struct {
	C time.Time `json:"c"`
	T string    `json:"t"`
	I int       `json:"i"`
}

func encodeCursor(created time.Time, typeStr string, id int) string {
	b, err := json.Marshal(cursorPayload{C: created, T: typeStr, I: id})
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// decodeCursor parses an opaque cursor into a repository.Cursor. Empty → nil
// (first page). A malformed cursor also → nil: a stale/garbage token just
// restarts from the newest rather than erroring the whole feed.
func decodeCursor(cursor string) *repository.Cursor {
	if cursor == "" {
		return nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return nil
	}
	var p cursorPayload
	if err := json.Unmarshal(raw, &p); err != nil || p.T == "" || p.I <= 0 {
		return nil
	}
	return &repository.Cursor{Created: p.C, TypeStr: p.T, ID: p.I}
}

// enrichAndHydrate runs the full row → item pipeline: galgame name/brief
// enrichment (which drops brief-missing rows) then OAuth identity hydration.
func (s *ActivityService) enrichAndHydrate(ctx context.Context, rows []repository.ActivityRow, isSFW bool) []dto.ActivityItem {
	items := rowsToItems(rows)
	items = s.enrichGalgameItems(ctx, rows, items, isSFW)
	s.hydrateActors(ctx, items)
	return items
}

// rowsToItems converts DB rows into response items (no enrichment yet).
// Identity is left blank — hydrated by hydrateActors after enrichGalgameItems
// has had a chance to inject GALGAME_CREATION actor IDs from the wiki brief.
func rowsToItems(rows []repository.ActivityRow) []dto.ActivityItem {
	items := make([]dto.ActivityItem, len(rows))
	for i, r := range rows {
		items[i] = dto.ActivityItem{
			ID:        r.ID,
			UniqueID:  fmt.Sprintf("%s-%d", r.TypeStr, r.ID),
			Type:      r.TypeStr,
			Content:   r.Content,
			Link:      r.Link,
			Timestamp: r.Created,
			Actor: dto.Actor{
				ID: r.UserID,
			},
		}
	}
	return items
}

// enrichGalgameItems batch-fetches names for every galgame-scoped activity
// row from the wiki service and rewrites the content string per type:
//
//	GALGAME_CREATION          → "<game name>"
//	GALGAME_RESOURCE_CREATION → "在《<game name>》发布了下载资源"
//	GALGAME_RATING_CREATION   → "<game name> · <short summary>" (name only when no summary)
//
// GALGAME_COMMENT_CREATION is intentionally NOT rewritten: it keeps the raw
// comment text (matching the legacy API) since the type chip + link already
// convey what it is and where it points — a "在《game》" prefix is just noise.
//
// Galgame-scoped rows whose galgame has NO wiki brief for this viewer — NSFW in
// SFW mode, or deleted — are DROPPED from the returned slice, so an NSFW galgame
// never leaks into the public feed as "galgame#N / 未知用户". On wiki failure the
// items are returned untouched (graceful degradation).
//
// Resource-less GALGAME_CREATION rows are NOT handled here anymore — they're
// excluded in SQL (sourceQuery) so they never spend a LIMIT slot; see the repo.
//
// Returns the kept items (a subset of `items`); rows/items must be
// index-aligned on entry.
func (s *ActivityService) enrichGalgameItems(
	ctx context.Context,
	rows []repository.ActivityRow,
	items []dto.ActivityItem,
	isSFW bool,
) []dto.ActivityItem {
	idSet := map[int]struct{}{}
	for _, r := range rows {
		if r.GalgameID > 0 {
			idSet[r.GalgameID] = struct{}{}
		}
	}
	if len(idSet) == 0 {
		return items
	}
	ids := make([]int, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}

	// Chunk: GetBatchPublic sends every id in one query param, and the over-fetch
	// window can surface more distinct galgames than a single batch should carry.
	briefMap := make(map[int]galgameClient.GalgameBrief, len(ids))
	for start := 0; start < len(ids); start += wikiBatchChunk {
		end := min(start+wikiBatchChunk, len(ids))
		m, appErr := s.wikiGC.GetBatchPublic(ctx, ids[start:end], isSFW)
		if appErr != nil {
			return items // graceful: wiki down, leave items untouched
		}
		maps.Copy(briefMap, m)
	}

	briefName := func(b galgameClient.GalgameBrief) string {
		// zh-cn > zh-tw > ja-jp > en-us — the FE getPreferredLanguageText zh-cn
		// default; en-US (VNDB romaji) last so a JP/CN game never shows it.
		for _, n := range []string{b.NameZhCn, b.NameZhTw, b.NameJaJp, b.NameEnUs} {
			if n != "" {
				return n
			}
		}
		return fmt.Sprintf("galgame#%d", b.ID)
	}

	kept := make([]dto.ActivityItem, 0, len(items))
	for i, r := range rows {
		if r.GalgameID == 0 {
			kept = append(kept, items[i]) // not galgame-scoped: always keep
			continue
		}
		b, ok := briefMap[r.GalgameID]
		if !ok {
			// NSFW (SFW viewer) or deleted → drop the whole activity.
			continue
		}
		name := briefName(b)
		switch r.TypeStr {
		case "GALGAME_CREATION":
			items[i].Content = name
			// galgame table has no local user_id; pull the creator from the brief.
			if items[i].Actor.ID == 0 {
				items[i].Actor.ID = b.UserID
			}
		case "GALGAME_RESOURCE_CREATION":
			items[i].Content = fmt.Sprintf("在《%s》发布了下载资源", name)
		case "GALGAME_EDIT":
			items[i].Content = fmt.Sprintf("编辑了《%s》", name)
		case "GALGAME_PR_CREATION":
			items[i].Content = fmt.Sprintf("对《%s》提出了更新请求", name)
		case "GALGAME_RATING_CREATION":
			if r.Content != "" {
				items[i].Content = fmt.Sprintf("%s · %s", name, r.Content)
			} else {
				items[i].Content = name
			}
		}
		kept = append(kept, items[i])
	}
	return kept
}

// hydrateActors batch-fetches identity (name/avatar) from OAuth for every
// non-zero actor id and writes it back into the items. Runs after
// enrichGalgameItems so GALGAME_CREATION rows whose actor was injected from
// the wiki brief are also hydrated.
func (s *ActivityService) hydrateActors(ctx context.Context, items []dto.ActivityItem) {
	uids := userclient.CollectIDs(items, func(it dto.ActivityItem) int { return it.Actor.ID })
	if len(uids) == 0 {
		return
	}
	userMap := s.userClient.Hydrate(ctx, uids)
	for i := range items {
		u := userMap[items[i].Actor.ID]
		items[i].Actor.Name = u.Name
		items[i].Actor.Avatar = u.Avatar
	}
}
