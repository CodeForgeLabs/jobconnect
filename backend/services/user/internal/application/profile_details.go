package application

import (
	"context"
	"time"

	"jobconnect/user/internal/domain"

	"github.com/google/uuid"
)

type PortfolioMedia struct {
	ID          int64
	MediaType   string
	StorageKey  string
	ExternalURL string
	FileName    string
	ContentType string
	SizeBytes   int64
	Width       int32
	Height      int32
}

type PortfolioItem struct {
	ID            int64
	UserID        uuid.UUID
	Title         string
	Description   string
	ProjectURL    string
	RoleInProject string
	CompletedAt   *time.Time
	Tags          []string
	Media         []PortfolioMedia
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CV struct {
	UserID      uuid.UUID
	FileName    string
	ContentType string
	StorageKey  string
	SizeBytes   int64
	UpdatedAt   time.Time
}

type WorkPreferences struct {
	PreferredProjectLength string
	MinBudgetUSD           float64
	MaxBudgetUSD           float64
	ContractTypes          []string
	WeeklyCapacityHours    uint32
}

type HiringPreferences struct {
	MinHourlyRate      float64
	MaxHourlyRate      float64
	PreferredLocations []string
}

type SavedFreelancer struct {
	FreelancerUserID uuid.UUID
	SavedAt          time.Time
}

type FreelancerNote struct {
	FreelancerUserID uuid.UUID
	Note             string
	UpdatedAt        time.Time
}

type ListResult[T any] struct {
	Items         []T
	NextPageToken string
}

// ProfileDetailsRepository stores profile resource collections.
type ProfileDetailsRepository interface {
	GetPortfolioItem(ctx context.Context, userID uuid.UUID, itemID int64) (PortfolioItem, error)
	CreatePortfolioItem(ctx context.Context, userID uuid.UUID, in PortfolioItem) (PortfolioItem, error)
	UpdatePortfolioItem(ctx context.Context, userID uuid.UUID, itemID int64, in PortfolioItem) (PortfolioItem, error)
	DeletePortfolioItem(ctx context.Context, userID uuid.UUID, itemID int64) (bool, error)
	ListMyPortfolioItems(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string) (ListResult[PortfolioItem], error)
	SetWorkPreferences(ctx context.Context, userID uuid.UUID, in WorkPreferences) (WorkPreferences, error)
	GetWorkPreferences(ctx context.Context, userID uuid.UUID) (WorkPreferences, error)

	GetHiringPreferences(ctx context.Context, userID uuid.UUID) (HiringPreferences, error)
	UpdateHiringPreferences(ctx context.Context, userID uuid.UUID, in HiringPreferences) (HiringPreferences, error)
	SaveFreelancer(ctx context.Context, userID uuid.UUID, freelancerUserID uuid.UUID) (SavedFreelancer, error)
	ListSavedFreelancers(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string) (ListResult[SavedFreelancer], error)
	RemoveSavedFreelancer(ctx context.Context, userID uuid.UUID, freelancerUserID uuid.UUID) (bool, error)
	UpsertFreelancerNote(ctx context.Context, userID uuid.UUID, freelancerUserID uuid.UUID, note string) (FreelancerNote, error)
	GetFreelancerNote(ctx context.Context, userID uuid.UUID, freelancerUserID uuid.UUID) (FreelancerNote, error)
}

// CVObjectStore persists CV binary content outside the primary database.
type CVObjectStore interface {
	PutCV(ctx context.Context, cv domain.CVObject) error
	DeleteCV(ctx context.Context, userID uuid.UUID, storageKey string) error
	PresignGetObject(ctx context.Context, storageKey string, ttl time.Duration) (string, error)
	PresignPutObject(ctx context.Context, storageKey string, contentType string, ttl time.Duration) (string, error)
	StatObject(ctx context.Context, storageKey string) (ObjectInfo, error)
}
