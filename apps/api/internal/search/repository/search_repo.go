package repository

import (
	"time"

	"gorm.io/gorm"
)

type SearchRepository struct {
	db *gorm.DB
}

func NewSearchRepository(db *gorm.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

// ──────────────────────────────────────────
// Row projections
// ──────────────────────────────────────────

// TopicRow / ReplyRow / CommentRow no longer carry user identity; the service
// layer hydrates name/avatar via userclient since the user table is no longer
// the source of truth.
type TopicRow struct {
	ID               int
	Title            string
	View             int
	Status           int
	LikeCount        int
	ReplyCount       int
	CommentCount     int
	StatusUpdateTime time.Time
	UserID           int
	IsNSFW           bool
	BestAnswerID     *int
	UpvoteTime       *time.Time
}

// TopicSectionRow + TopicTagRow share the shape used by home_repo:
// {topic_id, name}. Duplicated here rather than imported so the search
// module stays free of inter-module repo dependencies — the queries are
// trivially small and the data shapes won't drift independently.
type TopicSectionRow struct {
	TopicID     int    `gorm:"column:topic_id"`
	SectionName string `gorm:"column:name"`
}

type TopicTagRow struct {
	TopicID int    `gorm:"column:topic_id"`
	TagName string `gorm:"column:name"`
}

type ReplyRow struct {
	ID         int
	TopicID    int
	TopicTitle string
	Content    string
	Floor      int
	UserID     int
	Created    time.Time
}

type CommentRow struct {
	ID         int
	TopicID    int
	TopicTitle string
	Content    string
	UserID     int
	Created    time.Time
}

// ──────────────────────────────────────────
// Queries
// ──────────────────────────────────────────

// SearchTopics fulltext-searches topics by title/content/category. Identity
// is hydrated by the service layer via userclient.
//
// Selects the same superset of fields the FE HomeTopicCard expects so a
// search-topic result renders with all the badges (best-answer / poll /
// NSFW / upvote chip) instead of silently missing them — the card is
// shared with the /home and /topic feeds.
func (r *SearchRepository) SearchTopics(keywords []string, page, limit int) (rows []TopicRow, total int64) {
	query := r.db.Table("topic t").
		Select(`t.id, t.title, t.view, t.status, t.like_count, t.reply_count,
			t.comment_count, t.status_update_time, t.user_id,
			t.is_nsfw, t.best_answer_id, t.upvote_time`).
		Where("t.status != 1")
	for _, kw := range keywords {
		like := "%" + kw + "%"
		query = query.Where("(t.title ILIKE ? OR t.content ILIKE ? OR t.category ILIKE ?)",
			like, like, like)
	}

	query.Count(&total)
	query.Order("t.status_update_time DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&rows)
	return
}

// FindTopicSections groups section names by topic id (same shape as
// home_repo.FindTopicSections).
func (r *SearchRepository) FindTopicSections(topicIDs []int) []TopicSectionRow {
	if len(topicIDs) == 0 {
		return nil
	}
	var rows []TopicSectionRow
	r.db.Table("topic_section_relation tsr").
		Select("tsr.topic_id, ts.name").
		Joins("JOIN topic_section ts ON ts.id = tsr.topic_section_id").
		Where("tsr.topic_id IN ?", topicIDs).
		Find(&rows)
	return rows
}

// FindTopicTags groups tag names by topic id.
func (r *SearchRepository) FindTopicTags(topicIDs []int) []TopicTagRow {
	if len(topicIDs) == 0 {
		return nil
	}
	var rows []TopicTagRow
	r.db.Table("topic_tag_relation ttr").
		Select("ttr.topic_id, tt.name").
		Joins("JOIN topic_tag tt ON tt.id = ttr.tag_id").
		Where("ttr.topic_id IN ?", topicIDs).
		Find(&rows)
	return rows
}

// FindTopicIDsWithPoll returns the subset of topicIDs that have at
// least one row in topic_poll.
func (r *SearchRepository) FindTopicIDsWithPoll(topicIDs []int) map[int]bool {
	out := map[int]bool{}
	if len(topicIDs) == 0 {
		return out
	}
	var rows []struct {
		TopicID int `gorm:"column:topic_id"`
	}
	r.db.Table("topic_poll").
		Where("topic_id IN ?", topicIDs).
		Select("topic_id").
		Scan(&rows)
	for _, row := range rows {
		out[row.TopicID] = true
	}
	return out
}

// SearchReplies searches topic replies by content. Identity is hydrated by
// the service layer via userclient.
func (r *SearchRepository) SearchReplies(keywords []string, page, limit int) (rows []ReplyRow, total int64) {
	// Multi-target replies keep their text in topic_reply_target.content
	// (r.content is empty), so BOTH the snippet and the keyword match must
	// consider target content — otherwise such replies are unsearchable and
	// render blank. Mirrors the activity feed / user-reply list.
	query := r.db.Table("topic_reply r").
		Select(`r.id, r.topic_id, t.title AS topic_title,
			SUBSTRING(
				COALESCE(r.content, '') ||
				COALESCE(
					(SELECT STRING_AGG(trt.content, ' ' ORDER BY trt.id)
					 FROM topic_reply_target trt
					 WHERE trt.reply_id = r.id),
					''
				), 1, 233
			) AS content, r.floor,
			r.user_id, r.created`).
		Joins("LEFT JOIN topic t ON t.id = r.topic_id")
	for _, kw := range keywords {
		like := "%" + kw + "%"
		query = query.Where(
			`r.content ILIKE ? OR EXISTS (
				SELECT 1 FROM topic_reply_target trt
				WHERE trt.reply_id = r.id AND trt.content ILIKE ?
			)`, like, like)
	}

	query.Count(&total)
	query.Order("r.created DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&rows)
	return
}

// SearchComments searches topic comments by content. Identity is hydrated by
// the service layer via userclient.
func (r *SearchRepository) SearchComments(keywords []string, page, limit int) (rows []CommentRow, total int64) {
	query := r.db.Table("topic_comment c").
		Select(`c.id, c.topic_id, t.title AS topic_title,
			SUBSTRING(c.content, 1, 233) AS content,
			c.user_id, c.created`).
		Joins("LEFT JOIN topic t ON t.id = c.topic_id")
	for _, kw := range keywords {
		query = query.Where("c.content ILIKE ?", "%"+kw+"%")
	}

	query.Count(&total)
	query.Order("c.created DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&rows)
	return
}
