package application

import (
	"context"
	"fmt"
	"strings"

	"jobconnect/job/internal/domain"

	"github.com/google/uuid"
)

type RejectAllApplicants struct {
	Jobs      JobRepository
	Proposals ProposalClient
}

type RejectAllApplicantsInput struct {
	JobID    int64
	ClientID uuid.UUID
	Reason   string
}

type RejectAllApplicantsOutput struct {
	RejectedCount int32
}

func (uc *RejectAllApplicants) Execute(ctx context.Context, in RejectAllApplicantsInput) (RejectAllApplicantsOutput, error) {
	if in.JobID <= 0 {
		return RejectAllApplicantsOutput{}, fmt.Errorf("job_id is required")
	}
	if in.ClientID == uuid.Nil {
		return RejectAllApplicantsOutput{}, fmt.Errorf("client_id is required")
	}
	if _, err := uc.Jobs.GetByIDForClient(ctx, in.JobID, in.ClientID); err != nil {
		return RejectAllApplicantsOutput{}, err
	}

	proposals, err := uc.Proposals.ListProposalsByJob(ctx, in.JobID)
	if err != nil {
		return RejectAllApplicantsOutput{}, err
	}

	var rejected int32
	for _, p := range proposals {
		if p.Status == ApplicantStageHired || p.Status == ApplicantStageRejected {
			continue
		}
		if err := uc.Proposals.SetProposalStatus(ctx, p.ID, ApplicantStageRejected, strings.TrimSpace(in.Reason)); err == nil {
			rejected++
		}
	}
	return RejectAllApplicantsOutput{RejectedCount: rejected}, nil
}

type ReopenHiringForJob struct {
	Jobs      JobRepository
	Proposals ProposalClient
	Contracts ContractClient
	Clock     Clock
}

type ReopenHiringForJobInput struct {
	JobID    int64
	ClientID uuid.UUID
}

type ReopenHiringForJobOutput struct {
	Job domain.Job
}

func (uc *ReopenHiringForJob) Execute(ctx context.Context, in ReopenHiringForJobInput) (ReopenHiringForJobOutput, error) {
	if in.JobID <= 0 {
		return ReopenHiringForJobOutput{}, fmt.Errorf("job_id is required")
	}
	if in.ClientID == uuid.Nil {
		return ReopenHiringForJobOutput{}, fmt.Errorf("client_id is required")
	}
	job, err := uc.Jobs.GetByIDForClient(ctx, in.JobID, in.ClientID)
	if err != nil {
		return ReopenHiringForJobOutput{}, err
	}
	contractState, err := uc.Contracts.GetJobOfferState(ctx, in.JobID, in.ClientID)
	if err != nil {
		return ReopenHiringForJobOutput{}, err
	}
	if contractState.HasPendingOffer || contractState.HasActiveContract {
		return ReopenHiringForJobOutput{}, fmt.Errorf("revoke active offers before reopening hiring")
	}
	if strings.EqualFold(job.Status, domain.JobStatusOpen) {
		proposals, listErr := uc.Proposals.ListProposalsByJob(ctx, in.JobID)
		if listErr != nil {
			return ReopenHiringForJobOutput{}, listErr
		}
		for _, proposal := range proposals {
			if strings.EqualFold(proposal.Status, ApplicantStageHired) {
				if err := uc.Proposals.ReleaseHiredProposal(ctx, proposal.ID, in.ClientID, "hiring reopened"); err != nil {
					return ReopenHiringForJobOutput{}, err
				}
			}
		}
		return ReopenHiringForJobOutput{Job: job}, nil
	}
	job, err = uc.Jobs.ReopenHiring(ctx, in.JobID, in.ClientID, uc.Clock.Now())
	if err != nil {
		return ReopenHiringForJobOutput{}, err
	}
	return ReopenHiringForJobOutput{Job: job}, nil
}
