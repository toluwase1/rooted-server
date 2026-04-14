package moderation

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rooted-dating/rooted-server/internal/shared/config"
	"github.com/rooted-dating/rooted-server/internal/shared/middleware"
	"github.com/rooted-dating/rooted-server/internal/user"
)

type Handler struct {
	db          *pgxpool.Pool
	config      *config.DynamicConfig
	userService *user.Service
}

func NewHandler(db *pgxpool.Pool, cfg *config.DynamicConfig, us *user.Service) *Handler {
	return &Handler{db: db, config: cfg, userService: us}
}

func (h *Handler) RegisterRoutes(api fiber.Router) {
	api.Post("/report", h.Report)
}

func (h *Handler) Report(c *fiber.Ctx) error {
	tgUser, ok := c.Locals("telegram_user").(middleware.TelegramUser)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	u, _, err := h.userService.FindOrCreateUser(c.Context(), tgUser.ID, tgUser.Username)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	var body struct {
		ReportedID  string `json:"reported_id"`
		Category    string `json:"category"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if body.ReportedID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "reported_id is required"})
	}
	if body.Category == "" {
		return c.Status(400).JSON(fiber.Map{"error": "category is required"})
	}
	validCategories := map[string]bool{
		"harassment": true, "fake_profile": true, "scam": true,
		"inappropriate": true, "underage": true,
	}
	if !validCategories[body.Category] {
		return c.Status(400).JSON(fiber.Map{"error": "invalid category"})
	}
	if body.ReportedID == u.ID {
		return c.Status(400).JSON(fiber.Map{"error": "cannot report yourself"})
	}

	_, err = h.db.Exec(c.Context(), `
		INSERT INTO reports (reporter_id, reported_id, category, description)
		VALUES ($1, $2, $3, $4)
	`, u.ID, body.ReportedID, body.Category, body.Description)
	if err != nil {
		log.Printf("ERROR Report user=%s reported=%s: %v", u.ID, body.ReportedID, err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to submit report"})
	}

	// Auto-block
	h.userService.BlockUser(c.Context(), u.ID, body.ReportedID)

	// Check auto-ban
	var reportCount int
	h.db.QueryRow(c.Context(),
		"SELECT COUNT(DISTINCT reporter_id) FROM reports WHERE reported_id = $1 AND status = 'pending'",
		body.ReportedID,
	).Scan(&reportCount)

	threshold := h.config.GetInt(c.Context(), "auto_ban_report_threshold", 3)
	if reportCount >= threshold {
		h.userService.PauseProfile(c.Context(), body.ReportedID)
		log.Printf("WARN auto-suspended user=%s after %d reports", body.ReportedID, reportCount)
	}

	return c.JSON(fiber.Map{"status": "reported"})
}
