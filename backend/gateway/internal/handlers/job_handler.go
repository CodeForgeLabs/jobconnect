package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"jobconnect/gateway/internal/middleware"
	jobv1 "jobconnect/job/gen/job/v1"

	"github.com/gin-gonic/gin"
)

type JobHandler struct {
	client jobv1.JobServiceClient
}

func NewJobHandler(client jobv1.JobServiceClient) *JobHandler {
	return &JobHandler{client: client}
}

func (h *JobHandler) CreateJob(c *gin.Context) {
	var req jobv1.CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.CreateJob(c.Request.Context(), &req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "job", resp.GetJob())
}

func (h *JobHandler) GetJob(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
	if !ok {
		return
	}
	resp, err := h.client.GetJob(c.Request.Context(), &jobv1.GetJobRequest{JobId: jobID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "job", resp.GetJob())
}

func (h *JobHandler) UpdateJob(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
	if !ok {
		return
	}
	var req jobv1.UpdateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.JobId = jobID
	resp, err := h.client.UpdateJob(c.Request.Context(), &req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "job", resp.GetJob())
}

func (h *JobHandler) ListMyJobs(c *gin.Context) {
	statusEnum, ok := mapJobStatusQuery(c.Query("status"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}
	resp, err := h.client.ListMyJobs(c.Request.Context(), &jobv1.ListMyJobsRequest{
		StatusEnum: mapJobStatus(strings.TrimSpace(c.Query("status"))),
		PageSize:   int32(parseIntQuery(c, "page_size", 20)),
		PageToken:  strings.TrimSpace(c.Query("page_token")),
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	payload, convErr := protoSliceToAny(resp.GetJobs())
	if convErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"jobs": payload, "next_page_token": resp.GetNextPageToken()})
}

func (h *JobHandler) ListOpenJobs(c *gin.Context) {
	jobTypeEnum, ok := mapJobTypeQuery(c.Query("job_type"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job_type"})
		return
	}
	resp, err := h.client.ListOpenJobs(c.Request.Context(), &jobv1.ListOpenJobsRequest{
		PageSize:    int32(parseIntQuery(c, "page_size", 20)),
		PageToken:   strings.TrimSpace(c.Query("page_token")),
		SearchQuery: strings.TrimSpace(c.Query("query")),
		Skills:      c.QueryArray("skills"),
		JobTypeEnum: jobTypeEnum,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	payload, convErr := protoSliceToAny(resp.GetJobs())
	if convErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"jobs": payload, "next_page_token": resp.GetNextPageToken()})
}

func (h *JobHandler) SearchJobsV2(c *gin.Context) {
	resp, err := h.client.SearchJobsV2(c.Request.Context(), &jobv1.SearchJobsV2Request{
		Query:     strings.TrimSpace(c.Query("query")),
		Skills:    c.QueryArray("skills"),
		PageSize:  int32(parseIntQuery(c, "page_size", 20)),
		PageToken: strings.TrimSpace(c.Query("page_token")),
		SortBy:    mapSortByQuery(strings.TrimSpace(c.Query("sort_by"))),
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	payload, convErr := protoSliceToAny(resp.GetJobs())
	if convErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"jobs": payload, "next_page_token": resp.GetNextPageToken()})
}

func (h *JobHandler) SetJobVisibility(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
	if !ok {
		return
	}
	var body struct {
		Visibility string `json:"visibility"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	visibility, ok := mapVisibilityBody(body.Visibility)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visibility"})
		return
	}
	resp, err := h.client.SetJobVisibility(c.Request.Context(), &jobv1.SetJobVisibilityRequest{JobId: jobID, Visibility: visibility})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "job", resp.GetJob())
}

func (h *JobHandler) SetJobBudgetRange(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
	if !ok {
		return
	}
	var body struct {
		BudgetMin float64 `json:"budget_min"`
		BudgetMax float64 `json:"budget_max"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.SetJobBudgetRange(c.Request.Context(), &jobv1.SetJobBudgetRangeRequest{JobId: jobID, BudgetMin: body.BudgetMin, BudgetMax: body.BudgetMax})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "job", resp.GetJob())
}

func (h *JobHandler) PauseJob(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
	if !ok {
		return
	}
	resp, err := h.client.PauseJob(c.Request.Context(), &jobv1.PauseJobRequest{JobId: jobID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "job", resp.GetJob())
}

func (h *JobHandler) ReopenJob(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
	if !ok {
		return
	}
	resp, err := h.client.ReopenJob(c.Request.Context(), &jobv1.ReopenJobRequest{JobId: jobID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "job", resp.GetJob())
}

func (h *JobHandler) MarkJobFilled(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
	if !ok {
		return
	}
	resp, err := h.client.MarkJobFilled(c.Request.Context(), &jobv1.MarkJobFilledRequest{JobId: jobID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "job", resp.GetJob())
}

func (h *JobHandler) CloseJob(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
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
	reasonEnum, ok := mapCloseReasonBody(body.Reason)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reason"})
		return
	}
	resp, err := h.client.CloseJob(c.Request.Context(), &jobv1.CloseJobRequest{JobId: jobID, ReasonEnum: reasonEnum})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"closed": resp.GetClosed()})
}

func (h *JobHandler) ListJobApplicants(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
	if !ok {
		return
	}
	resp, err := h.client.ListJobApplicants(c.Request.Context(), &jobv1.ListJobApplicantsRequest{JobId: jobID, PageSize: int32(parseIntQuery(c, "page_size", 20)), PageToken: strings.TrimSpace(c.Query("page_token"))})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	payload, convErr := protoSliceToAny(resp.GetApplicants())
	if convErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"applicants": payload, "next_page_token": resp.GetNextPageToken()})
}

func (h *JobHandler) SetApplicantStage(c *gin.Context) {
	proposalID, ok := parseInt64Param(c, "proposalId")
	if !ok {
		return
	}
	var body struct {
		Stage  string `json:"stage"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.SetApplicantStage(c.Request.Context(), &jobv1.SetApplicantStageRequest{ProposalId: proposalID, Stage: mapApplicantStage(body.Stage), Reason: body.Reason})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated": resp.GetUpdated()})
}

func (h *JobHandler) InviteFreelancerToJob(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
	if !ok {
		return
	}
	var body struct {
		FreelancerID string `json:"freelancer_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.InviteFreelancerToJob(c.Request.Context(), &jobv1.InviteFreelancerToJobRequest{JobId: jobID, FreelancerId: body.FreelancerID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"invited": resp.GetInvited()})
}

func (h *JobHandler) GetJobStats(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
	if !ok {
		return
	}
	resp, err := h.client.GetJobStats(c.Request.Context(), &jobv1.GetJobStatsRequest{JobId: jobID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "stats", resp)
}

func (h *JobHandler) GetPublicJobDetail(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
	if !ok {
		return
	}
	resp, err := h.client.GetPublicJobDetail(c.Request.Context(), &jobv1.GetPublicJobDetailRequest{JobId: jobID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "job", resp.GetJob())
}

func (h *JobHandler) ListInvitedJobs(c *gin.Context) {
	resp, err := h.client.ListInvitedJobs(c.Request.Context(), &jobv1.ListInvitedJobsRequest{PageSize: int32(parseIntQuery(c, "page_size", 20)), PageToken: strings.TrimSpace(c.Query("page_token"))})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	payload, convErr := protoSliceToAny(resp.GetInvites())
	if convErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"jobs": payload, "next_page_token": resp.GetNextPageToken()})
}

func (h *JobHandler) RespondToJobInvite(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
	if !ok {
		return
	}
	var body struct {
		ResponseStatus string `json:"response_status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.RespondToJobInvite(c.Request.Context(), &jobv1.RespondToJobInviteRequest{JobId: jobID, ResponseStatus: mapInviteResponseStatus(body.ResponseStatus)})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated": resp.GetUpdated()})
}

func (h *JobHandler) SaveJob(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
	if !ok {
		return
	}
	resp, err := h.client.SaveJob(c.Request.Context(), &jobv1.SaveJobRequest{JobId: jobID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"saved": resp.GetSaved()})
}

func (h *JobHandler) UnsaveJob(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
	if !ok {
		return
	}
	resp, err := h.client.UnsaveJob(c.Request.Context(), &jobv1.UnsaveJobRequest{JobId: jobID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"removed": resp.GetRemoved()})
}

func (h *JobHandler) ListSavedJobs(c *gin.Context) {
	resp, err := h.client.ListSavedJobs(c.Request.Context(), &jobv1.ListSavedJobsRequest{PageSize: int32(parseIntQuery(c, "page_size", 20)), PageToken: strings.TrimSpace(c.Query("page_token"))})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	payload, convErr := protoSliceToAny(resp.GetJobs())
	if convErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"jobs": payload, "next_page_token": resp.GetNextPageToken()})
}

func (h *JobHandler) RejectAllApplicants(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
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
	resp, err := h.client.RejectAllApplicants(c.Request.Context(), &jobv1.RejectAllApplicantsRequest{JobId: jobID, Reason: body.Reason})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"rejected_count": resp.GetRejectedCount()})
}

func (h *JobHandler) ReopenHiringForJob(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
	if !ok {
		return
	}
	resp, err := h.client.ReopenHiringForJob(c.Request.Context(), &jobv1.ReopenHiringForJobRequest{JobId: jobID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "job", resp.GetJob())
}

func (h *JobHandler) MarkJobCompleted(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
	if !ok {
		return
	}
	resp, err := h.client.MarkJobCompleted(c.Request.Context(), &jobv1.MarkJobCompletedRequest{JobId: jobID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"completed": resp.GetCompleted()})
}

func (h *JobHandler) CancelJobWithSettlementPolicy(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
	if !ok {
		return
	}
	var body struct {
		SettlementPolicy string `json:"settlement_policy"`
		Reason           string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.CancelJobWithSettlementPolicy(c.Request.Context(), &jobv1.CancelJobWithSettlementPolicyRequest{JobId: jobID, SettlementPolicy: mapSettlementPolicy(body.SettlementPolicy), Reason: body.Reason})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"canceled": resp.GetCanceled()})
}

func (h *JobHandler) UploadJobAttachment(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
	if !ok {
		return
	}
	formFile, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	fh, err := formFile.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot open file"})
		return
	}
	defer fh.Close()
	data := make([]byte, formFile.Size)
	if _, err := fh.Read(data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read file"})
		return
	}
	contentType := strings.TrimSpace(c.PostForm("content_type"))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	resp, err := h.client.UploadJobAttachment(c.Request.Context(), &jobv1.UploadJobAttachmentRequest{JobId: jobID, FileName: formFile.Filename, ContentType: contentType, Content: data})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "attachment", resp.GetAttachment())
}

func (h *JobHandler) DeleteJobAttachment(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
	if !ok {
		return
	}
	attachmentID, ok := parseInt64Param(c, "attachmentId")
	if !ok {
		return
	}
	resp, err := h.client.DeleteJobAttachment(c.Request.Context(), &jobv1.DeleteJobAttachmentRequest{JobId: jobID, AttachmentId: attachmentID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": resp.GetDeleted()})
}

func (h *JobHandler) ListJobAttachments(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
	if !ok {
		return
	}
	resp, err := h.client.ListJobAttachments(c.Request.Context(), &jobv1.ListJobAttachmentsRequest{JobId: jobID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	payload, convErr := protoSliceToAny(resp.GetAttachments())
	if convErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"attachments": payload})
}

func (h *JobHandler) GetJobAttachmentDownloadURL(c *gin.Context) {
	jobID, ok := parseInt64Param(c, "jobId")
	if !ok {
		return
	}
	attachmentID, ok := parseInt64Param(c, "attachmentId")
	if !ok {
		return
	}
	resp, err := h.client.GetJobAttachmentDownloadUrl(c.Request.Context(), &jobv1.GetJobAttachmentDownloadUrlRequest{JobId: jobID, AttachmentId: attachmentID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": resp.GetUrl()})
}

func parseInt64Param(c *gin.Context, key string) (int64, bool) {
	v := strings.TrimSpace(c.Param(key))
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

func mapVisibilityBody(in string) (jobv1.Visibility, bool) {
	switch strings.ToLower(strings.TrimSpace(in)) {
	case "":
		return jobv1.Visibility_VISIBILITY_UNSPECIFIED, true
	case "public":
		return jobv1.Visibility_VISIBILITY_PUBLIC, true
	case "private":
		return jobv1.Visibility_VISIBILITY_PRIVATE, true
	case "invite_only":
		return jobv1.Visibility_VISIBILITY_INVITE_ONLY, true
	default:
		return jobv1.Visibility_VISIBILITY_UNSPECIFIED, false
	}
}

func mapJobStatusQuery(in string) (jobv1.JobStatus, bool) {
	switch strings.ToLower(strings.TrimSpace(in)) {
	case "":
		return jobv1.JobStatus_JOB_STATUS_UNSPECIFIED, true
	case "open":
		return jobv1.JobStatus_JOB_STATUS_OPEN, true
	case "paused":
		return jobv1.JobStatus_JOB_STATUS_PAUSED, true
	case "filled":
		return jobv1.JobStatus_JOB_STATUS_FILLED, true
	case "closed":
		return jobv1.JobStatus_JOB_STATUS_CLOSED, true
	case "completed":
		return jobv1.JobStatus_JOB_STATUS_COMPLETED, true
	case "canceled":
		return jobv1.JobStatus_JOB_STATUS_CANCELED, true
	default:
		return jobv1.JobStatus_JOB_STATUS_UNSPECIFIED, false
	}
}

func mapJobTypeQuery(in string) (jobv1.JobType, bool) {
	switch strings.ToLower(strings.TrimSpace(in)) {
	case "":
		return jobv1.JobType_JOB_TYPE_UNSPECIFIED, true
	case "fixed":
		return jobv1.JobType_JOB_TYPE_FIXED, true
	case "hourly":
		return jobv1.JobType_JOB_TYPE_HOURLY, true
	default:
		return jobv1.JobType_JOB_TYPE_UNSPECIFIED, false
	}
}

func mapCloseReasonBody(in string) (jobv1.CloseReason, bool) {
	switch strings.ToLower(strings.TrimSpace(in)) {
	case "":
		return jobv1.CloseReason_CLOSE_REASON_UNSPECIFIED, true
	case "canceled":
		return jobv1.CloseReason_CLOSE_REASON_CANCELED, true
	default:
		return jobv1.CloseReason_CLOSE_REASON_UNSPECIFIED, false
	}
}

func mapApplicantStage(in string) jobv1.ApplicantStage {
	switch strings.ToLower(strings.TrimSpace(in)) {
	case "shortlist":
		fallthrough
	case "shortlisted":
		return jobv1.ApplicantStage_APPLICANT_STAGE_SHORTLISTED
	case "reject":
		fallthrough
	case "rejected":
		return jobv1.ApplicantStage_APPLICANT_STAGE_REJECTED
	case "hired":
		return jobv1.ApplicantStage_APPLICANT_STAGE_HIRED
	default:
		return jobv1.ApplicantStage_APPLICANT_STAGE_SENT
	}
}

func mapInviteResponseStatus(in string) jobv1.InviteResponseStatus {
	switch strings.ToLower(strings.TrimSpace(in)) {
	case "accepted":
		return jobv1.InviteResponseStatus_INVITE_RESPONSE_STATUS_ACCEPTED
	case "declined":
		return jobv1.InviteResponseStatus_INVITE_RESPONSE_STATUS_DECLINED
	default:
		return jobv1.InviteResponseStatus_INVITE_RESPONSE_STATUS_UNSPECIFIED
	}
}

func mapSortByQuery(in string) jobv1.JobSortBy {
	switch strings.ToLower(strings.TrimSpace(in)) {
	case "newest":
		return jobv1.JobSortBy_JOB_SORT_BY_NEWEST
	case "oldest":
		return jobv1.JobSortBy_JOB_SORT_BY_OLDEST
	case "budget_high":
		return jobv1.JobSortBy_JOB_SORT_BY_BUDGET_HIGH
	case "budget_low":
		return jobv1.JobSortBy_JOB_SORT_BY_BUDGET_LOW
	default:
		return jobv1.JobSortBy_JOB_SORT_BY_RELEVANCE
	}
}

func mapSettlementPolicy(in string) jobv1.SettlementPolicy {
	switch strings.ToLower(strings.TrimSpace(in)) {
	case "no_refund":
		return jobv1.SettlementPolicy_SETTLEMENT_POLICY_NO_REFUND
	default:
		return jobv1.SettlementPolicy_SETTLEMENT_POLICY_REFUND_REMAINING
	}
}

func callerRole(c *gin.Context) (string, bool) {
	v, ok := c.Get(middleware.ContextRole)
	if !ok {
		return "", false
	}
	role, ok := v.(string)
	if !ok || strings.TrimSpace(role) == "" {
		return "", false
	}
	return role, true
}
