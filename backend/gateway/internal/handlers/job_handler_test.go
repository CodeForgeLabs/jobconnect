package handlers

import (
	"net/http"
	"testing"

	jobv1 "jobconnect/job/gen/job/v1"

	"github.com/gin-gonic/gin"
)

func TestGetJob_InvalidJobID_ReturnsBadRequest(t *testing.T) {
	h := &JobHandler{}
	ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/jobs/nope")
	ctx.Params = gin.Params{{Key: "jobId", Value: "nope"}}

	h.GetJob(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestUpdateJob_MissingJobID_ReturnsBadRequest(t *testing.T) {
	h := &JobHandler{}
	ctx, rec := newJSONTestContext(http.MethodPatch, "/api/v1/jobs/")
	ctx.Params = gin.Params{{Key: "jobId", Value: ""}}

	h.UpdateJob(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestDeleteJobAttachment_InvalidAttachmentID_ReturnsBadRequest(t *testing.T) {
	h := &JobHandler{}
	ctx, rec := newJSONTestContext(http.MethodDelete, "/api/v1/jobs/1/attachments/x")
	ctx.Params = gin.Params{{Key: "jobId", Value: "1"}, {Key: "attachmentId", Value: "x"}}

	h.DeleteJobAttachment(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestUploadJobAttachment_MissingFile_ReturnsBadRequest(t *testing.T) {
	h := &JobHandler{}
	ctx, rec := newJSONTestContext(http.MethodPost, "/api/v1/jobs/1/attachments")
	ctx.Params = gin.Params{{Key: "jobId", Value: "1"}}

	h.UploadJobAttachment(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestMapSortByQuery(t *testing.T) {
	cases := []struct {
		in   string
		want jobv1.JobSortBy
	}{
		{in: "newest", want: jobv1.JobSortBy_JOB_SORT_BY_NEWEST},
		{in: "oldest", want: jobv1.JobSortBy_JOB_SORT_BY_OLDEST},
		{in: "budget_high", want: jobv1.JobSortBy_JOB_SORT_BY_BUDGET_HIGH},
		{in: "budget_low", want: jobv1.JobSortBy_JOB_SORT_BY_BUDGET_LOW},
		{in: "anything", want: jobv1.JobSortBy_JOB_SORT_BY_RELEVANCE},
	}

	for _, tc := range cases {
		got := mapSortByQuery(tc.in)
		if got != tc.want {
			t.Fatalf("mapSortByQuery(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestMapVisibilityBody(t *testing.T) {
	cases := []struct {
		in     string
		want   jobv1.Visibility
		wantOK bool
	}{
		{in: "", want: jobv1.Visibility_VISIBILITY_UNSPECIFIED, wantOK: true},
		{in: "public", want: jobv1.Visibility_VISIBILITY_PUBLIC, wantOK: true},
		{in: "private", want: jobv1.Visibility_VISIBILITY_PRIVATE, wantOK: true},
		{in: "invite_only", want: jobv1.Visibility_VISIBILITY_INVITE_ONLY, wantOK: true},
		{in: "whatever", want: jobv1.Visibility_VISIBILITY_UNSPECIFIED, wantOK: false},
	}
	for _, tc := range cases {
		got, ok := mapVisibilityBody(tc.in)
		if got != tc.want || ok != tc.wantOK {
			t.Fatalf("mapVisibilityBody(%q) = (%v, %v), want (%v, %v)", tc.in, got, ok, tc.want, tc.wantOK)
		}
	}
}

func TestMapJobStatusQuery(t *testing.T) {
	cases := []struct {
		in     string
		want   jobv1.JobStatus
		wantOK bool
	}{
		{in: "", want: jobv1.JobStatus_JOB_STATUS_UNSPECIFIED, wantOK: true},
		{in: "open", want: jobv1.JobStatus_JOB_STATUS_OPEN, wantOK: true},
		{in: "paused", want: jobv1.JobStatus_JOB_STATUS_PAUSED, wantOK: true},
		{in: "filled", want: jobv1.JobStatus_JOB_STATUS_FILLED, wantOK: true},
		{in: "closed", want: jobv1.JobStatus_JOB_STATUS_CLOSED, wantOK: true},
		{in: "completed", want: jobv1.JobStatus_JOB_STATUS_COMPLETED, wantOK: true},
		{in: "canceled", want: jobv1.JobStatus_JOB_STATUS_CANCELED, wantOK: true},
		{in: "nope", want: jobv1.JobStatus_JOB_STATUS_UNSPECIFIED, wantOK: false},
	}
	for _, tc := range cases {
		got, ok := mapJobStatusQuery(tc.in)
		if got != tc.want || ok != tc.wantOK {
			t.Fatalf("mapJobStatusQuery(%q) = (%v, %v), want (%v, %v)", tc.in, got, ok, tc.want, tc.wantOK)
		}
	}
}

func TestMapJobTypeQuery(t *testing.T) {
	cases := []struct {
		in     string
		want   jobv1.JobType
		wantOK bool
	}{
		{in: "", want: jobv1.JobType_JOB_TYPE_UNSPECIFIED, wantOK: true},
		{in: "fixed", want: jobv1.JobType_JOB_TYPE_FIXED, wantOK: true},
		{in: "hourly", want: jobv1.JobType_JOB_TYPE_HOURLY, wantOK: true},
		{in: "contract", want: jobv1.JobType_JOB_TYPE_UNSPECIFIED, wantOK: false},
	}
	for _, tc := range cases {
		got, ok := mapJobTypeQuery(tc.in)
		if got != tc.want || ok != tc.wantOK {
			t.Fatalf("mapJobTypeQuery(%q) = (%v, %v), want (%v, %v)", tc.in, got, ok, tc.want, tc.wantOK)
		}
	}
}

func TestMapCloseReasonBody(t *testing.T) {
	cases := []struct {
		in     string
		want   jobv1.CloseReason
		wantOK bool
	}{
		{in: "", want: jobv1.CloseReason_CLOSE_REASON_UNSPECIFIED, wantOK: true},
		{in: "canceled", want: jobv1.CloseReason_CLOSE_REASON_CANCELED, wantOK: true},
		{in: "bogus", want: jobv1.CloseReason_CLOSE_REASON_UNSPECIFIED, wantOK: false},
	}
	for _, tc := range cases {
		got, ok := mapCloseReasonBody(tc.in)
		if got != tc.want || ok != tc.wantOK {
			t.Fatalf("mapCloseReasonBody(%q) = (%v, %v), want (%v, %v)", tc.in, got, ok, tc.want, tc.wantOK)
		}
	}
}

func TestListMyJobs_InvalidStatus_ReturnsBadRequest(t *testing.T) {
	h := &JobHandler{}
	ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/jobs/me?status=garbage")

	h.ListMyJobs(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestListOpenJobs_InvalidJobType_ReturnsBadRequest(t *testing.T) {
	h := &JobHandler{}
	ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/jobs?job_type=garbage")

	h.ListOpenJobs(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestCloseJob_InvalidReason_ReturnsBadRequest(t *testing.T) {
	h := &JobHandler{}
	ctx, rec := newJSONBodyTestContext(http.MethodPost, "/api/v1/jobs/1/close", `{"reason":"nope"}`)
	ctx.Params = gin.Params{{Key: "jobId", Value: "1"}}

	h.CloseJob(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSetJobVisibility_InvalidValue_ReturnsBadRequest(t *testing.T) {
	h := &JobHandler{}
	ctx, rec := newJSONBodyTestContext(http.MethodPost, "/api/v1/jobs/1/visibility", `{"visibility":"nope"}`)
	ctx.Params = gin.Params{{Key: "jobId", Value: "1"}}

	h.SetJobVisibility(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestMapSettlementPolicy(t *testing.T) {
	if got := mapSettlementPolicy("no_refund"); got != jobv1.SettlementPolicy_SETTLEMENT_POLICY_NO_REFUND {
		t.Fatalf("expected no_refund settlement enum, got %v", got)
	}
	if got := mapSettlementPolicy("other"); got != jobv1.SettlementPolicy_SETTLEMENT_POLICY_REFUND_REMAINING {
		t.Fatalf("expected fallback refund_remaining settlement enum, got %v", got)
	}
}
