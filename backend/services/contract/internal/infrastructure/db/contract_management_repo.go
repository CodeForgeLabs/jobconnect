package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"jobconnect/contract/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
			where id = $1 and client_id = $2 and status = 'active'
		`, contractID, clientID, status, at)
		err = execErr
		resTag = res.RowsAffected()
	case domain.StatusActive:
		res, execErr := r.pool.Exec(ctx, `
			update contracts
			set status = $3, updated_at = $4, paused_at = null
			where id = $1 and client_id = $2 and status = 'paused'
		`, contractID, clientID, status, at)
		err = execErr
		resTag = res.RowsAffected()
	case domain.StatusEnded:
		res, execErr := r.pool.Exec(ctx, `
			update contracts
			set status = $3, updated_at = $4, ended_at = $4
			where id = $1 and client_id = $2 and status in ('active', 'paused')
		`, contractID, clientID, status, at)
		err = execErr
		resTag = res.RowsAffected()
	case domain.StatusRevoked:
		res, execErr := r.pool.Exec(ctx, `
			update contracts
			set status = $3, updated_at = $4, revoked_at = $4
			where id = $1 and client_id = $2 and status = 'pending_acceptance'
		`, contractID, clientID, status, at)
		err = execErr
		resTag = res.RowsAffected()
	case domain.StatusPendingAcceptance:
		res, execErr := r.pool.Exec(ctx, `
			update contracts
			set status = $3, updated_at = $4, activated_at = null, declined_at = null, revoked_at = null
			where id = $1 and client_id = $2 and status in ('active', 'declined', 'revoked', 'pending_acceptance')
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

func (r *ContractRepo) ReplaceMilestones(ctx context.Context, contractID int64, milestones []domain.Milestone, at time.Time) error {
	milestonesRaw, err := json.Marshal(milestones)
	if err != nil {
		return err
	}
	res, err := r.pool.Exec(ctx, `
		update contracts
		set milestones = $2, updated_at = $3
		where id = $1
	`, contractID, milestonesRaw, at)
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
			note, evidence_urls, status, review_note, created_at, invoice_id
		)
		select $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12
		where exists (
			select 1 from contracts c
			where c.id = $1
			  and c.freelancer_id = $2
			  and c.contract_type = 'hourly'
			  and c.status = 'active'
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
		log.EvidenceURLs,
		log.Status,
		log.ReviewNote,
		log.CreatedAt,
		nullableInt64(log.InvoiceID),
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
			l.duration_min, l.note, l.evidence_urls, l.status, l.review_note, l.created_at, l.client_review_at, l.invoice_id
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
		item, err := scanHourlyLog(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *ContractRepo) ListHourlyLogsForActorInRange(ctx context.Context, contractID int64, actorID uuid.UUID, startAt time.Time, endAt time.Time) ([]domain.HourlyLog, error) {
	rows, err := r.pool.Query(ctx, `
		select l.id, l.contract_id, l.freelancer_id, l.work_date, l.start_at, l.end_at,
			l.duration_min, l.note, l.evidence_urls, l.status, l.review_note, l.created_at, l.client_review_at, l.invoice_id
		from contract_hourly_logs l
		where l.contract_id = $1
		  and l.start_at < $4
		  and l.end_at > $3
		  and exists (
			select 1 from contracts c
			where c.id = l.contract_id and (c.client_id = $2 or c.freelancer_id = $2)
		  )
		order by l.start_at asc
	`, contractID, actorID, startAt, endAt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.HourlyLog, 0)
	for rows.Next() {
		item, err := scanHourlyLog(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *ContractRepo) UpdateHourlyLogForFreelancer(ctx context.Context, log domain.HourlyLog) error {
	res, err := r.pool.Exec(ctx, `
		update contract_hourly_logs l
		set work_date = $3, start_at = $4, end_at = $5, duration_min = $6, note = $7, evidence_urls = $8
		where l.id = $1
		  and l.freelancer_id = $2
		  and l.status = 'pending'
		  and l.invoice_id is null
		  and exists (
			select 1 from contracts c
			where c.id = l.contract_id
			  and c.contract_type = 'hourly'
			  and c.status = 'active'
		  )
	`, log.ID, log.FreelancerID, log.WorkDate, log.StartAt, log.EndAt, log.DurationMin, log.Note, log.EvidenceURLs)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ContractRepo) DeleteHourlyLogForFreelancer(ctx context.Context, hourlyLogID int64, freelancerID uuid.UUID) error {
	res, err := r.pool.Exec(ctx, `
		delete from contract_hourly_logs l
		where l.id = $1
		  and l.freelancer_id = $2
		  and l.status = 'pending'
		  and l.invoice_id is null
		  and exists (
			select 1 from contracts c
			where c.id = l.contract_id
			  and c.contract_type = 'hourly'
			  and c.status = 'active'
		  )
	`, hourlyLogID, freelancerID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ContractRepo) ReviewHourlyLogForClient(ctx context.Context, hourlyLogID int64, clientID uuid.UUID, status string, note string, at time.Time) error {
	res, err := r.pool.Exec(ctx, `
		update contract_hourly_logs l
		set status = $3, review_note = $4, client_review_at = $5
		from contracts c
		where l.id = $1
		  and c.id = l.contract_id
		  and c.client_id = $2
		  and c.contract_type = 'hourly'
		  and l.status = 'pending'
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
	row := r.pool.QueryRow(ctx, `
		select l.id, l.contract_id, l.freelancer_id, l.work_date, l.start_at, l.end_at,
			l.duration_min, l.note, l.evidence_urls, l.status, l.review_note, l.created_at, l.client_review_at, l.invoice_id
		from contract_hourly_logs l
		join contracts c on c.id = l.contract_id
		where l.id = $1 and (c.client_id = $2 or c.freelancer_id = $2)
	`, hourlyLogID, actorID)
	item, err := scanHourlyLog(row)
	if isNoRows(err) {
		return domain.HourlyLog{}, ErrNotFound
	}
	if err != nil {
		return domain.HourlyLog{}, err
	}
	return item, nil
}

func scanHourlyLog(scanner rowScanner) (domain.HourlyLog, error) {
	var item domain.HourlyLog
	var invoiceID *int64
	if err := scanner.Scan(
		&item.ID,
		&item.ContractID,
		&item.FreelancerID,
		&item.WorkDate,
		&item.StartAt,
		&item.EndAt,
		&item.DurationMin,
		&item.Note,
		&item.EvidenceURLs,
		&item.Status,
		&item.ReviewNote,
		&item.CreatedAt,
		&item.ClientReviewAt,
		&invoiceID,
	); err != nil {
		return domain.HourlyLog{}, err
	}
	if invoiceID != nil {
		item.InvoiceID = *invoiceID
	}
	return item, nil
}

func (r *ContractRepo) CreateHourlyInvoice(ctx context.Context, invoice domain.HourlyInvoice) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var id int64
	created := false
	err = tx.QueryRow(ctx, `
		insert into contract_hourly_invoices (
			contract_id, client_id, freelancer_id, week_start, week_end, status,
			billable_minutes, hourly_rate, amount_minor, dispute_id, created_at, submitted_at
		)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		on conflict (contract_id, week_start) do nothing
		returning id
	`, invoice.ContractID, invoice.ClientID, invoice.FreelancerID, invoice.WeekStart, invoice.WeekEnd, invoice.Status, invoice.BillableMinutes, invoice.HourlyRate, invoice.AmountMinor, invoice.DisputeID, invoice.CreatedAt, invoice.SubmittedAt).Scan(&id)
	if isNoRows(err) {
		row := tx.QueryRow(ctx, `
			select id
			from contract_hourly_invoices
			where contract_id = $1 and week_start = $2
		`, invoice.ContractID, invoice.WeekStart)
		if getErr := row.Scan(&id); getErr != nil {
			if isNoRows(getErr) {
				return 0, ErrNotFound
			}
			return 0, getErr
		}
	} else if err != nil {
		return 0, err
	} else {
		created = true
	}

	if created {
		if _, err := tx.Exec(ctx, `
			update contract_hourly_logs
			set invoice_id = $4
			where contract_id = $1
			  and start_at >= $2
			  and end_at <= $3
			  and status in ('pending', 'approved')
			  and invoice_id is null
		`, invoice.ContractID, invoice.WeekStart, invoice.WeekEnd, id); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *ContractRepo) GetHourlyInvoiceForActor(ctx context.Context, invoiceID int64, actorID uuid.UUID) (domain.HourlyInvoice, error) {
	row := r.pool.QueryRow(ctx, hourlyInvoiceSelect+`
		where i.id = $1
		  and exists (
			select 1 from contracts c where c.id = i.contract_id and (c.client_id = $2 or c.freelancer_id = $2)
		  )
	`, invoiceID, actorID)
	item, err := scanHourlyInvoice(row)
	if isNoRows(err) {
		return domain.HourlyInvoice{}, ErrNotFound
	}
	return item, err
}

func (r *ContractRepo) GetHourlyInvoice(ctx context.Context, invoiceID int64) (domain.HourlyInvoice, error) {
	row := r.pool.QueryRow(ctx, hourlyInvoiceSelect+` where i.id = $1`, invoiceID)
	item, err := scanHourlyInvoice(row)
	if isNoRows(err) {
		return domain.HourlyInvoice{}, ErrNotFound
	}
	return item, err
}

func (r *ContractRepo) ListHourlyInvoicesForActor(ctx context.Context, contractID int64, actorID uuid.UUID, limit, offset int) ([]domain.HourlyInvoice, error) {
	rows, err := r.pool.Query(ctx, hourlyInvoiceSelect+`
		where i.contract_id = $1
		  and exists (
			select 1 from contracts c where c.id = i.contract_id and (c.client_id = $2 or c.freelancer_id = $2)
		  )
		order by i.week_start desc
		limit $3 offset $4
	`, contractID, actorID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.HourlyInvoice, 0)
	for rows.Next() {
		item, err := scanHourlyInvoice(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ContractRepo) GetHourlyInvoiceByContractWeek(ctx context.Context, contractID int64, weekStart time.Time) (domain.HourlyInvoice, error) {
	row := r.pool.QueryRow(ctx, hourlyInvoiceSelect+` where i.contract_id = $1 and i.week_start = $2`, contractID, weekStart)
	item, err := scanHourlyInvoice(row)
	if isNoRows(err) {
		return domain.HourlyInvoice{}, ErrNotFound
	}
	return item, err
}

func (r *ContractRepo) AttachHourlyLogsToInvoice(ctx context.Context, contractID int64, weekStart time.Time, weekEnd time.Time, invoiceID int64) error {
	_, err := r.pool.Exec(ctx, `
		update contract_hourly_logs
		set invoice_id = $4
		where contract_id = $1
		  and start_at >= $2
		  and end_at <= $3
		  and status in ('pending', 'approved')
		  and invoice_id is null
	`, contractID, weekStart, weekEnd, invoiceID)
	return err
}

func (r *ContractRepo) MarkHourlyInvoiceStatus(ctx context.Context, invoiceID int64, status string, disputeID string, at time.Time) error {
	res, err := r.pool.Exec(ctx, `
		update contract_hourly_invoices
		set status = $2,
		    dispute_id = case when $3 <> '' then $3 else dispute_id end,
		    approved_at = case when $2 = 'approved' then $4 else approved_at end,
		    paid_at = case when $2 = 'paid' then $4 else paid_at end,
		    failed_at = case when $2 = 'failed' then $4 else failed_at end
		where id = $1
	`, invoiceID, status, strings.TrimSpace(disputeID), at)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

const hourlyInvoiceSelect = `
	select i.id, i.contract_id, i.client_id, i.freelancer_id, i.week_start, i.week_end,
		i.status, i.billable_minutes, i.hourly_rate, i.amount_minor, i.dispute_id,
		i.created_at, i.submitted_at, i.approved_at, i.paid_at, i.failed_at
	from contract_hourly_invoices i
`

func scanHourlyInvoice(scanner rowScanner) (domain.HourlyInvoice, error) {
	var item domain.HourlyInvoice
	if err := scanner.Scan(
		&item.ID,
		&item.ContractID,
		&item.ClientID,
		&item.FreelancerID,
		&item.WeekStart,
		&item.WeekEnd,
		&item.Status,
		&item.BillableMinutes,
		&item.HourlyRate,
		&item.AmountMinor,
		&item.DisputeID,
		&item.CreatedAt,
		&item.SubmittedAt,
		&item.ApprovedAt,
		&item.PaidAt,
		&item.FailedAt,
	); err != nil {
		return domain.HourlyInvoice{}, err
	}
	return item, nil
}

func (r *ContractRepo) CreateContractBonus(ctx context.Context, bonus domain.ContractBonus) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		insert into contract_bonuses (
			contract_id, client_id, freelancer_id, amount_minor, payment_reference_id, note, status, created_at
		)
		select $1,$2,$3,$4,$5,$6,$7,$8
		where exists (
			select 1 from contracts c
			where c.id = $1 and c.client_id = $2 and c.freelancer_id = $3
			  and c.status in ('active', 'paused', 'ended')
		)
		returning id
	`, bonus.ContractID, bonus.ClientID, bonus.FreelancerID, bonus.AmountMinor, strings.TrimSpace(bonus.PaymentReferenceID), bonus.Note, bonus.Status, bonus.CreatedAt).Scan(&id)
	if isNoRows(err) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *ContractRepo) GetContractBonusForActor(ctx context.Context, bonusID int64, actorID uuid.UUID) (domain.ContractBonus, error) {
	row := r.pool.QueryRow(ctx, contractBonusSelect+`
		where b.id = $1
		  and exists (
			select 1 from contracts c where c.id = b.contract_id and (c.client_id = $2 or c.freelancer_id = $2)
		  )
	`, bonusID, actorID)
	item, err := scanContractBonus(row)
	if isNoRows(err) {
		return domain.ContractBonus{}, ErrNotFound
	}
	return item, err
}

func (r *ContractRepo) GetContractBonus(ctx context.Context, bonusID int64) (domain.ContractBonus, error) {
	row := r.pool.QueryRow(ctx, contractBonusSelect+` where b.id = $1`, bonusID)
	item, err := scanContractBonus(row)
	if isNoRows(err) {
		return domain.ContractBonus{}, ErrNotFound
	}
	return item, err
}

func (r *ContractRepo) ListContractBonusesForActor(ctx context.Context, contractID int64, actorID uuid.UUID, limit, offset int) ([]domain.ContractBonus, error) {
	rows, err := r.pool.Query(ctx, contractBonusSelect+`
		where b.contract_id = $1
		  and exists (
			select 1 from contracts c where c.id = b.contract_id and (c.client_id = $2 or c.freelancer_id = $2)
		  )
		order by b.created_at desc
		limit $3 offset $4
	`, contractID, actorID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.ContractBonus, 0)
	for rows.Next() {
		item, err := scanContractBonus(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ContractRepo) MarkContractBonusStatus(ctx context.Context, bonusID int64, status string, paymentReferenceID string, at time.Time) error {
	res, err := r.pool.Exec(ctx, `
		update contract_bonuses
		set status = $2,
		    payment_reference_id = case when $4 <> '' then $4 else payment_reference_id end,
		    paid_at = case when $2 = 'paid' then $3 else paid_at end,
		    failed_at = case when $2 = 'failed' then $3 else failed_at end
		where id = $1
	`, bonusID, status, at, strings.TrimSpace(paymentReferenceID))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

const contractBonusSelect = `
	select b.id, b.contract_id, b.client_id, b.freelancer_id, b.amount_minor, b.payment_reference_id,
		b.note, b.status, b.created_at, b.paid_at, b.failed_at
	from contract_bonuses b
`

func scanContractBonus(scanner rowScanner) (domain.ContractBonus, error) {
	var item domain.ContractBonus
	if err := scanner.Scan(
		&item.ID,
		&item.ContractID,
		&item.ClientID,
		&item.FreelancerID,
		&item.AmountMinor,
		&item.PaymentReferenceID,
		&item.Note,
		&item.Status,
		&item.CreatedAt,
		&item.PaidAt,
		&item.FailedAt,
	); err != nil {
		return domain.ContractBonus{}, err
	}
	return item, nil
}

func (r *ContractRepo) HasBlockingFinancialActivity(ctx context.Context, contractID int64) (bool, string, error) {
	var count int
	if err := r.pool.QueryRow(ctx, `
		select count(*)
		from contract_hourly_invoices
		where contract_id = $1 and status in ('submitted', 'in_review', 'disputed', 'charged', 'failed')
	`, contractID).Scan(&count); err != nil {
		return false, "", err
	}
	if count > 0 {
		return true, "contract has unresolved hourly invoices", nil
	}
	if err := r.pool.QueryRow(ctx, `
		select count(*)
		from contract_bonuses
		where contract_id = $1 and status in ('pending', 'failed')
	`, contractID).Scan(&count); err != nil {
		return false, "", err
	}
	if count > 0 {
		return true, "contract has unresolved bonuses", nil
	}
	return false, "", nil
}

func (r *ContractRepo) CreateAmendmentForActor(ctx context.Context, a domain.Amendment) (int64, error) {
	payloadRaw, err := marshalAmendmentPayload(a.Payload)
	if err != nil {
		return 0, err
	}
	var id int64
	err = r.pool.QueryRow(ctx, `
		insert into contract_amendments (
			contract_id, proposed_by, summary, payload_json, status, expires_at, created_at
		)
		select $1,$2,$3,$4::jsonb,$5,$6,$7
		where exists (
			select 1 from contracts c where c.id = $1 and (c.client_id = $2 or c.freelancer_id = $2)
		)
		returning id
	`, a.ContractID, a.ProposedBy, a.Summary, payloadRaw, a.Status, a.ExpiresAt, a.CreatedAt).Scan(&id)
	if isNoRows(err) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *ContractRepo) RespondAmendmentForActor(ctx context.Context, amendmentID int64, actorID uuid.UUID, status string, responseNote string, at time.Time) error {
	res, err := r.pool.Exec(ctx, `
		update contract_amendments a
		set status = $3, responded_at = $4, responded_by = $5, response_note = $6
		from contracts c
		where a.id = $1
		  and c.id = a.contract_id
		  and a.status = 'pending'
		  and (a.expires_at is null or a.expires_at > $4)
		  and (c.client_id = $2 or c.freelancer_id = $2)
	`, amendmentID, actorID, status, at, actorID, strings.TrimSpace(responseNote))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ContractRepo) RespondAmendmentAndApplyForActor(ctx context.Context, amendmentID int64, actorID uuid.UUID, responseNote string, at time.Time) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		contractID      int64
		proposedBy      uuid.UUID
		payloadRaw      []byte
		amendmentStatus string
		expiresAt       *time.Time
		contractType    string
		title           string
		description     string
		hourlyRate      float64
		fixedTotal      float64
		weeklyHourLimit int32
		milestonesRaw   []byte
	)
	if err := tx.QueryRow(ctx, `
		select a.contract_id, a.proposed_by, a.payload_json, a.status, a.expires_at,
		       c.contract_type, c.title, c.description, c.hourly_rate, c.fixed_total, c.weekly_hour_limit, c.milestones
		from contract_amendments a
		join contracts c on c.id = a.contract_id
		where a.id = $1 and (c.client_id = $2 or c.freelancer_id = $2)
		for update of a, c
	`, amendmentID, actorID).Scan(
		&contractID,
		&proposedBy,
		&payloadRaw,
		&amendmentStatus,
		&expiresAt,
		&contractType,
		&title,
		&description,
		&hourlyRate,
		&fixedTotal,
		&weeklyHourLimit,
		&milestonesRaw,
	); err != nil {
		if isNoRows(err) {
			return ErrNotFound
		}
		return err
	}
	if proposedBy == actorID {
		return fmt.Errorf("only the counterparty can respond to an amendment")
	}
	if amendmentStatus != domain.AmendmentStatusPending {
		return fmt.Errorf("can only respond to pending amendments")
	}
	if expiresAt != nil && !expiresAt.After(at) {
		if _, err := tx.Exec(ctx, `
			update contract_amendments
			set status = $2, responded_at = $3
			where id = $1 and status = $4
		`, amendmentID, domain.AmendmentStatusExpired, at, domain.AmendmentStatusPending); err != nil {
			return err
		}
		return fmt.Errorf("amendment has expired")
	}

	payload, err := unmarshalAmendmentPayload(payloadRaw)
	if err != nil {
		return err
	}
	if payload.ScopeChange != nil {
		if v := strings.TrimSpace(payload.ScopeChange.NewTitle); v != "" {
			title = v
		}
		if v := strings.TrimSpace(payload.ScopeChange.NewDescription); v != "" {
			description = v
		}
	}
	if payload.CompensationChange != nil {
		if payload.CompensationChange.NewHourlyRate > 0 {
			if _, err := domain.MoneyToMinorUnits(payload.CompensationChange.NewHourlyRate, "new_hourly_rate"); err != nil {
				return err
			}
			hourlyRate = payload.CompensationChange.NewHourlyRate
		}
		if payload.CompensationChange.NewFixedTotal > 0 {
			if _, err := domain.MoneyToMinorUnits(payload.CompensationChange.NewFixedTotal, "new_fixed_total"); err != nil {
				return err
			}
			if strings.TrimSpace(contractType) == domain.TypeFixed && payload.MilestonesChange == nil {
				return fmt.Errorf("fixed_total change requires milestones_change for fixed contracts")
			}
			fixedTotal = payload.CompensationChange.NewFixedTotal
		}
	}
	if payload.WeeklyLimitChange != nil && payload.WeeklyLimitChange.NewWeeklyHourLimit > 0 {
		weeklyHourLimit = payload.WeeklyLimitChange.NewWeeklyHourLimit
	}
	if payload.MilestonesChange != nil {
		nextMilestonesRaw, err := json.Marshal(payload.MilestonesChange.Milestones)
		if err != nil {
			return err
		}
		if strings.TrimSpace(contractType) == domain.TypeFixed {
			if err := validateFixedMilestoneTotal(fixedTotal, payload.MilestonesChange.Milestones); err != nil {
				return err
			}
		}
		milestonesRaw = nextMilestonesRaw
	}

	if _, err := tx.Exec(ctx, `
		update contracts
		set title = $3,
		    description = $4,
		    hourly_rate = $5,
		    fixed_total = $6,
		    weekly_hour_limit = $7,
		    milestones = $8,
		    updated_at = $9
		where id = $1 and contract_type = $2
	`, contractID, strings.TrimSpace(contractType), title, description, hourlyRate, fixedTotal, weeklyHourLimit, milestonesRaw, at); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		update contract_amendments
		set status = $2, responded_at = $3, responded_by = $4, response_note = $5
		where id = $1 and status = $6
	`, amendmentID, domain.AmendmentStatusAccepted, at, actorID, strings.TrimSpace(responseNote), domain.AmendmentStatusPending); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r *ContractRepo) GetAmendmentForActor(ctx context.Context, amendmentID int64, actorID uuid.UUID) (domain.Amendment, error) {
	var a domain.Amendment
	var payloadRaw []byte
	var respondedByRaw *string
	err := r.pool.QueryRow(ctx, `
		select a.id, a.contract_id, a.proposed_by, a.summary, a.payload_json, a.status,
			a.expires_at, a.responded_at, a.created_at, a.responded_by::text, a.response_note
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
		&respondedByRaw,
		&a.ResponseNote,
	)
	if isNoRows(err) {
		return domain.Amendment{}, ErrNotFound
	}
	if err != nil {
		return domain.Amendment{}, err
	}
	if respondedByRaw != nil && strings.TrimSpace(*respondedByRaw) != "" {
		id, parseErr := uuid.Parse(strings.TrimSpace(*respondedByRaw))
		if parseErr != nil {
			return domain.Amendment{}, parseErr
		}
		a.RespondedBy = &id
	}
	payload, err := unmarshalAmendmentPayload(payloadRaw)
	if err != nil {
		return domain.Amendment{}, err
	}
	a.Payload = payload
	return a, nil
}

func (r *ContractRepo) ListAmendmentsForActor(ctx context.Context, contractID int64, actorID uuid.UUID, limit, offset int) ([]domain.Amendment, error) {
	rows, err := r.pool.Query(ctx, `
		select a.id, a.contract_id, a.proposed_by, a.summary, a.payload_json, a.status,
			a.expires_at, a.responded_at, a.created_at, a.responded_by::text, a.response_note
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
		var respondedByRaw *string
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
			&respondedByRaw,
			&a.ResponseNote,
		); err != nil {
			return nil, err
		}
		if respondedByRaw != nil && strings.TrimSpace(*respondedByRaw) != "" {
			id, parseErr := uuid.Parse(strings.TrimSpace(*respondedByRaw))
			if parseErr != nil {
				return nil, parseErr
			}
			a.RespondedBy = &id
		}
		payload, err := unmarshalAmendmentPayload(payloadRaw)
		if err != nil {
			return nil, err
		}
		a.Payload = payload
		items = append(items, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ContractRepo) ExpirePendingAmendmentsForActor(ctx context.Context, contractID int64, actorID uuid.UUID, at time.Time) error {
	_, err := r.pool.Exec(ctx, `
		update contract_amendments a
		set status = $4, responded_at = $3
		where a.contract_id = $1
		  and a.status = $5
		  and a.expires_at is not null
		  and a.expires_at <= $3
		  and exists (
			select 1 from contracts c where c.id = a.contract_id and (c.client_id = $2 or c.freelancer_id = $2)
		  )
	`, contractID, actorID, at, domain.AmendmentStatusExpired, domain.AmendmentStatusPending)
	return err
}

func (r *ContractRepo) ExpireAmendmentForActor(ctx context.Context, amendmentID int64, actorID uuid.UUID, at time.Time) (bool, error) {
	res, err := r.pool.Exec(ctx, `
		update contract_amendments a
		set status = $4, responded_at = $3
		where a.id = $1
		  and a.status = $5
		  and a.expires_at is not null
		  and a.expires_at <= $3
		  and exists (
			select 1 from contracts c where c.id = a.contract_id and (c.client_id = $2 or c.freelancer_id = $2)
		  )
	`, amendmentID, actorID, at, domain.AmendmentStatusExpired, domain.AmendmentStatusPending)
	if err != nil {
		return false, err
	}
	return res.RowsAffected() > 0, nil
}

func marshalAmendmentPayload(payload domain.AmendmentPayload) ([]byte, error) {
	return json.Marshal(payload)
}

func unmarshalAmendmentPayload(raw []byte) (domain.AmendmentPayload, error) {
	if len(raw) == 0 {
		return domain.AmendmentPayload{}, nil
	}
	var out domain.AmendmentPayload
	if err := json.Unmarshal(raw, &out); err != nil {
		return domain.AmendmentPayload{}, err
	}
	return out, nil
}

func (r *ContractRepo) AppendStatusHistory(ctx context.Context, entry domain.StatusHistoryEntry) error {
	if strings.TrimSpace(entry.EventType) == "" {
		if err := domain.ValidateStatus(entry.Status); err == nil {
			entry.EventType = domain.StatusHistoryEventContractStatusChanged
		}
	}
	_, err := r.pool.Exec(ctx, `
		insert into contract_status_history (contract_id, status, reason, actor_id, created_at, event_type, milestone_id)
		values ($1,$2,$3,$4,$5,$6,$7)
	`, entry.ContractID, entry.Status, entry.Reason, entry.ActorID, entry.CreatedAt, entry.EventType, nullableInt64(entry.MilestoneID))
	return err
}

func (r *ContractRepo) ListStatusHistoryForActor(ctx context.Context, contractID int64, actorID uuid.UUID, limit, offset int) ([]domain.StatusHistoryEntry, error) {
	rows, err := r.pool.Query(ctx, `
		select h.id, h.contract_id, h.status, h.reason, h.actor_id, h.created_at, h.event_type, h.milestone_id
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
		var milestoneID *int64
		if err := rows.Scan(&h.ID, &h.ContractID, &h.Status, &h.Reason, &h.ActorID, &h.CreatedAt, &h.EventType, &milestoneID); err != nil {
			return nil, err
		}
		if milestoneID != nil {
			h.MilestoneID = *milestoneID
		}
		if strings.TrimSpace(h.EventType) == "" {
			if err := domain.ValidateStatus(h.Status); err == nil {
				h.EventType = domain.StatusHistoryEventContractStatusChanged
			}
		}
		items = append(items, h)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func nullableInt64(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}

func validateFixedMilestoneTotal(fixedTotal float64, milestones []domain.Milestone) error {
	if fixedTotal <= 0 {
		return fmt.Errorf("fixed_total must be greater than zero")
	}
	if len(milestones) == 0 {
		return fmt.Errorf("fixed contracts require at least one milestone")
	}
	var totalMinor int64
	for i, m := range milestones {
		amountMinor, err := domain.MoneyToMinorUnits(m.Amount, fmt.Sprintf("milestone amount at index %d", i))
		if err != nil {
			return err
		}
		if amountMinor <= 0 {
			return fmt.Errorf("milestone amount must be positive at index %d", i)
		}
		totalMinor += amountMinor
	}
	expectedMinor, err := domain.MoneyToMinorUnits(fixedTotal, "fixed_total")
	if err != nil {
		return err
	}
	if totalMinor != expectedMinor {
		return fmt.Errorf("milestone amounts must equal fixed_total")
	}
	return nil
}
