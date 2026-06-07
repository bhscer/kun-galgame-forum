package repository

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ActivityRepository struct {
	db *gorm.DB
}

func NewActivityRepository(db *gorm.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

// ──────────────────────────────────────────
// Source definitions
// ──────────────────────────────────────────

// ActivitySource defines a single SQL sub-query that produces activity rows.
// The query MUST SELECT these exact columns:
//
//	type_str, id, content, link, created, user_id, galgame_id
//
// `galgame_id` is 0 for activities that are not galgame-scoped. All column
// references should be fully qualified (t.xxx) to keep the UNION ALL query
// unambiguous.
type ActivitySource struct {
	TypeStr string
	Query   string
}

// Sources is the map of activity type → SQL source.
var Sources = map[string]ActivitySource{
	"TOPIC_CREATION": {
		TypeStr: "TOPIC_CREATION",
		Query: `SELECT 'TOPIC_CREATION' AS type_str, t.id,
			t.title AS content,
			'/topic/' || t.id AS link, t.created, t.user_id, 0 AS galgame_id
			FROM topic t WHERE t.status != 1`,
	},
	"TOPIC_REPLY_CREATION": {
		TypeStr: "TOPIC_REPLY_CREATION",
		// A reply can target multiple other replies. When it does, the
		// user's text is stored per-target in topic_reply_target.content
		// and topic_reply.content itself is empty. Concatenate both so
		// multi-target replies still show meaningful text.
		Query: `SELECT 'TOPIC_REPLY_CREATION' AS type_str, t.id,
			SUBSTRING(
				COALESCE(t.content, '') ||
				COALESCE(
					(SELECT STRING_AGG(trt.content, ' ' ORDER BY trt.id)
					 FROM topic_reply_target trt
					 WHERE trt.reply_id = t.id),
					''
				),
				1, 100
			) AS content,
			'/topic/' || t.topic_id AS link, t.created, t.user_id, 0 AS galgame_id
			FROM topic_reply t`,
	},
	"TOPIC_COMMENT_CREATION": {
		TypeStr: "TOPIC_COMMENT_CREATION",
		Query: `SELECT 'TOPIC_COMMENT_CREATION' AS type_str, t.id,
			SUBSTRING(t.content, 1, 100) AS content,
			'/topic/' || t.topic_id AS link, t.created, t.user_id, 0 AS galgame_id
			FROM topic_comment t`,
	},
	"GALGAME_CREATION": {
		TypeStr: "GALGAME_CREATION",
		// Local galgame table has no user_id (moved to wiki); actor is
		// filled in from the wiki brief during enrichment.
		Query: `SELECT 'GALGAME_CREATION' AS type_str, t.id,
			'' AS content,
			'/galgame/' || t.id AS link, t.created,
			0 AS user_id, t.id AS galgame_id
			FROM galgame t`,
	},
	"GALGAME_COMMENT_CREATION": {
		TypeStr: "GALGAME_COMMENT_CREATION",
		Query: `SELECT 'GALGAME_COMMENT_CREATION' AS type_str, t.id,
			SUBSTRING(t.content, 1, 100) AS content,
			'/galgame/' || t.galgame_id AS link, t.created, t.user_id, t.galgame_id
			FROM galgame_comment t`,
	},
	"GALGAME_RESOURCE_CREATION": {
		TypeStr: "GALGAME_RESOURCE_CREATION",
		Query: `SELECT 'GALGAME_RESOURCE_CREATION' AS type_str, t.id,
			'' AS content,
			'/galgame/' || t.galgame_id AS link, t.created, t.user_id, t.galgame_id
			FROM galgame_resource t`,
	},
	// GALGAME_EDIT: wiki merged-revision (edit) events, mirrored into
	// galgame_activity by the wiki-revision sync cron. user_id is the real
	// editor (from the revision), so unlike GALGAME_CREATION it needs no
	// brief-actor injection during enrichment.
	"GALGAME_EDIT": {
		TypeStr: "GALGAME_EDIT",
		Query: `SELECT 'GALGAME_EDIT' AS type_str, t.id,
			'' AS content,
			'/galgame/' || t.galgame_id AS link, t.created, t.user_id, t.galgame_id
			FROM galgame_activity t`,
	},
	"GALGAME_RATING_CREATION": {
		TypeStr: "GALGAME_RATING_CREATION",
		Query: `SELECT 'GALGAME_RATING_CREATION' AS type_str, t.id,
			SUBSTRING(COALESCE(t.short_summary,''), 1, 100) AS content,
			'/galgame/' || t.galgame_id AS link, t.created, t.user_id, t.galgame_id
			FROM galgame_rating t`,
	},
	"GALGAME_RATING_COMMENT_CREATION": {
		TypeStr: "GALGAME_RATING_COMMENT_CREATION",
		Query: `SELECT 'GALGAME_RATING_COMMENT_CREATION' AS type_str, t.id,
			SUBSTRING(t.content, 1, 100) AS content,
			'/galgame/' || t.galgame_rating_id AS link, t.created, t.user_id, 0 AS galgame_id
			FROM galgame_rating_comment t`,
	},
	// GALGAME_PR_CREATION removed: galgame_pr table moved to wiki service
	"GALGAME_WEBSITE_CREATION": {
		TypeStr: "GALGAME_WEBSITE_CREATION",
		Query: `SELECT 'GALGAME_WEBSITE_CREATION' AS type_str, t.id,
			t.name AS content,
			'/website/' || t.id AS link, t.created, t.user_id, 0 AS galgame_id
			FROM galgame_website t`,
	},
	"GALGAME_WEBSITE_COMMENT_CREATION": {
		TypeStr: "GALGAME_WEBSITE_COMMENT_CREATION",
		Query: `SELECT 'GALGAME_WEBSITE_COMMENT_CREATION' AS type_str, t.id,
			SUBSTRING(t.content, 1, 100) AS content,
			'/website/' || t.website_id AS link, t.created, t.user_id, 0 AS galgame_id
			FROM galgame_website_comment t`,
	},
	"TOOLSET_CREATION": {
		TypeStr: "TOOLSET_CREATION",
		Query: `SELECT 'TOOLSET_CREATION' AS type_str, t.id,
			t.name AS content,
			'/toolset/' || t.id AS link, t.created, t.user_id, 0 AS galgame_id
			FROM galgame_toolset t WHERE t.status != 1`,
	},
	"TOOLSET_RESOURCE_CREATION": {
		TypeStr: "TOOLSET_RESOURCE_CREATION",
		Query: `SELECT 'TOOLSET_RESOURCE_CREATION' AS type_str, t.id,
			COALESCE(NULLIF(t.note,''), t.content) AS content,
			'/toolset/' || t.toolset_id AS link, t.created, t.user_id, 0 AS galgame_id
			FROM galgame_toolset_resource t`,
	},
	"TOOLSET_COMMENT_CREATION": {
		TypeStr: "TOOLSET_COMMENT_CREATION",
		Query: `SELECT 'TOOLSET_COMMENT_CREATION' AS type_str, t.id,
			SUBSTRING(t.content, 1, 100) AS content,
			'/toolset/' || t.toolset_id AS link, t.created, t.user_id, 0 AS galgame_id
			FROM galgame_toolset_comment t`,
	},
	"TODO_CREATION": {
		TypeStr: "TODO_CREATION",
		Query: `SELECT 'TODO_CREATION' AS type_str, t.id,
			t.content_zh_cn AS content,
			'/update' AS link, t.created, t.user_id, 0 AS galgame_id
			FROM todo t`,
	},
	"UPDATE_LOG_CREATION": {
		TypeStr: "UPDATE_LOG_CREATION",
		Query: `SELECT 'UPDATE_LOG_CREATION' AS type_str, t.id,
			t.content_zh_cn AS content,
			'/update' AS link, t.created, t.user_id, 0 AS galgame_id
			FROM update_log t`,
	},
	"MESSAGE_UPVOTE": {
		TypeStr: "MESSAGE_UPVOTE",
		Query: `SELECT 'MESSAGE_UPVOTE' AS type_str, t.id, t.content,
			t.link, t.created, t.sender_id AS user_id, 0 AS galgame_id
			FROM message t WHERE t.type = 'upvoted'`,
	},
	"MESSAGE_SOLUTION": {
		TypeStr: "MESSAGE_SOLUTION",
		Query: `SELECT 'MESSAGE_SOLUTION' AS type_str, t.id, t.content,
			t.link, t.created, t.sender_id AS user_id, 0 AS galgame_id
			FROM message t WHERE t.type = 'solution'`,
	},
}

// ──────────────────────────────────────────
// Row projection
// ──────────────────────────────────────────

// ActivityRow is one row of the timeline. Identity (UserName/Avatar) is
// hydrated by the service layer via userclient.
type ActivityRow struct {
	TypeStr   string    `gorm:"column:type_str"`
	ID        int       `gorm:"column:id"`
	Content   string    `gorm:"column:content"`
	Link      string    `gorm:"column:link"`
	Created   time.Time `gorm:"column:created"`
	UserID    int       `gorm:"column:user_id"`
	GalgameID int       `gorm:"column:galgame_id"`
}

// ──────────────────────────────────────────
// Queries
// ──────────────────────────────────────────

// GetSource returns the source definition for the given type, or false if
// the type is not recognized.
func (r *ActivityRepository) GetSource(typeStr string) (ActivitySource, bool) {
	s, ok := Sources[typeStr]
	return s, ok
}

// GalgameIDsWithResources returns the subset of ids that have at least one
// download resource (a galgame_resource row). The activity feed uses it to drop
// GALGAME_CREATION rows for resource-less galgames when the viewer is hiding
// them (显示设置 → 显示没有下载资源的 Galgame, default off).
func (r *ActivityRepository) GalgameIDsWithResources(ids []int) map[int]bool {
	out := make(map[int]bool, len(ids))
	if len(ids) == 0 {
		return out
	}
	var found []int
	r.db.Table("galgame_resource").
		Where("galgame_id IN ?", ids).
		Distinct().
		Pluck("galgame_id", &found)
	for _, id := range found {
		out[id] = true
	}
	return out
}

// FetchSingleSource runs a single activity source with pagination and count.
// Identity is hydrated at the service layer via userclient.
func (r *ActivityRepository) FetchSingleSource(src ActivitySource, page, limit int) ([]ActivityRow, int64, error) {
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM (%s) AS sub`, src.Query)
	var total int64
	if err := r.db.Raw(countSQL).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	dataSQL := fmt.Sprintf(
		`SELECT sub.*
		FROM (%s) AS sub
		ORDER BY sub.created DESC
		LIMIT %d OFFSET %d`,
		src.Query, limit, (page-1)*limit,
	)
	var rows []ActivityRow
	if err := r.db.Raw(dataSQL).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// FetchTimeline runs a single UNION ALL query across all source tables,
// letting PostgreSQL handle sort and pagination in one pass. Identity is
// hydrated at the service layer via userclient.
func (r *ActivityRepository) FetchTimeline(page, limit int) ([]ActivityRow, int64, error) {
	// Exact total over the UNLIMITED union for pagination (the /activity page
	// renders ceil(total/limit) page count). COUNT(*) lets Postgres prune each
	// branch's target list to a bare row count — message's `WHERE type = ...`
	// rides idx_message_type_created, the rest are small tables — so this stays
	// cheap even though it spans every source.
	var total int64
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM (%s) AS u`, buildUnionAll())
	if err := r.db.Raw(countSQL).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	// Data: push `ORDER BY created DESC LIMIT (offset+limit)` INTO each source
	// before the UNION. The global top (offset+limit) rows are guaranteed to be
	// contained in the union of each source's own top (offset+limit) — a single
	// source can contribute at most that many rows to a global top-N — so the
	// outer ORDER BY/LIMIT/OFFSET still returns the exact page, but each branch
	// is now an index-backed top-N (idx_<table>_created) instead of a full scan
	// + global sort of every row across all ~17 source tables.
	offset := (page - 1) * limit
	perSource := offset + limit
	dataSQL := fmt.Sprintf(
		`SELECT u.*
		FROM (%s) AS u
		ORDER BY u.created DESC
		LIMIT %d OFFSET %d`,
		buildLimitedUnionAll(perSource), limit, offset,
	)
	var rows []ActivityRow
	if err := r.db.Raw(dataSQL).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// buildUnionAll joins all source queries with UNION ALL (no per-source limit).
// Used only for the exact COUNT(*) over the full activity set.
func buildUnionAll() string {
	parts := make([]string, 0, len(Sources))
	for _, src := range Sources {
		parts = append(parts, "("+src.Query+")")
	}
	return strings.Join(parts, " UNION ALL ")
}

// buildLimitedUnionAll joins all sources with UNION ALL, but first caps each
// source to its own `ORDER BY created DESC LIMIT perSource`. Every source SELECTs
// `t.created` off its base table aliased `t`, so each per-branch ORDER BY/LIMIT
// can use idx_<table>_created and read only the newest perSource rows. perSource
// MUST be >= the page's offset+limit (see FetchTimeline) for the page to be
// exact. Parenthesising each operand keeps the per-branch ORDER BY/LIMIT local
// to that SELECT (not applied to the whole UNION).
func buildLimitedUnionAll(perSource int) string {
	parts := make([]string, 0, len(Sources))
	for _, src := range Sources {
		parts = append(parts, fmt.Sprintf("(%s ORDER BY t.created DESC LIMIT %d)", src.Query, perSource))
	}
	return strings.Join(parts, " UNION ALL ")
}
