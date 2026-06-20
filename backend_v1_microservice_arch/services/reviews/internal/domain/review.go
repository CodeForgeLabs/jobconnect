package domain

import (
	"time"
)

// ReviewerRole type (safer than plain string)
type ReviewerRole string

const (
	RoleClient     ReviewerRole = "client"
	RoleFreelancer ReviewerRole = "freelancer"
)

// Review represents a row in reviews table
type Review struct {
	ID           int64        `db:"id"`
	ContractID   int64        `db:"contract_id"`
	ClientID     string       `db:"client_id"`
	FreelancerID string       `db:"freelancer_id"`
	ReviewerRole ReviewerRole `db:"reviewer_role"`
	Rating       int          `db:"rating"`
	Title        string       `db:"title"`
	Comment      string       `db:"comment"`
	CreatedAt    time.Time    `db:"created_at"`
	UpdatedAt    time.Time    `db:"updated_at"`
}
