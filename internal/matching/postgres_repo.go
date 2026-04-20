package matching

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepo struct {
	db *pgxpool.Pool
}

func NewPostgresRepo(db *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{db: db}
}

// FindCandidates runs the core matching query — candidate generation + scoring in a single SQL query.
// All scoring point values come from ScoringParams (loaded from admin_config).
func (r *PostgresRepo) FindCandidates(ctx context.Context, userID string, filters CandidateFilters, params ScoringParams) ([]ScoredCandidate, error) {
	// Convert age range to date of birth range
	now := time.Now()
	maxDOB := now.AddDate(-filters.MinAge, 0, 0)  // youngest allowed
	minDOB := now.AddDate(-filters.MaxAge, 0, 0)   // oldest allowed

	query := `
		SELECT
			p.user_id,
			p.first_name,
			EXTRACT(YEAR FROM AGE(p.date_of_birth))::int AS age,
			p.city,
			p.country,
			p.heritage,
			p.diaspora_tag,
			p.intention,
			p.faith,
			u.verification,
			u.trust_score,
			u.last_active_at,
			u.created_at,
			COALESCE(
				(SELECT url_medium FROM photos ph
				 WHERE ph.user_id = p.user_id AND ph.is_primary = true
				   AND ph.moderation_status = 'approved'
				 LIMIT 1),
				''
			) AS primary_photo,
			EXISTS(
				SELECT 1 FROM swipes s
				WHERE s.swiper_id = p.user_id AND s.swiped_id = $1 AND s.action = 'like'
			) AS has_liked_me,
			(
				-- Cultural Affinity
				(CASE WHEN p.heritage && $10 THEN $14
				      WHEN p.country = $11 THEN $15
				      ELSE 0 END)
				+
				(CASE WHEN p.faith = $12 AND p.faith_importance = 'very_important'
				           AND $13 = 'very_important' THEN $17
				      WHEN p.faith = $12 THEN $18
				      WHEN p.faith_importance = 'not_important'
				           AND $13 = 'not_important' THEN $19
				      ELSE 0 END)
				+
				-- Diaspora Fit
				(CASE WHEN p.diaspora_tag = $9 THEN $20
				      WHEN p.diaspora_tag IN ('diaspora_1st','diaspora_2nd')
				           AND $9 IN ('diaspora_1st','diaspora_2nd') THEN $21
				      WHEN p.heritage && $10 THEN $22
				      ELSE 2 END)
				+
				-- Proximity
				(CASE WHEN ST_DWithin(p.location, ST_MakePoint($7, $6)::geography, 10000) THEN $23
				      WHEN ST_DWithin(p.location, ST_MakePoint($7, $6)::geography, 50000) THEN $24
				      WHEN ST_DWithin(p.location, ST_MakePoint($7, $6)::geography, 200000) THEN $25
				      WHEN p.country = $11 THEN $26
				      ELSE 0 END)
				+
				-- Profile Quality
				(CASE WHEN u.verification = 'photo_verified' THEN 4 ELSE 0 END)
				+ LEAST(p.completeness / 34, 3)
				+ LEAST(
					(SELECT COUNT(*) FROM photos ph2
					 WHERE ph2.user_id = p.user_id AND ph2.moderation_status = 'approved')::int / 2, 2)
				+
				-- Recency
				(CASE WHEN u.last_active_at > NOW() - INTERVAL '1 hour' THEN 10
				      WHEN u.last_active_at > NOW() - INTERVAL '24 hours' THEN 7
				      WHEN u.last_active_at > NOW() - INTERVAL '3 days' THEN 4
				      WHEN u.last_active_at > NOW() - INTERVAL '7 days' THEN 2
				      ELSE 0 END)
			) AS match_score
		FROM profiles p
		JOIN users u ON u.id = p.user_id
		WHERE u.status = 'active'
			AND u.id != $1
			AND (p.gender = $2 OR $2 = 'everyone')
			AND (p.gender_pref = $3 OR p.gender_pref = 'everyone')
			AND (p.intention = $4 OR p.intention = 'both' OR $4 = 'both')
			AND p.date_of_birth BETWEEN $5 AND $16
			AND p.completeness >= $8
			AND u.id NOT IN (SELECT seen_user_id FROM seen_profiles WHERE user_id = $1)
			AND u.id NOT IN (SELECT blocked_user_id FROM blocklist WHERE user_id = $1)
		ORDER BY match_score DESC
		LIMIT $27
	`

	rows, err := r.db.Query(ctx, query,
		userID,                          // $1
		filters.GenderPref,              // $2  target gender
		filters.Gender,                  // $3  my gender (they must prefer)
		filters.Intention,               // $4
		minDOB,                          // $5  oldest DOB
		filters.Latitude,                // $6
		filters.Longitude,               // $7
		50,                              // $8  min completeness
		filters.DiasporaTag,             // $9  my diaspora tag
		filters.Heritage,                // $10 my heritage array
		filters.Country,                 // $11 my country
		filters.Faith,                   // $12 my faith
		filters.FaithImportance,         // $13 my faith importance
		params.HeritageExactPoints,      // $14
		params.HeritageCountryPoints,    // $15
		maxDOB,                          // $16 youngest DOB
		params.FaithBothImportantPoints, // $17
		params.FaithSamePoints,          // $18
		params.FaithBothUnimportantPoints, // $19
		params.DiasporaSamePoints,       // $20
		params.DiasporaSimilarPoints,    // $21
		params.DiasporaCrossPoints,      // $22
		params.Proximity10kmPoints,      // $23
		params.Proximity50kmPoints,      // $24
		params.Proximity200kmPoints,     // $25
		params.ProximityCountryPoints,   // $26
		filters.Limit,                   // $27
	)
	if err != nil {
		return nil, fmt.Errorf("finding candidates: %w", err)
	}
	defer rows.Close()

	var candidates []ScoredCandidate
	for rows.Next() {
		var c ScoredCandidate
		err := rows.Scan(
			&c.UserID, &c.FirstName, &c.Age, &c.City, &c.Country,
			&c.Heritage, &c.DiasporaTag, &c.Intention, &c.Faith,
			&c.Verification, &c.TrustScore, &c.LastActiveAt, &c.CreatedAt,
			&c.PrimaryPhoto, &c.HasLikedMe, &c.MatchScore,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning candidate: %w", err)
		}
		c.FinalScore = float64(c.MatchScore)
		candidates = append(candidates, c)
	}

	return candidates, nil
}

func (r *PostgresRepo) RecordSwipe(ctx context.Context, swipe Swipe) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO swipes (swiper_id, swiped_id, action, comment, liked_element)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (swiper_id, swiped_id) DO UPDATE SET action = $3, comment = $4
	`, swipe.SwiperID, swipe.SwipedID, swipe.Action, swipe.Comment, swipe.LikedElement)
	return err
}

func (r *PostgresRepo) HasSwiped(ctx context.Context, swiperID, swipedID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM swipes
			WHERE swiper_id = $1 AND swiped_id = $2 AND action = 'like'
		)
	`, swiperID, swipedID).Scan(&exists)
	return exists, err
}

func (r *PostgresRepo) GetLikesReceived(ctx context.Context, userID string, limit int) ([]Swipe, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, swiper_id, swiped_id, action, comment, liked_element, created_at
		FROM swipes
		WHERE swiped_id = $1 AND action = 'like'
		  AND swiper_id NOT IN (SELECT swiped_id FROM swipes WHERE swiper_id = $1)
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var swipes []Swipe
	for rows.Next() {
		var s Swipe
		if err := rows.Scan(&s.ID, &s.SwiperID, &s.SwipedID, &s.Action, &s.Comment, &s.LikedElement, &s.CreatedAt); err != nil {
			return nil, err
		}
		swipes = append(swipes, s)
	}
	return swipes, nil
}

func (r *PostgresRepo) CreateMatch(ctx context.Context, userAID, userBID string, compatibility int) (*Match, error) {
	// Ensure consistent ordering (smaller ID first)
	if userAID > userBID {
		userAID, userBID = userBID, userAID
	}

	var m Match
	err := r.db.QueryRow(ctx, `
		INSERT INTO matches (user_a_id, user_b_id, compatibility)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_a_id, user_b_id) DO UPDATE SET status = 'active'
		RETURNING id, user_a_id, user_b_id, compatibility, status, matched_at, last_message_at
	`, userAID, userBID, compatibility).Scan(
		&m.ID, &m.UserAID, &m.UserBID, &m.Compatibility, &m.Status, &m.MatchedAt, &m.LastMessageAt,
	)
	if err != nil {
		return nil, fmt.Errorf("creating match: %w", err)
	}
	return &m, nil
}

func (r *PostgresRepo) GetMatch(ctx context.Context, userAID, userBID string) (*Match, error) {
	if userAID > userBID {
		userAID, userBID = userBID, userAID
	}

	var m Match
	err := r.db.QueryRow(ctx, `
		SELECT id, user_a_id, user_b_id, compatibility, status, matched_at, last_message_at
		FROM matches WHERE user_a_id = $1 AND user_b_id = $2
	`, userAID, userBID).Scan(
		&m.ID, &m.UserAID, &m.UserBID, &m.Compatibility, &m.Status, &m.MatchedAt, &m.LastMessageAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *PostgresRepo) GetMatches(ctx context.Context, userID string) ([]Match, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_a_id, user_b_id, compatibility, status, matched_at, last_message_at
		FROM matches
		WHERE (user_a_id = $1 OR user_b_id = $1) AND status = 'active'
		ORDER BY COALESCE(last_message_at, matched_at) DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []Match
	for rows.Next() {
		var m Match
		if err := rows.Scan(&m.ID, &m.UserAID, &m.UserBID, &m.Compatibility, &m.Status, &m.MatchedAt, &m.LastMessageAt); err != nil {
			return nil, err
		}
		matches = append(matches, m)
	}
	return matches, nil
}

func (r *PostgresRepo) UpdateMatchStatus(ctx context.Context, matchID, status string) error {
	_, err := r.db.Exec(ctx,
		"UPDATE matches SET status = $2 WHERE id = $1",
		matchID, status,
	)
	return err
}

func (r *PostgresRepo) UpdateMatchLastMessage(ctx context.Context, matchID string) error {
	_, err := r.db.Exec(ctx,
		"UPDATE matches SET last_message_at = NOW() WHERE id = $1",
		matchID,
	)
	return err
}

func (r *PostgresRepo) MarkSeen(ctx context.Context, userID, seenUserID string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO seen_profiles (user_id, seen_user_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, userID, seenUserID)
	return err
}

func (r *PostgresRepo) GetSeenIDs(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.db.Query(ctx,
		"SELECT seen_user_id FROM seen_profiles WHERE user_id = $1",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		rows.Scan(&id)
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *PostgresRepo) SaveDailyCircle(ctx context.Context, circle DailyCircle) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO daily_circles (user_id, candidate_ids)
		VALUES ($1, $2)
	`, circle.UserID, circle.CandidateIDs)
	return err
}

func (r *PostgresRepo) GetLatestCircle(ctx context.Context, userID string) (*DailyCircle, error) {
	var dc DailyCircle
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, candidate_ids, delivered_at, opened, actions_taken
		FROM daily_circles WHERE user_id = $1
		ORDER BY delivered_at DESC LIMIT 1
	`, userID).Scan(&dc.ID, &dc.UserID, &dc.CandidateIDs, &dc.DeliveredAt, &dc.Opened, &dc.ActionsTaken)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &dc, nil
}
