package application

import (
	"context"
	"fmt"
	"strconv"

	"jobconnect/job/internal/domain"
)

type ListOpenJobs struct {
	Jobs JobRepository
}

type ListOpenJobsInput struct {
	PageSize  int32
	PageToken string
}

type ListOpenJobsOutput struct {
	Jobs          []domain.Job
	NextPageToken string
}

func (uc *ListOpenJobs) Execute(ctx context.Context, in ListOpenJobsInput) (ListOpenJobsOutput, error) {
	limit := normalizePageSize(in.PageSize)
	offset, err := parsePageToken(in.PageToken)
	if err != nil {
		return ListOpenJobsOutput{}, err
	}

	jobs, err := uc.Jobs.ListOpen(ctx, limit, offset)
	if err != nil {
		return ListOpenJobsOutput{}, err
	}

	next := ""
	if len(jobs) == limit {
		next = strconv.Itoa(offset + len(jobs))
	}

	return ListOpenJobsOutput{Jobs: jobs, NextPageToken: next}, nil
}

func ensureOpenOnlyStatus(status string) error {
	if status != "" && status != domain.JobStatusOpen {
		return fmt.Errorf("invalid status")
	}
	return nil
}
