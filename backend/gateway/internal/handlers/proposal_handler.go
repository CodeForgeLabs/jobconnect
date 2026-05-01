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

type ProposalErrorResponse struct {
	Error string `json:"error"`
}

type ProposalResponse struct {
	Proposal any `json:"proposal"`
}

type ProposalListResponse struct {
	Proposals     []any  `json:"proposals"`
	NextPageToken string `json:"next_page_token,omitempty"`
}

type ProposalDecisionRequest struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

type ProposalAttachmentUploadURLRequest struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
}

type ProposalAttachmentURLResponse struct {
	StorageKey  string `json:"storage_key,omitempty"`
	UploadURL   string `json:"upload_url,omitempty"`
	DownloadURL string `json:"download_url,omitempty"`
}

type ProposalCountResponse struct {
	Count int64 `json:"count"`
}

type ProposalHasAppliedResponse struct {
	HasApplied bool `json:"has_applied"`
}

// GetProposal godoc
// @Summary Get proposal by ID
// @Description Returns a proposal by proposal ID.
// @Tags Proposal
// @Produce json
// @Security BearerAuth
// @Param proposalId path int true "Proposal ID"
// @Success 200 {object} ProposalResponse
// @Failure 400 {object} ProposalErrorResponse
// @Failure 401 {object} ProposalErrorResponse
// @Failure 500 {object} ProposalErrorResponse
// @Router /api/v1/proposals/{proposalId} [get]
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

// GetMyProposalForJob godoc
// @Summary Get my proposal for a job
// @Description Returns the authenticated freelancer's proposal for the given job.
// @Tags Proposal
// @Produce json
// @Security BearerAuth
// @Param jobId path int true "Job ID"
// @Success 200 {object} ProposalResponse
// @Failure 400 {object} ProposalErrorResponse
// @Failure 401 {object} ProposalErrorResponse
// @Failure 500 {object} ProposalErrorResponse
// @Router /api/v1/proposals/me/jobs/{jobId} [get]
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

// HasAppliedToJob godoc
// @Summary Check if I applied to a job
// @Description Returns whether the authenticated freelancer has applied to the given job.
// @Tags Proposal
// @Produce json
// @Security BearerAuth
// @Param jobId path int true "Job ID"
// @Success 200 {object} ProposalHasAppliedResponse
// @Failure 400 {object} ProposalErrorResponse
// @Failure 401 {object} ProposalErrorResponse
// @Failure 500 {object} ProposalErrorResponse
// @Router /api/v1/proposals/me/jobs/{jobId}/has-applied [get]
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

// ListMyProposals godoc
// @Summary List my proposals
// @Description Lists proposals for the authenticated freelancer.
// @Tags Proposal
// @Produce json
// @Security BearerAuth
// @Param status query []string false "Proposal status filters"
// @Param job_id query int false "Job ID filter"
// @Param sort_by query string false "Sort by (newest|oldest|bid_high|bid_low)"
// @Param page_size query int false "Page size" default(20)
// @Param page_token query string false "Page token"
// @Success 200 {object} ProposalListResponse
// @Failure 401 {object} ProposalErrorResponse
// @Failure 500 {object} ProposalErrorResponse
// @Router /api/v1/proposals/me [get]
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

// ListClientProposals godoc
// @Summary List client proposals
// @Description Lists proposals for the authenticated client.
// @Tags Proposal
// @Produce json
// @Security BearerAuth
// @Param status query []string false "Proposal status filters"
// @Param job_id query int false "Job ID filter"
// @Param freelancer_id query string false "Freelancer ID filter"
// @Param sort_by query string false "Sort by (newest|oldest|bid_high|bid_low)"
// @Param page_size query int false "Page size" default(20)
// @Param page_token query string false "Page token"
// @Success 200 {object} ProposalListResponse
// @Failure 401 {object} ProposalErrorResponse
// @Failure 500 {object} ProposalErrorResponse
// @Router /api/v1/proposals/client [get]
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

// CountProposalsByJob godoc
// @Summary Count proposals by job
// @Description Returns proposal counts for a specific job.
// @Tags Proposal
// @Produce json
// @Security BearerAuth
// @Param jobId path int true "Job ID"
// @Success 200 {object} ProposalCountResponse
// @Failure 400 {object} ProposalErrorResponse
// @Failure 401 {object} ProposalErrorResponse
// @Failure 500 {object} ProposalErrorResponse
// @Router /api/v1/proposals/jobs/{jobId}/counts [get]
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

// CountClientProposalInbox godoc
// @Summary Count client proposal inbox
// @Description Returns inbox counts for the authenticated client.
// @Tags Proposal
// @Produce json
// @Security BearerAuth
// @Param status query []string false "Proposal status filters"
// @Success 200 {object} ProposalCountResponse
// @Failure 401 {object} ProposalErrorResponse
// @Failure 500 {object} ProposalErrorResponse
// @Router /api/v1/proposals/client/counts [get]
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

// SetProposalDecision godoc
// @Summary Set proposal decision
// @Description Sets client decision for a proposal.
// @Tags Proposal
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param proposalId path int true "Proposal ID"
// @Param request body ProposalDecisionRequest true "Decision payload"
// @Success 200 {object} ProposalResponse
// @Failure 400 {object} ProposalErrorResponse
// @Failure 401 {object} ProposalErrorResponse
// @Failure 500 {object} ProposalErrorResponse
// @Router /api/v1/proposals/{proposalId}/decision [post]
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

// GetProposalAttachmentUploadURL godoc
// @Summary Reserve proposal attachment upload URL
// @Description Returns a pre-signed upload URL for a proposal attachment.
// @Tags Proposal
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param proposalId path int true "Proposal ID"
// @Param request body ProposalAttachmentUploadURLRequest true "Attachment upload payload"
// @Success 200 {object} ProposalAttachmentURLResponse
// @Failure 400 {object} ProposalErrorResponse
// @Failure 401 {object} ProposalErrorResponse
// @Failure 500 {object} ProposalErrorResponse
// @Router /api/v1/proposals/{proposalId}/attachments/upload-url [post]
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

// GetProposalAttachmentDownloadURL godoc
// @Summary Get proposal attachment download URL
// @Description Returns a pre-signed download URL for a proposal attachment.
// @Tags Proposal
// @Produce json
// @Security BearerAuth
// @Param proposalId path int true "Proposal ID"
// @Param attachmentId path int true "Attachment ID"
// @Success 200 {object} ProposalAttachmentURLResponse
// @Failure 400 {object} ProposalErrorResponse
// @Failure 401 {object} ProposalErrorResponse
// @Failure 500 {object} ProposalErrorResponse
// @Router /api/v1/proposals/{proposalId}/attachments/{attachmentId}/download-url [get]
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
