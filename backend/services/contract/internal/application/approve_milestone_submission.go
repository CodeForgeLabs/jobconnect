package application

import (
	"context"
	"fmt"

	"jobconnect/contract/internal/domain"

	"github.com/google/uuid"
)

type ApproveMilestoneSubmission struct {
	UpdateMilestoneStatus *UpdateMilestoneStatus
}

type ApproveMilestoneSubmissionInput struct {
	ContractID  int64
	MilestoneID int64
	ActorID     uuid.UUID
	ActorRole   string
}

type ApproveMilestoneSubmissionOutput struct {
	Contract domain.Contract
}

func (uc *ApproveMilestoneSubmission) Execute(ctx context.Context, in ApproveMilestoneSubmissionInput) (ApproveMilestoneSubmissionOutput, error) {
	if uc.UpdateMilestoneStatus == nil {
		return ApproveMilestoneSubmissionOutput{}, fmt.Errorf("approve milestone dependencies are not configured")
	}
	out, err := uc.UpdateMilestoneStatus.Execute(ctx, UpdateMilestoneStatusInput{
		ContractID:  in.ContractID,
		MilestoneID: in.MilestoneID,
		ActorID:     in.ActorID,
		ActorRole:   in.ActorRole,
		Status:      domain.MilestoneStatusApproved,
	})
	if err != nil {
		return ApproveMilestoneSubmissionOutput{}, err
	}
	return ApproveMilestoneSubmissionOutput{Contract: out.Contract}, nil
}

