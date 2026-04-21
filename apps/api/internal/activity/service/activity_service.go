package service

import (
	"context"
	"fmt"

	"kun-galgame-api/internal/activity/dto"
	"kun-galgame-api/internal/activity/repository"
	galgameClient "kun-galgame-api/internal/galgame/client"
	"kun-galgame-api/pkg/errors"
)

type ActivityService struct {
	repo   *repository.ActivityRepository
	wikiGC *galgameClient.GalgameClient
}

func NewActivityService(
	repo *repository.ActivityRepository,
	gc *galgameClient.GalgameClient,
) *ActivityService {
	return &ActivityService{repo: repo, wikiGC: gc}
}

// Result holds a paginated activity list.
type Result struct {
	Items []dto.ActivityItem
	Total int64
}

// GetActivity returns a filtered activity feed. If the type is "all",
// it falls back to GetTimeline.
func (s *ActivityService) GetActivity(ctx context.Context, typeStr string, page, limit int) (*Result, *errors.AppError) {
	if typeStr == "all" {
		return s.GetTimeline(ctx, page, limit)
	}

	src, ok := s.repo.GetSource(typeStr)
	if !ok {
		return &Result{Items: []dto.ActivityItem{}, Total: 0}, nil
	}

	rows, total, err := s.repo.FetchSingleSource(src, page, limit)
	if err != nil {
		return nil, errors.ErrInternal("查询活动数据失败")
	}
	items := rowsToItems(rows)
	s.enrichGalgameItems(ctx, rows, items)
	return &Result{Items: items, Total: total}, nil
}

// GetTimeline returns a mixed activity timeline across all sources.
func (s *ActivityService) GetTimeline(ctx context.Context, page, limit int) (*Result, *errors.AppError) {
	rows, total, err := s.repo.FetchTimeline(page, limit)
	if err != nil {
		return nil, errors.ErrInternal("查询活动列表失败")
	}
	items := rowsToItems(rows)
	s.enrichGalgameItems(ctx, rows, items)
	return &Result{Items: items, Total: total}, nil
}

// rowsToItems converts DB rows into response items (no enrichment yet).
// The raw DB content is stashed in `Content` and may be replaced during
// enrichment with a galgame-name-aware string for galgame-scoped types.
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
				ID: r.UserID, Name: r.UserName, Avatar: r.Avatar,
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
//	GALGAME_RATING_CREATION   → "<game name> · <short summary>" (if summary)
//	GALGAME_COMMENT_CREATION  → "在《<game name>》<comment>"
//
// rows/items must be index-aligned; the caller guarantees this.
func (s *ActivityService) enrichGalgameItems(
	ctx context.Context,
	rows []repository.ActivityRow,
	items []dto.ActivityItem,
) {
	idSet := map[int]struct{}{}
	for _, r := range rows {
		if r.GalgameID > 0 {
			idSet[r.GalgameID] = struct{}{}
		}
	}
	if len(idSet) == 0 {
		return
	}
	ids := make([]int, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}

	briefMap, appErr := s.wikiGC.GetBatch(ctx, ids)
	if appErr != nil {
		return // graceful: leave raw content
	}

	pickName := func(id int) string {
		b, ok := briefMap[id]
		if !ok {
			return fmt.Sprintf("galgame#%d", id)
		}
		for _, n := range []string{b.NameZhCn, b.NameJaJp, b.NameEnUs, b.NameZhTw} {
			if n != "" {
				return n
			}
		}
		return fmt.Sprintf("galgame#%d", id)
	}

	for i, r := range rows {
		if r.GalgameID == 0 {
			continue
		}
		name := pickName(r.GalgameID)
		switch r.TypeStr {
		case "GALGAME_CREATION":
			items[i].Content = name
			// galgame table has no local user_id; pull the creator from
			// the wiki brief.
			if items[i].Actor.ID == 0 {
				if b, ok := briefMap[r.GalgameID]; ok {
					items[i].Actor.ID = b.UserID
				}
			}
		case "GALGAME_RESOURCE_CREATION":
			items[i].Content = fmt.Sprintf("在《%s》发布了下载资源", name)
		case "GALGAME_RATING_CREATION":
			if r.Content != "" {
				items[i].Content = fmt.Sprintf("%s · %s", name, r.Content)
			} else {
				items[i].Content = fmt.Sprintf("评价了《%s》", name)
			}
		case "GALGAME_COMMENT_CREATION":
			if r.Content != "" {
				items[i].Content = fmt.Sprintf("在《%s》%s", name, r.Content)
			} else {
				items[i].Content = fmt.Sprintf("评论了《%s》", name)
			}
		}
	}

	// Resolve display name/avatar for galgame-creation actors whose ID
	// was just injected from the wiki brief (LEFT JOIN missed them at
	// query time because user_id was 0).
	needUsers := make([]int, 0)
	for _, it := range items {
		if it.Type == "GALGAME_CREATION" && it.Actor.Name == "" && it.Actor.ID > 0 {
			needUsers = append(needUsers, it.Actor.ID)
		}
	}
	if len(needUsers) == 0 {
		return
	}
	userMap := map[int]repository.UserInfoRow{}
	for _, u := range s.repo.FindUsersByIDs(needUsers) {
		userMap[u.ID] = u
	}
	for i := range items {
		if items[i].Type == "GALGAME_CREATION" && items[i].Actor.Name == "" {
			if u, ok := userMap[items[i].Actor.ID]; ok {
				items[i].Actor.Name = u.Name
				items[i].Actor.Avatar = u.Avatar
			}
		}
	}
}
