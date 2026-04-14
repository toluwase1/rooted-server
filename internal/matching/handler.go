package matching

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rooted-dating/rooted-server/internal/chat"
	"github.com/rooted-dating/rooted-server/internal/notification"
	"github.com/rooted-dating/rooted-server/internal/shared/middleware"
	"github.com/rooted-dating/rooted-server/internal/user"
)

type Handler struct {
	service     *Service
	userService *user.Service
	chatService *chat.Service
	notifService *notification.Service
}

func NewHandler(s *Service, us *user.Service, cs *chat.Service, ns *notification.Service) *Handler {
	return &Handler{service: s, userService: us, chatService: cs, notifService: ns}
}

// RegisterRoutes registers all matching/discovery API routes.
func (h *Handler) RegisterRoutes(api fiber.Router) {
	api.Get("/circle", h.GetCircle)
	api.Get("/explore", h.GetExplore)
	api.Post("/swipe", h.Swipe)
	api.Get("/matches", h.GetMatches)
	api.Get("/likes", h.GetLikes)
	api.Post("/unmatch/:matchId", h.Unmatch)
}

func (h *Handler) getUser(c *fiber.Ctx) (*user.User, error) {
	tgUser := c.Locals("telegram_user").(middleware.TelegramUser)
	u, _, err := h.userService.FindOrCreateUser(c.Context(), tgUser.ID, tgUser.Username)
	return u, err
}

func (h *Handler) GetCircle(c *fiber.Ctx) error {
	u, _ := h.getUser(c)
	profile, err := h.userService.GetProfile(c.Context(), u.ID)
	if err != nil || profile == nil {
		return c.Status(400).JSON(fiber.Map{"error": "complete your profile first"})
	}

	filters := h.buildFilters(profile, c)
	candidates, err := h.service.GetDailyCircle(c.Context(), u.ID, filters)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to get matches"})
	}
	return c.JSON(fiber.Map{"candidates": candidates, "count": len(candidates)})
}

func (h *Handler) GetExplore(c *fiber.Ctx) error {
	u, _ := h.getUser(c)
	profile, err := h.userService.GetProfile(c.Context(), u.ID)
	if err != nil || profile == nil {
		return c.Status(400).JSON(fiber.Map{"error": "complete your profile first"})
	}

	remaining, _ := h.service.CheckSwipeLimit(c.Context(), u.ID, u.Subscription)
	if remaining <= 0 {
		return c.Status(429).JSON(fiber.Map{
			"error":   "daily swipe limit reached",
			"upgrade": "Upgrade to Plus for unlimited swipes",
		})
	}

	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	filters := h.buildFilters(profile, c)

	candidates, err := h.service.GetExploreFeed(c.Context(), u.ID, filters, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to load explore"})
	}
	return c.JSON(fiber.Map{
		"candidates":       candidates,
		"count":            len(candidates),
		"swipes_remaining": remaining,
	})
}

func (h *Handler) Swipe(c *fiber.Ctx) error {
	u, _ := h.getUser(c)
	var req SwipeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	match, err := h.service.Swipe(c.Context(), u.ID, req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "swipe failed"})
	}

	h.service.IncrementSwipeCount(c.Context(), u.ID)

	result := fiber.Map{"status": "recorded", "action": req.Action}

	if match != nil {
		result["match"] = match
		result["is_match"] = true

		// Create conversation
		conv, _ := h.chatService.CreateConversation(c.Context(), match.ID, match.UserAID, match.UserBID)
		if conv != nil {
			result["conversation_id"] = conv.ID
		}

		// Notify both users
		otherID := match.UserBID
		if u.ID == match.UserBID {
			otherID = match.UserAID
		}
		otherProfile, _ := h.userService.GetProfile(c.Context(), otherID)
		myProfile, _ := h.userService.GetProfile(c.Context(), u.ID)

		if otherProfile != nil && myProfile != nil {
			otherChatID, _ := h.chatService.GetRecipientTelegramChatID(c.Context(), otherID)
			myChatID, _ := h.chatService.GetRecipientTelegramChatID(c.Context(), u.ID)
			if otherChatID != 0 && myChatID != 0 {
				h.notifService.NotifyMatch(c.Context(), myChatID, otherChatID, myProfile.FirstName, otherProfile.FirstName)
			}
		}
	}

	return c.JSON(result)
}

func (h *Handler) GetMatches(c *fiber.Ctx) error {
	u, _ := h.getUser(c)
	matches, err := h.service.GetMatches(c.Context(), u.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to get matches"})
	}

	var enriched []fiber.Map
	for _, m := range matches {
		otherID := m.UserBID
		if u.ID == m.UserBID {
			otherID = m.UserAID
		}
		profile, _ := h.userService.GetProfile(c.Context(), otherID)
		enriched = append(enriched, fiber.Map{"match": m, "profile": profile})
	}

	return c.JSON(fiber.Map{"matches": enriched, "count": len(enriched)})
}

func (h *Handler) GetLikes(c *fiber.Ctx) error {
	u, _ := h.getUser(c)
	if u.Subscription == "free" {
		likes, _ := h.service.GetLikesReceived(c.Context(), u.ID, 100)
		return c.JSON(fiber.Map{"count": len(likes), "upgrade": "Upgrade to Plus to see who liked you"})
	}
	likes, _ := h.service.GetLikesReceived(c.Context(), u.ID, 50)
	return c.JSON(fiber.Map{"likes": likes, "count": len(likes)})
}

func (h *Handler) Unmatch(c *fiber.Ctx) error {
	u, _ := h.getUser(c)
	h.service.Unmatch(c.Context(), u.ID, c.Params("matchId"))
	return c.JSON(fiber.Map{"status": "unmatched"})
}

func (h *Handler) buildFilters(profile *user.Profile, c *fiber.Ctx) CandidateFilters {
	return CandidateFilters{
		Gender:          profile.Gender,
		GenderPref:      profile.GenderPref,
		Intention:       profile.Intention,
		MinAge:          intQuery(c, "min_age", 18),
		MaxAge:          intQuery(c, "max_age", 99),
		Country:         c.Query("country", profile.Country),
		Heritage:        profile.Heritage,
		DiasporaTag:     profile.DiasporaTag,
		Faith:           profile.Faith,
		FaithImportance: profile.FaithImportance,
		Latitude:        profile.Latitude,
		Longitude:       profile.Longitude,
	}
}

func intQuery(c *fiber.Ctx, key string, defaultVal int) int {
	v, err := strconv.Atoi(c.Query(key, ""))
	if err != nil {
		return defaultVal
	}
	return v
}

func splitQuery(c *fiber.Ctx, key string) []string {
	v := c.Query(key, "")
	if v == "" {
		return nil
	}
	return strings.Split(v, ",")
}
