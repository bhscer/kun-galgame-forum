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
	// SFWWhere, when non-empty, is a predicate on the source's base table `t`
	// appended as a fresh WHERE for SFW viewers only — hides r18 entities (e.g.
	// r18 galgame websites) from the default feed, mirroring the list endpoints'
	// sfwScope. Only set it on sources whose Query has NO base WHERE (it is
	// appended as `WHERE`, not `AND`).
	SFWWhere string
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
			FROM galgame_activity t WHERE t.type = 'GALGAME_EDIT'`,
	},
	// GALGAME_PR_CREATION: a user submitted an update request (PR). The galgame_pr
	// table lives in the wiki and can't be queried locally, so SubmitPR mirrors
	// each submission into galgame_activity — restoring the timeline entry that
	// was dropped in the wiki migration. user_id is the submitter; content gets
	// the game name during enrichment (same as GALGAME_EDIT).
	"GALGAME_PR_CREATION": {
		TypeStr: "GALGAME_PR_CREATION",
		Query: `SELECT 'GALGAME_PR_CREATION' AS type_str, t.id,
			'' AS content,
			'/galgame/' || t.galgame_id AS link, t.created, t.user_id, t.galgame_id
			FROM galgame_activity t WHERE t.type = 'GALGAME_PR_CREATION'`,
	},
	"GALGAME_RATING_CREATION": {
		TypeStr: "GALGAME_RATING_CREATION",
		// Link to the rating's OWN detail page (/galgame-rating/:ratingId =
		// galgame-rating/[id].vue, keyed by the rating id t.id), NOT the parent
		// galgame page — a "X 评分了 Y" activity should open that review.
		// galgame_id is still selected for actor/content enrichment.
		Query: `SELECT 'GALGAME_RATING_CREATION' AS type_str, t.id,
			SUBSTRING(COALESCE(t.short_summary,''), 1, 100) AS content,
			'/galgame-rating/' || t.id AS link, t.created, t.user_id, t.galgame_id
			FROM galgame_rating t`,
	},
	"GALGAME_RATING_COMMENT_CREATION": {
		TypeStr: "GALGAME_RATING_COMMENT_CREATION",
		// A comment ON a rating links to that rating's detail page
		// (/galgame-rating/:ratingId), NOT /galgame/:ratingId — the old prefix
		// fed a rating id into the galgame route, so the jump went nowhere.
		Query: `SELECT 'GALGAME_RATING_COMMENT_CREATION' AS type_str, t.id,
			SUBSTRING(t.content, 1, 100) AS content,
			'/galgame-rating/' || t.galgame_rating_id AS link, t.created, t.user_id, 0 AS galgame_id
			FROM galgame_rating_comment t`,
	},
	// GALGAME_PR_CREATION removed: galgame_pr table moved to wiki service
	"GALGAME_WEBSITE_CREATION": {
		TypeStr: "GALGAME_WEBSITE_CREATION",
		// Link key is the website's url (a bare domain), NOT its numeric id: the
		// FE route is /website/:domain and the API resolves it via `WHERE url = ?`
		// (website_repo.FindByDomain). Linking by t.id sent every click to a
		// non-existent /website/<id> → "跳转不过去".
		Query: `SELECT 'GALGAME_WEBSITE_CREATION' AS type_str, t.id,
			t.name AS content,
			'/website/' || t.url AS link, t.created, t.user_id, 0 AS galgame_id
			FROM galgame_website t`,
		// SFW viewers don't see r18 websites' activity (mirrors website-list sfwScope).
		SFWWhere: "t.age_limit = 'all'",
	},
	"GALGAME_WEBSITE_COMMENT_CREATION": {
		TypeStr: "GALGAME_WEBSITE_COMMENT_CREATION",
		// Same /website/:domain key as above — resolve the parent website's url
		// from website_id (COALESCE to '' if the website was deleted).
		Query: `SELECT 'GALGAME_WEBSITE_COMMENT_CREATION' AS type_str, t.id,
			SUBSTRING(t.content, 1, 100) AS content,
			'/website/' || COALESCE((SELECT w.url FROM galgame_website w WHERE w.id = t.website_id), '') AS link,
			t.created, t.user_id, 0 AS galgame_id
			FROM galgame_website_comment t`,
		// Hide comments on r18 websites from SFW viewers.
		SFWWhere: "EXISTS (SELECT 1 FROM galgame_website w WHERE w.id = t.website_id AND w.age_limit = 'all')",
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

// FetchTimelineRows returns the newest `n` activity rows across ALL sources
// (offset 0), resource-filtered per showNoResource. It returns no total: the
// service over-fetches a window with this and slices the requested page in
// *filtered* space (after enrichment drops brief-missing rows). That's the fix
// for the under-filled feed — the old fixed `LIMIT n` followed by a post-fetch
// drop left every page short, because a dropped galgame row had already spent a
// LIMIT slot.
func (r *ActivityRepository) FetchTimelineRows(n int, isSFW, showNoResource bool) ([]ActivityRow, error) {
	dataSQL := fmt.Sprintf(
		`SELECT u.* FROM (%s) AS u ORDER BY u.created DESC LIMIT %d`,
		buildLimitedUnionAll(n, isSFW, showNoResource), n,
	)
	var rows []ActivityRow
	if err := r.db.Raw(dataSQL).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// FetchTimelinePage is the legacy offset-paged fetch, kept for DEEP pages
// (beyond the service's over-fetch window) where filtered-space slicing from the
// top would be too expensive. Those pages keep the old "sparse but non-empty"
// behavior. Pushes `ORDER BY created DESC LIMIT (offset+limit)` into each source
// so every branch is an index-backed top-N, not a full scan.
func (r *ActivityRepository) FetchTimelinePage(page, limit int, isSFW, showNoResource bool) ([]ActivityRow, error) {
	offset := (page - 1) * limit
	perSource := offset + limit
	dataSQL := fmt.Sprintf(
		`SELECT u.* FROM (%s) AS u ORDER BY u.created DESC LIMIT %d OFFSET %d`,
		buildLimitedUnionAll(perSource, isSFW, showNoResource), limit, offset,
	)
	var rows []ActivityRow
	if err := r.db.Raw(dataSQL).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// CountTimeline is the exact total over the (resource-filtered) union, used for
// the /activity page bar (ceil(total/limit) pages). COUNT(*) lets Postgres prune
// each branch's target list to a bare row count, so it stays cheap even spanning
// every source.
func (r *ActivityRepository) CountTimeline(isSFW, showNoResource bool) (int64, error) {
	var total int64
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM (%s) AS u`, buildUnionAll(isSFW, showNoResource))
	if err := r.db.Raw(countSQL).Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// FetchSingleSourceRows / FetchSingleSourcePage / CountSingleSource mirror the
// timeline trio for a single type-filtered source.
func (r *ActivityRepository) FetchSingleSourceRows(src ActivitySource, n int, isSFW, showNoResource bool) ([]ActivityRow, error) {
	dataSQL := fmt.Sprintf(
		`SELECT sub.* FROM (%s) AS sub ORDER BY sub.created DESC LIMIT %d`,
		sourceQuery(src, isSFW, showNoResource), n,
	)
	var rows []ActivityRow
	if err := r.db.Raw(dataSQL).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *ActivityRepository) FetchSingleSourcePage(src ActivitySource, page, limit int, isSFW, showNoResource bool) ([]ActivityRow, error) {
	dataSQL := fmt.Sprintf(
		`SELECT sub.* FROM (%s) AS sub ORDER BY sub.created DESC LIMIT %d OFFSET %d`,
		sourceQuery(src, isSFW, showNoResource), limit, (page-1)*limit,
	)
	var rows []ActivityRow
	if err := r.db.Raw(dataSQL).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *ActivityRepository) CountSingleSource(src ActivitySource, isSFW, showNoResource bool) (int64, error) {
	var total int64
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM (%s) AS sub`, sourceQuery(src, isSFW, showNoResource))
	if err := r.db.Raw(countSQL).Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// sourceQuery returns src.Query with the dynamic predicates appended:
//
//   - Resource-less filter: for GALGAME_CREATION when the viewer hides
//     resource-less galgames (showNoResource=false, the default 显示设置), drop
//     creations of galgames with no download resource. Pushed into SQL so these
//     rows never occupy a LIMIT slot and then get dropped during enrichment —
//     that post-LIMIT drop was the root cause of the under-filled feed. Mirrors
//     the galgame list's default filter.
//   - SFW predicate: the source's SFWWhere (as a fresh WHERE) for SFW viewers.
//
// GALGAME_CREATION's base Query has no WHERE and no SFWWhere, so appending a
// single fresh WHERE for the resource filter is unambiguous.
func sourceQuery(src ActivitySource, isSFW, showNoResource bool) string {
	q := src.Query
	if !showNoResource && src.TypeStr == "GALGAME_CREATION" {
		q += " WHERE EXISTS (SELECT 1 FROM galgame_resource r WHERE r.galgame_id = t.id)"
	}
	if isSFW && src.SFWWhere != "" {
		q += " WHERE " + src.SFWWhere
	}
	return q
}

// buildUnionAll joins all source queries with UNION ALL (no per-source limit).
// Used only for the exact COUNT(*) over the full activity set.
func buildUnionAll(isSFW, showNoResource bool) string {
	parts := make([]string, 0, len(Sources))
	for _, src := range Sources {
		parts = append(parts, "("+sourceQuery(src, isSFW, showNoResource)+")")
	}
	return strings.Join(parts, " UNION ALL ")
}

// buildLimitedUnionAll joins all sources with UNION ALL, but first caps each
// source to its own `ORDER BY created DESC LIMIT perSource`. Every source SELECTs
// `t.created` off its base table aliased `t`, so each per-branch ORDER BY/LIMIT
// can use idx_<table>_created and read only the newest perSource rows. perSource
// MUST be >= the page's offset+limit for the page to be exact. Parenthesising
// each operand keeps the per-branch ORDER BY/LIMIT local to that SELECT (not
// applied to the whole UNION).
func buildLimitedUnionAll(perSource int, isSFW, showNoResource bool) string {
	parts := make([]string, 0, len(Sources))
	for _, src := range Sources {
		parts = append(parts, fmt.Sprintf("(%s ORDER BY t.created DESC LIMIT %d)", sourceQuery(src, isSFW, showNoResource), perSource))
	}
	return strings.Join(parts, " UNION ALL ")
}
