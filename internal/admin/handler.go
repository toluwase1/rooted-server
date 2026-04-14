package admin

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rooted-dating/rooted-server/internal/shared/config"
	"github.com/rooted-dating/rooted-server/internal/shared/middleware"
	"github.com/rooted-dating/rooted-server/internal/user"
)

type Handler struct {
	db          *pgxpool.Pool
	dynConfig   *config.DynamicConfig
	userService *user.Service
}

func NewHandler(db *pgxpool.Pool, dynConfig *config.DynamicConfig, userService *user.Service) *Handler {
	return &Handler{db: db, dynConfig: dynConfig, userService: userService}
}

func (h *Handler) RegisterRoutes(admin fiber.Router) {
	admin.Get("/stats", h.GetStats)
	admin.Get("/config", h.GetConfig)
	admin.Put("/config/:key", h.UpdateConfig)
	admin.Get("/config/history/:key", h.GetConfigHistory)
	admin.Get("/features", h.GetFeatures)
	admin.Put("/features/:key", h.UpdateFeature)
	admin.Get("/reports", h.GetReports)
	admin.Put("/reports/:id", h.ReviewReport)
}

func (h *Handler) getUser(c *fiber.Ctx) (*user.User, error) {
	tgUser := c.Locals("telegram_user").(middleware.TelegramUser)
	u, _, err := h.userService.FindOrCreateUser(c.Context(), tgUser.ID, tgUser.Username)
	return u, err
}

func (h *Handler) GetStats(c *fiber.Ctx) error {
	var totalUsers, activeUsers, totalMatches, pendingReports int
	h.db.QueryRow(c.Context(), "SELECT COUNT(*) FROM users").Scan(&totalUsers)
	h.db.QueryRow(c.Context(), "SELECT COUNT(*) FROM users WHERE status = 'active'").Scan(&activeUsers)
	h.db.QueryRow(c.Context(), "SELECT COUNT(*) FROM matches WHERE status = 'active'").Scan(&totalMatches)
	h.db.QueryRow(c.Context(), "SELECT COUNT(*) FROM reports WHERE status = 'pending'").Scan(&pendingReports)

	return c.JSON(fiber.Map{
		"total_users": totalUsers, "active_users": activeUsers,
		"total_matches": totalMatches, "pending_reports": pendingReports,
	})
}

func (h *Handler) GetConfig(c *fiber.Ctx) error {
	rows, err := h.db.Query(c.Context(), `
		SELECT key, value, category, description, value_type, updated_at
		FROM admin_config ORDER BY category, key
	`)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to load config"})
	}
	defer rows.Close()

	var configs []fiber.Map
	for rows.Next() {
		var key, category, description, valueType string
		var value []byte
		var updatedAt interface{}
		rows.Scan(&key, &value, &category, &description, &valueType, &updatedAt)
		configs = append(configs, fiber.Map{
			"key": key, "value": string(value), "category": category,
			"description": description, "value_type": valueType, "updated_at": updatedAt,
		})
	}
	return c.JSON(configs)
}

func (h *Handler) UpdateConfig(c *fiber.Ctx) error {
	key := c.Params("key")
	var body struct {
		Value  interface{} `json:"value"`
		Reason string      `json:"reason"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	if body.Reason == "" {
		return c.Status(400).JSON(fiber.Map{"error": "reason is required"})
	}

	u, _ := h.getUser(c)

	var oldValue []byte
	h.db.QueryRow(c.Context(), "SELECT value FROM admin_config WHERE key = $1", key).Scan(&oldValue)

	newValue, _ := json.Marshal(body.Value)
	h.db.Exec(c.Context(),
		"UPDATE admin_config SET value = $2, updated_by = $3, updated_at = NOW() WHERE key = $1",
		key, newValue, u.ID)

	h.db.Exec(c.Context(), `
		INSERT INTO admin_config_audit (config_key, old_value, new_value, changed_by, reason)
		VALUES ($1, $2, $3, $4, $5)
	`, key, oldValue, newValue, u.ID, body.Reason)

	h.dynConfig.Invalidate(c.Context(), key)
	return c.JSON(fiber.Map{"status": "updated", "key": key})
}

func (h *Handler) GetConfigHistory(c *fiber.Ctx) error {
	rows, err := h.db.Query(c.Context(), `
		SELECT old_value, new_value, changed_by, reason, created_at
		FROM admin_config_audit WHERE config_key = $1
		ORDER BY created_at DESC LIMIT 50
	`, c.Params("key"))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to load history"})
	}
	defer rows.Close()

	var history []fiber.Map
	for rows.Next() {
		var oldVal, newVal []byte
		var changedBy, reason string
		var createdAt interface{}
		rows.Scan(&oldVal, &newVal, &changedBy, &reason, &createdAt)
		history = append(history, fiber.Map{
			"old_value": string(oldVal), "new_value": string(newVal),
			"changed_by": changedBy, "reason": reason, "created_at": createdAt,
		})
	}
	return c.JSON(history)
}

func (h *Handler) GetFeatures(c *fiber.Ctx) error {
	rows, _ := h.db.Query(c.Context(), `
		SELECT key, enabled, description, rollout_percent, target_regions, target_plans, updated_at
		FROM feature_flags ORDER BY key
	`)
	defer rows.Close()

	var flags []fiber.Map
	for rows.Next() {
		var key, description string
		var enabled bool
		var rolloutPercent int
		var targetRegions, targetPlans []string
		var updatedAt interface{}
		rows.Scan(&key, &enabled, &description, &rolloutPercent, &targetRegions, &targetPlans, &updatedAt)
		flags = append(flags, fiber.Map{
			"key": key, "enabled": enabled, "description": description,
			"rollout_percent": rolloutPercent, "target_regions": targetRegions,
			"target_plans": targetPlans, "updated_at": updatedAt,
		})
	}
	return c.JSON(flags)
}

func (h *Handler) UpdateFeature(c *fiber.Ctx) error {
	var body struct {
		Enabled        *bool    `json:"enabled"`
		RolloutPercent *int     `json:"rollout_percent"`
		TargetRegions  []string `json:"target_regions"`
		TargetPlans    []string `json:"target_plans"`
	}
	c.BodyParser(&body)

	if body.Enabled != nil {
		h.db.Exec(c.Context(), "UPDATE feature_flags SET enabled = $2, updated_at = NOW() WHERE key = $1",
			c.Params("key"), *body.Enabled)
	}
	if body.RolloutPercent != nil {
		h.db.Exec(c.Context(), "UPDATE feature_flags SET rollout_percent = $2, updated_at = NOW() WHERE key = $1",
			c.Params("key"), *body.RolloutPercent)
	}

	h.dynConfig.InvalidateFeature(c.Context(), c.Params("key"))
	return c.JSON(fiber.Map{"status": "updated"})
}

func (h *Handler) GetReports(c *fiber.Ctx) error {
	rows, _ := h.db.Query(c.Context(), `
		SELECT r.id, r.reporter_id, r.reported_id, r.category, r.description,
		       r.status, r.created_at, p.first_name
		FROM reports r
		LEFT JOIN profiles p ON p.user_id = r.reported_id
		WHERE r.status = 'pending'
		ORDER BY r.created_at ASC LIMIT 50
	`)
	defer rows.Close()

	var reports []fiber.Map
	for rows.Next() {
		var id, reporterID, reportedID, category, description, status string
		var createdAt interface{}
		var firstName *string
		rows.Scan(&id, &reporterID, &reportedID, &category, &description,
			&status, &createdAt, &firstName)
		reports = append(reports, fiber.Map{
			"id": id, "reporter_id": reporterID, "reported_id": reportedID,
			"category": category, "description": description, "status": status,
			"created_at": createdAt, "reported_name": firstName,
		})
	}
	return c.JSON(fiber.Map{"reports": reports})
}

func (h *Handler) ReviewReport(c *fiber.Ctx) error {
	u, _ := h.getUser(c)
	var body struct {
		Action string `json:"action"`
	}
	c.BodyParser(&body)

	h.db.Exec(c.Context(), `
		UPDATE reports SET status = 'reviewed', reviewed_by = $2,
		       reviewed_at = NOW(), action_taken = $3
		WHERE id = $1
	`, c.Params("id"), u.ID, body.Action)

	if body.Action == "ban" || body.Action == "suspension" {
		var reportedID string
		h.db.QueryRow(c.Context(),
			"SELECT reported_id FROM reports WHERE id = $1", c.Params("id"),
		).Scan(&reportedID)
		if reportedID != "" {
			h.userService.PauseProfile(c.Context(), reportedID)
		}
	}

	return c.JSON(fiber.Map{"status": "reviewed"})
}
