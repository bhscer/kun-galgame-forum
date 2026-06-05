// Backfill friend_link.banner: re-upload the legacy static /friends/<name>.webp
// banners through image_service and replace the column with the returned CDN URL,
// so every friend link's image matches what the admin form now produces (the
// `topic` preset — WebP q77, EXIF stripped — via /image/topic).
//
// The static files live in the WEB container's public/ dir (not reachable from
// this Go process), so each is fetched over HTTP from the public site and the
// bytes are pushed to image_service. Run this from inside the cluster (the tools
// container reaches both postgres and the internal image_service at
// http://image:9278); fetching uses the PUBLIC base (default www.kungal.com),
// which Traefik routes back to the web container.
//
// Idempotent: only rows whose banner starts with "/friends/" are touched. Once
// migrated the banner is a CDN URL, so a re-run skips it. Per-row failures are
// logged and skipped (the row keeps its old static banner), never aborting.
//
//	docker compose -f docker-compose.prod.yml --profile jobs run --rm tools \
//	  backfill-friend-link-banners                 # do it (topic preset)
//	  backfill-friend-link-banners --dry-run        # fetch + report, no writes
//	  backfill-friend-link-banners --preset=galgame_banner --base=https://www.kungal.com
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
	"strings"
	"time"

	"kun-galgame-api/internal/infrastructure/database"
	"kun-galgame-api/pkg/config"
	"kun-galgame-api/pkg/imageclient"
	"kun-galgame-api/pkg/logger"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	dryRun := flag.Bool("dry-run", false, "Fetch + report planned changes but do not upload or write")
	base := flag.String("base", "https://www.kungal.com", "Public base URL the legacy /friends/<name>.webp banners are served from")
	preset := flag.String("preset", "topic", "image_service preset to upload under (topic | galgame_banner)")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("加载配置失败", "error", err)
		os.Exit(1)
	}
	logger.Init(cfg.Server.Mode)

	if cfg.ImageClient.ClientID == "" || cfg.ImageClient.ClientSecret == "" {
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
	baseURL := strings.TrimRight(*base, "/")

	type row struct {
		ID     int
		Name   string
		Banner string
	}
	var rows []row
	// Only legacy static banners; CDN URLs / empty / external are left as-is.
	if err := db.Table("friend_link").
		Select("id, name, banner").
		Where("banner LIKE ?", "/friends/%").
		Order("id ASC").
		Scan(&rows).Error; err != nil {
		slog.Error("查询友链失败", "error", err)
		os.Exit(1)
	}

	slog.Info("开始回填友链 banner",
		"candidates", len(rows), "base", baseURL, "preset", *preset, "dry_run", *dryRun)

	httpClient := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()
	updated, failed := 0, 0

	for _, r := range rows {
		src := baseURL + r.Banner // e.g. https://www.kungal.com/friends/acgngame.webp
		fname := r.Banner[strings.LastIndex(r.Banner, "/")+1:]

		body, ferr := fetch(ctx, httpClient, src)
		if ferr != nil {
			slog.Error("拉取静态 banner 失败, 跳过", "id", r.ID, "name", r.Name, "src", src, "error", ferr)
			failed++
			continue
		}

		if *dryRun {
			slog.Info("dry-run: 将上传", "id", r.ID, "name", r.Name, "src", src, "bytes", len(body))
			continue
		}

		res, uerr := imgCli.Upload(ctx, bytes.NewReader(body), fname, *preset)
		if uerr != nil {
			slog.Error("上传 image_service 失败, 跳过", "id", r.ID, "name", r.Name, "error", uerr)
			failed++
			continue
		}

		if err := db.Exec(
			"UPDATE friend_link SET banner = ?, updated = now() WHERE id = ?",
			res.URL, r.ID,
		).Error; err != nil {
			slog.Error("更新 banner 失败", "id", r.ID, "name", r.Name, "new", res.URL, "error", err)
			failed++
			continue
		}

		slog.Info("已迁移", "id", r.ID, "name", r.Name, "old", r.Banner, "new", res.URL)
		updated++
	}

	if *dryRun {
		fmt.Printf("dry-run 完成: %d 个候选 (banner LIKE '/friends/%%'), %d 个拉取失败\n", len(rows), failed)
	} else {
		fmt.Printf("回填完成: 成功迁移 %d, 失败/跳过 %d, 候选共 %d\n", updated, failed, len(rows))
	}
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
	return io.ReadAll(resp.Body)
}
