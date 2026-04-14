package application

import (
	"context"
	"time"

	"jobconnect/proposal/internal/domain"

	"github.com/google/uuid"
)

type ProposalRepository interface {
	Create(ctx context.Context, p domain.Proposal) (int64, error)
	GetByID(ctx context.Context, proposalID int64) (domain.Proposal, error)
	GetByIDForFreelancer(ctx context.Context, proposalID int64, freelancerID uuid.UUID) (domain.Proposal, error)
	GetLatestByJobForFreelancer(ctx context.Context, jobID int64, freelancerID uuid.UUID) (domain.Proposal, error)
	GetByIDForClient(ctx context.Context, proposalID int64, clientID uuid.UUID) (domain.Proposal, error)
	HasActiveProposal(ctx context.Context, jobID int64, freelancerID uuid.UUID) (bool, error)

	UpdateEditable(ctx context.Context, proposalID int64, freelancerID uuid.UUID, coverLetter string, bidAmount float64, estimatedDays int32, attachments []domain.Attachment, updatedAt time.Time) error
	Withdraw(ctx context.Context, proposalID int64, freelancerID uuid.UUID, reason string, at time.Time) error
	SetStatus(ctx context.Context, proposalID int64, clientID uuid.UUID, status string, reason string, at time.Time) error
	HireWithRequestID(ctx context.Context, proposalID int64, clientID uuid.UUID, requestID string, reason string, at time.Time) (domain.Proposal, bool, error)
	HasHiredProposalForJob(ctx context.Context, jobID int64) (bool, error)

	ListByJob(ctx context.Context, filter ListByJobFilter, limit, offset int) ([]domain.Proposal, error)
	ListByFreelancer(ctx context.Context, filter ListByFreelancerFilter, limit, offset int) ([]domain.Proposal, error)
}

type JobReader interface {
	GetJobSummary(ctx context.Context, jobID int64) (JobSummary, error)
}

type ConnectsClient interface {
	DeductConnects(ctx context.Context, userID uuid.UUID, amount int32, referenceID string) error
}

type Clock interface {
	Now() time.Time
}

type JobSummary struct {
	JobID    int64
	ClientID uuid.UUID
	Status   string
	IsOpen   bool
	Found    bool
}

type ListByJobFilter struct {
	ClientID     uuid.UUID
	JobID        int64
	Statuses     []string
	FreelancerID *uuid.UUID
	SortBy       string
}

type ListByFreelancerFilter struct {
	FreelancerID uuid.UUID
	Statuses     []string
	JobID        *int64
	SortBy       string
}
