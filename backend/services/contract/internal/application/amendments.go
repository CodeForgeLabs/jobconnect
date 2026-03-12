package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"jobconnect/contract/internal/domain"

	"github.com/google/uuid"
)

type ProposeAmendment struct {
	Contracts ContractRepository
	Clock     Clock
}

type ProposeAmendmentInput struct {
	ContractID  int64
	ActorID     uuid.UUID
	Summary     string
	PayloadJSON string
	ExpiresAt   *time.Time
}

type ProposeAmendmentOutput struct {
	Amendment domain.Amendment
}

func (uc *ProposeAmendment) Execute(ctx context.Context, in ProposeAmendmentInput) (ProposeAmendmentOutput, error) {
	if uc.Contracts == nil || uc.Clock == nil {
		return ProposeAmendmentOutput{}, fmt.Errorf("amendment dependencies are not configured")
	}
	if in.ContractID <= 0 || in.ActorID == uuid.Nil {
		return ProposeAmendmentOutput{}, fmt.Errorf("contract_id and actor_id are required")
	}
	if strings.TrimSpace(in.Summary) == "" {
		return ProposeAmendmentOutput{}, fmt.Errorf("summary is required")
	}
	now := uc.Clock.Now()
	a := domain.Amendment{
		ContractID:  in.ContractID,
		ProposedBy:  in.ActorID,
		Summary:     strings.TrimSpace(in.Summary),
		PayloadJSON: strings.TrimSpace(in.PayloadJSON),
		Status:      domain.AmendmentStatusPending,
		ExpiresAt:   in.ExpiresAt,
		CreatedAt:   now,
	}
	id, err := uc.Contracts.CreateAmendmentForActor(ctx, a)
	if err != nil {
		return ProposeAmendmentOutput{}, err
	}
	persisted, err := uc.Contracts.GetAmendmentForActor(ctx, id, in.ActorID)
	if err != nil {
		return ProposeAmendmentOutput{}, err
	}
	return ProposeAmendmentOutput{Amendment: persisted}, nil
}

type RespondAmendment struct {
	Contracts ContractRepository
	Clock     Clock
}

type RespondAmendmentInput struct {
	AmendmentID int64
	ActorID     uuid.UUID
	Status      string
}

type RespondAmendmentOutput struct {
	Amendment domain.Amendment
}

func (uc *RespondAmendment) Execute(ctx context.Context, in RespondAmendmentInput) (RespondAmendmentOutput, error) {
	if uc.Contracts == nil || uc.Clock == nil {
		return RespondAmendmentOutput{}, fmt.Errorf("amendment dependencies are not configured")
	}
	if in.AmendmentID <= 0 || in.ActorID == uuid.Nil {
		return RespondAmendmentOutput{}, fmt.Errorf("amendment_id and actor_id are required")
	}
	status := strings.ToLower(strings.TrimSpace(in.Status))
	if status != domain.AmendmentStatusAccepted && status != domain.AmendmentStatusRejected && status != domain.AmendmentStatusExpired {
		return RespondAmendmentOutput{}, fmt.Errorf("invalid amendment response status")
	}
	if err := uc.Contracts.RespondAmendmentForActor(ctx, in.AmendmentID, in.ActorID, status, uc.Clock.Now()); err != nil {
		return RespondAmendmentOutput{}, err
	}
	persisted, err := uc.Contracts.GetAmendmentForActor(ctx, in.AmendmentID, in.ActorID)
	if err != nil {
		return RespondAmendmentOutput{}, err
	}
	return RespondAmendmentOutput{Amendment: persisted}, nil
}

type ListAmendments struct {
	Contracts ContractRepository
}

type ListAmendmentsInput struct {
	ContractID int64
	ActorID    uuid.UUID
	PageSize   int32
	PageToken  string
}

type ListAmendmentsOutput struct {
	Amendments    []domain.Amendment
	NextPageToken string
}

func (uc *ListAmendments) Execute(ctx context.Context, in ListAmendmentsInput) (ListAmendmentsOutput, error) {
	if uc.Contracts == nil {
		return ListAmendmentsOutput{}, fmt.Errorf("amendment dependencies are not configured")
	}
	if in.ContractID <= 0 || in.ActorID == uuid.Nil {
		return ListAmendmentsOutput{}, fmt.Errorf("contract_id and actor_id are required")
	}
	pageSize := int(in.PageSize)
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := 0
	if strings.TrimSpace(in.PageToken) != "" {
		v, err := strconv.Atoi(strings.TrimSpace(in.PageToken))
		if err != nil || v < 0 {
			return ListAmendmentsOutput{}, fmt.Errorf("invalid page_token")
		}
		offset = v
	}
	items, err := uc.Contracts.ListAmendmentsForActor(ctx, in.ContractID, in.ActorID, pageSize, offset)
	if err != nil {
		return ListAmendmentsOutput{}, err
	}
	next := ""
	if len(items) == pageSize {
		next = strconv.Itoa(offset + len(items))
	}
	return ListAmendmentsOutput{Amendments: items, NextPageToken: next}, nil
}
