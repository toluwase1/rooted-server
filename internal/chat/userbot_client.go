package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// UserbotClient communicates with the Python userbot service.
type UserbotClient struct {
	baseURL string
	client  *http.Client
}

func NewUserbotClient(baseURL string) *UserbotClient {
	if baseURL == "" {
		return nil
	}
	return &UserbotClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type CreateGroupResponse struct {
	GroupID    int64    `json:"group_id"`
	InviteLink string  `json:"invite_link"`
	Added     []string `json:"added"`
}

// CreateGroup asks the userbot to create a Telegram supergroup for a match.
func (c *UserbotClient) CreateGroup(ctx context.Context, conversationID string, userATelegramID, userBTelegramID int64, userAName, userBName string) (*CreateGroupResponse, error) {
	if c == nil {
		return nil, fmt.Errorf("userbot not configured")
	}

	body, _ := json.Marshal(map[string]interface{}{
		"conversation_id":     conversationID,
		"user_a_telegram_id":  userATelegramID,
		"user_b_telegram_id":  userBTelegramID,
		"user_a_name":         userAName,
		"user_b_name":         userBName,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/create-group", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("userbot request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("userbot returned %d", resp.StatusCode)
	}

	var result CreateGroupResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return &result, nil
}

// SendToGroup sends a Mini App message to the Telegram group.
func (c *UserbotClient) SendToGroup(ctx context.Context, telegramGroupID int64, senderName, content, contentType string) error {
	if c == nil {
		return nil // silently skip if userbot not configured
	}

	body, _ := json.Marshal(map[string]interface{}{
		"telegram_group_id": telegramGroupID,
		"sender_name":       senderName,
		"content":           content,
		"content_type":      contentType,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/send-message", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("WARN userbot send failed: %v", err)
		return nil // don't fail the Mini App message
	}
	defer resp.Body.Close()

	return nil
}
