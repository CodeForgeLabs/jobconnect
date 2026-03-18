package application

import (
	"context"

	"github.com/google/uuid"
)

type ListSessionsInput struct {
	UserID uuid.UUID
}

type ListSessionsOutput struct {
	Sessions []SessionSummary
}

type ListSessions struct {
	Sessions SessionRepository
}

func (uc *ListSessions) Execute(ctx context.Context, in ListSessionsInput) (ListSessionsOutput, error) {
	rows, err := uc.Sessions.ListByUserID(ctx, in.UserID)
	if err != nil {
		return ListSessionsOutput{}, err
	}
	return ListSessionsOutput{Sessions: rows}, nil
}
