package service

import (
	"context"
	"io"

	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/imageclient"
)

// galgame_upload.go — U2 cover/screenshot upload path.
//
// Distinct from UploadTopicImage (legacy S3, no hash, no image_service):
// galgame covers/screenshots reference image_service by `image_hash`,
// which means we MUST upload through image_service first to obtain the
// hash before the user can submit a PUT/PR carrying that hash.
//
// Allowed presets are constrained at the handler boundary; this layer
// trusts its caller. image_service itself also gates presets against
// the calling site's `image_allowed_presets`, so a wrong preset returns
// a 4xx from upstream rather than silently mis-storing.
//
// Quota / size / format validation lives in image_service (see
// docs/image_service/03-api-design.md §1). We only forward + classify.

// UploadGalgameResult is the slim response surface the FE consumes for a
// new cover/screenshot row: the hash to attach to the galgame wire +
// the CDN URL it can render immediately (saves a hash→URL round-trip).
type UploadGalgameResult struct {
	Hash         string            `json:"hash"`
	URL          string            `json:"url"`
	Width        int               `json:"width"`
	Height       int               `json:"height"`
	SizeBytes    int64             `json:"sizeBytes"`
	VariantURLs  map[string]string `json:"variantUrls,omitempty"`
	Deduplicated bool              `json:"deduplicated"`
}

// UploadGalgameImage proxies a single file to image_service under the
// caller-chosen preset and adapts the response to UploadGalgameResult.
// userID is taken from the kungal session (caller passes it for logging
// + future per-user quota mirroring).
func (s *ImageService) UploadGalgameImage(
	ctx context.Context,
	userID int,
	r io.Reader,
	filename, preset string,
) (*UploadGalgameResult, *errors.AppError) {
	if s.imgCli == nil {
		// image_service credentials not configured. Surface a clear
		// error instead of crashing — matches the wiki side's behaviour
		// (see kun-oauth-admin's mapWriteBodyError fallback).
		return nil, errors.ErrBadRequest(
			"图片上传服务未配置 (KUN_IMAGE_CLIENT_ID / KUN_IMAGE_CLIENT_SECRET)",
		)
	}

	res, err := s.imgCli.Upload(ctx, r, filename, preset)
	if err != nil {
		// Forward image_service's own status + code if it returned a
		// structured Error; otherwise wrap as generic.
		if ie, ok := err.(*imageclient.Error); ok {
			return nil, errors.New(ie.Code, ie.Message, ie.StatusCode)
		}
		return nil, errors.ErrInternal("图片上传失败")
	}
	return &UploadGalgameResult{
		Hash:         res.Hash,
		URL:          res.URL,
		Width:        res.Width,
		Height:       res.Height,
		SizeBytes:    res.SizeBytes,
		VariantURLs:  res.VariantURLs,
		Deduplicated: res.Deduplicated,
	}, nil
}
