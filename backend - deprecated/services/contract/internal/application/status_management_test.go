package application

import (
	"context"
	"strings"
	"testing"
	"time"

	"jobconnect/contract/internal/domain"

	"github.com/google/uuid"
)

type endContractRepoStub struct {
	createContractRepoStub

	contract      domain.Contract
	blocked       bool
	blockedReason string
}

func (r *endContractRepoStub) GetByIDForActor(_ context.Context, contractID int64, actorID uuid.UUID) (domain.Contract, error) {
	if r.contract.ID == contractID && (r.contract.ClientID == actorID || r.contract.FreelancerID == actorID) {
		return r.contract, nil
	}
	return domain.Contract{}, errNotFoundForTest{}
}

func (r *endContractRepoStub) SetStatusForClient(_ context.Context, contractID int64, clientID uuid.UUID, status string, at time.Time) error {
	if r.contract.ID != contractID || r.contract.ClientID != clientID {
		return errNotFoundForTest{}
	}
	r.contract.Status = status
	r.contract.UpdatedAt = at
	if status == domain.StatusEnded {
		r.contract.EndedAt = &at
	}
	return nil
}

func (r *endContractRepoStub) SetStatusForFreelancer(_ context.Context, contractID int64, freelancerID uuid.UUID, status string, at time.Time) error {
	if r.contract.ID != contractID || r.contract.FreelancerID != freelancerID {
		return errNotFoundForTest{}
	}
	r.contract.Status = status
	r.contract.UpdatedAt = at
	if status == domain.StatusEnded {
		r.contract.EndedAt = &at
	}
	return nil
}

func (r *endContractRepoStub) HasBlockingFinancialActivity(context.Context, int64) (bool, string, error) {
	return r.blocked, r.blockedReason, nil
}

type errNotFoundForTest struct{}

func (errNotFoundForTest) Error() string { return "not found" }

func TestEndContract_BlocksUnresolvedMilestone(t *testing.T) {
	clientID := uuid.New()
	repo := &endContractRepoStub{
		contract: domain.Contract{
			ID:       10,
			ClientID: clientID,
			Status:   domain.StatusActive,
			Milestones: []domain.Milestone{
				{ID: 1, Status: domain.MilestoneStatusFunded},
			},
		},
	}
	uc := &EndContract{Contracts: repo, Clock: contractClockStub{now: time.Unix(1700000000, 0).UTC()}}

	_, err := uc.Execute(context.Background(), EndContractInput{ContractID: 10, ActorID: clientID})
	if err == nil || !strings.Contains(err.Error(), "unresolved milestones") {
		t.Fatalf("expected unresolved milestone error, got %v", err)
	}
}

func TestEndContract_BlocksUnresolvedFinancialActivity(t *testing.T) {
	clientID := uuid.New()
	repo := &endContractRepoStub{
		contract: domain.Contract{
			ID:       10,
			ClientID: clientID,
			Status:   domain.StatusActive,
		},
		blocked:       true,
		blockedReason: "contract has unresolved hourly invoices",
	}
	uc := &EndContract{Contracts: repo, Clock: contractClockStub{now: time.Unix(1700000000, 0).UTC()}}

	_, err := uc.Execute(context.Background(), EndContractInput{ContractID: 10, ActorID: clientID})
	if err == nil || !strings.Contains(err.Error(), "hourly invoices") {
		t.Fatalf("expected hourly invoice block, got %v", err)
	}
}

func TestEndContract_RejectsFreelancerActor(t *testing.T) {
	clientID := uuid.New()
	freelancerID := uuid.New()
	repo := &endContractRepoStub{
		contract: domain.Contract{
			ID:           10,
			ClientID:     clientID,
			FreelancerID: freelancerID,
			Status:       domain.StatusActive,
		},
	}
	uc := &EndContract{Contracts: repo, Clock: contractClockStub{now: time.Unix(1700000000, 0).UTC()}}

	_, err := uc.Execute(context.Background(), EndContractInput{ContractID: 10, ActorID: freelancerID, Reason: "freelancer ended"})
	if err == nil || !strings.Contains(err.Error(), "only client can end contract") {
		t.Fatalf("expected freelancer end to be denied, got %v", err)
	}
}

func TestPauseResumeContract_AllowsFreelancerActor(t *testing.T) {
	clientID := uuid.New()
	freelancerID := uuid.New()
	repo := &endContractRepoStub{
		contract: domain.Contract{
			ID:           10,
			ClientID:     clientID,
			FreelancerID: freelancerID,
			Status:       domain.StatusActive,
		},
	}
	clock := contractClockStub{now: time.Unix(1700000000, 0).UTC()}

	pauseUC := &PauseContract{Contracts: repo, Clock: clock}
	pauseOut, err := pauseUC.Execute(context.Background(), PauseContractInput{
		ContractID: 10,
		ActorID:    freelancerID,
		Reason:     "freelancer pause",
	})
	if err != nil {
		t.Fatalf("pause error: %v", err)
	}
	if pauseOut.Contract.Status != domain.StatusPaused {
		t.Fatalf("expected paused status, got %q", pauseOut.Contract.Status)
	}

	resumeUC := &ResumeContract{Contracts: repo, Clock: clock}
	resumeOut, err := resumeUC.Execute(context.Background(), ResumeContractInput{
		ContractID: 10,
		ActorID:    freelancerID,
		Reason:     "freelancer resume",
	})
	if err != nil {
		t.Fatalf("resume error: %v", err)
	}
	if resumeOut.Contract.Status != domain.StatusActive {
		t.Fatalf("expected active status, got %q", resumeOut.Contract.Status)
	}
}
