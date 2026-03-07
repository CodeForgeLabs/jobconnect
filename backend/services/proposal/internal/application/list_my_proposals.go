package application

import (
	"context"
	"fmt"

	"jobconnect/proposal/internal/domain"

	"github.com/google/uuid"
)

type ListMyProposals struct {
	Proposals ProposalRepository
}

type ListMyProposalsInput struct {
	FreelancerID uuid.UUID
	StatusFilter []string
	JobIDFilter  *int64
	SortBy       string
	PageSize     int32
	PageToken    string
}

type ListMyProposalsOutput struct {
	Proposals     []domain.Proposal
	NextPageToken string
}

func (uc *ListMyProposals) Execute(ctx context.Context, in ListMyProposalsInput) (ListMyProposalsOutput, error) {
	if in.FreelancerID == uuid.Nil {
		return ListMyProposalsOutput{}, fmt.Errorf("freelancer_id is required")
	}
	if err := domain.ValidateSortBy(in.SortBy); err != nil {
		return ListMyProposalsOutput{}, err
	}
	for _, s := range in.StatusFilter {
		if err := domain.ValidateStatus(s); err != nil {
			return ListMyProposalsOutput{}, err
		}
	}

	limit := normalizePageSize(in.PageSize, 20)
	offset, err := parsePageToken(in.PageToken)
	if err != nil {
		return ListMyProposalsOutput{}, err
	}

	items, err := uc.Proposals.ListByFreelancer(ctx, ListByFreelancerFilter{
		FreelancerID: in.FreelancerID,
		Statuses:     in.StatusFilter,
		JobID:        in.JobIDFilter,
		SortBy:       in.SortBy,
	}, limit, offset)
	if err != nil {
		return ListMyProposalsOutput{}, err
	}

	next := nextPageToken(offset, limit, len(items))
	return ListMyProposalsOutput{Proposals: items, NextPageToken: next}, nil
}
