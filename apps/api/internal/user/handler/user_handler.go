package handler

import (
	"strconv"

	"kun-galgame-api/internal/middleware"
	"kun-galgame-api/internal/user/dto"
	"kun-galgame-api/internal/user/service"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/response"
	"kun-galgame-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// UserHandler exposes the kungal-side user-facing endpoints. After the
// OAuth-as-truth migration, kungal no longer brokers identity changes
// (bio / username / email / avatar / ban / delete) — those live in the
// OAuth admin UI. The endpoints here only deal with kungal-specific
// state (check-in / moemoepoint / unread counts) and content listings
// keyed by user_id.
type UserHandler struct {
	userService        *service.UserService
	userContentService *service.UserContentService
}

func NewUserHandler(
	userService *service.UserService,
	userContentService *service.UserContentService,
) *UserHandler {
	return &UserHandler{
		userService:        userService,
		userContentService: userContentService,
	}
}

// GetProfile returns a user's public profile (identity from OAuth, stats
// from kungal local).
// GET /api/user/:userID
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("无效的用户 ID"))
	}
	profile, appErr := h.userService.GetUserProfile(c.Context(), userID)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, profile)
}

// CheckIn handles daily check-in.
// POST /api/user/check-in
func (h *UserHandler) CheckIn(c *fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	points, appErr := h.userService.CheckIn(c.Context(), user.ID)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, points)
}

// GetStatus returns the user's status (moemoepoints, check-in, unread messages).
// GET /api/user/status
func (h *UserHandler) GetStatus(c *fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	status, appErr := h.userService.GetUserStatus(c.Context(), user.ID)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, status)
}

// GetFloatingCard returns lightweight user info for hover card.
// GET /api/user/:userID/floating — target user is read from ?userId=N (legacy).
func (h *UserHandler) GetFloatingCard(c *fiber.Ctx) error {
	var req dto.FloatingCardRequest
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}
	card, appErr := h.userService.GetFloatingCard(c.Context(), req.UserID)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, card)
}

// GetUserGalgames returns a user's galgame list.
// GET /api/user/:userID/galgames
func (h *UserHandler) GetUserGalgames(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("无效的用户 ID"))
	}
	var req dto.UserGalgamesRequest
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}
	cards, total, appErr := h.userContentService.GetUserGalgameCards(c.Context(), userID, &req)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.Paginated(c, cards, total)
}

// GetUserTopics returns a user's topic list.
// GET /api/user/:userID/topics
func (h *UserHandler) GetUserTopics(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("无效的用户 ID"))
	}
	var req dto.UserTopicsRequest
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}
	items, total, appErr := h.userContentService.GetUserTopics(c.Context(), userID, &req)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, fiber.Map{"topics": items, "total": total})
}

// GetUserReplies returns a user's reply list.
// GET /api/user/:userID/replies
func (h *UserHandler) GetUserReplies(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("无效的用户 ID"))
	}
	var req dto.UserRepliesRequest
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}
	items, total, appErr := h.userContentService.GetUserReplies(c.Context(), userID, &req)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, fiber.Map{"replies": items, "total": total})
}

// GetUserComments returns a user's comment list.
// GET /api/user/:userID/comments
func (h *UserHandler) GetUserComments(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("无效的用户 ID"))
	}
	var req dto.UserCommentsRequest
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}
	items, total, appErr := h.userContentService.GetUserComments(c.Context(), userID, &req)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, fiber.Map{"comments": items, "total": total})
}

// GetUserResources returns a user's galgame resource list.
// GET /api/user/:userID/resources
func (h *UserHandler) GetUserResources(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("无效的用户 ID"))
	}
	var req dto.UserResourcesRequest
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}
	page, appErr := h.userContentService.GetUserResources(c.Context(), userID, &req)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, page)
}

// GetUserGalgameComments returns the "评论 / 被评论 / 点赞评论" comment
// rows for the three sub-tabs under /user/:id/galgame/. Replaces the
// old behaviour where these tabs surfaced the parent galgames as
// galgame-cards — what the user actually wanted was the comment-card
// view used by /user/:id/comment/.
// GET /api/user/:userID/galgame-comments
func (h *UserHandler) GetUserGalgameComments(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("无效的用户 ID"))
	}
	var req dto.UserGalgameCommentsRequest
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}
	items, total, appErr := h.userContentService.GetUserGalgameComments(c.Context(), userID, &req)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, fiber.Map{"comments": items, "total": total})
}

// GetUserRatings returns a user's galgame rating list.
// GET /api/user/:userID/ratings
func (h *UserHandler) GetUserRatings(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("无效的用户 ID"))
	}
	var req dto.UserRatingsRequest
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}
	page, appErr := h.userContentService.GetUserRatings(c.Context(), userID, &req)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, page)
}
