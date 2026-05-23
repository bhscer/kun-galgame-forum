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

// FindLatestDescendantsByRoots returns up to `perRoot` MOST RECENT
// descendants per root — any depth, partitioned by root_comment_id.
// Used by the inline list, which flattens the thread to a single
// visual tier: the 3 picked here are whichever 3 are freshest in the
// thread, regardless of DB depth.
//
// Sorted DESC by created so the newest reply appears at the top of
// the inline list (the user's own just-posted reply is immediately
// visible). Anything older lives behind the "查看更多 N 条回复"
// lazy drawer.
func (r *CommentRepository) FindLatestDescendantsByRoots(rootIDs []int, perRoot int) []CommentRow {
	if len(rootIDs) == 0 {
		return nil
	}
	var rows []CommentRow
	r.db.Raw(`
		SELECT id, content, galgame_id, user_id,
		       target_user_id, parent_comment_id, root_comment_id,
		       like_count, created_at
		FROM (
			SELECT gc.id, gc.content, gc.galgame_id, gc.user_id,
			       gc.target_user_id, gc.parent_comment_id, gc.root_comment_id,
			       gc.like_count, gc.created AS created_at,
			       ROW_NUMBER() OVER (
			           PARTITION BY gc.root_comment_id
			           ORDER BY gc.created DESC
			       ) AS rn
			FROM galgame_comment gc
			WHERE gc.root_comment_id IN ?
		) sub
		WHERE rn <= ?
		ORDER BY root_comment_id, created_at DESC
	`, rootIDs, perRoot).Scan(&rows)
	return rows
}

// CountDescendantsByRoots returns total descendant count per root
// (every comment whose root_comment_id is in rootIDs). Drives the
// root-level "查看更多 N 条回复" button: if the count exceeds what's
// rendered inline, the drawer opens and lazy-loads the full subtree.
func (r *CommentRepository) CountDescendantsByRoots(rootIDs []int) map[int]int {
	if len(rootIDs) == 0 {
		return map[int]int{}
	}
	var rows []struct {
		RootID int   `gorm:"column:root_comment_id"`
		Count  int64 `gorm:"column:count"`
	}
	r.db.Table("galgame_comment").
		Select("root_comment_id, COUNT(*) AS count").
		Where("root_comment_id IN ?", rootIDs).
		Group("root_comment_id").
		Scan(&rows)
	out := make(map[int]int, len(rows))
	for _, row := range rows {
		out[row.RootID] = int(row.Count)
	}
	return out
}

// FindThreadByRoot returns root + all descendants for a single root.
// Used by the "回复详情" drawer; depth is unbounded.
//
// Sorted DESC by created so the drawer also presents newest-first,
// matching the inline list's order. The service layer separates the
// root from descendants before returning, so the relative position of
// root in this slice doesn't matter.
func (r *CommentRepository) FindThreadByRoot(rootID int) []CommentRow {
	var rows []CommentRow
	r.db.Table("galgame_comment gc").
		Select(commentSelect).
		Where("gc.id = ? OR gc.root_comment_id = ?", rootID, rootID).
		Order("gc.created DESC").
		Find(&rows)
	return rows
}

// FindByID returns a comment by its primary key.
func (r *CommentRepository) FindByID(id int) (*model.GalgameComment, error) {
	var comment model.GalgameComment
	err := r.db.First(&comment, id).Error
	return &comment, err
}
