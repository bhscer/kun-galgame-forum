package cron

import (
	"context"
	"log/slog"
	"regexp"

	"kun-galgame-api/pkg/imageclient"

	"gorm.io/gorm"
)

// contentImageHashRe extracts the 64-hex hash from a /image/<hash> content token.
var contentImageHashRe = regexp.MustCompile(`/image/([0-9a-f]{64})`)

// refPingBatchSize caps each /image/reference-ping call (image_service accepts
// ≤1000 hashes per batch — docs/image_service/03-api-design.md).
const refPingBatchSize = 1000

// RunReferencePing refreshes last_referenced_at for every image_service hash the
// forum references in content, so image-gc never reclaims a live content image
// (TTL-driven: a hash >365d without a ping is soft-deleted, +30d physically —
// docs/image_service/06-integration-guide.md §保活).
//
// Scope = the four content tables' `/image/<hash>` tokens ONLY. Forum stores no
// local *_image_hash columns: avatars are OAuth's (site=account, pinged there);
// galgame cover/screenshot hashes belong to wiki and are pinged by infra's
// galgame-image-refping. The forum image client is site=kungal — the SAME site
// these content images were uploaded under — so the site-scoped pings hit (no
// risk of the wrong-client mismatch that bit the galgame path).
//
// Returns (distinct hashes seen, hashes image_service updated). A failed batch is
// logged and the remaining batches still run; the first error is returned so a
// one-off cmd can exit non-zero.
func RunReferencePing(
	ctx context.Context, db *gorm.DB, imgCli *imageclient.Client,
) (distinct int, updated int64, err error) {
	if imgCli == nil {
		return 0, 0, nil
	}

	// Pull every content row carrying a token across all token-bearing columns,
	// then extract + dedupe the hashes in memory (≈5.7k distinct — trivially
	// small). MUST cover the same columns as backfill-content-images' targets, so
	// every column we tokenize also gets kept alive (the snapshot columns —
	// message / topic_reply_target — and the doc/toolset bodies were added 2026-06-16).
	type contentRow struct{ Content string }
	var rows []contentRow
	if e := db.WithContext(ctx).Raw(`
		SELECT content          FROM topic              WHERE content          LIKE '%/image/%'
		UNION ALL SELECT content          FROM topic_reply        WHERE content          LIKE '%/image/%'
		UNION ALL SELECT content          FROM chat_message       WHERE content          LIKE '%/image/%'
		UNION ALL SELECT content          FROM galgame_comment    WHERE content          LIKE '%/image/%'
		UNION ALL SELECT content_markdown FROM doc_article        WHERE content_markdown LIKE '%/image/%'
		UNION ALL SELECT description      FROM galgame_toolset    WHERE description      LIKE '%/image/%'
		UNION ALL SELECT content          FROM message            WHERE content          LIKE '%/image/%'
		UNION ALL SELECT content          FROM topic_reply_target WHERE content          LIKE '%/image/%'
	`).Scan(&rows).Error; e != nil {
		return 0, 0, e
	}

	contents := make([]string, len(rows))
	for i, r := range rows {
		contents[i] = r.Content
	}
	hashes := extractContentImageHashes(contents)

	for start := 0; start < len(hashes); start += refPingBatchSize {
		end := min(start+refPingBatchSize, len(hashes))
		res, e := imgCli.ReferencePing(ctx, hashes[start:end])
		if e != nil {
			slog.Error("内容图 reference-ping 批次失败", "from", start, "to", end, "error", e)
			if err == nil {
				err = e
			}
			continue
		}
		updated += res.Updated
	}
	return len(hashes), updated, err
}

// extractContentImageHashes pulls every distinct /image/<hash> token hash out of
// a set of content strings. A row may carry several images (and the same image
// may repeat across rows), so the result is de-duplicated.
func extractContentImageHashes(contents []string) []string {
	seen := make(map[string]struct{})
	for _, c := range contents {
		for _, m := range contentImageHashRe.FindAllStringSubmatch(c, -1) {
			seen[m[1]] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for h := range seen {
		out = append(out, h)
	}
	return out
}
