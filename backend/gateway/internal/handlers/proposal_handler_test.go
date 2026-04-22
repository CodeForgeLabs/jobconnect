package handlers

import (
	"testing"

	proposalv1 "jobconnect/proposal/gen/proposal/v1"
)

func TestMapClientDecisionAliases(t *testing.T) {
	if got := mapClientDecision("shortlist"); got != proposalv1.ClientDecision_CLIENT_DECISION_SHORTLISTED {
		t.Fatalf("expected shortlist alias to map to shortlisted, got %v", got)
	}
	if got := mapClientDecision("shortlisted"); got != proposalv1.ClientDecision_CLIENT_DECISION_SHORTLISTED {
		t.Fatalf("expected shortlisted to map to shortlisted, got %v", got)
	}
	if got := mapClientDecision("reject"); got != proposalv1.ClientDecision_CLIENT_DECISION_REJECTED {
		t.Fatalf("expected reject alias to map to rejected, got %v", got)
	}
	if got := mapClientDecision("rejected"); got != proposalv1.ClientDecision_CLIENT_DECISION_REJECTED {
		t.Fatalf("expected rejected to map to rejected, got %v", got)
	}
	if got := mapClientDecision("something-else"); got != proposalv1.ClientDecision_CLIENT_DECISION_UNSPECIFIED {
		t.Fatalf("expected unknown value to map to unspecified, got %v", got)
	}
}

func TestMapProposalStatusAliases(t *testing.T) {
	if got := mapProposalStatus("shortlist"); got != proposalv1.ProposalStatus_PROPOSAL_STATUS_SHORTLISTED {
		t.Fatalf("expected shortlist alias to map to shortlisted, got %v", got)
	}
	if got := mapProposalStatus("reject"); got != proposalv1.ProposalStatus_PROPOSAL_STATUS_REJECTED {
		t.Fatalf("expected reject alias to map to rejected, got %v", got)
	}
}
