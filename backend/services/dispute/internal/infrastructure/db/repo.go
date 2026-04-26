package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"jobconnect/dispute/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) Create(ctx context.Context, d domain.Dispute) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		insert into disputes (
			reference_type, reference_id, opened_by, reason, status, decision, resolution_note, created_at
		) values ($1,$2,$3,$4,$5,$6,$7,$8)
		returning id
	`,
		strings.TrimSpace(d.ReferenceType),
		strings.TrimSpace(d.ReferenceID),
		d.OpenedBy,
		strings.TrimSpace(d.Reason),
		strings.TrimSpace(d.Status),
		strings.TrimSpace(d.Decision),
		strings.TrimSpace(d.ResolutionNote),
		d.CreatedAt,
	).Scan(&id)
	return id, err
}

func (r *Repo) GetByID(ctx context.Context, disputeID int64) (domain.Dispute, error) {
	row := r.pool.QueryRow(ctx, `
		select id, reference_type, reference_id, opened_by, reason, status, decision,
		       resolution_note, resolved_by, created_at, resolved_at
		from disputes where id = $1
	`, disputeID)
	item, err := scanDispute(row)
	if err != nil {
		return domain.Dispute{}, err
	}
	return item, nil
}

func (r *Repo) List(ctx context.Context, referenceType, referenceID, status string, limit, offset int) ([]domain.Dispute, error) {
	rows, err := r.pool.Query(ctx, `
		select id, reference_type, reference_id, opened_by, reason, status, decision,
		       resolution_note, resolved_by, created_at, resolved_at
		from disputes
		where ($1 = '' or reference_type = $1)
		  and ($2 = '' or reference_id = $2)
		  and ($3 = '' or status = $3)
		order by created_at desc
		limit $4 offset $5
	`, strings.TrimSpace(referenceType), strings.TrimSpace(referenceID), strings.TrimSpace(status), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.Dispute, 0)
	for rows.Next() {
		item, scanErr := scanDispute(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return items, nil
}

func (r *Repo) Resolve(ctx context.Context, disputeID int64, decision, note string, resolvedBy uuid.UUID, at time.Time) error {
	res, err := r.pool.Exec(ctx, `
		update disputes
		set status = $2,
		    decision = $3,
		    resolution_note = $4,
		    resolved_by = $5,
		    resolved_at = $6
		where id = $1 and status = 'open'
	`, disputeID, domain.StatusResolved, strings.TrimSpace(decision), strings.TrimSpace(note), resolvedBy, at)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("dispute not found or already resolved")
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanDispute(s scanner) (domain.Dispute, error) {
	var (
		item          domain.Dispute
		resolvedByRaw *string
	)
	if err := s.Scan(
		&item.ID,
		&item.ReferenceType,
		&item.ReferenceID,
		&item.OpenedBy,
		&item.Reason,
		&item.Status,
		&item.Decision,
		&item.ResolutionNote,
		&resolvedByRaw,
		&item.CreatedAt,
		&item.ResolvedAt,
	); err != nil {
		return domain.Dispute{}, err
	}
	if resolvedByRaw != nil && strings.TrimSpace(*resolvedByRaw) != "" {
		id, err := uuid.Parse(strings.TrimSpace(*resolvedByRaw))
		if err != nil {
			return domain.Dispute{}, err
		}
		item.ResolvedBy = &id
	}
	return item, nil
}
