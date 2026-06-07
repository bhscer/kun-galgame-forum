package dto

import "time"

// ──────────────────────────────────────────
// Requests
// ──────────────────────────────────────────

type ActivityRequest struct {
	Page  int    `query:"page" validate:"min=1"`
	Limit int    `query:"limit" validate:"min=1,max=50"`
	Type  string `query:"type" validate:"required"`
	// ShowNoResource mirrors the user's 显示设置 → 显示没有下载资源的 Galgame
	// preference. Default false (omitted) hides resource-less galgames, so their
	// GALGAME_CREATION activity is dropped from the feed too.
	ShowNoResource bool `query:"showNoResource"`
}

type TimelineRequest struct {
	Page           int  `query:"page" validate:"min=1"`
	Limit          int  `query:"limit" validate:"min=1,max=50"`
	ShowNoResource bool `query:"showNoResource"`
}

// ──────────────────────────────────────────
// Responses
// ──────────────────────────────────────────

type Actor struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type ActivityItem struct {
	UniqueID  string    `json:"uniqueId"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Actor     Actor     `json:"actor"`
	Link      string    `json:"link"`
	Content   string    `json:"content"`
}
