package chat

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepo struct {
	db *pgxpool.Pool
}

func NewPostgresRepo(db *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) CreateConversation(ctx context.Context, matchID, userAID, userBID string) (*Conversation, error) {
	var conv Conversation
	err := r.db.QueryRow(ctx, `
		INSERT INTO conversations (match_id, user_a_id, user_b_id,
			expires_at)
		VALUES ($1, $2, $3,
			NOW() + INTERVAL '14 days')
		RETURNING id, match_id, user_a_id, user_b_id, status,
		          last_message_at, message_count, created_at, expires_at
	`, matchID, userAID, userBID).Scan(
		&conv.ID, &conv.MatchID, &conv.UserAID, &conv.UserBID,
		&conv.Status, &conv.LastMessageAt, &conv.MessageCount,
		&conv.CreatedAt, &conv.ExpiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("creating conversation: %w", err)
	}
	return &conv, nil
}

func (r *PostgresRepo) GetConversation(ctx context.Context, id string) (*Conversation, error) {
	var conv Conversation
	err := r.db.QueryRow(ctx, `
		SELECT id, match_id, user_a_id, user_b_id, status,
		       last_message_at, message_count, created_at, expires_at
		FROM conversations WHERE id = $1
	`, id).Scan(
		&conv.ID, &conv.MatchID, &conv.UserAID, &conv.UserBID,
		&conv.Status, &conv.LastMessageAt, &conv.MessageCount,
		&conv.CreatedAt, &conv.ExpiresAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *PostgresRepo) GetConversationByMatch(ctx context.Context, matchID string) (*Conversation, error) {
	var conv Conversation
	err := r.db.QueryRow(ctx, `
		SELECT id, match_id, user_a_id, user_b_id, status,
		       last_message_at, message_count, created_at, expires_at
		FROM conversations WHERE match_id = $1 AND status = 'active'
	`, matchID).Scan(
		&conv.ID, &conv.MatchID, &conv.UserAID, &conv.UserBID,
		&conv.Status, &conv.LastMessageAt, &conv.MessageCount,
		&conv.CreatedAt, &conv.ExpiresAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *PostgresRepo) GetUserConversations(ctx context.Context, userID string) ([]Conversation, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, match_id, user_a_id, user_b_id, status,
		       last_message_at, message_count, created_at, expires_at
		FROM conversations
		WHERE (user_a_id = $1 OR user_b_id = $1) AND status = 'active'
		ORDER BY COALESCE(last_message_at, created_at) DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var convs []Conversation
	for rows.Next() {
		var conv Conversation
		if err := rows.Scan(&conv.ID, &conv.MatchID, &conv.UserAID, &conv.UserBID,
			&conv.Status, &conv.LastMessageAt, &conv.MessageCount,
			&conv.CreatedAt, &conv.ExpiresAt); err != nil {
			return nil, err
		}
		convs = append(convs, conv)
	}
	return convs, nil
}

func (r *PostgresRepo) UpdateConversationStatus(ctx context.Context, id, status string) error {
	_, err := r.db.Exec(ctx,
		"UPDATE conversations SET status = $2 WHERE id = $1", id, status)
	return err
}

func (r *PostgresRepo) SaveMessage(ctx context.Context, msg Message) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO messages (conversation_id, sender_id, content_type, content)
		VALUES ($1, $2, $3, $4)
	`, msg.ConversationID, msg.SenderID, msg.ContentType, msg.Content)
	if err != nil {
		return err
	}

	// Update conversation metadata
	_, err = r.db.Exec(ctx, `
		UPDATE conversations
		SET last_message_at = NOW(),
		    message_count = message_count + 1,
		    expires_at = NOW() + INTERVAL '14 days'
		WHERE id = $1
	`, msg.ConversationID)
	return err
}

func (r *PostgresRepo) GetMessages(ctx context.Context, conversationID string, limit, offset int) ([]Message, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, conversation_id, sender_id, content_type, content,
		       moderation_status, created_at
		FROM messages WHERE conversation_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, conversationID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderID,
			&m.ContentType, &m.Content, &m.ModerationStatus, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func (r *PostgresRepo) GetLatestMessage(ctx context.Context, conversationID string) (*Message, error) {
	var m Message
	err := r.db.QueryRow(ctx, `
		SELECT id, conversation_id, sender_id, content_type, content,
		       moderation_status, created_at
		FROM messages WHERE conversation_id = $1
		ORDER BY created_at DESC LIMIT 1
	`, conversationID).Scan(&m.ID, &m.ConversationID, &m.SenderID,
		&m.ContentType, &m.Content, &m.ModerationStatus, &m.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *PostgresRepo) SetChatRouting(ctx context.Context, routing ChatRouting) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO telegram_chat_routing (telegram_chat_id, user_id, active_conversation_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (telegram_chat_id) DO UPDATE
		SET user_id = $2, active_conversation_id = $3
	`, routing.TelegramChatID, routing.UserID, routing.ActiveConversationID)
	return err
}

func (r *PostgresRepo) GetChatRouting(ctx context.Context, telegramChatID int64) (*ChatRouting, error) {
	var routing ChatRouting
	err := r.db.QueryRow(ctx, `
		SELECT telegram_chat_id, user_id, active_conversation_id
		FROM telegram_chat_routing WHERE telegram_chat_id = $1
	`, telegramChatID).Scan(&routing.TelegramChatID, &routing.UserID, &routing.ActiveConversationID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &routing, nil
}

func (r *PostgresRepo) GetChatRoutingByUser(ctx context.Context, userID string) (*ChatRouting, error) {
	var routing ChatRouting
	err := r.db.QueryRow(ctx, `
		SELECT telegram_chat_id, user_id, active_conversation_id
		FROM telegram_chat_routing WHERE user_id = $1
	`, userID).Scan(&routing.TelegramChatID, &routing.UserID, &routing.ActiveConversationID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &routing, nil
}

func (r *PostgresRepo) SetActiveConversation(ctx context.Context, telegramChatID int64, conversationID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE telegram_chat_routing SET active_conversation_id = $2
		WHERE telegram_chat_id = $1
	`, telegramChatID, conversationID)
	return err
}

func (r *PostgresRepo) GetExpiredConversations(ctx context.Context) ([]Conversation, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, match_id, user_a_id, user_b_id, status,
		       last_message_at, message_count, created_at, expires_at
		FROM conversations
		WHERE status = 'active' AND expires_at < NOW()
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var convs []Conversation
	for rows.Next() {
		var conv Conversation
		rows.Scan(&conv.ID, &conv.MatchID, &conv.UserAID, &conv.UserBID,
			&conv.Status, &conv.LastMessageAt, &conv.MessageCount,
			&conv.CreatedAt, &conv.ExpiresAt)
		convs = append(convs, conv)
	}
	return convs, nil
}

func (r *PostgresRepo) GetConversationsNeedingNudge(ctx context.Context, nudgeDays int) ([]Conversation, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, match_id, user_a_id, user_b_id, status,
		       last_message_at, message_count, created_at, expires_at
		FROM conversations
		WHERE status = 'active'
		  AND last_message_at < NOW() - INTERVAL '1 day' * $1
		  AND last_message_at > NOW() - INTERVAL '1 day' * ($1 + 1)
	`, nudgeDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var convs []Conversation
	for rows.Next() {
		var conv Conversation
		rows.Scan(&conv.ID, &conv.MatchID, &conv.UserAID, &conv.UserBID,
			&conv.Status, &conv.LastMessageAt, &conv.MessageCount,
			&conv.CreatedAt, &conv.ExpiresAt)
		convs = append(convs, conv)
	}
	return convs, nil
}
