package service

import (
	"context"
	"fmt"

	"kun-galgame-api/internal/galgame/model"
	"kun-galgame-api/internal/galgame/repository"
	msgModel "kun-galgame-api/internal/message/model"
	userRepo "kun-galgame-api/internal/user/repository"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/userclient"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CommentService struct {
	commentRepo *repository.CommentRepository
	stateRepo   *userRepo.StateRepository
	userClient  *userclient.Client
	helpers     InteractionHelpers
}

func NewCommentService(
	commentRepo *repository.CommentRepository,
	stateRepo *userRepo.StateRepository,
	userClient *userclient.Client,
) *CommentService {
	return &CommentService{commentRepo: commentRepo, stateRepo: stateRepo, userClient: userClient}
}

// ──────────────────────────────────────────
// Response types
// ──────────────────────────────────────────

type UserObj struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

// CommentItem is the wire shape. Roots have Replies populated with the
// full descendant subtree; the frontend visually caps recursion at
// depth 3 and opens the drawer for deeper threads. ReplyCount counts
// only direct + transitive descendants of THIS node (not the whole
// thread under the root), so a mid-tree "查看更多" button can use it.
type CommentItem struct {
	ID              int            `json:"id"`
	Content         string         `json:"content"`
	GalgameID       int            `json:"galgameId"`
	User            UserObj        `json:"user"`
	TargetUser      *UserObj       `json:"targetUser"`
	ParentCommentID *int           `json:"parentCommentId"`
	RootCommentID   *int           `json:"rootCommentId"`
	LikeCount       int            `json:"likeCount"`
	Created         string         `json:"created"`
	Replies         []*CommentItem `json:"replies"`
	ReplyCount      int            `json:"replyCount"`
}

type CommentListResult struct {
	Items []*CommentItem
	Total int64
}

// ──────────────────────────────────────────
// GetComments
// ──────────────────────────────────────────

// GetComments returns paginated ROOT comments with their full
// descendant subtree attached. `total` is the root count (drives
// pagination); display-side counts can use CommentItem.ReplyCount to
// reason about nesting depth.
func (s *CommentService) GetComments(ctx context.Context, galgameID, page, limit int) *CommentListResult {
	total := s.commentRepo.CountRootsByGalgame(galgameID)
	roots := s.commentRepo.FindRootsPaginated(galgameID, page, limit)
	if len(roots) == 0 {
		return &CommentListResult{Items: []*CommentItem{}, Total: total}
	}

	rootIDs := make([]int, 0, len(roots))
	for _, r := range roots {
		rootIDs = append(rootIDs, r.ID)
	}
	replies := s.commentRepo.FindRepliesByRoots(rootIDs)

	all := make([]repository.CommentRow, 0, len(roots)+len(replies))
	all = append(all, roots...)
	all = append(all, replies...)

	userMap := s.hydrateUsers(ctx, all)
	items := s.buildItems(all, userMap)

	// Roots are returned in DESC order; preserve that.
	rootItems := make([]*CommentItem, 0, len(roots))
	for _, r := range roots {
		if it, ok := items[r.ID]; ok {
			rootItems = append(rootItems, it)
		}
	}

	return &CommentListResult{Items: rootItems, Total: total}
}

// ──────────────────────────────────────────
// GetThread
// ──────────────────────────────────────────

// GetThread returns the full thread starting at rootID. Used by the
// drawer; depth is unbounded. The returned CommentItem is the root
// with all descendants nested inside Replies.
func (s *CommentService) GetThread(ctx context.Context, rootID int) (*CommentItem, *errors.AppError) {
	rows := s.commentRepo.FindThreadByRoot(rootID)
	if len(rows) == 0 {
		return nil, errors.ErrNotFound("评论线程不存在")
	}
	userMap := s.hydrateUsers(ctx, rows)
	items := s.buildItems(rows, userMap)
	root, ok := items[rootID]
	if !ok {
		return nil, errors.ErrNotFound("评论线程不存在")
	}
	return root, nil
}

// ──────────────────────────────────────────
// Tree assembly
// ──────────────────────────────────────────

func (s *CommentService) hydrateUsers(ctx context.Context, rows []repository.CommentRow) map[int]userclient.User {
	uidSet := make(map[int]struct{}, len(rows)*2)
	for _, r := range rows {
		uidSet[r.UserID] = struct{}{}
		if r.TargetUserID != nil && *r.TargetUserID > 0 {
			uidSet[*r.TargetUserID] = struct{}{}
		}
	}
	uids := make([]int, 0, len(uidSet))
	for id := range uidSet {
		uids = append(uids, id)
	}
	return s.userClient.Hydrate(ctx, uids)
}

// buildItems converts flat rows to a parent→children tree. Returns a
// map keyed by comment ID for direct lookup; each item's Replies slice
// is populated, sorted by creation (already pre-sorted ASC by the
// repo). ReplyCount is filled in via a post-order walk.
//
// Rows whose author is no longer renderable (deleted user) are dropped
// silently; their replies are dropped with them — keeping orphans
// would render as ghost branches.
func (s *CommentService) buildItems(rows []repository.CommentRow, userMap map[int]userclient.User) map[int]*CommentItem {
	items := make(map[int]*CommentItem, len(rows))
	for _, r := range rows {
		u := userMap[r.UserID]
		if !userclient.IsRenderable(u) {
			continue
		}
		item := &CommentItem{
			ID:              r.ID,
			Content:         r.Content,
			GalgameID:       r.GalgameID,
			User:            UserObj{ID: u.ID, Name: u.Name, Avatar: u.Avatar},
			ParentCommentID: r.ParentCommentID,
			RootCommentID:   r.RootCommentID,
			LikeCount:       r.LikeCount,
			Created:         r.CreatedAt,
			Replies:         []*CommentItem{},
		}
		if r.TargetUserID != nil {
			if t, ok := userMap[*r.TargetUserID]; ok {
				item.TargetUser = &UserObj{ID: t.ID, Name: t.Name, Avatar: t.Avatar}
			}
		}
		items[r.ID] = item
	}

	// Attach each non-root item to its parent. Skips items whose parent
	// was filtered out above (e.g. parent author deleted).
	for _, item := range items {
		if item.ParentCommentID == nil {
			continue
		}
		parent, ok := items[*item.ParentCommentID]
		if !ok {
			continue
		}
		parent.Replies = append(parent.Replies, item)
	}

	// Post-order: sum descendants. Iterate in dependency-safe order by
	// walking from items with no replies first — but simpler is a
	// recursive count starting from roots. Since the row set is small
	// (one page or one thread) recursion depth is bounded.
	var count func(it *CommentItem) int
	count = func(it *CommentItem) int {
		n := 0
		for _, c := range it.Replies {
			n += 1 + count(c)
		}
		it.ReplyCount = n
		return n
	}
	for _, item := range items {
		if item.ParentCommentID == nil {
			count(item)
		}
	}

	return items
}

// ──────────────────────────────────────────
// CreateComment
// ──────────────────────────────────────────

// CreateComment creates a top-level or reply comment.
//   - parentCommentID == nil → root comment.
//   - parentCommentID != nil → reply; we look up the parent, verify
//     it belongs to the same galgame, and derive root_comment_id as
//     parent.RootCommentID (if parent is itself a reply) or parent.ID
//     (if parent is a root).
func (s *CommentService) CreateComment(
	ctx context.Context,
	userID, galgameID int,
	content string,
	targetUserID *int,
	parentCommentID *int,
) (*CommentItem, *errors.AppError) {
	var (
		parentID *int
		rootID   *int
	)
	if parentCommentID != nil {
		parent, err := s.commentRepo.FindByID(*parentCommentID)
		if err != nil || parent.GalgameID != galgameID {
			return nil, errors.ErrBadRequest("回复的评论不存在")
		}
		pid := parent.ID
		parentID = &pid
		if parent.RootCommentID != nil {
			rid := *parent.RootCommentID
			rootID = &rid
		} else {
			rid := parent.ID
			rootID = &rid
		}
	}

	comment := model.GalgameComment{
		Content:         content,
		GalgameID:       galgameID,
		UserID:          userID,
		TargetUserID:    targetUserID,
		ParentCommentID: parentID,
		RootCommentID:   rootID,
	}

	txErr := s.commentRepo.DB().Transaction(func(tx *gorm.DB) error {
		// Lazy-create stub: a pending submission has no kungal stub yet
		// (decision 2 in the wiki integration plan), so the very first
		// interaction must INSERT one or the comment_count UPDATE below
		// silently affects 0 rows.
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).
			Create(&model.GalgameLocal{ID: galgameID}).Error; err != nil {
			return err
		}
		if err := tx.Create(&comment).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.GalgameLocal{}).Where("id = ?", galgameID).
			Update("comment_count", gorm.Expr("comment_count + 1")).Error; err != nil {
			return err
		}

		if targetUserID != nil && *targetUserID != userID {
			if err := s.stateRepo.AdjustMoemoepointTx(tx, *targetUserID, 1); err != nil {
				return err
			}

			link := fmt.Sprintf("/galgame/%d", galgameID)
			if err := tx.Create(&msgModel.Message{
				SenderID: userID, ReceiverID: *targetUserID,
				Type: "commented", Content: truncate(content, 233),
				Link: link, Status: "unread",
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if txErr != nil {
		return nil, errors.ErrInternal("发表评论失败")
	}

	creator, _, _ := s.userClient.User(ctx, userID)
	resp := &CommentItem{
		ID:              comment.ID,
		Content:         comment.Content,
		GalgameID:       comment.GalgameID,
		User:            UserObj{ID: creator.ID, Name: creator.Name, Avatar: creator.Avatar},
		ParentCommentID: comment.ParentCommentID,
		RootCommentID:   comment.RootCommentID,
		LikeCount:       0,
		Created:         comment.CreatedAt.Format("2006-01-02T15:04:05Z"),
		Replies:         []*CommentItem{},
	}
	if targetUserID != nil {
		t, _, _ := s.userClient.User(ctx, *targetUserID)
		resp.TargetUser = &UserObj{ID: t.ID, Name: t.Name, Avatar: t.Avatar}
	}

	return resp, nil
}

// ──────────────────────────────────────────
// DeleteComment
// ──────────────────────────────────────────

// DeleteComment removes a comment. Cascading FK deletes child rows in
// galgame_comment (parent_comment_id ON DELETE CASCADE), so the
// comment_count adjustment subtracts the size of the destroyed subtree.
func (s *CommentService) DeleteComment(userID, role, commentID int) *errors.AppError {
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		return errors.ErrNotFound("未找到该评论")
	}
	if comment.UserID != userID && role < 2 {
		return errors.ErrForbidden("您没有权限删除此评论")
	}

	// Count this comment + all descendants so we can decrement the
	// galgame.comment_count by the full subtree size in one shot
	// (Postgres cascades the actual row deletes).
	var subtreeSize int64
	s.commentRepo.DB().Model(&model.GalgameComment{}).
		Where("id = ? OR root_comment_id = ?", commentID, commentID).
		Count(&subtreeSize)

	txErr := s.commentRepo.DB().Transaction(func(tx *gorm.DB) error {
		// Likes on descendant comments are cleared by their cascade
		// chain through galgame_comment → galgame_comment_like.
		tx.Where("galgame_comment_id = ?", commentID).Delete(&model.GalgameCommentLike{})
		tx.Delete(&comment)
		tx.Model(&model.GalgameLocal{}).Where("id = ?", comment.GalgameID).
			Update("comment_count", gorm.Expr("comment_count - ?", subtreeSize))
		return nil
	})
	if txErr != nil {
		return errors.ErrInternal("删除评论失败")
	}

	return nil
}

// ──────────────────────────────────────────
// ToggleCommentLike
// ──────────────────────────────────────────

func (s *CommentService) ToggleCommentLike(userID, commentID int) *errors.AppError {
	txErr := s.commentRepo.DB().Transaction(func(tx *gorm.DB) error {
		var comment model.GalgameComment
		tx.First(&comment, commentID)

		var existing model.GalgameCommentLike
		result := tx.Where("user_id = ? AND galgame_comment_id = ?", userID, commentID).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			tx.Create(&model.GalgameCommentLike{UserID: userID, CommentID: commentID})
			tx.Model(&model.GalgameComment{}).Where("id = ?", commentID).
				Update("like_count", gorm.Expr("like_count + 1"))
			if comment.UserID != userID {
				if err := s.stateRepo.AdjustMoemoepointTx(tx, comment.UserID, 1); err != nil {
					return err
				}
				s.helpers.CreateGalgameMessageWithContent(
					tx, userID, comment.UserID, "liked",
					truncate(comment.Content, 233),
					comment.GalgameID,
				)
			}
		} else {
			tx.Delete(&existing)
			tx.Model(&model.GalgameComment{}).Where("id = ?", commentID).
				Update("like_count", gorm.Expr("like_count - 1"))
			if comment.UserID != userID {
				if err := s.stateRepo.AdjustMoemoepointTx(tx, comment.UserID, -1); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if txErr != nil {
		return errors.ErrInternal("操作失败")
	}

	return nil
}

// ──────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen])
}
