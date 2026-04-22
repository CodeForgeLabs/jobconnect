package application

import (
	"context"
	"fmt"

	"jobconnect/proposal/internal/domain"

	"github.com/google/uuid"
)

type StatusCount struct {
	Status string
	Count  int64
}

type CountProposalsByJob struct {
	Proposals ProposalRepository
}

type CountProposalsByJobInput struct {
	ClientID uuid.UUID
	JobID    int64
}

type CountProposalsByJobOutput struct {
	Total    int64
	ByStatus []StatusCount
}

func (uc *CountProposalsByJob) Execute(ctx context.Context, in CountProposalsByJobInput) (CountProposalsByJobOutput, error) {
	if in.ClientID == uuid.Nil {
		return CountProposalsByJobOutput{}, fmt.Errorf("client_id is required")
	}
	if in.JobID <= 0 {
		return CountProposalsByJobOutput{}, fmt.Errorf("job_id is required")
	}
	total, byStatus, err := uc.Proposals.CountByJobForClient(ctx, in.ClientID, in.JobID)
	if err != nil {
		return CountProposalsByJobOutput{}, err
	}
	return CountProposalsByJobOutput{Total: total, ByStatus: toStatusCounts(byStatus)}, nil
}

type CountClientProposalInbox struct {
	Proposals ProposalRepository
}

type CountClientProposalInboxInput struct {
	ClientID      uuid.UUID
	StatusFilters []string
}

type CountClientProposalInboxOutput struct {
	Total    int64
	ByStatus []StatusCount
}

func (uc *CountClientProposalInbox) Execute(ctx context.Context, in CountClientProposalInboxInput) (CountClientProposalInboxOutput, error) {
	if in.ClientID == uuid.Nil {
		return CountClientProposalInboxOutput{}, fmt.Errorf("client_id is required")
	}
	for _, s := range in.StatusFilters {
		if err := domain.ValidateStatus(s); err != nil {
			return CountClientProposalInboxOutput{}, err
		}
	}
	total, byStatus, err := uc.Proposals.CountClientInbox(ctx, in.ClientID, in.StatusFilters)
	if err != nil {
		return CountClientProposalInboxOutput{}, err
	}
	return CountClientProposalInboxOutput{Total: total, ByStatus: toStatusCounts(byStatus)}, nil
}

func toStatusCounts(in map[string]int64) []StatusCount {
	if len(in) == 0 {
		return nil
	}
	statuses := []string{domain.StatusSent, domain.StatusShortlisted, domain.StatusRejected, domain.StatusOfferSent, domain.StatusHired, domain.StatusWithdrawn}
	out := make([]StatusCount, 0, len(in))
	for _, s := range statuses {
		if c, ok := in[s]; ok {
			out = append(out, StatusCount{Status: s, Count: c})
		}
	}
	for k, v := range in {
		known := false
		for _, s := range statuses {
			if s == k {
				known = true
				break
			}
		}
		if !known {
			out = append(out, StatusCount{Status: k, Count: v})
		}
	}
	return out
}
