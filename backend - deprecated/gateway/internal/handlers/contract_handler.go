package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	contractv1 "jobconnect/contract/gen/contract/v1"
	jobv1 "jobconnect/job/gen/job/v1"
	proposalv1 "jobconnect/proposal/gen/proposal/v1"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type contractProposalReader interface {
	GetProposal(ctx context.Context, in *proposalv1.GetProposalRequest, opts ...grpc.CallOption) (*proposalv1.GetProposalResponse, error)
}

type contractJobReader interface {
	GetJobSummary(ctx context.Context, in *jobv1.GetJobSummaryRequest, opts ...grpc.CallOption) (*jobv1.GetJobSummaryResponse, error)
}

type contractListReader interface {
	ListMyContracts(ctx context.Context, in *contractv1.ListMyContractsRequest, opts ...grpc.CallOption) (*contractv1.ListMyContractsResponse, error)
}

type contractCreateReader interface {
	CreateContract(ctx context.Context, in *contractv1.CreateContractRequest, opts ...grpc.CallOption) (*contractv1.CreateContractResponse, error)
	GetContract(ctx context.Context, in *contractv1.GetContractRequest, opts ...grpc.CallOption) (*contractv1.GetContractResponse, error)
	ListMyContracts(ctx context.Context, in *contractv1.ListMyContractsRequest, opts ...grpc.CallOption) (*contractv1.ListMyContractsResponse, error)
	AcceptContract(ctx context.Context, in *contractv1.AcceptContractRequest, opts ...grpc.CallOption) (*contractv1.AcceptContractResponse, error)
	DeclineContract(ctx context.Context, in *contractv1.DeclineContractRequest, opts ...grpc.CallOption) (*contractv1.DeclineContractResponse, error)
	RevokeContractOffer(ctx context.Context, in *contractv1.RevokeContractOfferRequest, opts ...grpc.CallOption) (*contractv1.RevokeContractOfferResponse, error)
	SubmitMilestoneWork(ctx context.Context, in *contractv1.SubmitMilestoneWorkRequest, opts ...grpc.CallOption) (*contractv1.SubmitMilestoneWorkResponse, error)
	RequestMilestoneChanges(ctx context.Context, in *contractv1.RequestMilestoneChangesRequest, opts ...grpc.CallOption) (*contractv1.RequestMilestoneChangesResponse, error)
	ApproveMilestoneSubmission(ctx context.Context, in *contractv1.ApproveMilestoneSubmissionRequest, opts ...grpc.CallOption) (*contractv1.ApproveMilestoneSubmissionResponse, error)
	LogHourlyWork(ctx context.Context, in *contractv1.LogHourlyWorkRequest, opts ...grpc.CallOption) (*contractv1.LogHourlyWorkResponse, error)
	GetHourlyLogEvidenceUploadUrl(ctx context.Context, in *contractv1.GetHourlyLogEvidenceUploadUrlRequest, opts ...grpc.CallOption) (*contractv1.GetHourlyLogEvidenceUploadUrlResponse, error)
	ListHourlyLogs(ctx context.Context, in *contractv1.ListHourlyLogsRequest, opts ...grpc.CallOption) (*contractv1.ListHourlyLogsResponse, error)
	GetHourlyWorkSummary(ctx context.Context, in *contractv1.GetHourlyWorkSummaryRequest, opts ...grpc.CallOption) (*contractv1.GetHourlyWorkSummaryResponse, error)
	UpdateHourlyLog(ctx context.Context, in *contractv1.UpdateHourlyLogRequest, opts ...grpc.CallOption) (*contractv1.UpdateHourlyLogResponse, error)
	DeleteHourlyLog(ctx context.Context, in *contractv1.DeleteHourlyLogRequest, opts ...grpc.CallOption) (*contractv1.DeleteHourlyLogResponse, error)
	ReviewHourlyLog(ctx context.Context, in *contractv1.ReviewHourlyLogRequest, opts ...grpc.CallOption) (*contractv1.ReviewHourlyLogResponse, error)
	GetHourlyInvoice(ctx context.Context, in *contractv1.GetHourlyInvoiceRequest, opts ...grpc.CallOption) (*contractv1.GetHourlyInvoiceResponse, error)
	ListHourlyInvoices(ctx context.Context, in *contractv1.ListHourlyInvoicesRequest, opts ...grpc.CallOption) (*contractv1.ListHourlyInvoicesResponse, error)
	CreateContractBonus(ctx context.Context, in *contractv1.CreateContractBonusRequest, opts ...grpc.CallOption) (*contractv1.CreateContractBonusResponse, error)
	ListContractBonuses(ctx context.Context, in *contractv1.ListContractBonusesRequest, opts ...grpc.CallOption) (*contractv1.ListContractBonusesResponse, error)
	ProposeAmendment(ctx context.Context, in *contractv1.ProposeAmendmentRequest, opts ...grpc.CallOption) (*contractv1.ProposeAmendmentResponse, error)
	RespondAmendment(ctx context.Context, in *contractv1.RespondAmendmentRequest, opts ...grpc.CallOption) (*contractv1.RespondAmendmentResponse, error)
	ListAmendments(ctx context.Context, in *contractv1.ListAmendmentsRequest, opts ...grpc.CallOption) (*contractv1.ListAmendmentsResponse, error)
	PauseContract(ctx context.Context, in *contractv1.PauseContractRequest, opts ...grpc.CallOption) (*contractv1.PauseContractResponse, error)
	ResumeContract(ctx context.Context, in *contractv1.ResumeContractRequest, opts ...grpc.CallOption) (*contractv1.ResumeContractResponse, error)
	EndContract(ctx context.Context, in *contractv1.EndContractRequest, opts ...grpc.CallOption) (*contractv1.EndContractResponse, error)
	GetStatusHistory(ctx context.Context, in *contractv1.GetStatusHistoryRequest, opts ...grpc.CallOption) (*contractv1.GetStatusHistoryResponse, error)
}

type ContractHandler struct {
	contractClient contractCreateReader
	jobClient      contractJobReader
	proposalClient contractProposalReader
}

type ContractErrorResponse struct {
	Error string `json:"error"`
}

type ContractResponse struct {
	Contract any `json:"contract"`
}

type ContractListResponse struct {
	Contracts     []any  `json:"contracts"`
	NextPageToken string `json:"next_page_token,omitempty"`
}

type ContractBooleanResponse struct {
	Deleted bool `json:"deleted,omitempty"`
}

type ContractBootstrapResponse struct {
	Proposal   any `json:"proposal"`
	JobSummary any `json:"job_summary"`
	OfferState any `json:"offer_state"`
	Contract   any `json:"contract,omitempty"`
}

type ContractReasonRequest struct {
	Reason string `json:"reason"`
}

type ContractMilestoneSubmitRequest struct {
	Note        string   `json:"note"`
	Attachments []string `json:"attachments"`
}

type ContractMilestoneRequestChangesRequest struct {
	Note string `json:"note"`
}

type ContractHourlyLogRequest struct {
	StartAtUnixSeconds int64    `json:"start_at_unix_seconds"`
	EndAtUnixSeconds   int64    `json:"end_at_unix_seconds"`
	Note               string   `json:"note"`
	EvidenceURLs       []string `json:"evidence_urls"`
}

type ContractHourlyLogReviewRequest struct {
	Status     string `json:"status"`
	ReviewNote string `json:"review_note"`
}

type ContractHourlyLogResponse struct {
	HourlyLog any `json:"hourly_log"`
}

type ContractHourlyLogListResponse struct {
	HourlyLogs    []any  `json:"hourly_logs"`
	NextPageToken string `json:"next_page_token,omitempty"`
}

type ContractUploadURLRequest struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
}

type ContractUploadURLResponse struct {
	StorageKey string `json:"storage_key"`
	UploadURL  string `json:"upload_url"`
}

type ContractSummaryResponse struct {
	Summary any `json:"summary"`
}

type ContractInvoiceResponse struct {
	Invoice any `json:"invoice"`
}

type ContractInvoiceListResponse struct {
	Invoices      []any  `json:"invoices"`
	NextPageToken string `json:"next_page_token,omitempty"`
}

type ContractBonusCreateRequest struct {
	AmountMinor int64  `json:"amount_minor"`
	Note        string `json:"note"`
}

type ContractBonusResponse struct {
	Bonus any `json:"bonus"`
}

type ContractBonusListResponse struct {
	Bonuses       []any  `json:"bonuses"`
	NextPageToken string `json:"next_page_token,omitempty"`
}

type ContractAmendmentCreateRequest struct {
	Summary   string                       `json:"summary"`
	Payload   *contractv1.AmendmentPayload `json:"payload"`
	ExpiresAt int64                        `json:"expires_at_unix_seconds"`
}

type ContractAmendmentRespondRequest struct {
	Status       string `json:"status"`
	ResponseNote string `json:"response_note"`
}

type ContractAmendmentResponse struct {
	Amendment any `json:"amendment"`
}

type ContractAmendmentListResponse struct {
	Amendments    []any  `json:"amendments"`
	NextPageToken string `json:"next_page_token,omitempty"`
}

type ContractStatusHistoryResponse struct {
	Entries       []any  `json:"entries"`
	NextPageToken string `json:"next_page_token,omitempty"`
}

func NewContractHandler(contractClient contractCreateReader, jobClient contractJobReader, proposalClient contractProposalReader) *ContractHandler {
	return &ContractHandler{contractClient: contractClient, jobClient: jobClient, proposalClient: proposalClient}
}

// Bootstrap godoc
// @Summary Bootstrap contract offer flow
// @Description Returns proposal, job summary, and offer-state metadata needed before creating a contract offer.
// @Tags Contract
// @Produce json
// @Security BearerAuth
// @Param job_id query int true "Job ID"
// @Param proposal_id query int true "Proposal ID"
// @Success 200 {object} ContractBootstrapResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/bootstrap [get]
func (h *ContractHandler) Bootstrap(c *gin.Context) {
	if h.contractClient == nil || h.jobClient == nil || h.proposalClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "contract dependencies are not configured"})
		return
	}

	jobID, ok := parseBootstrapID(c, "job_id")
	if !ok {
		return
	}
	proposalID, ok := parseBootstrapID(c, "proposal_id")
	if !ok {
		return
	}

	ctx := withAuthContext(c)
	proposalResp, err := h.proposalClient.GetProposal(ctx, &proposalv1.GetProposalRequest{ProposalId: proposalID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	proposal := proposalResp.GetProposal()
	if proposal == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "proposal not found"})
		return
	}
	if proposal.GetJobId() != jobID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "proposal does not belong to job"})
		return
	}

	jobResp, err := h.jobClient.GetJobSummary(ctx, &jobv1.GetJobSummaryRequest{JobId: jobID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	jobSummary := jobResp.GetSummary()
	if jobSummary == nil || !jobSummary.GetFound() {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}
	if strings.TrimSpace(jobSummary.GetClientId()) != strings.TrimSpace(proposal.GetClientId()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "proposal does not belong to job owner"})
		return
	}

	contractsResp, err := h.contractClient.ListMyContracts(ctx, &contractv1.ListMyContractsRequest{PageSize: 200})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	var matchedContract *contractv1.Contract
	var blockingContract *contractv1.Contract
	canOpenOfferForm := true
	blockingReason := ""
	proposalBlockingReason, proposalBlocksOfferForm := bootstrapProposalBlockingReason(proposal.GetStatus())
	var pendingOfferOnJob bool
	var activeContractOnJob bool
	for _, item := range contractsResp.GetContracts() {
		if item.GetJobId() != jobID {
			continue
		}
		if item.GetStatus() == contractv1.ContractStatus_CONTRACT_STATUS_ACTIVE {
			activeContractOnJob = true
			if blockingContract == nil {
				blockingContract = item
			}
		}
		if item.GetStatus() == contractv1.ContractStatus_CONTRACT_STATUS_PENDING_ACCEPTANCE {
			pendingOfferOnJob = true
			if blockingContract == nil {
				blockingContract = item
			}
		}
		if item.GetProposalId() == proposalID && item.GetJobId() == jobID {
			matchedContract = item
		}
	}

	switch {
	case !jobSummary.GetIsOpen():
		canOpenOfferForm = false
		blockingReason = "job_not_open"
	case activeContractOnJob:
		canOpenOfferForm = false
		blockingReason = "active_contract_exists"
	case matchedContract != nil && matchedContract.GetStatus() == contractv1.ContractStatus_CONTRACT_STATUS_PENDING_ACCEPTANCE:
		canOpenOfferForm = false
		blockingReason = "offer_already_sent"
	case pendingOfferOnJob:
		canOpenOfferForm = false
		blockingReason = "pending_offer_exists"
	case proposalBlocksOfferForm:
		canOpenOfferForm = false
		blockingReason = proposalBlockingReason
	}

	offerState := gin.H{
		"has_offer":            matchedContract != nil,
		"has_pending_offer":    proposal.GetStatus() == proposalv1.ProposalStatus_PROPOSAL_STATUS_OFFER_SENT || (matchedContract != nil && matchedContract.GetStatus() == contractv1.ContractStatus_CONTRACT_STATUS_PENDING_ACCEPTANCE) || pendingOfferOnJob,
		"has_active_contract":  proposal.GetStatus() == proposalv1.ProposalStatus_PROPOSAL_STATUS_HIRED || activeContractOnJob,
		"proposal_status":      proposal.GetStatus().String(),
		"job_status":           jobSummary.GetStatus().String(),
		"job_is_open":          jobSummary.GetIsOpen(),
		"can_open_offer_form":  canOpenOfferForm,
		"blocking_reason":      blockingReason,
		"blocking_contract_id": int64(0),
	}
	if blockingContract != nil {
		offerState["blocking_contract_id"] = blockingContract.GetId()
	}

	proposalPayload, err := protoToAny(proposal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	jobPayload, err := protoToAny(jobSummary)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}

	var contractPayload any
	if matchedContract != nil {
		contractPayload, err = protoToAny(matchedContract)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"proposal":    proposalPayload,
		"job_summary": jobPayload,
		"offer_state": offerState,
		"contract":    contractPayload,
	})
}

func mapContractStatus(v string) contractv1.ContractStatus {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "pending_acceptance":
		return contractv1.ContractStatus_CONTRACT_STATUS_PENDING_ACCEPTANCE
	case "active":
		return contractv1.ContractStatus_CONTRACT_STATUS_ACTIVE
	case "declined":
		return contractv1.ContractStatus_CONTRACT_STATUS_DECLINED
	case "revoked":
		return contractv1.ContractStatus_CONTRACT_STATUS_REVOKED
	case "paused":
		return contractv1.ContractStatus_CONTRACT_STATUS_PAUSED
	case "ended":
		return contractv1.ContractStatus_CONTRACT_STATUS_ENDED
	default:
		return contractv1.ContractStatus_CONTRACT_STATUS_UNSPECIFIED
	}
}

func bootstrapProposalBlockingReason(status proposalv1.ProposalStatus) (string, bool) {
	switch status {
	case proposalv1.ProposalStatus_PROPOSAL_STATUS_SENT,
		proposalv1.ProposalStatus_PROPOSAL_STATUS_SHORTLISTED:
		return "", false
	case proposalv1.ProposalStatus_PROPOSAL_STATUS_OFFER_SENT:
		return "offer_already_sent", true
	case proposalv1.ProposalStatus_PROPOSAL_STATUS_HIRED:
		return "active_contract_exists", true
	default:
		return "proposal_not_eligible", true
	}
}

// CreateContract godoc
// @Summary Create contract offer
// @Description Creates a new contract offer for a proposal.
// @Tags Contract
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body contractv1.CreateContractRequest true "Contract create payload"
// @Success 200 {object} ContractResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts [post]
func (h *ContractHandler) CreateContract(c *gin.Context) {
	var req contractv1.CreateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.contractClient.CreateContract(withAuthContext(c), &req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "contract", resp.GetContract())
}

// GetContract godoc
// @Summary Get contract by ID
// @Description Returns contract details by contract ID.
// @Tags Contract
// @Produce json
// @Security BearerAuth
// @Param contractId path int true "Contract ID"
// @Success 200 {object} ContractResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/{contractId} [get]
func (h *ContractHandler) GetContract(c *gin.Context) {
	contractID, ok := parseInt64Param(c, "contractId")
	if !ok {
		return
	}
	resp, err := h.contractClient.GetContract(withAuthContext(c), &contractv1.GetContractRequest{ContractId: contractID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "contract", resp.GetContract())
}

// ListMyContracts godoc
// @Summary List my contracts
// @Description Lists contracts visible to the authenticated caller.
// @Tags Contract
// @Produce json
// @Security BearerAuth
// @Param status query string false "Status filter"
// @Param page_size query int false "Page size" default(20)
// @Param page_token query string false "Page token"
// @Success 200 {object} ContractListResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts [get]
func (h *ContractHandler) ListMyContracts(c *gin.Context) {
	resp, err := h.contractClient.ListMyContracts(withAuthContext(c), &contractv1.ListMyContractsRequest{
		Status:    mapContractStatus(strings.TrimSpace(c.Query("status"))),
		PageSize:  int32(parseIntQuery(c, "page_size", 20)),
		PageToken: strings.TrimSpace(c.Query("page_token")),
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	payload, convErr := protoSliceToAny(resp.GetContracts())
	if convErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"contracts": payload, "next_page_token": resp.GetNextPageToken()})
}

// AcceptContract godoc
// @Summary Accept contract
// @Description Accepts a pending contract offer.
// @Tags Contract
// @Produce json
// @Security BearerAuth
// @Param contractId path int true "Contract ID"
// @Success 200 {object} ContractResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/{contractId}/accept [post]
func (h *ContractHandler) AcceptContract(c *gin.Context) {
	contractID, ok := parseInt64Param(c, "contractId")
	if !ok {
		return
	}
	resp, err := h.contractClient.AcceptContract(withAuthContext(c), &contractv1.AcceptContractRequest{ContractId: contractID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "contract", resp.GetContract())
}

// DeclineContract godoc
// @Summary Decline contract
// @Description Declines a pending contract offer.
// @Tags Contract
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contractId path int true "Contract ID"
// @Param request body ContractReasonRequest true "Decline reason payload"
// @Success 200 {object} ContractResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/{contractId}/decline [post]
func (h *ContractHandler) DeclineContract(c *gin.Context) {
	contractID, ok := parseInt64Param(c, "contractId")
	if !ok {
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.contractClient.DeclineContract(withAuthContext(c), &contractv1.DeclineContractRequest{ContractId: contractID, Reason: body.Reason})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "contract", resp.GetContract())
}

// RevokeContractOffer godoc
// @Summary Revoke contract offer
// @Description Revokes a pending contract offer.
// @Tags Contract
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contractId path int true "Contract ID"
// @Param request body ContractReasonRequest true "Revoke reason payload"
// @Success 200 {object} ContractResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/{contractId}/revoke [post]
func (h *ContractHandler) RevokeContractOffer(c *gin.Context) {
	contractID, ok := parseInt64Param(c, "contractId")
	if !ok {
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.contractClient.RevokeContractOffer(withAuthContext(c), &contractv1.RevokeContractOfferRequest{ContractId: contractID, Reason: body.Reason})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "contract", resp.GetContract())
}

// SubmitMilestoneWork godoc
// @Summary Submit milestone work
// @Description Submits milestone deliverables for freelancer review by client.
// @Tags Contract
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contractId path int true "Contract ID"
// @Param milestoneId path int true "Milestone ID"
// @Param request body ContractMilestoneSubmitRequest true "Milestone submission payload"
// @Success 200 {object} ContractResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/{contractId}/milestones/{milestoneId}/submit [post]
func (h *ContractHandler) SubmitMilestoneWork(c *gin.Context) {
	contractID, ok := parseInt64Param(c, "contractId")
	if !ok {
		return
	}
	milestoneID, ok := parseInt64Param(c, "milestoneId")
	if !ok {
		return
	}
	var body struct {
		Note        string   `json:"note"`
		Attachments []string `json:"attachments"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.contractClient.SubmitMilestoneWork(withAuthContext(c), &contractv1.SubmitMilestoneWorkRequest{
		ContractId:  contractID,
		MilestoneId: milestoneID,
		Note:        body.Note,
		Attachments: body.Attachments,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "contract", resp.GetContract())
}

// RequestMilestoneChanges godoc
// @Summary Request milestone changes
// @Description Requests changes on a milestone submission.
// @Tags Contract
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contractId path int true "Contract ID"
// @Param milestoneId path int true "Milestone ID"
// @Param request body ContractMilestoneRequestChangesRequest true "Change request payload"
// @Success 200 {object} ContractResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/{contractId}/milestones/{milestoneId}/request-changes [post]
func (h *ContractHandler) RequestMilestoneChanges(c *gin.Context) {
	contractID, ok := parseInt64Param(c, "contractId")
	if !ok {
		return
	}
	milestoneID, ok := parseInt64Param(c, "milestoneId")
	if !ok {
		return
	}
	var body struct {
		Note string `json:"note"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.contractClient.RequestMilestoneChanges(withAuthContext(c), &contractv1.RequestMilestoneChangesRequest{
		ContractId:  contractID,
		MilestoneId: milestoneID,
		Note:        body.Note,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "contract", resp.GetContract())
}

// ApproveMilestoneSubmission godoc
// @Summary Approve milestone submission
// @Description Approves a submitted milestone.
// @Tags Contract
// @Produce json
// @Security BearerAuth
// @Param contractId path int true "Contract ID"
// @Param milestoneId path int true "Milestone ID"
// @Success 200 {object} ContractResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/{contractId}/milestones/{milestoneId}/approve [post]
func (h *ContractHandler) ApproveMilestoneSubmission(c *gin.Context) {
	contractID, ok := parseInt64Param(c, "contractId")
	if !ok {
		return
	}
	milestoneID, ok := parseInt64Param(c, "milestoneId")
	if !ok {
		return
	}
	resp, err := h.contractClient.ApproveMilestoneSubmission(withAuthContext(c), &contractv1.ApproveMilestoneSubmissionRequest{
		ContractId:  contractID,
		MilestoneId: milestoneID,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "contract", resp.GetContract())
}

// LogHourlyWork godoc
// @Summary Log hourly work
// @Description Creates an hourly work log entry.
// @Tags Contract
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contractId path int true "Contract ID"
// @Param request body ContractHourlyLogRequest true "Hourly log payload"
// @Success 200 {object} ContractHourlyLogResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/{contractId}/hourly-logs [post]
func (h *ContractHandler) LogHourlyWork(c *gin.Context) {
	contractID, ok := parseInt64Param(c, "contractId")
	if !ok {
		return
	}
	var body struct {
		StartAtUnixSeconds int64    `json:"start_at_unix_seconds"`
		EndAtUnixSeconds   int64    `json:"end_at_unix_seconds"`
		Note               string   `json:"note"`
		EvidenceURLs       []string `json:"evidence_urls"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.contractClient.LogHourlyWork(withAuthContext(c), &contractv1.LogHourlyWorkRequest{
		ContractId:         contractID,
		StartAtUnixSeconds: body.StartAtUnixSeconds,
		EndAtUnixSeconds:   body.EndAtUnixSeconds,
		Note:               body.Note,
		EvidenceUrls:       body.EvidenceURLs,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "hourly_log", resp.GetHourlyLog())
}

// GetHourlyLogEvidenceUploadUrl godoc
// @Summary Reserve hourly log evidence upload URL
// @Description Returns a pre-signed upload URL for hourly log evidence.
// @Tags Contract
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contractId path int true "Contract ID"
// @Param request body ContractUploadURLRequest true "Upload URL payload"
// @Success 200 {object} ContractUploadURLResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/{contractId}/hourly-logs/evidence/upload-url [post]
func (h *ContractHandler) GetHourlyLogEvidenceUploadUrl(c *gin.Context) {
	contractID, ok := parseInt64Param(c, "contractId")
	if !ok {
		return
	}
	var body struct {
		FileName    string `json:"file_name"`
		ContentType string `json:"content_type"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.contractClient.GetHourlyLogEvidenceUploadUrl(withAuthContext(c), &contractv1.GetHourlyLogEvidenceUploadUrlRequest{
		ContractId:  contractID,
		FileName:    body.FileName,
		ContentType: body.ContentType,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"storage_key": resp.GetStorageKey(), "upload_url": resp.GetUploadUrl()})
}

// ListHourlyLogs godoc
// @Summary List hourly logs
// @Description Lists hourly logs for a contract.
// @Tags Contract
// @Produce json
// @Security BearerAuth
// @Param contractId path int true "Contract ID"
// @Param page_size query int false "Page size" default(20)
// @Param page_token query string false "Page token"
// @Success 200 {object} ContractHourlyLogListResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/{contractId}/hourly-logs [get]
func (h *ContractHandler) ListHourlyLogs(c *gin.Context) {
	contractID, ok := parseInt64Param(c, "contractId")
	if !ok {
		return
	}
	resp, err := h.contractClient.ListHourlyLogs(withAuthContext(c), &contractv1.ListHourlyLogsRequest{
		ContractId: contractID,
		PageSize:   int32(parseIntQuery(c, "page_size", 20)),
		PageToken:  strings.TrimSpace(c.Query("page_token")),
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	payload, convErr := protoSliceToAny(resp.GetHourlyLogs())
	if convErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"hourly_logs": payload, "next_page_token": resp.GetNextPageToken()})
}

// GetHourlyWorkSummary godoc
// @Summary Get hourly work summary
// @Description Returns weekly hourly work summary for a contract.
// @Tags Contract
// @Produce json
// @Security BearerAuth
// @Param contractId path int true "Contract ID"
// @Param week_start_unix_seconds query int false "Week start unix seconds"
// @Success 200 {object} ContractSummaryResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/{contractId}/hourly-summary [get]
func (h *ContractHandler) GetHourlyWorkSummary(c *gin.Context) {
	contractID, ok := parseInt64Param(c, "contractId")
	if !ok {
		return
	}
	resp, err := h.contractClient.GetHourlyWorkSummary(withAuthContext(c), &contractv1.GetHourlyWorkSummaryRequest{
		ContractId:           contractID,
		WeekStartUnixSeconds: int64(parseIntQuery(c, "week_start_unix_seconds", 0)),
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "summary", resp)
}

// UpdateHourlyLog godoc
// @Summary Update hourly log
// @Description Updates an existing hourly log entry.
// @Tags Contract
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param hourlyLogId path int true "Hourly log ID"
// @Param request body ContractHourlyLogRequest true "Hourly log update payload"
// @Success 200 {object} ContractHourlyLogResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/hourly-logs/{hourlyLogId} [patch]
func (h *ContractHandler) UpdateHourlyLog(c *gin.Context) {
	hourlyLogID, ok := parseInt64Param(c, "hourlyLogId")
	if !ok {
		return
	}
	var body struct {
		StartAtUnixSeconds int64    `json:"start_at_unix_seconds"`
		EndAtUnixSeconds   int64    `json:"end_at_unix_seconds"`
		Note               string   `json:"note"`
		EvidenceURLs       []string `json:"evidence_urls"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.contractClient.UpdateHourlyLog(withAuthContext(c), &contractv1.UpdateHourlyLogRequest{
		HourlyLogId:        hourlyLogID,
		StartAtUnixSeconds: body.StartAtUnixSeconds,
		EndAtUnixSeconds:   body.EndAtUnixSeconds,
		Note:               body.Note,
		EvidenceUrls:       body.EvidenceURLs,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "hourly_log", resp.GetHourlyLog())
}

// DeleteHourlyLog godoc
// @Summary Delete hourly log
// @Description Deletes an hourly log entry.
// @Tags Contract
// @Produce json
// @Security BearerAuth
// @Param hourlyLogId path int true "Hourly log ID"
// @Success 200 {object} ContractBooleanResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/hourly-logs/{hourlyLogId} [delete]
func (h *ContractHandler) DeleteHourlyLog(c *gin.Context) {
	hourlyLogID, ok := parseInt64Param(c, "hourlyLogId")
	if !ok {
		return
	}
	if _, err := h.contractClient.DeleteHourlyLog(withAuthContext(c), &contractv1.DeleteHourlyLogRequest{HourlyLogId: hourlyLogID}); err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

// ReviewHourlyLog godoc
// @Summary Review hourly log
// @Description Approves or rejects an hourly log.
// @Tags Contract
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param hourlyLogId path int true "Hourly log ID"
// @Param request body ContractHourlyLogReviewRequest true "Hourly log review payload"
// @Success 200 {object} ContractHourlyLogResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/hourly-logs/{hourlyLogId}/review [post]
func (h *ContractHandler) ReviewHourlyLog(c *gin.Context) {
	hourlyLogID, ok := parseInt64Param(c, "hourlyLogId")
	if !ok {
		return
	}
	var body struct {
		Status     string `json:"status"`
		ReviewNote string `json:"review_note"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	statusEnum, ok := contractv1.HourlyLogStatus_value[strings.ToUpper(strings.TrimSpace(body.Status))]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}
	resp, err := h.contractClient.ReviewHourlyLog(withAuthContext(c), &contractv1.ReviewHourlyLogRequest{
		HourlyLogId: hourlyLogID,
		Status:      contractv1.HourlyLogStatus(statusEnum),
		ReviewNote:  body.ReviewNote,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "hourly_log", resp.GetHourlyLog())
}

// GetHourlyInvoice godoc
// @Summary Get hourly invoice
// @Description Returns an hourly invoice by invoice ID.
// @Tags Contract
// @Produce json
// @Security BearerAuth
// @Param invoiceId path int true "Invoice ID"
// @Success 200 {object} ContractInvoiceResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/hourly-invoices/{invoiceId} [get]
func (h *ContractHandler) GetHourlyInvoice(c *gin.Context) {
	invoiceID, ok := parseInt64Param(c, "invoiceId")
	if !ok {
		return
	}
	resp, err := h.contractClient.GetHourlyInvoice(withAuthContext(c), &contractv1.GetHourlyInvoiceRequest{InvoiceId: invoiceID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "invoice", resp.GetInvoice())
}

// ListHourlyInvoices godoc
// @Summary List hourly invoices
// @Description Lists hourly invoices for a contract.
// @Tags Contract
// @Produce json
// @Security BearerAuth
// @Param contractId path int true "Contract ID"
// @Param page_size query int false "Page size" default(20)
// @Param page_token query string false "Page token"
// @Success 200 {object} ContractInvoiceListResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/{contractId}/hourly-invoices [get]
func (h *ContractHandler) ListHourlyInvoices(c *gin.Context) {
	contractID, ok := parseInt64Param(c, "contractId")
	if !ok {
		return
	}
	resp, err := h.contractClient.ListHourlyInvoices(withAuthContext(c), &contractv1.ListHourlyInvoicesRequest{
		ContractId: contractID,
		PageSize:   int32(parseIntQuery(c, "page_size", 20)),
		PageToken:  strings.TrimSpace(c.Query("page_token")),
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	payload, convErr := protoSliceToAny(resp.GetInvoices())
	if convErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"invoices": payload, "next_page_token": resp.GetNextPageToken()})
}

// CreateContractBonus godoc
// @Summary Create contract bonus
// @Description Creates a bonus payment entry for a contract.
// @Tags Contract
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contractId path int true "Contract ID"
// @Param request body ContractBonusCreateRequest true "Bonus payload"
// @Success 200 {object} ContractBonusResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/{contractId}/bonuses [post]
func (h *ContractHandler) CreateContractBonus(c *gin.Context) {
	contractID, ok := parseInt64Param(c, "contractId")
	if !ok {
		return
	}
	var body struct {
		AmountMinor int64  `json:"amount_minor"`
		Note        string `json:"note"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.contractClient.CreateContractBonus(withAuthContext(c), &contractv1.CreateContractBonusRequest{
		ContractId:  contractID,
		AmountMinor: body.AmountMinor,
		Note:        body.Note,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "bonus", resp.GetBonus())
}

// ListContractBonuses godoc
// @Summary List contract bonuses
// @Description Lists bonuses for a contract.
// @Tags Contract
// @Produce json
// @Security BearerAuth
// @Param contractId path int true "Contract ID"
// @Param page_size query int false "Page size" default(20)
// @Param page_token query string false "Page token"
// @Success 200 {object} ContractBonusListResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/{contractId}/bonuses [get]
func (h *ContractHandler) ListContractBonuses(c *gin.Context) {
	contractID, ok := parseInt64Param(c, "contractId")
	if !ok {
		return
	}
	resp, err := h.contractClient.ListContractBonuses(withAuthContext(c), &contractv1.ListContractBonusesRequest{
		ContractId: contractID,
		PageSize:   int32(parseIntQuery(c, "page_size", 20)),
		PageToken:  strings.TrimSpace(c.Query("page_token")),
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	payload, convErr := protoSliceToAny(resp.GetBonuses())
	if convErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bonuses": payload, "next_page_token": resp.GetNextPageToken()})
}

// ProposeAmendment godoc
// @Summary Propose contract amendment
// @Description Creates a new contract amendment proposal.
// @Tags Contract
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contractId path int true "Contract ID"
// @Param request body ContractAmendmentCreateRequest true "Amendment payload"
// @Success 200 {object} ContractAmendmentResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/{contractId}/amendments [post]
func (h *ContractHandler) ProposeAmendment(c *gin.Context) {
	contractID, ok := parseInt64Param(c, "contractId")
	if !ok {
		return
	}
	var body struct {
		Summary   string                       `json:"summary"`
		Payload   *contractv1.AmendmentPayload `json:"payload"`
		ExpiresAt int64                        `json:"expires_at_unix_seconds"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.contractClient.ProposeAmendment(withAuthContext(c), &contractv1.ProposeAmendmentRequest{
		ContractId:           contractID,
		Summary:              body.Summary,
		Payload:              body.Payload,
		ExpiresAtUnixSeconds: body.ExpiresAt,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "amendment", resp.GetAmendment())
}

// RespondAmendment godoc
// @Summary Respond to amendment
// @Description Accepts or rejects a contract amendment.
// @Tags Contract
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contractId path int true "Contract ID"
// @Param amendmentId path int true "Amendment ID"
// @Param request body ContractAmendmentRespondRequest true "Amendment response payload"
// @Success 200 {object} ContractAmendmentResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/{contractId}/amendments/{amendmentId}/respond [post]
func (h *ContractHandler) RespondAmendment(c *gin.Context) {
	_, ok := parseInt64Param(c, "contractId")
	if !ok {
		return
	}
	amendmentID, ok := parseInt64Param(c, "amendmentId")
	if !ok {
		return
	}
	var body struct {
		Status       string `json:"status"`
		ResponseNote string `json:"response_note"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	statusEnum, ok := contractv1.AmendmentStatus_value[strings.ToUpper(strings.TrimSpace(body.Status))]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}
	resp, err := h.contractClient.RespondAmendment(withAuthContext(c), &contractv1.RespondAmendmentRequest{
		AmendmentId:  amendmentID,
		Status:       contractv1.AmendmentStatus(statusEnum),
		ResponseNote: body.ResponseNote,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "amendment", resp.GetAmendment())
}

// ListAmendments godoc
// @Summary List amendments
// @Description Lists contract amendments.
// @Tags Contract
// @Produce json
// @Security BearerAuth
// @Param contractId path int true "Contract ID"
// @Param page_size query int false "Page size" default(20)
// @Param page_token query string false "Page token"
// @Success 200 {object} ContractAmendmentListResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/{contractId}/amendments [get]
func (h *ContractHandler) ListAmendments(c *gin.Context) {
	contractID, ok := parseInt64Param(c, "contractId")
	if !ok {
		return
	}
	resp, err := h.contractClient.ListAmendments(withAuthContext(c), &contractv1.ListAmendmentsRequest{
		ContractId: contractID,
		PageSize:   int32(parseIntQuery(c, "page_size", 20)),
		PageToken:  strings.TrimSpace(c.Query("page_token")),
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	payload, convErr := protoSliceToAny(resp.GetAmendments())
	if convErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"amendments": payload, "next_page_token": resp.GetNextPageToken()})
}

// PauseContract godoc
// @Summary Pause contract
// @Description Pauses an active contract.
// @Tags Contract
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contractId path int true "Contract ID"
// @Param request body ContractReasonRequest true "Pause reason payload"
// @Success 200 {object} ContractResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/{contractId}/pause [post]
func (h *ContractHandler) PauseContract(c *gin.Context) {
	contractID, ok := parseInt64Param(c, "contractId")
	if !ok {
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.contractClient.PauseContract(withAuthContext(c), &contractv1.PauseContractRequest{ContractId: contractID, Reason: body.Reason})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "contract", resp.GetContract())
}

// ResumeContract godoc
// @Summary Resume contract
// @Description Resumes a paused contract.
// @Tags Contract
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contractId path int true "Contract ID"
// @Param request body ContractReasonRequest true "Resume reason payload"
// @Success 200 {object} ContractResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/{contractId}/resume [post]
func (h *ContractHandler) ResumeContract(c *gin.Context) {
	contractID, ok := parseInt64Param(c, "contractId")
	if !ok {
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.contractClient.ResumeContract(withAuthContext(c), &contractv1.ResumeContractRequest{ContractId: contractID, Reason: body.Reason})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "contract", resp.GetContract())
}

// EndContract godoc
// @Summary End contract
// @Description Ends a contract.
// @Tags Contract
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contractId path int true "Contract ID"
// @Param request body ContractReasonRequest true "End reason payload"
// @Success 200 {object} ContractResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/{contractId}/end [post]
func (h *ContractHandler) EndContract(c *gin.Context) {
	contractID, ok := parseInt64Param(c, "contractId")
	if !ok {
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.contractClient.EndContract(withAuthContext(c), &contractv1.EndContractRequest{ContractId: contractID, Reason: body.Reason})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "contract", resp.GetContract())
}

// GetStatusHistory godoc
// @Summary Get contract status history
// @Description Returns status history entries for a contract.
// @Tags Contract
// @Produce json
// @Security BearerAuth
// @Param contractId path int true "Contract ID"
// @Param page_size query int false "Page size" default(20)
// @Param page_token query string false "Page token"
// @Success 200 {object} ContractStatusHistoryResponse
// @Failure 400 {object} ContractErrorResponse
// @Failure 401 {object} ContractErrorResponse
// @Failure 500 {object} ContractErrorResponse
// @Router /api/v1/contracts/{contractId}/status-history [get]
func (h *ContractHandler) GetStatusHistory(c *gin.Context) {
	contractID, ok := parseInt64Param(c, "contractId")
	if !ok {
		return
	}
	resp, err := h.contractClient.GetStatusHistory(withAuthContext(c), &contractv1.GetStatusHistoryRequest{
		ContractId: contractID,
		PageSize:   int32(parseIntQuery(c, "page_size", 20)),
		PageToken:  strings.TrimSpace(c.Query("page_token")),
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	payload, convErr := protoSliceToAny(resp.GetEntries())
	if convErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"entries": payload, "next_page_token": resp.GetNextPageToken()})
}

func parseBootstrapID(c *gin.Context, key string) (int64, bool) {
	v := strings.TrimSpace(c.Query(key))
	if v == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": key + " is required"})
		return 0, false
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil || n <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": key + " must be a positive integer"})
		return 0, false
	}
	return n, true
}
