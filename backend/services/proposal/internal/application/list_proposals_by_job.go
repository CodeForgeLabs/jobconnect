package application

import (
	"context"
	"fmt"

	"jobconnect/proposal/internal/domain"

	"github.com/google/uuid"
)

type ListProposalsByJob struct {
	Proposals ProposalRepository
}

type ListProposalsByJobInput struct {
	ClientID     uuid.UUID
	JobID        int64
	StatusFilter []string
	FreelancerID *uuid.UUID
	SortBy       string
	PageSize     int32
	PageToken    string
}

type ListProposalsByJobOutput struct {
	Proposals     []domain.Proposal
	NextPageToken string
}

func (uc *ListProposalsByJob) Execute(ctx context.Context, in ListProposalsByJobInput) (ListProposalsByJobOutput, error) {
	if in.ClientID == uuid.Nil {
		return ListProposalsByJobOutput{}, fmt.Errorf("client_id is required")
	}
	if in.JobID <= 0 {
		return ListProposalsByJobOutput{}, fmt.Errorf("job_id is required")
	}
	if err := domain.ValidateSortBy(in.SortBy); err != nil {
		return ListProposalsByJobOutput{}, err
	}
	for _, s := range in.StatusFilter {
		if err := domain.ValidateStatus(s); err != nil {
			return ListProposalsByJobOutput{}, err
		}
	}

	limit := normalizePageSize(in.PageSize, 20)
	offset, err := parsePageToken(in.PageToken)
	if err != nil {
		return ListProposalsByJobOutput{}, err
	}

	items, err := uc.Proposals.ListByJob(ctx, ListByJobFilter{
		ClientID:     in.ClientID,
		JobID:        in.JobID,
		Statuses:     in.StatusFilter,
		FreelancerID: in.FreelancerID,
		SortBy:       in.SortBy,
	}, limit, offset)
	if err != nil {
		return ListProposalsByJobOutput{}, err
	}

	next := nextPageToken(offset, limit, len(items))
	return ListProposalsByJobOutput{Proposals: items, NextPageToken: next}, nil
}
