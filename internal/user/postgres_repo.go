package user

import (
	"context"
	"encoding/json"
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

func (r *PostgresRepo) CreateUser(ctx context.Context, telegramID int64, telegramUsername string) (*User, error) {
	var u User
	err := r.db.QueryRow(ctx, `
		INSERT INTO users (telegram_id, telegram_username)
		VALUES ($1, $2)
		RETURNING id, telegram_id, telegram_username, status, verification, trust_score,
		          subscription, sub_expires_at, terms_accepted_at, privacy_accepted_at,
		          created_at, updated_at, last_active_at
	`, telegramID, telegramUsername).Scan(
		&u.ID, &u.TelegramID, &u.TelegramUsername, &u.Status, &u.Verification,
		&u.TrustScore, &u.Subscription, &u.SubExpiresAt, &u.TermsAcceptedAt, &u.PrivacyAcceptedAt,
		&u.CreatedAt, &u.UpdatedAt, &u.LastActiveAt,
	)
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}
	return &u, nil
}

func (r *PostgresRepo) GetUserByID(ctx context.Context, id string) (*User, error) {
	var u User
	err := r.db.QueryRow(ctx, `
		SELECT id, telegram_id, telegram_username, status, verification, trust_score,
		       subscription, sub_expires_at, terms_accepted_at, privacy_accepted_at,
		       created_at, updated_at, last_active_at
		FROM users WHERE id = $1
	`, id).Scan(
		&u.ID, &u.TelegramID, &u.TelegramUsername, &u.Status, &u.Verification,
		&u.TrustScore, &u.Subscription, &u.SubExpiresAt, &u.TermsAcceptedAt, &u.PrivacyAcceptedAt,
		&u.CreatedAt, &u.UpdatedAt, &u.LastActiveAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting user by ID: %w", err)
	}
	return &u, nil
}

func (r *PostgresRepo) GetUserByTelegramID(ctx context.Context, telegramID int64) (*User, error) {
	var u User
	err := r.db.QueryRow(ctx, `
		SELECT id, telegram_id, telegram_username, status, verification, trust_score,
		       subscription, sub_expires_at, terms_accepted_at, privacy_accepted_at,
		       created_at, updated_at, last_active_at
		FROM users WHERE telegram_id = $1
	`, telegramID).Scan(
		&u.ID, &u.TelegramID, &u.TelegramUsername, &u.Status, &u.Verification,
		&u.TrustScore, &u.Subscription, &u.SubExpiresAt, &u.TermsAcceptedAt, &u.PrivacyAcceptedAt,
		&u.CreatedAt, &u.UpdatedAt, &u.LastActiveAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting user by telegram ID: %w", err)
	}
	return &u, nil
}

func (r *PostgresRepo) UpdateUserStatus(ctx context.Context, id string, status string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users SET status = $2, updated_at = NOW() WHERE id = $1
	`, id, status)
	return err
}

func (r *PostgresRepo) UpdateUserVerification(ctx context.Context, id string, verification string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users SET verification = $2, updated_at = NOW() WHERE id = $1
	`, id, verification)
	return err
}

func (r *PostgresRepo) UpdateUserSubscription(ctx context.Context, id string, subscription string, expiresAt *string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users SET subscription = $2, sub_expires_at = $3, updated_at = NOW() WHERE id = $1
	`, id, subscription, expiresAt)
	return err
}

func (r *PostgresRepo) UpdateLastActive(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users SET last_active_at = NOW() WHERE id = $1
	`, id)
	return err
}

func (r *PostgresRepo) AcceptTerms(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users SET terms_accepted_at = NOW(), privacy_accepted_at = NOW(), updated_at = NOW() WHERE id = $1
	`, id)
	return err
}

func (r *PostgresRepo) DeleteUser(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users SET status = 'deleted', updated_at = NOW() WHERE id = $1
	`, id)
	return err
}

func (r *PostgresRepo) CreateProfile(ctx context.Context, userID string, req CreateProfileRequest) (*Profile, error) {
	culturalJSON, _ := json.Marshal(req.CulturalPrompts)
	personalityJSON, _ := json.Marshal(req.PersonalityPrompts)
	completeness := 0 // recalculated on read by service.recalcCompleteness

	_, err := r.db.Exec(ctx, `
		INSERT INTO profiles (
			user_id, first_name, date_of_birth, gender, gender_pref,
			city, country, latitude, longitude, location,
			heritage, diaspora_tag, intention, faith, faith_importance,
			bio, cultural_prompts, personality_prompts, completeness
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, ST_MakePoint($19::float8, $20::float8)::geography,
			$10, $11, $12, $13, $14,
			$15, $16, $17, $18
		)
		ON CONFLICT (user_id) DO UPDATE SET
			first_name = EXCLUDED.first_name,
			date_of_birth = EXCLUDED.date_of_birth,
			gender = EXCLUDED.gender,
			gender_pref = EXCLUDED.gender_pref,
			city = EXCLUDED.city,
			country = EXCLUDED.country,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			location = EXCLUDED.location,
			heritage = EXCLUDED.heritage,
			diaspora_tag = EXCLUDED.diaspora_tag,
			intention = EXCLUDED.intention,
			faith = EXCLUDED.faith,
			faith_importance = EXCLUDED.faith_importance,
			bio = EXCLUDED.bio,
			cultural_prompts = EXCLUDED.cultural_prompts,
			personality_prompts = EXCLUDED.personality_prompts,
			completeness = EXCLUDED.completeness,
			updated_at = NOW()
	`,
		userID, req.FirstName, req.DateOfBirth, req.Gender, req.GenderPref,
		req.City, req.Country, req.Latitude, req.Longitude,
		req.Heritage, req.DiasporaTag, req.Intention, req.Faith, req.FaithImportance,
		req.Bio, culturalJSON, personalityJSON, completeness,
		req.Longitude, req.Latitude,
	)
	if err != nil {
		return nil, fmt.Errorf("creating profile: %w", err)
	}

	// Fetch the created/updated profile with proper type scanning
	return r.GetProfile(ctx, userID)
}

func (r *PostgresRepo) GetProfile(ctx context.Context, userID string) (*Profile, error) {
	var p Profile
	var culturalJSON, personalityJSON []byte
	var dob time.Time

	err := r.db.QueryRow(ctx, `
		SELECT user_id, first_name, date_of_birth, gender, gender_pref,
		       COALESCE(city, ''), COALESCE(country, ''), COALESCE(latitude, 0), COALESCE(longitude, 0),
		       COALESCE(heritage, '{}'), COALESCE(diaspora_tag, ''), COALESCE(intention, ''),
		       COALESCE(faith, ''), COALESCE(faith_importance, ''),
		       COALESCE(bio, ''), COALESCE(audio_bio_url, ''),
		       COALESCE(cultural_prompts, '[]'::jsonb), COALESCE(personality_prompts, '[]'::jsonb),
		       COALESCE(completeness, 0),
		       created_at, updated_at
		FROM profiles WHERE user_id = $1
	`, userID).Scan(
		&p.UserID, &p.FirstName, &dob, &p.Gender, &p.GenderPref,
		&p.City, &p.Country, &p.Latitude, &p.Longitude,
		&p.Heritage, &p.DiasporaTag, &p.Intention, &p.Faith, &p.FaithImportance,
		&p.Bio, &p.AudioBioURL, &culturalJSON, &personalityJSON, &p.Completeness,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting profile: %w", err)
	}

	p.DateOfBirth = dob.Format("2006-01-02")
	json.Unmarshal(culturalJSON, &p.CulturalPrompts)
	json.Unmarshal(personalityJSON, &p.PersonalityPrompts)

	// Load photos
	photos, err := r.GetPhotos(ctx, userID)
	if err == nil {
		p.Photos = photos
	}

	return &p, nil
}

func (r *PostgresRepo) UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) error {
	// Build dynamic update query based on which fields are set
	setClauses := []string{}
	args := []interface{}{}
	argNum := 1

	addField := func(clause string, val interface{}) {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", clause, argNum))
		args = append(args, val)
		argNum++
	}

	if req.FirstName != nil {
		addField("first_name", *req.FirstName)
	}
	if req.City != nil {
		addField("city", *req.City)
	}
	if req.Country != nil {
		addField("country", *req.Country)
	}
	if req.Latitude != nil && req.Longitude != nil {
		addField("latitude", *req.Latitude)
		addField("longitude", *req.Longitude)
	}
	if req.Heritage != nil {
		addField("heritage", req.Heritage)
	}
	if req.DiasporaTag != nil {
		addField("diaspora_tag", *req.DiasporaTag)
	}
	if req.Intention != nil {
		addField("intention", *req.Intention)
	}
	if req.Faith != nil {
		addField("faith", *req.Faith)
	}
	if req.FaithImportance != nil {
		addField("faith_importance", *req.FaithImportance)
	}
	if req.Bio != nil {
		addField("bio", *req.Bio)
	}
	if req.CulturalPrompts != nil {
		j, _ := json.Marshal(req.CulturalPrompts)
		addField("cultural_prompts", j)
	}
	if req.PersonalityPrompts != nil {
		j, _ := json.Marshal(req.PersonalityPrompts)
		addField("personality_prompts", j)
	}

	if len(setClauses) == 0 {
		return nil
	}

	setClauses = append(setClauses, "updated_at = NOW()")

	query := "UPDATE profiles SET "
	for i, clause := range setClauses {
		if i > 0 {
			query += ", "
		}
		query += clause
	}
	query += fmt.Sprintf(" WHERE user_id = $%d", argNum)
	args = append(args, userID)

	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("updating profile: %w", err)
	}

	// Update location geography column if coordinates changed
	if req.Latitude != nil && req.Longitude != nil {
		_, err = r.db.Exec(ctx,
			"UPDATE profiles SET location = ST_MakePoint($2::float8, $3::float8)::geography WHERE user_id = $1",
			userID, *req.Longitude, *req.Latitude,
		)
		if err != nil {
			return fmt.Errorf("updating location: %w", err)
		}
	}

	return nil
}

func (r *PostgresRepo) UpdateCompleteness(ctx context.Context, userID string, completeness int) error {
	_, err := r.db.Exec(ctx,
		"UPDATE profiles SET completeness = $2, updated_at = NOW() WHERE user_id = $1",
		userID, completeness,
	)
	return err
}

func (r *PostgresRepo) AddPhoto(ctx context.Context, photo Photo) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO photos (id, user_id, url_thumbnail, url_medium, url_large, position, is_primary, moderation_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'approved')
	`, photo.ID, photo.UserID, photo.URLThumbnail, photo.URLMedium, photo.URLLarge, photo.Position, photo.IsPrimary)
	return err
}

func (r *PostgresRepo) GetPhotos(ctx context.Context, userID string) ([]Photo, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, url_thumbnail, url_medium, url_large, position, is_primary,
		       moderation_status, created_at
		FROM photos WHERE user_id = $1 ORDER BY position
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []Photo
	for rows.Next() {
		var p Photo
		if err := rows.Scan(&p.ID, &p.UserID, &p.URLThumbnail, &p.URLMedium, &p.URLLarge,
			&p.Position, &p.IsPrimary, &p.ModerationStatus, &p.CreatedAt); err != nil {
			return nil, err
		}
		photos = append(photos, p)
	}
	return photos, nil
}

func (r *PostgresRepo) DeletePhoto(ctx context.Context, photoID string, userID string) error {
	_, err := r.db.Exec(ctx,
		"DELETE FROM photos WHERE id = $1 AND user_id = $2",
		photoID, userID,
	)
	return err
}

func (r *PostgresRepo) ReorderPhotos(ctx context.Context, userID string, photoIDs []string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i, photoID := range photoIDs {
		_, err := tx.Exec(ctx,
			"UPDATE photos SET position = $3, is_primary = $4 WHERE id = $1 AND user_id = $2",
			photoID, userID, i, i == 0,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresRepo) UpdatePhotoModeration(ctx context.Context, photoID string, status string) error {
	_, err := r.db.Exec(ctx,
		"UPDATE photos SET moderation_status = $2 WHERE id = $1",
		photoID, status,
	)
	return err
}

func (r *PostgresRepo) BlockUser(ctx context.Context, blockerID, blockedID string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO blocklist (user_id, blocked_user_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, blockerID, blockedID)
	return err
}

func (r *PostgresRepo) UnblockUser(ctx context.Context, blockerID, blockedID string) error {
	_, err := r.db.Exec(ctx,
		"DELETE FROM blocklist WHERE user_id = $1 AND blocked_user_id = $2",
		blockerID, blockedID,
	)
	return err
}

func (r *PostgresRepo) IsBlocked(ctx context.Context, userA, userB string) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM blocklist
		WHERE (user_id = $1 AND blocked_user_id = $2)
		   OR (user_id = $2 AND blocked_user_id = $1)
	`, userA, userB).Scan(&count)
	return count > 0, err
}

func (r *PostgresRepo) GetBlockList(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.db.Query(ctx,
		"SELECT blocked_user_id FROM blocklist WHERE user_id = $1",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

