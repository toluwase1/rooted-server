package chat

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/rooted-dating/rooted-server/internal/shared/middleware"
	"github.com/rooted-dating/rooted-server/internal/user"
)

type Handler struct {
	service     *Service
	userService *user.Service
	hub         *Hub
}

func NewHandler(s *Service, us *user.Service, hub *Hub) *Handler {
	return &Handler{service: s, userService: us, hub: hub}
}

func (h *Handler) RegisterRoutes(api fiber.Router) {
	api.Get("/conversations", h.GetConversations)
	api.Get("/messages/:conversationId", h.GetMessages)
	api.Post("/messages/:conversationId", h.SendMessage)
}

func (h *Handler) SendMessage(c *fiber.Ctx) error {
	tgUser, ok := c.Locals("telegram_user").(middleware.TelegramUser)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	u, _, _ := h.userService.FindOrCreateUser(c.Context(), tgUser.ID, tgUser.Username)

	var body struct {
		Content     string `json:"content"`
		ContentType string `json:"content_type"`
	}
	if err := c.BodyParser(&body); err != nil || body.Content == "" {
		return c.Status(400).JSON(fiber.Map{"error": "content is required"})
	}
	if body.ContentType == "" {
		body.ContentType = "text"
	}

	msg := Message{
		ConversationID: c.Params("conversationId"),
		SenderID:       u.ID,
		ContentType:    body.ContentType,
		Content:        body.Content,
	}
	if err := h.service.repo.SaveMessage(c.Context(), msg); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Something went wrong. Please try again."})
	}

	// Push to recipient via WebSocket
	conv, _ := h.service.repo.GetConversation(c.Context(), c.Params("conversationId"))
	if conv != nil {
		recipientID := conv.UserBID
		if u.ID == conv.UserBID {
			recipientID = conv.UserAID
		}
		h.hub.Send(recipientID, WSOutgoing{
			Type:           "message",
			ConversationID: conv.ID,
			SenderID:       u.ID,
			ContentType:    body.ContentType,
			Content:        body.Content,
		})
	}

	return c.JSON(fiber.Map{"status": "sent"})
}

// RegisterWebSocket registers the WebSocket endpoint (called from main.go, outside auth group).
func (h *Handler) RegisterWebSocket(app *fiber.App) {
	app.Get("/ws/chat", HandleWebSocket(h.hub, h.service, func(c *fiber.Ctx) string {
		return c.Query("user_id")
	}))
}

func (h *Handler) GetConversations(c *fiber.Ctx) error {
	tgUser, ok := c.Locals("telegram_user").(middleware.TelegramUser)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	u, _, _ := h.userService.FindOrCreateUser(c.Context(), tgUser.ID, tgUser.Username)
	convs, _ := h.service.GetUserConversations(c.Context(), u.ID)
	if convs == nil {
		convs = []Conversation{}
	}

	// Enrich with other user's name
	var enriched []fiber.Map
	for _, conv := range convs {
		otherID := conv.UserBID
		if u.ID == conv.UserBID {
			otherID = conv.UserAID
		}
		otherProfile, _ := h.userService.GetProfile(c.Context(), otherID)
		otherName := ""
		if otherProfile != nil {
			otherName = otherProfile.FirstName
		}

		// Get last message preview
		lastMsg, _ := h.service.repo.GetLatestMessage(c.Context(), conv.ID)
		var preview string
		if lastMsg != nil {
			preview = lastMsg.Content
			if len(preview) > 50 {
				preview = preview[:50] + "..."
			}
		}

		enriched = append(enriched, fiber.Map{
			"conversation":    conv,
			"other_user_name": otherName,
			"other_user_id":   otherID,
			"last_message":    preview,
		})
	}
	if enriched == nil {
		enriched = []fiber.Map{}
	}

	return c.JSON(fiber.Map{"conversations": enriched})
}

func (h *Handler) GetMessages(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	msgs, _ := h.service.GetMessages(c.Context(), c.Params("conversationId"), limit, offset)
	if msgs == nil {
		msgs = []Message{}
	}
	return c.JSON(fiber.Map{"messages": msgs})
}
