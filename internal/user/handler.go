package user

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rooted-dating/rooted-server/internal/media"
	"github.com/rooted-dating/rooted-server/internal/shared/middleware"
)

type Handler struct {
	service      *Service
	mediaService *media.Service
}

func NewHandler(service *Service, mediaService *media.Service) *Handler {
	return &Handler{service: service, mediaService: mediaService}
}

func (h *Handler) RegisterRoutes(api fiber.Router) {
	api.Get("/me", h.GetMe)
	api.Get("/profile/:id", h.GetProfileByID)
	api.Post("/profile", h.CreateProfile)
	api.Put("/profile", h.UpdateProfile)
	api.Post("/profile/pause", h.PauseProfile)
	api.Post("/profile/resume", h.ResumeProfile)
	api.Delete("/account", h.DeleteAccount)
	api.Post("/block/:userId", h.BlockUser)
	api.Delete("/block/:userId", h.UnblockUser)
	api.Post("/verify", h.VerifyPhoto)
	api.Post("/photos", h.UploadPhoto)
	api.Delete("/photos/:photoId", h.DeletePhoto)
	api.Put("/photos/reorder", h.ReorderPhotos)
	api.Get("/photos/upload-url", h.GetUploadURL)
}

func (h *Handler) getUser(c *fiber.Ctx) (*User, error) {
	tgUser, ok := c.Locals("telegram_user").(middleware.TelegramUser)
	if !ok {
		return nil, fiber.NewError(401, "unauthorized")
	}
	u, _, err := h.service.FindOrCreateUser(c.Context(), tgUser.ID, tgUser.Username)
	if err != nil {
		log.Printf("ERROR getUser telegram_id=%d: %v", tgUser.ID, err)
		return nil, fiber.NewError(500, "failed to authenticate user")
	}
	return u, nil
}

func (h *Handler) GetMe(c *fiber.Ctx) error {
	tgUser, ok := c.Locals("telegram_user").(middleware.TelegramUser)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	u, isNew, err := h.service.FindOrCreateUser(c.Context(), tgUser.ID, tgUser.Username)
	if err != nil {
		log.Printf("ERROR GetMe telegram_id=%d: %v", tgUser.ID, err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to load user"})
	}
	var profile *Profile
	if !isNew {
		profile, _ = h.service.GetProfile(c.Context(), u.ID)
	}
	return c.JSON(fiber.Map{"user": u, "profile": profile, "is_new": isNew})
}

func (h *Handler) GetProfileByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "profile id required"})
	}
	profile, err := h.service.GetProfile(c.Context(), id)
	if err != nil {
		log.Printf("ERROR GetProfile id=%s: %v", id, err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to load profile"})
	}
	if profile == nil {
		return c.Status(404).JSON(fiber.Map{"error": "profile not found"})
	}
	return c.JSON(profile)
}

func (h *Handler) CreateProfile(c *fiber.Ctx) error {
	u, err := h.getUser(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req CreateProfileRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		return err
	}

	profile, err := h.service.CreateProfile(c.Context(), u.ID, req)
	if err != nil {
		log.Printf("ERROR CreateProfile user=%s: %v", u.ID, err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to create profile"})
	}
	return c.Status(201).JSON(profile)
}

func (h *Handler) UpdateProfile(c *fiber.Ctx) error {
	u, err := h.getUser(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req UpdateProfileRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.service.UpdateProfile(c.Context(), u.ID, req); err != nil {
		log.Printf("ERROR UpdateProfile user=%s: %v", u.ID, err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to update profile"})
	}
	return c.JSON(fiber.Map{"status": "updated"})
}

func (h *Handler) PauseProfile(c *fiber.Ctx) error {
	u, err := h.getUser(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	if err := h.service.PauseProfile(c.Context(), u.ID); err != nil {
		log.Printf("ERROR PauseProfile user=%s: %v", u.ID, err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to pause profile"})
	}
	return c.JSON(fiber.Map{"status": "paused"})
}

func (h *Handler) ResumeProfile(c *fiber.Ctx) error {
	u, err := h.getUser(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	if err := h.service.ResumeProfile(c.Context(), u.ID); err != nil {
		log.Printf("ERROR ResumeProfile user=%s: %v", u.ID, err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to resume profile"})
	}
	return c.JSON(fiber.Map{"status": "resumed"})
}

func (h *Handler) DeleteAccount(c *fiber.Ctx) error {
	u, err := h.getUser(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	if err := h.service.DeleteAccount(c.Context(), u.ID); err != nil {
		log.Printf("ERROR DeleteAccount user=%s: %v", u.ID, err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete account"})
	}
	return c.JSON(fiber.Map{"status": "deleted"})
}

func (h *Handler) BlockUser(c *fiber.Ctx) error {
	u, err := h.getUser(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	targetID := c.Params("userId")
	if targetID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "user id required"})
	}
	if targetID == u.ID {
		return c.Status(400).JSON(fiber.Map{"error": "cannot block yourself"})
	}
	if err := h.service.BlockUser(c.Context(), u.ID, targetID); err != nil {
		log.Printf("ERROR BlockUser user=%s target=%s: %v", u.ID, targetID, err)
	}
	return c.JSON(fiber.Map{"status": "blocked"})
}

func (h *Handler) UnblockUser(c *fiber.Ctx) error {
	u, err := h.getUser(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	targetID := c.Params("userId")
	if targetID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "user id required"})
	}
	if err := h.service.UnblockUser(c.Context(), u.ID, targetID); err != nil {
		log.Printf("ERROR UnblockUser user=%s target=%s: %v", u.ID, targetID, err)
	}
	return c.JSON(fiber.Map{"status": "unblocked"})
}

func (h *Handler) VerifyPhoto(c *fiber.Ctx) error {
	u, err := h.getUser(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	_, err = c.FormFile("selfie")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "selfie photo required"})
	}

	if err := h.service.VerifyPhoto(c.Context(), u.ID); err != nil {
		log.Printf("ERROR VerifyPhoto user=%s: %v", u.ID, err)
		return c.Status(500).JSON(fiber.Map{"error": "verification failed"})
	}

	return c.JSON(fiber.Map{"status": "verified", "verification": "photo_verified"})
}

func (h *Handler) UploadPhoto(c *fiber.Ctx) error {
	if h.mediaService == nil {
		return c.Status(503).JSON(fiber.Map{"error": "photo uploads not configured — R2 storage not set up"})
	}

	u, err := h.getUser(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	existing, _ := h.service.repo.GetPhotos(c.Context(), u.ID)
	if len(existing) >= 6 {
		return c.Status(400).JSON(fiber.Map{"error": "maximum 6 photos allowed"})
	}

	file, err := c.FormFile("photo")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "photo file required"})
	}

	if file.Size > 5*1024*1024 {
		return c.Status(400).JSON(fiber.Map{"error": "photo must be under 5MB"})
	}

	contentType := file.Header.Get("Content-Type")
	if contentType != "" && contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		return c.Status(400).JSON(fiber.Map{"error": "only JPEG, PNG, and WebP images are allowed"})
	}

	f, err := file.Open()
	if err != nil {
		log.Printf("ERROR UploadPhoto open file user=%s: %v", u.ID, err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to read file"})
	}
	defer f.Close()

	result, err := h.mediaService.UploadPhoto(c.Context(), u.ID, f, file.Filename)
	if err != nil {
		log.Printf("ERROR UploadPhoto upload user=%s: %v", u.ID, err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to upload photo"})
	}

	position := len(existing)
	photo := Photo{
		ID:           uuid.New().String(),
		UserID:       u.ID,
		URLThumbnail: result.URLThumbnail,
		URLMedium:    result.URLMedium,
		URLLarge:     result.URLLarge,
		Position:     position,
		IsPrimary:    position == 0,
	}

	if err := h.service.AddPhoto(c.Context(), photo); err != nil {
		log.Printf("ERROR UploadPhoto save user=%s: %v", u.ID, err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to save photo"})
	}

	return c.Status(201).JSON(photo)
}

func (h *Handler) GetUploadURL(c *fiber.Ctx) error {
	if h.mediaService == nil {
		return c.Status(503).JSON(fiber.Map{"error": "photo uploads not configured"})
	}

	u, err := h.getUser(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	ext := c.Query("ext", ".jpg")
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		return c.Status(400).JSON(fiber.Map{"error": "only .jpg, .png, .webp extensions allowed"})
	}

	url, fileID, err := h.mediaService.GeneratePresignedUploadURL(c.Context(), u.ID, ext)
	if err != nil {
		log.Printf("ERROR GetUploadURL user=%s: %v", u.ID, err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate upload URL"})
	}

	return c.JSON(fiber.Map{"upload_url": url, "file_id": fileID})
}

func (h *Handler) DeletePhoto(c *fiber.Ctx) error {
	u, err := h.getUser(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	photoID := c.Params("photoId")
	if photoID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "photo id required"})
	}

	if err := h.service.DeletePhoto(c.Context(), photoID, u.ID); err != nil {
		log.Printf("ERROR DeletePhoto user=%s photo=%s: %v", u.ID, photoID, err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete photo"})
	}
	return c.JSON(fiber.Map{"status": "deleted"})
}

func (h *Handler) ReorderPhotos(c *fiber.Ctx) error {
	u, err := h.getUser(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	var body struct {
		PhotoIDs []string `json:"photo_ids"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if len(body.PhotoIDs) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "photo_ids required"})
	}

	if err := h.service.ReorderPhotos(c.Context(), u.ID, body.PhotoIDs); err != nil {
		log.Printf("ERROR ReorderPhotos user=%s: %v", u.ID, err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to reorder photos"})
	}
	return c.JSON(fiber.Map{"status": "reordered"})
}
