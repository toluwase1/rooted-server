package chat

import "context"

type Repository interface {
	// Conversations
	CreateConversation(ctx context.Context, matchID, userAID, userBID string) (*Conversation, error)
	GetConversation(ctx context.Context, id string) (*Conversation, error)
	GetConversationByMatch(ctx context.Context, matchID string) (*Conversation, error)
	GetUserConversations(ctx context.Context, userID string) ([]Conversation, error)
	UpdateConversationStatus(ctx context.Context, id, status string) error

	// Messages
	SaveMessage(ctx context.Context, msg Message) error
	GetMessages(ctx context.Context, conversationID string, limit, offset int) ([]Message, error)
	GetLatestMessage(ctx context.Context, conversationID string) (*Message, error)

	// Chat routing (Telegram bot relay)
	SetChatRouting(ctx context.Context, routing ChatRouting) error
	GetChatRouting(ctx context.Context, telegramChatID int64) (*ChatRouting, error)
	GetChatRoutingByUser(ctx context.Context, userID string) (*ChatRouting, error)
	SetActiveConversation(ctx context.Context, telegramChatID int64, conversationID string) error

	// Expiry
	GetExpiredConversations(ctx context.Context) ([]Conversation, error)
	GetConversationsNeedingNudge(ctx context.Context, nudgeDays int) ([]Conversation, error)
}
