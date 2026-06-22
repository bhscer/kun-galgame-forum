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
	// Data is the per-type rich card payload (Activity Streams "object"),
	// discriminated by Type and populated during enrichment. nil for types that
	// have no rich card yet — the FE renders those with the generic card. Kept
	// `any` so each type carries only its own shape (see TopicActivityData).
	Data any `json:"data,omitempty"`
}

// TopicActivityData is the rich-card payload for TOPIC_CREATION: everything the
// feed's topic card shows beyond the envelope (the title lives in Content).
// Covers are /image/<hash> tokens (card shows the first few); Sections render in
// the stat row; the badge flags feed the shared TopicTagGroup; TopReply is the
// most-liked reply (omitted when none).
type TopicActivityData struct {
	Excerpt       string     `json:"excerpt"`
	Sections      []string   `json:"sections"`
	CoverImages   []string   `json:"coverImages"`
	View          int        `json:"view"`
	LikeCount     int        `json:"likeCount"`
	ReplyCount    int        `json:"replyCount"`
	CommentCount  int        `json:"commentCount"`
	UpvoteTime    *time.Time `json:"upvoteTime"`
	HasBestAnswer bool       `json:"hasBestAnswer"`
	IsPoll        bool       `json:"isPoll"`
	IsNSFW        bool       `json:"isNSFW"`
	TopReply      *TopReply  `json:"topReply,omitempty"`
}

// TopReply is a topic's most-liked reply (a short excerpt + its like count),
// shown on the feed's topic card. Only populated when a reply has >0 likes.
type TopReply struct {
	Content   string `json:"content"`
	LikeCount int    `json:"likeCount"`
}
