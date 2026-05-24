package dto

// KunLanguage is the four-language locale map shared across galgame-related
// responses: { en-us, ja-jp, zh-cn, zh-tw }.
type KunLanguage struct {
	EnUs string `json:"en-us"`
	JaJp string `json:"ja-jp"`
	ZhCn string `json:"zh-cn"`
	ZhTw string `json:"zh-tw"`
}

// UserBrief is the minimal user shape embedded in list responses.
type UserBrief struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

// UserGalgameCard is an item in GET /api/user/:userID/galgames.
type UserGalgameCard struct {
	ID                 int         `json:"id"`
	Name               KunLanguage `json:"name"`
	Banner             string      `json:"banner"`
	User               UserBrief   `json:"user"`
	ContentLimit       string      `json:"contentLimit"`
	View               int         `json:"view"`
	LikeCount          int         `json:"likeCount"`
	ResourceUpdateTime string      `json:"resourceUpdateTime"`
	Platform           []string    `json:"platform"`
	Language           []string    `json:"language"`
}

// UserGalgameComment is an item in GET /api/user/:userID/galgame-comments —
// the three "评论 / 被评论 / 点赞评论" sub-tabs on /user/:id/galgame/.
// Mirrors the topic comment card shape: enough info to render content
// + click-through to the parent galgame, no more.
type UserGalgameComment struct {
	ID          int       `json:"id"`
	GalgameID   int       `json:"galgameId"`
	Content     string    `json:"content"`
	ContentHtml string    `json:"contentHtml"`
	User        UserBrief `json:"user"`
	Created     string    `json:"created"`
}

// UserGalgameCommentsRequest is the query for GET /api/user/:userID/galgame-comments.
type UserGalgameCommentsRequest struct {
	Type  string `query:"type" validate:"required"`
	Page  int    `query:"page" validate:"min=1"`
	Limit int    `query:"limit" validate:"min=1,max=50"`
}
