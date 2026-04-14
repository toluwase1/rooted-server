package chat

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rooted-dating/rooted-server/internal/shared/middleware"
	"github.com/rooted-dating/rooted-server/internal/user"
)

type Handler struct {
	service     *Service
	userService *user.Service
}

func NewHandler(s *Service, us *user.Service) *Handler {
	return &Handler{service: s, userService: us}
}

func (h *Handler) RegisterRoutes(api fiber.Router) {
	api.Get("/conversations", h.GetConversations)
	api.Get("/messages/:conversationId", h.GetMessages)
}

func (h *Handler) GetConversations(c *fiber.Ctx) error {
	tgUser := c.Locals("telegram_user").(middleware.TelegramUser)
	u, _, _ := h.userService.FindOrCreateUser(c.Context(), tgUser.ID, tgUser.Username)
	convs, _ := h.service.GetUserConversations(c.Context(), u.ID)
	return c.JSON(fiber.Map{"conversations": convs})
}

func (h *Handler) GetMessages(c *fiber.Ctx) error {
	msgs, _ := h.service.GetMessages(c.Context(), c.Params("conversationId"), 50, 0)
	return c.JSON(fiber.Map{"messages": msgs})
}
