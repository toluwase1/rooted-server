package chat

import "time"

type Conversation struct {
	ID            string     `json:"id"`
	MatchID       string     `json:"match_id"`
	UserAID       string     `json:"user_a_id"`
	UserBID       string     `json:"user_b_id"`
	Status        string     `json:"status"` // active, expired, closed
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
	MessageCount  int        `json:"message_count"`
	CreatedAt     time.Time  `json:"created_at"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
}

type Message struct {
	ID               string    `json:"id"`
	ConversationID   string    `json:"conversation_id"`
	SenderID         string    `json:"sender_id"`
	ContentType      string    `json:"content_type"` // text, photo, voice_note
	Content          string    `json:"content"`
	ModerationStatus string    `json:"moderation_status"` // clean, flagged, blocked
	CreatedAt        time.Time `json:"created_at"`
}

// ChatRouting maps a Telegram chat_id to an internal user and their active conversation.
type ChatRouting struct {
	TelegramChatID       int64  `json:"telegram_chat_id"`
	UserID               string `json:"user_id"`
	ActiveConversationID string `json:"active_conversation_id"`
}
