package matching

import "time"

type Swipe struct {
	ID           string    `json:"id"`
	SwiperID     string    `json:"swiper_id"`
	SwipedID     string    `json:"swiped_id"`
	Action       string    `json:"action"`        // like, pass
	Comment      string    `json:"comment,omitempty"`
	LikedElement string    `json:"liked_element,omitempty"` // photo_1, cultural_prompt_0, etc.
	CreatedAt    time.Time `json:"created_at"`
}

type Match struct {
	ID            string     `json:"id"`
	UserAID       string     `json:"user_a_id"`
	UserBID       string     `json:"user_b_id"`
	Compatibility int        `json:"compatibility"` // 0-100
	Status        string     `json:"status"`        // active, unmatched_a, unmatched_b, expired
	MatchedAt     time.Time  `json:"matched_at"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
}

type ScoredCandidate struct {
	UserID       string   `json:"user_id"`
	FirstName    string   `json:"first_name"`
	Age          int      `json:"age"`
	City         string   `json:"city"`
	Country      string   `json:"country"`
	Heritage     []string `json:"heritage"`
	DiasporaTag  string   `json:"diaspora_tag"`
	Intention    string   `json:"intention"`
	Faith        string   `json:"faith"`
	Verification string   `json:"verification"`
	TrustScore   int      `json:"trust_score"`
	MatchScore   int      `json:"match_score"`   // raw score from SQL
	FinalScore   float64  `json:"final_score"`   // after re-ranking
	PrimaryPhoto string   `json:"primary_photo"`
	LastActiveAt time.Time `json:"-"`
	CreatedAt    time.Time `json:"-"`
	HasLikedMe   bool     `json:"-"` // for reciprocity boost
	HasSpotlight bool     `json:"-"` // for spotlight boost
	SentRose     bool     `json:"-"` // rose → top of queue
}

type SwipeRequest struct {
	CandidateID  string `json:"candidate_id" validate:"required"`
	Action       string `json:"action" validate:"required,oneof=like pass"`
	Comment      string `json:"comment,omitempty" validate:"max=200"`
	LikedElement string `json:"liked_element,omitempty"`
}

type CandidateFilters struct {
	Gender          string
	GenderPref      string
	Intention       string
	MinAge          int
	MaxAge          int
	Country         string   // optional filter
	Heritage        []string // optional filter (Plus+)
	DiasporaTag     string
	Faith           string   // optional filter (Plus+)
	FaithImportance string
	Latitude        float64
	Longitude       float64
	Limit           int
}

type DailyCircle struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	CandidateIDs []string  `json:"candidate_ids"`
	DeliveredAt  time.Time `json:"delivered_at"`
	Opened       bool      `json:"opened"`
	ActionsTaken int       `json:"actions_taken"`
}
