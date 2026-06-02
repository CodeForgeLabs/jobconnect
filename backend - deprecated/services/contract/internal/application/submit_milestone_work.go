package application

import (
	"context"
	"fmt"

	"jobconnect/contract/internal/domain"

	"github.com/google/uuid"
)

type SubmitMilestoneWork struct {
	UpdateMilestoneStatus *UpdateMilestoneStatus
}

type SubmitMilestoneWorkInput struct {
	ContractID  int64
	MilestoneID int64
	ActorID     uuid.UUID
	ActorRole   string
	Note        string
	Attachments []string
}

type SubmitMilestoneWorkOutput struct {
	Contract domain.Contract
}

func (uc *SubmitMilestoneWork) Execute(ctx context.Context, in SubmitMilestoneWorkInput) (SubmitMilestoneWorkOutput, error) {
	if uc.UpdateMilestoneStatus == nil {
		return SubmitMilestoneWorkOutput{}, fmt.Errorf("submit milestone dependencies are not configured")
	}
	out, err := uc.UpdateMilestoneStatus.Execute(ctx, UpdateMilestoneStatusInput{
		ContractID:     in.ContractID,
		MilestoneID:    in.MilestoneID,
		ActorID:        in.ActorID,
		ActorRole:      in.ActorRole,
		Status:         domain.MilestoneStatusSubmitted,
		SubmissionNote: in.Note,
		SubmissionURLs: in.Attachments,
	})
	if err != nil {
		return SubmitMilestoneWorkOutput{}, err
	}
	return SubmitMilestoneWorkOutput{Contract: out.Contract}, nil
}
