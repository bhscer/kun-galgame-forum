package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Backfill deep-links into OLD notification rows. Before the deep-link feature
// every reply/comment notification stored link = "/topic/<id>" (the topic root),
// so clicking an old notification couldn't jump to the actual post. We never
// recorded the target reply/comment id on the message — but the notification is
// written in the SAME transaction as the reply/comment, so notification.created
// ≈ the target's created. That lets us recover the target heuristically:
//
//   - replied   → the sender's reply on that topic, closest in time      → ?reply=<floor>
//   - commented → the sender's comment on that topic, closest in time     → ?comment=<id>
//   - mentioned → the sender's reply (else comment) on that topic         → ?reply / ?comment
//   - solution  → the topic's CURRENT best-answer reply                   → ?reply=<floor>
//   - pin-reply → the topic's CURRENT pinned reply                        → ?reply=<floor>
//
// Best-effort: a target deleted since (or a best-answer/pin changed since) won't
// match and the row keeps its /topic/<id> link — still valid, just not deep.
//
// Safe + idempotent: touches ONLY rows whose link is still the bare "/topic/<id>"
// (regex-anchored on `$`), so it never double-appends and re-running is a no-op
// for rows already deep-linked (by this tool or the new write path). Topic-level
// notifications (liked / favorite / upvoted) are intentionally left untouched —
// they target the topic, not a post.
//
//	go run ./cmd/backfill-message-links          # dry-run — counts only
//	go run ./cmd/backfill-message-links --apply  # write the deep-links
func main() {
	apply := flag.Bool("apply", false, "write the changes (default: dry-run, counts only)")
	flag.Parse()

	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(os.Getenv("KUN_DATABASE_URL")),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		fmt.Println("连接数据库失败:", err)
		os.Exit(1)
	}

	// A reply / comment by the notification's SENDER on the LINKED topic, closest
	// in time to the notification. Same-tx → created ≈ created (sub-second); the
	// ±60s window only guards clock/storage granularity, and ORDER BY time-
	// proximity picks the triggering post when a user authored several there.
	const replyMatch = `(SELECT r.floor FROM topic_reply r
		WHERE r.user_id = m.sender_id
		  AND r.topic_id = substring(m.link from '^/topic/([0-9]+)$')::int
		  AND r.created BETWEEN m.created - interval '60 seconds' AND m.created + interval '60 seconds'
		ORDER BY abs(extract(epoch FROM (r.created - m.created))) LIMIT 1)`
	const commentMatch = `(SELECT c.id FROM topic_comment c
		WHERE c.user_id = m.sender_id
		  AND c.topic_id = substring(m.link from '^/topic/([0-9]+)$')::int
		  AND c.created BETWEEN m.created - interval '60 seconds' AND m.created + interval '60 seconds'
		ORDER BY abs(extract(epoch FROM (c.created - m.created))) LIMIT 1)`

	// Only bare "/topic/<id>" links are candidates (anchored on `$` → never a
	// link that already carries ?reply=/?comment=).
	const bare = `m.link ~ '^/topic/[0-9]+$'`

	steps := []struct {
		label    string
		countSQL string
		applySQL string
	}{
		{
			"replied→reply",
			`SELECT count(*) FROM message m WHERE m.type='replied' AND ` + bare + ` AND ` + replyMatch + ` IS NOT NULL`,
			`UPDATE message m SET link = m.link || '?reply=' || ` + replyMatch + ` WHERE m.type='replied' AND ` + bare + ` AND ` + replyMatch + ` IS NOT NULL`,
		},
		{
			"commented→comment",
			`SELECT count(*) FROM message m WHERE m.type='commented' AND ` + bare + ` AND ` + commentMatch + ` IS NOT NULL`,
			`UPDATE message m SET link = m.link || '?comment=' || ` + commentMatch + ` WHERE m.type='commented' AND ` + bare + ` AND ` + commentMatch + ` IS NOT NULL`,
		},
		{
			"mentioned→reply",
			`SELECT count(*) FROM message m WHERE m.type='mentioned' AND ` + bare + ` AND ` + replyMatch + ` IS NOT NULL`,
			`UPDATE message m SET link = m.link || '?reply=' || ` + replyMatch + ` WHERE m.type='mentioned' AND ` + bare + ` AND ` + replyMatch + ` IS NOT NULL`,
		},
		{
			// Comment @mentions: only the ones NOT already resolved as a reply mention
			// above (so the two passes are disjoint — accurate dry-run + order-free).
			"mentioned→comment",
			`SELECT count(*) FROM message m WHERE m.type='mentioned' AND ` + bare + ` AND ` + replyMatch + ` IS NULL AND ` + commentMatch + ` IS NOT NULL`,
			`UPDATE message m SET link = m.link || '?comment=' || ` + commentMatch + ` WHERE m.type='mentioned' AND ` + bare + ` AND ` + replyMatch + ` IS NULL AND ` + commentMatch + ` IS NOT NULL`,
		},
		{
			// best_answer_id has OnDelete:SET NULL, so NOT NULL ⇒ the reply still exists.
			"solution→best-answer",
			`SELECT count(*) FROM message m JOIN topic t ON t.id = substring(m.link from '^/topic/([0-9]+)$')::int WHERE m.type='solution' AND ` + bare + ` AND t.best_answer_id IS NOT NULL`,
			`UPDATE message m SET link = m.link || '?reply=' || r.floor FROM topic t JOIN topic_reply r ON r.id = t.best_answer_id WHERE m.type='solution' AND ` + bare + ` AND t.id = substring(m.link from '^/topic/([0-9]+)$')::int AND t.best_answer_id IS NOT NULL`,
		},
		{
			"pin-reply→pinned",
			`SELECT count(*) FROM message m JOIN topic t ON t.id = substring(m.link from '^/topic/([0-9]+)$')::int WHERE m.type='pin-reply' AND ` + bare + ` AND t.pinned_reply_id IS NOT NULL`,
			`UPDATE message m SET link = m.link || '?reply=' || r.floor FROM topic t JOIN topic_reply r ON r.id = t.pinned_reply_id WHERE m.type='pin-reply' AND ` + bare + ` AND t.id = substring(m.link from '^/topic/([0-9]+)$')::int AND t.pinned_reply_id IS NOT NULL`,
		},
	}

	if *apply {
		fmt.Println("回填消息深链接（--apply 实际写入）:")
	} else {
		fmt.Println("回填消息深链接（dry-run，仅统计；加 --apply 实际写入）:")
	}

	var total int64
	for _, s := range steps {
		if *apply {
			res := db.Exec(s.applySQL)
			if res.Error != nil {
				fmt.Printf("  %-22s 失败: %v\n", s.label, res.Error)
				os.Exit(1)
			}
			fmt.Printf("  %-22s 回填 %d 条\n", s.label, res.RowsAffected)
			total += res.RowsAffected
		} else {
			var n int64
			if err := db.Raw(s.countSQL).Scan(&n).Error; err != nil {
				fmt.Printf("  %-22s 失败: %v\n", s.label, err)
				os.Exit(1)
			}
			fmt.Printf("  %-22s 可回填 %d 条\n", s.label, n)
			total += n
		}
	}

	// Reply/comment-ish notifications still on a bare topic link (target deleted,
	// or best-answer/pin changed) — not recoverable, kept as /topic/<id>.
	var leftover int64
	db.Raw(`SELECT count(*) FROM message m WHERE ` + bare +
		` AND m.type IN ('replied','commented','mentioned','solution','pin-reply')`).Scan(&leftover)

	if *apply {
		fmt.Printf("合计回填 %d 条；仍为话题根链接 %d 条（目标已删除/已变更，无法恢复）。\n", total, leftover)
	} else {
		fmt.Printf("合计可回填 %d 条；执行后预计仍为话题根链接 %d 条。\n", total, leftover-total)
	}
}
