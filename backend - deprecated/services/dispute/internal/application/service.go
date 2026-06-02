package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"jobconnect/dispute/internal/domain"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, d domain.Dispute) (int64, error)
	GetByID(ctx context.Context, disputeID int64) (domain.Dispute, error)
	List(ctx context.Context, referenceType, referenceID, status string, limit, offset int) ([]domain.Dispute, error)
	Resolve(ctx context.Context, disputeID int64, decision, note string, resolvedBy uuid.UUID, at time.Time) error
}

type WalletClient interface {
	GetHoldByReference(ctx context.Context, referenceType, referenceID string) (Hold, error)
	ReleaseHold(ctx context.Context, holdID int64, idempotencyKey, note string) error
	CaptureHold(ctx context.Context, holdID, amountMinor int64, idempotencyKey, referenceType, referenceID, note string) error
}

type Hold struct {
	ID            int64
	WalletID      int64
	AmountMinor   int64
	CapturedMinor int64
}

type Clock interface {
	Now() time.Time
}

type Service struct {
	Repo   Repository
	Wallet WalletClient
	Clock  Clock
}

func (s *Service) OpenDispute(ctx context.Context, referenceType, referenceID string, openedBy uuid.UUID, reason string) (domain.Dispute, error) {
	if s.Repo == nil || s.Clock == nil {
		return domain.Dispute{}, fmt.Errorf("dispute dependencies are not configured")
	}
	if openedBy == uuid.Nil {
		return domain.Dispute{}, fmt.Errorf("opened_by is required")
	}
	if err := domain.ValidateOpen(referenceType, referenceID, reason); err != nil {
		return domain.Dispute{}, err
	}
	now := s.Clock.Now()
	item := domain.Dispute{
		ReferenceType: strings.TrimSpace(referenceType),
		ReferenceID:   strings.TrimSpace(referenceID),
		OpenedBy:      openedBy,
		Reason:        strings.TrimSpace(reason),
		Status:        domain.StatusOpen,
		CreatedAt:     now,
	}
	id, err := s.Repo.Create(ctx, item)
	if err != nil {
		return domain.Dispute{}, err
	}
	return s.Repo.GetByID(ctx, id)
}

func (s *Service) GetDispute(ctx context.Context, disputeID int64) (domain.Dispute, error) {
	if s.Repo == nil {
		return domain.Dispute{}, fmt.Errorf("dispute dependencies are not configured")
	}
	if disputeID <= 0 {
		return domain.Dispute{}, fmt.Errorf("dispute_id is required")
	}
	return s.Repo.GetByID(ctx, disputeID)
}

func (s *Service) ListDisputes(ctx context.Context, referenceType, referenceID, status string, pageSize int32, pageToken string) ([]domain.Dispute, string, error) {
	if s.Repo == nil {
		return nil, "", fmt.Errorf("dispute dependencies are not configured")
	}
	limit := int(pageSize)
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := 0
	if strings.TrimSpace(pageToken) != "" {
		n, err := strconv.Atoi(strings.TrimSpace(pageToken))
		if err != nil || n < 0 {
			return nil, "", fmt.Errorf("invalid page_token")
		}
		offset = n
	}
	items, err := s.Repo.List(ctx, strings.TrimSpace(referenceType), strings.TrimSpace(referenceID), strings.TrimSpace(status), limit+1, offset)
	if err != nil {
		return nil, "", err
	}
	next := ""
	if len(items) > limit {
		next = strconv.Itoa(offset + limit)
		items = items[:limit]
	}
	return items, next, nil
}

func (s *Service) ResolveDispute(ctx context.Context, disputeID int64, decision, note string, resolvedBy uuid.UUID, resolverRole string) (domain.Dispute, error) {
	if s.Repo == nil || s.Wallet == nil || s.Clock == nil {
		return domain.Dispute{}, fmt.Errorf("dispute dependencies are not configured")
	}
	if resolvedBy == uuid.Nil {
		return domain.Dispute{}, fmt.Errorf("resolved_by is required")
	}
	role := strings.ToLower(strings.TrimSpace(resolverRole))
	if role != "admin" && role != "service" && role != "internal" {
		return domain.Dispute{}, fmt.Errorf("admin or internal role required")
	}
	if err := domain.ValidateDecision(decision); err != nil {
		return domain.Dispute{}, err
	}
	item, err := s.Repo.GetByID(ctx, disputeID)
	if err != nil {
		return domain.Dispute{}, err
	}
	if strings.ToLower(strings.TrimSpace(item.Status)) != domain.StatusOpen {
		return domain.Dispute{}, fmt.Errorf("dispute is not open")
	}

	hold, err := s.Wallet.GetHoldByReference(ctx, item.ReferenceType, item.ReferenceID)
	if err != nil {
		return domain.Dispute{}, fmt.Errorf("get hold by reference: %w", err)
	}
	remaining := hold.AmountMinor - hold.CapturedMinor
	if remaining <= 0 {
		return domain.Dispute{}, fmt.Errorf("hold has no remaining amount")
	}

	now := s.Clock.Now()
	idempotencyKey := fmt.Sprintf("dispute-resolve:%d:%s", disputeID, strings.ToLower(strings.TrimSpace(decision)))
	switch strings.ToLower(strings.TrimSpace(decision)) {
	case domain.DecisionRelease:
		if err := s.Wallet.CaptureHold(ctx, hold.ID, remaining, idempotencyKey, item.ReferenceType, item.ReferenceID, strings.TrimSpace(note)); err != nil {
			return domain.Dispute{}, fmt.Errorf("capture hold: %w", err)
		}
	case domain.DecisionRefund:
		if err := s.Wallet.ReleaseHold(ctx, hold.ID, idempotencyKey, strings.TrimSpace(note)); err != nil {
			return domain.Dispute{}, fmt.Errorf("release hold: %w", err)
		}
	}
	if err := s.Repo.Resolve(ctx, disputeID, strings.ToLower(strings.TrimSpace(decision)), strings.TrimSpace(note), resolvedBy, now); err != nil {
		return domain.Dispute{}, err
	}
	return s.Repo.GetByID(ctx, disputeID)
}
