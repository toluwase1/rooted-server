package moderation_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rooted-dating/rooted-server/internal/moderation"
	"github.com/rooted-dating/rooted-server/internal/shared/config"
	"github.com/rooted-dating/rooted-server/internal/shared/database"
)

func TestModerateText_Clean(t *testing.T) {
	rdb := database.NewSafeRedis(nil)
	cfg := config.NewDynamicConfig(nil, rdb)
	svc := moderation.NewService(cfg)

	result := svc.ModerateText(context.Background(), "Hello, how are you today?")
	assert.Equal(t, "clean", result)
}

func TestModerateText_ScamPatterns(t *testing.T) {
	rdb := database.NewSafeRedis(nil)
	cfg := config.NewDynamicConfig(nil, rdb)
	svc := moderation.NewService(cfg)

	tests := []struct {
		input    string
		expected string
	}{
		{"Send me money via Western Union", "flagged"},
		{"Buy me a gift card please", "flagged"},
		{"I have an investment opportunity for you", "flagged"},
		{"Send bitcoin to this address", "flagged"},
		{"Click this link http://scam.com/free", "flagged"},
		{"My number is 08012345678", "flagged"}, // phone number detection
		{"Whatsapp me at this number", "flagged"},
	}

	for _, tt := range tests {
		t.Run(tt.input[:20], func(t *testing.T) {
			result := svc.ModerateText(context.Background(), tt.input)
			assert.Equal(t, tt.expected, result, "input: %s", tt.input)
		})
	}
}

func TestModerateText_URLDetection(t *testing.T) {
	rdb := database.NewSafeRedis(nil)
	cfg := config.NewDynamicConfig(nil, rdb)
	svc := moderation.NewService(cfg)

	assert.Equal(t, "flagged", svc.ModerateText(context.Background(), "check out https://example.com/profile"))
	assert.Equal(t, "flagged", svc.ModerateText(context.Background(), "go to example.com/meet"))
	assert.Equal(t, "flagged", svc.ModerateText(context.Background(), "visit my site at sketchy.xyz"))
}

func TestModerateText_NormalConversation(t *testing.T) {
	rdb := database.NewSafeRedis(nil)
	cfg := config.NewDynamicConfig(nil, rdb)
	svc := moderation.NewService(cfg)

	clean := []string{
		"I love cooking jollof rice",
		"What do you do for fun?",
		"I'm an engineer based in Lagos",
		"That's really interesting!",
		"Would you like to meet up sometime?",
		"I enjoy reading and playing table tennis",
	}

	for _, msg := range clean {
		t.Run(msg[:20], func(t *testing.T) {
			result := svc.ModerateText(context.Background(), msg)
			assert.Equal(t, "clean", result, "should be clean: %s", msg)
		})
	}
}

func TestShouldAutoBan(t *testing.T) {
	rdb := database.NewSafeRedis(nil)
	cfg := config.NewDynamicConfig(nil, rdb)
	svc := moderation.NewService(cfg)

	// Default threshold is 3
	assert.False(t, svc.ShouldAutoBan(context.Background(), 1))
	assert.False(t, svc.ShouldAutoBan(context.Background(), 2))
	assert.True(t, svc.ShouldAutoBan(context.Background(), 3))
	assert.True(t, svc.ShouldAutoBan(context.Background(), 5))
}
