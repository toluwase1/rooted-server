package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const apiBase = "https://api.telegram.org/bot%s/%s"

// Bot wraps Telegram Bot API calls.
type Bot struct {
	token  string
	client *http.Client
}

func NewBot(token string) *Bot {
	return &Bot{
		token:  token,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Update represents an incoming Telegram update (message, callback, etc.).
type Update struct {
	UpdateID      int64          `json:"update_id"`
	Message       *BotMessage    `json:"message,omitempty"`
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
	PreCheckoutQuery *PreCheckoutQuery `json:"pre_checkout_query,omitempty"`
}

type BotMessage struct {
	MessageID int64    `json:"message_id"`
	From      *BotUser `json:"from"`
	Chat      *BotChat `json:"chat"`
	Date      int64    `json:"date"`
	Text      string   `json:"text,omitempty"`
	Photo     []PhotoSize `json:"photo,omitempty"`
	Voice     *Voice   `json:"voice,omitempty"`
	Document  *Document `json:"document,omitempty"`
}

type BotUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
}

type BotChat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"` // private, group, supergroup, channel
}

type CallbackQuery struct {
	ID      string      `json:"id"`
	From    *BotUser    `json:"from"`
	Message *BotMessage `json:"message,omitempty"`
	Data    string      `json:"data"`
}

type PreCheckoutQuery struct {
	ID               string   `json:"id"`
	From             *BotUser `json:"from"`
	Currency         string   `json:"currency"`
	TotalAmount      int      `json:"total_amount"`
	InvoicePayload   string   `json:"invoice_payload"`
}

type PhotoSize struct {
	FileID   string `json:"file_id"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	FileSize int    `json:"file_size,omitempty"`
}

type Voice struct {
	FileID   string `json:"file_id"`
	Duration int    `json:"duration"`
	FileSize int    `json:"file_size,omitempty"`
}

type Document struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
	FileSize int    `json:"file_size,omitempty"`
}

// SendMessage sends a text message.
func (b *Bot) SendMessage(ctx context.Context, chatID int64, text string) error {
	return b.callAPI(ctx, "sendMessage", map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	})
}

// SendMessageWithButtons sends a message with inline keyboard buttons.
func (b *Bot) SendMessageWithButtons(ctx context.Context, chatID int64, text string, buttons [][]InlineButton) error {
	return b.callAPI(ctx, "sendMessage", map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
		"reply_markup": map[string]interface{}{
			"inline_keyboard": buttons,
		},
	})
}

// SendMessageWithWebApp sends a message with a button that opens the Mini App.
func (b *Bot) SendMessageWithWebApp(ctx context.Context, chatID int64, text, buttonText, webAppURL string) error {
	return b.callAPI(ctx, "sendMessage", map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
		"reply_markup": map[string]interface{}{
			"inline_keyboard": [][]map[string]interface{}{
				{
					{
						"text":    buttonText,
						"web_app": map[string]string{"url": webAppURL},
					},
				},
			},
		},
	})
}

// ForwardPhoto forwards a photo by file_id to another chat.
func (b *Bot) ForwardPhoto(ctx context.Context, chatID int64, fileID, caption string) error {
	return b.callAPI(ctx, "sendPhoto", map[string]interface{}{
		"chat_id": chatID,
		"photo":   fileID,
		"caption": caption,
	})
}

// ForwardVoice forwards a voice note by file_id to another chat.
func (b *Bot) ForwardVoice(ctx context.Context, chatID int64, fileID string) error {
	return b.callAPI(ctx, "sendVoice", map[string]interface{}{
		"chat_id": chatID,
		"voice":   fileID,
	})
}

// AnswerCallbackQuery responds to a callback query (button press).
func (b *Bot) AnswerCallbackQuery(ctx context.Context, queryID, text string) error {
	return b.callAPI(ctx, "answerCallbackQuery", map[string]interface{}{
		"callback_query_id": queryID,
		"text":              text,
	})
}

// AnswerPreCheckoutQuery responds to a pre-checkout query (Star payment).
func (b *Bot) AnswerPreCheckoutQuery(ctx context.Context, queryID string, ok bool, errorMessage string) error {
	payload := map[string]interface{}{
		"pre_checkout_query_id": queryID,
		"ok":                    ok,
	}
	if !ok {
		payload["error_message"] = errorMessage
	}
	return b.callAPI(ctx, "answerPreCheckoutQuery", payload)
}

// SendInvoice sends a Star payment invoice.
func (b *Bot) SendInvoice(ctx context.Context, chatID int64, title, description, payload string, priceStars int) error {
	return b.callAPI(ctx, "sendInvoice", map[string]interface{}{
		"chat_id":     chatID,
		"title":       title,
		"description": description,
		"payload":     payload,
		"currency":    "XTR", // Telegram Stars
		"prices": []map[string]interface{}{
			{"label": title, "amount": priceStars},
		},
	})
}

// SetWebhook configures the Telegram webhook URL.
func (b *Bot) SetWebhook(ctx context.Context, webhookURL string) error {
	return b.callAPI(ctx, "setWebhook", map[string]interface{}{
		"url":             webhookURL,
		"allowed_updates": []string{"message", "callback_query", "pre_checkout_query"},
	})
}

// SetBotCommands registers the bot command menu.
func (b *Bot) SetBotCommands(ctx context.Context) error {
	commands := []map[string]string{
		{"command": "start", "description": "Start Rooted / Open profile"},
		{"command": "profile", "description": "View or edit your profile"},
		{"command": "matches", "description": "View your matches"},
		{"command": "explore", "description": "Browse profiles"},
		{"command": "settings", "description": "Manage settings"},
		{"command": "help", "description": "Help and FAQ"},
	}

	return b.callAPI(ctx, "setMyCommands", map[string]interface{}{
		"commands": commands,
	})
}

type InlineButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
	URL          string `json:"url,omitempty"`
}

func (b *Bot) callAPI(ctx context.Context, method string, payload map[string]interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling payload: %w", err)
	}

	url := fmt.Sprintf(apiBase, b.token, method)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("calling telegram API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		return fmt.Errorf("telegram API error %d: %v", resp.StatusCode, result)
	}

	return nil
}
