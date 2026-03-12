package jobgrpc

import (
	"context"

	"github.com/google/uuid"
)

type NoopJobClient struct{}

func NewNoopJobClient() *NoopJobClient {
	return &NoopJobClient{}
}

func (c *NoopJobClient) SetInProgress(_ context.Context, _ int64, _ uuid.UUID) error {
	return nil
}
