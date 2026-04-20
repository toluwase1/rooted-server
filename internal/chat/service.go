package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rooted-dating/rooted-server/internal/shared/config"
	"github.com/rooted-dating/rooted-server/internal/shared/database"
)

type Service struct {
	repo   Repository
	redis  *database.SafeRedis
	config *config.DynamicConfig
}

func NewService(repo Repository, redis *database.SafeRedis, cfg *config.DynamicConfig) *Service {
	return &Service{repo: repo, redis: redis, config: cfg}
}

// CreateConversation creates a new proxied chat between two matched users.
func (s *Service) CreateConversation(ctx context.Context, matchID, userAID, userBID string) (*Conversation, error) {
	expiryDays := s.config.GetInt(ctx, "conversation_expiry_days", 14)

	conv, err := s.repo.CreateConversation(ctx, matchID, userAID, userBID)
	if err != nil {
		return nil, err
	}

	// Set expiry
	expiresAt := time.Now().AddDate(0, 0, expiryDays)
	conv.ExpiresAt = &expiresAt

	return conv, nil
}

// RelayMessage handles the core bot-proxied chat flow:
// User A sends message to bot → we route it to User B via bot.
func (s *Service) RelayMessage(ctx context.Context, senderTelegramChatID int64, contentType, content string) (*Message, string, error) {
	// 1. Look up who is sending and their active conversation
	routing, err := s.getChatRouting(ctx, senderTelegramChatID)
	if err != nil || routing == nil {
		return nil, "", fmt.Errorf("no active chat routing for telegram chat %d", senderTelegramChatID)
	}

	if routing.ActiveConversationID == "" {
		return nil, "", fmt.Errorf("no active conversation selected")
	}

	// 2. Get the conversation to find the recipient
	conv, err := s.repo.GetConversation(ctx, routing.ActiveConversationID)
	if err != nil || conv == nil {
		return nil, "", fmt.Errorf("conversation not found")
	}

	if conv.Status != "active" {
		return nil, "", fmt.Errorf("conversation is %s", conv.Status)
	}

	// 3. Rate limit check
	rateLimitKey := fmt.Sprintf("ratelimit:%s:messages", routing.UserID)
	count, _ := s.redis.Get(ctx, rateLimitKey).Int()
	maxPerMin := s.config.GetInt(ctx, "message_rate_limit_per_minute", 30)
	if count >= maxPerMin {
		return nil, "", fmt.Errorf("message rate limit exceeded")
	}
	s.redis.Incr(ctx, rateLimitKey)
	s.redis.Expire(ctx, rateLimitKey, 1*time.Minute)

	// 4. Determine recipient
	recipientID := conv.UserBID
	if routing.UserID == conv.UserBID {
		recipientID = conv.UserAID
	}

	// 5. Save message
	msg := Message{
		ConversationID: conv.ID,
		SenderID:       routing.UserID,
		ContentType:    contentType,
		Content:        content,
	}
	if err := s.repo.SaveMessage(ctx, msg); err != nil {
		return nil, "", fmt.Errorf("saving message: %w", err)
	}

	// 6. Get recipient's Telegram chat ID for forwarding
	recipientRouting, err := s.repo.GetChatRoutingByUser(ctx, recipientID)
	if err != nil || recipientRouting == nil {
		return &msg, "", fmt.Errorf("recipient routing not found")
	}

	return &msg, recipientID, nil
}

// GetChatRouting returns the routing record for a telegram chat ID.
func (s *Service) GetChatRouting(ctx context.Context, telegramChatID int64) (*ChatRouting, error) {
	return s.getChatRouting(ctx, telegramChatID)
}

// GetRecipientTelegramChatID returns the Telegram chat ID for a user.
func (s *Service) GetRecipientTelegramChatID(ctx context.Context, userID string) (int64, error) {
	routing, err := s.repo.GetChatRoutingByUser(ctx, userID)
	if err != nil || routing == nil {
		return 0, fmt.Errorf("routing not found for user %s", userID)
	}
	return routing.TelegramChatID, nil
}

// SetActiveConversation sets which match the user is currently chatting with.
func (s *Service) SetActiveConversation(ctx context.Context, telegramChatID int64, conversationID string) error {
	if err := s.repo.SetActiveConversation(ctx, telegramChatID, conversationID); err != nil {
		return err
	}
	// Update cache
	s.redis.Del(ctx, fmt.Sprintf("chatroute:%d", telegramChatID))
	return nil
}

// RegisterChatRouting links a Telegram chat_id to an internal user.
func (s *Service) RegisterChatRouting(ctx context.Context, telegramChatID int64, userID string) error {
	routing := ChatRouting{
		TelegramChatID: telegramChatID,
		UserID:         userID,
	}
	if err := s.repo.SetChatRouting(ctx, routing); err != nil {
		return err
	}
	// Cache the routing
	data, _ := json.Marshal(routing)
	s.redis.Set(ctx, fmt.Sprintf("chatroute:%d", telegramChatID), data, 0) // no TTL
	return nil
}

// GetMessages returns paginated messages for a conversation.
func (s *Service) GetMessages(ctx context.Context, conversationID string, limit, offset int) ([]Message, error) {
	return s.repo.GetMessages(ctx, conversationID, limit, offset)
}

// GetUserConversations returns all active conversations for a user.
func (s *Service) GetUserConversations(ctx context.Context, userID string) ([]Conversation, error) {
	return s.repo.GetUserConversations(ctx, userID)
}

// CloseConversation ends a proxied chat (unmatch).
func (s *Service) CloseConversation(ctx context.Context, conversationID string) error {
	return s.repo.UpdateConversationStatus(ctx, conversationID, "closed")
}

// ExpireStaleConversations finds and expires conversations past their expiry date.
func (s *Service) ExpireStaleConversations(ctx context.Context) (int, error) {
	convs, err := s.repo.GetExpiredConversations(ctx)
	if err != nil {
		return 0, err
	}

	for _, conv := range convs {
		s.repo.UpdateConversationStatus(ctx, conv.ID, "expired")
	}

	return len(convs), nil
}

// GetConversationsNeedingNudge returns conversations that need a "don't let this fade" nudge.
func (s *Service) GetConversationsNeedingNudge(ctx context.Context) ([]Conversation, error) {
	nudgeDays := s.config.GetInt(ctx, "chat_nudge_day", 7)
	return s.repo.GetConversationsNeedingNudge(ctx, nudgeDays)
}

func (s *Service) getChatRouting(ctx context.Context, telegramChatID int64) (*ChatRouting, error) {
	cacheKey := fmt.Sprintf("chatroute:%d", telegramChatID)

	// Check cache
	cached, err := s.redis.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var routing ChatRouting
		if json.Unmarshal(cached, &routing) == nil {
			return &routing, nil
		}
	}

	// Cache miss
	routing, err := s.repo.GetChatRouting(ctx, telegramChatID)
	if err != nil {
		return nil, err
	}
	if routing == nil {
		return nil, nil
	}

	// Cache permanently (invalidated on change)
	data, _ := json.Marshal(routing)
	s.redis.Set(ctx, cacheKey, data, 0)

	return routing, nil
}
