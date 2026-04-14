package user

import (
	"time"
)

type User struct {
	ID               string    `json:"id"`
	TelegramID       int64     `json:"telegram_id"`
	TelegramUsername  string    `json:"telegram_username,omitempty"`
	Status           string    `json:"status"`    // active, paused, banned, deleted
	Verification     string    `json:"verification"` // unverified, photo_verified, id_verified
	TrustScore       int       `json:"trust_score"`
	Subscription     string    `json:"subscription"` // free, plus, premium
	SubExpiresAt     *time.Time `json:"sub_expires_at,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	LastActiveAt     time.Time `json:"last_active_at"`
}

type Profile struct {
	UserID           string            `json:"user_id"`
	FirstName        string            `json:"first_name"`
	DateOfBirth      string            `json:"date_of_birth"` // YYYY-MM-DD
	Gender           string            `json:"gender"`
	GenderPref       string            `json:"gender_pref"` // male, female, everyone
	City             string            `json:"city,omitempty"`
	Country          string            `json:"country,omitempty"` // ISO 3166-1 alpha-3
	Latitude         float64           `json:"latitude,omitempty"`
	Longitude        float64           `json:"longitude,omitempty"`
	Heritage         []string          `json:"heritage"`      // e.g. ["nigerian", "ghanaian"]
	DiasporaTag      string            `json:"diaspora_tag"`  // born_in_africa, diaspora_1st, diaspora_2nd, returnee, explorer
	Intention        string            `json:"intention"`     // dating, friendship, both
	Faith            string            `json:"faith,omitempty"`
	FaithImportance  string            `json:"faith_importance,omitempty"` // very_important, somewhat, not_important
	Bio              string            `json:"bio,omitempty"`
	AudioBioURL      string            `json:"audio_bio_url,omitempty"`
	CulturalPrompts  []PromptResponse  `json:"cultural_prompts"`
	PersonalityPrompts []PromptResponse `json:"personality_prompts"`
	Completeness     int               `json:"completeness"` // 0-100
	Photos           []Photo           `json:"photos,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

type PromptResponse struct {
	Prompt string `json:"prompt"`
	Answer string `json:"answer"`
}

type Photo struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	URLThumbnail    string    `json:"url_thumbnail"`
	URLMedium       string    `json:"url_medium"`
	URLLarge        string    `json:"url_large"`
	Position        int       `json:"position"`
	IsPrimary       bool      `json:"is_primary"`
	ModerationStatus string  `json:"moderation_status"` // pending, approved, rejected
	CreatedAt       time.Time `json:"created_at"`
}

type CreateProfileRequest struct {
	FirstName        string           `json:"first_name" validate:"required,min=2,max=50"`
	DateOfBirth      string           `json:"date_of_birth" validate:"required"`
	Gender           string           `json:"gender" validate:"required,oneof=male female"`
	GenderPref       string           `json:"gender_pref" validate:"required,oneof=male female everyone"`
	City             string           `json:"city"`
	Country          string           `json:"country"`
	Latitude         float64          `json:"latitude"`
	Longitude        float64          `json:"longitude"`
	Heritage         []string         `json:"heritage" validate:"required,min=1"`
	DiasporaTag      string           `json:"diaspora_tag" validate:"required,oneof=born_in_africa diaspora_1st diaspora_2nd returnee explorer"`
	Intention        string           `json:"intention" validate:"required,oneof=dating friendship both"`
	Faith            string           `json:"faith" validate:"omitempty,oneof=christian muslim traditional spiritual not_religious prefer_not_to_say"`
	FaithImportance  string           `json:"faith_importance" validate:"omitempty,oneof=very_important somewhat not_important"`
	Bio              string           `json:"bio" validate:"max=150"`
	CulturalPrompts  []PromptResponse `json:"cultural_prompts" validate:"omitempty"`
	PersonalityPrompts []PromptResponse `json:"personality_prompts" validate:"omitempty"`
}

type UpdateProfileRequest struct {
	FirstName        *string          `json:"first_name,omitempty"`
	City             *string          `json:"city,omitempty"`
	Country          *string          `json:"country,omitempty"`
	Latitude         *float64         `json:"latitude,omitempty"`
	Longitude        *float64         `json:"longitude,omitempty"`
	Heritage         []string         `json:"heritage,omitempty"`
	DiasporaTag      *string          `json:"diaspora_tag,omitempty"`
	Intention        *string          `json:"intention,omitempty"`
	Faith            *string          `json:"faith,omitempty"`
	FaithImportance  *string          `json:"faith_importance,omitempty"`
	Bio              *string          `json:"bio,omitempty"`
	CulturalPrompts  []PromptResponse `json:"cultural_prompts,omitempty"`
	PersonalityPrompts []PromptResponse `json:"personality_prompts,omitempty"`
}

// ProfileCard is the lightweight version shown in swipe feeds
type ProfileCard struct {
	UserID          string   `json:"user_id"`
	FirstName       string   `json:"first_name"`
	Age             int      `json:"age"`
	City            string   `json:"city"`
	Country         string   `json:"country"`
	Heritage        []string `json:"heritage"`
	DiasporaTag     string   `json:"diaspora_tag"`
	Intention       string   `json:"intention"`
	Faith           string   `json:"faith,omitempty"`
	Verification    string   `json:"verification"`
	PrimaryPhotoURL string   `json:"primary_photo_url"`
	CulturalPrompt  *PromptResponse `json:"cultural_prompt,omitempty"` // show one prompt on card
	MatchScore      int      `json:"match_score,omitempty"` // only for logged-in user context
}
