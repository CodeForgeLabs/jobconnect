package application

import (
	"context"
	"time"

	"jobconnect/job/internal/domain"

	"github.com/google/uuid"
)

type JobRepository interface {
	Create(ctx context.Context, job domain.Job) (int64, error)
	GetByID(ctx context.Context, jobID int64) (domain.Job, error)
	GetByIDForClient(ctx context.Context, jobID int64, clientID uuid.UUID) (domain.Job, error)
	Update(ctx context.Context, job domain.Job) (domain.Job, error)
	AddAttachment(ctx context.Context, jobID int64, clientID uuid.UUID, attachment domain.Attachment) (domain.Attachment, error)
	DeleteAttachment(ctx context.Context, jobID int64, attachmentID int64, clientID uuid.UUID) (domain.Attachment, error)
	ListAttachments(ctx context.Context, jobID int64, clientID uuid.UUID) ([]domain.Attachment, error)
	GetAttachment(ctx context.Context, jobID int64, attachmentID int64, clientID uuid.UUID) (domain.Attachment, error)
	ListByClient(ctx context.Context, clientID uuid.UUID, status string, limit, offset int) ([]domain.Job, error)
	ListOpen(ctx context.Context, limit, offset int) ([]domain.Job, error)
	ListOpenFiltered(ctx context.Context, filter ListOpenFilter) ([]domain.Job, error)
	CountOpenFiltered(ctx context.Context, filter ListOpenFilter) (int64, error)
	SetVisibility(ctx context.Context, jobID int64, clientID uuid.UUID, visibility string, updatedAt time.Time) (domain.Job, error)
	SetBudgetRange(ctx context.Context, jobID int64, clientID uuid.UUID, budgetMin, budgetMax float64, updatedAt time.Time) (domain.Job, error)
	SetExperienceLevel(ctx context.Context, jobID int64, clientID uuid.UUID, experienceLevel string, updatedAt time.Time) (domain.Job, error)
	InviteFreelancer(ctx context.Context, jobID int64, clientID uuid.UUID, freelancerID string, createdAt time.Time) (bool, error)
	Pause(ctx context.Context, jobID int64, clientID uuid.UUID, updatedAt time.Time) (domain.Job, error)
	Reopen(ctx context.Context, jobID int64, clientID uuid.UUID, updatedAt time.Time) (domain.Job, error)
	MarkFilled(ctx context.Context, jobID int64, clientID uuid.UUID, updatedAt time.Time) (domain.Job, error)
	Close(ctx context.Context, jobID int64, clientID uuid.UUID, reason string, closedAt time.Time) error
}

type AttachmentObjectStore interface {
	BuildObjectKey(jobID int64, fileName string) string
	PutObject(ctx context.Context, objectKey string, content []byte, contentType string) (string, error)
	DeleteObject(ctx context.Context, objectKey string) error
}

// ListOpenFilter contains optional filters for the ListOpenJobs query.
type ListOpenFilter struct {
	SearchQuery     string
	Skills          []string
	JobType         string
	Visibility      string
	ExperienceLevel string
	Limit           int
	Offset          int
}

type ConnectsClient interface {
	RefundConnects(ctx context.Context, userID string, amount int32, referenceID string) error
}

type Proposal struct {
	ID            int64
	ClientID      string
	FreelancerID  string
	ConnectsSpent int32
	Status        string
}

type ProposalClient interface {
	ListProposalsByJob(ctx context.Context, jobID int64) ([]Proposal, error)
	GetProposal(ctx context.Context, proposalID int64) (Proposal, error)
	SetProposalStatus(ctx context.Context, proposalID int64, status string, reason string) error
}

type Clock interface {
	Now() time.Time
}
