package application

import (
	"context"
	"time"

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
	SortOrder   int32
}

type PortfolioItem struct {
	ID            int64
	UserID        uuid.UUID
	Title         string
	Description   string
	ProjectURL    string
	RoleInProject string
	CompletedAt   *time.Time
	SortOrder     int32
	Visibility    string
	Tags          []string
	Media         []PortfolioMedia
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Employment struct {
	ID             int64
	UserID         uuid.UUID
	CompanyName    string
	Title          string
	EmploymentType string
	Location       string
	IsCurrent      bool
	StartDate      *time.Time
	EndDate        *time.Time
	Description    string
	SortOrder      int32
	Visibility     string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Education struct {
	ID           int64
	UserID       uuid.UUID
	SchoolName   string
	Degree       string
	FieldOfStudy string
	IsCurrent    bool
	StartDate    *time.Time
	EndDate      *time.Time
	Grade        string
	Description  string
	SortOrder    int32
	Visibility   string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Certification struct {
	ID                  int64
	UserID              uuid.UUID
	Name                string
	IssuingOrganization string
	CredentialID        string
	CredentialURL       string
	IssueDate           *time.Time
	ExpirationDate      *time.Time
	DoesNotExpire       bool
	Visibility          string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type LanguageProficiency struct {
	LanguageCode string
	Proficiency  string
	Visibility   string
}

type AvailabilitySettings struct {
	Availability        string
	WeeklyCapacityHours uint32
}

type RateSettings struct {
	HourlyRate float64
	Currency   string
}

type WorkPreferences struct {
	PreferredProjectLength string
	MinBudgetUSD           float64
	MaxBudgetUSD           float64
	ContractTypes          []string
}

type ClientProfileSettings struct {
	CompanyName        string
	BillingAddress     string
	TaxID              string
	VerificationStatus string
}

type CompanySettings struct {
	CompanyName    string
	BillingAddress string
	TaxID          string
}

type HiringPreferences struct {
	MinHourlyRate             float64
	MaxHourlyRate             float64
	PreferredExperienceLevels []string
	PreferredLocations        []string
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

type ListUsersFilter struct {
	Q         string
	Role      string
	Status    string
	PageSize  uint32
	PageToken string
}

type UserSummary struct {
	UserID      uuid.UUID
	Role        string
	Status      string
	Visibility  string
	FirstName   string
	LastName    string
	DisplayName string
	AvatarURL   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ImpersonationToken struct {
	Token     string
	ExpiresAt time.Time
}

type UserAuditSummary struct {
	UserID                uuid.UUID
	Status                string
	Visibility            string
	ProfileUpdatedAt      time.Time
	AvatarUpdatedAt       *time.Time
	SavedFreelancersCount uint32
	PortfolioItemsCount   uint32
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
	ListPublicPortfolioItems(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string) (ListResult[PortfolioItem], error)
	ReorderPortfolioItems(ctx context.Context, userID uuid.UUID, itemIDs []int64) ([]PortfolioItem, error)

	GetEmployment(ctx context.Context, userID uuid.UUID, employmentID int64) (Employment, error)
	CreateEmployment(ctx context.Context, userID uuid.UUID, in Employment) (Employment, error)
	UpdateEmployment(ctx context.Context, userID uuid.UUID, employmentID int64, in Employment) (Employment, error)
	DeleteEmployment(ctx context.Context, userID uuid.UUID, employmentID int64) (bool, error)
	ListMyEmployment(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string) (ListResult[Employment], error)
	ListPublicEmployment(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string) (ListResult[Employment], error)

	GetEducation(ctx context.Context, userID uuid.UUID, educationID int64) (Education, error)
	CreateEducation(ctx context.Context, userID uuid.UUID, in Education) (Education, error)
	UpdateEducation(ctx context.Context, userID uuid.UUID, educationID int64, in Education) (Education, error)
	DeleteEducation(ctx context.Context, userID uuid.UUID, educationID int64) (bool, error)
	ListMyEducation(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string) (ListResult[Education], error)
	ListPublicEducation(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string) (ListResult[Education], error)

	GetCertification(ctx context.Context, userID uuid.UUID, certificationID int64) (Certification, error)
	CreateCertification(ctx context.Context, userID uuid.UUID, in Certification) (Certification, error)
	UpdateCertification(ctx context.Context, userID uuid.UUID, certificationID int64, in Certification) (Certification, error)
	DeleteCertification(ctx context.Context, userID uuid.UUID, certificationID int64) (bool, error)
	ListMyCertifications(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string) (ListResult[Certification], error)
	ListPublicCertifications(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string) (ListResult[Certification], error)

	UpsertLanguages(ctx context.Context, userID uuid.UUID, languages []LanguageProficiency) ([]LanguageProficiency, error)
	GetMyLanguages(ctx context.Context, userID uuid.UUID) ([]LanguageProficiency, error)
	GetPublicLanguages(ctx context.Context, userID uuid.UUID) ([]LanguageProficiency, error)

	SetAvailability(ctx context.Context, userID uuid.UUID, in AvailabilitySettings) (AvailabilitySettings, error)
	GetAvailability(ctx context.Context, userID uuid.UUID) (AvailabilitySettings, error)
	SetRates(ctx context.Context, userID uuid.UUID, in RateSettings) (RateSettings, error)
	GetRates(ctx context.Context, userID uuid.UUID) (RateSettings, error)
	SetWorkPreferences(ctx context.Context, userID uuid.UUID, in WorkPreferences) (WorkPreferences, error)
	GetWorkPreferences(ctx context.Context, userID uuid.UUID) (WorkPreferences, error)

	GetClientProfile(ctx context.Context, userID uuid.UUID) (ClientProfileSettings, error)
	UpdateClientProfile(ctx context.Context, userID uuid.UUID, in ClientProfileSettings) (ClientProfileSettings, error)
	GetCompany(ctx context.Context, userID uuid.UUID) (CompanySettings, error)
	UpdateCompany(ctx context.Context, userID uuid.UUID, in CompanySettings) (CompanySettings, error)
	GetHiringPreferences(ctx context.Context, userID uuid.UUID) (HiringPreferences, error)
	UpdateHiringPreferences(ctx context.Context, userID uuid.UUID, in HiringPreferences) (HiringPreferences, error)
	SaveFreelancer(ctx context.Context, userID uuid.UUID, freelancerUserID uuid.UUID) (SavedFreelancer, error)
	ListSavedFreelancers(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string) (ListResult[SavedFreelancer], error)
	RemoveSavedFreelancer(ctx context.Context, userID uuid.UUID, freelancerUserID uuid.UUID) (bool, error)
	UpsertFreelancerNote(ctx context.Context, userID uuid.UUID, freelancerUserID uuid.UUID, note string) (FreelancerNote, error)
	GetFreelancerNote(ctx context.Context, userID uuid.UUID, freelancerUserID uuid.UUID) (FreelancerNote, error)

	ListUsers(ctx context.Context, requesterUserID uuid.UUID, filter ListUsersFilter) (ListResult[UserSummary], error)
	CreateImpersonationToken(ctx context.Context, requesterUserID uuid.UUID, targetUserID uuid.UUID, reason string, ttlSeconds uint32) (ImpersonationToken, error)
	GetUserAuditSummary(ctx context.Context, requesterUserID uuid.UUID, targetUserID uuid.UUID) (UserAuditSummary, error)
}
