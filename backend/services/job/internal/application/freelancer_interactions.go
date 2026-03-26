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
	InviteResponseAccepted = "accepted"
	InviteResponseDeclined = "declined"
)

type GetPublicJobDetail struct {
	Jobs JobRepository
}

type GetPublicJobDetailInput struct {
	JobID int64
}

type GetPublicJobDetailOutput struct {
	Job domain.Job
}

func (uc *GetPublicJobDetail) Execute(ctx context.Context, in GetPublicJobDetailInput) (GetPublicJobDetailOutput, error) {
	if in.JobID <= 0 {
		return GetPublicJobDetailOutput{}, fmt.Errorf("job_id is required")
	}
	job, err := uc.Jobs.GetPublicByID(ctx, in.JobID)
	if err != nil {
		return GetPublicJobDetailOutput{}, err
	}
	return GetPublicJobDetailOutput{Job: job}, nil
}

type ListInvitedJobs struct {
	Jobs JobRepository
}

type ListInvitedJobsInput struct {
	FreelancerID uuid.UUID
	PageSize     int32
	PageToken    string
}

type ListInvitedJobsOutput struct {
	Jobs          []domain.Job
	NextPageToken string
}

func (uc *ListInvitedJobs) Execute(ctx context.Context, in ListInvitedJobsInput) (ListInvitedJobsOutput, error) {
	if in.FreelancerID == uuid.Nil {
		return ListInvitedJobsOutput{}, fmt.Errorf("freelancer_id is required")
	}
	limit := normalizePageSize(in.PageSize)
	offset, err := parsePageToken(in.PageToken)
	if err != nil {
		return ListInvitedJobsOutput{}, err
	}
	jobs, err := uc.Jobs.ListInvitedJobs(ctx, in.FreelancerID, limit, offset)
	if err != nil {
		return ListInvitedJobsOutput{}, err
	}
	next := ""
	if len(jobs) == limit {
		next = strconv.Itoa(offset + len(jobs))
	}
	return ListInvitedJobsOutput{Jobs: jobs, NextPageToken: next}, nil
}

type RespondToJobInvite struct {
	Jobs  JobRepository
	Clock Clock
}

type RespondToJobInviteInput struct {
	JobID         int64
	FreelancerID  uuid.UUID
	ResponseState string
}

type RespondToJobInviteOutput struct {
	Updated bool
}

func (uc *RespondToJobInvite) Execute(ctx context.Context, in RespondToJobInviteInput) (RespondToJobInviteOutput, error) {
	if in.JobID <= 0 {
		return RespondToJobInviteOutput{}, fmt.Errorf("job_id is required")
	}
	if in.FreelancerID == uuid.Nil {
		return RespondToJobInviteOutput{}, fmt.Errorf("freelancer_id is required")
	}
	state := strings.ToLower(strings.TrimSpace(in.ResponseState))
	if state != InviteResponseAccepted && state != InviteResponseDeclined {
		return RespondToJobInviteOutput{}, fmt.Errorf("invalid response_status")
	}
	updated, err := uc.Jobs.RespondToInvite(ctx, in.JobID, in.FreelancerID, state, uc.Clock.Now())
	if err != nil {
		return RespondToJobInviteOutput{}, err
	}
	return RespondToJobInviteOutput{Updated: updated}, nil
}

type SaveJob struct {
	Jobs  JobRepository
	Clock Clock
}

type SaveJobInput struct {
	JobID        int64
	FreelancerID uuid.UUID
}

type SaveJobOutput struct {
	Saved bool
}

func (uc *SaveJob) Execute(ctx context.Context, in SaveJobInput) (SaveJobOutput, error) {
	if in.JobID <= 0 {
		return SaveJobOutput{}, fmt.Errorf("job_id is required")
	}
	if in.FreelancerID == uuid.Nil {
		return SaveJobOutput{}, fmt.Errorf("freelancer_id is required")
	}
	saved, err := uc.Jobs.SaveJob(ctx, in.JobID, in.FreelancerID, uc.Clock.Now())
	if err != nil {
		return SaveJobOutput{}, err
	}
	return SaveJobOutput{Saved: saved}, nil
}

type UnsaveJob struct {
	Jobs JobRepository
}

type UnsaveJobInput struct {
	JobID        int64
	FreelancerID uuid.UUID
}

type UnsaveJobOutput struct {
	Removed bool
}

func (uc *UnsaveJob) Execute(ctx context.Context, in UnsaveJobInput) (UnsaveJobOutput, error) {
	if in.JobID <= 0 {
		return UnsaveJobOutput{}, fmt.Errorf("job_id is required")
	}
	if in.FreelancerID == uuid.Nil {
		return UnsaveJobOutput{}, fmt.Errorf("freelancer_id is required")
	}
	removed, err := uc.Jobs.UnsaveJob(ctx, in.JobID, in.FreelancerID)
	if err != nil {
		return UnsaveJobOutput{}, err
	}
	return UnsaveJobOutput{Removed: removed}, nil
}

type ListSavedJobs struct {
	Jobs JobRepository
}

type ListSavedJobsInput struct {
	FreelancerID uuid.UUID
	PageSize     int32
	PageToken    string
}

type ListSavedJobsOutput struct {
	Jobs          []domain.Job
	NextPageToken string
}

func (uc *ListSavedJobs) Execute(ctx context.Context, in ListSavedJobsInput) (ListSavedJobsOutput, error) {
	if in.FreelancerID == uuid.Nil {
		return ListSavedJobsOutput{}, fmt.Errorf("freelancer_id is required")
	}
	limit := normalizePageSize(in.PageSize)
	offset, err := parsePageToken(in.PageToken)
	if err != nil {
		return ListSavedJobsOutput{}, err
	}
	jobs, err := uc.Jobs.ListSavedJobs(ctx, in.FreelancerID, limit, offset)
	if err != nil {
		return ListSavedJobsOutput{}, err
	}
	next := ""
	if len(jobs) == limit {
		next = strconv.Itoa(offset + len(jobs))
	}
	return ListSavedJobsOutput{Jobs: jobs, NextPageToken: next}, nil
}
