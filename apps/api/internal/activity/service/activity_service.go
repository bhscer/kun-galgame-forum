package service

import (
	"context"
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
	// activityShallowMax: pages whose page*limit ≤ this get the over-fetch+slice
	// treatment that fills them to `limit` despite enrichment drops. Deeper pages
	// fall back to the legacy offset fetch (cheap, sparse, but non-empty).
	activityShallowMax = 100
	// activityOverfetchFactor / Max: a shallow page fetches page*limit*factor raw
	// rows (capped at Max) so the post-fetch enrichment drop (resource-less rows
	// are already gone via SQL; this absorbs the deleted-from-wiki tail) still
	// leaves ≥ limit survivors to slice. Capped to bound enrichment cost and the
	// wiki-batch fan-out.
	activityOverfetchFactor = 2
	activityOverfetchMax    = 200
	// wikiBatchChunk caps ids per GetBatchPublic call — it sends every id in one
	// query param (no chunking), and over-fetch can surface more distinct
	// galgames than a single batch should carry. Matches the ≤100 batch convention.
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

// Result holds a paginated activity list.
type Result struct {
	Items []dto.ActivityItem
	Total int64
}

// GetActivity returns a filtered activity feed. If the type is "all",
// it falls back to GetTimeline. isSFW is forwarded to wiki so NSFW
// galgame names never enter the public activity stream
// (docs/galgame_wiki/00-handbook §16).
func (s *ActivityService) GetActivity(ctx context.Context, typeStr string, page, limit int, isSFW, showNoResource bool) (*Result, *errors.AppError) {
	// Cache-aside on the fully-enriched result. Keyed by isSFW (changes which
	// galgame names the wiki returns — NSFW filter) and showNoResource (changes
	// whether resource-less galgames' GALGAME_CREATION rows are dropped), so
	// viewers with different filters must not share an entry.
	cacheKey := fmt.Sprintf("activity:v1:%s:%d:%d:%t:%t", typeStr, page, limit, isSFW, showNoResource)
	if cached, ok := s.getCachedResult(ctx, cacheKey); ok {
		return cached, nil
	}

	result, appErr := s.computeActivity(ctx, typeStr, page, limit, isSFW, showNoResource)
	if appErr != nil {
		return nil, appErr
	}
	s.cacheResult(ctx, cacheKey, result)
	return result, nil
}

// computeActivity is GetActivity without the cache layer.
func (s *ActivityService) computeActivity(ctx context.Context, typeStr string, page, limit int, isSFW, showNoResource bool) (*Result, *errors.AppError) {
	if typeStr == "all" {
		return s.GetTimeline(ctx, page, limit, isSFW, showNoResource)
	}

	src, ok := s.repo.GetSource(typeStr)
	if !ok {
		return &Result{Items: []dto.ActivityItem{}, Total: 0}, nil
	}

	total, err := s.repo.CountSingleSource(src, isSFW, showNoResource)
	if err != nil {
		return nil, errors.ErrInternal("查询活动数据失败")
	}
	items, appErr := s.servePage(ctx, page, limit, isSFW,
		func(n int) ([]repository.ActivityRow, error) {
			return s.repo.FetchSingleSourceRows(src, n, isSFW, showNoResource)
		},
		func(p, l int) ([]repository.ActivityRow, error) {
			return s.repo.FetchSingleSourcePage(src, p, l, isSFW, showNoResource)
		},
	)
	if appErr != nil {
		return nil, appErr
	}
	return &Result{Items: items, Total: total}, nil
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

// GetTimeline returns a mixed activity timeline across all sources.
func (s *ActivityService) GetTimeline(ctx context.Context, page, limit int, isSFW, showNoResource bool) (*Result, *errors.AppError) {
	total, err := s.repo.CountTimeline(isSFW, showNoResource)
	if err != nil {
		return nil, errors.ErrInternal("查询活动列表失败")
	}
	items, appErr := s.servePage(ctx, page, limit, isSFW,
		func(n int) ([]repository.ActivityRow, error) {
			return s.repo.FetchTimelineRows(n, isSFW, showNoResource)
		},
		func(p, l int) ([]repository.ActivityRow, error) {
			return s.repo.FetchTimelinePage(p, l, isSFW, showNoResource)
		},
	)
	if appErr != nil {
		return nil, appErr
	}
	return &Result{Items: items, Total: total}, nil
}

// servePage fills a page to `limit` items despite enrichment drops. Shallow
// pages over-fetch a top-of-feed window and slice the requested page in
// *filtered* space (after enrichment), so a page no longer under-fills when
// galgame rows get dropped. Deep pages fall back to the legacy offset fetch
// (cheap, sparse, unchanged) — filtered-space slicing from the top would mean
// re-enriching everything before them.
func (s *ActivityService) servePage(
	ctx context.Context, page, limit int, isSFW bool,
	fetchNewest func(n int) ([]repository.ActivityRow, error),
	fetchPage func(p, l int) ([]repository.ActivityRow, error),
) ([]dto.ActivityItem, *errors.AppError) {
	if page*limit <= activityShallowMax {
		window := min(page*limit*activityOverfetchFactor, activityOverfetchMax)
		rows, err := fetchNewest(window)
		if err != nil {
			return nil, errors.ErrInternal("查询活动数据失败")
		}
		return pageSlice(s.enrichAndHydrate(ctx, rows, isSFW), page, limit), nil
	}

	rows, err := fetchPage(page, limit)
	if err != nil {
		return nil, errors.ErrInternal("查询活动数据失败")
	}
	return s.enrichAndHydrate(ctx, rows, isSFW), nil
}

// enrichAndHydrate runs the full row → item pipeline: galgame name/brief
// enrichment (which drops brief-missing rows) then OAuth identity hydration.
func (s *ActivityService) enrichAndHydrate(ctx context.Context, rows []repository.ActivityRow, isSFW bool) []dto.ActivityItem {
	items := rowsToItems(rows)
	items = s.enrichGalgameItems(ctx, rows, items, isSFW)
	s.hydrateActors(ctx, items)
	return items
}

// pageSlice returns the [(page-1)*limit, page*limit) slice of the enriched
// (filtered) items, clamped to the available length.
func pageSlice(items []dto.ActivityItem, page, limit int) []dto.ActivityItem {
	start := (page - 1) * limit
	if start >= len(items) {
		return []dto.ActivityItem{}
	}
	end := min(start+limit, len(items))
	return items[start:end]
}

// rowsToItems converts DB rows into response items (no enrichment yet).
// Identity is left blank — hydrated by hydrateActors after enrichGalgameItems
// has had a chance to inject GALGAME_CREATION actor IDs from the wiki brief.
func rowsToItems(rows []repository.ActivityRow) []dto.ActivityItem {
	items := make([]dto.ActivityItem, len(rows))
	for i, r := range rows {
		items[i] = dto.ActivityItem{
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
