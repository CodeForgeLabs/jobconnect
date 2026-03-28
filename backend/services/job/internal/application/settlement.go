package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const (
	SettlementPolicyRefundRemaining = "refund_remaining"
	SettlementPolicyNoRefund        = "no_refund"
)

type MarkJobCompleted struct {
	Jobs  JobRepository
	Clock Clock
}

type MarkJobCompletedInput struct {
	JobID    int64
	ClientID uuid.UUID
}

type MarkJobCompletedOutput struct {
	Completed bool
}

func (uc *MarkJobCompleted) Execute(ctx context.Context, in MarkJobCompletedInput) (MarkJobCompletedOutput, error) {
	if in.JobID <= 0 {
		return MarkJobCompletedOutput{}, fmt.Errorf("job_id is required")
	}
	if in.ClientID == uuid.Nil {
		return MarkJobCompletedOutput{}, fmt.Errorf("client_id is required")
	}
	completed, err := uc.Jobs.MarkJobCompleted(ctx, in.JobID, in.ClientID, uc.Clock.Now())
	if err != nil {
		return MarkJobCompletedOutput{}, err
	}
	return MarkJobCompletedOutput{Completed: completed}, nil
}

type CancelJobWithSettlementPolicy struct {
	Jobs      JobRepository
	Proposals ProposalClient
	Connects  ConnectsClient
	Clock     Clock
}

type CancelJobWithSettlementPolicyInput struct {
	JobID            int64
	ClientID         uuid.UUID
	SettlementPolicy string
	Reason           string
}

type CancelJobWithSettlementPolicyOutput struct {
	Canceled bool
}

func (uc *CancelJobWithSettlementPolicy) Execute(ctx context.Context, in CancelJobWithSettlementPolicyInput) (CancelJobWithSettlementPolicyOutput, error) {
	if in.JobID <= 0 {
		return CancelJobWithSettlementPolicyOutput{}, fmt.Errorf("job_id is required")
	}
	if in.ClientID == uuid.Nil {
		return CancelJobWithSettlementPolicyOutput{}, fmt.Errorf("client_id is required")
	}
	policy := strings.ToLower(strings.TrimSpace(in.SettlementPolicy))
	if policy != SettlementPolicyRefundRemaining && policy != SettlementPolicyNoRefund {
		return CancelJobWithSettlementPolicyOutput{}, fmt.Errorf("invalid settlement_policy")
	}
	reason := strings.TrimSpace(in.Reason)
	if reason == "" {
		reason = "canceled"
	}
	canceled, err := uc.Jobs.CancelJobWithSettlement(ctx, in.JobID, in.ClientID, policy, reason, uc.Clock.Now())
	if err != nil {
		return CancelJobWithSettlementPolicyOutput{}, err
	}

	if canceled && policy == SettlementPolicyRefundRemaining {
		proposals, propErr := uc.Proposals.ListProposalsByJob(ctx, in.JobID)
		if propErr == nil {
			for _, p := range proposals {
				if p.ConnectsSpent <= 0 {
					continue
				}
				refID := fmt.Sprintf("job_cancel_settle_%d_proposal_%d", in.JobID, p.ID)
				_ = uc.Connects.RefundConnects(ctx, p.FreelancerID, p.ConnectsSpent, refID)
			}
		}
	}

	return CancelJobWithSettlementPolicyOutput{Canceled: canceled}, nil
}
