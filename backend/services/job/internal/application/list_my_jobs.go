package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"jobconnect/job/internal/domain"

	"github.com/google/uuid"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

type ListMyJobs struct {
	Jobs JobRepository
}

type ListMyJobsInput struct {
	ClientID  uuid.UUID
	Status    string
	PageSize  int32
	PageToken string
}

type ListMyJobsOutput struct {
	Jobs          []domain.Job
	NextPageToken string
}

func (uc *ListMyJobs) Execute(ctx context.Context, in ListMyJobsInput) (ListMyJobsOutput, error) {
	if in.ClientID == uuid.Nil {
		return ListMyJobsOutput{}, fmt.Errorf("client_id is required")
	}
	status, err := domain.ValidateStatus(in.Status)
	if err != nil {
		return ListMyJobsOutput{}, err
	}
	limit := normalizePageSize(in.PageSize)
	offset, err := parsePageToken(in.PageToken)
	if err != nil {
		return ListMyJobsOutput{}, err
	}

	jobs, err := uc.Jobs.ListByClient(ctx, in.ClientID, status, limit, offset)
	if err != nil {
		return ListMyJobsOutput{}, err
	}

	next := ""
	if len(jobs) == limit {
		next = strconv.Itoa(offset + len(jobs))
	}

	return ListMyJobsOutput{Jobs: jobs, NextPageToken: next}, nil
}

func normalizePageSize(in int32) int {
	if in <= 0 {
		return defaultPageSize
	}
	if in > maxPageSize {
		return maxPageSize
	}
	return int(in)
}

func parsePageToken(token string) (int, error) {
	if strings.TrimSpace(token) == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(token)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("invalid page_token")
	}
	return n, nil
}
