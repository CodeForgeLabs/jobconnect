package application

import (
	"context"
	"fmt"
	"strings"

	"jobconnect/job/internal/domain"

	"github.com/google/uuid"
)

type HireApplicant struct {
	Jobs      JobRepository
	Proposals ProposalClient
	Contracts ContractClient
	Actors    ActorPolicy
	Clock     Clock
}

type HireApplicantInput struct {
	ProposalID int64
	ClientID   uuid.UUID
}

type HireApplicantOutput struct {
	Hired      bool
	JobID      int64
	ContractID int64
}

func (uc *HireApplicant) Execute(ctx context.Context, in HireApplicantInput) (HireApplicantOutput, error) {
	if in.ProposalID <= 0 {
		return HireApplicantOutput{}, fmt.Errorf("proposal_id is required")
	}
	if in.ClientID == uuid.Nil {
		return HireApplicantOutput{}, fmt.Errorf("client_id is required")
	}
	if uc.Contracts == nil || uc.Actors == nil {
		return HireApplicantOutput{}, fmt.Errorf("hire applicant dependencies are not configured")
	}

	proposal, err := uc.Proposals.GetProposal(ctx, in.ProposalID)
	if err != nil {
		return HireApplicantOutput{}, err
	}
	proposalClientID, parseErr := uuid.Parse(strings.TrimSpace(proposal.ClientID))
	if parseErr != nil || proposalClientID != in.ClientID {
		return HireApplicantOutput{}, fmt.Errorf("proposal not found")
	}
	if err := uc.Actors.EnsureClientCanHire(ctx, in.ClientID); err != nil {
		return HireApplicantOutput{}, err
	}
	freelancerID, parseErr := uuid.Parse(strings.TrimSpace(proposal.FreelancerID))
	if parseErr != nil {
		return HireApplicantOutput{}, fmt.Errorf("invalid freelancer_id on proposal")
	}
	if err := uc.Actors.EnsureFreelancerCanWork(ctx, freelancerID); err != nil {
		return HireApplicantOutput{}, err
	}
	job, err := uc.Jobs.GetByIDForClient(ctx, proposal.JobID, in.ClientID)
	if err != nil {
		return HireApplicantOutput{}, err
	}
	if !strings.EqualFold(job.Status, domain.JobStatusOpen) {
		return HireApplicantOutput{}, fmt.Errorf("job must be open to hire an applicant")
	}
	contractState, err := uc.Contracts.GetJobOfferState(ctx, proposal.JobID, in.ClientID)
	if err != nil {
		return HireApplicantOutput{}, err
	}
	if contractState.HasPendingOffer || contractState.HasActiveContract {
		return HireApplicantOutput{}, fmt.Errorf("job already has an active offer or contract")
	}

	requestID := fmt.Sprintf("job-service-hire-%d", in.ProposalID)
	if err := uc.Proposals.InternalHireProposal(ctx, in.ProposalID, in.ClientID, requestID, ""); err != nil {
		return HireApplicantOutput{}, err
	}
	return HireApplicantOutput{Hired: true, JobID: proposal.JobID, ContractID: 0}, nil
}

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
