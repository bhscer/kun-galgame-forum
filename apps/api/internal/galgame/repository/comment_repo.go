package repository

import (
	"kun-galgame-api/internal/galgame/model"

	"gorm.io/gorm"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) DB() *gorm.DB { return r.db }

// CommentRow holds a galgame comment row. Identity is hydrated by the
// service layer via userclient.
type CommentRow struct {
	ID              int
	Content         string
	GalgameID       int
	UserID          int
	TargetUserID    *int
	ParentCommentID *int
	RootCommentID   *int
	LikeCount       int
	CreatedAt       string
}

const commentSelect = `gc.id, gc.content, gc.galgame_id, gc.user_id,
	gc.target_user_id, gc.parent_comment_id, gc.root_comment_id,
	gc.like_count, gc.created AS created_at`

// CountByGalgame returns total comment count for a galgame (includes
// every nested reply). Used for the "X 条评论" header label.
func (r *CommentRepository) CountByGalgame(galgameID int) int64 {
	var total int64
	r.db.Model(&model.GalgameComment{}).
		Where("galgame_id = ?", galgameID).
		Count(&total)
	return total
}

// CountRootsByGalgame returns root-comment count for a galgame, used to
// drive pagination of the threaded view.
func (r *CommentRepository) CountRootsByGalgame(galgameID int) int64 {
	var total int64
	r.db.Model(&model.GalgameComment{}).
		Where("galgame_id = ? AND parent_comment_id IS NULL", galgameID).
		Count(&total)
	return total
}

// FindRootsPaginated returns paginated ROOT comments for a galgame.
// Replies are pulled in a second call via FindRepliesByRoots so the
// service can assemble the tree.
func (r *CommentRepository) FindRootsPaginated(galgameID, page, limit int) []CommentRow {
	var rows []CommentRow
	r.db.Table("galgame_comment gc").
		Select(commentSelect).
		Where("gc.galgame_id = ? AND gc.parent_comment_id IS NULL", galgameID).
		Order("gc.created DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&rows)
	return rows
}

// FindRepliesByRoots returns every descendant reply of the given roots
// in created-ASC order. We pull the whole subtree per root (typically
// small for galgame threads) and let the service build the parent→child
// tree in memory. Callers receive a flat list; map to tree by
// parent_comment_id.
func (r *CommentRepository) FindRepliesByRoots(rootIDs []int) []CommentRow {
	if len(rootIDs) == 0 {
		return nil
	}
	var rows []CommentRow
	r.db.Table("galgame_comment gc").
		Select(commentSelect).
		Where("gc.root_comment_id IN ?", rootIDs).
		Order("gc.created ASC").
		Find(&rows)
	return rows
}

// FindThreadByRoot returns root + all descendants for a single root.
// Used by the "查看完整线程" drawer; depth is unbounded.
func (r *CommentRepository) FindThreadByRoot(rootID int) []CommentRow {
	var rows []CommentRow
	r.db.Table("galgame_comment gc").
		Select(commentSelect).
		Where("gc.id = ? OR gc.root_comment_id = ?", rootID, rootID).
		Order("gc.created ASC").
		Find(&rows)
	return rows
}

// FindByID returns a comment by its primary key.
func (r *CommentRepository) FindByID(id int) (*model.GalgameComment, error) {
	var comment model.GalgameComment
	err := r.db.First(&comment, id).Error
	return &comment, err
}
