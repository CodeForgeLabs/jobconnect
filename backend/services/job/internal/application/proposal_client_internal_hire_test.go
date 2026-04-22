package application

import (
	"context"

	"github.com/google/uuid"
)

func (p *fakeProposalClient) InternalHireProposal(ctx context.Context, proposalID int64, clientID uuid.UUID, requestID string, reason string) error {
	if p.internalHireFn != nil {
		return p.internalHireFn(ctx, proposalID, clientID, requestID, reason)
	}
	return nil
}
