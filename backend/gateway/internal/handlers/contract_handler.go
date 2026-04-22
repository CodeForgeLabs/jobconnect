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

type ContractHandler struct {
	contractClient contractListReader
	jobClient      contractJobReader
	proposalClient contractProposalReader
}

func NewContractHandler(contractClient contractListReader, jobClient contractJobReader, proposalClient contractProposalReader) *ContractHandler {
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
	for _, item := range contractsResp.GetContracts() {
		if item.GetProposalId() == proposalID && item.GetJobId() == jobID {
			matchedContract = item
			break
		}
	}

	offerState := gin.H{
		"has_offer":           matchedContract != nil,
		"has_pending_offer":   proposal.GetStatus() == proposalv1.ProposalStatus_PROPOSAL_STATUS_OFFER_SENT || (matchedContract != nil && matchedContract.GetStatus() == contractv1.ContractStatus_CONTRACT_STATUS_PENDING_ACCEPTANCE),
		"has_active_contract": proposal.GetStatus() == proposalv1.ProposalStatus_PROPOSAL_STATUS_HIRED || (matchedContract != nil && matchedContract.GetStatus() == contractv1.ContractStatus_CONTRACT_STATUS_ACTIVE),
		"proposal_status":     proposal.GetStatus().String(),
		"job_status":          jobSummary.GetStatus().String(),
		"job_is_open":         jobSummary.GetIsOpen(),
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
