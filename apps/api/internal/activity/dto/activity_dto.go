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

// ReplyActivityData is the rich-card payload for TOPIC_REPLY_CREATION: the title
// of the topic the reply belongs to (shown at the bottom of the card) and, if the
// reply quoted another reply, that quoted reply. The reply body itself is
// ActivityItem.Content (with @/# tokens already resolved to readable text).
type ReplyActivityData struct {
	TopicTitle  string       `json:"topicTitle"`
	QuotedReply *QuotedReply `json:"quotedReply,omitempty"`
}

// QuotedReply is the reply this reply quoted (#floor → its body), shown as a
// nested block in the middle of the reply card. Content has its @/# tokens
// already resolved to readable text.
type QuotedReply struct {
	Floor   int    `json:"floor"`
	Content string `json:"content"`
}

// GalgameActivityData is the rich-card payload for galgame-scoped activity
// (creation / edit / PR / comment / rating / resource): the galgame's name +
// cover + a little metadata, all pulled from the wiki brief already fetched
// during enrichment (no extra query). CoverHash resolves to a CDN URL on the FE
// (imageTokenUrl). Release date is nullable (TBA / unknown).
type GalgameActivityData struct {
	Name        string  `json:"name"`
	CoverHash   string  `json:"coverHash"`
	Language    string  `json:"language"`
	AgeLimit    string  `json:"ageLimit"`
	ReleaseDate *string `json:"releaseDate"`
	// GalgameID lets the FE link / like / favorite without parsing the link.
	GalgameID int `json:"galgameId,omitempty"`
	// RevisionID is the wiki revision ROW id for GALGAME_EDIT — the legacy input
	// for the id→number diff resolution. RevisionNumber is the per-galgame
	// revision number the diff endpoint's :rev keys on; the card uses it directly
	// when present (>0) and falls back to RevisionID otherwise.
	RevisionID     int `json:"revisionId,omitempty"`
	RevisionNumber int `json:"revisionNumber,omitempty"`
	// Developer (制作会社) + Intro are set for the GALGAME_CREATION and GALGAME_EDIT
	// cards (which share the info area), from the wiki detail brief. Developer =
	// officials joined with 、; Intro is the preferred-language introduction,
	// truncated for a 3-line preview.
	Developer string `json:"developer,omitempty"`
	Intro     string `json:"intro,omitempty"`
	// Local rollups, only set for the GALGAME_CREATION rich card (omitempty → the
	// FE defaults each to 0). These are global counts (cache-safe); the viewer's
	// own liked/favorited state is NOT here (would break the shared feed cache).
	ResourceCount int `json:"resourceCount,omitempty"`
	LikeCount     int `json:"likeCount,omitempty"`
	FavoriteCount int `json:"favoriteCount,omitempty"`
	// Rating is only set for GALGAME_RATING_CREATION (the rating card).
	Rating *RatingInfo `json:"rating,omitempty"`
}

// RatingInfo is the GALGAME_RATING_CREATION rich-card payload — the galgame name
// + cover come from the surrounding GalgameActivityData. ShortSummary is blanked
// when SpoilerLevel != "none" so spoilers never cross the boundary.
type RatingInfo struct {
	RatingID     int    `json:"ratingId"`
	Overall      int    `json:"overall"`
	PlayStatus   string `json:"playStatus"`
	Recommend    string `json:"recommend"`
	ShortSummary string `json:"shortSummary"`
	SpoilerLevel string `json:"spoilerLevel"`
	LikeCount    int    `json:"likeCount"`
	AuthorID     int    `json:"authorId"`
}
