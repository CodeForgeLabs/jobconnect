package db

import (
	"context"
	"fmt"
	"time"

	"jobconnect/verification/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VerificationRepo struct {
	pool *pgxpool.Pool
}

func NewVerificationRepo(pool *pgxpool.Pool) *VerificationRepo {
	return &VerificationRepo{pool: pool}
}

func (r *VerificationRepo) CreateSubmission(ctx context.Context, req domain.VerificationRequest) (domain.VerificationRequest, error) {
	row := r.pool.QueryRow(ctx, `
		insert into verification_requests (
			user_id, request_version, status, legal_name, country_code, document_type,
			document_number_masked, evidence_url, submission_note, submitted_at, updated_at
		)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$10)
		returning id, user_id, request_version, status, legal_name, country_code, document_type,
			document_number_masked, evidence_url, submission_note, submitted_at, reviewed_at,
			reviewer_user_id, rejection_reason, internal_note, reverify_due_at, updated_at
	`, req.UserID, req.RequestVersion, req.Status, req.LegalName, req.CountryCode, req.DocumentType, req.DocumentNumberMasked, req.EvidenceURL, req.SubmissionNote, req.SubmittedAt)
	return scanRequest(row)
}

func (r *VerificationRepo) GetLatestByUserID(ctx context.Context, userID uuid.UUID) (domain.VerificationRequest, error) {
	row := r.pool.QueryRow(ctx, `
		select id, user_id, request_version, status, legal_name, country_code, document_type,
			document_number_masked, evidence_url, submission_note, submitted_at, reviewed_at,
			reviewer_user_id, rejection_reason, internal_note, reverify_due_at, updated_at
		from verification_requests
		where user_id = $1
		order by request_version desc
		limit 1
	`, userID)
	return scanRequest(row)
}

func (r *VerificationRepo) GetByID(ctx context.Context, id int64) (domain.VerificationRequest, error) {
	row := r.pool.QueryRow(ctx, `
		select id, user_id, request_version, status, legal_name, country_code, document_type,
			document_number_masked, evidence_url, submission_note, submitted_at, reviewed_at,
			reviewer_user_id, rejection_reason, internal_note, reverify_due_at, updated_at
		from verification_requests
		where id = $1
	`, id)
	return scanRequest(row)
}

func (r *VerificationRepo) ListPending(ctx context.Context, limit, offset int32) ([]domain.VerificationRequest, error) {
	rows, err := r.pool.Query(ctx, `
		select id, user_id, request_version, status, legal_name, country_code, document_type,
			document_number_masked, evidence_url, submission_note, submitted_at, reviewed_at,
			reviewer_user_id, rejection_reason, internal_note, reverify_due_at, updated_at
		from verification_requests
		where status in ('submitted', 'pending_review')
		order by submitted_at asc
		limit $1 offset $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.VerificationRequest, 0)
	for rows.Next() {
		item, err := scanRequest(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *VerificationRepo) Review(ctx context.Context, requestID int64, reviewerID uuid.UUID, decision, rejectionReason, internalNote string, reviewedAt time.Time) (domain.VerificationRequest, error) {
	status := domain.StatusRejected
	if decision == domain.DecisionApprove {
		status = domain.StatusVerified
		rejectionReason = ""
	}

	row := r.pool.QueryRow(ctx, `
		update verification_requests
		set status = $2,
			reviewer_user_id = $3,
			reviewed_at = $4,
			rejection_reason = $5,
			internal_note = $6,
			updated_at = $4
		where id = $1 and status in ('submitted', 'pending_review')
		returning id, user_id, request_version, status, legal_name, country_code, document_type,
			document_number_masked, evidence_url, submission_note, submitted_at, reviewed_at,
			reviewer_user_id, rejection_reason, internal_note, reverify_due_at, updated_at
	`, requestID, status, reviewerID, reviewedAt, rejectionReason, internalNote)

	out, err := scanRequest(row)
	if err != nil {
		if err == ErrNotFound {
			return domain.VerificationRequest{}, fmt.Errorf("request not found or not pending")
		}
		return domain.VerificationRequest{}, err
	}
	return out, nil
}

func (r *VerificationRepo) MarkReverificationRequired(ctx context.Context, userID uuid.UUID, reviewerID uuid.UUID, reason string, dueAt time.Time, at time.Time) (domain.VerificationRequest, error) {
	row := r.pool.QueryRow(ctx, `
		with latest as (
			select id
			from verification_requests
			where user_id = $1
			order by request_version desc
			limit 1
		)
		update verification_requests vr
		set status = 'reverification_required',
			reviewer_user_id = $2,
			internal_note = $3,
			reverify_due_at = $4,
			updated_at = $5
		from latest
		where vr.id = latest.id and vr.status = 'verified'
		returning vr.id, vr.user_id, vr.request_version, vr.status, vr.legal_name, vr.country_code, vr.document_type,
			vr.document_number_masked, vr.evidence_url, vr.submission_note, vr.submitted_at, vr.reviewed_at,
			vr.reviewer_user_id, vr.rejection_reason, vr.internal_note, vr.reverify_due_at, vr.updated_at
	`, userID, reviewerID, reason, dueAt, at)
	out, err := scanRequest(row)
	if err != nil {
		if err == ErrNotFound {
			return domain.VerificationRequest{}, fmt.Errorf("only verified users can be moved to reverification_required")
		}
		return domain.VerificationRequest{}, err
	}
	return out, nil
}

func (r *VerificationRepo) AppendEvent(ctx context.Context, event domain.VerificationEvent) error {
	_, err := r.pool.Exec(ctx, `
		insert into verification_events (request_id, user_id, event_type, actor_user_id, details_json, created_at)
		values ($1, $2, $3, $4, $5::jsonb, $6)
	`, event.RequestID, event.UserID, event.EventType, event.ActorUserID, event.DetailsJSON, event.CreatedAt)
	return err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRequest(row rowScanner) (domain.VerificationRequest, error) {
	var out domain.VerificationRequest
	var evidenceURL string
	var reviewerID *uuid.UUID
	var reviewedAt *time.Time
	var reverifyDueAt *time.Time
	if err := row.Scan(
		&out.ID,
		&out.UserID,
		&out.RequestVersion,
		&out.Status,
		&out.LegalName,
		&out.CountryCode,
		&out.DocumentType,
		&out.DocumentNumberMasked,
		&evidenceURL,
		&out.SubmissionNote,
		&out.SubmittedAt,
		&reviewedAt,
		&reviewerID,
		&out.RejectionReason,
		&out.InternalNote,
		&reverifyDueAt,
		&out.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return domain.VerificationRequest{}, ErrNotFound
		}
		return domain.VerificationRequest{}, err
	}
	out.EvidenceURL = evidenceURL
	out.ReviewerUserID = reviewerID
	out.ReviewedAt = reviewedAt
	out.ReverifyDueAt = reverifyDueAt
	return out, nil
}
