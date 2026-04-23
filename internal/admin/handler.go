package admin

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rooted-dating/rooted-server/internal/logging"
	"github.com/rooted-dating/rooted-server/internal/media"
	"github.com/rooted-dating/rooted-server/internal/shared/config"
)

type Handler struct {
	db           *pgxpool.Pool
	dynConfig    *config.DynamicConfig
	logClient    *logging.Client
	mediaService *media.Service
}

func NewHandler(db *pgxpool.Pool, dynConfig *config.DynamicConfig, logClient *logging.Client, mediaService *media.Service) *Handler {
	return &Handler{db: db, dynConfig: dynConfig, logClient: logClient, mediaService: mediaService}
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
	admin.Get("/users", h.ListUsers)
	admin.Get("/users/:id", h.GetUser)
	admin.Put("/users/:id/status", h.UpdateUserStatus)
	admin.Get("/logs", h.GetLogs)
	admin.Get("/logs/stats", h.GetLogStats)
}

func (h *Handler) GetStats(c *fiber.Ctx) error {
	var totalUsers, activeUsers, verifiedUsers, totalMatches, totalMessages, pendingReports int
	var todaySignups, todayMatches int
	var totalConversations, activeGroups, telegramMessages, miniappMessages int

	h.db.QueryRow(c.Context(), "SELECT COUNT(*) FROM users").Scan(&totalUsers)
	h.db.QueryRow(c.Context(), "SELECT COUNT(*) FROM users WHERE status = 'active'").Scan(&activeUsers)
	h.db.QueryRow(c.Context(), "SELECT COUNT(*) FROM users WHERE verification = 'photo_verified'").Scan(&verifiedUsers)
	h.db.QueryRow(c.Context(), "SELECT COUNT(*) FROM matches WHERE status = 'active'").Scan(&totalMatches)
	h.db.QueryRow(c.Context(), "SELECT COUNT(*) FROM messages").Scan(&totalMessages)
	h.db.QueryRow(c.Context(), "SELECT COUNT(*) FROM reports WHERE status = 'pending'").Scan(&pendingReports)
	h.db.QueryRow(c.Context(), "SELECT COUNT(*) FROM users WHERE created_at >= CURRENT_DATE").Scan(&todaySignups)
	h.db.QueryRow(c.Context(), "SELECT COUNT(*) FROM matches WHERE matched_at >= CURRENT_DATE").Scan(&todayMatches)

	// Chat & userbot stats
	h.db.QueryRow(c.Context(), "SELECT COUNT(*) FROM conversations").Scan(&totalConversations)
	h.db.QueryRow(c.Context(), "SELECT COUNT(*) FROM conversations WHERE telegram_group_id IS NOT NULL").Scan(&activeGroups)
	h.db.QueryRow(c.Context(), "SELECT COUNT(*) FROM messages WHERE source = 'telegram'").Scan(&telegramMessages)
	h.db.QueryRow(c.Context(), "SELECT COUNT(*) FROM messages WHERE source = 'miniapp'").Scan(&miniappMessages)

	return c.JSON(fiber.Map{
		"total_users":         totalUsers,
		"active_users":        activeUsers,
		"verified_users":      verifiedUsers,
		"total_matches":       totalMatches,
		"total_messages":      totalMessages,
		"pending_reports":     pendingReports,
		"today_signups":       todaySignups,
		"today_matches":       todayMatches,
		"total_conversations": totalConversations,
		"active_groups":       activeGroups,
		"telegram_messages":   telegramMessages,
		"miniapp_messages":    miniappMessages,
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
	admin := c.Locals("admin_user").(AdminUser)

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

	var oldValue []byte
	h.db.QueryRow(c.Context(), "SELECT value FROM admin_config WHERE key = $1", key).Scan(&oldValue)

	newValue, _ := json.Marshal(body.Value)
	h.db.Exec(c.Context(),
		"UPDATE admin_config SET value = $2, updated_by = $3, updated_at = NOW() WHERE key = $1",
		key, newValue, admin.ID)

	h.db.Exec(c.Context(), `
		INSERT INTO admin_config_audit (config_key, old_value, new_value, changed_by, reason)
		VALUES ($1, $2, $3, $4, $5)
	`, key, oldValue, newValue, admin.ID, body.Reason)

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
		Enabled        *bool `json:"enabled"`
		RolloutPercent *int  `json:"rollout_percent"`
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
		WHERE r.status = $1
		ORDER BY r.created_at ASC LIMIT 50
	`, c.Query("status", "pending"))
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
	admin := c.Locals("admin_user").(AdminUser)
	var body struct {
		Action string `json:"action"`
	}
	c.BodyParser(&body)

	h.db.Exec(c.Context(), `
		UPDATE reports SET status = 'reviewed', reviewed_by = $2,
		       reviewed_at = NOW(), action_taken = $3
		WHERE id = $1
	`, c.Params("id"), admin.ID, body.Action)

	if body.Action == "ban" || body.Action == "suspension" {
		var reportedID string
		h.db.QueryRow(c.Context(),
			"SELECT reported_id FROM reports WHERE id = $1", c.Params("id"),
		).Scan(&reportedID)
		if reportedID != "" {
			h.db.Exec(c.Context(), "UPDATE users SET status = 'banned', updated_at = NOW() WHERE id = $1", reportedID)
		}
	}

	return c.JSON(fiber.Map{"status": "reviewed"})
}

func (h *Handler) ListUsers(c *fiber.Ctx) error {
	search := c.Query("search", "")
	status := c.Query("status", "")
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	query := `
		SELECT u.id, u.telegram_id, u.status, u.verification, u.trust_score,
		       u.subscription, u.created_at, u.last_active_at,
		       p.first_name, p.city, p.country, p.heritage, p.completeness
		FROM users u
		LEFT JOIN profiles p ON p.user_id = u.id
		WHERE 1=1
	`
	args := []interface{}{}
	argNum := 1

	if search != "" {
		query += ` AND (p.first_name ILIKE $` + itoa(argNum) + ` OR u.telegram_id::text LIKE $` + itoa(argNum) + `)`
		args = append(args, "%"+search+"%")
		argNum++
	}
	if status != "" {
		query += ` AND u.status = $` + itoa(argNum)
		args = append(args, status)
		argNum++
	}

	query += ` ORDER BY u.created_at DESC LIMIT $` + itoa(argNum) + ` OFFSET $` + itoa(argNum+1)
	args = append(args, limit, offset)

	rows, err := h.db.Query(c.Context(), query, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to load users"})
	}
	defer rows.Close()

	var users []fiber.Map
	for rows.Next() {
		var id string
		var telegramID int64
		var userStatus, verification string
		var trustScore int
		var subscription string
		var createdAt, lastActiveAt interface{}
		var firstName, city, country *string
		var heritage []string
		var completeness *int

		rows.Scan(&id, &telegramID, &userStatus, &verification, &trustScore,
			&subscription, &createdAt, &lastActiveAt,
			&firstName, &city, &country, &heritage, &completeness)

		users = append(users, fiber.Map{
			"id": id, "telegram_id": telegramID, "status": userStatus,
			"verification": verification, "trust_score": trustScore,
			"subscription": subscription, "created_at": createdAt,
			"last_active_at": lastActiveAt, "first_name": firstName,
			"city": city, "country": country, "heritage": heritage,
			"completeness": completeness,
		})
	}

	// Get total count
	var total int
	h.db.QueryRow(c.Context(), "SELECT COUNT(*) FROM users").Scan(&total)

	return c.JSON(fiber.Map{"users": users, "total": total})
}

func (h *Handler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")

	var user fiber.Map
	var telegramID int64
	var status, verification, subscription string
	var trustScore int
	var createdAt, lastActiveAt interface{}

	err := h.db.QueryRow(c.Context(), `
		SELECT telegram_id, status, verification, trust_score, subscription, created_at, last_active_at
		FROM users WHERE id = $1
	`, id).Scan(&telegramID, &status, &verification, &trustScore, &subscription, &createdAt, &lastActiveAt)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	user = fiber.Map{
		"id": id, "telegram_id": telegramID, "status": status,
		"verification": verification, "trust_score": trustScore,
		"subscription": subscription, "created_at": createdAt,
		"last_active_at": lastActiveAt,
	}

	// Get profile
	var firstName, city, country, bio *string
	var heritage []string
	var completeness *int
	h.db.QueryRow(c.Context(), `
		SELECT first_name, city, country, heritage, bio, completeness
		FROM profiles WHERE user_id = $1
	`, id).Scan(&firstName, &city, &country, &heritage, &bio, &completeness)

	// Get photos
	photoRows, _ := h.db.Query(c.Context(), `
		SELECT id, url_thumbnail, url_medium, url_large, position, is_primary, moderation_status
		FROM photos WHERE user_id = $1 ORDER BY position
	`, id)
	var photos []fiber.Map
	if photoRows != nil {
		defer photoRows.Close()
		for photoRows.Next() {
			var pid, urlThumb, urlMed, urlLarge, modStatus string
			var pos int
			var isPrimary bool
			photoRows.Scan(&pid, &urlThumb, &urlMed, &urlLarge, &pos, &isPrimary, &modStatus)
			photos = append(photos, fiber.Map{
				"id": pid, "url_thumbnail": urlThumb, "url_medium": urlMed,
				"url_large": urlLarge, "position": pos, "is_primary": isPrimary,
				"moderation_status": modStatus,
			})
		}
	}
	if photos == nil {
		photos = []fiber.Map{}
	}

	// Enrich photo URLs with presigned URLs
	if h.mediaService != nil {
		for i, photo := range photos {
			if key, ok := photo["url_medium"].(string); ok && key != "" && !strings.HasPrefix(key, "http") {
				if url, err := h.mediaService.GetPresignedReadURL(c.Context(), key); err == nil {
					photos[i]["url_thumbnail"] = url
					photos[i]["url_medium"] = url
					photos[i]["url_large"] = url
				}
			}
		}
	}

	// Get stats
	var matchCount, messageCount, reportCount int
	h.db.QueryRow(c.Context(), "SELECT COUNT(*) FROM matches WHERE (user_a_id = $1 OR user_b_id = $1) AND status = 'active'", id).Scan(&matchCount)
	h.db.QueryRow(c.Context(), "SELECT COUNT(*) FROM messages WHERE sender_id = $1", id).Scan(&messageCount)
	h.db.QueryRow(c.Context(), "SELECT COUNT(*) FROM reports WHERE reported_id = $1", id).Scan(&reportCount)

	return c.JSON(fiber.Map{
		"user": user,
		"profile": fiber.Map{
			"first_name": firstName, "city": city, "country": country,
			"heritage": heritage, "bio": bio, "completeness": completeness,
		},
		"photos": photos,
		"stats": fiber.Map{
			"matches":  matchCount,
			"messages": messageCount,
			"reports":  reportCount,
		},
	})
}

func (h *Handler) UpdateUserStatus(c *fiber.Ctx) error {
	var body struct {
		Status string `json:"status"` // active, banned, suspended
	}
	c.BodyParser(&body)

	h.db.Exec(c.Context(), "UPDATE users SET status = $2, updated_at = NOW() WHERE id = $1",
		c.Params("id"), body.Status)

	return c.JSON(fiber.Map{"status": "updated"})
}

func (h *Handler) GetLogs(c *fiber.Ctx) error {
	if h.logClient == nil {
		return c.JSON(fiber.Map{
			"entries":     []interface{}{},
			"console_url": "https://console.cloud.google.com/logs/viewer?project=doodlegen-app-2026&resource=cloud_run_revision/service_name/rooted-api",
			"note":        "Cloud Logging client not initialized. View logs in GCP console.",
		})
	}

	// Parse time range
	timeRange := c.Query("range", "1h")
	var startTime time.Time
	switch timeRange {
	case "1h":
		startTime = time.Now().Add(-1 * time.Hour)
	case "6h":
		startTime = time.Now().Add(-6 * time.Hour)
	case "24h":
		startTime = time.Now().Add(-24 * time.Hour)
	case "7d":
		startTime = time.Now().Add(-7 * 24 * time.Hour)
	default:
		startTime = time.Now().Add(-1 * time.Hour)
	}

	result, err := h.logClient.QueryLogs(c.Context(), logging.QueryParams{
		Severity:  c.Query("severity", ""),
		Search:    c.Query("search", ""),
		StartTime: startTime,
		EndTime:   time.Now(),
		PageSize:  c.QueryInt("limit", 100),
		PageToken: c.Query("page_token", ""),
	})
	if err != nil {
		log.Printf("ERROR GetLogs: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to query logs"})
	}

	return c.JSON(result)
}

func (h *Handler) GetLogStats(c *fiber.Ctx) error {
	if h.logClient == nil {
		return c.JSON(fiber.Map{"note": "Cloud Logging client not initialized"})
	}

	timeRange := c.Query("range", "24h")
	var startTime time.Time
	switch timeRange {
	case "1h":
		startTime = time.Now().Add(-1 * time.Hour)
	case "6h":
		startTime = time.Now().Add(-6 * time.Hour)
	case "24h":
		startTime = time.Now().Add(-24 * time.Hour)
	case "7d":
		startTime = time.Now().Add(-7 * 24 * time.Hour)
	default:
		startTime = time.Now().Add(-24 * time.Hour)
	}

	stats, err := h.logClient.GetStats(c.Context(), startTime, time.Now())
	if err != nil {
		log.Printf("ERROR GetLogStats: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to get log stats"})
	}

	return c.JSON(stats)
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
