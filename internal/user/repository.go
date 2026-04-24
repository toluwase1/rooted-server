package user

import "context"

// Repository defines the data access interface for the user domain.
// This is the escape hatch — swap implementations without changing business logic.
// Today: postgres_repo.go implements this against PostgreSQL.
// Later: could be DynamoDB, gRPC call to a user microservice, etc.
type Repository interface {
	// Users
	CreateUser(ctx context.Context, telegramID int64, telegramUsername string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByTelegramID(ctx context.Context, telegramID int64) (*User, error)
	UpdateUserStatus(ctx context.Context, id string, status string) error
	UpdateUserVerification(ctx context.Context, id string, verification string) error
	UpdateUserSubscription(ctx context.Context, id string, subscription string, expiresAt *string) error
	UpdateLastActive(ctx context.Context, id string) error
	AcceptTerms(ctx context.Context, id string) error
	DeleteUser(ctx context.Context, id string) error

	// Profiles
	CreateProfile(ctx context.Context, userID string, req CreateProfileRequest) (*Profile, error)
	GetProfile(ctx context.Context, userID string) (*Profile, error)
	UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) error
	UpdateCompleteness(ctx context.Context, userID string, completeness int) error

	// Photos
	AddPhoto(ctx context.Context, photo Photo) error
	GetPhotos(ctx context.Context, userID string) ([]Photo, error)
	DeletePhoto(ctx context.Context, photoID string, userID string) error
	ReorderPhotos(ctx context.Context, userID string, photoIDs []string) error
	UpdatePhotoModeration(ctx context.Context, photoID string, status string) error

	// Block list
	BlockUser(ctx context.Context, blockerID, blockedID string) error
	UnblockUser(ctx context.Context, blockerID, blockedID string) error
	IsBlocked(ctx context.Context, userA, userB string) (bool, error)
	GetBlockList(ctx context.Context, userID string) ([]string, error)
}
