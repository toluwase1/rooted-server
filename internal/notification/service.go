package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rooted-dating/rooted-server/internal/shared/config"
)

const telegramAPI = "https://api.telegram.org/bot%s/%s"

type Service struct {
	botToken string
	redis    *redis.Client
	config   *config.DynamicConfig
	client   *http.Client
}

func NewService(botToken string, redis *redis.Client, cfg *config.DynamicConfig) *Service {
	return &Service{
		botToken: botToken,
		redis:    redis,
		config:   cfg,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

// NotifyMatch sends "It's a match!" to both users.
func (s *Service) NotifyMatch(ctx context.Context, userAChatID, userBChatID int64, userAName, userBName string) {
	msgA := fmt.Sprintf("You and %s are a match! Tap to start chatting.", userBName)
	msgB := fmt.Sprintf("You and %s are a match! Tap to start chatting.", userAName)

	go s.sendWithCooldown(ctx, userAChatID, msgA)
	go s.sendWithCooldown(ctx, userBChatID, msgB)
}

// NotifyNewLike sends "Someone liked your profile!" (vague for free, specific for Plus+).
func (s *Service) NotifyNewLike(ctx context.Context, chatID int64, likerName string, isPremium bool) {
	var msg string
	if isPremium {
		msg = fmt.Sprintf("%s liked your profile! Open Rooted to see their profile.", likerName)
	} else {
		msg = "Someone liked your profile! Open Rooted to see who."
	}

	go s.sendWithCooldown(ctx, chatID, msg)
}

// NotifyNewMessage forwards a chat message notification.
func (s *Service) NotifyNewMessage(ctx context.Context, chatID int64, senderName, preview string) {
	msg := fmt.Sprintf("%s: %s", senderName, truncate(preview, 100))
	go s.sendWithCooldown(ctx, chatID, msg)
}

// NotifyDailyCircle sends the daily "Your Circle is ready" notification.
func (s *Service) NotifyDailyCircle(ctx context.Context, chatID int64, count int) {
	msg := fmt.Sprintf("Your Circle is ready! %d new profiles curated for you today.", count)

	// Add inline keyboard button to open Mini App
	keyboard := InlineKeyboard{
		InlineKeyboard: [][]InlineButton{
			{
				{Text: "View Your Circle", WebApp: &WebAppInfo{URL: ""}}, // URL set at runtime
			},
		},
	}

	go s.sendMessageWithKeyboard(ctx, chatID, msg, keyboard)
}

// NotifyWeeklySummary sends the weekly engagement summary.
func (s *Service) NotifyWeeklySummary(ctx context.Context, chatID int64, matches, messages int) {
	msg := fmt.Sprintf("This week: %d matches, %d messages. Keep going!", matches, messages)
	go s.sendWithCooldown(ctx, chatID, msg)
}

// NotifyChatNudge sends a "don't let this fade" nudge.
func (s *Service) NotifyChatNudge(ctx context.Context, chatID int64, matchName string) {
	msg := fmt.Sprintf("Don't let this connection fade! Send %s a message.", matchName)
	go s.sendWithCooldown(ctx, chatID, msg)
}

// NotifyChatExpiring warns about an expiring conversation.
func (s *Service) NotifyChatExpiring(ctx context.Context, chatID int64, matchName string, daysLeft int) {
	msg := fmt.Sprintf("Your chat with %s expires in %d days. Don't lose this connection!", matchName, daysLeft)
	go s.sendWithCooldown(ctx, chatID, msg)
}

// NotifyVerificationReminder reminds unverified users to verify.
func (s *Service) NotifyVerificationReminder(ctx context.Context, chatID int64) {
	msg := "Get verified and stand out! It takes 30 seconds."
	go s.sendWithCooldown(ctx, chatID, msg)
}

// NotifyProfileIncomplete reminds users to complete their profile.
func (s *Service) NotifyProfileIncomplete(ctx context.Context, chatID int64, percent int) {
	msg := fmt.Sprintf("Your profile is %d%% complete. Add more to get better matches!", percent)
	go s.sendWithCooldown(ctx, chatID, msg)
}

// SendBotMessage sends a raw text message via the Telegram Bot API.
func (s *Service) SendBotMessage(ctx context.Context, chatID int64, text string) error {
	return s.sendMessage(ctx, chatID, text)
}

// sendWithCooldown sends a message only if the notification cooldown has passed.
func (s *Service) sendWithCooldown(ctx context.Context, chatID int64, text string) {
	cooldownKey := fmt.Sprintf("notif_cooldown:%d", chatID)
	cooldownMinutes := s.config.GetInt(ctx, "notification_cooldown_minutes", 5)

	// Check cooldown
	exists, _ := s.redis.Exists(ctx, cooldownKey).Result()
	if exists > 0 {
		// Queue it instead of dropping — store for batch delivery
		s.redis.RPush(ctx, fmt.Sprintf("notif_queue:%d", chatID), text)
		return
	}

	// Send and set cooldown
	s.sendMessage(ctx, chatID, text)
	s.redis.Set(ctx, cooldownKey, "1", time.Duration(cooldownMinutes)*time.Minute)
}

func (s *Service) sendMessage(ctx context.Context, chatID int64, text string) error {
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf(telegramAPI, s.botToken, "sendMessage")

	resp, err := s.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("sending telegram message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("telegram API returned %d", resp.StatusCode)
	}

	return nil
}

func (s *Service) sendMessageWithKeyboard(ctx context.Context, chatID int64, text string, keyboard InlineKeyboard) error {
	payload := map[string]interface{}{
		"chat_id":      chatID,
		"text":         text,
		"parse_mode":   "HTML",
		"reply_markup": keyboard,
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf(telegramAPI, s.botToken, "sendMessage")

	resp, err := s.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Telegram Bot API types

type InlineKeyboard struct {
	InlineKeyboard [][]InlineButton `json:"inline_keyboard"`
}

type InlineButton struct {
	Text    string      `json:"text"`
	URL     string      `json:"url,omitempty"`
	WebApp  *WebAppInfo `json:"web_app,omitempty"`
}

type WebAppInfo struct {
	URL string `json:"url"`
}
