package application

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"jobconnect/contract/internal/domain"

	"github.com/google/uuid"
)

type amendmentRepoStub struct {
	createContractRepoStub

	contract        domain.Contract
	amendment       domain.Amendment
	listItems       []domain.Amendment
	createAmendment domain.Amendment

	respondStatus      string
	respondNote        string
	respondCalled      bool
	respondApplyCalled bool
	expireCalled       bool
}

func (r *amendmentRepoStub) GetByIDForActor(_ context.Context, _ int64, _ uuid.UUID) (domain.Contract, error) {
	if r.contract.ID == 0 {
		return domain.Contract{}, fmt.Errorf("not found")
	}
	return r.contract, nil
}

func (r *amendmentRepoStub) CreateAmendmentForActor(_ context.Context, a domain.Amendment) (int64, error) {
	r.createAmendment = a
	if r.amendment.ID == 0 {
		r.amendment = a
		r.amendment.ID = 44
	}
	return r.amendment.ID, nil
}

func (r *amendmentRepoStub) GetAmendmentForActor(_ context.Context, _ int64, _ uuid.UUID) (domain.Amendment, error) {
	if r.amendment.ID == 0 {
		return domain.Amendment{}, fmt.Errorf("not found")
	}
	return r.amendment, nil
}

func (r *amendmentRepoStub) RespondAmendmentForActor(_ context.Context, _ int64, _ uuid.UUID, status string, responseNote string, _ time.Time) error {
	r.respondCalled = true
	r.respondStatus = status
	r.respondNote = responseNote
	r.amendment.Status = status
	r.amendment.ResponseNote = responseNote
	return nil
}

func (r *amendmentRepoStub) RespondAmendmentAndApplyForActor(_ context.Context, _ int64, _ uuid.UUID, responseNote string, _ time.Time) error {
	r.respondApplyCalled = true
	r.respondNote = responseNote
	r.amendment.Status = domain.AmendmentStatusAccepted
	r.amendment.ResponseNote = responseNote
	return nil
}

func (r *amendmentRepoStub) ListAmendmentsForActor(_ context.Context, _ int64, _ uuid.UUID, _, _ int) ([]domain.Amendment, error) {
	return r.listItems, nil
}

func (r *amendmentRepoStub) ExpirePendingAmendmentsForActor(_ context.Context, _ int64, _ uuid.UUID, _ time.Time) error {
	return nil
}

func (r *amendmentRepoStub) ExpireAmendmentForActor(_ context.Context, _ int64, _ uuid.UUID, _ time.Time) (bool, error) {
	r.expireCalled = true
	r.amendment.Status = domain.AmendmentStatusExpired
	return true, nil
}

func TestProposeAmendment_AllowsCompositePayload(t *testing.T) {
	actorID := uuid.New()
	repo := &amendmentRepoStub{
		contract: domain.Contract{
			ID:           10,
			ClientID:     actorID,
			FreelancerID: uuid.New(),
			ContractType: domain.TypeHourly,
			Status:       domain.StatusActive,
		},
	}
	uc := &ProposeAmendment{Contracts: repo, Clock: contractClockStub{now: time.Unix(1700000000, 0).UTC()}}

	_, err := uc.Execute(context.Background(), ProposeAmendmentInput{
		ContractID: repo.contract.ID,
		ActorID:    actorID,
		Summary:    "adjust terms",
		Payload: domain.AmendmentPayload{
			CompensationChange: &domain.CompensationChange{NewHourlyRate: 45},
			WeeklyLimitChange:  &domain.WeeklyLimitChange{NewWeeklyHourLimit: 30},
			ScopeChange:        &domain.ScopeChange{NewDescription: "Updated scope"},
		},
	})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if repo.createAmendment.Payload.CompensationChange == nil || repo.createAmendment.Payload.WeeklyLimitChange == nil || repo.createAmendment.Payload.ScopeChange == nil {
		t.Fatalf("expected composite payload to be persisted, got %+v", repo.createAmendment.Payload)
	}
}

func TestProposeAmendment_RejectsContractTypeMismatch(t *testing.T) {
	actorID := uuid.New()
	repo := &amendmentRepoStub{
		contract: domain.Contract{
			ID:           10,
			ClientID:     actorID,
			FreelancerID: uuid.New(),
			ContractType: domain.TypeHourly,
			Status:       domain.StatusActive,
		},
	}
	uc := &ProposeAmendment{Contracts: repo, Clock: contractClockStub{now: time.Unix(1700000000, 0).UTC()}}

	_, err := uc.Execute(context.Background(), ProposeAmendmentInput{
		ContractID: repo.contract.ID,
		ActorID:    actorID,
		Summary:    "bad mixed payload",
		Payload: domain.AmendmentPayload{
			MilestonesChange: &domain.MilestonesChange{Milestones: []domain.Milestone{{Title: "M1", Amount: 100}}},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "only allowed for fixed contracts") {
		t.Fatalf("expected hourly/milestone validation error, got %v", err)
	}
}

func TestRespondAmendment_AcceptUsesApplyPath(t *testing.T) {
	clientID := uuid.New()
	freelancerID := uuid.New()
	repo := &amendmentRepoStub{
		contract: domain.Contract{
			ID:           12,
			ClientID:     clientID,
			FreelancerID: freelancerID,
			ContractType: domain.TypeHourly,
			Status:       domain.StatusActive,
		},
		amendment: domain.Amendment{
			ID:         88,
			ContractID: 12,
			ProposedBy: freelancerID,
			Status:     domain.AmendmentStatusPending,
		},
	}
	uc := &RespondAmendment{Contracts: repo, Clock: contractClockStub{now: time.Unix(1700000000, 0).UTC()}}

	_, err := uc.Execute(context.Background(), RespondAmendmentInput{
		AmendmentID:  88,
		ActorID:      clientID,
		Status:       domain.AmendmentStatusAccepted,
		ResponseNote: "looks good",
	})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if !repo.respondApplyCalled {
		t.Fatal("expected accept to call RespondAmendmentAndApplyForActor")
	}
	if repo.respondCalled {
		t.Fatal("expected accept not to use plain RespondAmendmentForActor")
	}
}

func TestRespondAmendment_RejectRequiresNote(t *testing.T) {
	clientID := uuid.New()
	freelancerID := uuid.New()
	repo := &amendmentRepoStub{
		contract: domain.Contract{
			ID:           12,
			ClientID:     clientID,
			FreelancerID: freelancerID,
			ContractType: domain.TypeHourly,
			Status:       domain.StatusActive,
		},
		amendment: domain.Amendment{
			ID:         88,
			ContractID: 12,
			ProposedBy: freelancerID,
			Status:     domain.AmendmentStatusPending,
		},
	}
	uc := &RespondAmendment{Contracts: repo, Clock: contractClockStub{now: time.Unix(1700000000, 0).UTC()}}

	_, err := uc.Execute(context.Background(), RespondAmendmentInput{
		AmendmentID: 88,
		ActorID:     clientID,
		Status:      domain.AmendmentStatusRejected,
	})
	if err == nil || !strings.Contains(err.Error(), "response_note is required") {
		t.Fatalf("expected required response note error, got %v", err)
	}
}

func TestRespondAmendment_RejectsExpiredPending(t *testing.T) {
	clientID := uuid.New()
	freelancerID := uuid.New()
	expiredAt := time.Unix(1600000000, 0).UTC()
	now := time.Unix(1700000000, 0).UTC()
	repo := &amendmentRepoStub{
		contract: domain.Contract{
			ID:           12,
			ClientID:     clientID,
			FreelancerID: freelancerID,
			ContractType: domain.TypeHourly,
			Status:       domain.StatusActive,
		},
		amendment: domain.Amendment{
			ID:         88,
			ContractID: 12,
			ProposedBy: freelancerID,
			Status:     domain.AmendmentStatusPending,
			ExpiresAt:  &expiredAt,
		},
	}
	uc := &RespondAmendment{Contracts: repo, Clock: contractClockStub{now: now}}

	_, err := uc.Execute(context.Background(), RespondAmendmentInput{
		AmendmentID:  88,
		ActorID:      clientID,
		Status:       domain.AmendmentStatusAccepted,
		ResponseNote: "accept",
	})
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected expired error, got %v", err)
	}
	if !repo.expireCalled {
		t.Fatal("expected expiration path to run")
	}
}

func TestRespondAmendment_RejectsProposerAsResponder(t *testing.T) {
	actorID := uuid.New()
	repo := &amendmentRepoStub{
		contract: domain.Contract{
			ID:           12,
			ClientID:     actorID,
			FreelancerID: uuid.New(),
			ContractType: domain.TypeHourly,
			Status:       domain.StatusActive,
		},
		amendment: domain.Amendment{
			ID:         88,
			ContractID: 12,
			ProposedBy: actorID,
			Status:     domain.AmendmentStatusPending,
		},
	}
	uc := &RespondAmendment{Contracts: repo, Clock: contractClockStub{now: time.Unix(1700000000, 0).UTC()}}

	_, err := uc.Execute(context.Background(), RespondAmendmentInput{
		AmendmentID:  88,
		ActorID:      actorID,
		Status:       domain.AmendmentStatusRejected,
		ResponseNote: "no",
	})
	if err == nil || !strings.Contains(err.Error(), "counterparty") {
		t.Fatalf("expected counterparty restriction error, got %v", err)
	}
}
