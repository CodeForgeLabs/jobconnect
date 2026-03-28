package application

import (
	"context"
	"strconv"
	"strings"

	"jobconnect/job/internal/domain"
)

type SearchJobs struct {
	Jobs JobRepository
}

type SearchJobsInput struct {
	PageSize        int32
	PageToken       string
	Query           string
	Skills          []string
	JobType         string
	Visibility      string
	ExperienceLevel string
}

type SearchJobsOutput struct {
	Jobs          []domain.Job
	NextPageToken string
}

func (uc *SearchJobs) Execute(ctx context.Context, in SearchJobsInput) (SearchJobsOutput, error) {
	limit := normalizePageSize(in.PageSize)
	offset, err := parsePageToken(in.PageToken)
	if err != nil {
		return SearchJobsOutput{}, err
	}
	jobType := strings.ToLower(strings.TrimSpace(in.JobType))
	if err := domain.ValidateJobType(jobType); err != nil && jobType != "" {
		return SearchJobsOutput{}, err
	}
	visibility := strings.ToLower(strings.TrimSpace(in.Visibility))
	if err := domain.ValidateVisibility(visibility); err != nil {
		return SearchJobsOutput{}, err
	}
	level := strings.ToLower(strings.TrimSpace(in.ExperienceLevel))
	if err := domain.ValidateExperienceLevel(level); err != nil {
		return SearchJobsOutput{}, err
	}

	jobs, err := uc.Jobs.ListOpenFiltered(ctx, ListOpenFilter{
		SearchQuery:     strings.TrimSpace(in.Query),
		Skills:          in.Skills,
		JobType:         jobType,
		Visibility:      visibility,
		ExperienceLevel: level,
		Limit:           limit,
		Offset:          offset,
	})
	if err != nil {
		return SearchJobsOutput{}, err
	}

	next := ""
	if len(jobs) == limit {
		next = strconv.Itoa(offset + len(jobs))
	}
	return SearchJobsOutput{Jobs: jobs, NextPageToken: next}, nil
}

type FacetBucket struct {
	Value string
	Count int32
}

type ListJobFacets struct {
	Jobs JobRepository
}

type ListJobFacetsInput struct {
	Query string
}

type ListJobFacetsOutput struct {
	Skills           []FacetBucket
	JobTypes         []FacetBucket
	ExperienceLevels []FacetBucket
	Visibility       []FacetBucket
	Status           []FacetBucket
	Total            int64
}

func (uc *ListJobFacets) Execute(ctx context.Context, in ListJobFacetsInput) (ListJobFacetsOutput, error) {
	jobs, err := uc.Jobs.ListOpenFiltered(ctx, ListOpenFilter{
		SearchQuery: strings.TrimSpace(in.Query),
		Limit:       500,
		Offset:      0,
	})
	if err != nil {
		return ListJobFacetsOutput{}, err
	}

	skills := map[string]int32{}
	types := map[string]int32{}
	levels := map[string]int32{}
	vis := map[string]int32{}
	status := map[string]int32{}

	for _, j := range jobs {
		types[j.JobType]++
		levels[j.ExperienceLevel]++
		vis[j.Visibility]++
		status[j.Status]++
		for _, s := range j.RequiredSkills {
			k := strings.TrimSpace(s)
			if k == "" {
				continue
			}
			skills[k]++
		}
	}

	return ListJobFacetsOutput{
		Skills:           mapToFacetBuckets(skills),
		JobTypes:         mapToFacetBuckets(types),
		ExperienceLevels: mapToFacetBuckets(levels),
		Visibility:       mapToFacetBuckets(vis),
		Status:           mapToFacetBuckets(status),
		Total:            int64(len(jobs)),
	}, nil
}

func mapToFacetBuckets(src map[string]int32) []FacetBucket {
	out := make([]FacetBucket, 0, len(src))
	for k, v := range src {
		if strings.TrimSpace(k) == "" {
			continue
		}
		out = append(out, FacetBucket{Value: k, Count: v})
	}
	return out
}
