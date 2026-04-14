package middleware

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RequestLogger logs requests as structured JSON to stdout.
// Cloud Run picks these up automatically into Cloud Logging.
// No database writes — query Cloud Logging API from admin dashboard.
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		latencyMs := time.Since(start).Milliseconds()
		status := c.Response().StatusCode()
		method := c.Method()
		path := c.Path()

		if path == "/health" {
			return err
		}

		// Structured JSON log — Cloud Logging parses jsonPayload automatically
		entry := map[string]interface{}{
			"severity":    severityFromStatus(status),
			"method":      method,
			"path":        path,
			"status":      status,
			"latency_ms":  latencyMs,
			"ip":          c.IP(),
		}

		if tgUser, ok := c.Locals("telegram_user").(TelegramUser); ok {
			entry["telegram_id"] = tgUser.ID
		}

		if status >= 400 {
			body := c.Response().Body()
			if len(body) > 0 && len(body) < 1000 {
				entry["response_body"] = string(body)
			}
		}

		data, _ := json.Marshal(entry)
		log.Println(string(data))

		return err
	}
}

func severityFromStatus(status int) string {
	if status >= 500 {
		return "ERROR"
	}
	if status >= 400 {
		return "WARNING"
	}
	return "INFO"
}
