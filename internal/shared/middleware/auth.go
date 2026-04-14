package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

// TelegramAuth validates Telegram Mini App init data and extracts the user.
func TelegramAuth(botToken string, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		initData := c.Get("X-Telegram-Init-Data")
		if initData == "" {
			auth := c.Get("Authorization")
			if strings.HasPrefix(auth, "tma ") {
				initData = strings.TrimPrefix(auth, "tma ")
			}
		}

		if initData == "" {
			return c.Status(401).JSON(fiber.Map{"error": "missing Telegram init data"})
		}

		// Check session cache
		ctx := c.Context()
		sessionKey := fmt.Sprintf("session:%x", sha256Sum(initData))
		cached, err := redisClient.Get(ctx, sessionKey).Bytes()
		if err == nil {
			var user TelegramUser
			if json.Unmarshal(cached, &user) == nil {
				c.Locals("telegram_user", user)
				return c.Next()
			}
		}

		// Validate init data
		valid, err := validateInitData(initData, botToken)
		if err != nil || !valid {
			return c.Status(401).JSON(fiber.Map{"error": "invalid Telegram init data"})
		}

		// Extract user
		user, err := extractUser(initData)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "cannot extract user"})
		}

		// Cache session
		data, _ := json.Marshal(user)
		redisClient.Set(ctx, sessionKey, data, 1*time.Hour)

		c.Locals("telegram_user", user)
		return c.Next()
	}
}

// AdminOnly restricts access to configured admin Telegram IDs.
func AdminOnly(adminIDs []int64) fiber.Handler {
	adminSet := make(map[int64]bool)
	for _, id := range adminIDs {
		adminSet[id] = true
	}

	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("telegram_user").(TelegramUser)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
		}
		if !adminSet[user.ID] {
			return c.Status(403).JSON(fiber.Map{"error": "admin access required"})
		}
		return c.Next()
	}
}

// validateInitData verifies Telegram Mini App init data.
// See: https://core.telegram.org/bots/webapps#validating-data-received-via-the-mini-app
func validateInitData(initData string, botToken string) (bool, error) {
	parsed, err := url.ParseQuery(initData)
	if err != nil {
		return false, fmt.Errorf("parsing init data: %w", err)
	}

	receivedHash := parsed.Get("hash")
	if receivedHash == "" {
		return false, fmt.Errorf("missing hash")
	}

	parsed.Del("hash")

	var keys []string
	for k := range parsed {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, parsed.Get(k)))
	}
	dataCheckString := strings.Join(parts, "\n")

	secretKey := hmacSHA256([]byte("WebAppData"), []byte(botToken))
	hash := hmacSHA256(secretKey, []byte(dataCheckString))
	computedHash := hex.EncodeToString(hash)

	return computedHash == receivedHash, nil
}

func extractUser(initData string) (TelegramUser, error) {
	var user TelegramUser
	parsed, err := url.ParseQuery(initData)
	if err != nil {
		return user, err
	}

	userJSON := parsed.Get("user")
	if userJSON != "" {
		json.Unmarshal([]byte(userJSON), &user)
	}

	return user, nil
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func sha256Sum(s string) []byte {
	h := sha256.Sum256([]byte(s))
	return h[:]
}
