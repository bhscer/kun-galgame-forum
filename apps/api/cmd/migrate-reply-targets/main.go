// migrate-reply-targets folds the legacy multi-target reply model into the new
// single-body model: each topic_reply_target row becomes an inline
// "> 回复 [@](kungal-user:id) [#floor](kungal-reply:id)\n\n<note>" block prepended
// to its authoring reply's content, after which the target row is removed.
//
// Safety (docs/proj/mention.md §7):
//   - DRY-RUN BY DEFAULT — reports a sample + counts, writes nothing.
//     Pass -dry-run=false to apply.
//   - On apply it first snapshots topic_reply_target + the affected reply
//     contents into *_backup tables (CREATE TABLE IF NOT EXISTS, so a re-run
//     never clobbers the original snapshot — that's the rollback).
//   - Idempotent + resumable: each reply is migrated in its own tx (update
//     content + delete its targets), so a re-run only sees un-migrated rows.
//   - No notifications are emitted — this is a data move, not new activity.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"kun-galgame-api/internal/infrastructure/database"
	"kun-galgame-api/internal/topic/model"
	"kun-galgame-api/pkg/config"
	"kun-galgame-api/pkg/logger"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

// targetRow is one legacy topic_reply_target joined to its target reply (via
// LEFT JOIN, so a since-deleted target reply yields TargetExists=false and is
// treated as dangling — note kept, no @/# link).
type targetRow struct {
	ReplyID       int
	TargetReplyID int
	Note          string
	TargetFloor   int
	TargetUserID  int
	TargetExists  bool
}

// buildMigratedContent composes the new reply body: each target folded into a
// blockquote header above the original content. The mention name is left empty
// on purpose — the server resolves the CURRENT name at render
// (markdown.ResolveMentionNames), so the migration never has to touch OAuth.
func buildMigratedContent(original string, targets []targetRow) string {
	blocks := make([]string, 0, len(targets)+1)
	for _, t := range targets {
		note := strings.TrimSpace(t.Note)
		header := ""
		if t.TargetExists && t.TargetReplyID > 0 {
			header = fmt.Sprintf(
				"> 回复 [@](kungal-user:%d) [#%d](kungal-reply:%d)",
				t.TargetUserID, t.TargetFloor, t.TargetReplyID,
			)
		}
		switch {
		case header != "" && note != "":
			blocks = append(blocks, header+"\n\n"+note)
		case header != "":
			blocks = append(blocks, header)
		case note != "":
			// Dangling target: keep the note, drop the unresolvable @/# link.
			blocks = append(blocks, note)
		}
	}
	if strings.TrimSpace(original) != "" {
		blocks = append(blocks, original)
	}
	return strings.Join(blocks, "\n\n")
}

func main() {
	_ = godotenv.Load()

	dryRun := flag.Bool("dry-run", true,
		"TRUE (default): report what would change, write nothing. Pass -dry-run=false to apply.")
	limit := flag.Int("limit", 0,
		"Max authoring replies to process (0 = all). For testing against a DB copy.")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("加载配置失败", "error", err)
		os.Exit(1)
	}
	logger.Init(cfg.Server.Mode)
	db := database.NewPostgres(cfg.Database, cfg.Server.Mode)

	if err := run(db, *dryRun, *limit); err != nil {
		slog.Error("迁移失败", "error", err)
		os.Exit(1)
	}
}

func run(db *gorm.DB, dryRun bool, limit int) error {
	// 1. Load every legacy target row, joined to its target reply for floor+author.
	var rows []targetRow
	if err := db.Table("topic_reply_target AS tt").
		Select(`tt.reply_id AS reply_id,
			tt.target_reply_id AS target_reply_id,
			tt.content AS note,
			COALESCE(tr.floor, 0) AS target_floor,
			COALESCE(tr.user_id, 0) AS target_user_id,
			(tr.id IS NOT NULL) AS target_exists`).
		Joins("LEFT JOIN topic_reply AS tr ON tr.id = tt.target_reply_id").
		Order("tt.reply_id ASC, COALESCE(tr.floor, 0) ASC, tt.id ASC").
		Scan(&rows).Error; err != nil {
		return fmt.Errorf("加载 topic_reply_target 失败: %w", err)
	}

	if len(rows) == 0 {
		slog.Info("没有待迁移的 topic_reply_target，已是干净状态")
		return nil
	}

	// Group by authoring reply, preserving first-seen reply order.
	byReply := make(map[int][]targetRow)
	order := make([]int, 0)
	for _, r := range rows {
		if _, ok := byReply[r.ReplyID]; !ok {
			order = append(order, r.ReplyID)
		}
		byReply[r.ReplyID] = append(byReply[r.ReplyID], r)
	}
	if limit > 0 && limit < len(order) {
		order = order[:limit]
	}

	slog.Info("待迁移统计",
		"authoring_replies", len(byReply),
		"target_rows", len(rows),
		"processing", len(order),
		"dry_run", dryRun)

	// 2. Fetch the authoring replies' current bodies.
	type replyContent struct {
		ID      int
		Content string
	}
	var contents []replyContent
	if err := db.Table("topic_reply").
		Select("id, content").
		Where("id IN ?", order).
		Scan(&contents).Error; err != nil {
		return fmt.Errorf("加载 topic_reply 内容失败: %w", err)
	}
	contentByID := make(map[int]string, len(contents))
	for _, c := range contents {
		contentByID[c.ID] = c.Content
	}

	// 3a. Dry-run: print a sample + exit without writing.
	if dryRun {
		for i, replyID := range order {
			if i >= 3 {
				break
			}
			out := buildMigratedContent(contentByID[replyID], byReply[replyID])
			slog.Info("样例（不会写入）",
				"reply_id", replyID,
				"targets", len(byReply[replyID]),
				"before", truncate(contentByID[replyID], 200),
				"after", truncate(out, 400))
		}
		slog.Info("dry-run 结束：未写入任何数据。确认无误后用 -dry-run=false 执行")
		return nil
	}

	// 3b. Apply: snapshot first, then migrate each reply in its own tx.
	if err := backup(db); err != nil {
		return fmt.Errorf("备份失败（已中止，未做任何修改）: %w", err)
	}

	migrated := 0
	for _, replyID := range order {
		newContent := buildMigratedContent(contentByID[replyID], byReply[replyID])
		if err := db.Transaction(func(tx *gorm.DB) error {
			// UpdateColumn: skip GORM's auto `updated` bump / hooks — a data move
			// must not look like a user edit.
			if err := tx.Model(&model.TopicReply{}).
				Where("id = ?", replyID).
				UpdateColumn("content", newContent).Error; err != nil {
				return err
			}
			return tx.Where("reply_id = ?", replyID).
				Delete(&model.TopicReplyTarget{}).Error
		}); err != nil {
			return fmt.Errorf("迁移 reply %d 失败: %w", replyID, err)
		}
		migrated++
		if migrated%200 == 0 {
			slog.Info("进度", "migrated", migrated, "total", len(order))
		}
	}

	slog.Info("迁移完成",
		"migrated_replies", migrated,
		"note", "topic_reply_target 已清空；原始数据在 topic_reply_target_backup，确认无误后可手动 DROP")
	return nil
}

// backup snapshots the source tables before mutation. CREATE TABLE IF NOT EXISTS
// means a re-run keeps the FIRST (pre-migration) snapshot — that's the rollback.
func backup(db *gorm.DB) error {
	if err := db.Exec(
		`CREATE TABLE IF NOT EXISTS topic_reply_target_backup AS TABLE topic_reply_target`,
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`CREATE TABLE IF NOT EXISTS topic_reply_content_backup AS
		 SELECT id, content FROM topic_reply
		 WHERE id IN (SELECT DISTINCT reply_id FROM topic_reply_target)`,
	).Error; err != nil {
		return err
	}
	slog.Info("备份完成",
		"tables", "topic_reply_target_backup, topic_reply_content_backup")
	return nil
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
