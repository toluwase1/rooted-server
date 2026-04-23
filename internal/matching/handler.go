package matching

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/rooted-dating/rooted-server/internal/chat"
	"github.com/rooted-dating/rooted-server/internal/media"
	"github.com/rooted-dating/rooted-server/internal/notification"
	"github.com/rooted-dating/rooted-server/internal/shared/middleware"
	"github.com/rooted-dating/rooted-server/internal/user"
)

type Handler struct {
	service       *Service
	userService   *user.Service
	chatService   *chat.Service
	notifService  *notification.Service
	mediaService  *media.Service
	userbotClient *chat.UserbotClient
}

func NewHandler(s *Service, us *user.Service, cs *chat.Service, ns *notification.Service, ms *media.Service, ub *chat.UserbotClient) *Handler {
	return &Handler{service: s, userService: us, chatService: cs, notifService: ns, mediaService: ms, userbotClient: ub}
}

func (h *Handler) enrichCandidatePhotos(ctx context.Context, candidates []ScoredCandidate) {
	if h.mediaService == nil {
		return
	}
	for i := range candidates {
		candidates[i].PrimaryPhoto = h.mediaService.EnrichPhotoURL(ctx, candidates[i].PrimaryPhoto)
	}
}

func (h *Handler) RegisterRoutes(api fiber.Router) {
	api.Get("/circle", h.GetCircle)
	api.Get("/explore", h.GetExplore)
	api.Post("/swipe", h.Swipe)
	api.Get("/matches", h.GetMatches)
	api.Get("/likes", h.GetLikes)
	api.Post("/unmatch/:matchId", h.Unmatch)
}

func (h *Handler) getUser(c *fiber.Ctx) (*user.User, error) {
	tgUser, ok := c.Locals("telegram_user").(middleware.TelegramUser)
	if !ok {
		return nil, fiber.NewError(401, "unauthorized")
	}
	u, _, err := h.userService.FindOrCreateUser(c.Context(), tgUser.ID, tgUser.Username)
	if err != nil {
		log.Printf("ERROR getUser telegram_id=%d: %v", tgUser.ID, err)
		return nil, err
	}
	return u, nil
}

func (h *Handler) GetCircle(c *fiber.Ctx) error {
	u, err := h.getUser(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	profile, err := h.userService.GetProfile(c.Context(), u.ID)
	if err != nil {
		log.Printf("ERROR GetCircle get profile user=%s: %v", u.ID, err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to load profile"})
	}
	if profile == nil {
		return c.Status(400).JSON(fiber.Map{"error": "complete your profile first"})
	}

	filters := h.buildFilters(profile, c)
	candidates, err := h.service.GetDailyCircle(c.Context(), u.ID, filters)
	if err != nil {
		log.Printf("ERROR GetCircle matching user=%s: %v", u.ID, err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to load matches"})
	}
	if candidates == nil {
		candidates = []ScoredCandidate{}
	}
	h.enrichCandidatePhotos(c.Context(), candidates)
	return c.JSON(fiber.Map{"candidates": candidates, "count": len(candidates)})
}

func (h *Handler) GetExplore(c *fiber.Ctx) error {
	u, err := h.getUser(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	profile, err := h.userService.GetProfile(c.Context(), u.ID)
	if err != nil {
		log.Printf("ERROR GetExplore get profile user=%s: %v", u.ID, err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to load profile"})
	}
	if profile == nil {
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
		log.Printf("ERROR GetExplore matching user=%s: %v", u.ID, err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to load explore"})
	}
	if candidates == nil {
		candidates = []ScoredCandidate{}
	}
	h.enrichCandidatePhotos(c.Context(), candidates)
	return c.JSON(fiber.Map{
		"candidates":       candidates,
		"count":            len(candidates),
		"swipes_remaining": remaining,
	})
}

func (h *Handler) Swipe(c *fiber.Ctx) error {
	u, err := h.getUser(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req SwipeRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		return err
	}
	if req.CandidateID == u.ID {
		return c.Status(400).JSON(fiber.Map{"error": "cannot swipe on yourself"})
	}

	match, err := h.service.Swipe(c.Context(), u.ID, req)
	if err != nil {
		log.Printf("ERROR Swipe user=%s candidate=%s: %v", u.ID, req.CandidateID, err)
		return c.Status(500).JSON(fiber.Map{"error": "swipe failed"})
	}

	h.service.IncrementSwipeCount(c.Context(), u.ID)

	result := fiber.Map{"status": "recorded", "action": req.Action}

	if match != nil {
		result["match"] = match
		result["is_match"] = true

		conv, err := h.chatService.CreateConversation(c.Context(), match.ID, match.UserAID, match.UserBID)
		if err != nil {
			log.Printf("ERROR Swipe create conversation match=%s: %v", match.ID, err)
		}
		if conv != nil {
			result["conversation_id"] = conv.ID
		}

		// Set up chat channel + notify both users (best effort, async)
		go func() {
			ctx := context.Background()
			otherID := match.UserBID
			if u.ID == match.UserBID {
				otherID = match.UserAID
			}
			otherProfile, _ := h.userService.GetProfile(ctx, otherID)
			myProfile, _ := h.userService.GetProfile(ctx, u.ID)
			if otherProfile == nil || myProfile == nil {
				return
			}

			// Try supergroup first, fall back to bot relay
			var inviteLink string
			if h.userbotClient != nil && conv != nil {
				otherUser, _ := h.userService.GetUser(ctx, otherID)
				myUser, _ := h.userService.GetUser(ctx, u.ID)

				if myUser != nil && otherUser != nil {
					groupResp, groupCreationErr := h.userbotClient.CreateGroup(
						ctx, conv.ID,
						myUser.TelegramID, otherUser.TelegramID,
						myProfile.FirstName, otherProfile.FirstName,
					)
					if groupCreationErr != nil {
						log.Printf("WARN group creation failed, using bot relay: %v", groupCreationErr)
					} else {
						// Persist group info on the conversation
						h.chatService.SetGroupChat(ctx, conv.ID, groupResp.GroupID, groupResp.InviteLink)
						inviteLink = groupResp.InviteLink
						log.Printf("Group chat created for match %s (group %d)", match.ID, groupResp.GroupID)
					}
				}
			}

			// Notify both users
			myChatID, _ := h.chatService.GetRecipientTelegramChatID(ctx, u.ID)
			otherChatID, _ := h.chatService.GetRecipientTelegramChatID(ctx, otherID)

			if myChatID != 0 {
				h.notifService.SendBotMessage(ctx, myChatID, matchNotification(otherProfile.FirstName, inviteLink))
			}
			if otherChatID != 0 {
				h.notifService.SendBotMessage(ctx, otherChatID, matchNotification(myProfile.FirstName, inviteLink))
			}
		}()
	}

	return c.JSON(result)
}

func (h *Handler) GetMatches(c *fiber.Ctx) error {
	u, err := h.getUser(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	matches, err := h.service.GetMatches(c.Context(), u.ID)
	if err != nil {
		log.Printf("ERROR GetMatches user=%s: %v", u.ID, err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to load matches"})
	}

	var enriched []fiber.Map
	for _, m := range matches {
		otherID := m.UserBID
		if u.ID == m.UserBID {
			otherID = m.UserAID
		}
		profile, _ := h.userService.GetProfile(c.Context(), otherID)
		if profile != nil && h.mediaService != nil {
			for i := range profile.Photos {
				profile.Photos[i].URLThumbnail = h.mediaService.EnrichPhotoURL(c.Context(), profile.Photos[i].URLThumbnail)
				profile.Photos[i].URLMedium = h.mediaService.EnrichPhotoURL(c.Context(), profile.Photos[i].URLMedium)
				profile.Photos[i].URLLarge = h.mediaService.EnrichPhotoURL(c.Context(), profile.Photos[i].URLLarge)
			}
		}
		// Look up conversation ID for chat
		var convID string
		conv, _ := h.chatService.GetConversationByMatch(c.Context(), m.ID)
		if conv != nil {
			convID = conv.ID
		}
		enriched = append(enriched, fiber.Map{"match": m, "profile": profile, "conversation_id": convID})
	}
	if enriched == nil {
		enriched = []fiber.Map{}
	}

	return c.JSON(fiber.Map{"matches": enriched, "count": len(enriched)})
}

func (h *Handler) GetLikes(c *fiber.Ctx) error {
	u, err := h.getUser(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	if u.Subscription == "free" {
		likes, _ := h.service.GetLikesReceived(c.Context(), u.ID, 100)
		count := 0
		if likes != nil {
			count = len(likes)
		}
		return c.JSON(fiber.Map{"count": count, "upgrade": "Upgrade to Plus to see who liked you"})
	}
	likes, err := h.service.GetLikesReceived(c.Context(), u.ID, 50)
	if err != nil {
		log.Printf("ERROR GetLikes user=%s: %v", u.ID, err)
	}
	if likes == nil {
		likes = []Swipe{}
	}
	return c.JSON(fiber.Map{"likes": likes, "count": len(likes)})
}

func (h *Handler) Unmatch(c *fiber.Ctx) error {
	u, err := h.getUser(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	matchID := c.Params("matchId")
	if matchID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "match id required"})
	}
	if err := h.service.Unmatch(c.Context(), u.ID, matchID); err != nil {
		log.Printf("ERROR Unmatch user=%s match=%s: %v", u.ID, matchID, err)
	}
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

func matchNotification(otherName, inviteLink string) string {
	msg := fmt.Sprintf("You and %s are a match!", otherName)
	if inviteLink != "" {
		msg += fmt.Sprintf("\n\nJoin your private chat: %s", inviteLink)
	} else {
		msg += "\n\nSend them a message — I'll relay it for you."
	}
	return msg
}

func intQuery(c *fiber.Ctx, key string, defaultVal int) int {
	v, err := strconv.Atoi(c.Query(key, ""))
	if err != nil {
		return defaultVal
	}
	return v
}
