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

type LogHourlyWork struct {
	Contracts ContractRepository
	Clock     Clock
}

type LogHourlyWorkInput struct {
	ContractID   int64
	FreelancerID uuid.UUID
	StartAt      time.Time
	EndAt        time.Time
	Note         string
}

type LogHourlyWorkOutput struct {
	HourlyLog domain.HourlyLog
}

func (uc *LogHourlyWork) Execute(ctx context.Context, in LogHourlyWorkInput) (LogHourlyWorkOutput, error) {
	if uc.Contracts == nil || uc.Clock == nil {
		return LogHourlyWorkOutput{}, fmt.Errorf("hourly log dependencies are not configured")
	}
	if in.ContractID <= 0 || in.FreelancerID == uuid.Nil {
		return LogHourlyWorkOutput{}, fmt.Errorf("contract_id and freelancer_id are required")
	}
	if !in.EndAt.After(in.StartAt) {
		return LogHourlyWorkOutput{}, fmt.Errorf("end_at must be after start_at")
	}
	duration := int32(in.EndAt.Sub(in.StartAt).Minutes())
	if duration <= 0 {
		return LogHourlyWorkOutput{}, fmt.Errorf("duration must be positive")
	}
	now := uc.Clock.Now()
	log := domain.HourlyLog{
		ContractID:   in.ContractID,
		FreelancerID: in.FreelancerID,
		WorkDate:     time.Date(in.StartAt.Year(), in.StartAt.Month(), in.StartAt.Day(), 0, 0, 0, 0, time.UTC),
		StartAt:      in.StartAt.UTC(),
		EndAt:        in.EndAt.UTC(),
		DurationMin:  duration,
		Note:         strings.TrimSpace(in.Note),
		Status:       domain.HourlyLogStatusPending,
		CreatedAt:    now,
	}
	id, err := uc.Contracts.CreateHourlyLogForFreelancer(ctx, log)
	if err != nil {
		return LogHourlyWorkOutput{}, err
	}
	persisted, err := uc.Contracts.GetHourlyLogForActor(ctx, id, in.FreelancerID)
	if err != nil {
		return LogHourlyWorkOutput{}, err
	}
	return LogHourlyWorkOutput{HourlyLog: persisted}, nil
}

type ListHourlyLogs struct {
	Contracts ContractRepository
}

type ListHourlyLogsInput struct {
	ContractID int64
	ActorID    uuid.UUID
	PageSize   int32
	PageToken  string
}

type ListHourlyLogsOutput struct {
	HourlyLogs    []domain.HourlyLog
	NextPageToken string
}

func (uc *ListHourlyLogs) Execute(ctx context.Context, in ListHourlyLogsInput) (ListHourlyLogsOutput, error) {
	if uc.Contracts == nil {
		return ListHourlyLogsOutput{}, fmt.Errorf("hourly log dependencies are not configured")
	}
	if in.ContractID <= 0 || in.ActorID == uuid.Nil {
		return ListHourlyLogsOutput{}, fmt.Errorf("contract_id and actor_id are required")
	}
	pageSize := int(in.PageSize)
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := 0
	if strings.TrimSpace(in.PageToken) != "" {
		v, err := strconv.Atoi(strings.TrimSpace(in.PageToken))
		if err != nil || v < 0 {
			return ListHourlyLogsOutput{}, fmt.Errorf("invalid page_token")
		}
		offset = v
	}
	items, err := uc.Contracts.ListHourlyLogsForActor(ctx, in.ContractID, in.ActorID, pageSize, offset)
	if err != nil {
		return ListHourlyLogsOutput{}, err
	}
	next := ""
	if len(items) == pageSize {
		next = strconv.Itoa(offset + len(items))
	}
	return ListHourlyLogsOutput{HourlyLogs: items, NextPageToken: next}, nil
}

type ReviewHourlyLog struct {
	Contracts ContractRepository
	Clock     Clock
}

type ReviewHourlyLogInput struct {
	HourlyLogID int64
	ClientID    uuid.UUID
	Status      string
	ReviewNote  string
}

type ReviewHourlyLogOutput struct {
	HourlyLog domain.HourlyLog
}

func (uc *ReviewHourlyLog) Execute(ctx context.Context, in ReviewHourlyLogInput) (ReviewHourlyLogOutput, error) {
	if uc.Contracts == nil || uc.Clock == nil {
		return ReviewHourlyLogOutput{}, fmt.Errorf("hourly log dependencies are not configured")
	}
	if in.HourlyLogID <= 0 || in.ClientID == uuid.Nil {
		return ReviewHourlyLogOutput{}, fmt.Errorf("hourly_log_id and client_id are required")
	}
	status := strings.ToLower(strings.TrimSpace(in.Status))
	if status != domain.HourlyLogStatusApproved && status != domain.HourlyLogStatusRejected {
		return ReviewHourlyLogOutput{}, fmt.Errorf("status must be approved or rejected")
	}
	if err := uc.Contracts.ReviewHourlyLogForClient(ctx, in.HourlyLogID, in.ClientID, status, strings.TrimSpace(in.ReviewNote), uc.Clock.Now()); err != nil {
		return ReviewHourlyLogOutput{}, err
	}
	item, err := uc.Contracts.GetHourlyLogForActor(ctx, in.HourlyLogID, in.ClientID)
	if err != nil {
		return ReviewHourlyLogOutput{}, err
	}
	return ReviewHourlyLogOutput{HourlyLog: item}, nil
}
