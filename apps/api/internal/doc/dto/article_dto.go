package dto

import (
	"time"

	"kun-galgame-api/internal/infrastructure/markdown"
)

// ──────────────────────────────────────────
// Requests
// ──────────────────────────────────────────

// GetArticlesRequest is the query for GET /doc/article.
type GetArticlesRequest struct {
	Page       int    `query:"page" validate:"min=1"`
	Limit      int    `query:"limit" validate:"min=1,max=100"`
	CategoryID *int   `query:"categoryId"`
	TagID      *int   `query:"tagId"`
	Status     *int   `query:"status"`
	IsPin      *bool  `query:"isPin"`
	Keyword    string `query:"keyword"`
	OrderBy    string `query:"orderBy" validate:"omitempty,oneof=published_time created view updated"`
	SortOrder  string `query:"sortOrder" validate:"omitempty,oneof=asc desc"`
}

// CreateArticleRequest is the payload for POST /doc/article.
type CreateArticleRequest struct {
	Title           string `json:"title" validate:"required,max=233"`
	Slug            string `json:"slug" validate:"required,max=233"`
	Description     string `json:"description" validate:"max=1000"`
	Banner          string `json:"banner" validate:"max=500"`
	Status          int    `json:"status" validate:"oneof=0 1 2"`
	IsPin           bool   `json:"isPin"`
	ContentMarkdown string `json:"contentMarkdown" validate:"required"`
	CategoryID      int    `json:"categoryId" validate:"required,min=1"`
	TagIDs          []int  `json:"tagIds"`
}

// UpdateArticleRequest is the payload for PUT /doc/article.
type UpdateArticleRequest struct {
	ArticleID       int    `json:"articleId" validate:"required,min=1"`
	Title           string `json:"title" validate:"required,max=233"`
	Slug            string `json:"slug" validate:"required,max=233"`
	Description     string `json:"description" validate:"max=1000"`
	Banner          string `json:"banner" validate:"max=500"`
	Status          int    `json:"status" validate:"oneof=0 1 2"`
	IsPin           bool   `json:"isPin"`
	ContentMarkdown string `json:"contentMarkdown" validate:"required"`
	CategoryID      int    `json:"categoryId" validate:"required,min=1"`
	TagIDs          []int  `json:"tagIds"`
}

// DeleteArticleRequest is the query for DELETE /doc/article.
type DeleteArticleRequest struct {
	ArticleID int `query:"articleId" validate:"required,min=1"`
}

// ──────────────────────────────────────────
// Responses
// ──────────────────────────────────────────

// ArticleCategoryBrief is the nested category shape embedded in list/detail
// responses so the frontend can render category pills without a separate
// fetch. Matches the legacy Nitro relation output.
type ArticleCategoryBrief struct {
	ID    int    `json:"id"`
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

// ArticleSummary is the list-row shape of GET /doc/article. Mirrors
// model.DocArticle's JSON tags but adds the `category` relation.
type ArticleSummary struct {
	ID            int                  `json:"id"`
	Title         string               `json:"title"`
	Slug          string               `json:"slug"`
	Path          string               `json:"path"`
	Description   string               `json:"description"`
	Banner        string               `json:"banner"`
	Status        int                  `json:"status"`
	IsPin         bool                 `json:"is_pin"`
	View          int                  `json:"view"`
	PublishedTime time.Time            `json:"published_time"`
	EditedTime    *time.Time           `json:"edited_time"`
	CategoryID    int                  `json:"category_id"`
	AuthorID      int                  `json:"author_id"`
	Category      ArticleCategoryBrief `json:"category"`
	Created       time.Time            `json:"created"`
	Updated       time.Time            `json:"updated"`
}

// ArticleDetailResponse is the shape of GET /doc/article/:slug.
// Field names mirror the pre-refactor handler output exactly (mixing snake_case
// from the GORM model with one camelCase field for the rendered HTML).
type ArticleDetailResponse struct {
	ID              int        `json:"id"`
	Title           string     `json:"title"`
	Slug            string     `json:"slug"`
	Path            string     `json:"path"`
	Description     string     `json:"description"`
	Banner          string     `json:"banner"`
	Status          int        `json:"status"`
	IsPin           bool       `json:"is_pin"`
	View            int        `json:"view"`
	PublishedTime   time.Time  `json:"published_time"`
	EditedTime      *time.Time `json:"edited_time"`
	ContentMarkdown string               `json:"content_markdown"`
	ContentHTML     string               `json:"contentHtml"`
	Toc             []markdown.TocLink   `json:"toc"`
	CategoryID      int                  `json:"category_id"`
	AuthorID        int                  `json:"author_id"`
	Category        ArticleCategoryBrief `json:"category"`
	Created         time.Time            `json:"created"`
	Updated         time.Time            `json:"updated"`
}
