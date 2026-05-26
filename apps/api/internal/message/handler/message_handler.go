package handler

import (
	"strconv"

	"kun-galgame-api/internal/message/dto"
	"kun-galgame-api/internal/message/service"
	"kun-galgame-api/internal/middleware"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/response"
	"kun-galgame-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type MessageHandler struct {
	messageService *service.MessageService
}

func NewMessageHandler(messageService *service.MessageService) *MessageHandler {
	return &MessageHandler{messageService: messageService}
}

// GetMessages returns paginated notification messages.
// GET /api/message
func (h *MessageHandler) GetMessages(c *fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	var req dto.ListMessagesRequest
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	result, appErr := h.messageService.GetMessages(c.Context(), user.ID, &req)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, result)
}

// DeleteMessage deletes a single notification message.
// DELETE /api/message/:id
func (h *MessageHandler) DeleteMessage(c *fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("无效的消息 ID"))
	}

	if appErr := h.messageService.DeleteMessage(c.Context(), user.ID, id); appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OKMessage(c, "消息已删除")
}

// GetSystemMessages returns all system broadcast messages (public).
// GET /api/message/admin
func (h *MessageHandler) GetSystemMessages(c *fiber.Ctx) error {
	messages, appErr := h.messageService.GetSystemMessages(c.Context())
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, messages)
}

// MarkAdminRead marks all system broadcast messages as read.
// PUT /api/message/admin/read
//
// TODO(critical, schema-change): this endpoint is GLOBAL — any logged-in
// user calling it flips `system_message.status` from `unread` → `read`
// for the entire row, which means EVERY OTHER USER also loses their
// unread badge instantly. The schema currently has no per-user read
// state for system_message; fixing it properly needs a new
// `system_message_read_state` table modeled after
// `wiki_message_read_state` (high-water-mark cursor per user), then
// rewriting this handler to bump only the caller's cursor and the
// `GET /message/admin` handler to read unread relative to that cursor.
// See migrations/008 for the wiki-message precedent. Leaving this as a
// known footgun until the schema change is scheduled.
func (h *MessageHandler) MarkAdminRead(c *fiber.Ctx) error {
	if _, appErr := middleware.MustGetUser(c); appErr != nil {
		return response.Error(c, appErr)
	}

	if appErr := h.messageService.MarkAllSystemRead(c.Context()); appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OKMessage(c, "已标记全部已读")
}

// GetNavSummary returns nav-bar summary [notice, system].
// GET /api/message/nav/system
func (h *MessageHandler) GetNavSummary(c *fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	result, appErr := h.messageService.GetNavSummary(c.Context(), user.ID)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, result)
}

// MarkAllRead marks all user notification messages as read.
// PUT /api/message/system/read
func (h *MessageHandler) MarkAllRead(c *fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	if appErr := h.messageService.MarkAllRead(c.Context(), user.ID); appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OKMessage(c, "已标记全部已读")
}
