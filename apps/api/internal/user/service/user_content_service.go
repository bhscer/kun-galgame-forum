package service

import (
	"context"
	"encoding/json"

	galgameClient "kun-galgame-api/internal/galgame/client"
	"kun-galgame-api/internal/infrastructure/markdown"
	"kun-galgame-api/internal/user/dto"
	"kun-galgame-api/internal/user/repository"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/userclient"
)

type UserContentService struct {
	userContentRepo *repository.UserContentRepository
	wikiClient      *galgameClient.GalgameClient
	userClient      *userclient.Client
}

func NewUserContentService(
	userContentRepo *repository.UserContentRepository,
	wikiClient *galgameClient.GalgameClient,
	userClient *userclient.Client,
) *UserContentService {
	return &UserContentService{
		userContentRepo: userContentRepo,
		wikiClient:      wikiClient,
		userClient:      userClient,
	}
}

// ──────────────────────────────────────────
// User galgame list — GET /user/:userID/galgames
// ──────────────────────────────────────────

// GetUserGalgameCards returns enriched galgame cards for the user's list
// (created / liked / favorited / commented depending on req.Type).
//
// SFW filter applied at the service layer against each card's wiki brief —
// kungal's `galgame` table has no content_limit column, so SQL-level
// filtering would need a schema sync. `total` over-reports in SFW mode;
// same SEO-safe trade-off as galgame_service.GetList.
func (s *UserContentService) GetUserGalgameCards(
	ctx context.Context,
	userID int,
	req *dto.UserGalgamesRequest,
	isSFW bool,
) ([]dto.UserGalgameCard, int64, *errors.AppError) {
	ids, total, err := s.userContentRepo.FindUserGalgameIDs(userID, req.Type, req.Page, req.Limit)
	if err != nil {
		return nil, 0, errors.ErrInternal("获取用户 Galgame 列表失败")
	}
	if len(ids) == 0 {
		return []dto.UserGalgameCard{}, total, nil
	}

	briefMap, wikiErr := s.wikiClient.GetBatch(ctx, ids)
	if wikiErr != nil {
		// Wiki failure → return empty list but preserve total count.
		return []dto.UserGalgameCard{}, total, nil
	}

	localMap := s.userContentRepo.FindGalgameLocalStats(ids)
	metaRows := s.userContentRepo.FindResourceMetaByGalgameIDs(ids)
	platformMap, languageMap := groupResourceMeta(metaRows)

	userIDs := collectUniqueIDs(values(briefMap), func(b galgameClient.GalgameBrief) int { return b.UserID })
	userMap := s.userClient.Hydrate(ctx, userIDs)

	cards := make([]dto.UserGalgameCard, 0, len(ids))
	for _, id := range ids {
		b, ok := briefMap[id]
		if !ok {
			continue
		}
		if isSFW && b.ContentLimit != "sfw" {
			continue
		}
		l := localMap[id]
		u := userMap[b.UserID]
		cards = append(cards, dto.UserGalgameCard{
			ID:                 b.ID,
			Name:               briefToLocale(b),
			Banner:             b.Banner,
			User:               dto.UserBrief{ID: u.ID, Name: u.Name, Avatar: u.Avatar},
			ContentLimit:       b.ContentLimit,
			View:               l.View,
			LikeCount:          l.LikeCount,
			ResourceUpdateTime: b.ResourceUpdateTime,
			Platform:           emptyStrSlice(platformMap[id]),
			Language:           emptyStrSlice(languageMap[id]),
			// U2: pass through the wiki-derived banner so the FE card
			// can pick `_mini` instead of falling back to empty legacy
			// `banner` for newly-uploaded galgames.
			EffectiveBannerHash: b.EffectiveBannerHash,
			EffectiveBannerURL:  b.EffectiveBannerURL,
		})
	}
	return cards, total, nil
}

// values is a tiny helper that extracts map values into a slice.
func values[K comparable, V any](m map[K]V) []V {
	out := make([]V, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}

// ──────────────────────────────────────────
// Topics / replies / comments (already thin)
// ──────────────────────────────────────────

func (s *UserContentService) GetUserTopics(ctx context.Context, userID int, req *dto.UserTopicsRequest, isSFW bool) ([]dto.UserTopic, int64, *errors.AppError) {
	items, total, err := s.userContentRepo.FindUserTopics(userID, req.Type, req.Page, req.Limit, isSFW)
	if err != nil {
		return nil, 0, errors.ErrInternal("获取用户话题列表失败")
	}
	return items, total, nil
}

func (s *UserContentService) GetUserReplies(ctx context.Context, userID int, req *dto.UserRepliesRequest) ([]repository.UserReply, int64, *errors.AppError) {
	items, total, err := s.userContentRepo.FindUserReplies(userID, req.Type, req.Page, req.Limit)
	if err != nil {
		return nil, 0, errors.ErrInternal("获取用户回复列表失败")
	}
	return items, total, nil
}

func (s *UserContentService) GetUserComments(ctx context.Context, userID int, req *dto.UserCommentsRequest) ([]repository.UserComment, int64, *errors.AppError) {
	items, total, err := s.userContentRepo.FindUserComments(userID, req.Type, req.Page, req.Limit)
	if err != nil {
		return nil, 0, errors.ErrInternal("获取用户评论列表失败")
	}
	return items, total, nil
}

// GetUserGalgameComments returns the comment-card data for the
// "评论 / 被评论 / 点赞评论" tabs under /user/:id/galgame/.
// Author identity comes from userclient; content is rendered via the
// project goldmark pipeline so the frontend can drop it into
// <KunContent> consistently with the rest of the site.
func (s *UserContentService) GetUserGalgameComments(
	ctx context.Context,
	userID int,
	req *dto.UserGalgameCommentsRequest,
) ([]dto.UserGalgameComment, int64, *errors.AppError) {
	rows, total, err := s.userContentRepo.FindUserGalgameComments(userID, req.Type, req.Page, req.Limit)
	if err != nil {
		return nil, 0, errors.ErrInternal("获取用户 Galgame 评论列表失败")
	}
	if len(rows) == 0 {
		return []dto.UserGalgameComment{}, total, nil
	}

	uidSet := make(map[int]struct{}, len(rows))
	for _, r := range rows {
		uidSet[r.UserID] = struct{}{}
	}
	uids := make([]int, 0, len(uidSet))
	for id := range uidSet {
		uids = append(uids, id)
	}
	userMap := s.userClient.Hydrate(ctx, uids)

	items := make([]dto.UserGalgameComment, 0, len(rows))
	for _, r := range rows {
		u := userMap[r.UserID]
		items = append(items, dto.UserGalgameComment{
			ID:          r.ID,
			GalgameID:   r.GalgameID,
			Content:     r.Content,
			ContentHtml: markdown.Render(r.Content),
			User:        dto.UserBrief{ID: u.ID, Name: u.Name, Avatar: u.Avatar},
			Created:     r.CreatedAt,
		})
	}
	return items, total, nil
}

// ──────────────────────────────────────────
// Resources — GET /user/:userID/resources
// ──────────────────────────────────────────

func (s *UserContentService) GetUserResources(
	ctx context.Context,
	userID int,
	req *dto.UserResourcesRequest,
	isSFW bool,
) (*dto.UserResourcesResponse, *errors.AppError) {
	rows, total, err := s.userContentRepo.FindUserResources(userID, req.Type, req.Page, req.Limit)
	if err != nil {
		return nil, errors.ErrInternal("获取用户资源列表失败")
	}

	resourceIDs := make([]int, len(rows))
	galgameIDs := collectUniqueIDs(rows, func(r repository.UserResource) int { return r.GalgameID })
	for i, r := range rows {
		resourceIDs[i] = r.ID
	}

	var linkMap map[int][]string
	if len(resourceIDs) > 0 {
		linkMap, _ = s.userContentRepo.FindResourceLinks(resourceIDs)
	}

	var briefMap map[int]galgameClient.GalgameBrief
	if len(galgameIDs) > 0 {
		briefMap, _ = s.wikiClient.GetBatch(ctx, galgameIDs)
	}

	// SFW filter at service layer (content_limit lives on wiki briefs,
	// not on galgame_resource). Total over-reports in SFW mode — same
	// trade-off as galgame_service.GetList.
	items := make([]dto.UserResourceItem, 0, len(rows))
	for _, r := range rows {
		b, hasBrief := briefMap[r.GalgameID]
		if isSFW && (!hasBrief || b.ContentLimit != "sfw") {
			continue
		}
		links := linkMap[r.ID]
		if links == nil {
			links = []string{}
		}
		name := emptyLocale()
		if hasBrief {
			name = briefToLocale(b)
		}
		items = append(items, dto.UserResourceItem{
			ID:          r.ID,
			GalgameID:   r.GalgameID,
			GalgameName: name,
			Type:        r.Type,
			Language:    r.Language,
			Platform:    r.Platform,
			Size:        r.Size,
			Link:        links,
			Code:        r.Code,
			Password:    r.Password,
			Note:        r.Note,
			Status:      r.Status,
			Created:     r.Created,
		})
	}

	return &dto.UserResourcesResponse{Resources: items, Total: total}, nil
}

// ──────────────────────────────────────────
// Ratings — GET /user/:userID/ratings
// ──────────────────────────────────────────

func (s *UserContentService) GetUserRatings(
	ctx context.Context,
	userID int,
	req *dto.UserRatingsRequest,
	isSFW bool,
) (*dto.UserRatingsResponse, *errors.AppError) {
	rows, total, err := s.userContentRepo.FindUserRatings(userID, req.Page, req.Limit)
	if err != nil {
		return nil, errors.ErrInternal("获取用户评分列表失败")
	}

	galgameIDs := collectUniqueIDs(rows, func(r repository.UserRating) int { return r.GalgameID })
	var briefMap map[int]galgameClient.GalgameBrief
	if len(galgameIDs) > 0 {
		briefMap, _ = s.wikiClient.GetBatch(ctx, galgameIDs)
	}

	uids := collectUniqueIDs(rows, func(r repository.UserRating) int { return r.UserID })
	userMap := s.userClient.Hydrate(ctx, uids)

	// SFW filter at service layer — galgame_rating has no content_limit,
	// only the wiki brief does. Total over-reports in SFW mode.
	items := make([]dto.UserRatingItem, 0, len(rows))
	for _, r := range rows {
		b, hasBrief := briefMap[r.GalgameID]
		if isSFW && (!hasBrief || b.ContentLimit != "sfw") {
			continue
		}
		var galgameType []string
		if r.GalgameType != "" {
			_ = json.Unmarshal([]byte(r.GalgameType), &galgameType)
		}

		galgame := dto.UserRatingGalgame{ID: r.GalgameID}
		if hasBrief {
			galgame = dto.UserRatingGalgame{
				ID:           b.ID,
				Name:         briefToLocale(b),
				ContentLimit: b.ContentLimit,
			}
		}

		u := userMap[r.UserID]
		items = append(items, dto.UserRatingItem{
			ID:           r.ID,
			User:         dto.UserBrief{ID: u.ID, Name: u.Name, Avatar: u.Avatar},
			Recommend:    r.Recommend,
			Overall:      r.Overall,
			View:         r.View,
			GalgameType:  galgameType,
			PlayStatus:   r.PlayStatus,
			Art:          r.Art,
			Story:        r.Story,
			Music:        r.Music,
			Character:    r.Character,
			Route:        r.Route,
			System:       r.System,
			Voice:        r.Voice,
			ReplayValue:  r.ReplayValue,
			SpoilerLevel: r.SpoilerLevel,
			LikeCount:    r.LikeCount,
			Created:      r.Created,
			Updated:      r.Updated,
			Galgame:      galgame,
		})
	}

	return &dto.UserRatingsResponse{RatingData: items, Total: total}, nil
}
