package matching

import "context"

type Repository interface {
	// Candidates
	FindCandidates(ctx context.Context, userID string, userProfile CandidateFilters, scoringParams ScoringParams) ([]ScoredCandidate, error)

	// Swipes
	RecordSwipe(ctx context.Context, swipe Swipe) error
	HasSwiped(ctx context.Context, swiperID, swipedID string) (bool, error)
	GetLikesReceived(ctx context.Context, userID string, limit int) ([]Swipe, error)

	// Matches
	CreateMatch(ctx context.Context, userAID, userBID string, compatibility int) (*Match, error)
	GetMatch(ctx context.Context, userAID, userBID string) (*Match, error)
	GetMatches(ctx context.Context, userID string) ([]Match, error)
	UpdateMatchStatus(ctx context.Context, matchID, status string) error
	UpdateMatchLastMessage(ctx context.Context, matchID string) error

	// Seen profiles
	MarkSeen(ctx context.Context, userID, seenUserID string) error
	GetSeenIDs(ctx context.Context, userID string) ([]string, error)

	// Daily circles
	SaveDailyCircle(ctx context.Context, circle DailyCircle) error
	GetLatestCircle(ctx context.Context, userID string) (*DailyCircle, error)
}

// ScoringParams holds all the configurable algorithm weights and point values,
// loaded from admin_config before each matching run.
type ScoringParams struct {
	HeritageExactPoints      int
	HeritageCountryPoints    int
	HeritageRegionPoints     int
	FaithBothImportantPoints int
	FaithSamePoints          int
	FaithBothUnimportantPoints int
	DiasporaSamePoints       int
	DiasporaSimilarPoints    int
	DiasporaCrossPoints      int
	Proximity10kmPoints      int
	Proximity50kmPoints      int
	Proximity200kmPoints     int
	ProximityCountryPoints   int
}
