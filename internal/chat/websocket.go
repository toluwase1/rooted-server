package chat

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// Hub manages all active WebSocket connections.
type Hub struct {
	// connections maps user_id → WebSocket connection
	connections map[string]*websocket.Conn
	mu          sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		connections: make(map[string]*websocket.Conn),
	}
}

func (h *Hub) Register(userID string, conn *websocket.Conn) {
	h.mu.Lock()
	h.connections[userID] = conn
	h.mu.Unlock()
}

func (h *Hub) Unregister(userID string) {
	h.mu.Lock()
	delete(h.connections, userID)
	h.mu.Unlock()
}

// Send pushes a message to a specific user if they're connected.
func (h *Hub) Send(userID string, msg interface{}) {
	h.mu.RLock()
	conn, ok := h.connections[userID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	conn.WriteMessage(websocket.TextMessage, data)
}

// WebSocket message types
type WSIncoming struct {
	Type           string `json:"type"`            // "message", "typing"
	ConversationID string `json:"conversation_id"`
	ContentType    string `json:"content_type"`    // "text", "photo", "voice_note"
	Content        string `json:"content"`
}

type WSOutgoing struct {
	Type           string    `json:"type"`            // "message", "typing", "error"
	ConversationID string    `json:"conversation_id"`
	SenderID       string    `json:"sender_id"`
	SenderName     string    `json:"sender_name"`
	ContentType    string    `json:"content_type"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}

// HandleWebSocket upgrades HTTP to WebSocket for Mini App chat.
func HandleWebSocket(hub *Hub, chatService *Service, getUserID func(c *fiber.Ctx) string) fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		userID := c.Query("user_id")
		if userID == "" {
			c.Close()
			return
		}

		hub.Register(userID, c)
		defer hub.Unregister(userID)

		log.Printf("WebSocket connected: user=%s", userID)

		for {
			_, raw, err := c.ReadMessage()
			if err != nil {
				break
			}

			var msg WSIncoming
			if json.Unmarshal(raw, &msg) != nil {
				continue
			}

			switch msg.Type {
			case "message":
				handleWSMessage(hub, chatService, userID, msg)
			case "typing":
				handleWSTyping(hub, chatService, userID, msg)
			}
		}

		log.Printf("WebSocket disconnected: user=%s", userID)
	})
}

func handleWSMessage(hub *Hub, svc *Service, senderID string, msg WSIncoming) {
	ctx := context.Background()

	// Save message to DB
	dbMsg := Message{
		ConversationID: msg.ConversationID,
		SenderID:       senderID,
		ContentType:    msg.ContentType,
		Content:        msg.Content,
	}
	if err := svc.repo.SaveMessage(ctx, dbMsg); err != nil {
		log.Printf("ERROR WebSocket save message: %v", err)
		hub.Send(senderID, WSOutgoing{Type: "error", Content: "Failed to send message"})
		return
	}

	// Find the conversation to get the recipient
	conv, err := svc.repo.GetConversation(ctx, msg.ConversationID)
	if err != nil || conv == nil {
		return
	}

	recipientID := conv.UserBID
	if senderID == conv.UserBID {
		recipientID = conv.UserAID
	}

	// Push to recipient via WebSocket (if connected to Mini App)
	outgoing := WSOutgoing{
		Type:           "message",
		ConversationID: msg.ConversationID,
		SenderID:       senderID,
		ContentType:    msg.ContentType,
		Content:        msg.Content,
		CreatedAt:      time.Now(),
	}
	hub.Send(recipientID, outgoing)

	// Don't echo back to sender — they already have the optimistic update
}

func handleWSTyping(hub *Hub, svc *Service, senderID string, msg WSIncoming) {
	ctx := context.Background()

	conv, err := svc.repo.GetConversation(ctx, msg.ConversationID)
	if err != nil || conv == nil {
		return
	}

	recipientID := conv.UserBID
	if senderID == conv.UserBID {
		recipientID = conv.UserAID
	}

	hub.Send(recipientID, WSOutgoing{
		Type:           "typing",
		ConversationID: msg.ConversationID,
		SenderID:       senderID,
	})
}
