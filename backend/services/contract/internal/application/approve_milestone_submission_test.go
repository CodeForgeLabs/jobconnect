package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"jobconnect/contract/internal/domain"

	"github.com/google/uuid"
)

type milestoneRepoStub struct {
	contract domain.Contract
}

func (r *milestoneRepoStub) Create(context.Context, domain.Contract) (int64, error) { return 0, nil }
func (r *milestoneRepoStub) GetByID(context.Context, int64) (domain.Contract, error) {
	return r.contract, nil
}
func (r *milestoneRepoStub) GetByIDForActor(_ context.Context, contractID int64, actorID uuid.UUID) (domain.Contract, error) {
	if contractID != r.contract.ID {
		return domain.Contract{}, fmt.Errorf("not found")
	}
	if actorID != r.contract.ClientID && actorID != r.contract.FreelancerID {
		return domain.Contract{}, fmt.Errorf("not found")
	}
	return r.contract, nil
}
func (r *milestoneRepoStub) GetByProposalID(context.Context, int64) (domain.Contract, error) {
	return domain.Contract{}, fmt.Errorf("not found")
}
func (r *milestoneRepoStub) GetJobOfferState(context.Context, int64, uuid.UUID) (domain.JobOfferState, error) {
	return domain.JobOfferState{}, nil
}
func (r *milestoneRepoStub) ListByActor(context.Context, uuid.UUID, string, int, int) ([]domain.Contract, error) {
	return nil, nil
}
func (r *milestoneRepoStub) UpdateOfferForClient(context.Context, domain.Contract) error { return nil }
func (r *milestoneRepoStub) SetStatusForFreelancer(context.Context, int64, uuid.UUID, string, time.Time) error {
	return nil
}
func (r *milestoneRepoStub) SetStatusForClient(context.Context, int64, uuid.UUID, string, time.Time) error {
	return nil
}
func (r *milestoneRepoStub) ReplaceMilestonesForActor(_ context.Context, contractID int64, actorID uuid.UUID, milestones []domain.Milestone, at time.Time) error {
	if contractID != r.contract.ID {
		return fmt.Errorf("not found")
	}
	if actorID != r.contract.ClientID && actorID != r.contract.FreelancerID {
		return fmt.Errorf("not found")
	}
	r.contract.Milestones = milestones
	r.contract.UpdatedAt = at
	return nil
}
func (r *milestoneRepoStub) CreateHourlyLogForFreelancer(context.Context, domain.HourlyLog) (int64, error) {
	return 0, nil
}
func (r *milestoneRepoStub) ListHourlyLogsForActor(context.Context, int64, uuid.UUID, int, int) ([]domain.HourlyLog, error) {
	return nil, nil
}
func (r *milestoneRepoStub) ReviewHourlyLogForClient(context.Context, int64, uuid.UUID, string, string, time.Time) error {
	return nil
}
func (r *milestoneRepoStub) GetHourlyLogForActor(context.Context, int64, uuid.UUID) (domain.HourlyLog, error) {
	return domain.HourlyLog{}, nil
}
func (r *milestoneRepoStub) CreateAmendmentForActor(context.Context, domain.Amendment) (int64, error) {
	return 0, nil
}
func (r *milestoneRepoStub) RespondAmendmentForActor(context.Context, int64, uuid.UUID, string, string, time.Time) error {
	return nil
}
func (r *milestoneRepoStub) RespondAmendmentAndApplyForActor(context.Context, int64, uuid.UUID, string, time.Time) error {
	return nil
}
func (r *milestoneRepoStub) GetAmendmentForActor(context.Context, int64, uuid.UUID) (domain.Amendment, error) {
	return domain.Amendment{}, nil
}
func (r *milestoneRepoStub) ListAmendmentsForActor(context.Context, int64, uuid.UUID, int, int) ([]domain.Amendment, error) {
	return nil, nil
}
func (r *milestoneRepoStub) ExpirePendingAmendmentsForActor(context.Context, int64, uuid.UUID, time.Time) error {
	return nil
}
func (r *milestoneRepoStub) ExpireAmendmentForActor(context.Context, int64, uuid.UUID, time.Time) (bool, error) {
	return false, nil
}
func (r *milestoneRepoStub) AppendStatusHistory(context.Context, domain.StatusHistoryEntry) error {
	return nil
}
func (r *milestoneRepoStub) ListStatusHistoryForActor(context.Context, int64, uuid.UUID, int, int) ([]domain.StatusHistoryEntry, error) {
	return nil, nil
}

type disputeReaderStub struct {
	hasOpen bool
}

func (s disputeReaderStub) HasOpenDispute(context.Context, string, string) (bool, error) {
	return s.hasOpen, nil
}

type settlementDispatcherStub struct {
	err error
}

func (s settlementDispatcherStub) DispatchMilestoneApproved(context.Context, MilestoneApprovedSettlementCommand) error {
	return s.err
}

func TestApproveMilestoneSubmission_BlocksWhenOpenDisputeExists(t *testing.T) {
	clientID := uuid.New()
	freelancerID := uuid.New()
	repo := &milestoneRepoStub{
		contract: domain.Contract{
			ID:           99,
			ClientID:     clientID,
			FreelancerID: freelancerID,
			Milestones: []domain.Milestone{
				{ID: 7, Amount: 120, Status: domain.MilestoneStatusSubmitted},
			},
		},
	}
	clock := contractClockStub{now: time.Unix(1700000000, 0).UTC()}
	update := &UpdateMilestoneStatus{Contracts: repo, Clock: clock}
	uc := &ApproveMilestoneSubmission{
		UpdateMilestoneStatus: update,
		Disputes:              disputeReaderStub{hasOpen: true},
		Settlement:            settlementDispatcherStub{},
	}

	_, err := uc.Execute(context.Background(), ApproveMilestoneSubmissionInput{
		ContractID:  99,
		MilestoneID: 7,
		ActorID:     clientID,
		ActorRole:   "client",
	})
	if err == nil || err.Error() != "open dispute exists for milestone" {
		t.Fatalf("expected open dispute error, got %v", err)
	}
}

func TestApproveMilestoneSubmission_MarksPendingSettlementWhenDispatchFails(t *testing.T) {
	clientID := uuid.New()
	freelancerID := uuid.New()
	repo := &milestoneRepoStub{
		contract: domain.Contract{
			ID:           99,
			ClientID:     clientID,
			FreelancerID: freelancerID,
			Milestones: []domain.Milestone{
				{ID: 7, Amount: 120, Status: domain.MilestoneStatusSubmitted},
			},
		},
	}
	clock := contractClockStub{now: time.Unix(1700000000, 0).UTC()}
	update := &UpdateMilestoneStatus{Contracts: repo, Clock: clock}
	uc := &ApproveMilestoneSubmission{
		UpdateMilestoneStatus: update,
		Disputes:              disputeReaderStub{hasOpen: false},
		Settlement:            settlementDispatcherStub{err: fmt.Errorf("downstream failure")},
	}

	out, err := uc.Execute(context.Background(), ApproveMilestoneSubmissionInput{
		ContractID:  99,
		MilestoneID: 7,
		ActorID:     clientID,
		ActorRole:   "client",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := out.Contract.Milestones[0].Status; got != domain.MilestoneStatusApprovedPendingSettlement {
		t.Fatalf("expected %q, got %q", domain.MilestoneStatusApprovedPendingSettlement, got)
	}
}
