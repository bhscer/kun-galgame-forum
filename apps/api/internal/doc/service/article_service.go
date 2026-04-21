package service

import (
	"time"

	"kun-galgame-api/internal/doc/dto"
	"kun-galgame-api/internal/doc/model"
	"kun-galgame-api/internal/doc/repository"
	"kun-galgame-api/internal/infrastructure/markdown"
	"kun-galgame-api/pkg/errors"

	"gorm.io/gorm"
)

type ArticleService struct {
	articleRepo  *repository.ArticleRepository
	categoryRepo *repository.CategoryRepository
}

func NewArticleService(
	articleRepo *repository.ArticleRepository,
	categoryRepo *repository.CategoryRepository,
) *ArticleService {
	return &ArticleService{articleRepo: articleRepo, categoryRepo: categoryRepo}
}

// ──────────────────────────────────────────
// GetList — GET /doc/article
// ──────────────────────────────────────────

// ArticleListResult carries the list + total for paginated handler responses.
type ArticleListResult struct {
	Items []dto.ArticleSummary
	Total int64
}

func (s *ArticleService) GetList(req *dto.GetArticlesRequest) *ArticleListResult {
	if req.OrderBy == "" {
		req.OrderBy = "published_time"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	items, total := s.articleRepo.FindPaginated(req)

	categoryByID := s.loadCategoriesFor(items)
	summaries := make([]dto.ArticleSummary, 0, len(items))
	for _, a := range items {
		summaries = append(summaries, toArticleSummary(a, categoryByID[a.CategoryID]))
	}
	return &ArticleListResult{Items: summaries, Total: total}
}

func (s *ArticleService) loadCategoriesFor(articles []model.DocArticle) map[int]dto.ArticleCategoryBrief {
	out := map[int]dto.ArticleCategoryBrief{}
	if len(articles) == 0 {
		return out
	}
	idSet := map[int]struct{}{}
	ids := make([]int, 0, len(articles))
	for _, a := range articles {
		if _, ok := idSet[a.CategoryID]; ok {
			continue
		}
		idSet[a.CategoryID] = struct{}{}
		ids = append(ids, a.CategoryID)
	}
	var cats []model.DocCategory
	s.categoryRepo.DB().Where("id IN ?", ids).Find(&cats)
	for _, c := range cats {
		out[c.ID] = dto.ArticleCategoryBrief{ID: c.ID, Slug: c.Slug, Title: c.Title}
	}
	return out
}

func toArticleSummary(a model.DocArticle, cat dto.ArticleCategoryBrief) dto.ArticleSummary {
	if cat.ID == 0 {
		cat = dto.ArticleCategoryBrief{ID: a.CategoryID}
	}
	return dto.ArticleSummary{
		ID:            a.ID,
		Title:         a.Title,
		Slug:          a.Slug,
		Path:          a.Path,
		Description:   a.Description,
		Banner:        a.Banner,
		Status:        a.Status,
		IsPin:         a.IsPin,
		View:          a.View,
		PublishedTime: a.PublishedTime,
		EditedTime:    a.EditedTime,
		CategoryID:    a.CategoryID,
		AuthorID:      a.AuthorID,
		Category:      cat,
		Created:       a.CreatedAt,
		Updated:       a.UpdatedAt,
	}
}

// ──────────────────────────────────────────
// GetBySlug — GET /doc/article/:slug
// ──────────────────────────────────────────

func (s *ArticleService) GetBySlug(slug string) (*dto.ArticleDetailResponse, *errors.AppError) {
	article, err := s.articleRepo.FindBySlug(slug)
	if err != nil {
		return nil, errors.ErrNotFound("未找到该文章")
	}

	// Bump view asynchronously to preserve the old fire-and-forget behavior.
	go s.articleRepo.IncrementView(article.ID)

	html, toc := markdown.RenderWithTOC(article.ContentMarkdown)

	var cat dto.ArticleCategoryBrief
	var category model.DocCategory
	if err := s.categoryRepo.DB().
		Where("id = ?", article.CategoryID).First(&category).Error; err == nil {
		cat = dto.ArticleCategoryBrief{ID: category.ID, Slug: category.Slug, Title: category.Title}
	} else {
		cat = dto.ArticleCategoryBrief{ID: article.CategoryID}
	}

	return &dto.ArticleDetailResponse{
		ID:              article.ID,
		Title:           article.Title,
		Slug:            article.Slug,
		Path:            article.Path,
		Description:     article.Description,
		Banner:          article.Banner,
		Status:          article.Status,
		IsPin:           article.IsPin,
		View:            article.View,
		PublishedTime:   article.PublishedTime,
		EditedTime:      article.EditedTime,
		ContentMarkdown: article.ContentMarkdown,
		ContentHTML:     html,
		Toc:             toc,
		CategoryID:      article.CategoryID,
		AuthorID:        article.AuthorID,
		Category:        cat,
		Created:         article.CreatedAt,
		Updated:         article.UpdatedAt,
	}, nil
}

// ──────────────────────────────────────────
// Create — POST /doc/article
// ──────────────────────────────────────────

func (s *ArticleService) Create(uid int, req *dto.CreateArticleRequest) (*model.DocArticle, *errors.AppError) {
	now := time.Now()
	article := model.DocArticle{
		Title:           req.Title,
		Slug:            req.Slug,
		Path:            "/doc/" + req.Slug,
		Description:     req.Description,
		Banner:          req.Banner,
		Status:          req.Status,
		IsPin:           req.IsPin,
		ContentMarkdown: req.ContentMarkdown,
		CategoryID:      req.CategoryID,
		AuthorID:        uid,
		PublishedTime:   now,
	}

	txErr := s.articleRepo.DB().Transaction(func(tx *gorm.DB) error {
		if err := s.articleRepo.Create(tx, &article); err != nil {
			return err
		}
		s.articleRepo.InsertTagRelations(tx, article.ID, req.TagIDs)
		return nil
	})
	if txErr != nil {
		return nil, errors.ErrInternal("创建文章失败")
	}

	return &article, nil
}

// ──────────────────────────────────────────
// Update — PUT /doc/article
// ──────────────────────────────────────────

func (s *ArticleService) Update(req *dto.UpdateArticleRequest) *errors.AppError {
	now := time.Now()
	updates := map[string]any{
		"title":            req.Title,
		"slug":             req.Slug,
		"path":             "/doc/" + req.Slug,
		"description":      req.Description,
		"banner":           req.Banner,
		"status":           req.Status,
		"is_pin":           req.IsPin,
		"content_markdown": req.ContentMarkdown,
		"category_id":      req.CategoryID,
		"edited_time":      &now,
	}

	txErr := s.articleRepo.DB().Transaction(func(tx *gorm.DB) error {
		if err := s.articleRepo.UpdateFields(tx, req.ArticleID, updates); err != nil {
			return err
		}
		s.articleRepo.ReplaceTagRelations(tx, req.ArticleID, req.TagIDs)
		return nil
	})
	if txErr != nil {
		return errors.ErrInternal("更新文章失败")
	}

	return nil
}

// ──────────────────────────────────────────
// Delete — DELETE /doc/article
// ──────────────────────────────────────────

func (s *ArticleService) Delete(articleID int) *errors.AppError {
	s.articleRepo.DeleteTagRelationsByArticleID(articleID)
	s.articleRepo.DeleteByID(articleID)
	return nil
}
