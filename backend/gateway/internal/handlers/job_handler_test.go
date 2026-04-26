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

func TestMapVisibility(t *testing.T) {
	if got := mapVisibility("private"); got != jobv1.Visibility_VISIBILITY_PRIVATE {
		t.Fatalf("expected private visibility enum, got %v", got)
	}
	if got := mapVisibility("invite_only"); got != jobv1.Visibility_VISIBILITY_INVITE_ONLY {
		t.Fatalf("expected invite_only visibility enum, got %v", got)
	}
	if got := mapVisibility("whatever"); got != jobv1.Visibility_VISIBILITY_PUBLIC {
		t.Fatalf("expected fallback public visibility enum, got %v", got)
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

func TestMapApplicantStageAliases(t *testing.T) {
	if got := mapApplicantStage("shortlist"); got != jobv1.ApplicantStage_APPLICANT_STAGE_SHORTLISTED {
		t.Fatalf("expected shortlist alias to map to shortlisted, got %v", got)
	}
	if got := mapApplicantStage("shortlisted"); got != jobv1.ApplicantStage_APPLICANT_STAGE_SHORTLISTED {
		t.Fatalf("expected shortlisted to map to shortlisted, got %v", got)
	}
	if got := mapApplicantStage("reject"); got != jobv1.ApplicantStage_APPLICANT_STAGE_REJECTED {
		t.Fatalf("expected reject alias to map to rejected, got %v", got)
	}
	if got := mapApplicantStage("rejected"); got != jobv1.ApplicantStage_APPLICANT_STAGE_REJECTED {
		t.Fatalf("expected rejected to map to rejected, got %v", got)
	}
}
