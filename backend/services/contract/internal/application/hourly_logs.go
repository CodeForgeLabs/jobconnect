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
	EvidenceURLs []string
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
	startAt := in.StartAt.UTC()
	endAt := in.EndAt.UTC()
	now := uc.Clock.Now().UTC()
	if startAt.After(now) || endAt.After(now) {
		return LogHourlyWorkOutput{}, fmt.Errorf("hourly work cannot be logged in the future")
	}
	currentWeekStart, currentWeekEnd := weekBounds(now)
	if startAt.Before(currentWeekStart) || !endAt.Before(currentWeekEnd) {
		return LogHourlyWorkOutput{}, fmt.Errorf("only current-week hourly logs can be created")
	}
	contract, err := uc.Contracts.GetByIDForActor(ctx, in.ContractID, in.FreelancerID)
	if err != nil {
		return LogHourlyWorkOutput{}, err
	}
	if contract.ContractType != domain.TypeHourly {
		return LogHourlyWorkOutput{}, fmt.Errorf("hourly work can only be logged for hourly contracts")
	}
	if contract.Status != domain.StatusActive {
		return LogHourlyWorkOutput{}, fmt.Errorf("hourly work can only be logged for active contracts")
	}
	duration := int32(endAt.Sub(startAt).Minutes())
	if duration <= 0 {
		return LogHourlyWorkOutput{}, fmt.Errorf("duration must be positive")
	}
	weekStart, weekEnd := weekBounds(startAt)
	if endAt.After(weekEnd) {
		return LogHourlyWorkOutput{}, fmt.Errorf("hourly log cannot cross weekly boundary")
	}
	weeklyLogs, err := uc.Contracts.ListHourlyLogsForActorInRange(ctx, in.ContractID, in.FreelancerID, weekStart, weekEnd)
	if err != nil {
		return LogHourlyWorkOutput{}, err
	}
	var usedMinutes int32
	for _, existing := range weeklyLogs {
		if existing.Status == domain.HourlyLogStatusRejected {
			continue
		}
		if timeRangesOverlap(startAt, endAt, existing.StartAt.UTC(), existing.EndAt.UTC()) {
			return LogHourlyWorkOutput{}, fmt.Errorf("hourly log overlaps existing work log")
		}
		usedMinutes += existing.DurationMin
	}
	if contract.WeeklyHourLimit > 0 {
		limitMinutes := contract.WeeklyHourLimit * 60
		if usedMinutes+duration > limitMinutes {
			return LogHourlyWorkOutput{}, fmt.Errorf("weekly hour limit exceeded")
		}
	}
	log := domain.HourlyLog{
		ContractID:   in.ContractID,
		FreelancerID: in.FreelancerID,
		WorkDate:     time.Date(startAt.Year(), startAt.Month(), startAt.Day(), 0, 0, 0, 0, time.UTC),
		StartAt:      startAt,
		EndAt:        endAt,
		DurationMin:  duration,
		Note:         strings.TrimSpace(in.Note),
		EvidenceURLs: normalizeEvidenceURLs(in.EvidenceURLs),
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

type GetHourlyWorkSummary struct {
	Contracts ContractRepository
	Clock     Clock
}

type GetHourlyWorkSummaryInput struct {
	ContractID int64
	ActorID    uuid.UUID
	WeekStart  time.Time
}

type GetHourlyWorkSummaryOutput struct {
	Summary domain.HourlyWorkSummary
}

func (uc *GetHourlyWorkSummary) Execute(ctx context.Context, in GetHourlyWorkSummaryInput) (GetHourlyWorkSummaryOutput, error) {
	if uc.Contracts == nil || uc.Clock == nil {
		return GetHourlyWorkSummaryOutput{}, fmt.Errorf("hourly summary dependencies are not configured")
	}
	if in.ContractID <= 0 || in.ActorID == uuid.Nil {
		return GetHourlyWorkSummaryOutput{}, fmt.Errorf("contract_id and actor_id are required")
	}
	contract, err := uc.Contracts.GetByIDForActor(ctx, in.ContractID, in.ActorID)
	if err != nil {
		return GetHourlyWorkSummaryOutput{}, err
	}
	if contract.ContractType != domain.TypeHourly {
		return GetHourlyWorkSummaryOutput{}, fmt.Errorf("hourly summary is only available for hourly contracts")
	}
	anchor := in.WeekStart
	if anchor.IsZero() {
		anchor = uc.Clock.Now()
	}
	weekStart, weekEnd := weekBounds(anchor)
	logs, err := uc.Contracts.ListHourlyLogsForActorInRange(ctx, in.ContractID, in.ActorID, weekStart, weekEnd)
	if err != nil {
		return GetHourlyWorkSummaryOutput{}, err
	}
	summary := domain.HourlyWorkSummary{
		ContractID:      in.ContractID,
		WeekStart:       weekStart,
		WeekEnd:         weekEnd,
		WeeklyHourLimit: contract.WeeklyHourLimit,
		HourlyRate:      contract.HourlyRate,
	}
	for _, log := range logs {
		switch log.Status {
		case domain.HourlyLogStatusPending:
			summary.PendingMinutes += log.DurationMin
		case domain.HourlyLogStatusApproved:
			summary.ApprovedMinutes += log.DurationMin
		case domain.HourlyLogStatusRejected:
			summary.RejectedMinutes += log.DurationMin
		}
	}
	summary.BillableMinutes = summary.PendingMinutes + summary.ApprovedMinutes
	if summary.WeeklyHourLimit > 0 {
		remaining := summary.WeeklyHourLimit*60 - summary.BillableMinutes
		if remaining > 0 {
			summary.RemainingMinutes = remaining
		}
	}
	summary.EstimatedBillableAmount = float64(summary.BillableMinutes) / 60 * summary.HourlyRate
	return GetHourlyWorkSummaryOutput{Summary: summary}, nil
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

type UpdateHourlyLog struct {
	Contracts ContractRepository
	Clock     Clock
}

type UpdateHourlyLogInput struct {
	HourlyLogID  int64
	FreelancerID uuid.UUID
	StartAt      time.Time
	EndAt        time.Time
	Note         string
	EvidenceURLs []string
}

type DeleteHourlyLog struct {
	Contracts ContractRepository
	Clock     Clock
}

type DeleteHourlyLogInput struct {
	HourlyLogID  int64
	FreelancerID uuid.UUID
}

type GetHourlyInvoice struct {
	Contracts ContractRepository
}

type GetHourlyInvoiceInput struct {
	InvoiceID int64
	ActorID   uuid.UUID
}

type ListHourlyInvoices struct {
	Contracts ContractRepository
}

type ListHourlyInvoicesInput struct {
	ContractID int64
	ActorID    uuid.UUID
	PageSize   int32
	PageToken  string
}

type ListHourlyInvoicesOutput struct {
	Invoices      []domain.HourlyInvoice
	NextPageToken string
}

type InternalCloseHourlyWeek struct {
	Contracts ContractRepository
	Clock     Clock
}

type InternalCloseHourlyWeekInput struct {
	ContractID int64
	WeekStart  time.Time
}

type InternalSettleHourlyInvoice struct {
	Contracts ContractRepository
	Disputes  DisputeReader
	Clock     Clock
}

const hourlyInvoiceReviewWindow = 120 * time.Hour

type InternalSettleHourlyInvoiceInput struct {
	InvoiceID int64
}

type CreateContractBonus struct {
	Contracts ContractRepository
	Clock     Clock
}

type CreateContractBonusInput struct {
	ContractID  int64
	ClientID    uuid.UUID
	AmountMinor int64
	Note        string
}

type ListContractBonuses struct {
	Contracts ContractRepository
}

type InternalMarkContractBonusPaid struct {
	Contracts ContractRepository
	Clock     Clock
}

type ListContractBonusesInput struct {
	ContractID int64
	ActorID    uuid.UUID
	PageSize   int32
	PageToken  string
}

type ListContractBonusesOutput struct {
	Bonuses       []domain.ContractBonus
	NextPageToken string
}

type InternalMarkContractBonusPaidInput struct {
	BonusID            int64
	PaymentReferenceID string
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
	current, err := uc.Contracts.GetHourlyLogForActor(ctx, in.HourlyLogID, in.ClientID)
	if err != nil {
		return ReviewHourlyLogOutput{}, err
	}
	if current.Status != domain.HourlyLogStatusPending {
		return ReviewHourlyLogOutput{}, fmt.Errorf("can only review pending hourly logs")
	}
	contract, err := uc.Contracts.GetByIDForActor(ctx, current.ContractID, in.ClientID)
	if err != nil {
		return ReviewHourlyLogOutput{}, err
	}
	if contract.ContractType != domain.TypeHourly {
		return ReviewHourlyLogOutput{}, fmt.Errorf("hourly logs can only be reviewed for hourly contracts")
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

func (uc *UpdateHourlyLog) Execute(ctx context.Context, in UpdateHourlyLogInput) (LogHourlyWorkOutput, error) {
	if uc.Contracts == nil || uc.Clock == nil {
		return LogHourlyWorkOutput{}, fmt.Errorf("hourly log dependencies are not configured")
	}
	if in.HourlyLogID <= 0 || in.FreelancerID == uuid.Nil {
		return LogHourlyWorkOutput{}, fmt.Errorf("hourly_log_id and freelancer_id are required")
	}
	current, err := uc.Contracts.GetHourlyLogForActor(ctx, in.HourlyLogID, in.FreelancerID)
	if err != nil {
		return LogHourlyWorkOutput{}, err
	}
	if current.InvoiceID != 0 {
		return LogHourlyWorkOutput{}, fmt.Errorf("invoiced hourly logs cannot be edited")
	}
	if current.Status != domain.HourlyLogStatusPending {
		return LogHourlyWorkOutput{}, fmt.Errorf("only pending hourly logs can be edited")
	}
	log, err := validateHourlyLogWindow(ctx, uc.Contracts, uc.Clock, current.ContractID, in.FreelancerID, in.StartAt, in.EndAt, in.Note, in.EvidenceURLs, current.ID)
	if err != nil {
		return LogHourlyWorkOutput{}, err
	}
	log.ID = current.ID
	log.CreatedAt = current.CreatedAt
	if err := uc.Contracts.UpdateHourlyLogForFreelancer(ctx, log); err != nil {
		return LogHourlyWorkOutput{}, err
	}
	persisted, err := uc.Contracts.GetHourlyLogForActor(ctx, current.ID, in.FreelancerID)
	if err != nil {
		return LogHourlyWorkOutput{}, err
	}
	return LogHourlyWorkOutput{HourlyLog: persisted}, nil
}

func (uc *DeleteHourlyLog) Execute(ctx context.Context, in DeleteHourlyLogInput) error {
	if uc.Contracts == nil || uc.Clock == nil {
		return fmt.Errorf("hourly log dependencies are not configured")
	}
	if in.HourlyLogID <= 0 || in.FreelancerID == uuid.Nil {
		return fmt.Errorf("hourly_log_id and freelancer_id are required")
	}
	current, err := uc.Contracts.GetHourlyLogForActor(ctx, in.HourlyLogID, in.FreelancerID)
	if err != nil {
		return err
	}
	if current.InvoiceID != 0 {
		return fmt.Errorf("invoiced hourly logs cannot be deleted")
	}
	weekStart, weekEnd := weekBounds(uc.Clock.Now())
	if current.StartAt.Before(weekStart) || !current.EndAt.Before(weekEnd) {
		return fmt.Errorf("only current-week hourly logs can be deleted")
	}
	return uc.Contracts.DeleteHourlyLogForFreelancer(ctx, in.HourlyLogID, in.FreelancerID)
}

func (uc *GetHourlyInvoice) Execute(ctx context.Context, in GetHourlyInvoiceInput) (domain.HourlyInvoice, error) {
	if uc.Contracts == nil {
		return domain.HourlyInvoice{}, fmt.Errorf("hourly invoice dependencies are not configured")
	}
	if in.InvoiceID <= 0 || in.ActorID == uuid.Nil {
		return domain.HourlyInvoice{}, fmt.Errorf("invoice_id and actor_id are required")
	}
	return uc.Contracts.GetHourlyInvoiceForActor(ctx, in.InvoiceID, in.ActorID)
}

func (uc *ListHourlyInvoices) Execute(ctx context.Context, in ListHourlyInvoicesInput) (ListHourlyInvoicesOutput, error) {
	if uc.Contracts == nil {
		return ListHourlyInvoicesOutput{}, fmt.Errorf("hourly invoice dependencies are not configured")
	}
	if in.ContractID <= 0 || in.ActorID == uuid.Nil {
		return ListHourlyInvoicesOutput{}, fmt.Errorf("contract_id and actor_id are required")
	}
	limit, offset, err := pagination(in.PageSize, in.PageToken)
	if err != nil {
		return ListHourlyInvoicesOutput{}, err
	}
	items, err := uc.Contracts.ListHourlyInvoicesForActor(ctx, in.ContractID, in.ActorID, limit, offset)
	if err != nil {
		return ListHourlyInvoicesOutput{}, err
	}
	next := ""
	if len(items) == limit {
		next = strconv.Itoa(offset + len(items))
	}
	return ListHourlyInvoicesOutput{Invoices: items, NextPageToken: next}, nil
}

func (uc *InternalCloseHourlyWeek) Execute(ctx context.Context, in InternalCloseHourlyWeekInput) (domain.HourlyInvoice, error) {
	if uc.Contracts == nil || uc.Clock == nil {
		return domain.HourlyInvoice{}, fmt.Errorf("hourly invoice dependencies are not configured")
	}
	if in.ContractID <= 0 {
		return domain.HourlyInvoice{}, fmt.Errorf("contract_id is required")
	}
	contract, err := uc.Contracts.GetByID(ctx, in.ContractID)
	if err != nil {
		return domain.HourlyInvoice{}, err
	}
	if contract.ContractType != domain.TypeHourly {
		return domain.HourlyInvoice{}, fmt.Errorf("hourly invoices are only available for hourly contracts")
	}
	anchor := in.WeekStart
	if anchor.IsZero() {
		anchor = uc.Clock.Now().AddDate(0, 0, -7)
	}
	weekStart, weekEnd := weekBounds(anchor)
	existing, err := uc.Contracts.GetHourlyInvoiceByContractWeek(ctx, contract.ID, weekStart)
	if err == nil {
		return existing, nil
	}
	if !strings.Contains(strings.ToLower(err.Error()), "not found") {
		return domain.HourlyInvoice{}, err
	}
	logs, err := uc.Contracts.ListHourlyLogsForActorInRange(ctx, contract.ID, contract.ClientID, weekStart, weekEnd)
	if err != nil {
		return domain.HourlyInvoice{}, err
	}
	var billable int32
	for _, log := range logs {
		if log.Status == domain.HourlyLogStatusPending || log.Status == domain.HourlyLogStatusApproved {
			billable += log.DurationMin
		}
	}
	now := uc.Clock.Now()
	invoice := domain.HourlyInvoice{
		ContractID:      contract.ID,
		ClientID:        contract.ClientID,
		FreelancerID:    contract.FreelancerID,
		WeekStart:       weekStart,
		WeekEnd:         weekEnd,
		Status:          domain.HourlyInvoiceStatusInReview,
		BillableMinutes: billable,
		HourlyRate:      contract.HourlyRate,
		AmountMinor:     0,
		CreatedAt:       now,
		SubmittedAt:     &now,
	}
	estimatedAmount := (float64(billable) / 60) * contract.HourlyRate
	invoice.AmountMinor, err = domain.MoneyToMinorUnits(estimatedAmount, "hourly invoice amount")
	if err != nil {
		return domain.HourlyInvoice{}, err
	}
	id, err := uc.Contracts.CreateHourlyInvoice(ctx, invoice)
	if err != nil {
		return domain.HourlyInvoice{}, err
	}
	_ = uc.Contracts.AppendStatusHistory(ctx, domain.StatusHistoryEntry{ContractID: contract.ID, EventType: domain.StatusHistoryEventHourlyInvoiceCreated, ActorID: contract.ClientID, CreatedAt: now})
	return uc.Contracts.GetHourlyInvoice(ctx, id)
}

func (uc *InternalSettleHourlyInvoice) Execute(ctx context.Context, in InternalSettleHourlyInvoiceInput) (domain.HourlyInvoice, error) {
	if uc.Contracts == nil || uc.Clock == nil {
		return domain.HourlyInvoice{}, fmt.Errorf("hourly invoice dependencies are not configured")
	}
	if in.InvoiceID <= 0 {
		return domain.HourlyInvoice{}, fmt.Errorf("invoice_id is required")
	}
	invoice, err := uc.Contracts.GetHourlyInvoice(ctx, in.InvoiceID)
	if err != nil {
		return domain.HourlyInvoice{}, err
	}
	now := uc.Clock.Now().UTC()
	if invoice.WeekEnd.IsZero() {
		return domain.HourlyInvoice{}, fmt.Errorf("hourly invoice week_end is required")
	}
	reviewWindowEndsAt := invoice.WeekEnd.UTC().Add(hourlyInvoiceReviewWindow)
	if now.Before(reviewWindowEndsAt) {
		return domain.HourlyInvoice{}, fmt.Errorf("hourly invoice is still in review window until %s", reviewWindowEndsAt.Format(time.RFC3339))
	}
	if invoice.Status == domain.HourlyInvoiceStatusDisputed {
		return domain.HourlyInvoice{}, fmt.Errorf("disputed hourly invoice cannot be settled")
	}
	if invoice.Status != domain.HourlyInvoiceStatusSubmitted &&
		invoice.Status != domain.HourlyInvoiceStatusInReview &&
		invoice.Status != domain.HourlyInvoiceStatusApproved {
		return domain.HourlyInvoice{}, fmt.Errorf("hourly invoice status %q cannot be settled", invoice.Status)
	}
	if uc.Disputes != nil {
		openDisputeID, err := uc.Disputes.GetOpenDisputeID(ctx, "hourly_invoice", strconv.FormatInt(invoice.ID, 10))
		if err != nil {
			return domain.HourlyInvoice{}, err
		}
		if strings.TrimSpace(openDisputeID) != "" {
			_ = uc.Contracts.MarkHourlyInvoiceStatus(ctx, invoice.ID, domain.HourlyInvoiceStatusDisputed, openDisputeID, now)
			_ = uc.Contracts.AppendStatusHistory(ctx, domain.StatusHistoryEntry{ContractID: invoice.ContractID, EventType: domain.StatusHistoryEventHourlyInvoiceDisputed, ActorID: invoice.ClientID, CreatedAt: now})
			return domain.HourlyInvoice{}, fmt.Errorf("open dispute exists for hourly invoice")
		}
	}
	if invoice.Status == domain.HourlyInvoiceStatusInReview || invoice.Status == domain.HourlyInvoiceStatusSubmitted {
		if err := uc.Contracts.MarkHourlyInvoiceStatus(ctx, invoice.ID, domain.HourlyInvoiceStatusApproved, "", now); err != nil {
			return domain.HourlyInvoice{}, err
		}
	}
	if err := uc.Contracts.MarkHourlyInvoiceStatus(ctx, invoice.ID, domain.HourlyInvoiceStatusPaid, "", now); err != nil {
		return domain.HourlyInvoice{}, err
	}
	_ = uc.Contracts.AppendStatusHistory(ctx, domain.StatusHistoryEntry{ContractID: invoice.ContractID, EventType: domain.StatusHistoryEventHourlyInvoicePaid, ActorID: invoice.ClientID, CreatedAt: now})
	return uc.Contracts.GetHourlyInvoice(ctx, invoice.ID)
}

func (uc *CreateContractBonus) Execute(ctx context.Context, in CreateContractBonusInput) (domain.ContractBonus, error) {
	if uc.Contracts == nil || uc.Clock == nil {
		return domain.ContractBonus{}, fmt.Errorf("bonus dependencies are not configured")
	}
	if in.ContractID <= 0 || in.ClientID == uuid.Nil || in.AmountMinor <= 0 {
		return domain.ContractBonus{}, fmt.Errorf("contract_id, client_id, and positive amount_minor are required")
	}
	contract, err := uc.Contracts.GetByIDForActor(ctx, in.ContractID, in.ClientID)
	if err != nil {
		return domain.ContractBonus{}, err
	}
	bonus := domain.ContractBonus{
		ContractID:   contract.ID,
		ClientID:     contract.ClientID,
		FreelancerID: contract.FreelancerID,
		AmountMinor:  in.AmountMinor,
		Note:         strings.TrimSpace(in.Note),
		Status:       domain.ContractBonusStatusPending,
		CreatedAt:    uc.Clock.Now(),
	}
	id, err := uc.Contracts.CreateContractBonus(ctx, bonus)
	if err != nil {
		return domain.ContractBonus{}, err
	}
	return uc.Contracts.GetContractBonusForActor(ctx, id, in.ClientID)
}

func (uc *ListContractBonuses) Execute(ctx context.Context, in ListContractBonusesInput) (ListContractBonusesOutput, error) {
	if uc.Contracts == nil {
		return ListContractBonusesOutput{}, fmt.Errorf("bonus dependencies are not configured")
	}
	if in.ContractID <= 0 || in.ActorID == uuid.Nil {
		return ListContractBonusesOutput{}, fmt.Errorf("contract_id and actor_id are required")
	}
	limit, offset, err := pagination(in.PageSize, in.PageToken)
	if err != nil {
		return ListContractBonusesOutput{}, err
	}
	items, err := uc.Contracts.ListContractBonusesForActor(ctx, in.ContractID, in.ActorID, limit, offset)
	if err != nil {
		return ListContractBonusesOutput{}, err
	}
	next := ""
	if len(items) == limit {
		next = strconv.Itoa(offset + len(items))
	}
	return ListContractBonusesOutput{Bonuses: items, NextPageToken: next}, nil
}

func (uc *InternalMarkContractBonusPaid) Execute(ctx context.Context, in InternalMarkContractBonusPaidInput) (domain.ContractBonus, error) {
	if uc.Contracts == nil || uc.Clock == nil {
		return domain.ContractBonus{}, fmt.Errorf("bonus dependencies are not configured")
	}
	if in.BonusID <= 0 {
		return domain.ContractBonus{}, fmt.Errorf("bonus_id is required")
	}
	bonus, err := uc.Contracts.GetContractBonus(ctx, in.BonusID)
	if err != nil {
		return domain.ContractBonus{}, err
	}
	if bonus.Status == domain.ContractBonusStatusPaid {
		return bonus, nil
	}
	if bonus.Status != domain.ContractBonusStatusPending {
		return domain.ContractBonus{}, fmt.Errorf("only pending bonuses can be marked paid")
	}
	paymentReferenceID := strings.TrimSpace(in.PaymentReferenceID)
	if paymentReferenceID == "" {
		return domain.ContractBonus{}, fmt.Errorf("payment_reference_id is required")
	}
	if err := uc.Contracts.MarkContractBonusStatus(ctx, in.BonusID, domain.ContractBonusStatusPaid, paymentReferenceID, uc.Clock.Now()); err != nil {
		return domain.ContractBonus{}, err
	}
	return uc.Contracts.GetContractBonus(ctx, in.BonusID)
}

func weekBounds(t time.Time) (time.Time, time.Time) {
	d := t.UTC()
	day := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
	offset := (int(day.Weekday()) + 6) % 7
	start := day.AddDate(0, 0, -offset)
	return start, start.AddDate(0, 0, 7)
}

func timeRangesOverlap(startA, endA, startB, endB time.Time) bool {
	return startA.Before(endB) && endA.After(startB)
}

func validateHourlyLogWindow(ctx context.Context, repo ContractRepository, clock Clock, contractID int64, freelancerID uuid.UUID, start time.Time, end time.Time, note string, evidenceURLs []string, excludeLogID int64) (domain.HourlyLog, error) {
	if !end.After(start) {
		return domain.HourlyLog{}, fmt.Errorf("end_at must be after start_at")
	}
	startAt := start.UTC()
	endAt := end.UTC()
	now := clock.Now().UTC()
	if startAt.After(now) || endAt.After(now) {
		return domain.HourlyLog{}, fmt.Errorf("hourly work cannot be logged in the future")
	}
	contract, err := repo.GetByIDForActor(ctx, contractID, freelancerID)
	if err != nil {
		return domain.HourlyLog{}, err
	}
	if contract.ContractType != domain.TypeHourly {
		return domain.HourlyLog{}, fmt.Errorf("hourly work can only be logged for hourly contracts")
	}
	if contract.Status != domain.StatusActive {
		return domain.HourlyLog{}, fmt.Errorf("hourly work can only be logged for active contracts")
	}
	currentWeekStart, currentWeekEnd := weekBounds(now)
	if startAt.Before(currentWeekStart) || !endAt.Before(currentWeekEnd) {
		return domain.HourlyLog{}, fmt.Errorf("only current-week hourly logs can be edited")
	}
	duration := int32(endAt.Sub(startAt).Minutes())
	if duration <= 0 {
		return domain.HourlyLog{}, fmt.Errorf("duration must be positive")
	}
	weekStart, weekEnd := weekBounds(startAt)
	if endAt.After(weekEnd) {
		return domain.HourlyLog{}, fmt.Errorf("hourly log cannot cross weekly boundary")
	}
	weeklyLogs, err := repo.ListHourlyLogsForActorInRange(ctx, contractID, freelancerID, weekStart, weekEnd)
	if err != nil {
		return domain.HourlyLog{}, err
	}
	var usedMinutes int32
	for _, existing := range weeklyLogs {
		if existing.ID == excludeLogID || existing.Status == domain.HourlyLogStatusRejected {
			continue
		}
		if timeRangesOverlap(startAt, endAt, existing.StartAt.UTC(), existing.EndAt.UTC()) {
			return domain.HourlyLog{}, fmt.Errorf("hourly log overlaps existing work log")
		}
		usedMinutes += existing.DurationMin
	}
	if contract.WeeklyHourLimit > 0 && usedMinutes+duration > contract.WeeklyHourLimit*60 {
		return domain.HourlyLog{}, fmt.Errorf("weekly hour limit exceeded")
	}
	return domain.HourlyLog{
		ContractID:   contractID,
		FreelancerID: freelancerID,
		WorkDate:     time.Date(startAt.Year(), startAt.Month(), startAt.Day(), 0, 0, 0, 0, time.UTC),
		StartAt:      startAt,
		EndAt:        endAt,
		DurationMin:  duration,
		Note:         strings.TrimSpace(note),
		EvidenceURLs: normalizeEvidenceURLs(evidenceURLs),
		Status:       domain.HourlyLogStatusPending,
	}, nil
}

func normalizeEvidenceURLs(urls []string) []string {
	if len(urls) == 0 {
		return nil
	}
	out := make([]string, 0, len(urls))
	seen := make(map[string]struct{}, len(urls))
	for _, raw := range urls {
		v := strings.TrimSpace(raw)
		if v == "" {
			continue
		}
		if len(v) > 2048 {
			v = v[:2048]
		}
		if _, exists := seen[v]; exists {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
		if len(out) >= 20 {
			break
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func pagination(pageSize int32, pageToken string) (int, int, error) {
	limit := int(pageSize)
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := 0
	if strings.TrimSpace(pageToken) != "" {
		v, err := strconv.Atoi(strings.TrimSpace(pageToken))
		if err != nil || v < 0 {
			return 0, 0, fmt.Errorf("invalid page_token")
		}
		offset = v
	}
	return limit, offset, nil
}
