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
	tgUser := c.Locals("telegram_user").(middleware.TelegramUser)
	u, _, err := h.service.FindOrCreateUser(c.Context(), tgUser.ID, tgUser.Username)
	return u, err
}

func (h *Handler) GetMe(c *fiber.Ctx) error {
	tgUser := c.Locals("telegram_user").(middleware.TelegramUser)
	u, isNew, err := h.service.FindOrCreateUser(c.Context(), tgUser.ID, tgUser.Username)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to find/create user"})
	}
	var profile *Profile
	if !isNew {
		profile, _ = h.service.GetProfile(c.Context(), u.ID)
	}
	return c.JSON(fiber.Map{"user": u, "profile": profile, "is_new": isNew})
}

func (h *Handler) GetProfileByID(c *fiber.Ctx) error {
	profile, err := h.service.GetProfile(c.Context(), c.Params("id"))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to get profile"})
	}
	if profile == nil {
		return c.Status(404).JSON(fiber.Map{"error": "profile not found"})
	}
	return c.JSON(profile)
}

func (h *Handler) CreateProfile(c *fiber.Ctx) error {
	u, err := h.getUser(c)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "auth failed"})
	}
	var req CreateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
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
		return c.Status(500).JSON(fiber.Map{"error": "auth failed"})
	}
	var req UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.service.UpdateProfile(c.Context(), u.ID, req); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update profile"})
	}
	return c.JSON(fiber.Map{"status": "updated"})
}

func (h *Handler) PauseProfile(c *fiber.Ctx) error {
	u, _ := h.getUser(c)
	h.service.PauseProfile(c.Context(), u.ID)
	return c.JSON(fiber.Map{"status": "paused"})
}

func (h *Handler) ResumeProfile(c *fiber.Ctx) error {
	u, _ := h.getUser(c)
	h.service.ResumeProfile(c.Context(), u.ID)
	return c.JSON(fiber.Map{"status": "resumed"})
}

func (h *Handler) DeleteAccount(c *fiber.Ctx) error {
	u, _ := h.getUser(c)
	h.service.DeleteAccount(c.Context(), u.ID)
	return c.JSON(fiber.Map{"status": "deleted"})
}

func (h *Handler) BlockUser(c *fiber.Ctx) error {
	u, _ := h.getUser(c)
	h.service.BlockUser(c.Context(), u.ID, c.Params("userId"))
	return c.JSON(fiber.Map{"status": "blocked"})
}

func (h *Handler) UnblockUser(c *fiber.Ctx) error {
	u, _ := h.getUser(c)
	h.service.UnblockUser(c.Context(), u.ID, c.Params("userId"))
	return c.JSON(fiber.Map{"status": "unblocked"})
}

// VerifyPhoto handles photo verification — compares selfie to profile photos.
// For MVP: marks user as verified (face matching via Rekognition added later).
func (h *Handler) VerifyPhoto(c *fiber.Ctx) error {
	u, err := h.getUser(c)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "auth failed"})
	}

	_, err = c.FormFile("selfie")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "no selfie provided"})
	}

	// TODO: Compare selfie to profile photos using AWS Rekognition CompareFaces.
	// For MVP, we trust the selfie flow and mark as verified.
	// The selfie is not stored — just used for comparison.

	if err := h.service.VerifyPhoto(c.Context(), u.ID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "verification failed"})
	}

	return c.JSON(fiber.Map{"status": "verified", "verification": "photo_verified"})
}

// UploadPhoto handles multipart file upload for profile photos.
func (h *Handler) UploadPhoto(c *fiber.Ctx) error {
	if h.mediaService == nil {
		return c.Status(503).JSON(fiber.Map{"error": "photo uploads not configured"})
	}

	u, err := h.getUser(c)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "auth failed"})
	}

	// Check photo limit (max 6)
	existing, _ := h.service.repo.GetPhotos(c.Context(), u.ID)
	if len(existing) >= 6 {
		return c.Status(400).JSON(fiber.Map{"error": "maximum 6 photos allowed"})
	}

	file, err := c.FormFile("photo")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "no photo file provided"})
	}

	// Validate file size (5MB max)
	if file.Size > 5*1024*1024 {
		return c.Status(400).JSON(fiber.Map{"error": "photo must be under 5MB"})
	}

	// Open file
	f, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to read file"})
	}
	defer f.Close()

	// Upload to R2
	result, err := h.mediaService.UploadPhoto(c.Context(), u.ID, f, file.Filename)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to upload photo"})
	}

	// Save photo record
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
		return c.Status(500).JSON(fiber.Map{"error": "failed to save photo"})
	}

	return c.Status(201).JSON(photo)
}

// GetUploadURL returns a presigned URL for direct client-to-R2 upload.
func (h *Handler) GetUploadURL(c *fiber.Ctx) error {
	if h.mediaService == nil {
		return c.Status(503).JSON(fiber.Map{"error": "photo uploads not configured"})
	}

	u, err := h.getUser(c)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "auth failed"})
	}

	ext := c.Query("ext", ".jpg")
	url, fileID, err := h.mediaService.GeneratePresignedUploadURL(c.Context(), u.ID, ext)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate upload URL"})
	}

	return c.JSON(fiber.Map{"upload_url": url, "file_id": fileID})
}

// DeletePhoto removes a photo from profile and storage.
func (h *Handler) DeletePhoto(c *fiber.Ctx) error {
	u, _ := h.getUser(c)
	photoID := c.Params("photoId")

	if err := h.service.DeletePhoto(c.Context(), photoID, u.ID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete photo"})
	}
	return c.JSON(fiber.Map{"status": "deleted"})
}

// ReorderPhotos updates the display order of photos.
func (h *Handler) ReorderPhotos(c *fiber.Ctx) error {
	u, _ := h.getUser(c)
	var body struct {
		PhotoIDs []string `json:"photo_ids"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	if err := h.service.ReorderPhotos(c.Context(), u.ID, body.PhotoIDs); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to reorder photos"})
	}
	return c.JSON(fiber.Map{"status": "reordered"})
}
