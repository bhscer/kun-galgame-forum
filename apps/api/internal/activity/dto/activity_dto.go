package dto

import "time"

// ──────────────────────────────────────────
// Requests
// ──────────────────────────────────────────

type ActivityRequest struct {
	// Cursor is the opaque keyset position from the previous page's nextCursor;
	// empty = first page. Replaces the old `page` — offset paging duplicated /
	// skipped rows across pages (see repository.FetchKeyset).
	Cursor string `query:"cursor"`
	Limit  int    `query:"limit" validate:"min=1,max=50"`
	Type   string `query:"type" validate:"required"`
	// ShowNoResource mirrors the user's 显示设置 → 显示没有下载资源的 Galgame
	// preference. Default false (omitted) hides resource-less galgames, so their
	// GALGAME_CREATION activity is dropped from the feed too.
	ShowNoResource bool `query:"showNoResource"`
}

type TimelineRequest struct {
	Cursor         string `query:"cursor"`
	Limit          int    `query:"limit" validate:"min=1,max=50"`
	ShowNoResource bool   `query:"showNoResource"`
}

// TabRequest drives the home-page feed's five tab buckets. Tab is one of
// all/topic/galgame/resource/others (see service.homeTabTypes); "all" is every
// non-resource type.
type TabRequest struct {
	Tab            string `query:"tab" validate:"required,oneof=all topic galgame resource others"`
	Cursor         string `query:"cursor"`
	Limit          int    `query:"limit" validate:"min=1,max=50"`
	ShowNoResource bool   `query:"showNoResource"`
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
	// ID is the source row's id (galgame id for galgame-scoped rows). Kept
	// internal (json:"-") — the service uses (Timestamp, Type, ID) to build the
	// keyset nextCursor; clients consume nextCursor, never this.
	ID        int       `json:"-"`
	UniqueID  string    `json:"uniqueId"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Actor     Actor     `json:"actor"`
	Link      string    `json:"link"`
	Content   string    `json:"content"`
}
