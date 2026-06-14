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
	// for SFW viewers only — hides r18 entities (e.g. r18 galgame websites) from
	// the default feed, mirroring the list endpoints' sfwScope.
	SFWWhere string
	// HasBaseWhere must be true when Query already ends in a top-level WHERE
	// (e.g. `… FROM message t WHERE t.type = 'upvoted'`). sourceQuery then
	// AND-combines the resource / SFW / keyset predicates onto it instead of
	// emitting a second WHERE; forgetting it yields `WHERE … WHERE …` at runtime.
	HasBaseWhere bool
}

// Sources is the map of activity type → SQL source.
var Sources = map[string]ActivitySource{
	"TOPIC_CREATION": {
		TypeStr: "TOPIC_CREATION",
		Query: `SELECT 'TOPIC_CREATION' AS type_str, t.id,
			t.title AS content,
			'/topic/' || t.id AS link, t.created, t.user_id, 0 AS galgame_id
			FROM topic t WHERE t.status != 1`,
		HasBaseWhere: true,
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
		HasBaseWhere: true,
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
		HasBaseWhere: true,
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
		HasBaseWhere: true,
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
		HasBaseWhere: true,
	},
	"MESSAGE_SOLUTION": {
		TypeStr: "MESSAGE_SOLUTION",
		Query: `SELECT 'MESSAGE_SOLUTION' AS type_str, t.id, t.content,
			t.link, t.created, t.sender_id AS user_id, 0 AS galgame_id
			FROM message t WHERE t.type = 'solution'`,
		HasBaseWhere: true,
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

// Cursor is the keyset position for the activity feed: the (created, type_str,
// id) of the last row already returned. The feed is a UNION across many source
// tables, so `id` is unique only WITHIN a source — the deterministic total order
// (and thus the cursor) must include type_str. A nil Cursor means "from the
// newest" (first page).
type Cursor struct {
	Created time.Time
	TypeStr string
	ID      int
}

// AllSources returns every registered source (the mixed-timeline fetch). Order
// is irrelevant — FetchKeyset re-sorts the merged result deterministically.
func (r *ActivityRepository) AllSources() []ActivitySource {
	all := make([]ActivitySource, 0, len(Sources))
	for _, src := range Sources {
		all = append(all, src)
	}
	return all
}

// FetchKeyset returns up to `limit` rows from the given sources, strictly older
// than `cur` (or the newest when cur is nil), in the deterministic total order
// `created DESC, type_str DESC, id DESC`. This is the seek-method (keyset)
// pagination that replaces OFFSET: stable across inserts, O(1) regardless of
// depth, and — crucially — with a total-order tiebreaker so rows sharing a
// `created` timestamp can no longer reorder between pages (the old
// `ORDER BY created DESC` had no tiebreaker, so paging duplicated/skipped rows).
//
// Each source branch seeks its own idx_<table>_created via a coarse
// `t.created <= cur.Created` pre-filter plus a per-branch ORDER BY/LIMIT; the
// outer query then applies the EXACT row-value cut `(created,type_str,id) < cur`
// and merges the branches. One source = a type-filtered feed; all sources = the
// mixed timeline. Cursor values are bound as parameters (never interpolated).
func (r *ActivityRepository) FetchKeyset(sources []ActivitySource, limit int, cur *Cursor, isSFW, showNoResource bool) ([]ActivityRow, error) {
	sql, args := buildKeysetSQL(sources, limit, cur, isSFW, showNoResource)
	var rows []ActivityRow
	if err := r.db.Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// buildKeysetSQL assembles the keyset query + its bound args (split out from
// FetchKeyset so it can be unit-tested without a DB). The per-branch
// `t.created <= ?` placeholders bind in branch order; the outer row-value
// comparison's three placeholders bind last — matching their left-to-right
// position in the final SQL.
func buildKeysetSQL(sources []ActivitySource, limit int, cur *Cursor, isSFW, showNoResource bool) (string, []any) {
	parts := make([]string, 0, len(sources))
	args := make([]any, 0, len(sources)+3)
	for _, src := range sources {
		q := sourceQuery(src, isSFW, showNoResource, cur != nil)
		if cur != nil {
			args = append(args, cur.Created)
		}
		parts = append(parts, fmt.Sprintf("(%s ORDER BY t.created DESC LIMIT %d)", q, limit))
	}
	union := strings.Join(parts, " UNION ALL ")

	sql := fmt.Sprintf("SELECT u.* FROM (%s) AS u", union)
	if cur != nil {
		// Row-value comparison = the exact, deterministic keyset cut. The
		// per-branch `<= cur.Created` above is only a coarse, index-usable
		// pre-filter; this makes the boundary precise across every source.
		sql += " WHERE (u.created, u.type_str, u.id) < (?, ?, ?)"
		args = append(args, cur.Created, cur.TypeStr, cur.ID)
	}
	sql += fmt.Sprintf(" ORDER BY u.created DESC, u.type_str DESC, u.id DESC LIMIT %d", limit)
	return sql, args
}

// sourceQuery returns src.Query with the dynamic predicates AND-combined onto a
// single WHERE:
//
//   - Resource-less filter: for GALGAME_CREATION when the viewer hides
//     resource-less galgames (showNoResource=false, the default 显示设置), drop
//     creations of galgames with no download resource. Pushed into SQL so these
//     rows never occupy a LIMIT slot and then get dropped during enrichment.
//     Mirrors the galgame list's default filter.
//   - SFW predicate: the source's SFWWhere for SFW viewers.
//   - Keyset pre-filter: `t.created <= ?` (coarse, index-usable) when paginating
//     from a cursor; the exact cut is the outer row-value comparison in
//     FetchKeyset. The caller binds the parameter.
//
// The hasWhere tracker AND-combines predicates so a second/third one never emits
// a malformed `WHERE … WHERE …` (the old version assumed at most one applied).
func sourceQuery(src ActivitySource, isSFW, showNoResource, withCursor bool) string {
	q := src.Query
	hasWhere := src.HasBaseWhere
	add := func(pred string) {
		if hasWhere {
			q += " AND (" + pred + ")"
		} else {
			q += " WHERE " + pred
			hasWhere = true
		}
	}
	if !showNoResource && src.TypeStr == "GALGAME_CREATION" {
		add("EXISTS (SELECT 1 FROM galgame_resource r WHERE r.galgame_id = t.id)")
	}
	if isSFW && src.SFWWhere != "" {
		add(src.SFWWhere)
	}
	if withCursor {
		add("t.created <= ?")
	}
	return q
}
