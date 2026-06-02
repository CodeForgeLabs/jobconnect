package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"jobconnect/contract/internal/domain"

	"github.com/google/uuid"
)

type PauseContract struct {
	Contracts ContractRepository
	Clock     Clock
}

type ResumeContract struct {
	Contracts ContractRepository
	Clock     Clock
}

type EndContract struct {
	Contracts ContractRepository
	Disputes  DisputeReader
	Clock     Clock
}

type PauseContractInput struct {
	ContractID int64
	ActorID    uuid.UUID
	Reason     string
}

type ResumeContractInput struct {
	ContractID int64
	ActorID    uuid.UUID
	Reason     string
}

type EndContractInput struct {
	ContractID int64
	ActorID    uuid.UUID
	Reason     string
}

type PauseContractOutput struct{ Contract domain.Contract }
type ResumeContractOutput struct{ Contract domain.Contract }
type EndContractOutput struct{ Contract domain.Contract }

func (uc *PauseContract) Execute(ctx context.Context, in PauseContractInput) (PauseContractOutput, error) {
	if uc.Contracts == nil || uc.Clock == nil {
		return PauseContractOutput{}, fmt.Errorf("pause dependencies are not configured")
	}
	if in.ContractID <= 0 || in.ActorID == uuid.Nil {
		return PauseContractOutput{}, fmt.Errorf("contract_id and actor_id are required")
	}
	c, err := uc.Contracts.GetByIDForActor(ctx, in.ContractID, in.ActorID)
	if err != nil {
		return PauseContractOutput{}, err
	}
	if c.Status != domain.StatusActive {
		return PauseContractOutput{}, fmt.Errorf("can only pause from active status")
	}
	now := uc.Clock.Now()
	switch {
	case in.ActorID == c.ClientID:
		if err := uc.Contracts.SetStatusForClient(ctx, in.ContractID, in.ActorID, domain.StatusPaused, now); err != nil {
			return PauseContractOutput{}, err
		}
	case in.ActorID == c.FreelancerID:
		if err := uc.Contracts.SetStatusForFreelancer(ctx, in.ContractID, in.ActorID, domain.StatusPaused, now); err != nil {
			return PauseContractOutput{}, err
		}
	default:
		return PauseContractOutput{}, fmt.Errorf("contract not found")
	}
	_ = uc.Contracts.AppendStatusHistory(ctx, domain.StatusHistoryEntry{ContractID: in.ContractID, Status: domain.StatusPaused, Reason: strings.TrimSpace(in.Reason), ActorID: in.ActorID, CreatedAt: now})
	c, err = uc.Contracts.GetByIDForActor(ctx, in.ContractID, in.ActorID)
	if err != nil {
		return PauseContractOutput{}, err
	}
	return PauseContractOutput{Contract: c}, nil
}

func (uc *ResumeContract) Execute(ctx context.Context, in ResumeContractInput) (ResumeContractOutput, error) {
	if uc.Contracts == nil || uc.Clock == nil {
		return ResumeContractOutput{}, fmt.Errorf("resume dependencies are not configured")
	}
	if in.ContractID <= 0 || in.ActorID == uuid.Nil {
		return ResumeContractOutput{}, fmt.Errorf("contract_id and actor_id are required")
	}
	c, err := uc.Contracts.GetByIDForActor(ctx, in.ContractID, in.ActorID)
	if err != nil {
		return ResumeContractOutput{}, err
	}
	if c.Status != domain.StatusPaused {
		return ResumeContractOutput{}, fmt.Errorf("can only resume from paused status")
	}
	now := uc.Clock.Now()
	switch {
	case in.ActorID == c.ClientID:
		if err := uc.Contracts.SetStatusForClient(ctx, in.ContractID, in.ActorID, domain.StatusActive, now); err != nil {
			return ResumeContractOutput{}, err
		}
	case in.ActorID == c.FreelancerID:
		if err := uc.Contracts.SetStatusForFreelancer(ctx, in.ContractID, in.ActorID, domain.StatusActive, now); err != nil {
			return ResumeContractOutput{}, err
		}
	default:
		return ResumeContractOutput{}, fmt.Errorf("contract not found")
	}
	_ = uc.Contracts.AppendStatusHistory(ctx, domain.StatusHistoryEntry{ContractID: in.ContractID, Status: domain.StatusActive, Reason: strings.TrimSpace(in.Reason), ActorID: in.ActorID, CreatedAt: now})
	c, err = uc.Contracts.GetByIDForActor(ctx, in.ContractID, in.ActorID)
	if err != nil {
		return ResumeContractOutput{}, err
	}
	return ResumeContractOutput{Contract: c}, nil
}

func (uc *EndContract) Execute(ctx context.Context, in EndContractInput) (EndContractOutput, error) {
	if uc.Contracts == nil || uc.Clock == nil {
		return EndContractOutput{}, fmt.Errorf("end dependencies are not configured")
	}
	if in.ContractID <= 0 || in.ActorID == uuid.Nil {
		return EndContractOutput{}, fmt.Errorf("contract_id and actor_id are required")
	}
	c, err := uc.Contracts.GetByIDForActor(ctx, in.ContractID, in.ActorID)
	if err != nil {
		return EndContractOutput{}, err
	}
	if in.ActorID != c.ClientID {
		return EndContractOutput{}, fmt.Errorf("only client can end contract")
	}
	if c.Status != domain.StatusActive && c.Status != domain.StatusPaused {
		return EndContractOutput{}, fmt.Errorf("can only end from active or paused status")
	}
	now := uc.Clock.Now()
	for _, milestone := range c.Milestones {
		if uc.Disputes != nil {
			openDisputeID, err := uc.Disputes.GetOpenDisputeID(ctx, "milestone", fmt.Sprintf("%d:%d", c.ID, milestone.ID))
			if err != nil {
				return EndContractOutput{}, err
			}
			if strings.TrimSpace(openDisputeID) != "" {
				_ = uc.Contracts.AppendStatusHistory(ctx, domain.StatusHistoryEntry{ContractID: in.ContractID, EventType: domain.StatusHistoryEventContractEndBlocked, Reason: "contract has open milestone dispute", ActorID: in.ActorID, CreatedAt: now})
				return EndContractOutput{}, fmt.Errorf("contract has open milestone dispute")
			}
		}
		switch strings.ToLower(strings.TrimSpace(milestone.Status)) {
		case domain.MilestoneStatusFunded, domain.MilestoneStatusSubmitted, domain.MilestoneStatusApprovedPendingSettlement:
			_ = uc.Contracts.AppendStatusHistory(ctx, domain.StatusHistoryEntry{ContractID: in.ContractID, EventType: domain.StatusHistoryEventContractEndBlocked, Reason: "contract has unresolved milestones", ActorID: in.ActorID, CreatedAt: now})
			return EndContractOutput{}, fmt.Errorf("contract has unresolved milestones")
		}
	}
	if blocked, reason, err := uc.Contracts.HasBlockingFinancialActivity(ctx, in.ContractID); err != nil {
		return EndContractOutput{}, err
	} else if blocked {
		_ = uc.Contracts.AppendStatusHistory(ctx, domain.StatusHistoryEntry{ContractID: in.ContractID, EventType: domain.StatusHistoryEventContractEndBlocked, Reason: reason, ActorID: in.ActorID, CreatedAt: now})
		return EndContractOutput{}, fmt.Errorf("%s", reason)
	}
	if err := uc.Contracts.SetStatusForClient(ctx, in.ContractID, in.ActorID, domain.StatusEnded, now); err != nil {
		return EndContractOutput{}, err
	}
	_ = uc.Contracts.AppendStatusHistory(ctx, domain.StatusHistoryEntry{ContractID: in.ContractID, Status: domain.StatusEnded, Reason: strings.TrimSpace(in.Reason), ActorID: in.ActorID, CreatedAt: now})
	c, err = uc.Contracts.GetByIDForActor(ctx, in.ContractID, in.ActorID)
	if err != nil {
		return EndContractOutput{}, err
	}
	return EndContractOutput{Contract: c}, nil
}

type GetStatusHistory struct {
	Contracts ContractRepository
}

type GetStatusHistoryInput struct {
	ContractID int64
	ActorID    uuid.UUID
	PageSize   int32
	PageToken  string
}

type GetStatusHistoryOutput struct {
	Entries       []domain.StatusHistoryEntry
	NextPageToken string
}

func (uc *GetStatusHistory) Execute(ctx context.Context, in GetStatusHistoryInput) (GetStatusHistoryOutput, error) {
	if uc.Contracts == nil {
		return GetStatusHistoryOutput{}, fmt.Errorf("status history dependencies are not configured")
	}
	if in.ContractID <= 0 || in.ActorID == uuid.Nil {
		return GetStatusHistoryOutput{}, fmt.Errorf("contract_id and actor_id are required")
	}
	pageSize := int(in.PageSize)
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := 0
	if strings.TrimSpace(in.PageToken) != "" {
		v, err := strconv.Atoi(strings.TrimSpace(in.PageToken))
		if err != nil || v < 0 {
			return GetStatusHistoryOutput{}, fmt.Errorf("invalid page_token")
		}
		offset = v
	}
	items, err := uc.Contracts.ListStatusHistoryForActor(ctx, in.ContractID, in.ActorID, pageSize, offset)
	if err != nil {
		return GetStatusHistoryOutput{}, err
	}
	next := ""
	if len(items) == pageSize {
		next = strconv.Itoa(offset + len(items))
	}
	return GetStatusHistoryOutput{Entries: items, NextPageToken: next}, nil
}
