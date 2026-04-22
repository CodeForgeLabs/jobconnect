package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

const (
	ApplicantStageSent        = "sent"
	ApplicantStageShortlisted = "shortlisted"
	ApplicantStageRejected    = "rejected"
	ApplicantStageOfferSent   = "offer_sent"
	ApplicantStageHired       = "hired"
)

type InviteFreelancerToJob struct {
	Jobs  JobRepository
	Clock Clock
}

type InviteFreelancerToJobInput struct {
	JobID        int64
	ClientID     uuid.UUID
	FreelancerID string
}

type InviteFreelancerToJobOutput struct {
	Invited bool
}

func (uc *InviteFreelancerToJob) Execute(ctx context.Context, in InviteFreelancerToJobInput) (InviteFreelancerToJobOutput, error) {
	if in.JobID <= 0 {
		return InviteFreelancerToJobOutput{}, fmt.Errorf("job_id is required")
	}
	if in.ClientID == uuid.Nil {
		return InviteFreelancerToJobOutput{}, fmt.Errorf("client_id is required")
	}
	freelancerID := strings.TrimSpace(in.FreelancerID)
	if freelancerID == "" {
		return InviteFreelancerToJobOutput{}, fmt.Errorf("freelancer_id is required")
	}
	invited, err := uc.Jobs.InviteFreelancer(ctx, in.JobID, in.ClientID, freelancerID, uc.Clock.Now())
	if err != nil {
		return InviteFreelancerToJobOutput{}, err
	}
	return InviteFreelancerToJobOutput{Invited: invited}, nil
}

type Applicant struct {
	ProposalID    int64
	FreelancerID  string
	Stage         string
	ConnectsSpent int32
}

type ListJobApplicants struct {
	Jobs      JobRepository
	Proposals ProposalClient
}

type ListJobApplicantsInput struct {
	JobID     int64
	ClientID  uuid.UUID
	PageSize  int32
	PageToken string
}

type ListJobApplicantsOutput struct {
	Applicants    []Applicant
	NextPageToken string
}

func (uc *ListJobApplicants) Execute(ctx context.Context, in ListJobApplicantsInput) (ListJobApplicantsOutput, error) {
	if in.JobID <= 0 {
		return ListJobApplicantsOutput{}, fmt.Errorf("job_id is required")
	}
	if in.ClientID == uuid.Nil {
		return ListJobApplicantsOutput{}, fmt.Errorf("client_id is required")
	}
	if _, err := uc.Jobs.GetByIDForClient(ctx, in.JobID, in.ClientID); err != nil {
		return ListJobApplicantsOutput{}, err
	}

	proposals, err := uc.Proposals.ListProposalsByJob(ctx, in.JobID)
	if err != nil {
		return ListJobApplicantsOutput{}, err
	}

	limit := normalizePageSize(in.PageSize)
	offset, err := parsePageToken(in.PageToken)
	if err != nil {
		return ListJobApplicantsOutput{}, err
	}
	if offset >= len(proposals) {
		return ListJobApplicantsOutput{Applicants: []Applicant{}}, nil
	}

	end := offset + limit
	if end > len(proposals) {
		end = len(proposals)
	}

	outApplicants := make([]Applicant, 0, end-offset)
	for _, p := range proposals[offset:end] {
		outApplicants = append(outApplicants, Applicant{
			ProposalID:    p.ID,
			FreelancerID:  p.FreelancerID,
			Stage:         p.Status,
			ConnectsSpent: p.ConnectsSpent,
		})
	}

	next := ""
	if end < len(proposals) {
		next = strconv.Itoa(end)
	}
	return ListJobApplicantsOutput{Applicants: outApplicants, NextPageToken: next}, nil
}

type SetApplicantStage struct {
	Proposals ProposalClient
}

type SetApplicantStageInput struct {
	ProposalID int64
	ClientID   uuid.UUID
	Stage      string
	Reason     string
}

type SetApplicantStageOutput struct {
	Updated bool
}

func (uc *SetApplicantStage) Execute(ctx context.Context, in SetApplicantStageInput) (SetApplicantStageOutput, error) {
	if in.ProposalID <= 0 {
		return SetApplicantStageOutput{}, fmt.Errorf("proposal_id is required")
	}
	if in.ClientID == uuid.Nil {
		return SetApplicantStageOutput{}, fmt.Errorf("client_id is required")
	}
	stage := strings.ToLower(strings.TrimSpace(in.Stage))
	if !isValidApplicantStage(stage) {
		return SetApplicantStageOutput{}, fmt.Errorf("invalid stage")
	}
	if stage == ApplicantStageSent || stage == ApplicantStageHired {
		return SetApplicantStageOutput{}, fmt.Errorf("invalid stage")
	}

	proposal, err := uc.Proposals.GetProposal(ctx, in.ProposalID)
	if err != nil {
		return SetApplicantStageOutput{}, err
	}
	proposalClientID, parseErr := uuid.Parse(strings.TrimSpace(proposal.ClientID))
	if parseErr != nil || proposalClientID != in.ClientID {
		return SetApplicantStageOutput{}, fmt.Errorf("proposal not found")
	}

	if err := uc.Proposals.SetProposalStatus(ctx, in.ProposalID, stage, strings.TrimSpace(in.Reason)); err != nil {
		return SetApplicantStageOutput{}, err
	}
	return SetApplicantStageOutput{Updated: true}, nil
}

func isValidApplicantStage(stage string) bool {
	switch stage {
	case ApplicantStageSent, ApplicantStageShortlisted, ApplicantStageRejected:
		return true
	default:
		return false
	}
}
