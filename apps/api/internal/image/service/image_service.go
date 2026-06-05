package service

import (
	"context"
	"io"

	"kun-galgame-api/internal/image/repository"
	"kun-galgame-api/internal/infrastructure/storage"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/imageclient"
)

const (
	MaxImageSize    = 10 * 1024 * 1024 // 10MB
	dailyImageLimit = 50
)

type ImageService struct {
	repo *repository.ImageRepository
	// s3 is no longer used for image uploads — ALL uploads go through
	// image_service (imgCli) now. Kept only so the constructor wiring is
	// unchanged; safe to drop with its app.go wiring in a later cleanup.
	s3 *storage.S3Client
	// imgCli is the image_service client. ALL user image uploads — topic inline
	// images AND galgame covers/screenshots — go through it. Nil-able: when the
	// credentials (KUN_IMAGE_CLIENT_ID / KUN_IMAGE_CLIENT_SECRET) are unset, both
	// upload paths return a clear "未配置" error instead of falling back to S3.
	imgCli *imageclient.Client
}

func NewImageService(
	repo *repository.ImageRepository,
	s3 *storage.S3Client,
	imgCli *imageclient.Client,
) *ImageService {
	return &ImageService{repo: repo, s3: s3, imgCli: imgCli}
}

// UploadTopicImage routes a user's inline post image through image_service
// under the `topic` preset (WebP q77, ≤1920×1080, EXIF stripped — see infra
// configs/image_presets.yaml) and returns the full webp CDN URL, which the
// editor inserts directly as the image src.
//
// The kungal-local per-USER daily quota is kept on purpose: image_service's
// quota is per-SITE, so this still gives per-user fair-use limiting + a
// friendly message before we even hit image_service.
func (s *ImageService) UploadTopicImage(ctx context.Context, userID int, r io.Reader, filename string) (string, *errors.AppError) {
	if s.imgCli == nil {
		return "", errors.ErrBadRequest(
			"图片上传服务未配置 (KUN_IMAGE_CLIENT_ID / KUN_IMAGE_CLIENT_SECRET)",
		)
	}

	count, err := s.repo.GetDailyCount(userID)
	if err != nil {
		return "", errors.ErrInternal("查询用户失败")
	}
	if count >= dailyImageLimit {
		return "", errors.ErrBadRequest("今日图片上传次数已达上限")
	}

	res, uErr := s.imgCli.Upload(ctx, r, filename, "topic")
	if uErr != nil {
		// Forward image_service's structured error (preset denied, MIME, quota,
		// oversize, …) so the user sees the real reason; else generic.
		if ie, ok := uErr.(*imageclient.Error); ok {
			return "", errors.New(ie.Code, ie.Message, ie.StatusCode)
		}
		return "", errors.ErrInternal("图片上传失败")
	}

	s.repo.IncrementDailyCount(userID)
	return res.URL, nil
}
