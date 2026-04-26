package jobgrpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"

	jobv1 "jobconnect/job/gen/job/v1"
	"jobconnect/recommendation/internal/domain"
)

const (
	visibilityPublic = "public"
	jobTypeFixed     = "fixed"
	jobTypeHourly    = "hourly"
)

type Client struct {
	grpcClient jobv1.JobServiceClient
}

func NewClient(conn grpc.ClientConnInterface) *Client {
	return &Client{grpcClient: jobv1.NewJobServiceClient(conn)}
}

func (c *Client) ListRecentPublicOpenJobs(ctx context.Context, pageSize int32) ([]domain.JobData, error) {
	resp, err := c.grpcClient.SearchJobsV2(ctx, &jobv1.SearchJobsV2Request{
		PageSize:   pageSize,
		Visibility: jobv1.Visibility_VISIBILITY_PUBLIC,
		SortBy:     jobv1.JobSortBy_JOB_SORT_BY_NEWEST,
	})
	if err != nil {
		return nil, err
	}
	return mapJobs(resp.Jobs), nil
}

func (c *Client) GetJob(ctx context.Context, jobID int64) (domain.JobData, error) {
	resp, err := c.grpcClient.GetJob(ctx, &jobv1.GetJobRequest{JobId: jobID})
	if err != nil {
		return domain.JobData{}, err
	}
	job := resp.GetJob()
	if job == nil {
		return domain.JobData{}, fmt.Errorf("job %d not found", jobID)
	}
	return mapJob(job), nil
}

func (c *Client) SearchPublicOpenJobsBySkill(ctx context.Context, skill string, pageSize int32) ([]domain.JobData, error) {
	resp, err := c.grpcClient.SearchJobsV2(ctx, &jobv1.SearchJobsV2Request{
		PageSize:   pageSize,
		Visibility: jobv1.Visibility_VISIBILITY_PUBLIC,
		SortBy:     jobv1.JobSortBy_JOB_SORT_BY_NEWEST,
		Skills:     []string{strings.TrimSpace(skill)},
	})
	if err != nil {
		return nil, err
	}
	return mapJobs(resp.Jobs), nil
}

func mapJobs(jobs []*jobv1.Job) []domain.JobData {
	out := make([]domain.JobData, 0, len(jobs))
	for _, job := range jobs {
		if job == nil {
			continue
		}
		out = append(out, mapJob(job))
	}
	return out
}

func mapJob(job *jobv1.Job) domain.JobData {
	return domain.JobData{
		ID:             job.Id,
		ClientID:       job.ClientId,
		Title:          job.Title,
		Description:    job.Description,
		RequiredSkills: append([]string(nil), job.RequiredSkills...),
		BudgetMin:      job.BudgetMin,
		BudgetMax:      job.BudgetMax,
		HourlyRate:     job.HourlyRate,
		JobType:        mapJobType(job.JobTypeEnum),
		Visibility:     mapVisibility(job.Visibility),
		CreatedAt:      time.Unix(job.CreatedAtUnixSeconds, 0).UTC(),
	}
}

func mapJobType(jobType jobv1.JobType) string {
	switch jobType {
	case jobv1.JobType_JOB_TYPE_FIXED:
		return jobTypeFixed
	case jobv1.JobType_JOB_TYPE_HOURLY:
		return jobTypeHourly
	default:
		return ""
	}
}

func mapVisibility(visibility jobv1.Visibility) string {
	switch visibility {
	case jobv1.Visibility_VISIBILITY_PUBLIC:
		return visibilityPublic
	default:
		return strings.ToLower(strings.TrimPrefix(visibility.String(), "VISIBILITY_"))
	}
}
