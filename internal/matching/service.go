package matching

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/rooted-dating/rooted-server/internal/shared/config"
	"github.com/rooted-dating/rooted-server/internal/shared/database"
)

type Service struct {
	repo   Repository
	redis  *database.SafeRedis
	config *config.DynamicConfig
}

func NewService(repo Repository, redis *database.SafeRedis, cfg *config.DynamicConfig) *Service {
	return &Service{repo: repo, redis: redis, config: cfg}
}

// GetDailyCircle returns cached daily circle or generates a new one.
func (s *Service) GetDailyCircle(ctx context.Context, userID string, filters CandidateFilters) ([]ScoredCandidate, error) {
	cacheKey := "circle:" + userID

	// Check cache
	cached, err := s.redis.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var candidates []ScoredCandidate
		if json.Unmarshal(cached, &candidates) == nil {
			return candidates, nil
		}
	}

	// Generate new circle
	circleSize := s.config.GetIntRegional(ctx, "daily_circle_size", filters.Country, 8)
	filters.Limit = circleSize

	candidates, err := s.findAndScore(ctx, userID, filters)
	if err != nil {
		return nil, err
	}

	// Cache for 24 hours
	data, _ := json.Marshal(candidates)
	s.redis.Set(ctx, cacheKey, data, 24*time.Hour)

	// Save to DB for analytics
	candidateIDs := make([]string, len(candidates))
	for i, c := range candidates {
		candidateIDs[i] = c.UserID
	}
	s.repo.SaveDailyCircle(ctx, DailyCircle{
		UserID:       userID,
		CandidateIDs: candidateIDs,
	})

	return candidates, nil
}

// GetExploreFeed returns cached explore candidates or generates fresh ones.
func (s *Service) GetExploreFeed(ctx context.Context, userID string, filters CandidateFilters, offset int) ([]ScoredCandidate, error) {
	cacheKey := "explore:" + userID

	// Check cache
	cached, err := s.redis.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var candidates []ScoredCandidate
		if json.Unmarshal(cached, &candidates) == nil {
			if offset < len(candidates) {
				end := offset + 10
				if end > len(candidates) {
					end = len(candidates)
				}
				return candidates[offset:end], nil
			}
		}
	}

	// Generate fresh candidates (larger batch for explore)
	filters.Limit = 100
	candidates, err := s.findAndScore(ctx, userID, filters)
	if err != nil {
		return nil, err
	}

	// Cache for 30 minutes
	data, _ := json.Marshal(candidates)
	s.redis.Set(ctx, cacheKey, data, 30*time.Minute)

	end := 10
	if end > len(candidates) {
		end = len(candidates)
	}
	return candidates[:end], nil
}

// Swipe records a swipe and checks for mutual match.
func (s *Service) Swipe(ctx context.Context, userID string, req SwipeRequest) (*Match, error) {
	// Record the swipe
	swipe := Swipe{
		SwiperID:     userID,
		SwipedID:     req.CandidateID,
		Action:       req.Action,
		Comment:      req.Comment,
		LikedElement: req.LikedElement,
	}
	if err := s.repo.RecordSwipe(ctx, swipe); err != nil {
		return nil, fmt.Errorf("recording swipe: %w", err)
	}

	// Mark as seen
	s.repo.MarkSeen(ctx, userID, req.CandidateID)
	s.redis.SAdd(ctx, "seen:"+userID, req.CandidateID)

	// If it's a pass, no match check needed
	if req.Action == "pass" {
		return nil, nil
	}

	// Check if the other person already liked us
	otherSwipe, err := s.repo.HasSwiped(ctx, req.CandidateID, userID)
	if err != nil {
		return nil, fmt.Errorf("checking mutual like: %w", err)
	}

	if !otherSwipe {
		// Invalidate their "likes received" cache so they see the new like
		s.redis.Del(ctx, "likes:"+req.CandidateID)
		return nil, nil
	}

	// Mutual like — create match!
	match, err := s.repo.CreateMatch(ctx, userID, req.CandidateID, 0)
	if err != nil {
		return nil, fmt.Errorf("creating match: %w", err)
	}

	// Invalidate match caches for both users
	s.redis.Del(ctx, "matches:"+userID)
	s.redis.Del(ctx, "matches:"+req.CandidateID)

	return match, nil
}

// GetMatches returns cached match list.
func (s *Service) GetMatches(ctx context.Context, userID string) ([]Match, error) {
	cacheKey := "matches:" + userID

	cached, err := s.redis.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var matches []Match
		if json.Unmarshal(cached, &matches) == nil {
			return matches, nil
		}
	}

	matches, err := s.repo.GetMatches(ctx, userID)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(matches)
	s.redis.Set(ctx, cacheKey, data, 15*time.Minute)

	return matches, nil
}

// GetLikesReceived returns who liked this user (Plus+ feature).
func (s *Service) GetLikesReceived(ctx context.Context, userID string, limit int) ([]Swipe, error) {
	cacheKey := "likes:" + userID

	cached, err := s.redis.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var likes []Swipe
		if json.Unmarshal(cached, &likes) == nil {
			return likes, nil
		}
	}

	likes, err := s.repo.GetLikesReceived(ctx, userID, limit)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(likes)
	s.redis.Set(ctx, cacheKey, data, 15*time.Minute)

	return likes, nil
}

// Unmatch removes a match.
func (s *Service) Unmatch(ctx context.Context, userID, matchID string) error {
	status := "unmatched_a"
	if err := s.repo.UpdateMatchStatus(ctx, matchID, status); err != nil {
		return err
	}
	s.redis.Del(ctx, "matches:"+userID)
	return nil
}

// CheckSwipeLimit checks if user has remaining swipes for the day.
func (s *Service) CheckSwipeLimit(ctx context.Context, userID, subscription string) (int, error) {
	key := fmt.Sprintf("ratelimit:%s:swipes", userID)

	count, err := s.redis.Get(ctx, key).Int()
	if err != nil {
		count = 0
	}

	var limit int
	if subscription == "free" {
		limit = s.config.GetInt(ctx, "explore_swipe_limit_free", 15)
	} else {
		limit = s.config.GetInt(ctx, "explore_swipe_limit_plus", 999)
	}

	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}
	return remaining, nil
}

// IncrementSwipeCount tracks daily swipe usage.
func (s *Service) IncrementSwipeCount(ctx context.Context, userID string) {
	key := fmt.Sprintf("ratelimit:%s:swipes", userID)
	s.redis.Incr(ctx, key)
	// Set expiry to end of day if not already set
	ttl := s.redis.TTL(ctx, key).Val()
	if ttl < 0 {
		s.redis.Expire(ctx, key, untilEndOfDay())
	}
}

func (s *Service) findAndScore(ctx context.Context, userID string, filters CandidateFilters) ([]ScoredCandidate, error) {
	// Load all scoring params from admin config
	params := ScoringParams{
		HeritageExactPoints:        s.config.GetInt(ctx, "heritage_exact_match_points", 15),
		HeritageCountryPoints:      s.config.GetInt(ctx, "heritage_same_country_points", 7),
		HeritageRegionPoints:       s.config.GetInt(ctx, "heritage_same_region_points", 4),
		FaithBothImportantPoints:   s.config.GetInt(ctx, "faith_both_important_points", 12),
		FaithSamePoints:            s.config.GetInt(ctx, "faith_same_points", 8),
		FaithBothUnimportantPoints: s.config.GetInt(ctx, "faith_both_unimportant_points", 3),
		DiasporaSamePoints:         s.config.GetInt(ctx, "diaspora_same_tag_points", 10),
		DiasporaSimilarPoints:      s.config.GetInt(ctx, "diaspora_similar_tag_points", 7),
		DiasporaCrossPoints:        s.config.GetInt(ctx, "diaspora_cross_continental_points", 5),
		Proximity10kmPoints:        s.config.GetInt(ctx, "proximity_10km_points", 10),
		Proximity50kmPoints:        s.config.GetInt(ctx, "proximity_50km_points", 8),
		Proximity200kmPoints:       s.config.GetInt(ctx, "proximity_200km_points", 5),
		ProximityCountryPoints:     s.config.GetInt(ctx, "proximity_same_country_points", 3),
	}

	// Stage 1+2: Candidate generation + scoring (single SQL query)
	candidates, err := s.repo.FindCandidates(ctx, userID, filters, params)
	if err != nil {
		return nil, fmt.Errorf("finding candidates: %w", err)
	}

	// Stage 3: Re-ranking (business rules in Go)
	candidates = s.rerank(ctx, candidates)

	return candidates, nil
}

func (s *Service) rerank(ctx context.Context, candidates []ScoredCandidate) []ScoredCandidate {
	spotlightBoost := s.config.GetInt(ctx, "spotlight_boost_percent", 20)
	newUserBoostPct := s.config.GetInt(ctx, "new_user_boost_percent", 10)
	newUserBoostDays := s.config.GetInt(ctx, "new_user_boost_days", 7)
	reciprocityBoost := s.config.GetInt(ctx, "reciprocity_boost_percent", 15)
	trustThreshold := s.config.GetInt(ctx, "trust_score_penalty_threshold", 30)
	trustPenalty := s.config.GetFloat(ctx, "trust_score_penalty_factor", 0.5)

	for i := range candidates {
		c := &candidates[i]

		// Rose → always top of queue
		if c.SentRose {
			c.FinalScore = 99999
			continue
		}

		c.FinalScore = float64(c.MatchScore)

		// Spotlight boost
		if c.HasSpotlight {
			c.FinalScore *= (1 + float64(spotlightBoost)/100)
		}

		// New user boost
		if time.Since(c.CreatedAt) < time.Duration(newUserBoostDays)*24*time.Hour {
			c.FinalScore *= (1 + float64(newUserBoostPct)/100)
		}

		// Reciprocity boost
		if c.HasLikedMe {
			c.FinalScore *= (1 + float64(reciprocityBoost)/100)
		}

		// Trust penalty
		if c.TrustScore < trustThreshold {
			c.FinalScore *= trustPenalty
		}
	}

	// Sort by final score descending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].FinalScore > candidates[j].FinalScore
	})

	// Apply diversity rules
	candidates = s.applyDiversity(ctx, candidates)

	return candidates
}

func (s *Service) applyDiversity(ctx context.Context, candidates []ScoredCandidate) []ScoredCandidate {
	maxSameCity := s.config.GetInt(ctx, "diversity_max_same_city", 3)
	maxSameHeritage := s.config.GetInt(ctx, "diversity_max_same_heritage", 2)

	cityCounts := map[string]int{}
	heritageCounts := map[string]int{}
	var result []ScoredCandidate

	for _, c := range candidates {
		if cityCounts[c.City] >= maxSameCity {
			continue
		}
		primaryHeritage := ""
		if len(c.Heritage) > 0 {
			primaryHeritage = c.Heritage[0]
		}
		if primaryHeritage != "" && heritageCounts[primaryHeritage] >= maxSameHeritage {
			continue
		}
		cityCounts[c.City]++
		if primaryHeritage != "" {
			heritageCounts[primaryHeritage]++
		}
		result = append(result, c)
	}

	return result
}

func untilEndOfDay() time.Duration {
	now := time.Now()
	endOfDay := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return endOfDay.Sub(now)
}
