package application

import (
	"context"
	"fmt"
	"strings"

	"jobconnect/job/internal/domain"

	"github.com/google/uuid"
)

type CloseJob struct {
	Jobs  JobRepository
	Clock Clock
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

	err := uc.Jobs.Close(ctx, in.JobID, in.ClientID, strings.TrimSpace(in.Reason), uc.Clock.Now())
	if err != nil {
		return CloseJobOutput{}, err
	}
	return CloseJobOutput{Closed: true}, nil
}
