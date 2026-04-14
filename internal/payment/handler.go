package payment

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
	api.Get("/plans", h.GetPlans)
	api.Get("/subscription", h.GetSubscription)
	api.Get("/transactions", h.GetTransactions)
}

func (h *Handler) getUser(c *fiber.Ctx) (*user.User, error) {
	tgUser := c.Locals("telegram_user").(middleware.TelegramUser)
	u, _, err := h.userService.FindOrCreateUser(c.Context(), tgUser.ID, tgUser.Username)
	return u, err
}

func (h *Handler) GetPlans(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"plans": h.service.GetPlans(c.Context()),
		"items": h.service.GetPurchaseItems(c.Context()),
	})
}

func (h *Handler) GetSubscription(c *fiber.Ctx) error {
	u, _ := h.getUser(c)
	sub, _ := h.service.GetActiveSubscription(c.Context(), u.ID)
	return c.JSON(fiber.Map{"subscription": sub})
}

func (h *Handler) GetTransactions(c *fiber.Ctx) error {
	u, _ := h.getUser(c)
	txns, _ := h.service.GetTransactionHistory(c.Context(), u.ID)
	return c.JSON(fiber.Map{"transactions": txns})
}
