package jobgrpc

import (
	"context"
	"jobconnect/contract/internal/application"

	"github.com/google/uuid"
)

type NoopJobClient struct{}

func NewNoopJobClient() *NoopJobClient {
	return &NoopJobClient{}
}

func (c *NoopJobClient) GetSummary(_ context.Context, jobID int64, _ uuid.UUID) (application.JobSummary, error) {
	return application.JobSummary{JobID: jobID, Found: true, IsOpen: true}, nil
}

func (c *NoopJobClient) SetInProgress(_ context.Context, _ int64, _ uuid.UUID) error {
	return nil
}
