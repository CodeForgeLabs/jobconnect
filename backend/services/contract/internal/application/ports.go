package application

import (
	"context"
	"time"

	"jobconnect/contract/internal/domain"

	"github.com/google/uuid"
)

type ContractRepository interface {
	Create(ctx context.Context, c domain.Contract) (int64, error)
	GetByID(ctx context.Context, contractID int64) (domain.Contract, error)
	GetByIDForActor(ctx context.Context, contractID int64, actorID uuid.UUID) (domain.Contract, error)
	GetByProposalID(ctx context.Context, proposalID int64) (domain.Contract, error)
	GetJobOfferState(ctx context.Context, jobID int64, clientID uuid.UUID) (domain.JobOfferState, error)
	ListByActor(ctx context.Context, actorID uuid.UUID, status string, limit, offset int) ([]domain.Contract, error)
	UpdateOfferForClient(ctx context.Context, c domain.Contract) error
	SetStatusForFreelancer(ctx context.Context, contractID int64, freelancerID uuid.UUID, status string, at time.Time) error
	SetStatusForClient(ctx context.Context, contractID int64, clientID uuid.UUID, status string, at time.Time) error
	ReplaceMilestonesForActor(ctx context.Context, contractID int64, actorID uuid.UUID, milestones []domain.Milestone, at time.Time) error
	CreateHourlyLogForFreelancer(ctx context.Context, log domain.HourlyLog) (int64, error)
	ListHourlyLogsForActor(ctx context.Context, contractID int64, actorID uuid.UUID, limit, offset int) ([]domain.HourlyLog, error)
	ReviewHourlyLogForClient(ctx context.Context, hourlyLogID int64, clientID uuid.UUID, status string, note string, at time.Time) error
	GetHourlyLogForActor(ctx context.Context, hourlyLogID int64, actorID uuid.UUID) (domain.HourlyLog, error)
	CreateAmendmentForActor(ctx context.Context, a domain.Amendment) (int64, error)
	RespondAmendmentForActor(ctx context.Context, amendmentID int64, actorID uuid.UUID, status string, at time.Time) error
	GetAmendmentForActor(ctx context.Context, amendmentID int64, actorID uuid.UUID) (domain.Amendment, error)
	ListAmendmentsForActor(ctx context.Context, contractID int64, actorID uuid.UUID, limit, offset int) ([]domain.Amendment, error)
	AppendStatusHistory(ctx context.Context, entry domain.StatusHistoryEntry) error
	ListStatusHistoryForActor(ctx context.Context, contractID int64, actorID uuid.UUID, limit, offset int) ([]domain.StatusHistoryEntry, error)
}

type ProposalSummary struct {
	ID           int64
	JobID        int64
	ClientID     string
	FreelancerID string
	Status       string
}

type ProposalSync interface {
	GetProposal(ctx context.Context, proposalID int64, clientID uuid.UUID) (ProposalSummary, error)
	SetHired(ctx context.Context, proposalID int64, clientID uuid.UUID, reason string) error
	ReleaseHired(ctx context.Context, proposalID int64, clientID uuid.UUID, reason string) error
}

type JobStatusSync interface {
	SetInProgress(ctx context.Context, jobID int64, clientID uuid.UUID) error
}

type ActorPolicy interface {
	EnsureClientCanHire(ctx context.Context, userID uuid.UUID) error
	EnsureFreelancerCanWork(ctx context.Context, userID uuid.UUID) error
}

type Clock interface {
	Now() time.Time
}
