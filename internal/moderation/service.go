package moderation

import (
	"context"
	"strings"

	"github.com/rooted-dating/rooted-server/internal/shared/config"
)

type Service struct {
	config *config.DynamicConfig
}

func NewService(cfg *config.DynamicConfig) *Service {
	return &Service{config: cfg}
}

// ModerateText checks a message or bio for prohibited content.
// Returns: clean, flagged, or blocked.
func (s *Service) ModerateText(ctx context.Context, text string) string {
	lower := strings.ToLower(text)

	// Check banned words from admin config
	var bannedWords []string
	s.config.GetJSON(ctx, "banned_words", &bannedWords)
	for _, word := range bannedWords {
		if strings.Contains(lower, strings.ToLower(word)) {
			return "blocked"
		}
	}

	// Check scam patterns from admin config
	var scamPatterns []string
	s.config.GetJSON(ctx, "scam_patterns", &scamPatterns)
	for _, pattern := range scamPatterns {
		if strings.Contains(lower, strings.ToLower(pattern)) {
			return "flagged"
		}
	}

	// Basic scam detection (hardcoded patterns as fallback)
	scamIndicators := []string{
		"western union", "money gram", "gift card",
		"send me money", "cash app", "bitcoin",
		"investment opportunity", "crypto",
		"click this link", "whatsapp me",
	}
	for _, indicator := range scamIndicators {
		if strings.Contains(lower, indicator) {
			return "flagged"
		}
	}

	// URL detection (suspicious in early dating conversations)
	if strings.Contains(lower, "http://") || strings.Contains(lower, "https://") ||
		strings.Contains(lower, ".com/") || strings.Contains(lower, ".xyz") {
		return "flagged"
	}

	// Phone number sharing detection (early warning)
	digitCount := 0
	for _, c := range text {
		if c >= '0' && c <= '9' {
			digitCount++
		}
	}
	if digitCount >= 8 {
		// Could be a phone number — flag but don't block
		return "flagged"
	}

	return "clean"
}

// Report represents a user report.
type Report struct {
	ID          string `json:"id"`
	ReporterID  string `json:"reporter_id"`
	ReportedID  string `json:"reported_id"`
	Category    string `json:"category"` // harassment, fake_profile, scam, inappropriate, underage
	Description string `json:"description"`
}

// ShouldAutoBan checks if a user has exceeded the report threshold.
func (s *Service) ShouldAutoBan(ctx context.Context, reportCount int) bool {
	threshold := s.config.GetInt(ctx, "auto_ban_report_threshold", 3)
	return reportCount >= threshold
}
