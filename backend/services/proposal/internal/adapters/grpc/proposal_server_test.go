package grpcadapter

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	proposalv1 "jobconnect/proposal/gen/proposal/v1"
	"jobconnect/proposal/internal/application"
	"jobconnect/proposal/internal/domain"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type fakeTokenParser struct{}

func (f fakeTokenParser) ParseAccessToken(token string) (uuid.UUID, string, error) {
	switch strings.TrimSpace(token) {
	case "client-token":
		return testClientID, "client", nil
	case "freelancer-token":
		return testFreelancerID, "freelancer", nil
	default:
		return uuid.Nil, "", fmt.Errorf("invalid token")
	}
}

type fakeClock struct {
	now time.Time
}

func (f fakeClock) Now() time.Time { return f.now }

type fakeJobReader struct {
	getJobSummaryFn func(ctx context.Context, jobID int64) (application.JobSummary, error)
}

func (f *fakeJobReader) GetJobSummary(ctx context.Context, jobID int64) (application.JobSummary, error) {
	if f.getJobSummaryFn == nil {
		return application.JobSummary{}, fmt.Errorf("get job summary not implemented")
	}
	return f.getJobSummaryFn(ctx, jobID)
}

type fakeConnectsClient struct {
	deductFn func(ctx context.Context, userID uuid.UUID, amount int32, referenceID string) error
}

func (f *fakeConnectsClient) DeductConnects(ctx context.Context, userID uuid.UUID, amount int32, referenceID string) error {
	if f.deductFn == nil {
		return nil
	}
	return f.deductFn(ctx, userID, amount, referenceID)
}

type fakeJobLifecycle struct {
	markJobFilledFn func(ctx context.Context, jobID int64) error
}

func (f *fakeJobLifecycle) MarkJobFilled(ctx context.Context, jobID int64) error {
	if f.markJobFilledFn == nil {
		return nil
	}
	return f.markJobFilledFn(ctx, jobID)
}

type fakeContractCreator struct {
	createFn func(ctx context.Context, in application.CreateContractFromProposalInput) error
}

func (f *fakeContractCreator) CreateFromProposal(ctx context.Context, in application.CreateContractFromProposalInput) error {
	if f.createFn == nil {
		return nil
	}
	return f.createFn(ctx, in)
}

type fakeAttachmentStore struct {
	buildObjectKeyFn   func(proposalID int64, fileName string) string
	presignPutObjectFn func(ctx context.Context, storageKey string, ttl time.Duration) (string, error)
	presignGetObjectFn func(ctx context.Context, storageKey string, ttl time.Duration) (string, error)
}

func (f *fakeAttachmentStore) BuildObjectKey(proposalID int64, fileName string) string {
	if f.buildObjectKeyFn == nil {
		return fmt.Sprintf("proposal/%d/%s", proposalID, fileName)
	}
	return f.buildObjectKeyFn(proposalID, fileName)
}

func (f *fakeAttachmentStore) PresignPutObject(ctx context.Context, storageKey string, ttl time.Duration) (string, error) {
	if f.presignPutObjectFn == nil {
		return "https://upload.local/" + storageKey, nil
	}
	return f.presignPutObjectFn(ctx, storageKey, ttl)
}

func (f *fakeAttachmentStore) PresignGetObject(ctx context.Context, storageKey string, ttl time.Duration) (string, error) {
	if f.presignGetObjectFn == nil {
		return "https://download.local/" + storageKey, nil
	}
	return f.presignGetObjectFn(ctx, storageKey, ttl)
}

type fakeProposalRepo struct {
	createFn                      func(ctx context.Context, p domain.Proposal) (int64, error)
	getByIDFn                     func(ctx context.Context, proposalID int64) (domain.Proposal, error)
	getByIDForFreelancerFn        func(ctx context.Context, proposalID int64, freelancerID uuid.UUID) (domain.Proposal, error)
	getLatestByJobForFreelancerFn func(ctx context.Context, jobID int64, freelancerID uuid.UUID) (domain.Proposal, error)
	getByIDForClientFn            func(ctx context.Context, proposalID int64, clientID uuid.UUID) (domain.Proposal, error)
	hasActiveProposalFn           func(ctx context.Context, jobID int64, freelancerID uuid.UUID) (bool, error)
	updateEditableFn              func(ctx context.Context, proposalID int64, freelancerID uuid.UUID, coverLetter string, bidAmount float64, estimatedDays int32, attachments []domain.Attachment, updatedAt time.Time) error
	withdrawFn                    func(ctx context.Context, proposalID int64, freelancerID uuid.UUID, reason string, at time.Time) error
	setStatusFn                   func(ctx context.Context, proposalID int64, clientID uuid.UUID, status string, reason string, at time.Time) error
	markOfferSentFn               func(ctx context.Context, proposalID int64, clientID uuid.UUID, reason string, at time.Time) (domain.Proposal, error)
	revertHireFn                  func(ctx context.Context, proposalID int64, clientID uuid.UUID, reason string, at time.Time) error
	hireWithRequestIDFn           func(ctx context.Context, proposalID int64, clientID uuid.UUID, requestID string, reason string, at time.Time) (domain.Proposal, bool, error)
	hasHiredProposalForJobFn      func(ctx context.Context, jobID int64) (bool, error)
	listByJobFn                   func(ctx context.Context, filter application.ListByJobFilter, pageSize int, pageToken string) ([]domain.Proposal, string, error)
	listByFreelancerFn            func(ctx context.Context, filter application.ListByFreelancerFilter, pageSize int, pageToken string) ([]domain.Proposal, string, error)
	listByClientFn                func(ctx context.Context, filter application.ListByClientFilter, pageSize int, pageToken string) ([]domain.Proposal, string, error)
	countByJobForClientFn         func(ctx context.Context, clientID uuid.UUID, jobID int64) (int64, map[string]int64, error)
	countClientInboxFn            func(ctx context.Context, clientID uuid.UUID, statuses []string) (int64, map[string]int64, error)
}

func (f *fakeProposalRepo) Create(ctx context.Context, p domain.Proposal) (int64, error) {
	if f.createFn == nil {
		return 0, fmt.Errorf("create not implemented")
	}
	return f.createFn(ctx, p)
}

func (f *fakeProposalRepo) GetByID(ctx context.Context, proposalID int64) (domain.Proposal, error) {
	if f.getByIDFn == nil {
		return domain.Proposal{}, fmt.Errorf("get by id not implemented")
	}
	return f.getByIDFn(ctx, proposalID)
}

func (f *fakeProposalRepo) GetByIDForFreelancer(ctx context.Context, proposalID int64, freelancerID uuid.UUID) (domain.Proposal, error) {
	if f.getByIDForFreelancerFn == nil {
		return domain.Proposal{}, fmt.Errorf("get by id for freelancer not implemented")
	}
	return f.getByIDForFreelancerFn(ctx, proposalID, freelancerID)
}

func (f *fakeProposalRepo) GetLatestByJobForFreelancer(ctx context.Context, jobID int64, freelancerID uuid.UUID) (domain.Proposal, error) {
	if f.getLatestByJobForFreelancerFn == nil {
		return domain.Proposal{}, fmt.Errorf("get latest by job for freelancer not implemented")
	}
	return f.getLatestByJobForFreelancerFn(ctx, jobID, freelancerID)
}

func (f *fakeProposalRepo) GetByIDForClient(ctx context.Context, proposalID int64, clientID uuid.UUID) (domain.Proposal, error) {
	if f.getByIDForClientFn == nil {
		return domain.Proposal{}, fmt.Errorf("get by id for client not implemented")
	}
	return f.getByIDForClientFn(ctx, proposalID, clientID)
}

func (f *fakeProposalRepo) HasActiveProposal(ctx context.Context, jobID int64, freelancerID uuid.UUID) (bool, error) {
	if f.hasActiveProposalFn == nil {
		return false, nil
	}
	return f.hasActiveProposalFn(ctx, jobID, freelancerID)
}

func (f *fakeProposalRepo) UpdateEditable(ctx context.Context, proposalID int64, freelancerID uuid.UUID, coverLetter string, bidAmount float64, estimatedDays int32, attachments []domain.Attachment, updatedAt time.Time) error {
	if f.updateEditableFn == nil {
		return nil
	}
	return f.updateEditableFn(ctx, proposalID, freelancerID, coverLetter, bidAmount, estimatedDays, attachments, updatedAt)
}

func (f *fakeProposalRepo) Withdraw(ctx context.Context, proposalID int64, freelancerID uuid.UUID, reason string, at time.Time) error {
	if f.withdrawFn == nil {
		return nil
	}
	return f.withdrawFn(ctx, proposalID, freelancerID, reason, at)
}

func (f *fakeProposalRepo) SetStatus(ctx context.Context, proposalID int64, clientID uuid.UUID, status string, reason string, at time.Time) error {
	if f.setStatusFn == nil {
		return nil
	}
	return f.setStatusFn(ctx, proposalID, clientID, status, reason, at)
}

func (f *fakeProposalRepo) MarkOfferSent(ctx context.Context, proposalID int64, clientID uuid.UUID, reason string, at time.Time) (domain.Proposal, error) {
	if f.markOfferSentFn == nil {
		return domain.Proposal{}, fmt.Errorf("mark offer sent not implemented")
	}
	return f.markOfferSentFn(ctx, proposalID, clientID, reason, at)
}

func (f *fakeProposalRepo) RevertHire(ctx context.Context, proposalID int64, clientID uuid.UUID, reason string, at time.Time) error {
	if f.revertHireFn == nil {
		return nil
	}
	return f.revertHireFn(ctx, proposalID, clientID, reason, at)
}

func (f *fakeProposalRepo) HireWithRequestID(ctx context.Context, proposalID int64, clientID uuid.UUID, requestID string, reason string, at time.Time) (domain.Proposal, bool, error) {
	if f.hireWithRequestIDFn == nil {
		return domain.Proposal{}, false, fmt.Errorf("hire with request id not implemented")
	}
	return f.hireWithRequestIDFn(ctx, proposalID, clientID, requestID, reason, at)
}

func (f *fakeProposalRepo) HasHiredProposalForJob(ctx context.Context, jobID int64) (bool, error) {
	if f.hasHiredProposalForJobFn == nil {
		return false, nil
	}
	return f.hasHiredProposalForJobFn(ctx, jobID)
}

func (f *fakeProposalRepo) ListByJob(ctx context.Context, filter application.ListByJobFilter, pageSize int, pageToken string) ([]domain.Proposal, string, error) {
	if f.listByJobFn == nil {
		return nil, "", nil
	}
	return f.listByJobFn(ctx, filter, pageSize, pageToken)
}

func (f *fakeProposalRepo) ListByFreelancer(ctx context.Context, filter application.ListByFreelancerFilter, pageSize int, pageToken string) ([]domain.Proposal, string, error) {
	if f.listByFreelancerFn == nil {
		return nil, "", nil
	}
	return f.listByFreelancerFn(ctx, filter, pageSize, pageToken)
}

func (f *fakeProposalRepo) ListByClient(ctx context.Context, filter application.ListByClientFilter, pageSize int, pageToken string) ([]domain.Proposal, string, error) {
	if f.listByClientFn == nil {
		return nil, "", nil
	}
	return f.listByClientFn(ctx, filter, pageSize, pageToken)
}

func (f *fakeProposalRepo) CountByJobForClient(ctx context.Context, clientID uuid.UUID, jobID int64) (int64, map[string]int64, error) {
	if f.countByJobForClientFn == nil {
		return 0, map[string]int64{}, nil
	}
	return f.countByJobForClientFn(ctx, clientID, jobID)
}

func (f *fakeProposalRepo) CountClientInbox(ctx context.Context, clientID uuid.UUID, statuses []string) (int64, map[string]int64, error) {
	if f.countClientInboxFn == nil {
		return 0, map[string]int64{}, nil
	}
	return f.countClientInboxFn(ctx, clientID, statuses)
}

var (
	testClientID     = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	testFreelancerID = uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	testNow          = time.Unix(1710000000, 0).UTC()
)

func makeProposal(id int64, status string) domain.Proposal {
	attachment := domain.Attachment{
		ID:          9,
		FileName:    "cv.pdf",
		ContentType: "application/pdf",
		URL:         "https://cdn.local/cv.pdf",
		SizeBytes:   2048,
		StorageKey:  "proposal/1/cv.pdf",
	}
	return domain.Proposal{
		ID:            id,
		JobID:         1001,
		ClientID:      testClientID,
		FreelancerID:  testFreelancerID,
		CoverLetter:   "Experienced freelancer",
		BidType:       domain.BidTypeFixed,
		BidAmount:     250,
		EstimatedDays: 7,
		Attachments:   []domain.Attachment{attachment},
		Status:        status,
		ConnectsSpent: 8,
		CreatedAt:     testNow,
		UpdatedAt:     testNow,
	}
}

func authCtx(token string) context.Context {
	md := metadata.Pairs("authorization", "Bearer "+token)
	return metadata.NewIncomingContext(context.Background(), md)
}

func buildServer(repo *fakeProposalRepo, jobs *fakeJobReader, connects *fakeConnectsClient, lifecycle *fakeJobLifecycle, contracts *fakeContractCreator, store *fakeAttachmentStore) *ProposalServer {
	if jobs == nil {
		jobs = &fakeJobReader{getJobSummaryFn: func(ctx context.Context, jobID int64) (application.JobSummary, error) {
			return application.JobSummary{JobID: jobID, ClientID: testClientID, IsOpen: true, Found: true, Status: "open"}, nil
		}}
	}
	if connects == nil {
		connects = &fakeConnectsClient{}
	}
	if lifecycle == nil {
		lifecycle = &fakeJobLifecycle{}
	}
	if contracts == nil {
		contracts = &fakeContractCreator{}
	}
	if store == nil {
		store = &fakeAttachmentStore{}
	}

	clk := fakeClock{now: testNow}

	return NewProposalServer(
		&application.SubmitProposal{Proposals: repo, Jobs: jobs, Connects: connects, Clock: clk},
		&application.ModifyProposal{Proposals: repo, Clock: clk},
		&application.WithdrawProposal{Proposals: repo, Clock: clk},
		&application.GetProposal{Proposals: repo},
		&application.GetMyProposalForJob{Proposals: repo},
		&application.HasAppliedToJob{Proposals: repo},
		&application.GetProposalAttachmentUploadURL{Proposals: repo, Store: store, PutTTL: 15 * time.Minute},
		&application.GetProposalAttachmentDownloadURL{Proposals: repo, Store: store, GetTTL: 30 * time.Minute},
		&application.ListProposalsByJob{Proposals: repo},
		&application.ListMyProposals{Proposals: repo},
		&application.ListClientProposals{Proposals: repo},
		&application.CountProposalsByJob{Proposals: repo},
		&application.CountClientProposalInbox{Proposals: repo},
		&application.SetProposalStatus{Proposals: repo, Clock: clk},
		&application.InternalMarkProposalOfferSent{Proposals: repo, Clock: clk},
		&application.InternalHireProposal{Proposals: repo, Clock: clk},
		&application.ReleaseHiredProposal{Proposals: repo, Clock: clk},
		fakeTokenParser{},
	)
}

func assertCode(t *testing.T, err error, code codes.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with code %s", code)
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected grpc status error, got: %v", err)
	}
	if st.Code() != code {
		t.Fatalf("expected code %s, got %s (%s)", code, st.Code(), st.Message())
	}
}

func TestProposalServer_RPCs(t *testing.T) {
	t.Run("SubmitProposal success and role denial", func(t *testing.T) {
		repo := &fakeProposalRepo{}
		repo.createFn = func(ctx context.Context, p domain.Proposal) (int64, error) {
			if p.ConnectsSpent != 8 {
				t.Fatalf("expected connects_spent 8, got %d", p.ConnectsSpent)
			}
			return 101, nil
		}
		repo.getByIDFn = func(ctx context.Context, proposalID int64) (domain.Proposal, error) {
			p := makeProposal(proposalID, domain.StatusSent)
			p.ConnectsSpent = 8
			return p, nil
		}

		srv := buildServer(repo, nil, nil, nil, nil, nil)
		resp, err := srv.SubmitProposal(authCtx("freelancer-token"), &proposalv1.SubmitProposalRequest{
			JobId:         1001,
			CoverLetter:   "Experienced freelancer",
			BidType:       "fixed",
			BidAmount:     250,
			EstimatedDays: 7,
			Attachments:   []*proposalv1.ProposalAttachment{{FileName: "cv.pdf", ContentType: "application/pdf", Url: "https://cdn.local/cv.pdf", SizeBytes: 2048}},
			ConnectsSpent: 8,
		})
		if err != nil {
			t.Fatalf("SubmitProposal returned error: %v", err)
		}
		if resp.GetProposal().GetId() != 101 {
			t.Fatalf("expected proposal id 101, got %d", resp.GetProposal().GetId())
		}

		_, err = srv.SubmitProposal(authCtx("client-token"), &proposalv1.SubmitProposalRequest{JobId: 1001, CoverLetter: "x", BidType: "fixed", BidAmount: 10, EstimatedDays: 1, ConnectsSpent: 1})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("ModifyProposal success and invalid status", func(t *testing.T) {
		repo := &fakeProposalRepo{}
		call := 0
		repo.getByIDForFreelancerFn = func(ctx context.Context, proposalID int64, freelancerID uuid.UUID) (domain.Proposal, error) {
			call++
			if call == 1 {
				return makeProposal(300, domain.StatusSent), nil
			}
			p := makeProposal(300, domain.StatusSent)
			p.CoverLetter = "Updated"
			return p, nil
		}
		repo.updateEditableFn = func(ctx context.Context, proposalID int64, freelancerID uuid.UUID, coverLetter string, bidAmount float64, estimatedDays int32, attachments []domain.Attachment, updatedAt time.Time) error {
			if coverLetter != "Updated" {
				t.Fatalf("expected trimmed cover letter to be updated, got %q", coverLetter)
			}
			return nil
		}

		srv := buildServer(repo, nil, nil, nil, nil, nil)
		resp, err := srv.ModifyProposal(authCtx("freelancer-token"), &proposalv1.ModifyProposalRequest{ProposalId: 300, CoverLetter: " Updated ", BidAmount: 500, EstimatedDays: 10})
		if err != nil {
			t.Fatalf("ModifyProposal returned error: %v", err)
		}
		if resp.GetProposal().GetCoverLetter() != "Updated" {
			t.Fatalf("expected updated cover letter, got %q", resp.GetProposal().GetCoverLetter())
		}

		repo2 := &fakeProposalRepo{getByIDForFreelancerFn: func(ctx context.Context, proposalID int64, freelancerID uuid.UUID) (domain.Proposal, error) {
			return makeProposal(301, domain.StatusRejected), nil
		}}
		srv2 := buildServer(repo2, nil, nil, nil, nil, nil)
		_, err = srv2.ModifyProposal(authCtx("freelancer-token"), &proposalv1.ModifyProposalRequest{ProposalId: 301, CoverLetter: "x", BidAmount: 10, EstimatedDays: 1})
		assertCode(t, err, codes.Internal)
	})

	t.Run("WithdrawProposal success and invalid reason", func(t *testing.T) {
		repo := &fakeProposalRepo{}
		repo.getByIDForFreelancerFn = func(ctx context.Context, proposalID int64, freelancerID uuid.UUID) (domain.Proposal, error) {
			return makeProposal(400, domain.StatusShortlisted), nil
		}
		repo.withdrawFn = func(ctx context.Context, proposalID int64, freelancerID uuid.UUID, reason string, at time.Time) error {
			if reason != "No longer available" {
				t.Fatalf("unexpected reason %q", reason)
			}
			return nil
		}

		srv := buildServer(repo, nil, nil, nil, nil, nil)
		resp, err := srv.WithdrawProposal(authCtx("freelancer-token"), &proposalv1.WithdrawProposalRequest{ProposalId: 400, Reason: "No longer available"})
		if err != nil {
			t.Fatalf("WithdrawProposal returned error: %v", err)
		}
		if !resp.GetWithdrawn() {
			t.Fatalf("expected withdrawn=true")
		}

		_, err = srv.WithdrawProposal(authCtx("freelancer-token"), &proposalv1.WithdrawProposalRequest{ProposalId: 400, Reason: strings.Repeat("x", 501)})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("GetProposal supports client and freelancer", func(t *testing.T) {
		repo := &fakeProposalRepo{}
		repo.getByIDForClientFn = func(ctx context.Context, proposalID int64, clientID uuid.UUID) (domain.Proposal, error) {
			return makeProposal(500, domain.StatusSent), nil
		}
		repo.getByIDForFreelancerFn = func(ctx context.Context, proposalID int64, freelancerID uuid.UUID) (domain.Proposal, error) {
			return makeProposal(501, domain.StatusSent), nil
		}

		srv := buildServer(repo, nil, nil, nil, nil, nil)
		clientResp, err := srv.GetProposal(authCtx("client-token"), &proposalv1.GetProposalRequest{ProposalId: 500})
		if err != nil {
			t.Fatalf("GetProposal client error: %v", err)
		}
		if clientResp.GetProposal().GetId() != 500 {
			t.Fatalf("expected id 500, got %d", clientResp.GetProposal().GetId())
		}

		freelancerResp, err := srv.GetProposal(authCtx("freelancer-token"), &proposalv1.GetProposalRequest{ProposalId: 501})
		if err != nil {
			t.Fatalf("GetProposal freelancer error: %v", err)
		}
		if freelancerResp.GetProposal().GetId() != 501 {
			t.Fatalf("expected id 501, got %d", freelancerResp.GetProposal().GetId())
		}

		_, err = srv.GetProposal(authCtx("bad-token"), &proposalv1.GetProposalRequest{ProposalId: 500})
		assertCode(t, err, codes.Unauthenticated)
	})

	t.Run("GetMyProposalForJob success and role denied", func(t *testing.T) {
		repo := &fakeProposalRepo{getLatestByJobForFreelancerFn: func(ctx context.Context, jobID int64, freelancerID uuid.UUID) (domain.Proposal, error) {
			return makeProposal(600, domain.StatusSent), nil
		}}

		srv := buildServer(repo, nil, nil, nil, nil, nil)
		resp, err := srv.GetMyProposalForJob(authCtx("freelancer-token"), &proposalv1.GetMyProposalForJobRequest{JobId: 1001})
		if err != nil {
			t.Fatalf("GetMyProposalForJob error: %v", err)
		}
		if resp.GetProposal().GetId() != 600 {
			t.Fatalf("expected id 600, got %d", resp.GetProposal().GetId())
		}

		_, err = srv.GetMyProposalForJob(authCtx("client-token"), &proposalv1.GetMyProposalForJobRequest{JobId: 1001})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("HasAppliedToJob supports found and not found", func(t *testing.T) {
		repo := &fakeProposalRepo{getLatestByJobForFreelancerFn: func(ctx context.Context, jobID int64, freelancerID uuid.UUID) (domain.Proposal, error) {
			if jobID == 2002 {
				return domain.Proposal{}, fmt.Errorf("proposal not found")
			}
			return makeProposal(700, domain.StatusShortlisted), nil
		}}
		srv := buildServer(repo, nil, nil, nil, nil, nil)

		yes, err := srv.HasAppliedToJob(authCtx("freelancer-token"), &proposalv1.HasAppliedToJobRequest{JobId: 1001})
		if err != nil {
			t.Fatalf("HasAppliedToJob found error: %v", err)
		}
		if !yes.GetHasApplied() || yes.GetProposalId() != 700 {
			t.Fatalf("unexpected has-applied response: %+v", yes)
		}

		no, err := srv.HasAppliedToJob(authCtx("freelancer-token"), &proposalv1.HasAppliedToJobRequest{JobId: 2002})
		if err != nil {
			t.Fatalf("HasAppliedToJob not-found error: %v", err)
		}
		if no.GetHasApplied() {
			t.Fatalf("expected has_applied=false")
		}
	})

	t.Run("ListProposalsByJob validates filter and maps payload", func(t *testing.T) {
		repo := &fakeProposalRepo{listByJobFn: func(ctx context.Context, filter application.ListByJobFilter, pageSize int, pageToken string) ([]domain.Proposal, string, error) {
			if filter.JobID != 1001 {
				t.Fatalf("expected job_id filter 1001")
			}
			if filter.FreelancerID == nil || *filter.FreelancerID != testFreelancerID {
				t.Fatalf("expected freelancer filter to be set")
			}
			if pageSize != 20 || pageToken != "" {
				t.Fatalf("unexpected pagination pageSize=%d pageToken=%q", pageSize, pageToken)
			}
			return []domain.Proposal{makeProposal(800, domain.StatusSent)}, "", nil
		}}
		srv := buildServer(repo, nil, nil, nil, nil, nil)

		resp, err := srv.ListProposalsByJob(authCtx("client-token"), &proposalv1.ListProposalsByJobRequest{
			JobId:              1001,
			StatusFilter:       []proposalv1.ProposalStatus{proposalv1.ProposalStatus_PROPOSAL_STATUS_SENT},
			FreelancerIdFilter: &[]string{testFreelancerID.String()}[0],
		})
		if err != nil {
			t.Fatalf("ListProposalsByJob error: %v", err)
		}
		if len(resp.GetProposals()) != 1 || resp.GetProposals()[0].GetId() != 800 {
			t.Fatalf("unexpected proposals payload: %+v", resp.GetProposals())
		}

		bad := "not-a-uuid"
		_, err = srv.ListProposalsByJob(authCtx("client-token"), &proposalv1.ListProposalsByJobRequest{JobId: 1001, FreelancerIdFilter: &bad})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("ListMyProposals supports pagination validation", func(t *testing.T) {
		repo := &fakeProposalRepo{listByFreelancerFn: func(ctx context.Context, filter application.ListByFreelancerFilter, pageSize int, pageToken string) ([]domain.Proposal, string, error) {
			if pageToken == "bad" {
				return nil, "", fmt.Errorf("invalid page_token")
			}
			if pageToken != "5" {
				t.Fatalf("expected page token 5, got %q", pageToken)
			}
			return []domain.Proposal{makeProposal(900, domain.StatusSent)}, "", nil
		}}
		srv := buildServer(repo, nil, nil, nil, nil, nil)

		resp, err := srv.ListMyProposals(authCtx("freelancer-token"), &proposalv1.ListMyProposalsRequest{PageToken: "5"})
		if err != nil {
			t.Fatalf("ListMyProposals error: %v", err)
		}
		if len(resp.GetProposals()) != 1 || resp.GetProposals()[0].GetId() != 900 {
			t.Fatalf("unexpected list my proposals response")
		}

		_, err = srv.ListMyProposals(authCtx("freelancer-token"), &proposalv1.ListMyProposalsRequest{PageToken: "bad"})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("ListClientProposals validates freelancer filter", func(t *testing.T) {
		repo := &fakeProposalRepo{listByClientFn: func(ctx context.Context, filter application.ListByClientFilter, pageSize int, pageToken string) ([]domain.Proposal, string, error) {
			if filter.FreelancerID == nil || *filter.FreelancerID != testFreelancerID {
				t.Fatalf("expected freelancer filter")
			}
			return []domain.Proposal{makeProposal(1000, domain.StatusShortlisted)}, "", nil
		}}
		srv := buildServer(repo, nil, nil, nil, nil, nil)

		resp, err := srv.ListClientProposals(authCtx("client-token"), &proposalv1.ListClientProposalsRequest{FreelancerIdFilter: &[]string{testFreelancerID.String()}[0]})
		if err != nil {
			t.Fatalf("ListClientProposals error: %v", err)
		}
		if len(resp.GetProposals()) != 1 || resp.GetProposals()[0].GetId() != 1000 {
			t.Fatalf("unexpected list client proposals response")
		}

		bad := "broken-uuid"
		_, err = srv.ListClientProposals(authCtx("client-token"), &proposalv1.ListClientProposalsRequest{FreelancerIdFilter: &bad})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("CountProposalsByJob returns totals", func(t *testing.T) {
		repo := &fakeProposalRepo{countByJobForClientFn: func(ctx context.Context, clientID uuid.UUID, jobID int64) (int64, map[string]int64, error) {
			return 6, map[string]int64{domain.StatusSent: 4, domain.StatusShortlisted: 2}, nil
		}}
		srv := buildServer(repo, nil, nil, nil, nil, nil)

		resp, err := srv.CountProposalsByJob(authCtx("client-token"), &proposalv1.CountProposalsByJobRequest{JobId: 1001})
		if err != nil {
			t.Fatalf("CountProposalsByJob error: %v", err)
		}
		if resp.GetTotal() != 6 || len(resp.GetByStatus()) != 2 {
			t.Fatalf("unexpected count by job response: %+v", resp)
		}
	})

	t.Run("CountClientProposalInbox supports status filters", func(t *testing.T) {
		repo := &fakeProposalRepo{countClientInboxFn: func(ctx context.Context, clientID uuid.UUID, statuses []string) (int64, map[string]int64, error) {
			if len(statuses) != 1 || statuses[0] != domain.StatusSent {
				t.Fatalf("unexpected status filters: %+v", statuses)
			}
			return 3, map[string]int64{domain.StatusSent: 3}, nil
		}}
		srv := buildServer(repo, nil, nil, nil, nil, nil)

		resp, err := srv.CountClientProposalInbox(authCtx("client-token"), &proposalv1.CountClientProposalInboxRequest{StatusFilter: []proposalv1.ProposalStatus{proposalv1.ProposalStatus_PROPOSAL_STATUS_SENT}})
		if err != nil {
			t.Fatalf("CountClientProposalInbox error: %v", err)
		}
		if resp.GetTotal() != 3 || len(resp.GetByStatus()) != 1 {
			t.Fatalf("unexpected inbox count response: %+v", resp)
		}
	})

	t.Run("GetProposalAttachmentUploadUrl success and validation", func(t *testing.T) {
		repo := &fakeProposalRepo{getByIDForFreelancerFn: func(ctx context.Context, proposalID int64, freelancerID uuid.UUID) (domain.Proposal, error) {
			return makeProposal(1200, domain.StatusSent), nil
		}}
		store := &fakeAttachmentStore{buildObjectKeyFn: func(proposalID int64, fileName string) string {
			return fmt.Sprintf("p/%d/%s", proposalID, fileName)
		}}
		srv := buildServer(repo, nil, nil, nil, nil, store)

		resp, err := srv.GetProposalAttachmentUploadUrl(authCtx("freelancer-token"), &proposalv1.GetProposalAttachmentUploadUrlRequest{ProposalId: 1200, FileName: "design.pdf", ContentType: "application/pdf"})
		if err != nil {
			t.Fatalf("GetProposalAttachmentUploadUrl error: %v", err)
		}
		if resp.GetStorageKey() != "p/1200/design.pdf" {
			t.Fatalf("unexpected storage key: %s", resp.GetStorageKey())
		}

		_, err = srv.GetProposalAttachmentUploadUrl(authCtx("freelancer-token"), &proposalv1.GetProposalAttachmentUploadUrlRequest{ProposalId: 1200, FileName: "design.pdf"})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("GetProposalAttachmentDownloadUrl enforces role and returns presigned URL", func(t *testing.T) {
		repo := &fakeProposalRepo{getByIDForClientFn: func(ctx context.Context, proposalID int64, clientID uuid.UUID) (domain.Proposal, error) {
			p := makeProposal(1300, domain.StatusSent)
			p.Attachments = []domain.Attachment{{ID: 22, StorageKey: "p/1300/design.pdf"}}
			return p, nil
		}}
		store := &fakeAttachmentStore{presignGetObjectFn: func(ctx context.Context, storageKey string, ttl time.Duration) (string, error) {
			if storageKey != "p/1300/design.pdf" {
				t.Fatalf("unexpected storage key for download: %s", storageKey)
			}
			return "https://download.local/p/1300/design.pdf", nil
		}}
		srv := buildServer(repo, nil, nil, nil, nil, store)

		resp, err := srv.GetProposalAttachmentDownloadUrl(authCtx("client-token"), &proposalv1.GetProposalAttachmentDownloadUrlRequest{ProposalId: 1300, AttachmentId: 22})
		if err != nil {
			t.Fatalf("GetProposalAttachmentDownloadUrl error: %v", err)
		}
		if resp.GetDownloadUrl() == "" {
			t.Fatalf("expected non-empty download url")
		}

		_, err = srv.GetProposalAttachmentDownloadUrl(authCtx("bad-token"), &proposalv1.GetProposalAttachmentDownloadUrlRequest{ProposalId: 1300, AttachmentId: 22})
		assertCode(t, err, codes.Unauthenticated)
	})

	t.Run("SetProposalStatus supports client decisions", func(t *testing.T) {
		repo := &fakeProposalRepo{}
		call := 0
		repo.getByIDForClientFn = func(ctx context.Context, proposalID int64, clientID uuid.UUID) (domain.Proposal, error) {
			call++
			if call == 1 {
				return makeProposal(1400, domain.StatusSent), nil
			}
			return makeProposal(1400, domain.StatusShortlisted), nil
		}
		repo.setStatusFn = func(ctx context.Context, proposalID int64, clientID uuid.UUID, status string, reason string, at time.Time) error {
			if status != domain.StatusShortlisted {
				t.Fatalf("expected shortlist status, got %s", status)
			}
			return nil
		}
		srv := buildServer(repo, nil, nil, nil, nil, nil)

		resp, err := srv.SetProposalStatus(authCtx("client-token"), &proposalv1.SetProposalStatusRequest{ProposalId: 1400, Decision: proposalv1.ClientDecision_CLIENT_DECISION_SHORTLISTED, Reason: "good fit"})
		if err != nil {
			t.Fatalf("SetProposalStatus error: %v", err)
		}
		if resp.GetProposal().GetStatus() != proposalv1.ProposalStatus_PROPOSAL_STATUS_SHORTLISTED {
			t.Fatalf("expected shortlisted status, got %v", resp.GetProposal().GetStatus())
		}

		_, err = srv.SetProposalStatus(authCtx("client-token"), &proposalv1.SetProposalStatusRequest{ProposalId: 1400, Decision: proposalv1.ClientDecision_CLIENT_DECISION_UNSPECIFIED})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("InternalMarkProposalOfferSent requires contract service caller", func(t *testing.T) {
		repo := &fakeProposalRepo{}
		repo.markOfferSentFn = func(ctx context.Context, proposalID int64, clientID uuid.UUID, reason string, at time.Time) (domain.Proposal, error) {
			p := makeProposal(proposalID, domain.StatusOfferSent)
			p.ClientID = clientID
			return p, nil
		}

		srv := buildServer(repo, nil, nil, nil, nil, nil)

		_, err := srv.InternalMarkProposalOfferSent(authCtx("client-token"), &proposalv1.InternalMarkProposalOfferSentRequest{
			ProposalId: 1440,
			ClientId:   testClientID.String(),
			Note:       "offer sent",
		})
		assertCode(t, err, codes.PermissionDenied)

		internalMD := metadata.Pairs("authorization", "Bearer client-token", "x-jobconnect-internal", "contract-service")
		internalCtx := metadata.NewIncomingContext(context.Background(), internalMD)

		resp, err := srv.InternalMarkProposalOfferSent(internalCtx, &proposalv1.InternalMarkProposalOfferSentRequest{
			ProposalId: 1440,
			ClientId:   testClientID.String(),
			Note:       "offer sent",
		})
		if err != nil {
			t.Fatalf("InternalMarkProposalOfferSent error: %v", err)
		}
		if got := resp.GetProposal().GetStatus(); got != proposalv1.ProposalStatus_PROPOSAL_STATUS_OFFER_SENT {
			t.Fatalf("expected offer_sent status, got %v", got)
		}
	})

	t.Run("InternalHireProposal requires internal marker and supports idempotency", func(t *testing.T) {
		repo := &fakeProposalRepo{}
		repo.getByIDForClientFn = func(ctx context.Context, proposalID int64, clientID uuid.UUID) (domain.Proposal, error) {
			return makeProposal(1450, domain.StatusOfferSent), nil
		}
		repo.hireWithRequestIDFn = func(ctx context.Context, proposalID int64, clientID uuid.UUID, requestID string, reason string, at time.Time) (domain.Proposal, bool, error) {
			return makeProposal(1450, domain.StatusHired), requestID == "dup", nil
		}

		srv := buildServer(repo, nil, nil, nil, nil, nil)

		_, err := srv.InternalHireProposal(authCtx("client-token"), &proposalv1.InternalHireProposalRequest{
			ProposalId: 1450,
			ClientId:   testClientID.String(),
			RequestId:  "req-1",
			Note:       "hire",
		})
		assertCode(t, err, codes.PermissionDenied)

		internalMD := metadata.Pairs("authorization", "Bearer client-token", "x-jobconnect-internal", "job-service")
		internalCtx := metadata.NewIncomingContext(context.Background(), internalMD)

		okResp, err := srv.InternalHireProposal(internalCtx, &proposalv1.InternalHireProposalRequest{
			ProposalId: 1450,
			ClientId:   testClientID.String(),
			RequestId:  "req-1",
			Note:       "hire",
		})
		if err != nil {
			t.Fatalf("InternalHireProposal error: %v", err)
		}
		if okResp.GetReusedIdempotentResult() {
			t.Fatalf("expected non-idempotent fresh response")
		}

		dupResp, err := srv.InternalHireProposal(internalCtx, &proposalv1.InternalHireProposalRequest{
			ProposalId: 1450,
			ClientId:   testClientID.String(),
			RequestId:  "dup",
			Note:       "retry",
		})
		if err != nil {
			t.Fatalf("InternalHireProposal duplicate error: %v", err)
		}
		if !dupResp.GetReusedIdempotentResult() {
			t.Fatalf("expected reused idempotent result")
		}
	})

	t.Run("InternalReleaseHiredProposal accepts contract service caller", func(t *testing.T) {
		repo := &fakeProposalRepo{}
		call := 0
		repo.getByIDForClientFn = func(ctx context.Context, proposalID int64, clientID uuid.UUID) (domain.Proposal, error) {
			call++
			if call == 1 {
				return makeProposal(1460, domain.StatusHired), nil
			}
			return makeProposal(1460, domain.StatusShortlisted), nil
		}
		repo.revertHireFn = func(ctx context.Context, proposalID int64, clientID uuid.UUID, reason string, at time.Time) error {
			if proposalID != 1460 {
				t.Fatalf("unexpected proposal id: %d", proposalID)
			}
			if clientID != testClientID {
				t.Fatalf("unexpected client id: %s", clientID)
			}
			if reason != "offer revoked" {
				t.Fatalf("unexpected reason: %q", reason)
			}
			return nil
		}

		srv := buildServer(repo, nil, nil, nil, nil, nil)

		internalMD := metadata.Pairs("authorization", "Bearer client-token", "x-jobconnect-internal", "contract-service")
		internalCtx := metadata.NewIncomingContext(context.Background(), internalMD)
		resp, err := srv.InternalReleaseHiredProposal(internalCtx, &proposalv1.InternalReleaseHiredProposalRequest{
			ProposalId: 1460,
			ClientId:   testClientID.String(),
			Reason:     "offer revoked",
		})
		if err != nil {
			t.Fatalf("InternalReleaseHiredProposal error: %v", err)
		}
		if got := resp.GetProposal().GetStatus(); got != proposalv1.ProposalStatus_PROPOSAL_STATUS_SHORTLISTED {
			t.Fatalf("expected shortlisted status, got %v", got)
		}
	})

	t.Run("all RPCs reject nil request", func(t *testing.T) {
		repo := &fakeProposalRepo{}
		srv := buildServer(repo, nil, nil, nil, nil, nil)

		calls := []struct {
			name string
			fn   func() error
		}{
			{name: "SubmitProposal", fn: func() error { _, err := srv.SubmitProposal(authCtx("freelancer-token"), nil); return err }},
			{name: "ModifyProposal", fn: func() error { _, err := srv.ModifyProposal(authCtx("freelancer-token"), nil); return err }},
			{name: "WithdrawProposal", fn: func() error { _, err := srv.WithdrawProposal(authCtx("freelancer-token"), nil); return err }},
			{name: "GetProposal", fn: func() error { _, err := srv.GetProposal(authCtx("client-token"), nil); return err }},
			{name: "GetMyProposalForJob", fn: func() error { _, err := srv.GetMyProposalForJob(authCtx("freelancer-token"), nil); return err }},
			{name: "HasAppliedToJob", fn: func() error { _, err := srv.HasAppliedToJob(authCtx("freelancer-token"), nil); return err }},
			{name: "ListProposalsByJob", fn: func() error { _, err := srv.ListProposalsByJob(authCtx("client-token"), nil); return err }},
			{name: "ListMyProposals", fn: func() error { _, err := srv.ListMyProposals(authCtx("freelancer-token"), nil); return err }},
			{name: "ListClientProposals", fn: func() error { _, err := srv.ListClientProposals(authCtx("client-token"), nil); return err }},
			{name: "CountProposalsByJob", fn: func() error { _, err := srv.CountProposalsByJob(authCtx("client-token"), nil); return err }},
			{name: "CountClientProposalInbox", fn: func() error { _, err := srv.CountClientProposalInbox(authCtx("client-token"), nil); return err }},
			{name: "InternalMarkProposalOfferSent", fn: func() error { _, err := srv.InternalMarkProposalOfferSent(authCtx("client-token"), nil); return err }},
			{name: "InternalHireProposal", fn: func() error { _, err := srv.InternalHireProposal(authCtx("client-token"), nil); return err }},
			{name: "InternalReleaseHiredProposal", fn: func() error { _, err := srv.InternalReleaseHiredProposal(authCtx("client-token"), nil); return err }},
			{name: "GetProposalAttachmentUploadUrl", fn: func() error {
				_, err := srv.GetProposalAttachmentUploadUrl(authCtx("freelancer-token"), nil)
				return err
			}},
			{name: "GetProposalAttachmentDownloadUrl", fn: func() error { _, err := srv.GetProposalAttachmentDownloadUrl(authCtx("client-token"), nil); return err }},
			{name: "SetProposalStatus", fn: func() error { _, err := srv.SetProposalStatus(authCtx("client-token"), nil); return err }},
		}

		for _, tc := range calls {
			err := tc.fn()
			if err == nil {
				t.Fatalf("%s: expected error", tc.name)
			}
			st, ok := status.FromError(err)
			if !ok || st.Code() != codes.InvalidArgument {
				t.Fatalf("%s: expected InvalidArgument, got %v", tc.name, err)
			}
		}
	})
}
