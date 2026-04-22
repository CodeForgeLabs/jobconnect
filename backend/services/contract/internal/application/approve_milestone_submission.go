package application

import (
	"context"
	"fmt"
	"math"
	"strings"

	"jobconnect/contract/internal/domain"

	"github.com/google/uuid"
)

type ApproveMilestoneSubmission struct {
	UpdateMilestoneStatus *UpdateMilestoneStatus
	Settlement            MilestoneSettlementDispatcher
	Disputes              DisputeReader
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
	referenceID := fmt.Sprintf("%d:%d", in.ContractID, in.MilestoneID)
	if uc.Disputes != nil {
		hasOpen, err := uc.Disputes.HasOpenDispute(ctx, "milestone", referenceID)
		if err != nil {
			return ApproveMilestoneSubmissionOutput{}, err
		}
		if hasOpen {
			return ApproveMilestoneSubmissionOutput{}, fmt.Errorf("open dispute exists for milestone")
		}
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
	if uc.Settlement == nil {
		return ApproveMilestoneSubmissionOutput{Contract: out.Contract}, nil
	}
	var amountMinor int64
	for _, m := range out.Contract.Milestones {
		if m.ID == in.MilestoneID {
			amountMinor = int64(math.Round(m.Amount * 100))
			break
		}
	}
	if amountMinor <= 0 {
		return ApproveMilestoneSubmissionOutput{}, fmt.Errorf("milestone amount must be greater than zero")
	}

	command := MilestoneApprovedSettlementCommand{
		EventID:      fmt.Sprintf("milestone-approved:%d:%d", in.ContractID, in.MilestoneID),
		ContractID:   in.ContractID,
		MilestoneID:  in.MilestoneID,
		ReferenceID:  referenceID,
		AmountMinor:  amountMinor,
		FreelancerID: strings.TrimSpace(out.Contract.FreelancerID.String()),
	}
	if err := uc.Settlement.DispatchMilestoneApproved(ctx, command); err != nil {
		pendingOut, pendingErr := uc.UpdateMilestoneStatus.Execute(ctx, UpdateMilestoneStatusInput{
			ContractID:  in.ContractID,
			MilestoneID: in.MilestoneID,
			ActorID:     in.ActorID,
			ActorRole:   "internal",
			Status:      domain.MilestoneStatusApprovedPendingSettlement,
		})
		if pendingErr == nil {
			return ApproveMilestoneSubmissionOutput{Contract: pendingOut.Contract}, nil
		}
		return ApproveMilestoneSubmissionOutput{}, fmt.Errorf("dispatch milestone settlement: %w", err)
	}
	return ApproveMilestoneSubmissionOutput{Contract: out.Contract}, nil
}
