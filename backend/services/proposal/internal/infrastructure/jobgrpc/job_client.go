package jobgrpc

import (
	"context"
	"fmt"

	jobv1 "jobconnect/job/gen/job/v1"
	"jobconnect/proposal/internal/application"

	"github.com/google/uuid"
)

type JobClient struct {
	client jobv1.JobServiceClient
}

func NewJobClient(client jobv1.JobServiceClient) *JobClient {
	return &JobClient{client: client}
}

func (c *JobClient) GetJobSummary(ctx context.Context, jobID int64) (application.JobSummary, error) {
	if c == nil || c.client == nil {
		return application.JobSummary{}, fmt.Errorf("job client is nil")
	}
	if jobID <= 0 {
		return application.JobSummary{}, fmt.Errorf("job_id is required")
	}

	pageToken := ""
	for i := 0; i < 200; i++ {
		resp, err := c.client.ListOpenJobs(ctx, &jobv1.ListOpenJobsRequest{
			PageSize:  100,
			PageToken: pageToken,
		})
		if err != nil {
			return application.JobSummary{}, fmt.Errorf("list open jobs: %w", err)
		}

		for _, j := range resp.GetJobs() {
			if j == nil || j.GetId() != jobID {
				continue
			}
			clientID, err := uuid.Parse(j.GetClientId())
			if err != nil {
				return application.JobSummary{}, fmt.Errorf("invalid client_id from job service")
			}
			statusEnum := j.GetStatusEnum()
			return application.JobSummary{
				JobID:    j.GetId(),
				ClientID: clientID,
				Status:   statusEnum.String(),
				IsOpen:   statusEnum == jobv1.JobStatus_JOB_STATUS_OPEN,
				Found:    true,
			}, nil
		}

		next := resp.GetNextPageToken()
		if next == "" || next == pageToken {
			break
		}
		pageToken = next
	}

	return application.JobSummary{JobID: jobID, Found: false}, nil
}
