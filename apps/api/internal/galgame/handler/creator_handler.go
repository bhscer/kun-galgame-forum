package handler

import (
	"encoding/json"

	"kun-galgame-api/internal/galgame/service"
	"kun-galgame-api/internal/middleware"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/response"

	"github.com/gofiber/fiber/v2"
)

// CreatorHandler exposes the forum-side creator-role application endpoints:
// eligibility/status (read) and apply (which checks the forum's eligibility
// gate, then files the application on the central OAuth queue). Role grant +
// admin review live in OAuth. See docs/auth/01-creator-role-design.md.
type CreatorHandler struct {
	svc *service.CreatorService
}

func NewCreatorHandler(svc *service.CreatorService) *CreatorHandler {
	return &CreatorHandler{svc: svc}
}

// Status — GET /api/user/creator/status: eligibility snapshot + current application.
func (h *CreatorHandler) Status(c *fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	token := middleware.GetAccessToken(c)
	if token == "" {
		return response.Error(c, errors.ErrAuthExpired())
	}
	elig, app, appErr := h.svc.Status(c.Context(), user.ID, token)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, fiber.Map{"eligibility": elig, "application": app})
}

// Apply — POST /api/user/creator/apply {message?}.
func (h *CreatorHandler) Apply(c *fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	token := middleware.GetAccessToken(c)
	if token == "" {
		return response.Error(c, errors.ErrAuthExpired())
	}
	var body struct {
		Message string `json:"message"`
	}
	if len(c.Body()) > 0 {
		_ = json.Unmarshal(c.Body(), &body)
	}
	app, appErr := h.svc.Apply(c.Context(), user.ID, token, body.Message)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, app)
}
