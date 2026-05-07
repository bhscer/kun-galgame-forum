package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	galgameClient "kun-galgame-api/internal/galgame/client"
	galgameDto "kun-galgame-api/internal/galgame/dto"
	galgameService "kun-galgame-api/internal/galgame/service"
	"kun-galgame-api/internal/search/dto"
	"kun-galgame-api/internal/search/repository"
	"kun-galgame-api/pkg/errors"
)

type SearchService struct {
	repo       *repository.SearchRepository
	wikiClient *galgameClient.GalgameClient
	enricher   *galgameService.GalgameEnricher
}

func NewSearchService(
	repo *repository.SearchRepository,
	wikiClient *galgameClient.GalgameClient,
	enricher *galgameService.GalgameEnricher,
) *SearchService {
	return &SearchService{repo: repo, wikiClient: wikiClient, enricher: enricher}
}

// tokenize splits a keyword string into trimmed non-empty tokens.
// Returns an error if the result is empty.
func tokenize(raw string) ([]string, *errors.AppError) {
	keywords := strings.Fields(strings.TrimSpace(raw))
	if len(keywords) == 0 {
		return nil, errors.ErrBadRequest("搜索关键词不能为空")
	}
	return keywords, nil
}

// SearchTopics returns topic search results.
func (s *SearchService) SearchTopics(raw string, page, limit int) (*dto.PaginatedResult[dto.TopicItem], *errors.AppError) {
	keywords, appErr := tokenize(raw)
	if appErr != nil {
		return nil, appErr
	}
	rows, total := s.repo.SearchTopics(keywords, page, limit)

	items := make([]dto.TopicItem, len(rows))
	for i, r := range rows {
		items[i] = dto.TopicItem{
			ID: r.ID, Title: r.Title, View: r.View, Status: r.Status,
			LikeCount: r.LikeCount, ReplyCount: r.ReplyCount,
			CommentCount: r.CommentCount, StatusUpdateTime: r.StatusUpdateTime,
			User: dto.UserBrief{ID: r.UserID, Name: r.UserName, Avatar: r.UserAvatar},
		}
	}
	return &dto.PaginatedResult[dto.TopicItem]{Items: items, Total: total}, nil
}

// SearchUsers returns user search results.
func (s *SearchService) SearchUsers(raw string, page, limit int) (*dto.PaginatedResult[dto.UserItem], *errors.AppError) {
	keywords, appErr := tokenize(raw)
	if appErr != nil {
		return nil, appErr
	}
	rows, total := s.repo.SearchUsers(keywords, page, limit)

	items := make([]dto.UserItem, len(rows))
	for i, r := range rows {
		items[i] = dto.UserItem{
			ID: r.ID, Name: r.Name, Avatar: r.Avatar, Bio: r.Bio,
			Moemoepoint: r.Moemoepoint, Created: r.Created,
		}
	}
	return &dto.PaginatedResult[dto.UserItem]{Items: items, Total: total}, nil
}

// SearchReplies returns reply search results.
func (s *SearchService) SearchReplies(raw string, page, limit int) (*dto.PaginatedResult[dto.ReplyItem], *errors.AppError) {
	keywords, appErr := tokenize(raw)
	if appErr != nil {
		return nil, appErr
	}
	rows, total := s.repo.SearchReplies(keywords, page, limit)

	items := make([]dto.ReplyItem, len(rows))
	for i, r := range rows {
		items[i] = dto.ReplyItem{
			ID: r.ID, TopicID: r.TopicID, TopicTitle: r.TopicTitle,
			Content: r.Content, Floor: r.Floor,
			User:    dto.UserBrief{ID: r.UserID, Name: r.UserName, Avatar: r.UserAvatar},
			Created: r.Created,
		}
	}
	return &dto.PaginatedResult[dto.ReplyItem]{Items: items, Total: total}, nil
}

// SearchGalgames returns galgame search results from the wiki Meilisearch
// index, enriched with local interaction counts. Uses the wiki `fields`
// parameter so the heavy `intro_*` markdown isn't sent over the wire on
// list pages — saves hundreds of KB per request.
func (s *SearchService) SearchGalgames(
	ctx context.Context,
	raw string,
	page, limit int,
	isSFW bool,
) (*dto.PaginatedResult[galgameDto.GalgameCard], *errors.AppError) {
	if _, appErr := tokenize(raw); appErr != nil {
		return nil, appErr
	}
	if s.wikiClient == nil || s.enricher == nil {
		return nil, errors.ErrInternal("Galgame 搜索未启用")
	}

	q := url.Values{}
	q.Set("q", raw)
	q.Set("page", strconv.Itoa(page))
	q.Set("limit", strconv.Itoa(limit))
	// Field projection: list view only needs the basics. Drop intro_* (1–10KB
	// each, 4 langs) + tag/official/engine name arrays. Saves ~95% bandwidth.
	q.Set("fields", "id,vndb_id,name_zh_cn,name_ja_jp,name_en_us,name_zh_tw,banner,banner_image_hash,content_limit,view,resource_update_time,user_id,original_language,age_limit")
	if isSFW {
		q.Set("content_limit", "sfw")
	}

	data, appErr := s.wikiClient.Get(ctx, "/galgame/search", q)
	if appErr != nil {
		return nil, appErr
	}

	var resp struct {
		Items []galgameDto.WikiGalgameItem `json:"items"`
		Total int64                        `json:"total"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, errors.ErrInternal(fmt.Sprintf("解析 Wiki 搜索响应失败: %v", err))
	}

	// Defensive: if wiki ignores the SFW filter we still strip NSFW here.
	filtered := s.enricher.FilterSFW(resp.Items, isSFW)
	cards := s.enricher.ToCards(filtered)

	return &dto.PaginatedResult[galgameDto.GalgameCard]{
		Items: cards,
		Total: resp.Total,
	}, nil
}

// SearchComments returns comment search results.
func (s *SearchService) SearchComments(raw string, page, limit int) (*dto.PaginatedResult[dto.CommentItem], *errors.AppError) {
	keywords, appErr := tokenize(raw)
	if appErr != nil {
		return nil, appErr
	}
	rows, total := s.repo.SearchComments(keywords, page, limit)

	items := make([]dto.CommentItem, len(rows))
	for i, r := range rows {
		items[i] = dto.CommentItem{
			ID: r.ID, TopicID: r.TopicID, TopicTitle: r.TopicTitle,
			Content: r.Content,
			User:    dto.UserBrief{ID: r.UserID, Name: r.UserName, Avatar: r.UserAvatar},
			Created: r.Created,
		}
	}
	return &dto.PaginatedResult[dto.CommentItem]{Items: items, Total: total}, nil
}
