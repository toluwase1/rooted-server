package matching_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/rooted-dating/rooted-server/internal/matching"
	"github.com/rooted-dating/rooted-server/internal/shared/config"
	"github.com/rooted-dating/rooted-server/internal/shared/database"
)

var (
	testDB      *pgxpool.Pool
	testRepo    *matching.PostgresRepo
	testService *matching.Service
	testCtx     = context.Background()
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Start PostGIS container
	pgContainer, err := postgres.Run(ctx, "postgis/postgis:14-3.4-alpine",
		postgres.WithDatabase("rooted_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to start postgres container: %v", err))
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(fmt.Sprintf("Failed to get connection string: %v", err))
	}

	testDB, err = database.NewPostgres(ctx, connStr)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect: %v", err))
	}

	// Run migrations
	if err := runMigrations(ctx, testDB); err != nil {
		panic(fmt.Sprintf("Failed to run migrations: %v", err))
	}

	// Setup repo and service
	testRepo = matching.NewPostgresRepo(testDB)
	rdb := database.NewSafeRedis(nil) // no Redis in tests
	dynConfig := config.NewDynamicConfig(testDB, rdb)
	testService = matching.NewService(testRepo, rdb, dynConfig)

	code := m.Run()

	testDB.Close()
	pgContainer.Terminate(ctx)
	os.Exit(code)
}

func runMigrations(ctx context.Context, db *pgxpool.Pool) error {
	migration := `
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
		CREATE EXTENSION IF NOT EXISTS "postgis";

		CREATE TABLE users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			telegram_id BIGINT UNIQUE NOT NULL,
			telegram_username VARCHAR(64),
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			last_active_at TIMESTAMPTZ DEFAULT NOW(),
			status VARCHAR(20) DEFAULT 'active',
			verification VARCHAR(20) DEFAULT 'unverified',
			trust_score SMALLINT DEFAULT 50,
			subscription VARCHAR(20) DEFAULT 'free',
			sub_expires_at TIMESTAMPTZ
		);

		CREATE TABLE profiles (
			user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			first_name VARCHAR(50) NOT NULL,
			date_of_birth DATE NOT NULL,
			gender VARCHAR(20) NOT NULL,
			gender_pref VARCHAR(20) NOT NULL,
			city VARCHAR(100),
			country VARCHAR(3),
			latitude DECIMAL(10, 7),
			longitude DECIMAL(10, 7),
			location GEOGRAPHY(POINT, 4326),
			heritage VARCHAR(50)[] DEFAULT '{}',
			diaspora_tag VARCHAR(30),
			intention VARCHAR(20),
			faith VARCHAR(30),
			faith_importance VARCHAR(20),
			bio VARCHAR(150),
			audio_bio_url TEXT,
			cultural_prompts JSONB DEFAULT '[]',
			personality_prompts JSONB DEFAULT '[]',
			completeness SMALLINT DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE INDEX idx_profiles_location ON profiles USING GIST(location);
		CREATE INDEX idx_profiles_heritage ON profiles USING GIN(heritage);
		CREATE INDEX idx_profiles_gender_pref ON profiles(gender, gender_pref);

		CREATE TABLE photos (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			url_thumbnail TEXT NOT NULL,
			url_medium TEXT NOT NULL,
			url_large TEXT NOT NULL,
			position SMALLINT NOT NULL,
			is_primary BOOLEAN DEFAULT FALSE,
			moderation_status VARCHAR(20) DEFAULT 'approved',
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE TABLE swipes (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			swiper_id UUID NOT NULL REFERENCES users(id),
			swiped_id UUID NOT NULL REFERENCES users(id),
			action VARCHAR(10) NOT NULL,
			comment TEXT,
			liked_element VARCHAR(50),
			created_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(swiper_id, swiped_id)
		);

		CREATE TABLE matches (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_a_id UUID NOT NULL REFERENCES users(id),
			user_b_id UUID NOT NULL REFERENCES users(id),
			compatibility SMALLINT,
			status VARCHAR(20) DEFAULT 'active',
			matched_at TIMESTAMPTZ DEFAULT NOW(),
			last_message_at TIMESTAMPTZ,
			UNIQUE(user_a_id, user_b_id)
		);

		CREATE TABLE seen_profiles (
			user_id UUID NOT NULL REFERENCES users(id),
			seen_user_id UUID NOT NULL REFERENCES users(id),
			seen_at TIMESTAMPTZ DEFAULT NOW(),
			PRIMARY KEY (user_id, seen_user_id)
		);

		CREATE TABLE blocklist (
			user_id UUID NOT NULL REFERENCES users(id),
			blocked_user_id UUID NOT NULL REFERENCES users(id),
			created_at TIMESTAMPTZ DEFAULT NOW(),
			PRIMARY KEY (user_id, blocked_user_id)
		);

		CREATE TABLE daily_circles (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL REFERENCES users(id),
			candidate_ids UUID[] NOT NULL,
			delivered_at TIMESTAMPTZ DEFAULT NOW(),
			opened BOOLEAN DEFAULT FALSE,
			actions_taken INT DEFAULT 0
		);

		CREATE TABLE admin_config (
			key VARCHAR(100) PRIMARY KEY,
			value JSONB NOT NULL,
			category VARCHAR(50) NOT NULL,
			description TEXT,
			value_type VARCHAR(20) NOT NULL,
			constraints JSONB,
			updated_by UUID,
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		-- Seed default config for matching
		INSERT INTO admin_config (key, value, category, description, value_type) VALUES
		('daily_circle_size', '8', 'matching', 'Daily matches', 'int'),
		('explore_swipe_limit_free', '15', 'matching', 'Daily swipes free', 'int'),
		('explore_swipe_limit_plus', '999', 'matching', 'Daily swipes paid', 'int'),
		('heritage_exact_match_points', '15', 'matching', '', 'int'),
		('heritage_same_country_points', '7', 'matching', '', 'int'),
		('heritage_same_region_points', '4', 'matching', '', 'int'),
		('faith_both_important_points', '12', 'matching', '', 'int'),
		('faith_same_points', '8', 'matching', '', 'int'),
		('faith_both_unimportant_points', '3', 'matching', '', 'int'),
		('diaspora_same_tag_points', '10', 'matching', '', 'int'),
		('diaspora_similar_tag_points', '7', 'matching', '', 'int'),
		('diaspora_cross_continental_points', '5', 'matching', '', 'int'),
		('proximity_10km_points', '10', 'matching', '', 'int'),
		('proximity_50km_points', '8', 'matching', '', 'int'),
		('proximity_200km_points', '5', 'matching', '', 'int'),
		('proximity_same_country_points', '3', 'matching', '', 'int'),
		('new_user_boost_percent', '10', 'matching', '', 'int'),
		('new_user_boost_days', '7', 'matching', '', 'int'),
		('spotlight_boost_percent', '20', 'matching', '', 'int'),
		('reciprocity_boost_percent', '15', 'matching', '', 'int'),
		('diversity_max_same_city', '3', 'matching', '', 'int'),
		('diversity_max_same_heritage', '2', 'matching', '', 'int'),
		('trust_score_penalty_threshold', '30', 'matching', '', 'int'),
		('trust_score_penalty_factor', '0.5', 'matching', '', 'float');
	`
	_, err := db.Exec(ctx, migration)
	return err
}

// ============================================================
// HELPERS
// ============================================================

type testUser struct {
	id              string
	telegramID      int64
	firstName       string
	gender          string
	genderPref      string
	city            string
	country         string
	lat, lon        float64
	heritage        []string
	diasporaTag     string
	intention       string
	faith           string
	faithImportance string
	verification    string
	trustScore      int
}

func createTestUser(t *testing.T, u testUser) string {
	t.Helper()

	var id string
	err := testDB.QueryRow(testCtx, `
		INSERT INTO users (telegram_id, telegram_username, status, verification, trust_score)
		VALUES ($1, $2, 'active', $3, $4)
		RETURNING id
	`, u.telegramID, u.firstName, u.verification, u.trustScore).Scan(&id)
	require.NoError(t, err)

	_, err = testDB.Exec(testCtx, `
		INSERT INTO profiles (user_id, first_name, date_of_birth, gender, gender_pref,
			city, country, latitude, longitude, location,
			heritage, diaspora_tag, intention, faith, faith_importance,
			cultural_prompts, personality_prompts, completeness)
		VALUES ($1, $2, '1995-01-01', $3, $4,
			$5, $6, $7, $8, ST_MakePoint($9::float8, $10::float8)::geography,
			$11, $12, $13, $14, $15,
			'[]'::jsonb, '[]'::jsonb, 70)
	`, id, u.firstName, u.gender, u.genderPref,
		u.city, u.country, u.lat, u.lon, u.lon, u.lat,
		u.heritage, u.diasporaTag, u.intention, u.faith, u.faithImportance)
	require.NoError(t, err)

	return id
}

func cleanDB(t *testing.T) {
	t.Helper()
	testDB.Exec(testCtx, "DELETE FROM daily_circles")
	testDB.Exec(testCtx, "DELETE FROM seen_profiles")
	testDB.Exec(testCtx, "DELETE FROM swipes")
	testDB.Exec(testCtx, "DELETE FROM matches")
	testDB.Exec(testCtx, "DELETE FROM blocklist")
	testDB.Exec(testCtx, "DELETE FROM photos")
	testDB.Exec(testCtx, "DELETE FROM profiles")
	testDB.Exec(testCtx, "DELETE FROM users")
}

func defaultFilters(gender, genderPref, intention, faith, faithImportance, diasporaTag string, heritage []string, lat, lon float64, country string) matching.CandidateFilters {
	return matching.CandidateFilters{
		Gender:          gender,
		GenderPref:      genderPref,
		Intention:       intention,
		MinAge:          18,
		MaxAge:          99,
		Country:         country,
		Heritage:        heritage,
		DiasporaTag:     diasporaTag,
		Faith:           faith,
		FaithImportance: faithImportance,
		Latitude:        lat,
		Longitude:       lon,
		Limit:           20,
	}
}

// ============================================================
// TESTS: Gender Matching
// ============================================================

func TestMatching_GenderFilter(t *testing.T) {
	cleanDB(t)

	maleID := createTestUser(t, testUser{
		telegramID: 1, firstName: "Tolu", gender: "male", genderPref: "female",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", faith: "christian", faithImportance: "somewhat",
		verification: "unverified", trustScore: 50,
	})

	createTestUser(t, testUser{
		telegramID: 2, firstName: "Ada", gender: "female", genderPref: "male",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", faith: "christian", faithImportance: "somewhat",
		verification: "unverified", trustScore: 50,
	})

	createTestUser(t, testUser{
		telegramID: 3, firstName: "Emeka", gender: "male", genderPref: "female",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", faith: "christian", faithImportance: "somewhat",
		verification: "unverified", trustScore: 50,
	})

	filters := defaultFilters("male", "female", "dating", "christian", "somewhat", "born_in_africa", []string{"nigerian"}, 6.52, 3.37, "NG")
	candidates, err := testService.GetExploreFeed(testCtx, maleID, filters, 0)
	require.NoError(t, err)

	// Should only see Ada (female), not Emeka (male)
	assert.Len(t, candidates, 1)
	assert.Equal(t, "Ada", candidates[0].FirstName)
}

func TestMatching_GenderPrefEveryone(t *testing.T) {
	cleanDB(t)

	userID := createTestUser(t, testUser{
		telegramID: 1, firstName: "Alex", gender: "male", genderPref: "everyone",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", faith: "christian", faithImportance: "somewhat",
		verification: "unverified", trustScore: 50,
	})

	createTestUser(t, testUser{
		telegramID: 2, firstName: "Ada", gender: "female", genderPref: "male",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", faith: "christian", faithImportance: "somewhat",
		verification: "unverified", trustScore: 50,
	})

	createTestUser(t, testUser{
		telegramID: 3, firstName: "Emeka", gender: "male", genderPref: "everyone",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", faith: "christian", faithImportance: "somewhat",
		verification: "unverified", trustScore: 50,
	})

	// gender_pref = everyone → should see both genders (that also accept male)
	filters := defaultFilters("male", "everyone", "dating", "christian", "somewhat", "born_in_africa", []string{"nigerian"}, 6.52, 3.37, "NG")
	candidates, err := testService.GetExploreFeed(testCtx, userID, filters, 0)
	require.NoError(t, err)

	assert.Len(t, candidates, 2)
}

// ============================================================
// TESTS: Intention Matching
// ============================================================

func TestMatching_IntentionDating(t *testing.T) {
	cleanDB(t)

	userID := createTestUser(t, testUser{
		telegramID: 1, firstName: "Tolu", gender: "male", genderPref: "female",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "unverified", trustScore: 50,
	})

	createTestUser(t, testUser{
		telegramID: 2, firstName: "Ada", gender: "female", genderPref: "male",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "unverified", trustScore: 50,
	})

	createTestUser(t, testUser{
		telegramID: 3, firstName: "Bola", gender: "female", genderPref: "male",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "friendship", verification: "unverified", trustScore: 50,
	})

	createTestUser(t, testUser{
		telegramID: 4, firstName: "Chi", gender: "female", genderPref: "male",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "both", verification: "unverified", trustScore: 50,
	})

	filters := defaultFilters("male", "female", "dating", "", "", "born_in_africa", []string{"nigerian"}, 6.52, 3.37, "NG")
	candidates, err := testService.GetExploreFeed(testCtx, userID, filters, 0)
	require.NoError(t, err)

	// Should see Ada (dating) and Chi (both), NOT Bola (friendship)
	names := candidateNames(candidates)
	assert.Contains(t, names, "Ada")
	assert.Contains(t, names, "Chi")
	assert.NotContains(t, names, "Bola")
}

func TestMatching_IntentionBothSeesEveryone(t *testing.T) {
	cleanDB(t)

	userID := createTestUser(t, testUser{
		telegramID: 1, firstName: "Tolu", gender: "male", genderPref: "female",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "both", verification: "unverified", trustScore: 50,
	})

	createTestUser(t, testUser{
		telegramID: 2, firstName: "Ada", gender: "female", genderPref: "male",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "unverified", trustScore: 50,
	})

	createTestUser(t, testUser{
		telegramID: 3, firstName: "Bola", gender: "female", genderPref: "male",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "friendship", verification: "unverified", trustScore: 50,
	})

	createTestUser(t, testUser{
		telegramID: 4, firstName: "Chi", gender: "female", genderPref: "male",
		city: "Accra", country: "GH", lat: 5.60, lon: -0.19,
		heritage: []string{"ghanaian"}, diasporaTag: "born_in_africa",
		intention: "both", verification: "unverified", trustScore: 50,
	})

	// User with intention "both" should see ALL intentions
	filters := defaultFilters("male", "female", "both", "", "", "born_in_africa", []string{"nigerian"}, 6.52, 3.37, "NG")
	candidates, err := testService.GetExploreFeed(testCtx, userID, filters, 0)
	require.NoError(t, err)

	names := candidateNames(candidates)
	assert.Len(t, candidates, 3)
	assert.Contains(t, names, "Ada")
	assert.Contains(t, names, "Bola")
	assert.Contains(t, names, "Chi")
}

// ============================================================
// TESTS: Seen / Blocked / Self Exclusion
// ============================================================

func TestMatching_ExcludesSelf(t *testing.T) {
	cleanDB(t)

	userID := createTestUser(t, testUser{
		telegramID: 1, firstName: "Tolu", gender: "male", genderPref: "everyone",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "both", verification: "unverified", trustScore: 50,
	})

	filters := defaultFilters("male", "everyone", "both", "", "", "born_in_africa", []string{"nigerian"}, 6.52, 3.37, "NG")
	candidates, err := testService.GetExploreFeed(testCtx, userID, filters, 0)
	require.NoError(t, err)

	assert.Len(t, candidates, 0)
}

func TestMatching_ExcludesSeenProfiles(t *testing.T) {
	cleanDB(t)

	userID := createTestUser(t, testUser{
		telegramID: 1, firstName: "Tolu", gender: "male", genderPref: "female",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "unverified", trustScore: 50,
	})

	adaID := createTestUser(t, testUser{
		telegramID: 2, firstName: "Ada", gender: "female", genderPref: "male",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "unverified", trustScore: 50,
	})

	createTestUser(t, testUser{
		telegramID: 3, firstName: "Bola", gender: "female", genderPref: "male",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "unverified", trustScore: 50,
	})

	// Mark Ada as seen
	testRepo.MarkSeen(testCtx, userID, adaID)

	filters := defaultFilters("male", "female", "dating", "", "", "born_in_africa", []string{"nigerian"}, 6.52, 3.37, "NG")
	candidates, err := testService.GetExploreFeed(testCtx, userID, filters, 0)
	require.NoError(t, err)

	// Should only see Bola, not Ada
	assert.Len(t, candidates, 1)
	assert.Equal(t, "Bola", candidates[0].FirstName)
}

func TestMatching_ExcludesBlockedUsers(t *testing.T) {
	cleanDB(t)

	userID := createTestUser(t, testUser{
		telegramID: 1, firstName: "Tolu", gender: "male", genderPref: "female",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "unverified", trustScore: 50,
	})

	adaID := createTestUser(t, testUser{
		telegramID: 2, firstName: "Ada", gender: "female", genderPref: "male",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "unverified", trustScore: 50,
	})

	// Block Ada
	testDB.Exec(testCtx, "INSERT INTO blocklist (user_id, blocked_user_id) VALUES ($1, $2)", userID, adaID)

	filters := defaultFilters("male", "female", "dating", "", "", "born_in_africa", []string{"nigerian"}, 6.52, 3.37, "NG")
	candidates, err := testService.GetExploreFeed(testCtx, userID, filters, 0)
	require.NoError(t, err)

	assert.Len(t, candidates, 0)
}

func TestMatching_ExcludesInactiveUsers(t *testing.T) {
	cleanDB(t)

	userID := createTestUser(t, testUser{
		telegramID: 1, firstName: "Tolu", gender: "male", genderPref: "female",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "unverified", trustScore: 50,
	})

	bannedID := createTestUser(t, testUser{
		telegramID: 2, firstName: "Banned", gender: "female", genderPref: "male",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "unverified", trustScore: 50,
	})
	testDB.Exec(testCtx, "UPDATE users SET status = 'banned' WHERE id = $1", bannedID)

	filters := defaultFilters("male", "female", "dating", "", "", "born_in_africa", []string{"nigerian"}, 6.52, 3.37, "NG")
	candidates, err := testService.GetExploreFeed(testCtx, userID, filters, 0)
	require.NoError(t, err)

	assert.Len(t, candidates, 0)
}

func TestMatching_ExcludesLowCompleteness(t *testing.T) {
	cleanDB(t)

	userID := createTestUser(t, testUser{
		telegramID: 1, firstName: "Tolu", gender: "male", genderPref: "female",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "unverified", trustScore: 50,
	})

	lowID := createTestUser(t, testUser{
		telegramID: 2, firstName: "Low", gender: "female", genderPref: "male",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "unverified", trustScore: 50,
	})
	testDB.Exec(testCtx, "UPDATE profiles SET completeness = 20 WHERE user_id = $1", lowID)

	filters := defaultFilters("male", "female", "dating", "", "", "born_in_africa", []string{"nigerian"}, 6.52, 3.37, "NG")
	candidates, err := testService.GetExploreFeed(testCtx, userID, filters, 0)
	require.NoError(t, err)

	assert.Len(t, candidates, 0)
}

// ============================================================
// TESTS: Scoring
// ============================================================

func TestMatching_HeritageScoring(t *testing.T) {
	cleanDB(t)

	userID := createTestUser(t, testUser{
		telegramID: 1, firstName: "Tolu", gender: "male", genderPref: "female",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "unverified", trustScore: 50,
	})

	// Same heritage → highest score
	createTestUser(t, testUser{
		telegramID: 2, firstName: "Ada", gender: "female", genderPref: "male",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "unverified", trustScore: 50,
	})

	// Same country, different heritage
	createTestUser(t, testUser{
		telegramID: 3, firstName: "Bola", gender: "female", genderPref: "male",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"ghanaian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "unverified", trustScore: 50,
	})

	// Different country entirely
	createTestUser(t, testUser{
		telegramID: 4, firstName: "Fatima", gender: "female", genderPref: "male",
		city: "New York", country: "US", lat: 40.71, lon: -74.00,
		heritage: []string{"somali"}, diasporaTag: "diaspora_2nd",
		intention: "dating", verification: "unverified", trustScore: 50,
	})

	filters := defaultFilters("male", "female", "dating", "", "", "born_in_africa", []string{"nigerian"}, 6.52, 3.37, "NG")
	candidates, err := testService.GetExploreFeed(testCtx, userID, filters, 0)
	require.NoError(t, err)

	assert.Len(t, candidates, 3)
	// Ada should score highest (same heritage + same location)
	assert.Equal(t, "Ada", candidates[0].FirstName)
}

func TestMatching_VerifiedUsersScoreHigher(t *testing.T) {
	cleanDB(t)

	userID := createTestUser(t, testUser{
		telegramID: 1, firstName: "Tolu", gender: "male", genderPref: "female",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "unverified", trustScore: 50,
	})

	createTestUser(t, testUser{
		telegramID: 2, firstName: "Verified", gender: "female", genderPref: "male",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "photo_verified", trustScore: 50,
	})

	createTestUser(t, testUser{
		telegramID: 3, firstName: "Unverified", gender: "female", genderPref: "male",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "unverified", trustScore: 50,
	})

	filters := defaultFilters("male", "female", "dating", "", "", "born_in_africa", []string{"nigerian"}, 6.52, 3.37, "NG")
	candidates, err := testService.GetExploreFeed(testCtx, userID, filters, 0)
	require.NoError(t, err)

	assert.Len(t, candidates, 2)
	// Verified should rank higher (4 extra points for verification)
	assert.Equal(t, "Verified", candidates[0].FirstName)
}

// ============================================================
// TESTS: Swipe + Match
// ============================================================

func TestSwipe_MutualLikeCreatesMatch(t *testing.T) {
	cleanDB(t)

	toluID := createTestUser(t, testUser{
		telegramID: 1, firstName: "Tolu", gender: "male", genderPref: "female",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "unverified", trustScore: 50,
	})

	adaID := createTestUser(t, testUser{
		telegramID: 2, firstName: "Ada", gender: "female", genderPref: "male",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "unverified", trustScore: 50,
	})

	// Tolu likes Ada → no match yet
	match1, err := testService.Swipe(testCtx, toluID, matching.SwipeRequest{
		CandidateID: adaID, Action: "like",
	})
	require.NoError(t, err)
	assert.Nil(t, match1)

	// Ada likes Tolu → MATCH!
	match2, err := testService.Swipe(testCtx, adaID, matching.SwipeRequest{
		CandidateID: toluID, Action: "like",
	})
	require.NoError(t, err)
	assert.NotNil(t, match2)
	assert.Equal(t, "active", match2.Status)
}

func TestSwipe_PassDoesNotMatch(t *testing.T) {
	cleanDB(t)

	toluID := createTestUser(t, testUser{
		telegramID: 1, firstName: "Tolu", gender: "male", genderPref: "female",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "unverified", trustScore: 50,
	})

	adaID := createTestUser(t, testUser{
		telegramID: 2, firstName: "Ada", gender: "female", genderPref: "male",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "unverified", trustScore: 50,
	})

	// Tolu likes Ada
	testService.Swipe(testCtx, toluID, matching.SwipeRequest{CandidateID: adaID, Action: "like"})

	// Ada passes on Tolu → no match
	match, err := testService.Swipe(testCtx, adaID, matching.SwipeRequest{CandidateID: toluID, Action: "pass"})
	require.NoError(t, err)
	assert.Nil(t, match)
}

func TestSwipe_OneLikeNoMatch(t *testing.T) {
	cleanDB(t)

	toluID := createTestUser(t, testUser{
		telegramID: 1, firstName: "Tolu", gender: "male", genderPref: "female",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "unverified", trustScore: 50,
	})

	adaID := createTestUser(t, testUser{
		telegramID: 2, firstName: "Ada", gender: "female", genderPref: "male",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "unverified", trustScore: 50,
	})

	// Only Tolu likes Ada → no match
	match, err := testService.Swipe(testCtx, toluID, matching.SwipeRequest{CandidateID: adaID, Action: "like"})
	require.NoError(t, err)
	assert.Nil(t, match)

	// Verify no match exists in DB
	matches, err := testService.GetMatches(testCtx, toluID)
	require.NoError(t, err)
	assert.Len(t, matches, 0)
}

// ============================================================
// TESTS: Cross-Continental Matching
// ============================================================

func TestMatching_CrossContinental(t *testing.T) {
	cleanDB(t)

	// Nigerian in Lagos
	lagosID := createTestUser(t, testUser{
		telegramID: 1, firstName: "Tolu", gender: "male", genderPref: "female",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "unverified", trustScore: 50,
	})

	// Nigerian in London (diaspora)
	createTestUser(t, testUser{
		telegramID: 2, firstName: "London", gender: "female", genderPref: "male",
		city: "London", country: "GB", lat: 51.51, lon: -0.13,
		heritage: []string{"nigerian"}, diasporaTag: "diaspora_1st",
		intention: "dating", verification: "unverified", trustScore: 50,
	})

	// Ghanaian in Accra (different heritage)
	createTestUser(t, testUser{
		telegramID: 3, firstName: "Accra", gender: "female", genderPref: "male",
		city: "Accra", country: "GH", lat: 5.60, lon: -0.19,
		heritage: []string{"ghanaian"}, diasporaTag: "born_in_africa",
		intention: "dating", verification: "unverified", trustScore: 50,
	})

	filters := defaultFilters("male", "female", "dating", "", "", "born_in_africa", []string{"nigerian"}, 6.52, 3.37, "NG")
	candidates, err := testService.GetExploreFeed(testCtx, lagosID, filters, 0)
	require.NoError(t, err)

	// Should see both — London (same heritage, different continent) and Accra (different heritage)
	assert.Len(t, candidates, 2)

	// London should rank higher (shared Nigerian heritage)
	assert.Equal(t, "London", candidates[0].FirstName)
}

// ============================================================
// TESTS: Faith Matching
// ============================================================

func TestMatching_FaithBothImportantSameReligion(t *testing.T) {
	cleanDB(t)

	userID := createTestUser(t, testUser{
		telegramID: 1, firstName: "Tolu", gender: "male", genderPref: "female",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", faith: "christian", faithImportance: "very_important",
		verification: "unverified", trustScore: 50,
	})

	// Same faith, both very important → high score
	createTestUser(t, testUser{
		telegramID: 2, firstName: "SameFaith", gender: "female", genderPref: "male",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", faith: "christian", faithImportance: "very_important",
		verification: "unverified", trustScore: 50,
	})

	// Different faith, both very important → low score
	createTestUser(t, testUser{
		telegramID: 3, firstName: "DiffFaith", gender: "female", genderPref: "male",
		city: "Lagos", country: "NG", lat: 6.52, lon: 3.37,
		heritage: []string{"nigerian"}, diasporaTag: "born_in_africa",
		intention: "dating", faith: "muslim", faithImportance: "very_important",
		verification: "unverified", trustScore: 50,
	})

	filters := defaultFilters("male", "female", "dating", "christian", "very_important", "born_in_africa", []string{"nigerian"}, 6.52, 3.37, "NG")
	candidates, err := testService.GetExploreFeed(testCtx, userID, filters, 0)
	require.NoError(t, err)

	assert.Len(t, candidates, 2)
	// Same faith should score higher
	assert.Equal(t, "SameFaith", candidates[0].FirstName)
}

// ============================================================
// HELPERS
// ============================================================

func candidateNames(candidates []matching.ScoredCandidate) []string {
	names := make([]string, len(candidates))
	for i, c := range candidates {
		names[i] = c.FirstName
	}
	return names
}
