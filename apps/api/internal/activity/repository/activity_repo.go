package repository

import (
	"fmt"
	"strings"
	"time"

	topicModel "kun-galgame-api/internal/topic/model"

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
		// A reply's text is in topic_reply.content; the legacy multi-target
		// rows were folded into it by the Phase-4 migration. The full body is
		// sent (not truncated) — the reply card renders it as Markdown.
		Query: `SELECT 'TOPIC_REPLY_CREATION' AS type_str, t.id,
			COALESCE(t.content, '') AS content,
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
	"TOPIC_UPVOTE": {
		TypeStr: "TOPIC_UPVOTE",
		// A user pushed (推) a topic. id = the upvote row id (unique per push, so
		// the cursor + uniqueId don't collide when many users push one topic).
		// content = the optional "why" one-liner; the card pulls the topic's
		// title / excerpt / footer via topic_id during enrichment.
		Query: `SELECT 'TOPIC_UPVOTE' AS type_str, t.id,
			COALESCE(t.description, '') AS content,
			'/topic/' || t.topic_id AS link, t.created, t.user_id, 0 AS galgame_id
			FROM topic_upvote t`,
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
		// Full comment body (not truncated) — the card renders it as Markdown.
		Query: `SELECT 'GALGAME_COMMENT_CREATION' AS type_str, t.id,
			t.content AS content,
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
		// Spoiler-flagged ratings (spoiler_level <> 'none') don't leak their
		// summary into the public feed — the reader opens the detail page to
		// see it. Wording mirrors KUN_GALGAME_RATING_SPOILER_WARNING on the FE.
		Query: `SELECT 'GALGAME_RATING_CREATION' AS type_str, t.id,
			CASE WHEN t.spoiler_level <> 'none'
			     THEN '⚠️ 该评分可能含有剧透内容，点进查看'
			     ELSE SUBSTRING(COALESCE(t.short_summary,''), 1, 100) END AS content,
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

// GetSources returns the sources for the given type strings (unknown ones are
// skipped). Used by the home feed's tab buckets — a fixed subset of the
// timeline merged with the same keyset machinery as AllSources.
func (r *ActivityRepository) GetSources(types []string) []ActivitySource {
	out := make([]ActivitySource, 0, len(types))
	for _, t := range types {
		if s, ok := Sources[t]; ok {
			out = append(out, s)
		}
	}
	return out
}

// TopicCardData is the per-topic core enrichment for the feed's rich topic card
// (one row from the topic table). Sections/tags/poll/top-reply are separate
// batch loaders below. Excerpt is a server-truncated slice of the body.
type TopicCardData struct {
	ID            int                    `gorm:"column:id"`
	Title         string                 `gorm:"column:title"`
	Excerpt       string                 `gorm:"column:excerpt"`
	CoverImages   topicModel.ImageTokens `gorm:"column:cover_images"`
	View          int                    `gorm:"column:view"`
	LikeCount     int                    `gorm:"column:like_count"`
	FavoriteCount int                    `gorm:"column:favorite_count"`
	ReplyCount    int                    `gorm:"column:reply_count"`
	CommentCount  int                    `gorm:"column:comment_count"`
	IsNSFW        bool                   `gorm:"column:is_nsfw"`
	UpvoteTime    *time.Time             `gorm:"column:upvote_time"`
	BestAnswerID  *int                   `gorm:"column:best_answer_id"`
	AuthorID      int                    `gorm:"column:user_id"`
}

// FetchTopicActivityData batch-loads the topic core row for the given ids (one
// query, no N+1), keyed by id. The body is truncated to a preview in SQL so a
// 100k-char topic never bloats the feed payload. Empty ids → empty map.
func (r *ActivityRepository) FetchTopicActivityData(ids []int) (map[int]TopicCardData, error) {
	out := make(map[int]TopicCardData, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	var rows []TopicCardData
	if err := r.db.Table("topic").
		Select("id, title, user_id, SUBSTRING(content, 1, 300) AS excerpt, cover_images, view, like_count, favorite_count, reply_count, comment_count, is_nsfw, upvote_time, best_answer_id").
		Where("id IN ?", ids).
		Scan(&rows).Error; err != nil {
		return out, err
	}
	for _, row := range rows {
		out[row.ID] = row
	}
	return out, nil
}

// FetchUpvoteTopics maps each topic_upvote row id → its topic id, so the 推话题
// card can pull the pushed topic's data via the topic enrichment. Empty → empty.
func (r *ActivityRepository) FetchUpvoteTopics(upvoteIDs []int) (map[int]int, error) {
	out := map[int]int{}
	if len(upvoteIDs) == 0 {
		return out, nil
	}
	var rows []struct {
		ID      int `gorm:"column:id"`
		TopicID int `gorm:"column:topic_id"`
	}
	if err := r.db.Table("topic_upvote").Select("id, topic_id").
		Where("id IN ?", upvoteIDs).Scan(&rows).Error; err != nil {
		return out, err
	}
	for _, row := range rows {
		out[row.ID] = row.TopicID
	}
	return out, nil
}

// TopicReactionCountRow is one windowed reactor row for the feed cards: the
// (topic, reaction key) with its total count (cnt, repeated on every row) plus
// ONE reactor's user_id — up to feedReactionAvatarCap rows per key.
type TopicReactionCountRow struct {
	TopicID  int    `gorm:"column:topic_id"`
	Reaction string `gorm:"column:reaction"`
	Count    int    `gorm:"column:cnt"`
	UserID   int    `gorm:"column:user_id"`
}

// feedReactionAvatarCap bounds the reactor avatars fetched per (topic, reaction)
// for the feed cards; the rest collapse to a "+N" (mirrors the detail's cap of 3).
const feedReactionAvatarCap = 3

// FetchTopicsReactions batch-loads each topic's reaction keys with their total
// count AND up to feedReactionAvatarCap reactor ids per key (a windowed query),
// so the feed card can render ≤3 avatars + a "+N". Empty ids → empty slice.
func (r *ActivityRepository) FetchTopicsReactions(ids []int) ([]TopicReactionCountRow, error) {
	out := []TopicReactionCountRow{}
	if len(ids) == 0 {
		return out, nil
	}
	err := r.db.Raw(`
		SELECT topic_id, reaction, user_id, cnt FROM (
			SELECT topic_id, reaction, user_id,
				COUNT(*) OVER (PARTITION BY topic_id, reaction) AS cnt,
				ROW_NUMBER() OVER (PARTITION BY topic_id, reaction ORDER BY id) AS rn
			FROM topic_reaction WHERE topic_id IN ?
		) t WHERE rn <= ? ORDER BY topic_id, reaction, rn`, ids, feedReactionAvatarCap).
		Scan(&out).Error
	return out, err
}

// FetchTodoStatuses maps todo id → status (0待处理/1进行中/2已完成/3已废弃) for the
// 其他-tab Note card. Empty ids → empty map.
func (r *ActivityRepository) FetchTodoStatuses(ids []int) (map[int]int, error) {
	out := map[int]int{}
	if len(ids) == 0 {
		return out, nil
	}
	var rows []struct {
		ID     int `gorm:"column:id"`
		Status int `gorm:"column:status"`
	}
	if err := r.db.Table("todo").Select("id, status").
		Where("id IN ?", ids).Scan(&rows).Error; err != nil {
		return out, err
	}
	for _, row := range rows {
		out[row.ID] = row.Status
	}
	return out, nil
}

// FetchUpdateLogVersions maps update_log id → version string for the Note card.
// Empty ids → empty map.
func (r *ActivityRepository) FetchUpdateLogVersions(ids []int) (map[int]string, error) {
	out := map[int]string{}
	if len(ids) == 0 {
		return out, nil
	}
	var rows []struct {
		ID      int    `gorm:"column:id"`
		Version string `gorm:"column:version"`
	}
	if err := r.db.Table("update_log").Select("id, version").
		Where("id IN ?", ids).Scan(&rows).Error; err != nil {
		return out, err
	}
	for _, row := range rows {
		out[row.ID] = row.Version
	}
	return out, nil
}

// TopicCommentContext is the owning topic's title + the reply (被评论的评论) for a
// TOPIC_COMMENT_CREATION card.
type TopicCommentContext struct {
	CommentID    int    `gorm:"column:comment_id"`
	TopicTitle   string `gorm:"column:topic_title"`
	ReplyFloor   int    `gorm:"column:reply_floor"`
	ReplyContent string `gorm:"column:reply_content"`
}

// FetchTopicCommentContext loads, per comment id, the parent reply's floor +
// content and the owning topic's title (one JOIN). Empty ids → empty map.
func (r *ActivityRepository) FetchTopicCommentContext(ids []int) (map[int]TopicCommentContext, error) {
	out := map[int]TopicCommentContext{}
	if len(ids) == 0 {
		return out, nil
	}
	var rows []TopicCommentContext
	if err := r.db.Table("topic_comment c").
		Select(`c.id AS comment_id, tp.title AS topic_title,
			r.floor AS reply_floor, r.content AS reply_content`).
		Joins("JOIN topic_reply r ON r.id = c.topic_reply_id").
		Joins("JOIN topic tp ON tp.id = r.topic_id").
		Where("c.id IN ?", ids).Scan(&rows).Error; err != nil {
		return out, err
	}
	for _, row := range rows {
		out[row.CommentID] = row
	}
	return out, nil
}

// FetchGalgameCommentParents loads, per comment id, the parent comment's content
// (被评论的评论). Only comments WITH a parent appear in the result.
func (r *ActivityRepository) FetchGalgameCommentParents(ids []int) (map[int]string, error) {
	out := map[int]string{}
	if len(ids) == 0 {
		return out, nil
	}
	var rows []struct {
		CommentID int    `gorm:"column:comment_id"`
		Content   string `gorm:"column:content"`
	}
	if err := r.db.Table("galgame_comment c").
		Select("c.id AS comment_id, p.content AS content").
		Joins("JOIN galgame_comment p ON p.id = c.parent_comment_id").
		Where("c.id IN ?", ids).Scan(&rows).Error; err != nil {
		return out, err
	}
	for _, row := range rows {
		out[row.CommentID] = row.Content
	}
	return out, nil
}

// fetchParentNames is the shared id→parent-name lookup for the toolset / website
// cards: `childTable c JOIN parentTable p ON p.id = c.<fk>` selecting p.name.
func (r *ActivityRepository) fetchParentNames(childTable, parentTable, fk string, ids []int) (map[int]string, error) {
	out := map[int]string{}
	if len(ids) == 0 {
		return out, nil
	}
	var rows []struct {
		ID   int    `gorm:"column:id"`
		Name string `gorm:"column:name"`
	}
	if err := r.db.Table(childTable+" c").
		Select("c.id, p.name").
		Joins("JOIN "+parentTable+" p ON p.id = c."+fk).
		Where("c.id IN ?", ids).Scan(&rows).Error; err != nil {
		return out, err
	}
	for _, row := range rows {
		out[row.ID] = row.Name
	}
	return out, nil
}

// FetchToolsetResourceParents maps each toolset-resource id → its toolset name.
func (r *ActivityRepository) FetchToolsetResourceParents(ids []int) (map[int]string, error) {
	return r.fetchParentNames("galgame_toolset_resource", "galgame_toolset", "toolset_id", ids)
}

// FetchToolsetCommentParents maps each toolset-comment id → its toolset name.
func (r *ActivityRepository) FetchToolsetCommentParents(ids []int) (map[int]string, error) {
	return r.fetchParentNames("galgame_toolset_comment", "galgame_toolset", "toolset_id", ids)
}

// FetchWebsiteCommentParents maps each website-comment id → its website name.
func (r *ActivityRepository) FetchWebsiteCommentParents(ids []int) (map[int]string, error) {
	return r.fetchParentNames("galgame_website_comment", "galgame_website", "website_id", ids)
}

// GalgameResourceRow is one resource's feed-card spec (no download link / codes).
type GalgameResourceRow struct {
	ID        int    `gorm:"column:id"`
	Type      string `gorm:"column:type"`
	Language  string `gorm:"column:language"`
	Platform  string `gorm:"column:platform"`
	Size      string `gorm:"column:size"`
	Note      string `gorm:"column:note"`
	LikeCount int    `gorm:"column:like_count"`
}

// FetchGalgameResourceDetails batch-loads resource specs by id. Empty → empty.
func (r *ActivityRepository) FetchGalgameResourceDetails(ids []int) (map[int]GalgameResourceRow, error) {
	out := map[int]GalgameResourceRow{}
	if len(ids) == 0 {
		return out, nil
	}
	var rows []GalgameResourceRow
	if err := r.db.Table("galgame_resource").
		Select("id, type, language, platform, size, note, like_count").
		Where("id IN ?", ids).Scan(&rows).Error; err != nil {
		return out, err
	}
	for _, row := range rows {
		out[row.ID] = row
	}
	return out, nil
}

// idNameRow is a (topic_id, name) pair for the section/tag batch joins.
type idNameRow struct {
	TopicID int    `gorm:"column:topic_id"`
	Name    string `gorm:"column:name"`
}

func collectIDNames(rows []idNameRow) map[int][]string {
	out := map[int][]string{}
	for _, row := range rows {
		out[row.TopicID] = append(out[row.TopicID], row.Name)
	}
	return out
}

// FetchTopicSections batch-loads section names per topic id (topic_id → names).
func (r *ActivityRepository) FetchTopicSections(ids []int) (map[int][]string, error) {
	if len(ids) == 0 {
		return map[int][]string{}, nil
	}
	var rows []idNameRow
	if err := r.db.Table("topic_section_relation tsr").
		Select("tsr.topic_id, ts.name").
		Joins("JOIN topic_section ts ON ts.id = tsr.topic_section_id").
		Where("tsr.topic_id IN ?", ids).
		Scan(&rows).Error; err != nil {
		return map[int][]string{}, err
	}
	return collectIDNames(rows), nil
}

// FetchTopicPolls returns the set of given topic ids that have a poll attached.
func (r *ActivityRepository) FetchTopicPolls(ids []int) (map[int]bool, error) {
	out := map[int]bool{}
	if len(ids) == 0 {
		return out, nil
	}
	var rows []struct {
		TopicID int `gorm:"column:topic_id"`
	}
	if err := r.db.Table("topic_poll").
		Select("DISTINCT topic_id").
		Where("topic_id IN ?", ids).
		Scan(&rows).Error; err != nil {
		return out, err
	}
	for _, row := range rows {
		out[row.TopicID] = true
	}
	return out, nil
}

// FetchReplyTopicTitles batch-loads the parent topic title for each reply id
// (reply_id → topic.title) — the feed's reply card shows it at the bottom.
func (r *ActivityRepository) FetchReplyTopicTitles(replyIDs []int) (map[int]string, error) {
	out := map[int]string{}
	if len(replyIDs) == 0 {
		return out, nil
	}
	var rows []struct {
		ID    int    `gorm:"column:id"`
		Title string `gorm:"column:title"`
	}
	if err := r.db.Raw(`
		SELECT r.id, t.title
		FROM topic_reply r JOIN topic t ON t.id = r.topic_id
		WHERE r.id IN ?`, replyIDs).Scan(&rows).Error; err != nil {
		return out, err
	}
	for _, row := range rows {
		out[row.ID] = row.Title
	}
	return out, nil
}

// ReplyContent is a quoted reply's floor + raw body (tokens unresolved), for the
// feed's reply card.
type ReplyContent struct {
	Floor   int
	Content string
}

// FetchReplyContents batch-loads floor + content for reply ids — the quoted
// replies referenced by feed reply cards (#floor tokens).
func (r *ActivityRepository) FetchReplyContents(replyIDs []int) (map[int]ReplyContent, error) {
	out := map[int]ReplyContent{}
	if len(replyIDs) == 0 {
		return out, nil
	}
	var rows []struct {
		ID      int    `gorm:"column:id"`
		Floor   int    `gorm:"column:floor"`
		Content string `gorm:"column:content"`
	}
	if err := r.db.Raw(`
		SELECT id, floor, content
		FROM topic_reply
		WHERE id IN ?`, replyIDs).Scan(&rows).Error; err != nil {
		return out, err
	}
	for _, row := range rows {
		out[row.ID] = ReplyContent{Floor: row.Floor, Content: row.Content}
	}
	return out, nil
}

// GalgameCounts holds a galgame's local rollups for the new-galgame feed card.
type GalgameCounts struct {
	ResourceCount int
	LikeCount     int
	FavoriteCount int
}

// FetchGalgameCounts batch-loads resource/like/favorite counts for galgame ids
// from the local galgame table (global counts — cache-safe).
func (r *ActivityRepository) FetchGalgameCounts(galgameIDs []int) (map[int]GalgameCounts, error) {
	out := map[int]GalgameCounts{}
	if len(galgameIDs) == 0 {
		return out, nil
	}
	var rows []struct {
		ID            int `gorm:"column:id"`
		ResourceCount int `gorm:"column:resource_count"`
		LikeCount     int `gorm:"column:like_count"`
		FavoriteCount int `gorm:"column:favorite_count"`
	}
	if err := r.db.Raw(`
		SELECT id, resource_count, like_count, favorite_count
		FROM galgame
		WHERE id IN ?`, galgameIDs).Scan(&rows).Error; err != nil {
		return out, err
	}
	for _, row := range rows {
		out[row.ID] = GalgameCounts{
			ResourceCount: row.ResourceCount,
			LikeCount:     row.LikeCount,
			FavoriteCount: row.FavoriteCount,
		}
	}
	return out, nil
}

// EditRevision pairs a GALGAME_EDIT activity's wiki revision row id (global —
// the input for the id→number fallback) with the per-galgame revision NUMBER
// (what the diff endpoint's :rev expects). RevisionNumber is 0 for rows synced
// before the wiki feed started carrying `revision`.
type EditRevision struct {
	RevisionID     int
	RevisionNumber int
}

// FetchEditRevisions maps galgame_activity ids → their wiki revision ref, for the
// feed's edit card to lazily load the diff (directly via the number, or via the
// id→number resolution fallback for legacy rows where the number is unknown).
func (r *ActivityRepository) FetchEditRevisions(activityIDs []int) (map[int]EditRevision, error) {
	out := map[int]EditRevision{}
	if len(activityIDs) == 0 {
		return out, nil
	}
	var rows []struct {
		ID         int  `gorm:"column:id"`
		RevisionID int  `gorm:"column:wiki_revision_id"`
		RevisionNo *int `gorm:"column:wiki_revision_number"`
	}
	if err := r.db.Raw(`
		SELECT id, wiki_revision_id, wiki_revision_number
		FROM galgame_activity
		WHERE id IN ?`, activityIDs).Scan(&rows).Error; err != nil {
		return out, err
	}
	for _, row := range rows {
		num := 0
		if row.RevisionNo != nil {
			num = *row.RevisionNo
		}
		out[row.ID] = EditRevision{RevisionID: row.RevisionID, RevisionNumber: num}
	}
	return out, nil
}

// RatingActivity is a galgame rating's feed-card fields. ShortSummary is blanked
// when there's a spoiler so it never leaves the boundary.
type RatingActivity struct {
	Overall      int
	PlayStatus   string
	Recommend    string
	ShortSummary string
	SpoilerLevel string
	LikeCount    int
	AuthorID     int
}

// FetchRatingActivityData batch-loads rating fields by galgame_rating id for the
// feed's rating card. Summaries of spoiler-flagged ratings are dropped.
func (r *ActivityRepository) FetchRatingActivityData(ratingIDs []int) (map[int]RatingActivity, error) {
	out := map[int]RatingActivity{}
	if len(ratingIDs) == 0 {
		return out, nil
	}
	var rows []struct {
		ID           int    `gorm:"column:id"`
		Overall      int    `gorm:"column:overall"`
		PlayStatus   string `gorm:"column:play_status"`
		Recommend    string `gorm:"column:recommend"`
		ShortSummary string `gorm:"column:short_summary"`
		SpoilerLevel string `gorm:"column:spoiler_level"`
		LikeCount    int    `gorm:"column:like_count"`
		UserID       int    `gorm:"column:user_id"`
	}
	if err := r.db.Raw(`
		SELECT id, overall, play_status, recommend, short_summary, spoiler_level, like_count, user_id
		FROM galgame_rating
		WHERE id IN ?`, ratingIDs).Scan(&rows).Error; err != nil {
		return out, err
	}
	for _, row := range rows {
		summary := row.ShortSummary
		if row.SpoilerLevel != "none" {
			summary = "" // never leak a spoiler-flagged summary into the feed
		}
		out[row.ID] = RatingActivity{
			Overall:      row.Overall,
			PlayStatus:   row.PlayStatus,
			Recommend:    row.Recommend,
			ShortSummary: summary,
			SpoilerLevel: row.SpoilerLevel,
			LikeCount:    row.LikeCount,
			AuthorID:     row.UserID,
		}
	}
	return out, nil
}

// TopReplyRow is a topic's most-liked reply (excerpt + like count).
type TopReplyRow struct {
	TopicID   int    `gorm:"column:topic_id"`
	Content   string `gorm:"column:content"`
	LikeCount int    `gorm:"column:like_count"`
	UserID    int    `gorm:"column:user_id"`
}

// FetchTopicTopReply batch-loads each topic's MOST-LIKED reply via DISTINCT ON
// (one row per topic, highest like_count, id as the tiebreaker), restricted to
// replies that actually have likes (>0) so the card only surfaces a notable one.
func (r *ActivityRepository) FetchTopicTopReply(ids []int) (map[int]TopReplyRow, error) {
	out := map[int]TopReplyRow{}
	if len(ids) == 0 {
		return out, nil
	}
	var rows []TopReplyRow
	if err := r.db.Raw(`
		SELECT DISTINCT ON (topic_id) topic_id,
			SUBSTRING(content, 1, 200) AS content, like_count, user_id
		FROM topic_reply
		WHERE topic_id IN ? AND like_count > 0
		ORDER BY topic_id, like_count DESC, id DESC`, ids).
		Scan(&rows).Error; err != nil {
		return out, err
	}
	for _, row := range rows {
		out[row.TopicID] = row
	}
	return out, nil
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
func (r *ActivityRepository) FetchKeyset(sources []ActivitySource, limit int, cur *Cursor, isSFW, showNoResource bool, tab string) ([]ActivityRow, error) {
	sql, args := buildKeysetSQL(sources, limit, cur, isSFW, showNoResource, tab)
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
func buildKeysetSQL(sources []ActivitySource, limit int, cur *Cursor, isSFW, showNoResource bool, tab string) (string, []any) {
	parts := make([]string, 0, len(sources))
	args := make([]any, 0, len(sources)+3)
	for _, src := range sources {
		q := sourceQuery(src, isSFW, showNoResource, cur != nil, tab)
		if cur != nil {
			// Args for the precise per-branch keyset cut emitted by sourceQuery:
			// t.created < ? · t.created = ? · (?<?) branch<cursor type ·
			// (?=?) branch=cursor type · t.id < ?. type_str is bound (not Go-
			// compared) so PG uses the same collation as the outer row-value cut.
			args = append(args, cur.Created, cur.Created, src.TypeStr, cur.TypeStr, src.TypeStr, cur.TypeStr, cur.ID)
		}
		// id DESC tiebreaker (type_str is constant within a branch) so the
		// LIMIT slice picks the rows DIRECTLY below the cursor — without it,
		// equal-`created` rows ordered arbitrarily and the slice could be all
		// already-seen rows, returning 0 and freezing the feed.
		parts = append(parts, fmt.Sprintf("(%s ORDER BY t.created DESC, t.id DESC LIMIT %d)", q, limit))
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
func sourceQuery(src ActivitySource, isSFW, showNoResource, withCursor bool, tab string) string {
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
	// 资源和求助 topics — sections g-seeking / g-other / t-help — belong ONLY in the
	// 资源和求助 tab (tab="resource", alongside the galgame resources). Every other
	// feed (全部 / 话题 / timeline / single-type) excludes them.
	if src.TypeStr == "TOPIC_CREATION" {
		const inHelp = "EXISTS (SELECT 1 FROM topic_section_relation tsr " +
			"JOIN topic_section ts ON ts.id = tsr.topic_section_id " +
			"WHERE tsr.topic_id = t.id AND ts.name IN ('g-seeking','g-other','t-help'))"
		if tab == "resource" {
			add(inHelp)
		} else {
			add("NOT " + inHelp)
		}
	}
	if isSFW && src.SFWWhere != "" {
		add(src.SFWWhere)
	}
	if withCursor {
		// EXACT total-order cut for this branch — equivalent to the outer
		// (created, type_str, id) < cursor, specialised to this branch's
		// constant type_str (bound as a param, compared by PG). Pushing it into
		// the branch means its `ORDER BY … LIMIT` returns only qualifying rows,
		// so a cluster of equal-`created` rows can't waste the whole window and
		// return 0 (which froze single-source feeds like 资源). The outer cut in
		// FetchKeyset stays as a redundant safety net + the cross-branch merge.
		add("t.created < ? OR (t.created = ? AND (? < ? OR (? = ? AND t.id < ?)))")
	}
	return q
}

// FetchFeed is the materialized-table replacement for FetchKeyset: instead of a
// UNION across ~18 source tables, it keyset-paginates the single feed_activity
// table (maintained by triggers — see migration 034), so per-page cost is flat
// regardless of depth / source count. `types` is the tab's activity-type set
// (nil/empty = the whole timeline). The dynamic filters that used to live in each
// UNION branch are applied once, here:
//
//   - SFW: drop is_nsfw rows (r18 website activity) for SFW viewers.
//   - resource-less: drop GALGAME_CREATION of galgames with no download resource
//     (unless showNoResource) so they never occupy a LIMIT slot.
//   - section: 资源/求助 topics (g-seeking/g-other/t-help) belong ONLY in the
//     资源和求助 tab (tab="resource"); every other feed excludes them.
//
// Order + cut use the SAME (created, type, source_id) total order as the cursor
// (feed_activity has UNIQUE(type, source_id)), so serveKeyset / encode / decode
// are reused unchanged. Galgame NSFW is still dropped at enrichment (wiki-owned),
// bounded by activityMaxRounds.
func (r *ActivityRepository) FetchFeed(types []string, limit int, cur *Cursor, isSFW, showNoResource bool, tab string) ([]ActivityRow, error) {
	conds := make([]string, 0, 6)
	args := make([]any, 0, 6)
	if len(types) > 0 {
		conds = append(conds, "fa.type IN ?")
		args = append(args, types)
	}
	if isSFW {
		conds = append(conds, "NOT fa.is_nsfw")
	}
	if !showNoResource {
		conds = append(conds, "(fa.type <> 'GALGAME_CREATION' OR EXISTS (SELECT 1 FROM galgame_resource r WHERE r.galgame_id = fa.galgame_id))")
	}
	const inHelp = "EXISTS (SELECT 1 FROM topic_section_relation tsr " +
		"JOIN topic_section ts ON ts.id = tsr.topic_section_id " +
		"WHERE tsr.topic_id = fa.source_id AND ts.name IN ('g-seeking','g-other','t-help'))"
	if tab == "resource" {
		conds = append(conds, "(fa.type <> 'TOPIC_CREATION' OR "+inHelp+")")
	} else {
		conds = append(conds, "(fa.type <> 'TOPIC_CREATION' OR NOT "+inHelp+")")
	}
	if cur != nil {
		conds = append(conds, "(fa.created, fa.type, fa.source_id) < (?, ?, ?)")
		args = append(args, cur.Created, cur.TypeStr, cur.ID)
	}

	sql := "SELECT fa.type AS type_str, fa.source_id AS id, fa.content, fa.link, fa.created, fa.user_id, fa.galgame_id FROM feed_activity fa"
	if len(conds) > 0 {
		sql += " WHERE " + strings.Join(conds, " AND ")
	}
	sql += fmt.Sprintf(" ORDER BY fa.created DESC, fa.type DESC, fa.source_id DESC LIMIT %d", limit)

	var rows []ActivityRow
	if err := r.db.Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
