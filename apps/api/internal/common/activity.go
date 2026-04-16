package common

import (
	"fmt"
	"sort"
	"time"

	"kun-galgame-api/pkg/response"
	"kun-galgame-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ActivityHandler struct {
	db *gorm.DB
}

func NewActivityHandler(db *gorm.DB) *ActivityHandler {
	return &ActivityHandler{db: db}
}

type activityItem struct {
	UniqueID  string    `json:"uniqueId"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Actor     struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Avatar string `json:"avatar"`
	} `json:"actor"`
	Link    string `json:"link"`
	Content string `json:"content"`
}

// activitySource defines a single SQL query that produces activity rows.
// The query MUST return columns: id, content, link, created, user_id.
// All column references must be fully qualified (t.xxx) to avoid
// ambiguity with the JOIN on the "user" table.
type activitySource struct {
	typeStr string
	query   string // SELECT ... FROM ... WHERE ... (no ORDER/LIMIT)
}

var sources = map[string]activitySource{
	"TOPIC_CREATION": {
		typeStr: "TOPIC_CREATION",
		query: `SELECT t.id, t.title AS content,
			'/topic/' || t.id AS link, t.created, t.user_id
			FROM topic t WHERE t.status != 1`,
	},
	"TOPIC_REPLY_CREATION": {
		typeStr: "TOPIC_REPLY_CREATION",
		query: `SELECT t.id, SUBSTRING(t.content, 1, 100) AS content,
			'/topic/' || t.topic_id AS link, t.created, t.user_id
			FROM topic_reply t`,
	},
	"TOPIC_COMMENT_CREATION": {
		typeStr: "TOPIC_COMMENT_CREATION",
		query: `SELECT t.id, SUBSTRING(t.content, 1, 100) AS content,
			'/topic/' || t.topic_id AS link, t.created, t.user_id
			FROM topic_comment t`,
	},
	"GALGAME_CREATION": {
		typeStr: "GALGAME_CREATION",
		query: `SELECT t.id,
			COALESCE(NULLIF(t.name_zh_cn,''), NULLIF(t.name_ja_jp,''),
				NULLIF(t.name_en_us,''), t.name_zh_tw) AS content,
			'/galgame/' || t.id AS link, t.created, t.user_id
			FROM galgame t WHERE t.status != 1`,
	},
	"GALGAME_COMMENT_CREATION": {
		typeStr: "GALGAME_COMMENT_CREATION",
		query: `SELECT t.id, SUBSTRING(t.content, 1, 100) AS content,
			'/galgame/' || t.galgame_id AS link, t.created, t.user_id
			FROM galgame_comment t`,
	},
	"GALGAME_RESOURCE_CREATION": {
		typeStr: "GALGAME_RESOURCE_CREATION",
		query: `SELECT t.id,
			COALESCE(NULLIF(t.note,''), t.type) AS content,
			'/galgame/' || t.galgame_id AS link, t.created, t.user_id
			FROM galgame_resource t`,
	},
	"GALGAME_RATING_CREATION": {
		typeStr: "GALGAME_RATING_CREATION",
		query: `SELECT t.id,
			SUBSTRING(COALESCE(t.short_summary,''), 1, 100) AS content,
			'/galgame/' || t.galgame_id AS link, t.created, t.user_id
			FROM galgame_rating t`,
	},
	"GALGAME_RATING_COMMENT_CREATION": {
		typeStr: "GALGAME_RATING_COMMENT_CREATION",
		query: `SELECT t.id, SUBSTRING(t.content, 1, 100) AS content,
			'/galgame/' || t.galgame_rating_id AS link, t.created, t.user_id
			FROM galgame_rating_comment t`,
	},
	"GALGAME_PR_CREATION": {
		typeStr: "GALGAME_PR_CREATION",
		query: `SELECT t.id,
			COALESCE(NULLIF(t.note,''), 'PR') AS content,
			'/galgame/' || t.galgame_id AS link, t.created, t.user_id
			FROM galgame_pr t`,
	},
	"GALGAME_WEBSITE_CREATION": {
		typeStr: "GALGAME_WEBSITE_CREATION",
		query: `SELECT t.id, t.name AS content,
			'/website/' || t.id AS link, t.created, t.user_id
			FROM galgame_website t`,
	},
	"GALGAME_WEBSITE_COMMENT_CREATION": {
		typeStr: "GALGAME_WEBSITE_COMMENT_CREATION",
		query: `SELECT t.id, SUBSTRING(t.content, 1, 100) AS content,
			'/website/' || t.website_id AS link, t.created, t.user_id
			FROM galgame_website_comment t`,
	},
	"TOOLSET_CREATION": {
		typeStr: "TOOLSET_CREATION",
		query: `SELECT t.id, t.name AS content,
			'/toolset/' || t.id AS link, t.created, t.user_id
			FROM galgame_toolset t WHERE t.status != 1`,
	},
	"TOOLSET_RESOURCE_CREATION": {
		typeStr: "TOOLSET_RESOURCE_CREATION",
		query: `SELECT t.id,
			COALESCE(NULLIF(t.note,''), t.content) AS content,
			'/toolset/' || t.toolset_id AS link, t.created, t.user_id
			FROM galgame_toolset_resource t`,
	},
	"TOOLSET_COMMENT_CREATION": {
		typeStr: "TOOLSET_COMMENT_CREATION",
		query: `SELECT t.id, SUBSTRING(t.content, 1, 100) AS content,
			'/toolset/' || t.toolset_id AS link, t.created, t.user_id
			FROM galgame_toolset_comment t`,
	},
	"TODO_CREATION": {
		typeStr: "TODO_CREATION",
		query: `SELECT t.id, t.content_zh_cn AS content,
			'/update' AS link, t.created, t.user_id
			FROM todo t`,
	},
	"UPDATE_LOG_CREATION": {
		typeStr: "UPDATE_LOG_CREATION",
		query: `SELECT t.id, t.content_zh_cn AS content,
			'/update' AS link, t.created, t.user_id
			FROM update_log t`,
	},
	"MESSAGE_UPVOTE": {
		typeStr: "MESSAGE_UPVOTE",
		query: `SELECT t.id, t.content,
			t.link, t.created, t.sender_id AS user_id
			FROM message t WHERE t.type = 'upvoted'`,
	},
	"MESSAGE_SOLUTION": {
		typeStr: "MESSAGE_SOLUTION",
		query: `SELECT t.id, t.content,
			t.link, t.created, t.sender_id AS user_id
			FROM message t WHERE t.type = 'solution'`,
	},
}

// GetActivity returns activity feed filtered by type.
// GET /api/activity
func (h *ActivityHandler) GetActivity(c *fiber.Ctx) error {
	var req struct {
		Page  int    `query:"page" validate:"min=1"`
		Limit int    `query:"limit" validate:"min=1,max=50"`
		Type  string `query:"type" validate:"required"`
	}
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	if req.Type == "all" {
		return h.getTimeline(c, req.Page, req.Limit)
	}

	src, ok := sources[req.Type]
	if !ok {
		return response.Paginated(c, []activityItem{}, 0)
	}

	items, total := h.fetch(src, req.Page, req.Limit)
	return response.Paginated(c, items, total)
}

// GetTimeline returns mixed activity timeline.
// GET /api/activity/timeline
func (h *ActivityHandler) GetTimeline(c *fiber.Ctx) error {
	var req struct {
		Page  int `query:"page" validate:"min=1"`
		Limit int `query:"limit" validate:"min=1,max=50"`
	}
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	return h.getTimeline(c, req.Page, req.Limit)
}

func (h *ActivityHandler) getTimeline(
	c *fiber.Ctx, page, limit int,
) error {
	var all []activityItem
	for _, src := range sources {
		items, _ := h.fetch(src, 1, 50)
		all = append(all, items...)
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].Timestamp.After(all[j].Timestamp)
	})

	total := int64(len(all))
	start := (page - 1) * limit
	if start >= len(all) {
		return response.Paginated(c, []activityItem{}, total)
	}
	end := min(start+limit, len(all))
	return response.Paginated(c, all[start:end], total)
}

// fetch wraps the source query with user JOIN, pagination, and
// count. The inner query (src.query) selects: id, content, link,
// created, user_id — all fully qualified to avoid column ambiguity.
func (h *ActivityHandler) fetch(
	src activitySource, page, limit int,
) ([]activityItem, int64) {
	type row struct {
		ID        int       `gorm:"column:id"`
		Content   string    `gorm:"column:content"`
		Link      string    `gorm:"column:link"`
		Created   time.Time `gorm:"column:created"`
		UserID    int       `gorm:"column:user_id"`
		UserName  string    `gorm:"column:user_name"`
		Avatar    string    `gorm:"column:avatar"`
	}

	countSQL := fmt.Sprintf(
		`SELECT COUNT(*) FROM (%s) AS sub`, src.query,
	)
	var total int64
	h.db.Raw(countSQL).Scan(&total)

	dataSQL := fmt.Sprintf(
		`SELECT sub.*, u.name AS user_name, u.avatar
		FROM (%s) AS sub
		LEFT JOIN "user" u ON u.id = sub.user_id
		ORDER BY sub.created DESC
		LIMIT %d OFFSET %d`,
		src.query, limit, (page-1)*limit,
	)
	var rows []row
	h.db.Raw(dataSQL).Scan(&rows)

	items := make([]activityItem, len(rows))
	for i, r := range rows {
		items[i] = activityItem{
			UniqueID:  fmt.Sprintf("%s-%d", src.typeStr, r.ID),
			Type:      src.typeStr,
			Content:   r.Content,
			Link:      r.Link,
			Timestamp: r.Created,
		}
		items[i].Actor.ID = r.UserID
		items[i].Actor.Name = r.UserName
		items[i].Actor.Avatar = r.Avatar
	}
	return items, total
}
