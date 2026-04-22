package db

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"jobconnect/contract/internal/domain"

	"github.com/google/uuid"
)

func (r *ContractRepo) SetStatusForClient(ctx context.Context, contractID int64, clientID uuid.UUID, status string, at time.Time) error {
	status = strings.ToLower(strings.TrimSpace(status))
	var resTag int64
	var err error
	switch status {
	case domain.StatusPaused:
		res, execErr := r.pool.Exec(ctx, `
			update contracts
			set status = $3, updated_at = $4, paused_at = $4
			where id = $1 and client_id = $2
		`, contractID, clientID, status, at)
		err = execErr
		resTag = res.RowsAffected()
	case domain.StatusActive:
		res, execErr := r.pool.Exec(ctx, `
			update contracts
			set status = $3, updated_at = $4, paused_at = null
			where id = $1 and client_id = $2
		`, contractID, clientID, status, at)
		err = execErr
		resTag = res.RowsAffected()
	case domain.StatusEnded:
		res, execErr := r.pool.Exec(ctx, `
			update contracts
			set status = $3, updated_at = $4, ended_at = $4
			where id = $1 and client_id = $2
		`, contractID, clientID, status, at)
		err = execErr
		resTag = res.RowsAffected()
	case domain.StatusRevoked:
		res, execErr := r.pool.Exec(ctx, `
			update contracts
			set status = $3, updated_at = $4, revoked_at = $4
			where id = $1 and client_id = $2
		`, contractID, clientID, status, at)
		err = execErr
		resTag = res.RowsAffected()
	case domain.StatusPendingAcceptance:
		res, execErr := r.pool.Exec(ctx, `
			update contracts
			set status = $3, updated_at = $4, activated_at = null, declined_at = null, revoked_at = null
			where id = $1 and client_id = $2
		`, contractID, clientID, status, at)
		err = execErr
		resTag = res.RowsAffected()
	default:
		res, execErr := r.pool.Exec(ctx, `
			update contracts
			set status = $3, updated_at = $4
			where id = $1 and client_id = $2
		`, contractID, clientID, status, at)
		err = execErr
		resTag = res.RowsAffected()
	}
	if err != nil {
		return err
	}
	if resTag == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ContractRepo) ReplaceMilestonesForActor(ctx context.Context, contractID int64, actorID uuid.UUID, milestones []domain.Milestone, at time.Time) error {
	milestonesRaw, err := json.Marshal(milestones)
	if err != nil {
		return err
	}
	res, err := r.pool.Exec(ctx, `
		update contracts
		set milestones = $3, updated_at = $4
		where id = $1 and (client_id = $2 or freelancer_id = $2)
	`, contractID, actorID, milestonesRaw, at)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ContractRepo) CreateHourlyLogForFreelancer(ctx context.Context, log domain.HourlyLog) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		insert into contract_hourly_logs (
			contract_id, freelancer_id, work_date, start_at, end_at, duration_min,
			note, status, review_note, created_at
		)
		select $1,$2,$3,$4,$5,$6,$7,$8,$9,$10
		where exists (
			select 1 from contracts c where c.id = $1 and c.freelancer_id = $2
		)
		returning id
	`,
		log.ContractID,
		log.FreelancerID,
		log.WorkDate,
		log.StartAt,
		log.EndAt,
		log.DurationMin,
		log.Note,
		log.Status,
		log.ReviewNote,
		log.CreatedAt,
	).Scan(&id)
	if isNoRows(err) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *ContractRepo) ListHourlyLogsForActor(ctx context.Context, contractID int64, actorID uuid.UUID, limit, offset int) ([]domain.HourlyLog, error) {
	rows, err := r.pool.Query(ctx, `
		select l.id, l.contract_id, l.freelancer_id, l.work_date, l.start_at, l.end_at,
			l.duration_min, l.note, l.status, l.review_note, l.created_at, l.client_review_at
		from contract_hourly_logs l
		where l.contract_id = $1
		  and exists (
			select 1 from contracts c
			where c.id = l.contract_id and (c.client_id = $2 or c.freelancer_id = $2)
		  )
		order by l.created_at desc
		limit $3 offset $4
	`, contractID, actorID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.HourlyLog, 0)
	for rows.Next() {
		var item domain.HourlyLog
		if err := rows.Scan(
			&item.ID,
			&item.ContractID,
			&item.FreelancerID,
			&item.WorkDate,
			&item.StartAt,
			&item.EndAt,
			&item.DurationMin,
			&item.Note,
			&item.Status,
			&item.ReviewNote,
			&item.CreatedAt,
			&item.ClientReviewAt,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *ContractRepo) ReviewHourlyLogForClient(ctx context.Context, hourlyLogID int64, clientID uuid.UUID, status string, note string, at time.Time) error {
	res, err := r.pool.Exec(ctx, `
		update contract_hourly_logs l
		set status = $3, review_note = $4, client_review_at = $5
		from contracts c
		where l.id = $1 and c.id = l.contract_id and c.client_id = $2
	`, hourlyLogID, clientID, status, note, at)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ContractRepo) GetHourlyLogForActor(ctx context.Context, hourlyLogID int64, actorID uuid.UUID) (domain.HourlyLog, error) {
	var item domain.HourlyLog
	err := r.pool.QueryRow(ctx, `
		select l.id, l.contract_id, l.freelancer_id, l.work_date, l.start_at, l.end_at,
			l.duration_min, l.note, l.status, l.review_note, l.created_at, l.client_review_at
		from contract_hourly_logs l
		join contracts c on c.id = l.contract_id
		where l.id = $1 and (c.client_id = $2 or c.freelancer_id = $2)
	`, hourlyLogID, actorID).Scan(
		&item.ID,
		&item.ContractID,
		&item.FreelancerID,
		&item.WorkDate,
		&item.StartAt,
		&item.EndAt,
		&item.DurationMin,
		&item.Note,
		&item.Status,
		&item.ReviewNote,
		&item.CreatedAt,
		&item.ClientReviewAt,
	)
	if isNoRows(err) {
		return domain.HourlyLog{}, ErrNotFound
	}
	if err != nil {
		return domain.HourlyLog{}, err
	}
	return item, nil
}

func (r *ContractRepo) CreateAmendmentForActor(ctx context.Context, a domain.Amendment) (int64, error) {
	payload := a.PayloadJSON
	if strings.TrimSpace(payload) == "" {
		payload = "{}"
	}
	var id int64
	err := r.pool.QueryRow(ctx, `
		insert into contract_amendments (
			contract_id, proposed_by, summary, payload_json, status, expires_at, created_at
		)
		select $1,$2,$3,$4::jsonb,$5,$6,$7
		where exists (
			select 1 from contracts c where c.id = $1 and (c.client_id = $2 or c.freelancer_id = $2)
		)
		returning id
	`, a.ContractID, a.ProposedBy, a.Summary, payload, a.Status, a.ExpiresAt, a.CreatedAt).Scan(&id)
	if isNoRows(err) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *ContractRepo) RespondAmendmentForActor(ctx context.Context, amendmentID int64, actorID uuid.UUID, status string, at time.Time) error {
	res, err := r.pool.Exec(ctx, `
		update contract_amendments a
		set status = $3, responded_at = $4
		from contracts c
		where a.id = $1 and c.id = a.contract_id and (c.client_id = $2 or c.freelancer_id = $2)
	`, amendmentID, actorID, status, at)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ContractRepo) GetAmendmentForActor(ctx context.Context, amendmentID int64, actorID uuid.UUID) (domain.Amendment, error) {
	var a domain.Amendment
	var payloadRaw []byte
	err := r.pool.QueryRow(ctx, `
		select a.id, a.contract_id, a.proposed_by, a.summary, a.payload_json, a.status,
			a.expires_at, a.responded_at, a.created_at
		from contract_amendments a
		join contracts c on c.id = a.contract_id
		where a.id = $1 and (c.client_id = $2 or c.freelancer_id = $2)
	`, amendmentID, actorID).Scan(
		&a.ID,
		&a.ContractID,
		&a.ProposedBy,
		&a.Summary,
		&payloadRaw,
		&a.Status,
		&a.ExpiresAt,
		&a.RespondedAt,
		&a.CreatedAt,
	)
	if isNoRows(err) {
		return domain.Amendment{}, ErrNotFound
	}
	if err != nil {
		return domain.Amendment{}, err
	}
	a.PayloadJSON = string(payloadRaw)
	return a, nil
}

func (r *ContractRepo) ListAmendmentsForActor(ctx context.Context, contractID int64, actorID uuid.UUID, limit, offset int) ([]domain.Amendment, error) {
	rows, err := r.pool.Query(ctx, `
		select a.id, a.contract_id, a.proposed_by, a.summary, a.payload_json, a.status,
			a.expires_at, a.responded_at, a.created_at
		from contract_amendments a
		where a.contract_id = $1
		  and exists (
			select 1 from contracts c where c.id = a.contract_id and (c.client_id = $2 or c.freelancer_id = $2)
		  )
		order by a.created_at desc
		limit $3 offset $4
	`, contractID, actorID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.Amendment, 0)
	for rows.Next() {
		var a domain.Amendment
		var payloadRaw []byte
		if err := rows.Scan(
			&a.ID,
			&a.ContractID,
			&a.ProposedBy,
			&a.Summary,
			&payloadRaw,
			&a.Status,
			&a.ExpiresAt,
			&a.RespondedAt,
			&a.CreatedAt,
		); err != nil {
			return nil, err
		}
		a.PayloadJSON = string(payloadRaw)
		items = append(items, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ContractRepo) AppendStatusHistory(ctx context.Context, entry domain.StatusHistoryEntry) error {
	_, err := r.pool.Exec(ctx, `
		insert into contract_status_history (contract_id, status, reason, actor_id, created_at)
		values ($1,$2,$3,$4,$5)
	`, entry.ContractID, entry.Status, entry.Reason, entry.ActorID, entry.CreatedAt)
	return err
}

func (r *ContractRepo) ListStatusHistoryForActor(ctx context.Context, contractID int64, actorID uuid.UUID, limit, offset int) ([]domain.StatusHistoryEntry, error) {
	rows, err := r.pool.Query(ctx, `
		select h.id, h.contract_id, h.status, h.reason, h.actor_id, h.created_at
		from contract_status_history h
		where h.contract_id = $1
		  and exists (
			select 1 from contracts c where c.id = h.contract_id and (c.client_id = $2 or c.freelancer_id = $2)
		  )
		order by h.created_at desc
		limit $3 offset $4
	`, contractID, actorID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.StatusHistoryEntry, 0)
	for rows.Next() {
		var h domain.StatusHistoryEntry
		if err := rows.Scan(&h.ID, &h.ContractID, &h.Status, &h.Reason, &h.ActorID, &h.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, h)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
