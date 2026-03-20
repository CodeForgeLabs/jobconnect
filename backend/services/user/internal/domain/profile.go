package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	RoleClient     = "client"
	RoleFreelancer = "freelancer"
	RoleAdmin      = "admin"

	AccountStatusActive    = "ACTIVE"
	AccountStatusSuspended = "SUSPENDED"
	AccountStatusDeleted   = "DELETED"

	ProfileVisibilityPublic  = "PUBLIC"
	ProfileVisibilityPrivate = "PRIVATE"

	VerificationStatusPending  = "PENDING"
	VerificationStatusVerified = "VERIFIED"
	VerificationStatusRejected = "REJECTED"
	VerificationStatusExpired  = "EXPIRED"

	AvailabilityFullTime    = "FULL_TIME"
	AvailabilityPartTime    = "PART_TIME"
	AvailabilityAsNeeded    = "AS_NEEDED"
	AvailabilityUnavailable = "UNAVAILABLE"
)

type Profile struct {
	ID               int64
	UserID           uuid.UUID
	Role             string
	FirstName        string
	LastName         string
	DisplayName      string
	AvatarURL        string
	Language         string
	ContactEmail     string
	ContactPhone     string
	Bio              string
	AccountStatus    string
	SuspensionReason string
	Visibility       string
	DeletedAt        *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type ClientProfile struct {
	CompanyName        string
	BillingAddress     string
	TaxID              string
	VerificationStatus string
}

type FreelancerProfile struct {
	Headline           string
	Bio                string
	Skills             []string
	ExperienceLevel    string
	Rating             float64
	VerificationStatus string
	Reputation         Reputation
	HourlyRate         float64
	Availability       string
	Location           string
	LastActiveAt       *time.Time
}

type Reputation struct {
	JobSuccessScore  float64
	AvgRating        float64
	TotalReviews     uint32
	TotalJobs        uint32
	TotalEarningsUSD float64
}

type Avatar struct {
	UserID      uuid.UUID
	FileName    string
	ContentType string
	Content     []byte
	Width       int
	Height      int
	SizeBytes   int64
	UpdatedAt   time.Time
}
