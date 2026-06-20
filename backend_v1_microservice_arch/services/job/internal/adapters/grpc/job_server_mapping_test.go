package grpcadapter

import (
	"testing"

	jobv1 "jobconnect/job/gen/job/v1"
	"jobconnect/job/internal/application"
)

func TestInviteResponseFromEnum(t *testing.T) {
	got, err := inviteResponseFromEnum(jobv1.InviteResponseStatus_INVITE_RESPONSE_STATUS_ACCEPTED)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != application.InviteResponseAccepted {
		t.Fatalf("expected %q, got %q", application.InviteResponseAccepted, got)
	}

	got, err = inviteResponseFromEnum(jobv1.InviteResponseStatus_INVITE_RESPONSE_STATUS_DECLINED)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != application.InviteResponseDeclined {
		t.Fatalf("expected %q, got %q", application.InviteResponseDeclined, got)
	}

	if _, err := inviteResponseFromEnum(jobv1.InviteResponseStatus_INVITE_RESPONSE_STATUS_UNSPECIFIED); err == nil {
		t.Fatalf("expected error for unspecified invite response")
	}
}

func TestSortByFromEnum(t *testing.T) {
	cases := []struct {
		in   jobv1.JobSortBy
		want string
	}{
		{in: jobv1.JobSortBy_JOB_SORT_BY_RELEVANCE, want: "relevance"},
		{in: jobv1.JobSortBy_JOB_SORT_BY_NEWEST, want: "newest"},
		{in: jobv1.JobSortBy_JOB_SORT_BY_OLDEST, want: "oldest"},
		{in: jobv1.JobSortBy_JOB_SORT_BY_BUDGET_HIGH, want: "budget_high"},
		{in: jobv1.JobSortBy_JOB_SORT_BY_BUDGET_LOW, want: "budget_low"},
	}

	for _, tc := range cases {
		got := sortByFromEnum(tc.in)
		if got != tc.want {
			t.Fatalf("sortByFromEnum(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSettlementPolicyFromEnum(t *testing.T) {
	got, err := settlementPolicyFromEnum(jobv1.SettlementPolicy_SETTLEMENT_POLICY_REFUND_REMAINING)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != application.SettlementPolicyRefundRemaining {
		t.Fatalf("expected %q, got %q", application.SettlementPolicyRefundRemaining, got)
	}

	got, err = settlementPolicyFromEnum(jobv1.SettlementPolicy_SETTLEMENT_POLICY_NO_REFUND)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != application.SettlementPolicyNoRefund {
		t.Fatalf("expected %q, got %q", application.SettlementPolicyNoRefund, got)
	}

	if _, err := settlementPolicyFromEnum(jobv1.SettlementPolicy_SETTLEMENT_POLICY_UNSPECIFIED); err == nil {
		t.Fatalf("expected error for unspecified settlement policy")
	}
}
