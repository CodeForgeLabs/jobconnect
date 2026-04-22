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
}

type ContractHandler struct {
	contractClient contractCreateReader
	jobClient      contractJobReader
	proposalClient contractProposalReader
}

func NewContractHandler(contractClient contractCreateReader, jobClient contractJobReader, proposalClient contractProposalReader) *ContractHandler {
	return &ContractHandler{contractClient: contractClient, jobClient: jobClient, proposalClient: proposalClient}
}

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
