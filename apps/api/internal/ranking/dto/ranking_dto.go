package dto

// ──────────────────────────────────────────
// Requests
// ──────────────────────────────────────────

type GalgameRankingRequest struct {
	Page      int    `query:"page" validate:"min=1"`
	Limit     int    `query:"limit" validate:"min=1,max=50"`
	SortField string `query:"sortField" validate:"required,oneof=view like_count favorite_count resource_count"`
	SortOrder string `query:"sortOrder" validate:"required,oneof=asc desc"`
}

type TopicRankingRequest struct {
	Page      int    `query:"page" validate:"min=1"`
	Limit     int    `query:"limit" validate:"min=1,max=50"`
	SortField string `query:"sortField" validate:"required,oneof=view upvote_count like_count reply_count comment_count favorite_count"`
	SortOrder string `query:"sortOrder" validate:"required,oneof=asc desc"`
}

type UserRankingRequest struct {
	Page      int    `query:"page" validate:"min=1"`
	Limit     int    `query:"limit" validate:"min=1,max=50"`
	SortField string `query:"sortField" validate:"required,oneof=moemoepoint"`
	SortOrder string `query:"sortOrder" validate:"required,oneof=asc desc"`
}

// ──────────────────────────────────────────
// Responses
// ──────────────────────────────────────────

type UserBrief struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type LocaleName struct {
	EnUS string `json:"en-us"`
	JaJP string `json:"ja-jp"`
	ZhCN string `json:"zh-cn"`
	ZhTW string `json:"zh-tw"`
}

type GalgameRankingItem struct {
	ID        int        `json:"id"`
	Name      LocaleName `json:"name"`
	User      UserBrief  `json:"user"`
	Banner    string     `json:"banner"`
	Value     int        `json:"value"`
	SortField string     `json:"sortField"`
	// U2: derived banner so FE `getEffectiveBanner` can pick `_mini` for
	// covers-only galgames; without these the ranking card falls back to
	// the legacy `banner` URL which is empty for post-PR5 entries.
	// Mirrors HomeGalgame / GalgameListCard.
	EffectiveBannerHash string `json:"effective_banner_hash,omitempty"`
	EffectiveBannerURL  string `json:"effective_banner_url,omitempty"`
}

type TopicRankingItem struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	User      UserBrief `json:"user"`
	Value     int       `json:"value"`
	SortField string    `json:"sortField"`
}

type UserRankingItem struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Bio    string `json:"bio"`
	Value  int    `json:"value"`
	// SortField must travel back so the FE can render the right
	// sort-icon next to the value. Galgame/Topic already set this; user
	// was missing it (silent FE icon dropout).
	SortField string `json:"sortField"`
}
