package application

import (
	"context"
	"fmt"
	"strings"

	"jobconnect/job/internal/domain"

	"github.com/google/uuid"
)

type CloseJob struct {
	Jobs      JobRepository
	Proposals ProposalClient
	Connects  ConnectsClient
	Clock     Clock
}

type CloseJobInput struct {
	JobID    int64
	ClientID uuid.UUID
	Reason   string
}

type CloseJobOutput struct {
	Closed bool
}

func (uc *CloseJob) Execute(ctx context.Context, in CloseJobInput) (CloseJobOutput, error) {
	if in.JobID <= 0 {
		return CloseJobOutput{}, fmt.Errorf("job_id is required")
	}
	if in.ClientID == uuid.Nil {
		return CloseJobOutput{}, fmt.Errorf("client_id is required")
	}
	if err := domain.ValidateCloseReason(in.Reason); err != nil {
		return CloseJobOutput{}, err
	}
	normalizedReason := strings.ToLower(strings.TrimSpace(in.Reason))

	err := uc.Jobs.Close(ctx, in.JobID, in.ClientID, normalizedReason, uc.Clock.Now())
	if err != nil {
		return CloseJobOutput{}, err
	}

	// Refund Connects (Best effort refund loop for MVP)
	// In production, this might be handled via an async event/message queue
	// to avoid blocking the client request and ensure strict retry semantics on the consumer side.
	if normalizedReason == domain.CloseReasonCanceled {
		proposals, err := uc.Proposals.ListProposalsByJob(ctx, in.JobID)
		if err == nil {
			for _, p := range proposals {
				if p.ConnectsSpent > 0 {
					refID := fmt.Sprintf("job_canceled_%d_proposal_%d", in.JobID, p.ID)
					// We ignore errors on individual refunds for MVP to not fail the loop.
					// A real implementation would queue these.
					_ = uc.Connects.RefundConnects(ctx, p.FreelancerID, p.ConnectsSpent, refID)
				}
			}
		} else {
			// Log error, but job is already closed
			fmt.Printf("failed to fetch proposals for refunds: %v\n", err)
		}
	}

	return CloseJobOutput{Closed: true}, nil
}
