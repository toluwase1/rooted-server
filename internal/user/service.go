package user

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rooted-dating/rooted-server/internal/shared/database"
)

const profileCacheTTL = 1 * time.Hour

type Service struct {
	repo  Repository
	redis *database.SafeRedis
}

func NewService(repo Repository, redis *database.SafeRedis) *Service {
	return &Service{repo: repo, redis: redis}
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
	s.invalidateProfileCache(ctx, photo.UserID)
	return nil
}

func (s *Service) DeletePhoto(ctx context.Context, photoID, userID string) error {
	if err := s.repo.DeletePhoto(ctx, photoID, userID); err != nil {
		return err
	}
	s.invalidateProfileCache(ctx, userID)
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
	total := 12

	if p.FirstName != "" { score++ }
	if p.DateOfBirth != "" { score++ }
	if p.Gender != "" { score++ }
	if p.City != "" { score++ }
	if len(p.Heritage) > 0 { score++ }
	if p.DiasporaTag != "" { score++ }
	if p.Intention != "" { score++ }
	if p.Faith != "" { score++ }
	if len(p.CulturalPrompts) >= 2 { score++ }
	if len(p.PersonalityPrompts) >= 1 { score++ }
	if len(p.Photos) >= 1 { score++ }
	if p.Bio != "" { score++ }

	_ = fmt.Sprintf("")
	return (score * 100) / total
}
