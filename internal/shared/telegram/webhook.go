package telegram

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rooted-dating/rooted-server/internal/chat"
	"github.com/rooted-dating/rooted-server/internal/user"
)

// WebhookHandler processes incoming Telegram bot updates.
type WebhookHandler struct {
	bot         *Bot
	userService *user.Service
	chatService *chat.Service
	webAppURL   string
}

func NewWebhookHandler(bot *Bot, userSvc *user.Service, chatSvc *chat.Service, webAppURL string) *WebhookHandler {
	return &WebhookHandler{
		bot:         bot,
		userService: userSvc,
		chatService: chatSvc,
		webAppURL:   webAppURL,
	}
}

// Handle processes an incoming Telegram update.
func (h *WebhookHandler) Handle(c *fiber.Ctx) error {
	var update Update
	if err := c.BodyParser(&update); err != nil {
		return c.SendStatus(200) // Always return 200 to Telegram
	}

	ctx := context.Background()

	// Handle pre-checkout queries (Star payments) — must respond within 10 seconds
	if update.PreCheckoutQuery != nil {
		h.bot.AnswerPreCheckoutQuery(ctx, update.PreCheckoutQuery.ID, true, "")
		return c.SendStatus(200)
	}

	// Handle callback queries (button presses)
	if update.CallbackQuery != nil {
		h.handleCallback(ctx, update.CallbackQuery)
		return c.SendStatus(200)
	}

	// Handle messages
	if update.Message != nil {
		h.handleMessage(ctx, update.Message)
	}

	return c.SendStatus(200)
}

func (h *WebhookHandler) handleMessage(ctx context.Context, msg *BotMessage) {
	if msg.Chat.Type != "private" {
		return // Only handle private messages
	}

	chatID := msg.Chat.ID
	telegramID := msg.From.ID

	// Ensure user exists
	u, isNew, err := h.userService.FindOrCreateUser(ctx, telegramID, msg.From.Username)
	if err != nil {
		return
	}

	// Register chat routing (link telegram chat_id to internal user)
	h.chatService.RegisterChatRouting(ctx, chatID, u.ID)

	// Handle commands
	if msg.Text != "" && strings.HasPrefix(msg.Text, "/") {
		h.handleCommand(ctx, chatID, u, msg.Text, isNew)
		return
	}

	// Handle regular messages — relay to active conversation
	h.relayMessage(ctx, chatID, msg)
}

func (h *WebhookHandler) handleCommand(ctx context.Context, chatID int64, u *user.User, text string, isNew bool) {
	cmd := strings.Split(strings.TrimPrefix(text, "/"), " ")[0]
	cmd = strings.Split(cmd, "@")[0] // Remove @botname suffix

	switch cmd {
	case "start":
		if isNew {
			h.bot.SendMessageWithWebApp(ctx, chatID,
				"Welcome to Rooted! 🌍\n\nDating, rooted in what matters.\n\nLet's create your profile to get started.",
				"Create Your Profile",
				h.webAppURL+"/onboarding",
			)
		} else {
			h.bot.SendMessageWithWebApp(ctx, chatID,
				"Welcome back! Your Circle is waiting.",
				"Open Rooted",
				h.webAppURL,
			)
		}

	case "profile":
		h.bot.SendMessageWithWebApp(ctx, chatID,
			"View or edit your profile:",
			"Open Profile",
			h.webAppURL+"/profile",
		)

	case "matches":
		h.bot.SendMessageWithWebApp(ctx, chatID,
			"View your matches:",
			"Open Matches",
			h.webAppURL+"/matches",
		)

	case "explore":
		h.bot.SendMessageWithWebApp(ctx, chatID,
			"Browse profiles:",
			"Open Explore",
			h.webAppURL+"/explore",
		)

	case "settings":
		h.bot.SendMessageWithWebApp(ctx, chatID,
			"Manage your settings:",
			"Open Settings",
			h.webAppURL+"/settings",
		)

	case "help":
		helpText := `<b>Rooted Help</b>

/start — Open Rooted
/profile — View or edit your profile
/matches — View your matches
/explore — Browse profiles
/settings — Manage settings

<b>How chat works:</b>
When you match with someone, I'll notify you. You can chat right here — I'll relay messages between you and your match. Neither of you will see each other's phone number or Telegram username.

<b>Need help?</b>
Use /settings to report issues or contact support.`

		h.bot.SendMessage(ctx, chatID, helpText)

	case "chat":
		// Show list of active conversations to switch between
		h.showConversationList(ctx, chatID, u.ID)

	default:
		h.bot.SendMessage(ctx, chatID, "I don't recognize that command. Try /help for available commands.")
	}
}

func (h *WebhookHandler) relayMessage(ctx context.Context, senderChatID int64, msg *BotMessage) {
	var contentType, content string

	switch {
	case msg.Text != "":
		contentType = "text"
		content = msg.Text

	case len(msg.Photo) > 0:
		contentType = "photo"
		// Use the largest photo variant
		content = msg.Photo[len(msg.Photo)-1].FileID

	case msg.Voice != nil:
		contentType = "voice_note"
		content = msg.Voice.FileID

	default:
		h.bot.SendMessage(ctx, senderChatID, "Sorry, I can only relay text messages, photos, and voice notes.")
		return
	}

	// Relay through chat service
	relayedMsg, recipientID, err := h.chatService.RelayMessage(ctx, senderChatID, contentType, content)
	if err != nil {
		if strings.Contains(err.Error(), "no active conversation") || strings.Contains(err.Error(), "no active chat routing") {
			h.bot.SendMessage(ctx, senderChatID,
				"You don't have an active chat. Match with someone first, then I'll relay your messages!\n\nUse /explore to browse profiles.")
		} else if strings.Contains(err.Error(), "rate limit") {
			h.bot.SendMessage(ctx, senderChatID, "You're sending messages too fast. Please slow down.")
		} else {
			h.bot.SendMessage(ctx, senderChatID, "Something went wrong. Please try again.")
		}
		return
	}

	// Get recipient's Telegram chat ID
	recipientChatID, err := h.chatService.GetRecipientTelegramChatID(ctx, recipientID)
	if err != nil {
		return
	}

	// Forward the message to the recipient via bot
	switch relayedMsg.ContentType {
	case "text":
		// Get sender's first name for display
		senderName := msg.From.FirstName
		h.bot.SendMessage(ctx, recipientChatID, fmt.Sprintf("<b>%s:</b> %s", senderName, content))

	case "photo":
		h.bot.ForwardPhoto(ctx, recipientChatID, content, msg.From.FirstName)

	case "voice_note":
		h.bot.SendMessage(ctx, recipientChatID, fmt.Sprintf("<b>%s</b> sent a voice note:", msg.From.FirstName))
		h.bot.ForwardVoice(ctx, recipientChatID, content)
	}
}

func (h *WebhookHandler) showConversationList(ctx context.Context, chatID int64, userID string) {
	convs, err := h.chatService.GetUserConversations(ctx, userID)
	if err != nil || len(convs) == 0 {
		h.bot.SendMessage(ctx, chatID, "You don't have any active conversations yet. Match with someone first!")
		return
	}

	// Build inline keyboard with conversation options
	var buttons [][]InlineButton
	for _, conv := range convs {
		// Determine the other person's name
		otherID := conv.UserBID
		if userID == conv.UserBID {
			otherID = conv.UserAID
		}

		// Get other person's profile for name
		profile, _ := h.userService.GetProfile(ctx, otherID)
		name := "Match"
		if profile != nil {
			name = profile.FirstName
		}

		buttons = append(buttons, []InlineButton{
			{Text: name, CallbackData: "switch_chat:" + conv.ID},
		})
	}

	h.bot.SendMessageWithButtons(ctx, chatID,
		"Switch to a conversation:",
		buttons,
	)
}

func (h *WebhookHandler) handleCallback(ctx context.Context, query *CallbackQuery) {
	chatID := query.Message.Chat.ID
	data := query.Data

	if strings.HasPrefix(data, "switch_chat:") {
		convID := strings.TrimPrefix(data, "switch_chat:")
		h.chatService.SetActiveConversation(ctx, chatID, convID)
		h.bot.AnswerCallbackQuery(ctx, query.ID, "Switched! Send your message.")
		h.bot.SendMessage(ctx, chatID, "Chat switched. Your next messages will go to this match.")
	}
}
