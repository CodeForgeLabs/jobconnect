package application

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"jobconnect/contract/internal/domain"

	"github.com/google/uuid"
)

type hourlyLogRepoStub struct {
	createContractRepoStub

	contract    domain.Contract
	hourlyLog   domain.HourlyLog
	rangeLogs   []domain.HourlyLog
	invoice     domain.HourlyInvoice
	bonus       domain.ContractBonus
	createdLogs []domain.HourlyLog
	reviewCalls int
}

func (r *hourlyLogRepoStub) GetByID(_ context.Context, contractID int64) (domain.Contract, error) {
	if r.contract.ID != contractID {
		return domain.Contract{}, fmt.Errorf("not found")
	}
	return r.contract, nil
}

func (r *hourlyLogRepoStub) GetByIDForActor(_ context.Context, contractID int64, actorID uuid.UUID) (domain.Contract, error) {
	if r.contract.ID != contractID {
		return domain.Contract{}, fmt.Errorf("not found")
	}
	if actorID != r.contract.ClientID && actorID != r.contract.FreelancerID {
		return domain.Contract{}, fmt.Errorf("not found")
	}
	return r.contract, nil
}

func (r *hourlyLogRepoStub) CreateHourlyLogForFreelancer(_ context.Context, log domain.HourlyLog) (int64, error) {
	r.createdLogs = append(r.createdLogs, log)
	r.hourlyLog = log
	r.hourlyLog.ID = 55
	return 55, nil
}

func (r *hourlyLogRepoStub) ListHourlyLogsForActorInRange(_ context.Context, contractID int64, actorID uuid.UUID, startAt time.Time, endAt time.Time) ([]domain.HourlyLog, error) {
	if r.contract.ID != contractID {
		return nil, fmt.Errorf("not found")
	}
	if actorID != r.contract.ClientID && actorID != r.contract.FreelancerID {
		return nil, fmt.Errorf("not found")
	}
	out := make([]domain.HourlyLog, 0, len(r.rangeLogs))
	for _, log := range r.rangeLogs {
		if log.StartAt.Before(endAt) && log.EndAt.After(startAt) {
			out = append(out, log)
		}
	}
	return out, nil
}

func (r *hourlyLogRepoStub) GetHourlyLogForActor(_ context.Context, hourlyLogID int64, actorID uuid.UUID) (domain.HourlyLog, error) {
	if r.hourlyLog.ID != hourlyLogID {
		return domain.HourlyLog{}, fmt.Errorf("not found")
	}
	if actorID != r.contract.ClientID && actorID != r.contract.FreelancerID {
		return domain.HourlyLog{}, fmt.Errorf("not found")
	}
	return r.hourlyLog, nil
}

func (r *hourlyLogRepoStub) ReviewHourlyLogForClient(_ context.Context, hourlyLogID int64, clientID uuid.UUID, status string, note string, at time.Time) error {
	r.reviewCalls++
	if hourlyLogID != r.hourlyLog.ID || clientID != r.contract.ClientID {
		return fmt.Errorf("not found")
	}
	r.hourlyLog.Status = status
	r.hourlyLog.ReviewNote = note
	r.hourlyLog.ClientReviewAt = &at
	return nil
}

func (r *hourlyLogRepoStub) UpdateHourlyLogForFreelancer(_ context.Context, log domain.HourlyLog) error {
	if r.hourlyLog.ID != log.ID || r.hourlyLog.FreelancerID != log.FreelancerID {
		return fmt.Errorf("not found")
	}
	r.hourlyLog.StartAt = log.StartAt
	r.hourlyLog.EndAt = log.EndAt
	r.hourlyLog.WorkDate = log.WorkDate
	r.hourlyLog.DurationMin = log.DurationMin
	r.hourlyLog.Note = log.Note
	return nil
}

func (r *hourlyLogRepoStub) DeleteHourlyLogForFreelancer(_ context.Context, hourlyLogID int64, freelancerID uuid.UUID) error {
	if r.hourlyLog.ID != hourlyLogID || r.hourlyLog.FreelancerID != freelancerID {
		return fmt.Errorf("not found")
	}
	r.hourlyLog = domain.HourlyLog{}
	return nil
}

func (r *hourlyLogRepoStub) CreateHourlyInvoice(_ context.Context, invoice domain.HourlyInvoice) (int64, error) {
	r.invoice = invoice
	r.invoice.ID = 77
	for i := range r.rangeLogs {
		if r.rangeLogs[i].ContractID == invoice.ContractID &&
			!r.rangeLogs[i].StartAt.Before(invoice.WeekStart) &&
			!r.rangeLogs[i].EndAt.After(invoice.WeekEnd) &&
			r.rangeLogs[i].Status != domain.HourlyLogStatusRejected &&
			r.rangeLogs[i].InvoiceID == 0 {
			r.rangeLogs[i].InvoiceID = r.invoice.ID
		}
	}
	return 77, nil
}

func (r *hourlyLogRepoStub) GetHourlyInvoice(_ context.Context, invoiceID int64) (domain.HourlyInvoice, error) {
	if r.invoice.ID != invoiceID {
		return domain.HourlyInvoice{}, fmt.Errorf("not found")
	}
	return r.invoice, nil
}

func (r *hourlyLogRepoStub) GetHourlyInvoiceByContractWeek(_ context.Context, contractID int64, weekStart time.Time) (domain.HourlyInvoice, error) {
	if r.invoice.ID == 0 {
		return domain.HourlyInvoice{}, fmt.Errorf("not found")
	}
	if r.invoice.ContractID != contractID || !r.invoice.WeekStart.Equal(weekStart) {
		return domain.HourlyInvoice{}, fmt.Errorf("not found")
	}
	return r.invoice, nil
}

func (r *hourlyLogRepoStub) AttachHourlyLogsToInvoice(_ context.Context, contractID int64, startAt time.Time, endAt time.Time, invoiceID int64) error {
	for i := range r.rangeLogs {
		if r.rangeLogs[i].ContractID == contractID && !r.rangeLogs[i].StartAt.Before(startAt) && !r.rangeLogs[i].EndAt.After(endAt) && r.rangeLogs[i].Status != domain.HourlyLogStatusRejected {
			r.rangeLogs[i].InvoiceID = invoiceID
		}
	}
	return nil
}

func (r *hourlyLogRepoStub) MarkHourlyInvoiceStatus(_ context.Context, invoiceID int64, status string, disputeID string, at time.Time) error {
	if r.invoice.ID != invoiceID {
		return fmt.Errorf("not found")
	}
	r.invoice.Status = status
	if disputeID != "" {
		r.invoice.DisputeID = disputeID
	}
	if status == domain.HourlyInvoiceStatusPaid {
		r.invoice.PaidAt = &at
	}
	if status == domain.HourlyInvoiceStatusApproved {
		r.invoice.ApprovedAt = &at
	}
	return nil
}

func (r *hourlyLogRepoStub) GetContractBonus(_ context.Context, bonusID int64) (domain.ContractBonus, error) {
	if r.bonus.ID != bonusID {
		return domain.ContractBonus{}, fmt.Errorf("not found")
	}
	return r.bonus, nil
}

func (r *hourlyLogRepoStub) MarkContractBonusStatus(_ context.Context, bonusID int64, status string, paymentReferenceID string, at time.Time) error {
	if r.bonus.ID != bonusID {
		return fmt.Errorf("not found")
	}
	r.bonus.Status = status
	r.bonus.PaymentReferenceID = paymentReferenceID
	if status == domain.ContractBonusStatusPaid {
		r.bonus.PaidAt = &at
	}
	return nil
}

func TestLogHourlyWork_RequiresActiveHourlyContract(t *testing.T) {
	freelancerID := uuid.New()
	repo := &hourlyLogRepoStub{
		contract: domain.Contract{
			ID:           10,
			FreelancerID: freelancerID,
			ContractType: domain.TypeFixed,
			Status:       domain.StatusActive,
		},
	}
	uc := &LogHourlyWork{Contracts: repo, Clock: contractClockStub{now: time.Unix(1700000000, 0).UTC()}}

	_, err := uc.Execute(context.Background(), LogHourlyWorkInput{
		ContractID:   10,
		FreelancerID: freelancerID,
		StartAt:      time.Unix(1699992800, 0).UTC(),
		EndAt:        time.Unix(1699996400, 0).UTC(),
	})
	if err == nil || !strings.Contains(err.Error(), "hourly contracts") {
		t.Fatalf("expected hourly contract error, got %v", err)
	}
}

func TestLogHourlyWork_CreatesPendingLogForActiveHourlyContract(t *testing.T) {
	freelancerID := uuid.New()
	repo := &hourlyLogRepoStub{
		contract: domain.Contract{
			ID:           10,
			FreelancerID: freelancerID,
			ContractType: domain.TypeHourly,
			Status:       domain.StatusActive,
		},
	}
	uc := &LogHourlyWork{Contracts: repo, Clock: contractClockStub{now: time.Unix(1700000000, 0).UTC()}}

	out, err := uc.Execute(context.Background(), LogHourlyWorkInput{
		ContractID:   10,
		FreelancerID: freelancerID,
		StartAt:      time.Unix(1699992800, 0).UTC(),
		EndAt:        time.Unix(1699996400, 0).UTC(),
		Note:         " work ",
	})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if out.HourlyLog.Status != domain.HourlyLogStatusPending {
		t.Fatalf("expected pending log, got %q", out.HourlyLog.Status)
	}
	if len(repo.createdLogs) != 1 || repo.createdLogs[0].Note != "work" {
		t.Fatalf("unexpected created logs: %+v", repo.createdLogs)
	}
}

func TestLogHourlyWork_RejectsOverlappingLog(t *testing.T) {
	freelancerID := uuid.New()
	now := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	existingStart := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	repo := &hourlyLogRepoStub{
		contract: domain.Contract{
			ID:           10,
			FreelancerID: freelancerID,
			ContractType: domain.TypeHourly,
			Status:       domain.StatusActive,
		},
		rangeLogs: []domain.HourlyLog{{
			ContractID:  10,
			StartAt:     existingStart,
			EndAt:       existingStart.Add(2 * time.Hour),
			DurationMin: 120,
			Status:      domain.HourlyLogStatusPending,
		}},
	}
	uc := &LogHourlyWork{Contracts: repo, Clock: contractClockStub{now: now}}

	_, err := uc.Execute(context.Background(), LogHourlyWorkInput{
		ContractID:   10,
		FreelancerID: freelancerID,
		StartAt:      existingStart.Add(time.Hour),
		EndAt:        existingStart.Add(3 * time.Hour),
	})
	if err == nil || !strings.Contains(err.Error(), "overlaps") {
		t.Fatalf("expected overlap error, got %v", err)
	}
	if len(repo.createdLogs) != 0 {
		t.Fatalf("expected no created log, got %+v", repo.createdLogs)
	}
}

func TestLogHourlyWork_EnforcesWeeklyLimit(t *testing.T) {
	freelancerID := uuid.New()
	now := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	weekStart := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)
	repo := &hourlyLogRepoStub{
		contract: domain.Contract{
			ID:              10,
			FreelancerID:    freelancerID,
			ContractType:    domain.TypeHourly,
			Status:          domain.StatusActive,
			WeeklyHourLimit: 2,
		},
		rangeLogs: []domain.HourlyLog{{
			ContractID:  10,
			StartAt:     weekStart.Add(9 * time.Hour),
			EndAt:       weekStart.Add(10*time.Hour + 30*time.Minute),
			DurationMin: 90,
			Status:      domain.HourlyLogStatusApproved,
		}},
	}
	uc := &LogHourlyWork{Contracts: repo, Clock: contractClockStub{now: now}}

	_, err := uc.Execute(context.Background(), LogHourlyWorkInput{
		ContractID:   10,
		FreelancerID: freelancerID,
		StartAt:      weekStart.Add(24 * time.Hour),
		EndAt:        weekStart.Add(25 * time.Hour),
	})
	if err == nil || !strings.Contains(err.Error(), "weekly hour limit") {
		t.Fatalf("expected weekly limit error, got %v", err)
	}
	if len(repo.createdLogs) != 0 {
		t.Fatalf("expected no created log, got %+v", repo.createdLogs)
	}
}

func TestLogHourlyWork_RejectsNonCurrentWeek(t *testing.T) {
	freelancerID := uuid.New()
	now := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	oldStart := time.Date(2026, 4, 10, 9, 0, 0, 0, time.UTC)
	repo := &hourlyLogRepoStub{
		contract: domain.Contract{
			ID:           10,
			FreelancerID: freelancerID,
			ContractType: domain.TypeHourly,
			Status:       domain.StatusActive,
		},
	}
	uc := &LogHourlyWork{Contracts: repo, Clock: contractClockStub{now: now}}

	_, err := uc.Execute(context.Background(), LogHourlyWorkInput{
		ContractID:   10,
		FreelancerID: freelancerID,
		StartAt:      oldStart,
		EndAt:        oldStart.Add(1 * time.Hour),
	})
	if err == nil || !strings.Contains(err.Error(), "current-week") {
		t.Fatalf("expected current-week restriction error, got %v", err)
	}
}

func TestGetHourlyWorkSummary_CountsWeeklyMinutes(t *testing.T) {
	clientID := uuid.New()
	freelancerID := uuid.New()
	weekStart := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)
	repo := &hourlyLogRepoStub{
		contract: domain.Contract{
			ID:              10,
			ClientID:        clientID,
			FreelancerID:    freelancerID,
			ContractType:    domain.TypeHourly,
			Status:          domain.StatusActive,
			WeeklyHourLimit: 10,
			HourlyRate:      50,
		},
		rangeLogs: []domain.HourlyLog{
			{ContractID: 10, StartAt: weekStart.Add(9 * time.Hour), EndAt: weekStart.Add(10 * time.Hour), DurationMin: 60, Status: domain.HourlyLogStatusPending},
			{ContractID: 10, StartAt: weekStart.Add(24 * time.Hour), EndAt: weekStart.Add(26 * time.Hour), DurationMin: 120, Status: domain.HourlyLogStatusApproved},
			{ContractID: 10, StartAt: weekStart.Add(48 * time.Hour), EndAt: weekStart.Add(49 * time.Hour), DurationMin: 60, Status: domain.HourlyLogStatusRejected},
		},
	}
	uc := &GetHourlyWorkSummary{Contracts: repo, Clock: contractClockStub{now: weekStart}}

	out, err := uc.Execute(context.Background(), GetHourlyWorkSummaryInput{
		ContractID: 10,
		ActorID:    clientID,
		WeekStart:  weekStart.Add(2 * 24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if out.Summary.BillableMinutes != 180 || out.Summary.PendingMinutes != 60 || out.Summary.ApprovedMinutes != 120 || out.Summary.RejectedMinutes != 60 {
		t.Fatalf("unexpected summary minutes: %+v", out.Summary)
	}
	if out.Summary.RemainingMinutes != 420 {
		t.Fatalf("expected 420 remaining minutes, got %d", out.Summary.RemainingMinutes)
	}
	if out.Summary.EstimatedBillableAmount != 150 {
		t.Fatalf("expected estimated billable amount 150, got %.2f", out.Summary.EstimatedBillableAmount)
	}
}

func TestInternalCloseHourlyWeek_CreatesInvoiceExcludingRejectedLogs(t *testing.T) {
	clientID := uuid.New()
	freelancerID := uuid.New()
	weekStart := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)
	repo := &hourlyLogRepoStub{
		contract: domain.Contract{
			ID:           10,
			ClientID:     clientID,
			FreelancerID: freelancerID,
			ContractType: domain.TypeHourly,
			Status:       domain.StatusActive,
			HourlyRate:   50,
		},
		rangeLogs: []domain.HourlyLog{
			{ContractID: 10, StartAt: weekStart.Add(9 * time.Hour), EndAt: weekStart.Add(10 * time.Hour), DurationMin: 60, Status: domain.HourlyLogStatusPending},
			{ContractID: 10, StartAt: weekStart.Add(24 * time.Hour), EndAt: weekStart.Add(26 * time.Hour), DurationMin: 120, Status: domain.HourlyLogStatusApproved},
			{ContractID: 10, StartAt: weekStart.Add(48 * time.Hour), EndAt: weekStart.Add(49 * time.Hour), DurationMin: 60, Status: domain.HourlyLogStatusRejected},
		},
	}
	uc := &InternalCloseHourlyWeek{Contracts: repo, Clock: contractClockStub{now: weekStart.AddDate(0, 0, 7)}}

	invoice, err := uc.Execute(context.Background(), InternalCloseHourlyWeekInput{ContractID: 10, WeekStart: weekStart})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if invoice.BillableMinutes != 180 || invoice.AmountMinor != 15000 {
		t.Fatalf("unexpected invoice totals: %+v", invoice)
	}
	if invoice.Status != domain.HourlyInvoiceStatusInReview {
		t.Fatalf("expected in_review invoice, got %q", invoice.Status)
	}
}

func TestInternalCloseHourlyWeek_ReturnsExistingInvoiceWithoutOverwrite(t *testing.T) {
	clientID := uuid.New()
	freelancerID := uuid.New()
	weekStart := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)
	repo := &hourlyLogRepoStub{
		contract: domain.Contract{
			ID:           10,
			ClientID:     clientID,
			FreelancerID: freelancerID,
			ContractType: domain.TypeHourly,
			Status:       domain.StatusActive,
			HourlyRate:   50,
		},
		invoice: domain.HourlyInvoice{
			ID:              77,
			ContractID:      10,
			ClientID:        clientID,
			FreelancerID:    freelancerID,
			WeekStart:       weekStart,
			WeekEnd:         weekStart.AddDate(0, 0, 7),
			Status:          domain.HourlyInvoiceStatusPaid,
			BillableMinutes: 180,
			AmountMinor:     15000,
		},
	}
	uc := &InternalCloseHourlyWeek{Contracts: repo, Clock: contractClockStub{now: weekStart.AddDate(0, 0, 7)}}

	invoice, err := uc.Execute(context.Background(), InternalCloseHourlyWeekInput{ContractID: 10, WeekStart: weekStart})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if invoice.ID != 77 || invoice.Status != domain.HourlyInvoiceStatusPaid {
		t.Fatalf("expected existing paid invoice, got %+v", invoice)
	}
}

func TestInternalSettleHourlyInvoice_BlocksOpenDispute(t *testing.T) {
	clientID := uuid.New()
	now := time.Unix(1700000000, 0).UTC()
	repo := &hourlyLogRepoStub{
		invoice: domain.HourlyInvoice{
			ID:         77,
			ContractID: 10,
			ClientID:   clientID,
			WeekEnd:    now.Add(-6 * 24 * time.Hour),
			Status:     domain.HourlyInvoiceStatusInReview,
		},
	}
	uc := &InternalSettleHourlyInvoice{Contracts: repo, Disputes: disputeReaderStub{hasOpen: true}, Clock: contractClockStub{now: now}}

	_, err := uc.Execute(context.Background(), InternalSettleHourlyInvoiceInput{InvoiceID: 77})
	if err == nil || !strings.Contains(err.Error(), "open dispute") {
		t.Fatalf("expected open dispute error, got %v", err)
	}
	if repo.invoice.Status != domain.HourlyInvoiceStatusDisputed {
		t.Fatalf("expected disputed invoice, got %q", repo.invoice.Status)
	}
	if repo.invoice.DisputeID == "" {
		t.Fatalf("expected dispute id to be persisted, got empty value")
	}
}

func TestInternalSettleHourlyInvoice_MarksPaid(t *testing.T) {
	clientID := uuid.New()
	now := time.Unix(1700000000, 0).UTC()
	repo := &hourlyLogRepoStub{
		invoice: domain.HourlyInvoice{
			ID:         77,
			ContractID: 10,
			ClientID:   clientID,
			WeekEnd:    now.Add(-6 * 24 * time.Hour),
			Status:     domain.HourlyInvoiceStatusInReview,
		},
	}
	uc := &InternalSettleHourlyInvoice{Contracts: repo, Disputes: disputeReaderStub{hasOpen: false}, Clock: contractClockStub{now: now}}

	invoice, err := uc.Execute(context.Background(), InternalSettleHourlyInvoiceInput{InvoiceID: 77})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if invoice.Status != domain.HourlyInvoiceStatusPaid || invoice.PaidAt == nil {
		t.Fatalf("expected paid invoice, got %+v", invoice)
	}
}

func TestInternalSettleHourlyInvoice_BlocksDuringReviewWindow(t *testing.T) {
	clientID := uuid.New()
	now := time.Unix(1700000000, 0).UTC()
	repo := &hourlyLogRepoStub{
		invoice: domain.HourlyInvoice{
			ID:         77,
			ContractID: 10,
			ClientID:   clientID,
			WeekEnd:    now.Add(-48 * time.Hour),
			Status:     domain.HourlyInvoiceStatusInReview,
		},
	}
	uc := &InternalSettleHourlyInvoice{Contracts: repo, Disputes: disputeReaderStub{hasOpen: false}, Clock: contractClockStub{now: now}}

	_, err := uc.Execute(context.Background(), InternalSettleHourlyInvoiceInput{InvoiceID: 77})
	if err == nil || !strings.Contains(err.Error(), "still in review window") {
		t.Fatalf("expected review window block, got %v", err)
	}
}

func TestInternalSettleHourlyInvoice_BlocksFailedInvoice(t *testing.T) {
	clientID := uuid.New()
	now := time.Unix(1700000000, 0).UTC()
	repo := &hourlyLogRepoStub{
		invoice: domain.HourlyInvoice{
			ID:         77,
			ContractID: 10,
			ClientID:   clientID,
			WeekEnd:    now.Add(-6 * 24 * time.Hour),
			Status:     domain.HourlyInvoiceStatusFailed,
		},
	}
	uc := &InternalSettleHourlyInvoice{Contracts: repo, Disputes: disputeReaderStub{hasOpen: false}, Clock: contractClockStub{now: now}}

	_, err := uc.Execute(context.Background(), InternalSettleHourlyInvoiceInput{InvoiceID: 77})
	if err == nil || !strings.Contains(err.Error(), "cannot be settled") {
		t.Fatalf("expected un-settleable status error, got %v", err)
	}
	if repo.invoice.Status != domain.HourlyInvoiceStatusFailed {
		t.Fatalf("expected failed status to remain unchanged, got %q", repo.invoice.Status)
	}
}

func TestReviewHourlyLog_RejectsAlreadyReviewedLog(t *testing.T) {
	clientID := uuid.New()
	freelancerID := uuid.New()
	repo := &hourlyLogRepoStub{
		contract: domain.Contract{
			ID:           10,
			ClientID:     clientID,
			FreelancerID: freelancerID,
			ContractType: domain.TypeHourly,
			Status:       domain.StatusActive,
		},
		hourlyLog: domain.HourlyLog{
			ID:         55,
			ContractID: 10,
			Status:     domain.HourlyLogStatusApproved,
		},
	}
	uc := &ReviewHourlyLog{Contracts: repo, Clock: contractClockStub{now: time.Unix(1700000000, 0).UTC()}}

	_, err := uc.Execute(context.Background(), ReviewHourlyLogInput{
		HourlyLogID: 55,
		ClientID:    clientID,
		Status:      domain.HourlyLogStatusRejected,
	})
	if err == nil || !strings.Contains(err.Error(), "pending") {
		t.Fatalf("expected pending-only review error, got %v", err)
	}
	if repo.reviewCalls != 0 {
		t.Fatalf("expected review write to be skipped, got %d calls", repo.reviewCalls)
	}
}

func TestInternalMarkContractBonusPaid_RequiresPaymentReferenceID(t *testing.T) {
	repo := &hourlyLogRepoStub{
		bonus: domain.ContractBonus{
			ID:     12,
			Status: domain.ContractBonusStatusPending,
		},
	}
	uc := &InternalMarkContractBonusPaid{Contracts: repo, Clock: contractClockStub{now: time.Unix(1700000000, 0).UTC()}}

	_, err := uc.Execute(context.Background(), InternalMarkContractBonusPaidInput{BonusID: 12})
	if err == nil || !strings.Contains(err.Error(), "payment_reference_id is required") {
		t.Fatalf("expected payment reference requirement error, got %v", err)
	}
}

func TestInternalMarkContractBonusPaid_PersistsPaymentReferenceID(t *testing.T) {
	repo := &hourlyLogRepoStub{
		bonus: domain.ContractBonus{
			ID:     12,
			Status: domain.ContractBonusStatusPending,
		},
	}
	uc := &InternalMarkContractBonusPaid{Contracts: repo, Clock: contractClockStub{now: time.Unix(1700000000, 0).UTC()}}

	bonus, err := uc.Execute(context.Background(), InternalMarkContractBonusPaidInput{
		BonusID:            12,
		PaymentReferenceID: "pay_abc123",
	})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if bonus.Status != domain.ContractBonusStatusPaid {
		t.Fatalf("expected paid bonus status, got %q", bonus.Status)
	}
	if bonus.PaymentReferenceID != "pay_abc123" {
		t.Fatalf("expected persisted payment reference, got %q", bonus.PaymentReferenceID)
	}
}
