package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"jobconnect/proposal/internal/application"
	"jobconnect/proposal/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProposalRepo struct {
	pool *pgxpool.Pool
}

func NewProposalRepo(pool *pgxpool.Pool) *ProposalRepo {
	return &ProposalRepo{pool: pool}
}

func (r *ProposalRepo) Create(ctx context.Context, p domain.Proposal) (int64, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var id int64
	err = tx.QueryRow(ctx, `
		insert into proposals (
			job_id, client_id, freelancer_id, cover_letter, bid_type, bid_amount, estimated_days,
			status, status_reason, created_at, updated_at, connects_spent
		)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		returning id
	`, p.JobID, p.ClientID, p.FreelancerID, p.CoverLetter, p.BidType, p.BidAmount, p.EstimatedDays,
		p.Status, p.StatusReason, p.CreatedAt, p.UpdatedAt, p.ConnectsSpent).Scan(&id)
	if err != nil {
		if isUniqueViolation(err, "uq_proposals_active_per_job_freelancer") {
			return 0, fmt.Errorf("active proposal already exists for this job: %w", ErrConflict)
		}
		return 0, err
	}

	if err := r.replaceAttachmentsTx(ctx, tx, id, p.Attachments); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *ProposalRepo) GetByID(ctx context.Context, proposalID int64) (domain.Proposal, error) {
	return r.getByWhere(ctx, `where id = $1`, proposalID)
}

func (r *ProposalRepo) GetByIDForFreelancer(ctx context.Context, proposalID int64, freelancerID uuid.UUID) (domain.Proposal, error) {
	return r.getByWhere(ctx, `where id = $1 and freelancer_id = $2`, proposalID, freelancerID)
}

func (r *ProposalRepo) GetLatestByJobForFreelancer(ctx context.Context, jobID int64, freelancerID uuid.UUID) (domain.Proposal, error) {
	return r.getByWhere(ctx, `where job_id = $1 and freelancer_id = $2 order by created_at desc limit 1`, jobID, freelancerID)
}

func (r *ProposalRepo) GetByIDForClient(ctx context.Context, proposalID int64, clientID uuid.UUID) (domain.Proposal, error) {
	return r.getByWhere(ctx, `where id = $1 and client_id = $2`, proposalID, clientID)
}

func (r *ProposalRepo) HasActiveProposal(ctx context.Context, jobID int64, freelancerID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		select exists(
			select 1 from proposals
			where job_id = $1 and freelancer_id = $2 and status in ('sent','shortlisted')
		)
	`, jobID, freelancerID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *ProposalRepo) UpdateEditable(ctx context.Context, proposalID int64, freelancerID uuid.UUID, coverLetter string, bidAmount float64, estimatedDays int32, attachments []domain.Attachment, updatedAt time.Time) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	res, err := tx.Exec(ctx, `
		update proposals
		set cover_letter = $3,
			bid_amount = $4,
			estimated_days = $5,
			updated_at = $6
		where id = $1 and freelancer_id = $2
	`, proposalID, freelancerID, coverLetter, bidAmount, estimatedDays, updatedAt)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}

	if err := r.replaceAttachmentsTx(ctx, tx, proposalID, attachments); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *ProposalRepo) Withdraw(ctx context.Context, proposalID int64, freelancerID uuid.UUID, reason string, at time.Time) error {
	res, err := r.pool.Exec(ctx, `
		update proposals
		set status = 'withdrawn',
			status_reason = $3,
			withdrawn_at = $4,
			updated_at = $4
		where id = $1 and freelancer_id = $2
	`, proposalID, freelancerID, reason, at)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ProposalRepo) SetStatus(ctx context.Context, proposalID int64, clientID uuid.UUID, status string, reason string, at time.Time) error {
	status = strings.ToLower(strings.TrimSpace(status))
	res, err := r.pool.Exec(ctx, `
		update proposals
		set status = $3,
			status_reason = $4,
			shortlisted_at = case when $3 = 'shortlisted' then $5 else shortlisted_at end,
			rejected_at = case when $3 = 'rejected' then $5 else rejected_at end,
			hired_at = case when $3 = 'hired' then $5 else hired_at end,
			updated_at = $5
		where id = $1 and client_id = $2
	`, proposalID, clientID, status, reason, at)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ProposalRepo) HasHiredProposalForJob(ctx context.Context, jobID int64) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		select exists(
			select 1 from proposals
			where job_id = $1 and status = 'hired'
		)
	`, jobID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *ProposalRepo) HireWithRequestID(ctx context.Context, proposalID int64, clientID uuid.UUID, requestID string, reason string, at time.Time) (domain.Proposal, bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Proposal{}, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var existingProposalID int64
	err = tx.QueryRow(ctx, `
		select proposal_id
		from proposal_hire_requests
		where client_id = $1 and proposal_id = $2 and request_id = $3
	`, clientID, proposalID, requestID).Scan(&existingProposalID)
	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			return domain.Proposal{}, false, err
		}
		p, getErr := r.GetByIDForClient(ctx, existingProposalID, clientID)
		if getErr != nil {
			return domain.Proposal{}, false, getErr
		}
		return p, true, nil
	}
	if !isNoRows(err) {
		return domain.Proposal{}, false, err
	}

	res, err := tx.Exec(ctx, `
		update proposals
		set status = 'hired',
			status_reason = $3,
			hired_at = $4,
			updated_at = $4
		where id = $1 and client_id = $2 and status in ('sent', 'shortlisted')
	`, proposalID, clientID, reason, at)
	if err != nil {
		if isUniqueViolation(err, "uq_proposals_single_hired_per_job") {
			return domain.Proposal{}, false, fmt.Errorf("job already has a hired proposal: %w", ErrConflict)
		}
		return domain.Proposal{}, false, err
	}
	if res.RowsAffected() == 0 {
		return domain.Proposal{}, false, fmt.Errorf("proposal cannot be hired in current status: %w", ErrConflict)
	}

	_, err = tx.Exec(ctx, `
		insert into proposal_hire_requests (client_id, proposal_id, request_id, created_at)
		values ($1, $2, $3, $4)
	`, clientID, proposalID, requestID, at)
	if err != nil {
		if isUniqueViolation(err, "proposal_hire_requests_client_id_proposal_id_request_id_key") {
			if err := tx.Commit(ctx); err != nil {
				return domain.Proposal{}, false, err
			}
			p, getErr := r.GetByIDForClient(ctx, proposalID, clientID)
			if getErr != nil {
				return domain.Proposal{}, false, getErr
			}
			return p, true, nil
		}
		return domain.Proposal{}, false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Proposal{}, false, err
	}

	p, err := r.GetByIDForClient(ctx, proposalID, clientID)
	if err != nil {
		return domain.Proposal{}, false, err
	}
	return p, false, nil
}

func (r *ProposalRepo) ListByJob(ctx context.Context, filter application.ListByJobFilter, limit, offset int) ([]domain.Proposal, error) {
	order := mapSortBy(filter.SortBy)
	args := []any{filter.ClientID, filter.JobID}
	idx := 3
	query := `
		select id, job_id, client_id, freelancer_id, cover_letter, bid_type, bid_amount, estimated_days,
			status, status_reason, created_at, updated_at, shortlisted_at, rejected_at, hired_at, withdrawn_at, connects_spent
		from proposals
		where client_id = $1 and job_id = $2
	`

	if len(filter.Statuses) > 0 {
		query += fmt.Sprintf(` and status = any($%d)`, idx)
		args = append(args, filter.Statuses)
		idx++
	}
	if filter.FreelancerID != nil {
		query += fmt.Sprintf(` and freelancer_id = $%d`, idx)
		args = append(args, *filter.FreelancerID)
		idx++
	}
	query += fmt.Sprintf(` order by %s limit $%d offset $%d`, order, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.Proposal, 0)
	for rows.Next() {
		p, err := scanProposal(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	if err := r.loadAttachments(ctx, items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ProposalRepo) ListByFreelancer(ctx context.Context, filter application.ListByFreelancerFilter, limit, offset int) ([]domain.Proposal, error) {
	order := mapSortBy(filter.SortBy)
	args := []any{filter.FreelancerID}
	idx := 2
	query := `
		select id, job_id, client_id, freelancer_id, cover_letter, bid_type, bid_amount, estimated_days,
			status, status_reason, created_at, updated_at, shortlisted_at, rejected_at, hired_at, withdrawn_at, connects_spent
		from proposals
		where freelancer_id = $1
	`

	if len(filter.Statuses) > 0 {
		query += fmt.Sprintf(` and status = any($%d)`, idx)
		args = append(args, filter.Statuses)
		idx++
	}
	if filter.JobID != nil {
		query += fmt.Sprintf(` and job_id = $%d`, idx)
		args = append(args, *filter.JobID)
		idx++
	}
	query += fmt.Sprintf(` order by %s limit $%d offset $%d`, order, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.Proposal, 0)
	for rows.Next() {
		p, err := scanProposal(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	if err := r.loadAttachments(ctx, items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ProposalRepo) getByWhere(ctx context.Context, where string, args ...any) (domain.Proposal, error) {
	query := fmt.Sprintf(`
		select id, job_id, client_id, freelancer_id, cover_letter, bid_type, bid_amount, estimated_days,
			status, status_reason, created_at, updated_at, shortlisted_at, rejected_at, hired_at, withdrawn_at, connects_spent
		from proposals
		%s
	`, where)

	row := r.pool.QueryRow(ctx, query, args...)
	p, err := scanProposal(row)
	if err != nil {
		if isNoRows(err) {
			return domain.Proposal{}, ErrNotFound
		}
		return domain.Proposal{}, err
	}
	attachments, err := r.getAttachmentsByProposalID(ctx, p.ID)
	if err != nil {
		return domain.Proposal{}, err
	}
	p.Attachments = attachments
	return p, nil
}

func (r *ProposalRepo) replaceAttachmentsTx(ctx context.Context, tx pgx.Tx, proposalID int64, attachments []domain.Attachment) error {
	if _, err := tx.Exec(ctx, `delete from proposal_attachments where proposal_id = $1`, proposalID); err != nil {
		return err
	}
	for _, a := range attachments {
		if _, err := tx.Exec(ctx, `
			insert into proposal_attachments (proposal_id, file_name, content_type, url, size_bytes)
			values ($1,$2,$3,$4,$5)
		`, proposalID, a.FileName, a.ContentType, a.URL, a.SizeBytes); err != nil {
			return err
		}
	}
	return nil
}

func (r *ProposalRepo) loadAttachments(ctx context.Context, items []domain.Proposal) error {
	if len(items) == 0 {
		return nil
	}
	for i := range items {
		attachments, err := r.getAttachmentsByProposalID(ctx, items[i].ID)
		if err != nil {
			return err
		}
		items[i].Attachments = attachments
	}
	return nil
}

func (r *ProposalRepo) getAttachmentsByProposalID(ctx context.Context, proposalID int64) ([]domain.Attachment, error) {
	rows, err := r.pool.Query(ctx, `
		select id, file_name, content_type, url, size_bytes
		from proposal_attachments
		where proposal_id = $1
		order by id asc
	`, proposalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attachments := make([]domain.Attachment, 0)
	for rows.Next() {
		var a domain.Attachment
		if err := rows.Scan(&a.ID, &a.FileName, &a.ContentType, &a.URL, &a.SizeBytes); err != nil {
			return nil, err
		}
		attachments = append(attachments, a)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return attachments, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanProposal(scanner rowScanner) (domain.Proposal, error) {
	var p domain.Proposal
	err := scanner.Scan(
		&p.ID,
		&p.JobID,
		&p.ClientID,
		&p.FreelancerID,
		&p.CoverLetter,
		&p.BidType,
		&p.BidAmount,
		&p.EstimatedDays,
		&p.Status,
		&p.StatusReason,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.ShortlistedAt,
		&p.RejectedAt,
		&p.HiredAt,
		&p.WithdrawnAt,
		&p.ConnectsSpent,
	)
	if err != nil {
		return domain.Proposal{}, err
	}
	return p, nil
}

func mapSortBy(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case domain.SortOldest:
		return "created_at asc"
	case domain.SortBidHigh:
		return "bid_amount desc, created_at desc"
	case domain.SortBidLow:
		return "bid_amount asc, created_at desc"
	case domain.SortNewest, "":
		fallthrough
	default:
		return "created_at desc"
	}
}
