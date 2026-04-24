package user

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rooted-dating/rooted-server/internal/shared/config"
	"github.com/rooted-dating/rooted-server/internal/shared/database"
)

const profileCacheTTL = 1 * time.Hour

type Service struct {
	repo   Repository
	redis  *database.SafeRedis
	config *config.DynamicConfig
}

func NewService(repo Repository, redis *database.SafeRedis, cfg *config.DynamicConfig) *Service {
	return &Service{repo: repo, redis: redis, config: cfg}
}

// GetMinAge returns the admin-configurable minimum age for registration.
func (s *Service) GetMinAge(ctx context.Context) int {
	return s.config.GetInt(ctx, "min_registration_age", 18)
}

// FindOrCreateUser returns existing user or creates a new one from Telegram auth.
func (s *Service) FindOrCreateUser(ctx context.Context, telegramID int64, username string) (*User, bool, error) {
	user, err := s.repo.GetUserByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, false, err
	}
	if user != nil {
		// Update last active
		s.repo.UpdateLastActive(ctx, user.ID)
		s.setOnlineStatus(ctx, user.ID)
		return user, false, nil // false = not new
	}

	// Create new user
	user, err = s.repo.CreateUser(ctx, telegramID, username)
	if err != nil {
		return nil, false, err
	}
	return user, true, nil // true = new user
}

func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *Service) GetProfile(ctx context.Context, userID string) (*Profile, error) {
	cacheKey := "profile:" + userID

	// Check cache
	cached, err := s.redis.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var profile Profile
		if json.Unmarshal(cached, &profile) == nil {
			return &profile, nil
		}
	}

	// Cache miss
	profile, err := s.repo.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, nil
	}

	// Recalculate completeness on every load (cheap, ensures accuracy)
	newCompleteness := s.recalcCompleteness(profile)
	if newCompleteness != profile.Completeness {
		profile.Completeness = newCompleteness
		s.repo.UpdateCompleteness(ctx, userID, newCompleteness)
	}

	// Cache for next time
	data, _ := json.Marshal(profile)
	s.redis.Set(ctx, cacheKey, data, profileCacheTTL)

	return profile, nil
}

func (s *Service) CreateProfile(ctx context.Context, userID string, req CreateProfileRequest) (*Profile, error) {
	profile, err := s.repo.CreateProfile(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	// Recalculate completeness and persist
	profile.Completeness = s.recalcCompleteness(profile)
	s.repo.UpdateCompleteness(ctx, userID, profile.Completeness)

	// Cache the new profile
	data, _ := json.Marshal(profile)
	s.redis.Set(ctx, "profile:"+userID, data, profileCacheTTL)

	return profile, nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) error {
	if err := s.repo.UpdateProfile(ctx, userID, req); err != nil {
		return err
	}

	// Recalculate completeness
	profile, err := s.repo.GetProfile(ctx, userID)
	if err == nil && profile != nil {
		completeness := s.recalcCompleteness(profile)
		s.repo.UpdateCompleteness(ctx, userID, completeness)
	}

	// Invalidate cache
	s.redis.Del(ctx, "profile:"+userID)

	return nil
}

func (s *Service) AcceptTerms(ctx context.Context, userID string) error {
	return s.repo.AcceptTerms(ctx, userID)
}

func (s *Service) DeleteAccount(ctx context.Context, userID string) error {
	if err := s.repo.DeleteUser(ctx, userID); err != nil {
		return err
	}
	s.redis.Del(ctx, "profile:"+userID)
	s.redis.Del(ctx, "matches:"+userID)
	s.redis.Del(ctx, "circle:"+userID)
	s.redis.Del(ctx, "explore:"+userID)
	return nil
}

func (s *Service) PauseProfile(ctx context.Context, userID string) error {
	if err := s.repo.UpdateUserStatus(ctx, userID, "paused"); err != nil {
		return err
	}
	s.redis.Del(ctx, "profile:"+userID)
	return nil
}

func (s *Service) ResumeProfile(ctx context.Context, userID string) error {
	if err := s.repo.UpdateUserStatus(ctx, userID, "active"); err != nil {
		return err
	}
	s.redis.Del(ctx, "profile:"+userID)
	return nil
}

func (s *Service) VerifyPhoto(ctx context.Context, userID string) error {
	if err := s.repo.UpdateUserVerification(ctx, userID, "photo_verified"); err != nil {
		return err
	}
	s.redis.Del(ctx, "profile:"+userID)
	return nil
}

func (s *Service) BlockUser(ctx context.Context, blockerID, blockedID string) error {
	return s.repo.BlockUser(ctx, blockerID, blockedID)
}

func (s *Service) UnblockUser(ctx context.Context, blockerID, blockedID string) error {
	return s.repo.UnblockUser(ctx, blockerID, blockedID)
}

// AddPhoto adds a photo and invalidates profile cache.
func (s *Service) AddPhoto(ctx context.Context, photo Photo) error {
	if err := s.repo.AddPhoto(ctx, photo); err != nil {
		return err
	}
	s.RecalcAndSaveCompleteness(ctx, photo.UserID)
	return nil
}

func (s *Service) DeletePhoto(ctx context.Context, photoID, userID string) error {
	if err := s.repo.DeletePhoto(ctx, photoID, userID); err != nil {
		return err
	}
	s.RecalcAndSaveCompleteness(ctx, userID)
	return nil
}

func (s *Service) ReorderPhotos(ctx context.Context, userID string, photoIDs []string) error {
	if err := s.repo.ReorderPhotos(ctx, userID, photoIDs); err != nil {
		return err
	}
	s.invalidateProfileCache(ctx, userID)
	return nil
}

func (s *Service) setOnlineStatus(ctx context.Context, userID string) {
	s.redis.Set(ctx, "online:"+userID, "1", 60*time.Second)
}

func (s *Service) IsOnline(ctx context.Context, userID string) bool {
	val, err := s.redis.Get(ctx, "online:"+userID).Result()
	return err == nil && val == "1"
}

func (s *Service) invalidateProfileCache(ctx context.Context, userID string) {
	s.redis.Del(ctx, "profile:"+userID)
}

func (s *Service) recalcCompleteness(p *Profile) int {
	score := 0
	total := 14

	if p.FirstName != "" { score++ }        // 1
	if p.DateOfBirth != "" { score++ }       // 2
	if p.Gender != "" { score++ }            // 3
	if p.City != "" { score++ }              // 4
	if len(p.Heritage) > 0 { score++ }       // 5
	if p.DiasporaTag != "" { score++ }       // 6
	if p.Intention != "" { score++ }         // 7
	if p.Faith != "" { score++ }             // 8
	if len(p.CulturalPrompts) >= 2 { score++ } // 9
	if len(p.PersonalityPrompts) >= 1 { score++ } // 10
	if len(p.Photos) >= 1 { score++ }        // 11 — has at least 1 photo
	if len(p.Photos) >= 4 { score++ }        // 12 — has 4+ photos (bonus)
	if p.Bio != "" { score++ }               // 13
	if p.AudioBioURL != "" { score++ }       // 14

	return (score * 100) / total
}

// RecalcAndSaveCompleteness recalculates and persists completeness for a user.
func (s *Service) RecalcAndSaveCompleteness(ctx context.Context, userID string) {
	profile, err := s.repo.GetProfile(ctx, userID)
	if err != nil || profile == nil {
		return
	}
	completeness := s.recalcCompleteness(profile)
	s.repo.UpdateCompleteness(ctx, userID, completeness)
	s.invalidateProfileCache(ctx, userID)
}
