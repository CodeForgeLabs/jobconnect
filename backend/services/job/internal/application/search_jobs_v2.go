package application

import (
	"context"
	"strconv"
	"strings"

	"jobconnect/job/internal/domain"
)

type SearchJobsV2 struct {
	Jobs JobRepository
}

type SearchJobsV2Input struct {
	PageSize        int32
	PageToken       string
	Query           string
	Skills          []string
	JobType         string
	Visibility      string
	ExperienceLevel string
	SortBy          string
}

type SearchJobsV2Output struct {
	Jobs          []domain.Job
	NextPageToken string
}

func (uc *SearchJobsV2) Execute(ctx context.Context, in SearchJobsV2Input) (SearchJobsV2Output, error) {
	limit := normalizePageSize(in.PageSize)
	offset, err := parsePageToken(in.PageToken)
	if err != nil {
		return SearchJobsV2Output{}, err
	}
	jobType := strings.ToLower(strings.TrimSpace(in.JobType))
	if err := domain.ValidateJobType(jobType); err != nil && jobType != "" {
		return SearchJobsV2Output{}, err
	}
	visibility := strings.ToLower(strings.TrimSpace(in.Visibility))
	if err := domain.ValidateVisibility(visibility); err != nil {
		return SearchJobsV2Output{}, err
	}
	level := strings.ToLower(strings.TrimSpace(in.ExperienceLevel))
	if err := domain.ValidateExperienceLevel(level); err != nil {
		return SearchJobsV2Output{}, err
	}
	jobs, err := uc.Jobs.ListOpenFilteredV2(ctx, ListOpenFilter{
		SearchQuery:     strings.TrimSpace(in.Query),
		Skills:          in.Skills,
		JobType:         jobType,
		Visibility:      visibility,
		ExperienceLevel: level,
		Limit:           limit,
		Offset:          offset,
	}, strings.ToLower(strings.TrimSpace(in.SortBy)))
	if err != nil {
		return SearchJobsV2Output{}, err
	}
	next := ""
	if len(jobs) == limit {
		next = strconv.Itoa(offset + len(jobs))
	}
	return SearchJobsV2Output{Jobs: jobs, NextPageToken: next}, nil
}
