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
	ContractID int64
	ActorID    uuid.UUID
	Summary    string
	Payload    domain.AmendmentPayload
	ExpiresAt  *time.Time
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
	if in.ExpiresAt != nil && !in.ExpiresAt.After(now) {
		return ProposeAmendmentOutput{}, fmt.Errorf("expires_at must be in the future")
	}
	// Fetch contract and check state
	contract, err := uc.Contracts.GetByIDForActor(ctx, in.ContractID, in.ActorID)
	if err != nil {
		return ProposeAmendmentOutput{}, fmt.Errorf("contract not found or access denied")
	}
	if contract.Status != domain.StatusActive && contract.Status != domain.StatusPaused {
		return ProposeAmendmentOutput{}, fmt.Errorf("amendments can only be proposed when contract is active or paused")
	}
	payload, err := normalizeAndValidateAmendmentPayload(contract, in.Payload)
	if err != nil {
		return ProposeAmendmentOutput{}, err
	}
	a := domain.Amendment{
		ContractID: in.ContractID,
		ProposedBy: in.ActorID,
		Summary:    strings.TrimSpace(in.Summary),
		Payload:    payload,
		Status:     domain.AmendmentStatusPending,
		ExpiresAt:  in.ExpiresAt,
		CreatedAt:  now,
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
	AmendmentID  int64
	ActorID      uuid.UUID
	Status       string
	ResponseNote string
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
	if status != domain.AmendmentStatusAccepted && status != domain.AmendmentStatusRejected {
		return RespondAmendmentOutput{}, fmt.Errorf("invalid amendment response status: only accept/reject allowed")
	}
	// Fetch amendment and contract
	now := uc.Clock.Now()
	amendment, err := uc.Contracts.GetAmendmentForActor(ctx, in.AmendmentID, in.ActorID)
	if err != nil {
		return RespondAmendmentOutput{}, fmt.Errorf("amendment not found or access denied")
	}
	if amendment.Status == domain.AmendmentStatusPending && amendment.ExpiresAt != nil && !amendment.ExpiresAt.After(now) {
		if _, expireErr := uc.Contracts.ExpireAmendmentForActor(ctx, in.AmendmentID, in.ActorID, now); expireErr != nil {
			return RespondAmendmentOutput{}, expireErr
		}
		return RespondAmendmentOutput{}, fmt.Errorf("amendment has expired")
	}
	contract, err := uc.Contracts.GetByIDForActor(ctx, amendment.ContractID, in.ActorID)
	if err != nil {
		return RespondAmendmentOutput{}, fmt.Errorf("contract not found or access denied")
	}
	// Only counterparty can respond
	if in.ActorID == amendment.ProposedBy {
		return RespondAmendmentOutput{}, fmt.Errorf("only the counterparty can respond to an amendment")
	}
	// Only pending amendments can be responded to
	if amendment.Status != domain.AmendmentStatusPending {
		return RespondAmendmentOutput{}, fmt.Errorf("can only respond to pending amendments")
	}
	// Only allow if contract is active or paused
	if contract.Status != domain.StatusActive && contract.Status != domain.StatusPaused {
		return RespondAmendmentOutput{}, fmt.Errorf("can only respond to amendments when contract is active or paused")
	}
	note := strings.TrimSpace(in.ResponseNote)
	if status == domain.AmendmentStatusRejected && note == "" {
		return RespondAmendmentOutput{}, fmt.Errorf("response_note is required when rejecting an amendment")
	}
	if status == domain.AmendmentStatusAccepted {
		if err := uc.Contracts.RespondAmendmentAndApplyForActor(ctx, in.AmendmentID, in.ActorID, note, now); err != nil {
			return RespondAmendmentOutput{}, err
		}
	} else {
		if err := uc.Contracts.RespondAmendmentForActor(ctx, in.AmendmentID, in.ActorID, status, note, now); err != nil {
			return RespondAmendmentOutput{}, err
		}
	}
	persisted, err := uc.Contracts.GetAmendmentForActor(ctx, in.AmendmentID, in.ActorID)
	if err != nil {
		return RespondAmendmentOutput{}, err
	}
	return RespondAmendmentOutput{Amendment: persisted}, nil
}

type ListAmendments struct {
	Contracts ContractRepository
	Clock     Clock
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
	if uc.Contracts == nil || uc.Clock == nil {
		return ListAmendmentsOutput{}, fmt.Errorf("amendment dependencies are not configured")
	}
	if in.ContractID <= 0 || in.ActorID == uuid.Nil {
		return ListAmendmentsOutput{}, fmt.Errorf("contract_id and actor_id are required")
	}
	if err := uc.Contracts.ExpirePendingAmendmentsForActor(ctx, in.ContractID, in.ActorID, uc.Clock.Now()); err != nil {
		return ListAmendmentsOutput{}, err
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

func normalizeAndValidateAmendmentPayload(contract domain.Contract, payload domain.AmendmentPayload) (domain.AmendmentPayload, error) {
	hasAny := payload.CompensationChange != nil ||
		payload.MilestonesChange != nil ||
		payload.WeeklyLimitChange != nil ||
		payload.ScopeChange != nil
	if !hasAny {
		return domain.AmendmentPayload{}, fmt.Errorf("payload must include at least one change section")
	}

	contractType := strings.ToLower(strings.TrimSpace(contract.ContractType))
	nextFixedTotal := contract.FixedTotal

	if payload.CompensationChange != nil {
		hourly := payload.CompensationChange.NewHourlyRate
		fixed := payload.CompensationChange.NewFixedTotal
		if hourly <= 0 && fixed <= 0 {
			return domain.AmendmentPayload{}, fmt.Errorf("compensation_change requires a positive new value")
		}
		if hourly > 0 && fixed > 0 {
			return domain.AmendmentPayload{}, fmt.Errorf("compensation_change cannot set both hourly and fixed values")
		}
		if contractType == domain.TypeHourly && fixed > 0 {
			return domain.AmendmentPayload{}, fmt.Errorf("fixed total cannot be set for hourly contracts")
		}
		if contractType == domain.TypeFixed && hourly > 0 {
			return domain.AmendmentPayload{}, fmt.Errorf("hourly rate cannot be set for fixed contracts")
		}
		if hourly > 0 {
			if _, err := domain.MoneyToMinorUnits(hourly, "new_hourly_rate"); err != nil {
				return domain.AmendmentPayload{}, err
			}
		}
		if fixed > 0 {
			if _, err := domain.MoneyToMinorUnits(fixed, "new_fixed_total"); err != nil {
				return domain.AmendmentPayload{}, err
			}
		}
		if contractType == domain.TypeFixed && fixed > 0 {
			if payload.MilestonesChange == nil {
				return domain.AmendmentPayload{}, fmt.Errorf("fixed_total change requires milestones_change for fixed contracts")
			}
			nextFixedTotal = fixed
		}
	}

	if payload.MilestonesChange != nil {
		if contractType != domain.TypeFixed {
			return domain.AmendmentPayload{}, fmt.Errorf("milestones_change is only allowed for fixed contracts")
		}
		normalized, err := domain.NormalizeMilestonesForContract(contractType, nextFixedTotal, payload.MilestonesChange.Milestones)
		if err != nil {
			return domain.AmendmentPayload{}, err
		}
		payload.MilestonesChange.Milestones = normalized
	}

	if payload.WeeklyLimitChange != nil {
		if contractType != domain.TypeHourly {
			return domain.AmendmentPayload{}, fmt.Errorf("weekly_limit_change is only allowed for hourly contracts")
		}
		if payload.WeeklyLimitChange.NewWeeklyHourLimit <= 0 {
			return domain.AmendmentPayload{}, fmt.Errorf("new_weekly_hour_limit must be greater than zero")
		}
	}

	if payload.ScopeChange != nil {
		title := strings.TrimSpace(payload.ScopeChange.NewTitle)
		description := strings.TrimSpace(payload.ScopeChange.NewDescription)
		if title == "" && description == "" {
			return domain.AmendmentPayload{}, fmt.Errorf("scope_change requires a new title or description")
		}
		if title != "" && len(title) > 200 {
			return domain.AmendmentPayload{}, fmt.Errorf("new_title too long")
		}
		payload.ScopeChange.NewTitle = title
		payload.ScopeChange.NewDescription = description
	}

	return payload, nil
}
