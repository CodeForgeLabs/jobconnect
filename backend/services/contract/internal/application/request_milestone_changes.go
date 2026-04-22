package application

import (
	"context"
	"fmt"

	"jobconnect/contract/internal/domain"

	"github.com/google/uuid"
)

type RequestMilestoneChanges struct {
	UpdateMilestoneStatus *UpdateMilestoneStatus
}

type RequestMilestoneChangesInput struct {
	ContractID  int64
	MilestoneID int64
	ActorID     uuid.UUID
	ActorRole   string
	Note        string
}

type RequestMilestoneChangesOutput struct {
	Contract domain.Contract
}

func (uc *RequestMilestoneChanges) Execute(ctx context.Context, in RequestMilestoneChangesInput) (RequestMilestoneChangesOutput, error) {
	if uc.UpdateMilestoneStatus == nil {
		return RequestMilestoneChangesOutput{}, fmt.Errorf("request changes dependencies are not configured")
	}
	out, err := uc.UpdateMilestoneStatus.Execute(ctx, UpdateMilestoneStatusInput{
		ContractID:  in.ContractID,
		MilestoneID: in.MilestoneID,
		ActorID:     in.ActorID,
		ActorRole:   in.ActorRole,
		Status:      domain.MilestoneStatusChangesRequested,
		ReviewNote:  in.Note,
	})
	if err != nil {
		return RequestMilestoneChangesOutput{}, err
	}
	return RequestMilestoneChangesOutput{Contract: out.Contract}, nil
}
