package handlers

import (
	"context"
	"net/http"
	"strings"

	proposalv1 "jobconnect/proposal/gen/proposal/v1"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
)

type ProposalHandler struct {
	client proposalv1.ProposalServiceClient
}

func NewProposalHandler(client proposalv1.ProposalServiceClient) *ProposalHandler {
	return &ProposalHandler{client: client}
}

func (h *ProposalHandler) GetProposal(c *gin.Context) {
	proposalID, ok := parseInt64Param(c, "proposalId")
	if !ok {
		return
	}
	resp, err := h.client.GetProposal(withAuthContext(c), &proposalv1.GetProposalRequest{ProposalId: proposalID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "proposal", resp.GetProposal())
}

func (h *ProposalHandler) GetMyProposalForJob(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
	if !ok {
		return
	}
	resp, err := h.client.GetMyProposalForJob(withAuthContext(c), &proposalv1.GetMyProposalForJobRequest{JobId: jobID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "proposal", resp.GetProposal())
}

func (h *ProposalHandler) HasAppliedToJob(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
	if !ok {
		return
	}
	resp, err := h.client.HasAppliedToJob(withAuthContext(c), &proposalv1.HasAppliedToJobRequest{JobId: jobID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	payload, convErr := protoToAny(resp)
	if convErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, payload)
}

func (h *ProposalHandler) ListMyProposals(c *gin.Context) {
	statuses := parseProposalStatusFilters(c.QueryArray("status"))
	var jobFilter *int64
	if strings.TrimSpace(c.Query("job_id")) != "" {
		jobID := parseIntQuery(c, "job_id", 0)
		v := int64(jobID)
		jobFilter = &v
	}
	resp, err := h.client.ListMyProposals(withAuthContext(c), &proposalv1.ListMyProposalsRequest{
		StatusFilter: statuses,
		JobIdFilter:  jobFilter,
		SortBy:       mapProposalSort(strings.TrimSpace(c.Query("sort_by"))),
		PageSize:     int32(parseIntQuery(c, "page_size", 20)),
		PageToken:    strings.TrimSpace(c.Query("page_token")),
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	payload, convErr := protoSliceToAny(resp.GetProposals())
	if convErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"proposals": payload, "next_page_token": resp.GetNextPageToken()})
}

func (h *ProposalHandler) ListClientProposals(c *gin.Context) {
	statuses := parseProposalStatusFilters(c.QueryArray("status"))
	var jobFilter *int64
	if strings.TrimSpace(c.Query("job_id")) != "" {
		jobID := parseIntQuery(c, "job_id", 0)
		v := int64(jobID)
		jobFilter = &v
	}
	var freelancerFilter *string
	if v := strings.TrimSpace(c.Query("freelancer_id")); v != "" {
		freelancerFilter = &v
	}
	resp, err := h.client.ListClientProposals(withAuthContext(c), &proposalv1.ListClientProposalsRequest{
		StatusFilter:       statuses,
		JobIdFilter:        jobFilter,
		FreelancerIdFilter: freelancerFilter,
		SortBy:             mapProposalSort(strings.TrimSpace(c.Query("sort_by"))),
		PageSize:           int32(parseIntQuery(c, "page_size", 20)),
		PageToken:          strings.TrimSpace(c.Query("page_token")),
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	payload, convErr := protoSliceToAny(resp.GetProposals())
	if convErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"proposals": payload, "next_page_token": resp.GetNextPageToken()})
}

func (h *ProposalHandler) CountProposalsByJob(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
	if !ok {
		return
	}
	resp, err := h.client.CountProposalsByJob(withAuthContext(c), &proposalv1.CountProposalsByJobRequest{JobId: jobID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	payload, convErr := protoToAny(resp)
	if convErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, payload)
}

func (h *ProposalHandler) CountClientProposalInbox(c *gin.Context) {
	statuses := parseProposalStatusFilters(c.QueryArray("status"))
	resp, err := h.client.CountClientProposalInbox(withAuthContext(c), &proposalv1.CountClientProposalInboxRequest{StatusFilter: statuses})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	payload, convErr := protoToAny(resp)
	if convErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, payload)
}

func (h *ProposalHandler) SetProposalDecision(c *gin.Context) {
	proposalID, ok := parseInt64Param(c, "proposalId")
	if !ok {
		return
	}
	var body struct {
		Decision string `json:"decision"`
		Reason   string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.SetProposalStatus(withAuthContext(c), &proposalv1.SetProposalStatusRequest{
		ProposalId: proposalID,
		Decision:   mapClientDecision(body.Decision),
		Reason:     body.Reason,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "proposal", resp.GetProposal())
}

func (h *ProposalHandler) GetProposalAttachmentUploadURL(c *gin.Context) {
	proposalID, ok := parseInt64Param(c, "proposalId")
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
	resp, err := h.client.GetProposalAttachmentUploadUrl(withAuthContext(c), &proposalv1.GetProposalAttachmentUploadUrlRequest{
		ProposalId:  proposalID,
		FileName:    body.FileName,
		ContentType: body.ContentType,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	payload, convErr := protoToAny(resp)
	if convErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, payload)
}

func (h *ProposalHandler) GetProposalAttachmentDownloadURL(c *gin.Context) {
	proposalID, ok := parseInt64Param(c, "proposalId")
	if !ok {
		return
	}
	attachmentID, ok := parseInt64Param(c, "attachmentId")
	if !ok {
		return
	}
	resp, err := h.client.GetProposalAttachmentDownloadUrl(withAuthContext(c), &proposalv1.GetProposalAttachmentDownloadUrlRequest{
		ProposalId:   proposalID,
		AttachmentId: attachmentID,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	payload, convErr := protoToAny(resp)
	if convErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, payload)
}

func parseProposalStatusFilters(values []string) []proposalv1.ProposalStatus {
	out := make([]proposalv1.ProposalStatus, 0, len(values))
	for _, v := range values {
		s := mapProposalStatus(v)
		if s == proposalv1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED {
			continue
		}
		out = append(out, s)
	}
	return out
}

func mapProposalSort(v string) proposalv1.SortBy {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "oldest":
		return proposalv1.SortBy_SORT_BY_OLDEST
	case "bid_high":
		return proposalv1.SortBy_SORT_BY_BID_HIGH
	case "bid_low":
		return proposalv1.SortBy_SORT_BY_BID_LOW
	case "newest":
		fallthrough
	default:
		return proposalv1.SortBy_SORT_BY_NEWEST
	}
}

func mapProposalStatus(v string) proposalv1.ProposalStatus {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "sent":
		return proposalv1.ProposalStatus_PROPOSAL_STATUS_SENT
	case "shortlist":
		fallthrough
	case "shortlisted":
		return proposalv1.ProposalStatus_PROPOSAL_STATUS_SHORTLISTED
	case "reject":
		fallthrough
	case "rejected":
		return proposalv1.ProposalStatus_PROPOSAL_STATUS_REJECTED
	case "hired":
		return proposalv1.ProposalStatus_PROPOSAL_STATUS_HIRED
	case "withdrawn":
		return proposalv1.ProposalStatus_PROPOSAL_STATUS_WITHDRAWN
	default:
		return proposalv1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED
	}
}

func mapClientDecision(v string) proposalv1.ClientDecision {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "shortlist":
		fallthrough
	case "shortlisted":
		return proposalv1.ClientDecision_CLIENT_DECISION_SHORTLISTED
	case "reject":
		fallthrough
	case "rejected":
		return proposalv1.ClientDecision_CLIENT_DECISION_REJECTED
	default:
		return proposalv1.ClientDecision_CLIENT_DECISION_UNSPECIFIED
	}
}

func withAuthContext(c *gin.Context) context.Context {
	ctx := c.Request.Context()
	authz := strings.TrimSpace(c.GetHeader("Authorization"))
	if authz == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "authorization", authz)
}
