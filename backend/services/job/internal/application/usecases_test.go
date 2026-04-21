package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"jobconnect/job/internal/domain"

	"github.com/google/uuid"
)

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time { return c.now }

type refundCall struct {
	userID      string
	amount      int32
	referenceID string
}

type fakeConnectsClient struct {
	refunds []refundCall
}

func (c *fakeConnectsClient) RefundConnects(ctx context.Context, userID string, amount int32, referenceID string) error {
	c.refunds = append(c.refunds, refundCall{userID: userID, amount: amount, referenceID: referenceID})
	return nil
}

type fakeProposalClient struct {
	proposalsByJob []Proposal
	listCalls      []int64
}

func (p *fakeProposalClient) ListProposalsByJob(ctx context.Context, jobID int64) ([]Proposal, error) {
	p.listCalls = append(p.listCalls, jobID)
	return p.proposalsByJob, nil
}

func (p *fakeProposalClient) GetProposal(ctx context.Context, proposalID int64) (Proposal, error) {
	return Proposal{}, nil
}

func (p *fakeProposalClient) SetProposalStatus(ctx context.Context, proposalID int64, status string, reason string) error {
	return nil
}

func (p *fakeProposalClient) HireProposal(ctx context.Context, proposalID int64, reason string) error {
	return nil
}

type fakeJobRepo struct {
	createID int64

	lastCreateJob          domain.Job
	lastGetByIDJobID       int64
	lastGetByIDForClientID int64
	lastGetByIDClientID    uuid.UUID
	lastListByClientStatus string
	lastListByClientLimit  int
	lastListByClientOffset int
	lastListOpenLimit      int
	lastListOpenOffset     int
	lastListOpenFilter     ListOpenFilter
	lastListOpenV2Filter   ListOpenFilter
	lastListOpenV2SortBy   string
	lastCloseJobID         int64
	lastCloseClientID      uuid.UUID
	lastCloseReason        string
	lastPauseJobID         int64
	lastPauseClientID      uuid.UUID
	lastReopenJobID        int64
	lastReopenClientID     uuid.UUID
	lastMarkFilledJobID    int64
	lastMarkFilledClientID uuid.UUID
	lastSaveJobID          int64
	lastSaveFreelancerID   uuid.UUID
	lastUnsaveJobID        int64
	lastUnsaveFreelancerID uuid.UUID
	lastRespondJobID       int64
	lastRespondFreelancer  uuid.UUID
	lastRespondStatus      string
	lastInviteJobID        int64
	lastInviteFreelancer   string
	lastMarkCompletedJobID int64
	lastCancelJobID        int64
	lastCancelPolicy       string
	lastCancelReason       string
	lastSetVisibilityJobID int64
	lastSetVisibilityValue string
	lastSetBudgetJobID     int64
	lastSetBudgetMin       float64
	lastSetBudgetMax       float64
	lastReopenHiringJobID  int64
	lastFacetQuery         string

	getByIDJob           domain.Job
	getByIDErr           error
	getByIDForClientJob  domain.Job
	getByIDForClientErr  error
	getPublicJob         domain.Job
	getPublicErr         error
	updateJob            domain.Job
	updateErr            error
	listByClientJobs     []domain.Job
	listByClientErr      error
	listOpenJobs         []domain.Job
	listOpenErr          error
	listOpenFilteredJobs []domain.Job
	listOpenFilteredErr  error
	listOpenV2Jobs       []domain.Job
	listOpenV2Err        error
	listInvitedJobs      []domain.InvitedJob
	listInvitedErr       error
	savedJobs            []domain.Job
	savedJobsErr         error
	respondUpdated       bool
	respondErr           error
	saved                bool
	saveErr              error
	removed              bool
	unsaveErr            error
	inviteStats          InviteStats
	inviteStatsErr       error
	completed            bool
	completedErr         error
	canceled             bool
	cancelErr            error
	visibilityJob        domain.Job
	visibilityErr        error
	budgetJob            domain.Job
	budgetErr            error
	pauseJob             domain.Job
	pauseErr             error
	reopenJob            domain.Job
	reopenErr            error
	markFilledJob        domain.Job
	markFilledErr        error
	reopenHiringJob      domain.Job
	reopenHiringErr      error
	closeErr             error
	facetCounts          FacetCountsResult
	facetCountsErr       error
	attachment           domain.Attachment
	attachmentErr        error
	attachments          []domain.Attachment
	attachmentsErr       error
	deletedAttachment    domain.Attachment
	deleteAttachmentErr  error
	addedAttachment      domain.Attachment
	addAttachmentErr     error
}

func (r *fakeJobRepo) Create(ctx context.Context, job domain.Job) (int64, error) {
	r.lastCreateJob = job
	if r.createID == 0 {
		r.createID = 1
	}
	return r.createID, nil
}

func (r *fakeJobRepo) GetByID(ctx context.Context, jobID int64) (domain.Job, error) {
	r.lastGetByIDJobID = jobID
	return r.getByIDJob, r.getByIDErr
}

func (r *fakeJobRepo) GetByIDForClient(ctx context.Context, jobID int64, clientID uuid.UUID) (domain.Job, error) {
	r.lastGetByIDForClientID = jobID
	r.lastGetByIDClientID = clientID
	return r.getByIDForClientJob, r.getByIDForClientErr
}

func (r *fakeJobRepo) GetPublicByID(ctx context.Context, jobID int64) (domain.Job, error) {
	return r.getPublicJob, r.getPublicErr
}

func (r *fakeJobRepo) Update(ctx context.Context, job domain.Job) (domain.Job, error) {
	r.updateJob = job
	return r.updateJob, r.updateErr
}

func (r *fakeJobRepo) AddAttachment(ctx context.Context, jobID int64, clientID uuid.UUID, attachment domain.Attachment) (domain.Attachment, error) {
	r.addedAttachment = attachment
	return r.addedAttachment, r.addAttachmentErr
}

func (r *fakeJobRepo) DeleteAttachment(ctx context.Context, jobID int64, attachmentID int64, clientID uuid.UUID) (domain.Attachment, error) {
	r.deletedAttachment = domain.Attachment{ID: attachmentID}
	return r.deletedAttachment, r.deleteAttachmentErr
}

func (r *fakeJobRepo) ListAttachments(ctx context.Context, jobID int64, clientID uuid.UUID) ([]domain.Attachment, error) {
	return r.attachments, r.attachmentsErr
}

func (r *fakeJobRepo) GetAttachment(ctx context.Context, jobID int64, attachmentID int64, clientID uuid.UUID) (domain.Attachment, error) {
	return r.attachment, r.attachmentErr
}

func (r *fakeJobRepo) ListByClient(ctx context.Context, clientID uuid.UUID, status string, limit, offset int) ([]domain.Job, error) {
	r.lastListByClientStatus = status
	r.lastListByClientLimit = limit
	r.lastListByClientOffset = offset
	return r.listByClientJobs, r.listByClientErr
}

func (r *fakeJobRepo) ListInvitedJobs(ctx context.Context, freelancerID uuid.UUID, limit, offset int) ([]domain.InvitedJob, error) {
	return r.listInvitedJobs, r.listInvitedErr
}

func (r *fakeJobRepo) RespondToInvite(ctx context.Context, jobID int64, freelancerID uuid.UUID, responseStatus string, respondedAt time.Time) (bool, error) {
	r.lastRespondJobID = jobID
	r.lastRespondFreelancer = freelancerID
	r.lastRespondStatus = responseStatus
	return r.respondUpdated, r.respondErr
}

func (r *fakeJobRepo) SaveJob(ctx context.Context, jobID int64, freelancerID uuid.UUID, createdAt time.Time) (bool, error) {
	r.lastSaveJobID = jobID
	r.lastSaveFreelancerID = freelancerID
	return r.saved, r.saveErr
}

func (r *fakeJobRepo) UnsaveJob(ctx context.Context, jobID int64, freelancerID uuid.UUID) (bool, error) {
	r.lastUnsaveJobID = jobID
	r.lastUnsaveFreelancerID = freelancerID
	return r.removed, r.unsaveErr
}

func (r *fakeJobRepo) ListSavedJobs(ctx context.Context, freelancerID uuid.UUID, limit, offset int) ([]domain.Job, error) {
	return r.savedJobs, r.savedJobsErr
}

func (r *fakeJobRepo) ListOpen(ctx context.Context, limit, offset int) ([]domain.Job, error) {
	r.lastListOpenLimit = limit
	r.lastListOpenOffset = offset
	return r.listOpenJobs, r.listOpenErr
}

func (r *fakeJobRepo) ListOpenFiltered(ctx context.Context, filter ListOpenFilter) ([]domain.Job, error) {
	r.lastListOpenFilter = filter
	return r.listOpenFilteredJobs, r.listOpenFilteredErr
}

func (r *fakeJobRepo) ListOpenFilteredV2(ctx context.Context, filter ListOpenFilter, sortBy string) ([]domain.Job, error) {
	r.lastListOpenV2Filter = filter
	r.lastListOpenV2SortBy = sortBy
	return r.listOpenV2Jobs, r.listOpenV2Err
}

func (r *fakeJobRepo) CountOpenFiltered(ctx context.Context, filter ListOpenFilter) (int64, error) {
	return 0, nil
}

func (r *fakeJobRepo) GetInviteStats(ctx context.Context, jobID int64) (InviteStats, error) {
	return r.inviteStats, r.inviteStatsErr
}

func (r *fakeJobRepo) MarkJobCompleted(ctx context.Context, jobID int64, clientID uuid.UUID, completedAt time.Time) (bool, error) {
	r.lastMarkCompletedJobID = jobID
	return r.completed, r.completedErr
}

func (r *fakeJobRepo) CancelJobWithSettlement(ctx context.Context, jobID int64, clientID uuid.UUID, settlementPolicy string, reason string, canceledAt time.Time) (bool, error) {
	r.lastCancelJobID = jobID
	r.lastCancelPolicy = settlementPolicy
	r.lastCancelReason = reason
	return r.canceled, r.cancelErr
}

func (r *fakeJobRepo) SetVisibility(ctx context.Context, jobID int64, clientID uuid.UUID, visibility string, updatedAt time.Time) (domain.Job, error) {
	r.lastSetVisibilityJobID = jobID
	r.lastSetVisibilityValue = visibility
	return r.visibilityJob, r.visibilityErr
}

func (r *fakeJobRepo) SetBudgetRange(ctx context.Context, jobID int64, clientID uuid.UUID, budgetMin, budgetMax float64, updatedAt time.Time) (domain.Job, error) {
	r.lastSetBudgetJobID = jobID
	r.lastSetBudgetMin = budgetMin
	r.lastSetBudgetMax = budgetMax
	return r.budgetJob, r.budgetErr
}

func (r *fakeJobRepo) InviteFreelancer(ctx context.Context, jobID int64, clientID uuid.UUID, freelancerID string, createdAt time.Time) (bool, error) {
	r.lastInviteJobID = jobID
	r.lastInviteFreelancer = freelancerID
	return true, nil
}

func (r *fakeJobRepo) Pause(ctx context.Context, jobID int64, clientID uuid.UUID, updatedAt time.Time) (domain.Job, error) {
	r.lastPauseJobID = jobID
	r.lastPauseClientID = clientID
	return r.pauseJob, r.pauseErr
}

func (r *fakeJobRepo) Reopen(ctx context.Context, jobID int64, clientID uuid.UUID, updatedAt time.Time) (domain.Job, error) {
	r.lastReopenJobID = jobID
	r.lastReopenClientID = clientID
	return r.reopenJob, r.reopenErr
}

func (r *fakeJobRepo) MarkFilled(ctx context.Context, jobID int64, clientID uuid.UUID, updatedAt time.Time) (domain.Job, error) {
	r.lastMarkFilledJobID = jobID
	r.lastMarkFilledClientID = clientID
	return r.markFilledJob, r.markFilledErr
}

func (r *fakeJobRepo) ReopenHiring(ctx context.Context, jobID int64, clientID uuid.UUID, updatedAt time.Time) (domain.Job, error) {
	r.lastReopenHiringJobID = jobID
	return r.reopenHiringJob, r.reopenHiringErr
}

func (r *fakeJobRepo) Close(ctx context.Context, jobID int64, clientID uuid.UUID, reason string, closedAt time.Time) error {
	r.lastCloseJobID = jobID
	r.lastCloseClientID = clientID
	r.lastCloseReason = reason
	return r.closeErr
}

func (r *fakeJobRepo) FacetCounts(ctx context.Context, query string) (FacetCountsResult, error) {
	r.lastFacetQuery = query
	return r.facetCounts, r.facetCountsErr
}

func TestCreateJob_Execute_HappyPath(t *testing.T) {
	now := time.Date(2026, time.April, 9, 12, 0, 0, 0, time.UTC)
	deadline := now.Add(24 * time.Hour).Unix()
	clientID := uuid.New()
	repo := &fakeJobRepo{
		createID: 99,
		getByIDJob: domain.Job{
			ID:          99,
			ClientID:    clientID,
			Title:       "Build API",
			Description: "Integrate billing",
			JobType:     domain.JobTypeFixed,
			BudgetFixed: 500,
			BudgetMin:   500,
			BudgetMax:   500,
			Currency:    "USD",
			Status:      domain.JobStatusOpen,
		},
	}

	uc := &CreateJob{Jobs: repo, Clock: fixedClock{now: now}}
	out, err := uc.Execute(context.Background(), CreateJobInput{
		ClientID:       clientID,
		Title:          "  Build API  ",
		Description:    "  Integrate billing  ",
		RequiredSkills: []string{"go", "grpc"},
		JobType:        " Fixed ",
		BudgetFixed:    500,
		Currency:       "usd",
		Deadline:       &deadline,
	})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if repo.lastCreateJob.Title != "Build API" {
		t.Fatalf("expected trimmed title, got %q", repo.lastCreateJob.Title)
	}
	if repo.lastCreateJob.Description != "Integrate billing" {
		t.Fatalf("expected trimmed description, got %q", repo.lastCreateJob.Description)
	}
	if repo.lastCreateJob.JobType != domain.JobTypeFixed {
		t.Fatalf("expected job type %q, got %q", domain.JobTypeFixed, repo.lastCreateJob.JobType)
	}
	if repo.lastCreateJob.Currency != "USD" {
		t.Fatalf("expected uppercased currency, got %q", repo.lastCreateJob.Currency)
	}
	if repo.lastCreateJob.Status != domain.JobStatusOpen {
		t.Fatalf("expected open status, got %q", repo.lastCreateJob.Status)
	}
	if repo.lastCreateJob.Visibility != domain.VisibilityPublic {
		t.Fatalf("expected public visibility, got %q", repo.lastCreateJob.Visibility)
	}
	if repo.lastCreateJob.BudgetMin != 500 || repo.lastCreateJob.BudgetMax != 500 {
		t.Fatalf("expected fixed budget range of 500, got min=%v max=%v", repo.lastCreateJob.BudgetMin, repo.lastCreateJob.BudgetMax)
	}
	if repo.lastCreateJob.Deadline == nil || !repo.lastCreateJob.Deadline.Equal(time.Unix(deadline, 0).UTC()) {
		t.Fatalf("expected deadline to be normalized")
	}
	if out.Job.ID != 99 {
		t.Fatalf("expected persisted job to be returned, got id %d", out.Job.ID)
	}
	if repo.lastGetByIDJobID != 99 {
		t.Fatalf("expected GetByID to be called with created id 99, got %d", repo.lastGetByIDJobID)
	}
}

func TestCreateJob_Execute_ExplicitPrivateVisibilityIsPreserved(t *testing.T) {
	now := time.Date(2026, time.April, 9, 12, 0, 0, 0, time.UTC)
	deadline := now.Add(24 * time.Hour).Unix()
	clientID := uuid.New()
	repo := &fakeJobRepo{
		createID:   1,
		getByIDJob: domain.Job{ID: 1, ClientID: clientID, Status: domain.JobStatusOpen},
	}

	uc := &CreateJob{Jobs: repo, Clock: fixedClock{now: now}}
	_, err := uc.Execute(context.Background(), CreateJobInput{
		ClientID:    clientID,
		Title:       "Build API",
		Description: "Integrate billing",
		JobType:     domain.JobTypeFixed,
		BudgetFixed: 500,
		Currency:    "USD",
		Visibility:  domain.VisibilityPrivate,
		Deadline:    &deadline,
	})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if repo.lastCreateJob.Visibility != domain.VisibilityPrivate {
		t.Fatalf("expected private visibility, got %q", repo.lastCreateJob.Visibility)
	}
}

func TestCreateJob_Execute_InvalidVisibilityFails(t *testing.T) {
	now := time.Date(2026, time.April, 9, 12, 0, 0, 0, time.UTC)
	deadline := now.Add(24 * time.Hour).Unix()
	clientID := uuid.New()
	repo := &fakeJobRepo{}

	uc := &CreateJob{Jobs: repo, Clock: fixedClock{now: now}}
	_, err := uc.Execute(context.Background(), CreateJobInput{
		ClientID:    clientID,
		Title:       "Build API",
		Description: "Integrate billing",
		JobType:     domain.JobTypeFixed,
		BudgetFixed: 500,
		Currency:    "USD",
		Visibility:  "weird",
		Deadline:    &deadline,
	})
	if err == nil {
		t.Fatal("expected validation error for invalid visibility")
	}
	if repo.lastCreateJob.ClientID != uuid.Nil {
		t.Fatal("expected repo not to be called on visibility validation failure")
	}
}

func TestCreateJob_Execute_ValidationError(t *testing.T) {
	repo := &fakeJobRepo{}
	uc := &CreateJob{Jobs: repo, Clock: fixedClock{now: time.Now().UTC()}}

	_, err := uc.Execute(context.Background(), CreateJobInput{
		ClientID:    uuid.New(),
		Title:       "Need support",
		Description: "Support needed",
		JobType:     domain.JobTypeHourly,
		Currency:    "USD",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if repo.lastCreateJob.ClientID != uuid.Nil {
		t.Fatal("expected repo not to be called on validation failure")
	}
}

func TestCloseJob_Execute_RefundsConnectsForCanceledReason(t *testing.T) {
	now := time.Date(2026, time.April, 9, 13, 0, 0, 0, time.UTC)
	clientID := uuid.New()
	proposals := &fakeProposalClient{proposalsByJob: []Proposal{
		{ID: 11, FreelancerID: "freelancer-1", ConnectsSpent: 5},
		{ID: 12, FreelancerID: "freelancer-2", ConnectsSpent: 0},
		{ID: 13, FreelancerID: "freelancer-3", ConnectsSpent: 2},
	}}
	connects := &fakeConnectsClient{}
	repo := &fakeJobRepo{}

	uc := &CloseJob{Jobs: repo, Proposals: proposals, Connects: connects, Clock: fixedClock{now: now}}
	out, err := uc.Execute(context.Background(), CloseJobInput{JobID: 42, ClientID: clientID, Reason: "  CANCELED  "})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if !out.Closed {
		t.Fatal("expected closed response")
	}
	if repo.lastCloseJobID != 42 || repo.lastCloseClientID != clientID {
		t.Fatalf("unexpected close call: jobID=%d clientID=%s", repo.lastCloseJobID, repo.lastCloseClientID)
	}
	if repo.lastCloseReason != domain.CloseReasonCanceled {
		t.Fatalf("expected normalized close reason %q, got %q", domain.CloseReasonCanceled, repo.lastCloseReason)
	}
	if len(connects.refunds) != 2 {
		t.Fatalf("expected 2 refund calls, got %d", len(connects.refunds))
	}
	if connects.refunds[0].referenceID != "job_canceled_42_proposal_11" || connects.refunds[1].referenceID != "job_canceled_42_proposal_13" {
		t.Fatalf("unexpected refund references: %+v", connects.refunds)
	}
	if proposals.listCalls[0] != 42 {
		t.Fatalf("expected proposal lookup for job 42, got %d", proposals.listCalls[0])
	}
}

func TestPauseReopenAndMarkFilled_Execute(t *testing.T) {
	now := time.Date(2026, time.April, 9, 14, 0, 0, 0, time.UTC)
	clientID := uuid.New()
	repo := &fakeJobRepo{
		pauseJob:      domain.Job{ID: 1, Status: domain.JobStatusPaused},
		reopenJob:     domain.Job{ID: 1, Status: domain.JobStatusOpen},
		markFilledJob: domain.Job{ID: 1, Status: domain.JobStatusFilled},
	}

	pauseUC := &PauseJob{Jobs: repo, Clock: fixedClock{now: now}}
	pauseOut, err := pauseUC.Execute(context.Background(), PauseJobInput{JobID: 1, ClientID: clientID})
	if err != nil {
		t.Fatalf("PauseJob error: %v", err)
	}
	if pauseOut.Job.Status != domain.JobStatusPaused {
		t.Fatalf("expected paused job, got %q", pauseOut.Job.Status)
	}
	if repo.lastPauseJobID != 1 || repo.lastPauseClientID != clientID {
		t.Fatalf("unexpected pause call: jobID=%d clientID=%s", repo.lastPauseJobID, repo.lastPauseClientID)
	}

	reopenUC := &ReopenJob{Jobs: repo, Clock: fixedClock{now: now}}
	reopenOut, err := reopenUC.Execute(context.Background(), ReopenJobInput{JobID: 1, ClientID: clientID})
	if err != nil {
		t.Fatalf("ReopenJob error: %v", err)
	}
	if reopenOut.Job.Status != domain.JobStatusOpen {
		t.Fatalf("expected reopened job, got %q", reopenOut.Job.Status)
	}
	if repo.lastReopenJobID != 1 || repo.lastReopenClientID != clientID {
		t.Fatalf("unexpected reopen call: jobID=%d clientID=%s", repo.lastReopenJobID, repo.lastReopenClientID)
	}

	markFilledUC := &MarkJobFilled{Jobs: repo, Clock: fixedClock{now: now}}
	markFilledOut, err := markFilledUC.Execute(context.Background(), MarkJobFilledInput{JobID: 1, ClientID: clientID})
	if err != nil {
		t.Fatalf("MarkJobFilled error: %v", err)
	}
	if markFilledOut.Job.Status != domain.JobStatusFilled {
		t.Fatalf("expected filled job, got %q", markFilledOut.Job.Status)
	}
	if repo.lastMarkFilledJobID != 1 || repo.lastMarkFilledClientID != clientID {
		t.Fatalf("unexpected mark filled call: jobID=%d clientID=%s", repo.lastMarkFilledJobID, repo.lastMarkFilledClientID)
	}
}

func TestListMyJobs_Execute_PaginatesAndNormalizesStatus(t *testing.T) {
	clientID := uuid.New()
	repo := &fakeJobRepo{listByClientJobs: make([]domain.Job, 20)}
	for i := range repo.listByClientJobs {
		repo.listByClientJobs[i] = domain.Job{ID: int64(i + 1)}
	}

	uc := &ListMyJobs{Jobs: repo}
	out, err := uc.Execute(context.Background(), ListMyJobsInput{
		ClientID:  clientID,
		Status:    " OPEN ",
		PageSize:  0,
		PageToken: "20",
	})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if repo.lastListByClientStatus != domain.JobStatusOpen {
		t.Fatalf("expected normalized status %q, got %q", domain.JobStatusOpen, repo.lastListByClientStatus)
	}
	if repo.lastListByClientLimit != defaultPageSize {
		t.Fatalf("expected default page size %d, got %d", defaultPageSize, repo.lastListByClientLimit)
	}
	if repo.lastListByClientOffset != 20 {
		t.Fatalf("expected offset 20, got %d", repo.lastListByClientOffset)
	}
	if out.NextPageToken != fmt.Sprintf("%d", 20+len(repo.listByClientJobs)) {
		t.Fatalf("unexpected next page token %q", out.NextPageToken)
	}
}

func TestSearchJobs_Execute_UsesFilters(t *testing.T) {
	repo := &fakeJobRepo{listOpenFilteredJobs: []domain.Job{{ID: 7}}}
	uc := &SearchJobs{Jobs: repo}

	out, err := uc.Execute(context.Background(), SearchJobsInput{
		PageSize:   5,
		PageToken:  "3",
		Query:      "  go grpc  ",
		Skills:     []string{"go", "grpc"},
		JobType:    " FIXED ",
		Visibility: " PUBLIC ",
	})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if repo.lastListOpenFilter.SearchQuery != "go grpc" {
		t.Fatalf("expected trimmed query, got %q", repo.lastListOpenFilter.SearchQuery)
	}
	if repo.lastListOpenFilter.JobType != domain.JobTypeFixed {
		t.Fatalf("expected normalized job type %q, got %q", domain.JobTypeFixed, repo.lastListOpenFilter.JobType)
	}
	if repo.lastListOpenFilter.Visibility != domain.VisibilityPublic {
		t.Fatalf("expected normalized visibility %q, got %q", domain.VisibilityPublic, repo.lastListOpenFilter.Visibility)
	}
	if repo.lastListOpenFilter.Limit != 5 || repo.lastListOpenFilter.Offset != 3 {
		t.Fatalf("unexpected paging values: %+v", repo.lastListOpenFilter)
	}
	if len(out.Jobs) != 1 || out.Jobs[0].ID != 7 {
		t.Fatalf("unexpected output jobs: %+v", out.Jobs)
	}
}
