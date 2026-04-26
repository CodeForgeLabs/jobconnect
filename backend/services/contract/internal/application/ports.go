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
	ReplaceMilestones(ctx context.Context, contractID int64, milestones []domain.Milestone, at time.Time) error
	CreateHourlyLogForFreelancer(ctx context.Context, log domain.HourlyLog) (int64, error)
	ListHourlyLogsForActor(ctx context.Context, contractID int64, actorID uuid.UUID, limit, offset int) ([]domain.HourlyLog, error)
	ListHourlyLogsForActorInRange(ctx context.Context, contractID int64, actorID uuid.UUID, startAt time.Time, endAt time.Time) ([]domain.HourlyLog, error)
	UpdateHourlyLogForFreelancer(ctx context.Context, log domain.HourlyLog) error
	DeleteHourlyLogForFreelancer(ctx context.Context, hourlyLogID int64, freelancerID uuid.UUID) error
	ReviewHourlyLogForClient(ctx context.Context, hourlyLogID int64, clientID uuid.UUID, status string, note string, at time.Time) error
	GetHourlyLogForActor(ctx context.Context, hourlyLogID int64, actorID uuid.UUID) (domain.HourlyLog, error)
	CreateHourlyInvoice(ctx context.Context, invoice domain.HourlyInvoice) (int64, error)
	GetHourlyInvoice(ctx context.Context, invoiceID int64) (domain.HourlyInvoice, error)
	GetHourlyInvoiceForActor(ctx context.Context, invoiceID int64, actorID uuid.UUID) (domain.HourlyInvoice, error)
	ListHourlyInvoicesForActor(ctx context.Context, contractID int64, actorID uuid.UUID, limit, offset int) ([]domain.HourlyInvoice, error)
	GetHourlyInvoiceByContractWeek(ctx context.Context, contractID int64, weekStart time.Time) (domain.HourlyInvoice, error)
	AttachHourlyLogsToInvoice(ctx context.Context, contractID int64, weekStart time.Time, weekEnd time.Time, invoiceID int64) error
	MarkHourlyInvoiceStatus(ctx context.Context, invoiceID int64, status string, disputeID string, at time.Time) error
	CreateContractBonus(ctx context.Context, bonus domain.ContractBonus) (int64, error)
	GetContractBonusForActor(ctx context.Context, bonusID int64, actorID uuid.UUID) (domain.ContractBonus, error)
	GetContractBonus(ctx context.Context, bonusID int64) (domain.ContractBonus, error)
	ListContractBonusesForActor(ctx context.Context, contractID int64, actorID uuid.UUID, limit, offset int) ([]domain.ContractBonus, error)
	MarkContractBonusStatus(ctx context.Context, bonusID int64, status string, paymentReferenceID string, at time.Time) error
	HasBlockingFinancialActivity(ctx context.Context, contractID int64) (bool, string, error)
	CreateAmendmentForActor(ctx context.Context, a domain.Amendment) (int64, error)
	RespondAmendmentForActor(ctx context.Context, amendmentID int64, actorID uuid.UUID, status string, responseNote string, at time.Time) error
	RespondAmendmentAndApplyForActor(ctx context.Context, amendmentID int64, actorID uuid.UUID, responseNote string, at time.Time) error
	GetAmendmentForActor(ctx context.Context, amendmentID int64, actorID uuid.UUID) (domain.Amendment, error)
	ListAmendmentsForActor(ctx context.Context, contractID int64, actorID uuid.UUID, limit, offset int) ([]domain.Amendment, error)
	ExpirePendingAmendmentsForActor(ctx context.Context, contractID int64, actorID uuid.UUID, at time.Time) error
	ExpireAmendmentForActor(ctx context.Context, amendmentID int64, actorID uuid.UUID, at time.Time) (bool, error)
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
	MarkOfferSent(ctx context.Context, proposalID int64, clientID uuid.UUID, reason string) error
	SetHired(ctx context.Context, proposalID int64, clientID uuid.UUID, reason string) error
	ReleaseOffer(ctx context.Context, proposalID int64, clientID uuid.UUID, reason string) error
}

type JobSummary struct {
	JobID    int64
	ClientID string
	Status   string
	IsOpen   bool
	Found    bool
}

type JobReader interface {
	GetSummary(ctx context.Context, jobID int64, clientID uuid.UUID) (JobSummary, error)
}

type JobStatusSync interface {
	SetInProgress(ctx context.Context, jobID int64, clientID uuid.UUID) error
}

type ActorPolicy interface {
	EnsureClientCanHire(ctx context.Context, userID uuid.UUID) error
	EnsureFreelancerCanWork(ctx context.Context, userID uuid.UUID) error
}

type MilestoneApprovedSettlementCommand struct {
	EventID      string
	ContractID   int64
	MilestoneID  int64
	ReferenceID  string
	AmountMinor  int64
	FreelancerID string
}

type MilestoneSettlementDispatcher interface {
	DispatchMilestoneApproved(ctx context.Context, cmd MilestoneApprovedSettlementCommand) error
}

type MilestoneFundingNotifier interface {
	MarkMilestoneFunded(ctx context.Context, contractID int64, milestoneID int64) error
}

type DisputeReader interface {
	GetOpenDisputeID(ctx context.Context, referenceType, referenceID string) (string, error)
}

type Clock interface {
	Now() time.Time
}

type HourlyEvidenceObjectStore interface {
	BuildObjectKey(contractID int64, fileName string) string
	PresignPutObject(ctx context.Context, storageKey string, ttl time.Duration) (string, error)
}
