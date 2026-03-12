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
	PageSize    int32
	PageToken   string
	SearchQuery string
	Skills      []string
	JobType     string
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

	hasFilter := in.SearchQuery != "" || len(in.Skills) > 0 || in.JobType != ""

	var jobs []domain.Job
	if hasFilter {
		jobs, err = uc.Jobs.ListOpenFiltered(ctx, ListOpenFilter{
			SearchQuery: in.SearchQuery,
			Skills:      in.Skills,
			JobType:     in.JobType,
			Limit:       limit,
			Offset:      offset,
		})
	} else {
		jobs, err = uc.Jobs.ListOpen(ctx, limit, offset)
	}
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
