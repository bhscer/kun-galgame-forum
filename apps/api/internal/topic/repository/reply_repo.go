package repository

import (
	"kun-galgame-api/internal/topic/model"

	"gorm.io/gorm"
)

// ReplyRepository owns topic-reply rows: CRUD, paginated lists, reply-level
// interactions (like/dislike + count adjustments), cascade delete and the
// reply-author brief projection.
//
// Comment rows (TopicComment) have their own sibling repo in this package:
//   - CommentRepository (comment_repo.go)
type ReplyRepository struct {
	db *gorm.DB
}

func NewReplyRepository(db *gorm.DB) *ReplyRepository {
	return &ReplyRepository{db: db}
}

func (r *ReplyRepository) DB() *gorm.DB {
	return r.db
}

// ──────────────────────────────────────────
// Core CRUD
// ──────────────────────────────────────────

func (r *ReplyRepository) FindByID(id int) (*model.TopicReply, error) {
	var reply model.TopicReply
	err := r.db.First(&reply, id).Error
	return &reply, err
}

func (r *ReplyRepository) GetMaxFloor(tx *gorm.DB, topicID int) (int, error) {
	var maxFloor *int
	err := tx.Model(&model.TopicReply{}).
		Where("topic_id = ?", topicID).
		Select("COALESCE(MAX(floor), 0)").
		Scan(&maxFloor).Error
	if err != nil || maxFloor == nil {
		return 0, err
	}
	return *maxFloor, nil
}

type ReplyRow struct {
	model.TopicReply
	UserName        string
	UserAvatar      string
	UserMoemoepoint int
}

func (r *ReplyRepository) FindRepliesPaginated(
	topicID int,
	excludeIDs []int,
	page, limit int,
	sortOrder string,
) ([]ReplyRow, error) {
	var rows []ReplyRow
	query := r.db.Table("topic_reply").
		Select(`topic_reply.*`).
		Where("topic_reply.topic_id = ?", topicID)

	if len(excludeIDs) > 0 {
		query = query.Where("topic_reply.id NOT IN ?", excludeIDs)
	}

	err := query.
		Order("topic_reply.floor " + sortOrder).
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *ReplyRepository) FindRepliesByIDs(ids []int) ([]ReplyRow, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []ReplyRow
	err := r.db.Table("topic_reply").
		Select(`topic_reply.*`).
		Where("topic_reply.id IN ?", ids).
		Find(&rows).Error
	return rows, err
}

// ──────────────────────────────────────────
// Interaction status (batch)
// ──────────────────────────────────────────

func (r *ReplyRepository) FindReplyLikeStatus(userID int, replyIDs []int) (map[int]bool, error) {
	return findInteractionStatus(r.db, "topic_reply_like", "topic_reply_id", userID, replyIDs)
}

func (r *ReplyRepository) FindReplyDislikeStatus(userID int, replyIDs []int) (map[int]bool, error) {
	return findInteractionStatus(r.db, "topic_reply_dislike", "topic_reply_id", userID, replyIDs)
}

// findInteractionStatus is shared by both ReplyRepository and CommentRepository
// (see comment_repo.go) to resolve a user's boolean interaction state across
// an ID batch.
func findInteractionStatus(db *gorm.DB, table, fkCol string, userID int, ids []int) (map[int]bool, error) {
	if len(ids) == 0 || userID == 0 {
		return make(map[int]bool), nil
	}
	var foundIDs []int
	err := db.Table(table).
		Where("user_id = ? AND "+fkCol+" IN ?", userID, ids).
		Pluck(fkCol, &foundIDs).Error
	if err != nil {
		return nil, err
	}
	result := make(map[int]bool, len(foundIDs))
	for _, id := range foundIDs {
		result[id] = true
	}
	return result, nil
}

// ──────────────────────────────────────────
// Cascade delete helpers
// ──────────────────────────────────────────

func (r *ReplyRepository) DeleteRepliesByIDs(tx *gorm.DB, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	// Delete child rows (comment likes, comments, likes, dislikes) first
	tx.Exec("DELETE FROM topic_comment_like WHERE topic_comment_id IN (SELECT id FROM topic_comment WHERE topic_reply_id IN ?)", ids)
	tx.Where("topic_reply_id IN ?", ids).Delete(&model.TopicComment{})
	tx.Where("topic_reply_id IN ?", ids).Delete(&model.TopicReplyLike{})
	tx.Where("topic_reply_id IN ?", ids).Delete(&model.TopicReplyDislike{})

	return tx.Where("id IN ?", ids).Delete(&model.TopicReply{}).Error
}

// CountReplyRelated returns counts used for moemoepoint penalty calculation.
func (r *ReplyRepository) CountReplyRelated(replyID int) (commentCount, likeCount int64, err error) {
	r.db.Model(&model.TopicComment{}).Where("topic_reply_id = ?", replyID).Count(&commentCount)
	r.db.Model(&model.TopicReplyLike{}).Where("topic_reply_id = ?", replyID).Count(&likeCount)
	return
}

// ──────────────────────────────────────────
// Reply interaction helpers (for tx participation)
// ──────────────────────────────────────────

// FindByIDTx loads a reply inside a transaction.
func (r *ReplyRepository) FindByIDTx(tx *gorm.DB, replyID int) (*model.TopicReply, error) {
	var reply model.TopicReply
	err := tx.First(&reply, replyID).Error
	return &reply, err
}

// FindReplyLike returns an existing reply-like row (or ErrRecordNotFound).
func (r *ReplyRepository) FindReplyLike(tx *gorm.DB, userID, replyID int) (*model.TopicReplyLike, error) {
	var existing model.TopicReplyLike
	err := tx.Where("user_id = ? AND topic_reply_id = ?", userID, replyID).First(&existing).Error
	return &existing, err
}

// CreateReplyLike inserts a new reply-like row.
func (r *ReplyRepository) CreateReplyLike(tx *gorm.DB, userID, replyID int) error {
	return tx.Create(&model.TopicReplyLike{UserID: userID, TopicReplyID: replyID}).Error
}

// DeleteReplyLike removes a previously fetched reply-like row.
func (r *ReplyRepository) DeleteReplyLike(tx *gorm.DB, like *model.TopicReplyLike) error {
	return tx.Delete(like).Error
}

// FindReplyDislike returns an existing reply-dislike row (or ErrRecordNotFound).
func (r *ReplyRepository) FindReplyDislike(tx *gorm.DB, userID, replyID int) (*model.TopicReplyDislike, error) {
	var existing model.TopicReplyDislike
	err := tx.Where("user_id = ? AND topic_reply_id = ?", userID, replyID).First(&existing).Error
	return &existing, err
}

// CreateReplyDislike inserts a new reply-dislike row.
func (r *ReplyRepository) CreateReplyDislike(tx *gorm.DB, userID, replyID int) error {
	return tx.Create(&model.TopicReplyDislike{UserID: userID, TopicReplyID: replyID}).Error
}

// DeleteReplyDislike removes a previously fetched reply-dislike row.
func (r *ReplyRepository) DeleteReplyDislike(tx *gorm.DB, dislike *model.TopicReplyDislike) error {
	return tx.Delete(dislike).Error
}

// AdjustReplyLikeCount adjusts the topic_reply.like_count by `delta`.
func (r *ReplyRepository) AdjustReplyLikeCount(tx *gorm.DB, replyID, delta int) error {
	return tx.Model(&model.TopicReply{}).Where("id = ?", replyID).
		Update("like_count", gorm.Expr("like_count + ?", delta)).Error
}

// AdjustReplyDislikeCount adjusts the topic_reply.dislike_count by `delta`.
func (r *ReplyRepository) AdjustReplyDislikeCount(tx *gorm.DB, replyID, delta int) error {
	return tx.Model(&model.TopicReply{}).Where("id = ?", replyID).
		Update("dislike_count", gorm.Expr("dislike_count + ?", delta)).Error
}

// ──────────────────────────────────────────
// Reply CRUD helpers (tx-aware)
// ──────────────────────────────────────────

// CreateReply inserts a TopicReply inside the caller tx.
func (r *ReplyRepository) CreateReply(tx *gorm.DB, reply *model.TopicReply) error {
	return tx.Create(reply).Error
}

// UpdateReplyContent updates content + edited timestamp for a reply.
func (r *ReplyRepository) UpdateReplyContent(tx *gorm.DB, replyID int, fields map[string]any) error {
	return tx.Model(&model.TopicReply{}).Where("id = ?", replyID).Updates(fields).Error
}
