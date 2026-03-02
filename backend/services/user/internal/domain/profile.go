package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	RoleClient     = "client"
	RoleFreelancer = "freelancer"
	RoleAdmin      = "admin"
)

type Profile struct {
	ID           int64
	UserID       uuid.UUID
	Role         string
	FirstName    string
	LastName     string
	DisplayName  string
	AvatarURL    string
	Language     string
	ContactEmail string
	ContactPhone string
	Bio          string
	DeletedAt    *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
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
