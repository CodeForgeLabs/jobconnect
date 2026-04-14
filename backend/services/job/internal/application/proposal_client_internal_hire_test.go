package application

import (
	"context"

	"github.com/google/uuid"
)

func (p *fakeProposalClient) InternalHireProposal(ctx context.Context, proposalID int64, clientID uuid.UUID, requestID string, reason string) error {
	return nil
}
