package repository

import (
	"fmt"

	"kun-galgame-api/internal/admin/dto"

	"gorm.io/gorm"
)

// PurgeRepository hard-deletes every piece of content + interaction a user
// has produced in kungal, in one transaction, fixing the denormalized
// counters on any surviving parent rows afterwards.
//
// Scope note: this only touches kungal-owned tables. Galgame *entries*
// (the wiki owns those — local `galgame` has no user_id) and account
// identity (OAuth) are out of reach by design. Collateral is deliberately
// minimized: other users' replies/comments that merely *referenced* the
// purged user are kept (their dangling "in reply to" link is dropped), we
// only remove rows that cannot survive without the purged parent.
type PurgeRepository struct {
	db *gorm.DB
}

func NewPurgeRepository(db *gorm.DB) *PurgeRepository {
	return &PurgeRepository{db: db}
}

// CountUserContent returns a read-only breakdown of what a purge would
// delete — used to preview before the admin confirms.
func (r *PurgeRepository) CountUserContent(userID int) dto.UserContentStats {
	return r.counts(r.db, userID)
}

// counts runs the per-type COUNTs (works on either the base DB or a tx).
func (r *PurgeRepository) counts(q *gorm.DB, userID int) dto.UserContentStats {
	var s dto.UserContentStats
	countBy := func(table string) int64 {
		var n int64
		q.Table(table).Where("user_id = ?", userID).Count(&n)
		return n
	}

	s.Topics = countBy("topic")
	s.Replies = countBy("topic_reply")
	s.TopicComments = countBy("topic_comment")
	s.GalgameComments = countBy("galgame_comment")
	s.Ratings = countBy("galgame_rating")
	s.RatingComments = countBy("galgame_rating_comment")
	s.Resources = countBy("galgame_resource")
	s.Websites = countBy("galgame_website")
	s.WebsiteComments = countBy("galgame_website_comment")
	s.Toolsets = countBy("galgame_toolset")
	s.ToolsetResources = countBy("galgame_toolset_resource")
	s.ToolsetComments = countBy("galgame_toolset_comment")

	// Interactions: every like/dislike/favorite/upvote/vote the user cast.
	for _, t := range interactionTables {
		s.Interactions += countBy(t)
	}

	s.Total = s.Topics + s.Replies + s.TopicComments + s.GalgameComments +
		s.Ratings + s.RatingComments + s.Resources + s.Websites +
		s.WebsiteComments + s.Toolsets + s.ToolsetResources +
		s.ToolsetComments + s.Interactions
	return s
}

// interactionTables are the user-cast interaction rows cleared in the purge.
// (Order doesn't matter for counting.)
var interactionTables = []string{
	"topic_like", "topic_dislike", "topic_favorite", "topic_upvote",
	"topic_reply_like", "topic_reply_dislike", "topic_comment_like",
	"topic_poll_vote",
	"galgame_like", "galgame_favorite", "galgame_comment_like",
	"galgame_rating_like", "galgame_resource_like",
	"galgame_website_like", "galgame_website_favorite",
	"galgame_toolset_practicality", "galgame_toolset_contributor",
}

// PurgeUserContent deletes everything (content + interactions) authored/cast
// by userID and returns the pre-delete breakdown of what was removed.
func (r *PurgeRepository) PurgeUserContent(userID int) (dto.UserContentStats, error) {
	stats := r.counts(r.db, userID)

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Accumulators for surviving parents whose denormalized counts must
		// be recomputed after their child rows are removed.
		var affTopics, affReplies, affGalgames, affRatings, affWebsites []int

		// ── A. Topics authored by the user (delete the whole subtree) ──
		ownTopics := pluck(tx, "SELECT id FROM topic WHERE user_id = ?", userID)
		if len(ownTopics) > 0 {
			ownReplies := pluck(tx, "SELECT id FROM topic_reply WHERE topic_id IN ?", ownTopics)
			ownPolls := pluck(tx, "SELECT id FROM topic_poll WHERE topic_id IN ?", ownTopics)

			execIn(tx, "DELETE FROM topic_comment_like WHERE topic_comment_id IN (SELECT id FROM topic_comment WHERE topic_id IN ?)", ownTopics)
			execIn(tx, "DELETE FROM topic_comment WHERE topic_id IN ?", ownTopics)
			if len(ownReplies) > 0 {
				execIn(tx, "DELETE FROM topic_reply_target WHERE reply_id IN ? OR target_reply_id IN ?", ownReplies, ownReplies)
				execIn(tx, "DELETE FROM topic_reply_like WHERE topic_reply_id IN ?", ownReplies)
				execIn(tx, "DELETE FROM topic_reply_dislike WHERE topic_reply_id IN ?", ownReplies)
			}
			execIn(tx, "DELETE FROM topic_reply WHERE topic_id IN ?", ownTopics)
			if len(ownPolls) > 0 {
				execIn(tx, "DELETE FROM topic_poll_vote WHERE poll_id IN ?", ownPolls)
				execIn(tx, "DELETE FROM topic_poll_option WHERE poll_id IN ?", ownPolls)
				execIn(tx, "DELETE FROM topic_poll WHERE id IN ?", ownPolls)
			}
			execIn(tx, "DELETE FROM topic_like WHERE topic_id IN ?", ownTopics)
			execIn(tx, "DELETE FROM topic_dislike WHERE topic_id IN ?", ownTopics)
			execIn(tx, "DELETE FROM topic_favorite WHERE topic_id IN ?", ownTopics)
			execIn(tx, "DELETE FROM topic_upvote WHERE topic_id IN ?", ownTopics)
			execIn(tx, "DELETE FROM topic_tag_relation WHERE topic_id IN ?", ownTopics)
			execIn(tx, "DELETE FROM topic_section_relation WHERE topic_id IN ?", ownTopics)
			execIn(tx, "DELETE FROM topic WHERE id IN ?", ownTopics)
		}

		// ── B. Replies authored by the user on OTHER topics ──
		affTopics = append(affTopics, pluck(tx, "SELECT DISTINCT topic_id FROM topic_reply WHERE user_id = ?", userID)...)
		myReplies := pluck(tx, "SELECT id FROM topic_reply WHERE user_id = ?", userID)
		if len(myReplies) > 0 {
			// surviving topics also lose the comments that lived on these replies
			execIn(tx, "DELETE FROM topic_comment_like WHERE topic_comment_id IN (SELECT id FROM topic_comment WHERE topic_reply_id IN ?)", myReplies)
			execIn(tx, "DELETE FROM topic_comment WHERE topic_reply_id IN ?", myReplies)
			// drop both directions of the target link (keep the targeting reply)
			execIn(tx, "DELETE FROM topic_reply_target WHERE reply_id IN ? OR target_reply_id IN ?", myReplies, myReplies)
			execIn(tx, "DELETE FROM topic_reply_like WHERE topic_reply_id IN ?", myReplies)
			execIn(tx, "DELETE FROM topic_reply_dislike WHERE topic_reply_id IN ?", myReplies)
			execIn(tx, "DELETE FROM topic_reply WHERE id IN ?", myReplies)
		}

		// ── C. Topic comments authored by the user on surviving replies ──
		affTopics = append(affTopics, pluck(tx, "SELECT DISTINCT topic_id FROM topic_comment WHERE user_id = ?", userID)...)
		affReplies = append(affReplies, pluck(tx, "SELECT DISTINCT topic_reply_id FROM topic_comment WHERE user_id = ?", userID)...)
		myComments := pluck(tx, "SELECT id FROM topic_comment WHERE user_id = ?", userID)
		if len(myComments) > 0 {
			execIn(tx, "DELETE FROM topic_comment_like WHERE topic_comment_id IN ?", myComments)
			execIn(tx, "DELETE FROM topic_comment WHERE id IN ?", myComments)
		}

		// Recompute topic-domain counts on surviving parents.
		recount(tx, "topic", "reply_count", "topic_reply", "topic_id", affTopics)
		recount(tx, "topic", "comment_count", "topic_comment", "topic_id", affTopics)
		recount(tx, "topic_reply", "comment_count", "topic_comment", "topic_reply_id", affReplies)

		// ── E. Galgame ratings authored by the user ──
		affGalgames = append(affGalgames, pluck(tx, "SELECT DISTINCT galgame_id FROM galgame_rating WHERE user_id = ?", userID)...)
		myRatings := pluck(tx, "SELECT id FROM galgame_rating WHERE user_id = ?", userID)
		if len(myRatings) > 0 {
			execIn(tx, "DELETE FROM galgame_rating_like WHERE galgame_rating_id IN ?", myRatings)
			execIn(tx, "DELETE FROM galgame_rating_comment WHERE galgame_rating_id IN ?", myRatings)
			execIn(tx, "DELETE FROM galgame_rating WHERE id IN ?", myRatings)
		}

		// ── F. Rating comments authored by the user on surviving ratings ──
		affRatings = append(affRatings, pluck(tx, "SELECT DISTINCT galgame_rating_id FROM galgame_rating_comment WHERE user_id = ?", userID)...)
		execIn(tx, "DELETE FROM galgame_rating_comment WHERE user_id = ?", userID)

		// ── G. Galgame comments by the user, plus every nested reply
		// beneath them (the whole subtree) so no orphaned child is left
		// pointing at a deleted parent — spam threads vanish entirely. ──
		galCommentIDs := subtreeIDs(tx, "galgame_comment", "parent_comment_id", userID)
		if len(galCommentIDs) > 0 {
			affGalgames = append(affGalgames, pluck(tx, "SELECT DISTINCT galgame_id FROM galgame_comment WHERE id IN ?", galCommentIDs)...)
			execIn(tx, "DELETE FROM galgame_comment_like WHERE galgame_comment_id IN ?", galCommentIDs)
			execIn(tx, "DELETE FROM galgame_comment WHERE id IN ?", galCommentIDs)
		}

		// ── H. Galgame resources authored by the user ──
		affGalgames = append(affGalgames, pluck(tx, "SELECT DISTINCT galgame_id FROM galgame_resource WHERE user_id = ?", userID)...)
		myResources := pluck(tx, "SELECT id FROM galgame_resource WHERE user_id = ?", userID)
		if len(myResources) > 0 {
			execIn(tx, "DELETE FROM galgame_resource_link WHERE galgame_resource_id IN ?", myResources)
			execIn(tx, "DELETE FROM galgame_resource_like WHERE galgame_resource_id IN ?", myResources)
			execIn(tx, "DELETE FROM galgame_resource WHERE id IN ?", myResources)
		}

		// Recompute galgame-domain counts on surviving parents.
		recount(tx, "galgame", "rating_count", "galgame_rating", "galgame_id", affGalgames)
		recount(tx, "galgame", "comment_count", "galgame_comment", "galgame_id", affGalgames)
		recount(tx, "galgame", "resource_count", "galgame_resource", "galgame_id", affGalgames)
		recount(tx, "galgame_rating", "comment_count", "galgame_rating_comment", "galgame_rating_id", affRatings)

		// ── J. Websites authored by the user (whole subtree) ──
		ownWebsites := pluck(tx, "SELECT id FROM galgame_website WHERE user_id = ?", userID)
		if len(ownWebsites) > 0 {
			execIn(tx, "DELETE FROM galgame_website_comment WHERE website_id IN ?", ownWebsites)
			execIn(tx, "DELETE FROM galgame_website_like WHERE website_id IN ?", ownWebsites)
			execIn(tx, "DELETE FROM galgame_website_favorite WHERE website_id IN ?", ownWebsites)
			execIn(tx, "DELETE FROM galgame_website_tag_relation WHERE galgame_website_id IN ?", ownWebsites)
			execIn(tx, "DELETE FROM galgame_website WHERE id IN ?", ownWebsites)
		}

		// ── K. Website comments by the user + their nested sub-comment
		// subtree (galgame_website_comment.parent_id). ──
		wcIDs := subtreeIDs(tx, "galgame_website_comment", "parent_id", userID)
		if len(wcIDs) > 0 {
			affWebsites = append(affWebsites, pluck(tx, "SELECT DISTINCT website_id FROM galgame_website_comment WHERE id IN ?", wcIDs)...)
			execIn(tx, "DELETE FROM galgame_website_comment WHERE id IN ?", wcIDs)
		}
		recount(tx, "galgame_website", "comment_count", "galgame_website_comment", "website_id", affWebsites)

		// ── L. Toolsets authored by the user (whole subtree) ──
		ownToolsets := pluck(tx, "SELECT id FROM galgame_toolset WHERE user_id = ?", userID)
		if len(ownToolsets) > 0 {
			execIn(tx, "DELETE FROM galgame_toolset_resource WHERE toolset_id IN ?", ownToolsets)
			execIn(tx, "DELETE FROM galgame_toolset_comment WHERE toolset_id IN ?", ownToolsets)
			execIn(tx, "DELETE FROM galgame_toolset_contributor WHERE toolset_id IN ?", ownToolsets)
			execIn(tx, "DELETE FROM galgame_toolset_practicality WHERE toolset_id IN ?", ownToolsets)
			execIn(tx, "DELETE FROM galgame_toolset_alias WHERE toolset_id IN ?", ownToolsets)
			execIn(tx, "DELETE FROM galgame_toolset_category_relation WHERE toolset_id IN ?", ownToolsets)
			execIn(tx, "DELETE FROM galgame_toolset WHERE id IN ?", ownToolsets)
		}

		// ── M. Toolset resources authored on surviving toolsets ──
		execIn(tx, "DELETE FROM galgame_toolset_resource WHERE user_id = ?", userID)
		// ── N. Toolset comments by the user + their nested sub-comment
		// subtree (galgame_toolset_comment.parent_id). ──
		if tcIDs := subtreeIDs(tx, "galgame_toolset_comment", "parent_id", userID); len(tcIDs) > 0 {
			execIn(tx, "DELETE FROM galgame_toolset_comment WHERE id IN ?", tcIDs)
		}

		// ── O. Interactions cast by the user (recompute each parent count) ──
		for _, it := range interactionDeletes {
			ids := pluck(tx, "SELECT DISTINCT "+it.parentCol+" FROM "+it.table+" WHERE user_id = ?", userID)
			if err := tx.Exec("DELETE FROM "+it.table+" WHERE user_id = ?", userID).Error; err != nil {
				return err
			}
			if it.parentTable != "" {
				recount(tx, it.parentTable, it.countCol, it.table, it.parentCol, ids)
			}
		}

		return nil
	})

	return stats, err
}

// interactionDelete describes one interaction table and the parent counter to
// recompute after the user's rows are removed. parentTable == "" means the
// parent has no denormalized counter (just delete).
type interactionDelete struct {
	table       string
	parentCol   string
	parentTable string
	countCol    string
}

var interactionDeletes = []interactionDelete{
	{"topic_like", "topic_id", "topic", "like_count"},
	{"topic_dislike", "topic_id", "topic", "dislike_count"},
	{"topic_favorite", "topic_id", "topic", "favorite_count"},
	{"topic_upvote", "topic_id", "topic", "upvote_count"},
	{"topic_reply_like", "topic_reply_id", "topic_reply", "like_count"},
	{"topic_reply_dislike", "topic_reply_id", "topic_reply", "dislike_count"},
	{"topic_comment_like", "topic_comment_id", "", ""},
	{"topic_poll_vote", "option_id", "topic_poll_option", "vote_count"},
	{"galgame_like", "galgame_id", "galgame", "like_count"},
	{"galgame_favorite", "galgame_id", "galgame", "favorite_count"},
	{"galgame_comment_like", "galgame_comment_id", "galgame_comment", "like_count"},
	{"galgame_rating_like", "galgame_rating_id", "galgame_rating", "like_count"},
	{"galgame_resource_like", "galgame_resource_id", "galgame_resource", "like_count"},
	{"galgame_website_like", "website_id", "galgame_website", "like_count"},
	{"galgame_website_favorite", "website_id", "galgame_website", "favorite_count"},
	{"galgame_toolset_practicality", "toolset_id", "", ""},
	{"galgame_toolset_contributor", "toolset_id", "", ""},
}

// pluck runs a single-int-column SELECT and returns the ids (nil on empty).
func pluck(tx *gorm.DB, query string, args ...any) []int {
	var ids []int
	tx.Raw(query, args...).Scan(&ids)
	return ids
}

// subtreeIDs returns the ids of every comment authored by userID in a
// self-referential comment table PLUS all descendants nested beneath them
// (children of any author), walked via parentCol. Used so purging a spammer
// removes whole comment threads rather than leaving children orphaned at a
// deleted parent. table/parentCol are caller constants — never user input.
func subtreeIDs(tx *gorm.DB, table, parentCol string, userID int) []int {
	q := fmt.Sprintf(`WITH RECURSIVE tree AS (
		SELECT id FROM %s WHERE user_id = ?
		UNION
		SELECT c.id FROM %s c JOIN tree t ON c.%s = t.id
	) SELECT id FROM tree`, table, table, parentCol)
	var ids []int
	tx.Raw(q, userID).Scan(&ids)
	return ids
}

// execIn runs a DELETE; ignored when the caller already guards the id slice.
func execIn(tx *gorm.DB, query string, args ...any) {
	tx.Exec(query, args...)
}

// recount sets parentTable.countCol = COUNT(childTable WHERE childTable.parentCol = parentTable.id)
// for the given (surviving) parent ids. Table/column names are caller-owned
// constants — never user input — so the fmt.Sprintf is injection-safe.
func recount(tx *gorm.DB, parentTable, countCol, childTable, parentCol string, ids []int) {
	if len(ids) == 0 {
		return
	}
	sql := fmt.Sprintf(
		"UPDATE %s SET %s = (SELECT COUNT(*) FROM %s WHERE %s.%s = %s.id) WHERE %s.id IN ?",
		parentTable, countCol, childTable, childTable, parentCol, parentTable, parentTable,
	)
	tx.Exec(sql, ids)
}
