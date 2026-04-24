package user_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rooted-dating/rooted-server/internal/shared/config"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/rooted-dating/rooted-server/internal/shared/database"
	"github.com/rooted-dating/rooted-server/internal/user"
)

var (
	testDB      *pgxpool.Pool
	testRepo    *user.PostgresRepo
	testService *user.Service
	testCtx     = context.Background()
)

func TestMain(m *testing.M) {
	ctx := context.Background()

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
		panic(fmt.Sprintf("Failed to start postgres: %v", err))
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(fmt.Sprintf("Failed to get connection string: %v", err))
	}

	testDB, err = database.NewPostgres(ctx, connStr)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect: %v", err))
	}

	if err := runMigrations(ctx, testDB); err != nil {
		panic(fmt.Sprintf("Failed to run migrations: %v", err))
	}

	testRepo = user.NewPostgresRepo(testDB)
	rdb := database.NewSafeRedis(nil)
	dynConfig := config.NewDynamicConfig(testDB, rdb)
	testService = user.NewService(testRepo, rdb, dynConfig)

	code := m.Run()

	testDB.Close()
	pgContainer.Terminate(ctx)
	os.Exit(code)
}

func runMigrations(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
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
			sub_expires_at TIMESTAMPTZ,
			terms_accepted_at TIMESTAMPTZ,
			privacy_accepted_at TIMESTAMPTZ
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

		CREATE TABLE blocklist (
			user_id UUID NOT NULL REFERENCES users(id),
			blocked_user_id UUID NOT NULL REFERENCES users(id),
			created_at TIMESTAMPTZ DEFAULT NOW(),
			PRIMARY KEY (user_id, blocked_user_id)
		);
	`)
	return err
}

func cleanDB(t *testing.T) {
	t.Helper()
	testDB.Exec(testCtx, "DELETE FROM photos")
	testDB.Exec(testCtx, "DELETE FROM blocklist")
	testDB.Exec(testCtx, "DELETE FROM profiles")
	testDB.Exec(testCtx, "DELETE FROM users")
}

// ============================================================
// USER CREATION
// ============================================================

func TestFindOrCreateUser_NewUser(t *testing.T) {
	cleanDB(t)

	u, isNew, err := testService.FindOrCreateUser(testCtx, 12345, "testuser")
	require.NoError(t, err)
	assert.True(t, isNew)
	assert.NotEmpty(t, u.ID)
	assert.Equal(t, int64(12345), u.TelegramID)
	assert.Equal(t, "active", u.Status)
	assert.Equal(t, "unverified", u.Verification)
	assert.Equal(t, "free", u.Subscription)
	assert.Equal(t, 50, u.TrustScore)
}

func TestFindOrCreateUser_ExistingUser(t *testing.T) {
	cleanDB(t)

	u1, isNew1, _ := testService.FindOrCreateUser(testCtx, 12345, "testuser")
	assert.True(t, isNew1)

	u2, isNew2, err := testService.FindOrCreateUser(testCtx, 12345, "testuser")
	require.NoError(t, err)
	assert.False(t, isNew2)
	assert.Equal(t, u1.ID, u2.ID)
}

// ============================================================
// PROFILE CRUD
// ============================================================

func TestCreateProfile(t *testing.T) {
	cleanDB(t)

	u, _, _ := testService.FindOrCreateUser(testCtx, 100, "tolu")

	req := user.CreateProfileRequest{
		FirstName:   "Tolu",
		DateOfBirth: "1994-04-23",
		Gender:      "male",
		GenderPref:  "female",
		City:        "Lagos",
		Country:     "NG",
		Latitude:    6.52,
		Longitude:   3.37,
		Heritage:    []string{"nigerian"},
		DiasporaTag: "born_in_africa",
		Intention:   "dating",
		Faith:       "christian",
		FaithImportance: "very_important",
		Bio:         "Hello world",
		CulturalPrompts: []user.PromptResponse{
			{Prompt: "Home is...", Answer: "Lagos"},
			{Prompt: "I'm proudly...", Answer: "Nigerian"},
		},
		PersonalityPrompts: []user.PromptResponse{
			{Prompt: "My ideal weekend...", Answer: "Chill"},
		},
	}

	profile, err := testService.CreateProfile(testCtx, u.ID, req)
	require.NoError(t, err)
	assert.Equal(t, "Tolu", profile.FirstName)
	assert.Equal(t, "male", profile.Gender)
	assert.Equal(t, "Lagos", profile.City)
	assert.Equal(t, []string{"nigerian"}, profile.Heritage)
	assert.Equal(t, "born_in_africa", profile.DiasporaTag)
	assert.Equal(t, 2, len(profile.CulturalPrompts))
	assert.Equal(t, 1, len(profile.PersonalityPrompts))
	assert.True(t, profile.Completeness > 0)
}

func TestCreateProfile_Upsert(t *testing.T) {
	cleanDB(t)

	u, _, _ := testService.FindOrCreateUser(testCtx, 100, "tolu")

	req1 := user.CreateProfileRequest{
		FirstName: "Tolu", DateOfBirth: "1994-04-23",
		Gender: "male", GenderPref: "female",
		Heritage: []string{"nigerian"}, DiasporaTag: "born_in_africa",
		Intention: "dating",
	}

	_, err := testService.CreateProfile(testCtx, u.ID, req1)
	require.NoError(t, err)

	// Create again with different name → should update, not error
	req2 := user.CreateProfileRequest{
		FirstName: "Toluwase", DateOfBirth: "1994-04-23",
		Gender: "male", GenderPref: "female",
		Heritage: []string{"nigerian"}, DiasporaTag: "born_in_africa",
		Intention: "both",
	}

	profile, err := testService.CreateProfile(testCtx, u.ID, req2)
	require.NoError(t, err)
	assert.Equal(t, "Toluwase", profile.FirstName)
	assert.Equal(t, "both", profile.Intention)
}

func TestUpdateProfile(t *testing.T) {
	cleanDB(t)

	u, _, _ := testService.FindOrCreateUser(testCtx, 100, "tolu")
	testService.CreateProfile(testCtx, u.ID, user.CreateProfileRequest{
		FirstName: "Tolu", DateOfBirth: "1994-04-23",
		Gender: "male", GenderPref: "female",
		Heritage: []string{"nigerian"}, DiasporaTag: "born_in_africa",
		Intention: "dating", City: "Lagos", Country: "NG",
	})

	newCity := "London"
	newCountry := "GB"
	err := testService.UpdateProfile(testCtx, u.ID, user.UpdateProfileRequest{
		City:    &newCity,
		Country: &newCountry,
	})
	require.NoError(t, err)

	profile, err := testService.GetProfile(testCtx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, "London", profile.City)
	assert.Equal(t, "GB", profile.Country)
	assert.Equal(t, "Tolu", profile.FirstName) // unchanged
}

func TestGetProfile_NotFound(t *testing.T) {
	cleanDB(t)

	profile, err := testService.GetProfile(testCtx, "aaaaaaaa-1111-1111-1111-111111111111")
	require.NoError(t, err)
	assert.Nil(t, profile)
}

// ============================================================
// PHOTOS
// ============================================================

func TestAddPhoto(t *testing.T) {
	cleanDB(t)

	u, _, _ := testService.FindOrCreateUser(testCtx, 100, "tolu")
	testService.CreateProfile(testCtx, u.ID, user.CreateProfileRequest{
		FirstName: "Tolu", DateOfBirth: "1994-04-23",
		Gender: "male", GenderPref: "female",
		Heritage: []string{"nigerian"}, DiasporaTag: "born_in_africa",
		Intention: "dating",
	})

	photo := user.Photo{
		ID: uuid.New().String(), UserID: u.ID,
		URLThumbnail: "key/thumb.jpg", URLMedium: "key/med.jpg", URLLarge: "key/large.jpg",
		Position: 0, IsPrimary: true,
	}
	err := testService.AddPhoto(testCtx, photo)
	require.NoError(t, err)

	profile, _ := testService.GetProfile(testCtx, u.ID)
	assert.Len(t, profile.Photos, 1)
	assert.Equal(t, photo.ID, profile.Photos[0].ID)
	assert.True(t, profile.Photos[0].IsPrimary)
}

func TestDeletePhoto(t *testing.T) {
	cleanDB(t)

	u, _, _ := testService.FindOrCreateUser(testCtx, 100, "tolu")
	testService.CreateProfile(testCtx, u.ID, user.CreateProfileRequest{
		FirstName: "Tolu", DateOfBirth: "1994-04-23",
		Gender: "male", GenderPref: "female",
		Heritage: []string{"nigerian"}, DiasporaTag: "born_in_africa",
		Intention: "dating",
	})

	photo1ID := uuid.New().String()
	photo2ID := uuid.New().String()

	testService.AddPhoto(testCtx, user.Photo{
		ID: photo1ID, UserID: u.ID,
		URLThumbnail: "k1", URLMedium: "k1", URLLarge: "k1",
		Position: 0, IsPrimary: true,
	})
	testService.AddPhoto(testCtx, user.Photo{
		ID: photo2ID, UserID: u.ID,
		URLThumbnail: "k2", URLMedium: "k2", URLLarge: "k2",
		Position: 1, IsPrimary: false,
	})

	err := testService.DeletePhoto(testCtx, photo1ID, u.ID)
	require.NoError(t, err)

	profile, _ := testService.GetProfile(testCtx, u.ID)
	assert.Len(t, profile.Photos, 1)
	assert.Equal(t, photo2ID, profile.Photos[0].ID)
}

func TestDeletePhoto_WrongUser(t *testing.T) {
	cleanDB(t)

	u1, _, _ := testService.FindOrCreateUser(testCtx, 100, "user1")
	u2, _, _ := testService.FindOrCreateUser(testCtx, 200, "user2")
	testService.CreateProfile(testCtx, u1.ID, user.CreateProfileRequest{
		FirstName: "User1", DateOfBirth: "1994-01-01",
		Gender: "male", GenderPref: "female",
		Heritage: []string{"nigerian"}, DiasporaTag: "born_in_africa",
		Intention: "dating",
	})

	photoID := uuid.New().String()
	testService.AddPhoto(testCtx, user.Photo{
		ID: photoID, UserID: u1.ID,
		URLThumbnail: "k1", URLMedium: "k1", URLLarge: "k1",
		Position: 0, IsPrimary: true,
	})

	// User2 tries to delete User1's photo — should not delete
	testService.DeletePhoto(testCtx, photoID, u2.ID)

	profile, _ := testService.GetProfile(testCtx, u1.ID)
	assert.Len(t, profile.Photos, 1) // still there
}

// ============================================================
// COMPLETENESS
// ============================================================

func TestCompleteness_MinimalProfile(t *testing.T) {
	cleanDB(t)

	u, _, _ := testService.FindOrCreateUser(testCtx, 100, "tolu")
	testService.CreateProfile(testCtx, u.ID, user.CreateProfileRequest{
		FirstName: "Tolu", DateOfBirth: "1994-04-23",
		Gender: "male", GenderPref: "female",
		Heritage: []string{"nigerian"}, DiasporaTag: "born_in_africa",
		Intention: "dating",
	})

	profile, _ := testService.GetProfile(testCtx, u.ID)
	// Has: firstName, dob, gender, heritage, diasporaTag, intention = 6/14
	// Missing: city, faith, culturalPrompts, personalityPrompts, photos(1), photos(4+), bio, audioBio
	assert.True(t, profile.Completeness < 50, "minimal profile should be under 50%%")
}

func TestCompleteness_FullProfile(t *testing.T) {
	cleanDB(t)

	u, _, _ := testService.FindOrCreateUser(testCtx, 100, "tolu")
	testService.CreateProfile(testCtx, u.ID, user.CreateProfileRequest{
		FirstName: "Tolu", DateOfBirth: "1994-04-23",
		Gender: "male", GenderPref: "female",
		City: "Lagos", Country: "NG",
		Heritage: []string{"nigerian"}, DiasporaTag: "born_in_africa",
		Intention: "dating", Faith: "christian", FaithImportance: "very_important",
		Bio: "Hello",
		CulturalPrompts: []user.PromptResponse{
			{Prompt: "A", Answer: "B"}, {Prompt: "C", Answer: "D"},
		},
		PersonalityPrompts: []user.PromptResponse{
			{Prompt: "E", Answer: "F"},
		},
	})

	// Add 4 photos
	for i := 0; i < 4; i++ {
		testService.AddPhoto(testCtx, user.Photo{
			ID: uuid.New().String(), UserID: u.ID,
			URLThumbnail: "k", URLMedium: "k", URLLarge: "k",
			Position: i, IsPrimary: i == 0,
		})
	}

	profile, _ := testService.GetProfile(testCtx, u.ID)
	// Has everything except audioBio = 13/14 = 92%
	assert.True(t, profile.Completeness >= 90, "full profile should be >= 90%%, got %d%%", profile.Completeness)
}

func TestCompleteness_PhotosUpdateIt(t *testing.T) {
	cleanDB(t)

	u, _, _ := testService.FindOrCreateUser(testCtx, 100, "tolu")
	testService.CreateProfile(testCtx, u.ID, user.CreateProfileRequest{
		FirstName: "Tolu", DateOfBirth: "1994-04-23",
		Gender: "male", GenderPref: "female",
		Heritage: []string{"nigerian"}, DiasporaTag: "born_in_africa",
		Intention: "dating",
	})

	before, _ := testService.GetProfile(testCtx, u.ID)
	beforeScore := before.Completeness

	testService.AddPhoto(testCtx, user.Photo{
		ID: uuid.New().String(), UserID: u.ID,
		URLThumbnail: "k", URLMedium: "k", URLLarge: "k",
		Position: 0, IsPrimary: true,
	})

	after, _ := testService.GetProfile(testCtx, u.ID)
	t.Logf("before=%d after=%d photos=%d", beforeScore, after.Completeness, len(after.Photos))
	assert.True(t, after.Completeness > beforeScore, "adding photo should increase completeness (before=%d, after=%d, photos=%d)", beforeScore, after.Completeness, len(after.Photos))
}

// ============================================================
// BLOCK / UNBLOCK
// ============================================================

func TestBlockUser(t *testing.T) {
	cleanDB(t)

	u1, _, _ := testService.FindOrCreateUser(testCtx, 100, "user1")
	u2, _, _ := testService.FindOrCreateUser(testCtx, 200, "user2")

	err := testService.BlockUser(testCtx, u1.ID, u2.ID)
	require.NoError(t, err)

	blocked, err := testRepo.IsBlocked(testCtx, u1.ID, u2.ID)
	require.NoError(t, err)
	assert.True(t, blocked)

	// Reverse check also returns true
	blocked2, _ := testRepo.IsBlocked(testCtx, u2.ID, u1.ID)
	assert.True(t, blocked2)
}

func TestUnblockUser(t *testing.T) {
	cleanDB(t)

	u1, _, _ := testService.FindOrCreateUser(testCtx, 100, "user1")
	u2, _, _ := testService.FindOrCreateUser(testCtx, 200, "user2")

	testService.BlockUser(testCtx, u1.ID, u2.ID)
	testService.UnblockUser(testCtx, u1.ID, u2.ID)

	blocked, _ := testRepo.IsBlocked(testCtx, u1.ID, u2.ID)
	assert.False(t, blocked)
}

func TestBlockUser_Duplicate(t *testing.T) {
	cleanDB(t)

	u1, _, _ := testService.FindOrCreateUser(testCtx, 100, "user1")
	u2, _, _ := testService.FindOrCreateUser(testCtx, 200, "user2")

	testService.BlockUser(testCtx, u1.ID, u2.ID)
	err := testService.BlockUser(testCtx, u1.ID, u2.ID) // duplicate — should not error
	assert.NoError(t, err)
}

// ============================================================
// PAUSE / RESUME / DELETE
// ============================================================

func TestPauseResume(t *testing.T) {
	cleanDB(t)

	u, _, _ := testService.FindOrCreateUser(testCtx, 100, "tolu")

	testService.PauseProfile(testCtx, u.ID)
	paused, _ := testService.GetUser(testCtx, u.ID)
	assert.Equal(t, "paused", paused.Status)

	testService.ResumeProfile(testCtx, u.ID)
	resumed, _ := testService.GetUser(testCtx, u.ID)
	assert.Equal(t, "active", resumed.Status)
}

func TestDeleteAccount(t *testing.T) {
	cleanDB(t)

	u, _, _ := testService.FindOrCreateUser(testCtx, 100, "tolu")
	testService.CreateProfile(testCtx, u.ID, user.CreateProfileRequest{
		FirstName: "Tolu", DateOfBirth: "1994-04-23",
		Gender: "male", GenderPref: "female",
		Heritage: []string{"nigerian"}, DiasporaTag: "born_in_africa",
		Intention: "dating",
	})

	testService.DeleteAccount(testCtx, u.ID)

	deleted, _ := testService.GetUser(testCtx, u.ID)
	assert.Equal(t, "deleted", deleted.Status)
}

func TestVerifyPhoto(t *testing.T) {
	cleanDB(t)

	u, _, _ := testService.FindOrCreateUser(testCtx, 100, "tolu")
	assert.Equal(t, "unverified", u.Verification)

	testService.VerifyPhoto(testCtx, u.ID)

	verified, _ := testService.GetUser(testCtx, u.ID)
	assert.Equal(t, "photo_verified", verified.Verification)
}
