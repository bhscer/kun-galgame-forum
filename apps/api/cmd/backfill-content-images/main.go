// Backfill legacy content images: rehost every `https://image.kungal.com/...`
// image embedded in user content onto image_service, then rewrite the stored
// content to the new content-addressed CDN URL.
//
// Background: before image_service, kungal's image uploader put files at the
// PATH-based legacy host `image.kungal.com/topic/user_<id>/<name>-<ts>.webp`.
// image_service (image.kungal.iloveren.link/{aa}/{bb}/{hash}.webp) replaced it,
// but old posts still reference the legacy host. This command migrates those.
// (Avatars are NOT here — OAuth owns them; kungal doesn't store avatars.)
//
// Scope (content columns that can embed `image.kungal.com` URLs):
//   topic.content · topic_reply.content · chat_message.content · galgame_comment.content
// galgame cover/screenshot already moved to image_service (image_hash); doc
// banners are repo static assets; website icons are external favicons — none here.
//
// Per distinct old URL: HTTP-fetch the original from `-base` (default the public
// legacy host) → upload to image_service under the `topic` preset → cache the
// new URL → string-replace every occurrence in each row's content. The same
// image referenced by many rows is uploaded once (cross-table cache). A URL that
// 404s / fails to fetch is logged and SKIPPED (its rows keep the old URL — safe,
// and a re-run retries). Content is the ONLY column written (we deliberately do
// NOT touch `updated`, so topics aren't bounced to the top of "recently updated").
//
// SAFE BY DEFAULT: -dry-run defaults to TRUE (reports distinct URLs + per-table
// counts, no network writes, no DB writes). Pass -dry-run=false to actually run.
// Idempotent: once rewritten, content no longer matches `%image.kungal.com%`.
//
//	docker compose -f docker-compose.prod.yml --profile jobs run --rm tools \
//	  backfill-content-images                       # dry-run: report only
//	  backfill-content-images -dry-run=false        # actually rehost + rewrite
//	  backfill-content-images -dry-run=false -limit=20   # smoke-test 20 rows/table
//	  backfill-content-images -base=http://legacy-image:80   # fetch originals elsewhere
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"kun-galgame-api/internal/infrastructure/database"
	"kun-galgame-api/pkg/config"
	"kun-galgame-api/pkg/imageclient"
	"kun-galgame-api/pkg/logger"

	"github.com/joho/godotenv"
)

// Matches a legacy image URL up to the first markdown/whitespace/quote delimiter.
// Filenames may contain non-ASCII (e.g. 茅羽耶-...webp), which these bytes allow.
var legacyImageRe = regexp.MustCompile(`https://image\.kungal\.com/[^\s)"'>\]\\]+`)

// table + its content column to scan/rewrite.
type target struct {
	table string
	col   string
}

var targets = []target{
	{"topic", "content"},
	{"topic_reply", "content"},
	{"chat_message", "content"},
	{"galgame_comment", "content"},
}

func main() {
	_ = godotenv.Load()

	dryRun := flag.Bool("dry-run", true, "TRUE (default): report only, no fetch/upload/DB writes. Pass -dry-run=false to apply.")
	base := flag.String("base", "https://image.kungal.com", "Base to HTTP-fetch legacy originals from (override if the old host isn't reachable from here)")
	limit := flag.Int("limit", 0, "Max rows per table (0 = all); for smoke-testing -dry-run=false on a small batch")
	preset := flag.String("preset", "topic", "image_service preset to rehost under")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("加载配置失败", "error", err)
		os.Exit(1)
	}
	logger.Init(cfg.Server.Mode)

	if !*dryRun && (cfg.ImageClient.ClientID == "" || cfg.ImageClient.ClientSecret == "") {
		slog.Error("image_service 未配置 (KUN_IMAGE_CLIENT_ID / KUN_IMAGE_CLIENT_SECRET), 无法上传")
		os.Exit(1)
	}
	imgCli := imageclient.New(imageclient.Config{
		BaseURL:      cfg.ImageClient.BaseURL,
		CDNBase:      cfg.GalgameWiki.ImageCDNBase,
		ClientID:     cfg.ImageClient.ClientID,
		ClientSecret: cfg.ImageClient.ClientSecret,
	})
	db := database.NewPostgres(cfg.Database, cfg.Server.Mode)

	httpClient := &http.Client{Timeout: 60 * time.Second}
	ctx := context.Background()

	migrated := map[string]string{} // oldURL -> new image_service URL (uploaded)
	dead := map[string]bool{}       // oldURL that failed to fetch/upload (skipped)
	seen := map[string]bool{}       // distinct oldURL seen (dry-run accounting)

	var rowsUpdated, rowsSkipped int

	slog.Info("开始迁移 content 老图", "dry_run", *dryRun, "base", *base, "preset", *preset, "limit", *limit)

	for _, t := range targets {
		type row struct {
			ID      int64
			Content string
		}
		var rows []row
		q := db.Table(t.table).
			Select("id, "+t.col+" AS content").
			Where(t.col+" LIKE ?", "%image.kungal.com%").
			Order("id ASC")
		if *limit > 0 {
			q = q.Limit(*limit)
		}
		if err := q.Scan(&rows).Error; err != nil {
			slog.Error("扫描失败", "table", t.table, "error", err)
			os.Exit(1)
		}
		slog.Info("扫描完成", "table", t.table, "含老图行数", len(rows))

		for _, r := range rows {
			urls := dedupe(legacyImageRe.FindAllString(r.Content, -1))
			newContent := r.Content
			changed := false

			for _, old := range urls {
				seen[old] = true
				if dead[old] {
					continue
				}
				newURL, done := migrated[old]
				if !done {
					body, ferr := fetch(ctx, httpClient, rewriteHost(old, *base))
					if ferr != nil {
						slog.Warn("抓取原图失败, 跳过(保留旧 URL)", "url", old, "error", ferr)
						dead[old] = true
						continue
					}
					if *dryRun {
						// reachable; in dry-run we don't upload or rewrite.
						continue
					}
					res, uerr := imgCli.Upload(ctx, bytes.NewReader(body), path.Base(old), *preset)
					if uerr != nil {
						slog.Error("上传 image_service 失败, 跳过", "url", old, "error", uerr)
						dead[old] = true
						continue
					}
					newURL = res.URL
					migrated[old] = newURL
					slog.Info("已重托管", "old", old, "new", newURL)
				}
				if !*dryRun && newURL != "" && newURL != old {
					newContent = strings.ReplaceAll(newContent, old, newURL)
					changed = true
				}
			}

			if *dryRun {
				continue
			}
			if !changed {
				rowsSkipped++
				continue
			}
			if err := db.Exec(
				"UPDATE "+t.table+" SET "+t.col+" = ? WHERE id = ?",
				newContent, r.ID,
			).Error; err != nil {
				slog.Error("更新行失败", "table", t.table, "id", r.ID, "error", err)
				rowsSkipped++
				continue
			}
			rowsUpdated++
		}
	}

	if *dryRun {
		reachable := len(seen) - len(dead)
		fmt.Printf("dry-run 完成: 发现去重老图 %d 张 (可抓取 %d, 抓取失败/404 %d)。重跑加 -dry-run=false 执行。\n",
			len(seen), reachable, len(dead))
	} else {
		fmt.Printf("迁移完成: 重托管 %d 张老图, 改写 %d 行, 跳过 %d 行, 失败/404 老图 %d 张(保留旧 URL)。\n",
			len(migrated), rowsUpdated, rowsSkipped, len(dead))
	}
}

// rewriteHost swaps the legacy host for -base so originals can be fetched from
// an internal mirror when image.kungal.com isn't reachable from the job.
func rewriteHost(url, base string) string {
	if base == "" || base == "https://image.kungal.com" {
		return url
	}
	return strings.Replace(url, "https://image.kungal.com", strings.TrimRight(base, "/"), 1)
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

func fetch(ctx context.Context, c *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 32<<20)) // 32MB guard
}
