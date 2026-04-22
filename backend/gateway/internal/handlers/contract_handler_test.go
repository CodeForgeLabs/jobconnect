package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	contractv1 "jobconnect/contract/gen/contract/v1"
	jobv1 "jobconnect/job/gen/job/v1"
	proposalv1 "jobconnect/proposal/gen/proposal/v1"

	"google.golang.org/grpc"
)

type contractHandlerProposalStub struct {
	proposalv1.UnimplementedProposalServiceServer
	response *proposalv1.GetProposalResponse
	lastReq  *proposalv1.GetProposalRequest
}

func (s *contractHandlerProposalStub) GetProposal(ctx context.Context, in *proposalv1.GetProposalRequest, opts ...grpc.CallOption) (*proposalv1.GetProposalResponse, error) {
	s.lastReq = in
	if s.response != nil {
		return s.response, nil
	}
	return &proposalv1.GetProposalResponse{Proposal: &proposalv1.Proposal{Id: in.GetProposalId(), JobId: 21, ClientId: "client-1", FreelancerId: "freelancer-1", Status: proposalv1.ProposalStatus_PROPOSAL_STATUS_SENT}}, nil
}

func (s *contractHandlerProposalStub) GetMyProposalForJob(context.Context, *proposalv1.GetMyProposalForJobRequest, ...grpc.CallOption) (*proposalv1.GetMyProposalForJobResponse, error) {
	return nil, nil
}
func (s *contractHandlerProposalStub) ListMyProposals(context.Context, *proposalv1.ListMyProposalsRequest, ...grpc.CallOption) (*proposalv1.ListMyProposalsResponse, error) {
	return nil, nil
}
func (s *contractHandlerProposalStub) ListClientProposals(context.Context, *proposalv1.ListClientProposalsRequest, ...grpc.CallOption) (*proposalv1.ListClientProposalsResponse, error) {
	return nil, nil
}
func (s *contractHandlerProposalStub) HasAppliedToJob(context.Context, *proposalv1.HasAppliedToJobRequest, ...grpc.CallOption) (*proposalv1.HasAppliedToJobResponse, error) {
	return nil, nil
}
func (s *contractHandlerProposalStub) CountProposalsByJob(context.Context, *proposalv1.CountProposalsByJobRequest, ...grpc.CallOption) (*proposalv1.CountProposalsByJobResponse, error) {
	return nil, nil
}
func (s *contractHandlerProposalStub) CountClientProposalInbox(context.Context, *proposalv1.CountClientProposalInboxRequest, ...grpc.CallOption) (*proposalv1.CountClientProposalInboxResponse, error) {
	return nil, nil
}
func (s *contractHandlerProposalStub) SetProposalStatus(context.Context, *proposalv1.SetProposalStatusRequest, ...grpc.CallOption) (*proposalv1.SetProposalStatusResponse, error) {
	return nil, nil
}
func (s *contractHandlerProposalStub) GetProposalAttachmentUploadUrl(context.Context, *proposalv1.GetProposalAttachmentUploadUrlRequest, ...grpc.CallOption) (*proposalv1.GetProposalAttachmentUploadUrlResponse, error) {
	return nil, nil
}
func (s *contractHandlerProposalStub) GetProposalAttachmentDownloadUrl(context.Context, *proposalv1.GetProposalAttachmentDownloadUrlRequest, ...grpc.CallOption) (*proposalv1.GetProposalAttachmentDownloadUrlResponse, error) {
	return nil, nil
}

type contractHandlerJobStub struct {
	jobv1.UnimplementedJobServiceServer
	response *jobv1.GetJobSummaryResponse
	lastReq  *jobv1.GetJobSummaryRequest
}

func (s *contractHandlerJobStub) GetJobSummary(ctx context.Context, in *jobv1.GetJobSummaryRequest, opts ...grpc.CallOption) (*jobv1.GetJobSummaryResponse, error) {
	s.lastReq = in
	if s.response != nil {
		return s.response, nil
	}
	return &jobv1.GetJobSummaryResponse{Summary: &jobv1.JobSummary{JobId: in.GetJobId(), ClientId: "client-1", IsOpen: true, Found: true}}, nil
}
func (s *contractHandlerJobStub) CreateJob(context.Context, *jobv1.CreateJobRequest, ...grpc.CallOption) (*jobv1.CreateJobResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) GetJob(context.Context, *jobv1.GetJobRequest, ...grpc.CallOption) (*jobv1.GetJobResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) UpdateJob(context.Context, *jobv1.UpdateJobRequest, ...grpc.CallOption) (*jobv1.UpdateJobResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) ListMyJobs(context.Context, *jobv1.ListMyJobsRequest, ...grpc.CallOption) (*jobv1.ListMyJobsResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) ListOpenJobs(context.Context, *jobv1.ListOpenJobsRequest, ...grpc.CallOption) (*jobv1.ListOpenJobsResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) CloseJob(context.Context, *jobv1.CloseJobRequest, ...grpc.CallOption) (*jobv1.CloseJobResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) UploadJobAttachment(context.Context, *jobv1.UploadJobAttachmentRequest, ...grpc.CallOption) (*jobv1.UploadJobAttachmentResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) DeleteJobAttachment(context.Context, *jobv1.DeleteJobAttachmentRequest, ...grpc.CallOption) (*jobv1.DeleteJobAttachmentResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) SetJobVisibility(context.Context, *jobv1.SetJobVisibilityRequest, ...grpc.CallOption) (*jobv1.SetJobVisibilityResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) SetJobBudgetRange(context.Context, *jobv1.SetJobBudgetRangeRequest, ...grpc.CallOption) (*jobv1.SetJobBudgetRangeResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) InviteFreelancerToJob(context.Context, *jobv1.InviteFreelancerToJobRequest, ...grpc.CallOption) (*jobv1.InviteFreelancerToJobResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) ListJobApplicants(context.Context, *jobv1.ListJobApplicantsRequest, ...grpc.CallOption) (*jobv1.ListJobApplicantsResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) SetApplicantStage(context.Context, *jobv1.SetApplicantStageRequest, ...grpc.CallOption) (*jobv1.SetApplicantStageResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) PauseJob(context.Context, *jobv1.PauseJobRequest, ...grpc.CallOption) (*jobv1.PauseJobResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) ReopenJob(context.Context, *jobv1.ReopenJobRequest, ...grpc.CallOption) (*jobv1.ReopenJobResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) MarkJobFilled(context.Context, *jobv1.MarkJobFilledRequest, ...grpc.CallOption) (*jobv1.MarkJobFilledResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) SearchJobs(context.Context, *jobv1.SearchJobsRequest, ...grpc.CallOption) (*jobv1.SearchJobsResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) ListJobFacets(context.Context, *jobv1.ListJobFacetsRequest, ...grpc.CallOption) (*jobv1.ListJobFacetsResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) ListJobAttachments(context.Context, *jobv1.ListJobAttachmentsRequest, ...grpc.CallOption) (*jobv1.ListJobAttachmentsResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) GetJobAttachmentDownloadUrl(context.Context, *jobv1.GetJobAttachmentDownloadUrlRequest, ...grpc.CallOption) (*jobv1.GetJobAttachmentDownloadUrlResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) GetPublicJobDetail(context.Context, *jobv1.GetPublicJobDetailRequest, ...grpc.CallOption) (*jobv1.GetPublicJobDetailResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) ListInvitedJobs(context.Context, *jobv1.ListInvitedJobsRequest, ...grpc.CallOption) (*jobv1.ListInvitedJobsResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) RespondToJobInvite(context.Context, *jobv1.RespondToJobInviteRequest, ...grpc.CallOption) (*jobv1.RespondToJobInviteResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) SaveJob(context.Context, *jobv1.SaveJobRequest, ...grpc.CallOption) (*jobv1.SaveJobResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) UnsaveJob(context.Context, *jobv1.UnsaveJobRequest, ...grpc.CallOption) (*jobv1.UnsaveJobResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) ListSavedJobs(context.Context, *jobv1.ListSavedJobsRequest, ...grpc.CallOption) (*jobv1.ListSavedJobsResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) RejectAllApplicants(context.Context, *jobv1.RejectAllApplicantsRequest, ...grpc.CallOption) (*jobv1.RejectAllApplicantsResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) ReopenHiringForJob(context.Context, *jobv1.ReopenHiringForJobRequest, ...grpc.CallOption) (*jobv1.ReopenHiringForJobResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) GetJobStats(context.Context, *jobv1.GetJobStatsRequest, ...grpc.CallOption) (*jobv1.GetJobStatsResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) SearchJobsV2(context.Context, *jobv1.SearchJobsV2Request, ...grpc.CallOption) (*jobv1.SearchJobsV2Response, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) MarkJobCompleted(context.Context, *jobv1.MarkJobCompletedRequest, ...grpc.CallOption) (*jobv1.MarkJobCompletedResponse, error) {
	return nil, nil
}
func (s *contractHandlerJobStub) CancelJobWithSettlementPolicy(context.Context, *jobv1.CancelJobWithSettlementPolicyRequest, ...grpc.CallOption) (*jobv1.CancelJobWithSettlementPolicyResponse, error) {
	return nil, nil
}

type contractHandlerContractStub struct {
	contractv1.UnimplementedContractServiceServer
	contracts []*contractv1.Contract
	lastReq   *contractv1.ListMyContractsRequest
}

func (s *contractHandlerContractStub) ListMyContracts(ctx context.Context, in *contractv1.ListMyContractsRequest, opts ...grpc.CallOption) (*contractv1.ListMyContractsResponse, error) {
	s.lastReq = in
	return &contractv1.ListMyContractsResponse{Contracts: s.contracts}, nil
}
func (s *contractHandlerContractStub) GetContract(context.Context, *contractv1.GetContractRequest, ...grpc.CallOption) (*contractv1.GetContractResponse, error) {
	return nil, nil
}
func (s *contractHandlerContractStub) CreateContract(context.Context, *contractv1.CreateContractRequest, ...grpc.CallOption) (*contractv1.CreateContractResponse, error) {
	return nil, nil
}
func (s *contractHandlerContractStub) GetJobOfferState(context.Context, *contractv1.GetJobOfferStateRequest, ...grpc.CallOption) (*contractv1.GetJobOfferStateResponse, error) {
	return nil, nil
}
func (s *contractHandlerContractStub) AcceptContract(context.Context, *contractv1.AcceptContractRequest, ...grpc.CallOption) (*contractv1.AcceptContractResponse, error) {
	return nil, nil
}
func (s *contractHandlerContractStub) DeclineContract(context.Context, *contractv1.DeclineContractRequest, ...grpc.CallOption) (*contractv1.DeclineContractResponse, error) {
	return nil, nil
}
func (s *contractHandlerContractStub) RevokeContractOffer(context.Context, *contractv1.RevokeContractOfferRequest, ...grpc.CallOption) (*contractv1.RevokeContractOfferResponse, error) {
	return nil, nil
}
func (s *contractHandlerContractStub) SubmitMilestoneWork(context.Context, *contractv1.SubmitMilestoneWorkRequest, ...grpc.CallOption) (*contractv1.SubmitMilestoneWorkResponse, error) {
	return nil, nil
}
func (s *contractHandlerContractStub) RequestMilestoneChanges(context.Context, *contractv1.RequestMilestoneChangesRequest, ...grpc.CallOption) (*contractv1.RequestMilestoneChangesResponse, error) {
	return nil, nil
}
func (s *contractHandlerContractStub) ApproveMilestoneSubmission(context.Context, *contractv1.ApproveMilestoneSubmissionRequest, ...grpc.CallOption) (*contractv1.ApproveMilestoneSubmissionResponse, error) {
	return nil, nil
}
func (s *contractHandlerContractStub) UpdateMilestoneStatus(context.Context, *contractv1.UpdateMilestoneStatusRequest, ...grpc.CallOption) (*contractv1.UpdateMilestoneStatusResponse, error) {
	return nil, nil
}
func (s *contractHandlerContractStub) LogHourlyWork(context.Context, *contractv1.LogHourlyWorkRequest, ...grpc.CallOption) (*contractv1.LogHourlyWorkResponse, error) {
	return nil, nil
}
func (s *contractHandlerContractStub) ListHourlyLogs(context.Context, *contractv1.ListHourlyLogsRequest, ...grpc.CallOption) (*contractv1.ListHourlyLogsResponse, error) {
	return nil, nil
}
func (s *contractHandlerContractStub) ReviewHourlyLog(context.Context, *contractv1.ReviewHourlyLogRequest, ...grpc.CallOption) (*contractv1.ReviewHourlyLogResponse, error) {
	return nil, nil
}
func (s *contractHandlerContractStub) ProposeAmendment(context.Context, *contractv1.ProposeAmendmentRequest, ...grpc.CallOption) (*contractv1.ProposeAmendmentResponse, error) {
	return nil, nil
}
func (s *contractHandlerContractStub) RespondAmendment(context.Context, *contractv1.RespondAmendmentRequest, ...grpc.CallOption) (*contractv1.RespondAmendmentResponse, error) {
	return nil, nil
}
func (s *contractHandlerContractStub) ListAmendments(context.Context, *contractv1.ListAmendmentsRequest, ...grpc.CallOption) (*contractv1.ListAmendmentsResponse, error) {
	return nil, nil
}
func (s *contractHandlerContractStub) PauseContract(context.Context, *contractv1.PauseContractRequest, ...grpc.CallOption) (*contractv1.PauseContractResponse, error) {
	return nil, nil
}
func (s *contractHandlerContractStub) ResumeContract(context.Context, *contractv1.ResumeContractRequest, ...grpc.CallOption) (*contractv1.ResumeContractResponse, error) {
	return nil, nil
}
func (s *contractHandlerContractStub) EndContract(context.Context, *contractv1.EndContractRequest, ...grpc.CallOption) (*contractv1.EndContractResponse, error) {
	return nil, nil
}
func (s *contractHandlerContractStub) GetStatusHistory(context.Context, *contractv1.GetStatusHistoryRequest, ...grpc.CallOption) (*contractv1.GetStatusHistoryResponse, error) {
	return nil, nil
}

func TestContractBootstrap_ReturnsRawData(t *testing.T) {
	proposalClient := &contractHandlerProposalStub{}
	jobClient := &contractHandlerJobStub{}
	contractClient := &contractHandlerContractStub{contracts: []*contractv1.Contract{{Id: 77, JobId: 21, ProposalId: 44, ClientId: "client-1", FreelancerId: "freelancer-1", Status: contractv1.ContractStatus_CONTRACT_STATUS_PENDING_ACCEPTANCE}}}
	h := NewContractHandler(contractClient, jobClient, proposalClient)

	ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/contracts/bootstrap?job_id=21&proposal_id=44")
	h.Bootstrap(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if proposalClient.lastReq == nil || proposalClient.lastReq.GetProposalId() != 44 {
		t.Fatalf("expected proposal lookup for 44, got %#v", proposalClient.lastReq)
	}
	if jobClient.lastReq == nil || jobClient.lastReq.GetJobId() != 21 {
		t.Fatalf("expected job summary lookup for 21, got %#v", jobClient.lastReq)
	}
	if contractClient.lastReq == nil {
		t.Fatal("expected contract list call")
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["proposal"] == nil || body["job_summary"] == nil || body["offer_state"] == nil {
		t.Fatalf("expected proposal, job_summary, and offer_state in response, got %#v", body)
	}
	if body["contract"] == nil {
		t.Fatalf("expected contract payload in response, got %#v", body)
	}
}

func TestContractBootstrap_RejectsMismatchedProposalJob(t *testing.T) {
	h := NewContractHandler(&contractHandlerContractStub{}, &contractHandlerJobStub{}, &contractHandlerProposalStub{response: &proposalv1.GetProposalResponse{Proposal: &proposalv1.Proposal{Id: 44, JobId: 99, ClientId: "client-1", FreelancerId: "freelancer-1", Status: proposalv1.ProposalStatus_PROPOSAL_STATUS_SENT}}})

	ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/contracts/bootstrap?job_id=21&proposal_id=44")
	h.Bootstrap(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestContractBootstrap_MissingJobID_ReturnsBadRequest(t *testing.T) {
	h := NewContractHandler(&contractHandlerContractStub{}, &contractHandlerJobStub{}, &contractHandlerProposalStub{})

	ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/contracts/bootstrap?proposal_id=44")
	h.Bootstrap(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestContractBootstrap_BlocksWhenAnotherOfferExistsOnJob(t *testing.T) {
	proposalClient := &contractHandlerProposalStub{}
	jobClient := &contractHandlerJobStub{}
	contractClient := &contractHandlerContractStub{contracts: []*contractv1.Contract{
		{Id: 77, JobId: 21, ProposalId: 44, ClientId: "client-1", FreelancerId: "freelancer-1", Status: contractv1.ContractStatus_CONTRACT_STATUS_PENDING_ACCEPTANCE},
		{Id: 78, JobId: 21, ProposalId: 45, ClientId: "client-1", FreelancerId: "freelancer-2", Status: contractv1.ContractStatus_CONTRACT_STATUS_PENDING_ACCEPTANCE},
	}}
	h := NewContractHandler(contractClient, jobClient, proposalClient)

	ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/contracts/bootstrap?job_id=21&proposal_id=46")
	h.Bootstrap(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	offerState := body["offer_state"].(map[string]any)
	if canOpen, ok := offerState["can_open_offer_form"].(bool); !ok || canOpen {
		t.Fatalf("expected bootstrap to block opening the form, got %#v", offerState)
	}
	if reason, _ := offerState["blocking_reason"].(string); reason != "pending_offer_exists" {
		t.Fatalf("expected pending_offer_exists reason, got %#v", offerState)
	}
}

func TestContractBootstrap_BlocksWhenSameProposalAlreadyHasOffer(t *testing.T) {
	proposalClient := &contractHandlerProposalStub{}
	jobClient := &contractHandlerJobStub{}
	contractClient := &contractHandlerContractStub{contracts: []*contractv1.Contract{
		{Id: 77, JobId: 21, ProposalId: 44, ClientId: "client-1", FreelancerId: "freelancer-1", Status: contractv1.ContractStatus_CONTRACT_STATUS_PENDING_ACCEPTANCE},
	}}
	h := NewContractHandler(contractClient, jobClient, proposalClient)

	ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/contracts/bootstrap?job_id=21&proposal_id=44")
	h.Bootstrap(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	offerState := body["offer_state"].(map[string]any)
	if canOpen, ok := offerState["can_open_offer_form"].(bool); !ok || canOpen {
		t.Fatalf("expected bootstrap to block opening the form, got %#v", offerState)
	}
	if reason, _ := offerState["blocking_reason"].(string); reason != "offer_already_sent" {
		t.Fatalf("expected offer_already_sent reason, got %#v", offerState)
	}
}

func TestContractBootstrap_BlocksWhenProposalIsNotEligible(t *testing.T) {
	proposalClient := &contractHandlerProposalStub{response: &proposalv1.GetProposalResponse{Proposal: &proposalv1.Proposal{Id: 44, JobId: 21, ClientId: "client-1", FreelancerId: "freelancer-1", Status: proposalv1.ProposalStatus_PROPOSAL_STATUS_REJECTED}}}
	jobClient := &contractHandlerJobStub{}
	contractClient := &contractHandlerContractStub{}
	h := NewContractHandler(contractClient, jobClient, proposalClient)

	ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/contracts/bootstrap?job_id=21&proposal_id=44")
	h.Bootstrap(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	offerState := body["offer_state"].(map[string]any)
	if canOpen, ok := offerState["can_open_offer_form"].(bool); !ok || canOpen {
		t.Fatalf("expected bootstrap to block opening the form, got %#v", offerState)
	}
	if reason, _ := offerState["blocking_reason"].(string); reason != "proposal_not_eligible" {
		t.Fatalf("expected proposal_not_eligible reason, got %#v", offerState)
	}
}
