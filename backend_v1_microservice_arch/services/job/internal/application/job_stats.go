package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type GetJobStats struct {
	Jobs      JobRepository
	Proposals ProposalClient
}

type GetJobStatsInput struct {
	JobID    int64
	ClientID uuid.UUID
}

type GetJobStatsOutput struct {
	InviteCount         int32
	InviteAcceptedCount int32
	InviteDeclinedCount int32
	ApplicantCount      int32
	ShortlistedCount    int32
	RejectedCount       int32
	HiredCount          int32
}

func (uc *GetJobStats) Execute(ctx context.Context, in GetJobStatsInput) (GetJobStatsOutput, error) {
	if in.JobID <= 0 {
		return GetJobStatsOutput{}, fmt.Errorf("job_id is required")
	}
	if in.ClientID == uuid.Nil {
		return GetJobStatsOutput{}, fmt.Errorf("client_id is required")
	}
	if _, err := uc.Jobs.GetByIDForClient(ctx, in.JobID, in.ClientID); err != nil {
		return GetJobStatsOutput{}, err
	}

	inviteStats, err := uc.Jobs.GetInviteStats(ctx, in.JobID)
	if err != nil {
		return GetJobStatsOutput{}, err
	}
	proposals, err := uc.Proposals.ListProposalsByJob(ctx, in.JobID)
	if err != nil {
		return GetJobStatsOutput{}, err
	}

	out := GetJobStatsOutput{
		InviteCount:         inviteStats.Total,
		InviteAcceptedCount: inviteStats.Accepted,
		InviteDeclinedCount: inviteStats.Declined,
		ApplicantCount:      int32(len(proposals)),
	}
	for _, p := range proposals {
		switch p.Status {
		case ApplicantStageShortlisted:
			out.ShortlistedCount++
		case ApplicantStageRejected:
			out.RejectedCount++
		case ApplicantStageOfferSent:
			out.ShortlistedCount++
		case ApplicantStageHired:
			out.HiredCount++
		}
	}
	return out, nil
}
