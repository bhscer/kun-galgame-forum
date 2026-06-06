package handler

import (
	"kun-galgame-api/internal/image/service"
	"kun-galgame-api/internal/middleware"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type ImageHandler struct {
	imageService *service.ImageService
}

func NewImageHandler(imageService *service.ImageService) *ImageHandler {
	return &ImageHandler{imageService: imageService}
}

// allowedGalgamePresets restricts which image_service presets the galgame
// upload proxy can request. Keeps the proxy from doubling as a generic
// image-service tunnel; presets here MUST also be in this site's
// image_allowed_presets on the image_service side.
//
// AUDIT FIX: wiki's image_presets.yaml only registers `galgame_banner`
// (no `galgame_screenshot` preset). Both cover and screenshot uploads
// flow through the same preset — image_service's main pipeline (fit
// 1920x1080 webp@77) is suitable for both. The mini variant (460x259)
// generated alongside is unused on screenshots; wasteful but not
// broken. If wiki later registers a dedicated screenshot preset, add
// it here and route Screenshots.vue to it; until then a single preset
// keeps the contract honest.
var allowedGalgamePresets = map[string]struct{}{
	"galgame_banner": {}, // cover (sort_order=0 pinned) + screenshot rows
}

// UploadGalgameImage handles cover/screenshot upload (U2). Multipart form:
//   - file:   image binary (required)
//   - preset: "galgame_banner" (required; only registered galgame
//             preset on wiki — see allowedGalgamePresets above)
//
// Returns the image_service {hash, url, ...} payload so the FE can
// immediately add a new cover/screenshot row referencing the hash and
// submit it on the next PUT /galgame/:gid or POST /galgame/:gid/prs
// (presence-replace arrays — see GalgameEditStoreTemp note).
//
// POST /api/image/galgame
func (h *ImageHandler) UploadGalgameImage(c *fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	preset := c.FormValue("preset")
	if _, ok := allowedGalgamePresets[preset]; !ok {
		return response.Error(c, errors.ErrBadRequest(
			"preset 必须为 galgame_banner"))
	}

	file, err := c.FormFile("file")
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("请选择要上传的图片"))
	}
	if file.Size > service.MaxImageSize {
		return response.Error(c, errors.ErrBadRequest("图片大小不能超过 10MB"))
	}

	f, err := file.Open()
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("读取图片失败"))
	}
	defer f.Close()

	res, sErr := h.imageService.UploadGalgameImage(
		c.Context(), user.ID, f, file.Filename, preset,
	)
	if sErr != nil {
		return response.Error(c, sErr)
	}
	return response.OK(c, res)
}

// UploadTopicImage handles topic image upload.
// POST /api/image/topic
func (h *ImageHandler) UploadTopicImage(c *fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	file, err := c.FormFile("image")
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("请选择要上传的图片"))
	}
	if file.Size > service.MaxImageSize {
		return response.Error(c, errors.ErrBadRequest("图片大小不能超过 10MB"))
	}

	f, err := file.Open()
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("读取图片失败"))
	}
	defer f.Close()

	key, sErr := h.imageService.UploadTopicImage(c.Context(), user.ID, f, file.Filename)
	if sErr != nil {
		return response.Error(c, sErr)
	}
	return response.OK(c, key)
}

// POST /api/image/message
// Uploads a chat / private-message inline image and returns its CDN URL, which
// the client inserts into the message as `![name](url)`. Mirror of
// UploadTopicImage but under the `message` preset.
func (h *ImageHandler) UploadMessageImage(c *fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	file, err := c.FormFile("image")
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("请选择要上传的图片"))
	}
	if file.Size > service.MaxImageSize {
		return response.Error(c, errors.ErrBadRequest("图片大小不能超过 10MB"))
	}

	f, err := file.Open()
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("读取图片失败"))
	}
	defer f.Close()

	key, sErr := h.imageService.UploadMessageImage(c.Context(), user.ID, f, file.Filename)
	if sErr != nil {
		return response.Error(c, sErr)
	}
	return response.OK(c, key)
}
