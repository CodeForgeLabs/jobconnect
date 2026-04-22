package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"jobconnect/contract/internal/domain"

	"github.com/google/uuid"
)

type createContractRepoStub struct {
	createID       int64
	existing       domain.Contract
	existingErr    error
	updateErr      error
	offerState     domain.JobOfferState
	created        []domain.Contract
	updated        []domain.Contract
	statusChanges  []string
	historyReasons []string
}

func (r *createContractRepoStub) Create(_ context.Context, c domain.Contract) (int64, error) {
	r.created = append(r.created, c)
	if r.createID == 0 {
		r.createID = 101
	}
	r.existing = c
	r.existing.ID = r.createID
	return r.createID, nil
}

func (r *createContractRepoStub) GetByID(_ context.Context, contractID int64) (domain.Contract, error) {
	if r.existing.ID == contractID {
		return r.existing, nil
	}
	return domain.Contract{}, fmt.Errorf("not found")
}

func (r *createContractRepoStub) GetByIDForActor(_ context.Context, _ int64, _ uuid.UUID) (domain.Contract, error) {
	return domain.Contract{}, fmt.Errorf("not implemented")
}

func (r *createContractRepoStub) GetByProposalID(_ context.Context, _ int64) (domain.Contract, error) {
	if r.existingErr != nil {
		return domain.Contract{}, r.existingErr
	}
	return r.existing, nil
}

func (r *createContractRepoStub) GetJobOfferState(_ context.Context, _ int64, _ uuid.UUID) (domain.JobOfferState, error) {
	return r.offerState, nil
}

func (r *createContractRepoStub) ListByActor(_ context.Context, _ uuid.UUID, _ string, _, _ int) ([]domain.Contract, error) {
	return nil, nil
}

func (r *createContractRepoStub) UpdateOfferForClient(_ context.Context, c domain.Contract) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	r.updated = append(r.updated, c)
	r.existing = c
	return nil
}

func (r *createContractRepoStub) SetStatusForFreelancer(_ context.Context, _ int64, _ uuid.UUID, _ string, _ time.Time) error {
	return nil
}

func (r *createContractRepoStub) SetStatusForClient(_ context.Context, _ int64, _ uuid.UUID, status string, _ time.Time) error {
	r.statusChanges = append(r.statusChanges, status)
	return nil
}

func (r *createContractRepoStub) ReplaceMilestonesForActor(_ context.Context, _ int64, _ uuid.UUID, _ []domain.Milestone, _ time.Time) error {
	return nil
}

func (r *createContractRepoStub) CreateHourlyLogForFreelancer(_ context.Context, _ domain.HourlyLog) (int64, error) {
	return 0, nil
}

func (r *createContractRepoStub) ListHourlyLogsForActor(_ context.Context, _ int64, _ uuid.UUID, _, _ int) ([]domain.HourlyLog, error) {
	return nil, nil
}

func (r *createContractRepoStub) ReviewHourlyLogForClient(_ context.Context, _ int64, _ uuid.UUID, _, _ string, _ time.Time) error {
	return nil
}

func (r *createContractRepoStub) GetHourlyLogForActor(_ context.Context, _ int64, _ uuid.UUID) (domain.HourlyLog, error) {
	return domain.HourlyLog{}, nil
}

func (r *createContractRepoStub) CreateAmendmentForActor(_ context.Context, _ domain.Amendment) (int64, error) {
	return 0, nil
}

func (r *createContractRepoStub) RespondAmendmentForActor(_ context.Context, _ int64, _ uuid.UUID, _ string, _ time.Time) error {
	return nil
}

func (r *createContractRepoStub) GetAmendmentForActor(_ context.Context, _ int64, _ uuid.UUID) (domain.Amendment, error) {
	return domain.Amendment{}, nil
}

func (r *createContractRepoStub) ListAmendmentsForActor(_ context.Context, _ int64, _ uuid.UUID, _, _ int) ([]domain.Amendment, error) {
	return nil, nil
}

func (r *createContractRepoStub) AppendStatusHistory(_ context.Context, entry domain.StatusHistoryEntry) error {
	r.historyReasons = append(r.historyReasons, entry.Reason)
	return nil
}

func (r *createContractRepoStub) ListStatusHistoryForActor(_ context.Context, _ int64, _ uuid.UUID, _, _ int) ([]domain.StatusHistoryEntry, error) {
	return nil, nil
}

type proposalSyncStub struct {
	proposal        ProposalSummary
	markOfferCalls  []int64
	markOfferErr    error
	markOfferReason []string
	setHiredCalls   []int64
	setHiredReason  []string
	setHiredErr     error
	releaseCalls    []int64
	releaseReason   []string
	releaseErr      error
}

func (p *proposalSyncStub) GetProposal(_ context.Context, _ int64, _ uuid.UUID) (ProposalSummary, error) {
	return p.proposal, nil
}

func (p *proposalSyncStub) MarkOfferSent(_ context.Context, proposalID int64, _ uuid.UUID, reason string) error {
	p.markOfferCalls = append(p.markOfferCalls, proposalID)
	p.markOfferReason = append(p.markOfferReason, reason)
	return p.markOfferErr
}

func (p *proposalSyncStub) SetHired(_ context.Context, proposalID int64, _ uuid.UUID, reason string) error {
	p.setHiredCalls = append(p.setHiredCalls, proposalID)
	p.setHiredReason = append(p.setHiredReason, reason)
	return p.setHiredErr
}

func (p *proposalSyncStub) ReleaseOffer(_ context.Context, proposalID int64, _ uuid.UUID, reason string) error {
	p.releaseCalls = append(p.releaseCalls, proposalID)
	p.releaseReason = append(p.releaseReason, reason)
	return p.releaseErr
}

type actorPolicyStub struct{}

func (a *actorPolicyStub) EnsureClientCanHire(context.Context, uuid.UUID) error     { return nil }
func (a *actorPolicyStub) EnsureFreelancerCanWork(context.Context, uuid.UUID) error { return nil }

type jobReaderStub struct {
	summary JobSummary
}

func (j *jobReaderStub) GetSummary(_ context.Context, _ int64, _ uuid.UUID) (JobSummary, error) {
	return j.summary, nil
}

type contractClockStub struct {
	now time.Time
}

func (c contractClockStub) Now() time.Time { return c.now }

func TestCreateContract_Execute_SendsOfferFromShortlistedProposal(t *testing.T) {
	clientID := uuid.New()
	freelancerID := uuid.New()
	repo := &createContractRepoStub{existingErr: fmt.Errorf("not found")}
	proposals := &proposalSyncStub{proposal: ProposalSummary{
		ID:           8,
		JobID:        21,
		ClientID:     clientID.String(),
		FreelancerID: freelancerID.String(),
		Status:       "shortlisted",
	}}
	jobs := &jobReaderStub{summary: JobSummary{JobID: 21, ClientID: clientID.String(), IsOpen: true, Found: true}}
	uc := &CreateContract{
		Contracts: repo,
		Proposals: proposals,
		Jobs:      jobs,
		Actors:    &actorPolicyStub{},
		Clock:     contractClockStub{now: time.Unix(1700000000, 0).UTC()},
	}

	out, err := uc.Execute(context.Background(), CreateContractInput{
		ClientID:     clientID,
		FreelancerID: freelancerID,
		JobID:        21,
		ProposalID:   8,
		ContractType: domain.TypeFixed,
		Title:        "Build API",
		FixedTotal:   500,
	})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if out.Contract.Status != domain.StatusPendingAcceptance {
		t.Fatalf("expected pending contract, got %q", out.Contract.Status)
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected one contract create, got %d", len(repo.created))
	}
	if len(proposals.markOfferCalls) != 1 || proposals.markOfferCalls[0] != 8 {
		t.Fatalf("expected proposal to be marked offer_sent, got %+v", proposals.markOfferCalls)
	}
	if len(repo.historyReasons) != 1 || repo.historyReasons[0] != "offer sent" {
		t.Fatalf("unexpected history reasons: %+v", repo.historyReasons)
	}
}

func TestCreateContract_Execute_BlocksWhenPendingOfferExists(t *testing.T) {
	clientID := uuid.New()
	freelancerID := uuid.New()
	repo := &createContractRepoStub{
		existingErr: fmt.Errorf("not found"),
		offerState:  domain.JobOfferState{JobID: 21, HasPendingOffer: true, PendingContractID: 77},
	}
	proposals := &proposalSyncStub{proposal: ProposalSummary{
		ID:           8,
		JobID:        21,
		ClientID:     clientID.String(),
		FreelancerID: freelancerID.String(),
		Status:       "sent",
	}}
	jobs := &jobReaderStub{summary: JobSummary{JobID: 21, ClientID: clientID.String(), IsOpen: true, Found: true}}
	uc := &CreateContract{
		Contracts: repo,
		Proposals: proposals,
		Jobs:      jobs,
		Actors:    &actorPolicyStub{},
		Clock:     contractClockStub{now: time.Unix(1700000000, 0).UTC()},
	}

	_, err := uc.Execute(context.Background(), CreateContractInput{
		ClientID:     clientID,
		FreelancerID: freelancerID,
		JobID:        21,
		ProposalID:   8,
		ContractType: domain.TypeFixed,
		Title:        "Build API",
		FixedTotal:   500,
	})
	if err == nil || err.Error() != "job already has a pending offer" {
		t.Fatalf("expected pending offer conflict, got %v", err)
	}
}

func TestCreateContract_Execute_ReopensRevokedOfferAndMarksProposalHired(t *testing.T) {
	clientID := uuid.New()
	freelancerID := uuid.New()
	repo := &createContractRepoStub{
		existing: domain.Contract{
			ID:           91,
			ClientID:     clientID,
			FreelancerID: freelancerID,
			JobID:        21,
			ProposalID:   8,
			Status:       domain.StatusRevoked,
		},
	}
	proposals := &proposalSyncStub{proposal: ProposalSummary{
		ID:           8,
		JobID:        21,
		ClientID:     clientID.String(),
		FreelancerID: freelancerID.String(),
		Status:       "shortlisted",
	}}
	jobs := &jobReaderStub{summary: JobSummary{JobID: 21, ClientID: clientID.String(), IsOpen: true, Found: true}}
	uc := &CreateContract{
		Contracts: repo,
		Proposals: proposals,
		Jobs:      jobs,
		Actors:    &actorPolicyStub{},
		Clock:     contractClockStub{now: time.Unix(1700000000, 0).UTC()},
	}

	out, err := uc.Execute(context.Background(), CreateContractInput{
		ClientID:     clientID,
		FreelancerID: freelancerID,
		JobID:        21,
		ProposalID:   8,
		ContractType: domain.TypeFixed,
		Title:        "Build API",
		FixedTotal:   500,
	})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if out.Contract.ID != 91 {
		t.Fatalf("expected reused contract id 91, got %d", out.Contract.ID)
	}
	if len(repo.updated) != 1 {
		t.Fatalf("expected one offer update, got %d", len(repo.updated))
	}
	if len(proposals.markOfferCalls) != 1 || proposals.markOfferCalls[0] != 8 {
		t.Fatalf("expected proposal to be marked offer_sent on resend, got %+v", proposals.markOfferCalls)
	}
	if len(repo.historyReasons) != 1 || repo.historyReasons[0] != "offer resent" {
		t.Fatalf("unexpected history reasons: %+v", repo.historyReasons)
	}
}

func TestCreateContract_Execute_DoesNotPersistResendWhenProposalSyncFails(t *testing.T) {
	clientID := uuid.New()
	freelancerID := uuid.New()
	repo := &createContractRepoStub{
		existing: domain.Contract{
			ID:           91,
			ClientID:     clientID,
			FreelancerID: freelancerID,
			JobID:        21,
			ProposalID:   8,
			Status:       domain.StatusRevoked,
			Title:        "Old title",
		},
	}
	proposals := &proposalSyncStub{
		proposal: ProposalSummary{
			ID:           8,
			JobID:        21,
			ClientID:     clientID.String(),
			FreelancerID: freelancerID.String(),
			Status:       "shortlisted",
		},
		markOfferErr: fmt.Errorf("proposal service unavailable"),
	}
	jobs := &jobReaderStub{summary: JobSummary{JobID: 21, ClientID: clientID.String(), IsOpen: true, Found: true}}
	uc := &CreateContract{
		Contracts: repo,
		Proposals: proposals,
		Jobs:      jobs,
		Actors:    &actorPolicyStub{},
		Clock:     contractClockStub{now: time.Unix(1700000000, 0).UTC()},
	}

	_, err := uc.Execute(context.Background(), CreateContractInput{
		ClientID:     clientID,
		FreelancerID: freelancerID,
		JobID:        21,
		ProposalID:   8,
		ContractType: domain.TypeFixed,
		Title:        "New title",
		FixedTotal:   500,
	})
	if err == nil {
		t.Fatal("expected resend to fail when proposal sync fails")
	}
	if len(repo.updated) != 0 {
		t.Fatalf("expected offer update to be skipped, got %d updates", len(repo.updated))
	}
	if repo.existing.Title != "Old title" {
		t.Fatalf("expected existing contract to remain unchanged, got title %q", repo.existing.Title)
	}
	if len(proposals.releaseCalls) != 0 {
		t.Fatalf("expected no proposal release compensation, got %+v", proposals.releaseCalls)
	}
}

func TestCreateContract_Execute_ReleasesProposalIfResendUpdateFailsAfterSync(t *testing.T) {
	clientID := uuid.New()
	freelancerID := uuid.New()
	repo := &createContractRepoStub{
		existing: domain.Contract{
			ID:           91,
			ClientID:     clientID,
			FreelancerID: freelancerID,
			JobID:        21,
			ProposalID:   8,
			Status:       domain.StatusRevoked,
		},
		updateErr: fmt.Errorf("db write failed"),
	}
	proposals := &proposalSyncStub{proposal: ProposalSummary{
		ID:           8,
		JobID:        21,
		ClientID:     clientID.String(),
		FreelancerID: freelancerID.String(),
		Status:       "shortlisted",
	}}
	jobs := &jobReaderStub{summary: JobSummary{JobID: 21, ClientID: clientID.String(), IsOpen: true, Found: true}}
	uc := &CreateContract{
		Contracts: repo,
		Proposals: proposals,
		Jobs:      jobs,
		Actors:    &actorPolicyStub{},
		Clock:     contractClockStub{now: time.Unix(1700000000, 0).UTC()},
	}

	_, err := uc.Execute(context.Background(), CreateContractInput{
		ClientID:     clientID,
		FreelancerID: freelancerID,
		JobID:        21,
		ProposalID:   8,
		ContractType: domain.TypeFixed,
		Title:        "Build API",
		FixedTotal:   500,
	})
	if err == nil {
		t.Fatal("expected resend to fail when offer update fails")
	}
	if len(proposals.markOfferCalls) != 1 || proposals.markOfferCalls[0] != 8 {
		t.Fatalf("expected proposal to be marked offer_sent before update failure, got %+v", proposals.markOfferCalls)
	}
	if len(proposals.releaseCalls) != 1 || proposals.releaseCalls[0] != 8 {
		t.Fatalf("expected proposal hire release compensation, got %+v", proposals.releaseCalls)
	}
	if len(repo.updated) != 0 {
		t.Fatalf("expected no contract update to persist, got %d updates", len(repo.updated))
	}
}
