package application

import (
	"context"
	"fmt"

	"jobconnect/proposal/internal/domain"

	"github.com/google/uuid"
)

type ListClientProposals struct {
	Proposals ProposalRepository
}

type ListClientProposalsInput struct {
	ClientID         uuid.UUID
	StatusFilter     []string
	JobIDFilter      *int64
	FreelancerIDFilter *uuid.UUID
	SortBy           string
	PageSize         int32
	PageToken        string
}

type ListClientProposalsOutput struct {
	Proposals     []domain.Proposal
	NextPageToken string
}

func (uc *ListClientProposals) Execute(ctx context.Context, in ListClientProposalsInput) (ListClientProposalsOutput, error) {
	if in.ClientID == uuid.Nil {
		return ListClientProposalsOutput{}, fmt.Errorf("client_id is required")
	}
	if err := domain.ValidateSortBy(in.SortBy); err != nil {
		return ListClientProposalsOutput{}, err
	}
	for _, s := range in.StatusFilter {
		if err := domain.ValidateStatus(s); err != nil {
			return ListClientProposalsOutput{}, err
		}
	}

	limit := normalizePageSize(in.PageSize, 20)
	offset, err := parsePageToken(in.PageToken)
	if err != nil {
		return ListClientProposalsOutput{}, err
	}

	items, err := uc.Proposals.ListByClient(ctx, ListByClientFilter{
		ClientID:     in.ClientID,
		Statuses:     in.StatusFilter,
		JobID:        in.JobIDFilter,
		FreelancerID: in.FreelancerIDFilter,
		SortBy:       in.SortBy,
	}, limit, offset)
	if err != nil {
		return ListClientProposalsOutput{}, err
	}

	next := nextPageToken(offset, limit, len(items))
	return ListClientProposalsOutput{Proposals: items, NextPageToken: next}, nil
}
