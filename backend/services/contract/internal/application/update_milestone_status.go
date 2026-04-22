package application

import (
	"context"
	"fmt"
	"strings"

	"jobconnect/contract/internal/domain"

	"github.com/google/uuid"
)

type UpdateMilestoneStatus struct {
	Contracts ContractRepository
	Clock     Clock
}

type UpdateMilestoneStatusInput struct {
	ContractID     int64
	MilestoneID    int64
	ActorID        uuid.UUID
	ActorRole      string
	Status         string
	SubmissionNote string
	SubmissionURLs []string
	ReviewNote     string
}

type UpdateMilestoneStatusOutput struct {
	Contract domain.Contract
}

func (uc *UpdateMilestoneStatus) Execute(ctx context.Context, in UpdateMilestoneStatusInput) (UpdateMilestoneStatusOutput, error) {
	if uc.Contracts == nil || uc.Clock == nil {
		return UpdateMilestoneStatusOutput{}, fmt.Errorf("milestone dependencies are not configured")
	}
	if in.ContractID <= 0 || in.MilestoneID <= 0 {
		return UpdateMilestoneStatusOutput{}, fmt.Errorf("contract_id and milestone_id are required")
	}
	if in.ActorID == uuid.Nil {
		return UpdateMilestoneStatusOutput{}, fmt.Errorf("actor_id is required")
	}
	status := strings.ToLower(strings.TrimSpace(in.Status))
	if err := domain.ValidateMilestoneStatus(status); err != nil {
		return UpdateMilestoneStatusOutput{}, err
	}
	role := strings.ToLower(strings.TrimSpace(in.ActorRole))
	isInternal := role == "service" || role == "system" || role == "internal"
	if role != "client" && role != "freelancer" && !isInternal {
		return UpdateMilestoneStatusOutput{}, fmt.Errorf("client, freelancer, or internal role required")
	}
	if role == "freelancer" && status != domain.MilestoneStatusSubmitted {
		return UpdateMilestoneStatusOutput{}, fmt.Errorf("freelancer can only submit milestones")
	}
	if role == "client" && status != domain.MilestoneStatusApproved && status != domain.MilestoneStatusChangesRequested {
		return UpdateMilestoneStatusOutput{}, fmt.Errorf("client can only approve or request changes")
	}
	if isInternal && status != domain.MilestoneStatusFunded && status != domain.MilestoneStatusApprovedPendingSettlement {
		return UpdateMilestoneStatusOutput{}, fmt.Errorf("internal role can only set funded or approved_pending_settlement")
	}

	current, err := uc.Contracts.GetByIDForActor(ctx, in.ContractID, in.ActorID)
	if err != nil {
		return UpdateMilestoneStatusOutput{}, err
	}
	now := uc.Clock.Now()
	found := false
	for i := range current.Milestones {
		if current.Milestones[i].ID == in.MilestoneID {
			curr := strings.ToLower(strings.TrimSpace(current.Milestones[i].Status))
			if role == "freelancer" {
				if curr != domain.MilestoneStatusPending && curr != domain.MilestoneStatusChangesRequested {
					return UpdateMilestoneStatusOutput{}, fmt.Errorf("milestone can only be submitted from pending or changes_requested")
				}
				note := strings.TrimSpace(in.SubmissionNote)
				urls := make([]string, 0, len(in.SubmissionURLs))
				for _, u := range in.SubmissionURLs {
					u = strings.TrimSpace(u)
					if u != "" {
						urls = append(urls, u)
					}
				}
				if note == "" && len(urls) == 0 {
					return UpdateMilestoneStatusOutput{}, fmt.Errorf("submission_note or submission_urls is required when submitting milestone")
				}
				current.Milestones[i].SubmissionNote = note
				current.Milestones[i].SubmissionURLs = urls
				submittedAt := now
				current.Milestones[i].SubmittedAt = &submittedAt
				current.Milestones[i].ReviewNote = ""
				current.Milestones[i].ReviewedAt = nil
			} else if role == "client" {
				if curr != domain.MilestoneStatusSubmitted {
					return UpdateMilestoneStatusOutput{}, fmt.Errorf("client can only review submitted milestones")
				}
				reviewNote := strings.TrimSpace(in.ReviewNote)
				if status == domain.MilestoneStatusChangesRequested && reviewNote == "" {
					return UpdateMilestoneStatusOutput{}, fmt.Errorf("review_note is required when requesting changes")
				}
				current.Milestones[i].ReviewNote = reviewNote
				reviewedAt := now
				current.Milestones[i].ReviewedAt = &reviewedAt
				if status == domain.MilestoneStatusChangesRequested {
					current.Milestones[i].RevisionCount++
				}
			} else {
				switch status {
				case domain.MilestoneStatusApprovedPendingSettlement:
					if curr != domain.MilestoneStatusApproved {
						return UpdateMilestoneStatusOutput{}, fmt.Errorf("approved_pending_settlement can only be set from approved")
					}
				case domain.MilestoneStatusFunded:
					if curr != domain.MilestoneStatusApproved && curr != domain.MilestoneStatusApprovedPendingSettlement {
						return UpdateMilestoneStatusOutput{}, fmt.Errorf("funded can only be set from approved or approved_pending_settlement")
					}
				}
			}
			current.Milestones[i].Status = status
			found = true
			break
		}
	}
	if !found {
		return UpdateMilestoneStatusOutput{}, fmt.Errorf("milestone not found")
	}
	if err := uc.Contracts.ReplaceMilestonesForActor(ctx, in.ContractID, in.ActorID, current.Milestones, now); err != nil {
		return UpdateMilestoneStatusOutput{}, err
	}
	reason := ""
	if reason == "" {
		if role == "freelancer" {
			reason = strings.TrimSpace(in.SubmissionNote)
		} else {
			reason = strings.TrimSpace(in.ReviewNote)
		}
	}
	_ = uc.Contracts.AppendStatusHistory(ctx, domain.StatusHistoryEntry{
		ContractID: in.ContractID,
		Status:     "milestone_" + status,
		Reason:     reason,
		ActorID:    in.ActorID,
		CreatedAt:  now,
	})
	updated, err := uc.Contracts.GetByIDForActor(ctx, in.ContractID, in.ActorID)
	if err != nil {
		return UpdateMilestoneStatusOutput{}, err
	}
	return UpdateMilestoneStatusOutput{Contract: updated}, nil
}
