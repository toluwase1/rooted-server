package chat_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/rooted-dating/rooted-server/internal/chat"
	"github.com/rooted-dating/rooted-server/internal/shared/config"
	"github.com/rooted-dating/rooted-server/internal/shared/database"
)

var (
	testDB      *pgxpool.Pool
	testRepo    *chat.PostgresRepo
	testService *chat.Service
	testCtx     = context.Background()
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	pgContainer, err := postgres.Run(ctx, "postgis/postgis:14-3.4-alpine",
		postgres.WithDatabase("rooted_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to start postgres: %v", err))
	}

	connStr, _ := pgContainer.ConnectionString(ctx, "sslmode=disable")
	testDB, err = database.NewPostgres(ctx, connStr)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect: %v", err))
	}

	if err := runMigrations(ctx, testDB); err != nil {
		panic(fmt.Sprintf("Migrations failed: %v", err))
	}

	testRepo = chat.NewPostgresRepo(testDB)
	rdb := database.NewSafeRedis(nil)
	dynConfig := config.NewDynamicConfig(testDB, rdb)
	testService = chat.NewService(testRepo, rdb, dynConfig)

	code := m.Run()
	testDB.Close()
	pgContainer.Terminate(ctx)
	os.Exit(code)
}

func runMigrations(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

		CREATE TABLE users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			telegram_id BIGINT UNIQUE NOT NULL,
			status VARCHAR(20) DEFAULT 'active',
			verification VARCHAR(20) DEFAULT 'unverified',
			trust_score SMALLINT DEFAULT 50,
			subscription VARCHAR(20) DEFAULT 'free',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			last_active_at TIMESTAMPTZ DEFAULT NOW(),
			telegram_username VARCHAR(64),
			sub_expires_at TIMESTAMPTZ
		);

		CREATE TABLE matches (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_a_id UUID NOT NULL REFERENCES users(id),
			user_b_id UUID NOT NULL REFERENCES users(id),
			compatibility SMALLINT,
			status VARCHAR(20) DEFAULT 'active',
			matched_at TIMESTAMPTZ DEFAULT NOW(),
			last_message_at TIMESTAMPTZ,
			UNIQUE(user_a_id, user_b_id)
		);

		CREATE TABLE conversations (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			match_id UUID NOT NULL REFERENCES matches(id),
			user_a_id UUID NOT NULL,
			user_b_id UUID NOT NULL,
			status VARCHAR(20) DEFAULT 'active',
			last_message_at TIMESTAMPTZ,
			message_count INT DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			expires_at TIMESTAMPTZ
		);

		CREATE TABLE messages (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			conversation_id UUID NOT NULL REFERENCES conversations(id),
			sender_id UUID NOT NULL REFERENCES users(id),
			content_type VARCHAR(20) NOT NULL,
			content TEXT,
			moderation_status VARCHAR(20) DEFAULT 'clean',
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE TABLE telegram_chat_routing (
			telegram_chat_id BIGINT PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id),
			active_conversation_id UUID REFERENCES conversations(id)
		);

		CREATE TABLE admin_config (
			key VARCHAR(100) PRIMARY KEY,
			value JSONB NOT NULL,
			category VARCHAR(50) NOT NULL,
			description TEXT,
			value_type VARCHAR(20) NOT NULL,
			constraints JSONB,
			updated_by UUID,
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		INSERT INTO admin_config (key, value, category, description, value_type) VALUES
		('conversation_expiry_days', '14', 'chat', '', 'int'),
		('chat_nudge_day', '7', 'chat', '', 'int'),
		('message_rate_limit_per_minute', '30', 'chat', '', 'int');
	`)
	return err
}

func createUsers(t *testing.T) (string, string) {
	t.Helper()
	var userA, userB string
	testDB.QueryRow(testCtx, "INSERT INTO users (telegram_id) VALUES (100) RETURNING id").Scan(&userA)
	testDB.QueryRow(testCtx, "INSERT INTO users (telegram_id) VALUES (200) RETURNING id").Scan(&userB)
	return userA, userB
}

func createMatch(t *testing.T, userA, userB string) string {
	t.Helper()
	var matchID string
	testDB.QueryRow(testCtx, "INSERT INTO matches (user_a_id, user_b_id, compatibility) VALUES ($1, $2, 80) RETURNING id",
		userA, userB).Scan(&matchID)
	return matchID
}

func cleanDB(t *testing.T) {
	t.Helper()
	testDB.Exec(testCtx, "DELETE FROM telegram_chat_routing")
	testDB.Exec(testCtx, "DELETE FROM messages")
	testDB.Exec(testCtx, "DELETE FROM conversations")
	testDB.Exec(testCtx, "DELETE FROM matches")
	testDB.Exec(testCtx, "DELETE FROM users")
}

// ============================================================
// CONVERSATION TESTS
// ============================================================

func TestCreateConversation(t *testing.T) {
	cleanDB(t)
	userA, userB := createUsers(t)
	matchID := createMatch(t, userA, userB)

	conv, err := testService.CreateConversation(testCtx, matchID, userA, userB)
	require.NoError(t, err)
	assert.NotEmpty(t, conv.ID)
	assert.Equal(t, "active", conv.Status)
	assert.Equal(t, userA, conv.UserAID)
	assert.Equal(t, userB, conv.UserBID)
	assert.Equal(t, 0, conv.MessageCount)
}

func TestGetUserConversations(t *testing.T) {
	cleanDB(t)
	userA, userB := createUsers(t)
	matchID := createMatch(t, userA, userB)

	testService.CreateConversation(testCtx, matchID, userA, userB)

	convs, err := testService.GetUserConversations(testCtx, userA)
	require.NoError(t, err)
	assert.Len(t, convs, 1)

	// UserB should also see it
	convsB, _ := testService.GetUserConversations(testCtx, userB)
	assert.Len(t, convsB, 1)
}

func TestCloseConversation(t *testing.T) {
	cleanDB(t)
	userA, userB := createUsers(t)
	matchID := createMatch(t, userA, userB)

	conv, _ := testService.CreateConversation(testCtx, matchID, userA, userB)
	testService.CloseConversation(testCtx, conv.ID)

	convs, _ := testService.GetUserConversations(testCtx, userA)
	assert.Len(t, convs, 0) // closed conversations not returned
}

// ============================================================
// MESSAGE TESTS
// ============================================================

func TestSaveAndGetMessages(t *testing.T) {
	cleanDB(t)
	userA, userB := createUsers(t)
	matchID := createMatch(t, userA, userB)
	conv, _ := testService.CreateConversation(testCtx, matchID, userA, userB)

	// Save messages directly via repo
	testRepo.SaveMessage(testCtx, chat.Message{
		ConversationID: conv.ID, SenderID: userA,
		ContentType: "text", Content: "Hello!",
	})
	testRepo.SaveMessage(testCtx, chat.Message{
		ConversationID: conv.ID, SenderID: userB,
		ContentType: "text", Content: "Hi there!",
	})

	msgs, err := testService.GetMessages(testCtx, conv.ID, 50, 0)
	require.NoError(t, err)
	assert.Len(t, msgs, 2)
}

func TestMessageUpdatesConversation(t *testing.T) {
	cleanDB(t)
	userA, userB := createUsers(t)
	matchID := createMatch(t, userA, userB)
	conv, _ := testService.CreateConversation(testCtx, matchID, userA, userB)

	testRepo.SaveMessage(testCtx, chat.Message{
		ConversationID: conv.ID, SenderID: userA,
		ContentType: "text", Content: "Hey",
	})

	// Conversation should have updated message count and last_message_at
	updated, _ := testRepo.GetConversation(testCtx, conv.ID)
	assert.Equal(t, 1, updated.MessageCount)
	assert.NotNil(t, updated.LastMessageAt)
}

// ============================================================
// CHAT ROUTING TESTS
// ============================================================

func TestChatRouting(t *testing.T) {
	cleanDB(t)
	userA, userB := createUsers(t)
	matchID := createMatch(t, userA, userB)
	conv, _ := testService.CreateConversation(testCtx, matchID, userA, userB)

	// Register routing
	testService.RegisterChatRouting(testCtx, 1001, userA)
	testService.SetActiveConversation(testCtx, 1001, conv.ID)

	// Get routing
	chatID, err := testService.GetRecipientTelegramChatID(testCtx, userA)
	require.NoError(t, err)
	assert.Equal(t, int64(1001), chatID)
}

func TestChatRouting_NotFound(t *testing.T) {
	cleanDB(t)

	_, err := testService.GetRecipientTelegramChatID(testCtx, "nonexistent-id")
	assert.Error(t, err)
}

// ============================================================
// EXPIRY TESTS
// ============================================================

func TestExpireStaleConversations(t *testing.T) {
	cleanDB(t)
	userA, userB := createUsers(t)
	matchID := createMatch(t, userA, userB)

	conv, _ := testService.CreateConversation(testCtx, matchID, userA, userB)

	// Set expiry to the past
	testDB.Exec(testCtx, "UPDATE conversations SET expires_at = NOW() - INTERVAL '1 day' WHERE id = $1", conv.ID)

	count, err := testService.ExpireStaleConversations(testCtx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Should not show in active conversations
	convs, _ := testService.GetUserConversations(testCtx, userA)
	assert.Len(t, convs, 0)
}

func TestGetConversationsNeedingNudge(t *testing.T) {
	cleanDB(t)
	userA, userB := createUsers(t)
	matchID := createMatch(t, userA, userB)

	conv, _ := testService.CreateConversation(testCtx, matchID, userA, userB)

	// Set last message to 7 days ago
	testDB.Exec(testCtx, "UPDATE conversations SET last_message_at = NOW() - INTERVAL '7 days' WHERE id = $1", conv.ID)

	convs, err := testService.GetConversationsNeedingNudge(testCtx)
	require.NoError(t, err)
	assert.Len(t, convs, 1)
	assert.Equal(t, conv.ID, convs[0].ID)
}
