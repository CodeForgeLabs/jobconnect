package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"jobconnect/job/internal/application"
	"jobconnect/job/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type JobRepo struct {
	pool *pgxpool.Pool
}

func NewJobRepo(pool *pgxpool.Pool) *JobRepo {
	return &JobRepo{pool: pool}
}

func (r *JobRepo) Create(ctx context.Context, job domain.Job) (int64, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	skills, err := json.Marshal(job.RequiredSkills)
	if err != nil {
		return 0, err
	}

	var id int64
	err = tx.QueryRow(ctx, `
		insert into jobs (
			client_id, title, description, required_skills, job_type, budget_fixed, hourly_rate,
			currency, budget_min, budget_max, visibility, experience_level, deadline, status, created_at, updated_at
		)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		returning id
	`, job.ClientID, job.Title, job.Description, skills, job.JobType, job.BudgetFixed, job.HourlyRate,
		job.Currency, job.BudgetMin, job.BudgetMax, job.Visibility, job.ExperienceLevel, job.Deadline, job.Status, job.CreatedAt, job.UpdatedAt).Scan(&id)
	if err != nil {
		return 0, err
	}

	for _, att := range job.Attachments {
		_, err = tx.Exec(ctx, `
			insert into job_attachments (job_id, file_name, content_type, storage_key, url, size_bytes)
			values ($1,$2,$3,$4,$5,$6)
		`, id, att.FileName, att.ContentType, att.StorageKey, att.URL, att.SizeBytes)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *JobRepo) GetByID(ctx context.Context, jobID int64) (domain.Job, error) {
	job, err := r.getJobByQuery(ctx, `where id = $1`, jobID)
	if err != nil {
		return domain.Job{}, err
	}
	return job, nil
}

func (r *JobRepo) GetByIDForClient(ctx context.Context, jobID int64, clientID uuid.UUID) (domain.Job, error) {
	job, err := r.getJobByQuery(ctx, `where id = $1 and client_id = $2`, jobID, clientID)
	if err != nil {
		return domain.Job{}, err
	}
	return job, nil
}

func (r *JobRepo) ListByClient(ctx context.Context, clientID uuid.UUID, status string, limit, offset int) ([]domain.Job, error) {
	query := `
		select id, client_id, title, description, required_skills, job_type,
			budget_fixed, hourly_rate, currency, budget_min, budget_max, visibility, experience_level,
			deadline, status, close_reason, created_at, updated_at, closed_at, paused_at, filled_at
		from jobs
		where client_id = $1
	`
	args := []any{clientID}
	if status != "" {
		query += ` and status = $2`
		args = append(args, status)
		query += ` order by created_at desc limit $3 offset $4`
		args = append(args, limit, offset)
	} else {
		query += ` order by created_at desc limit $2 offset $3`
		args = append(args, limit, offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := make([]domain.Job, 0)
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	if err := r.hydrateAttachments(ctx, jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}

func (r *JobRepo) ListOpen(ctx context.Context, limit, offset int) ([]domain.Job, error) {
	rows, err := r.pool.Query(ctx, `
		select id, client_id, title, description, required_skills, job_type,
			budget_fixed, hourly_rate, currency, budget_min, budget_max, visibility, experience_level,
			deadline, status, close_reason, created_at, updated_at, closed_at, paused_at, filled_at
		from jobs
		where status = $1
		order by created_at desc
		limit $2 offset $3
	`, domain.JobStatusOpen, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := make([]domain.Job, 0)
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	if err := r.hydrateAttachments(ctx, jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}

func (r *JobRepo) Close(ctx context.Context, jobID int64, clientID uuid.UUID, reason string, closedAt time.Time) error {
	res, err := r.pool.Exec(ctx, `
		update jobs
		set status = $3,
			close_reason = $4,
			closed_at = $5,
			updated_at = $5
		where id = $1 and client_id = $2 and status = $6
	`, jobID, clientID, domain.JobStatusClosed, reason, closedAt, domain.JobStatusOpen)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *JobRepo) AddAttachment(ctx context.Context, jobID int64, clientID uuid.UUID, attachment domain.Attachment) (domain.Attachment, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		insert into job_attachments (job_id, file_name, content_type, storage_key, url, size_bytes)
		select j.id, $3, $4, $5, $6, $7
		from jobs j
		where j.id = $1 and j.client_id = $2 and j.status = $8
		returning id
	`, jobID, clientID, attachment.FileName, attachment.ContentType, attachment.StorageKey, attachment.URL, attachment.SizeBytes, domain.JobStatusOpen).Scan(&id)
	if err != nil {
		if isNoRows(err) {
			return domain.Attachment{}, ErrNotFound
		}
		return domain.Attachment{}, err
	}

	attachment.ID = id
	return attachment, nil
}

func (r *JobRepo) DeleteAttachment(ctx context.Context, jobID int64, attachmentID int64, clientID uuid.UUID) (domain.Attachment, error) {
	var att domain.Attachment
	err := r.pool.QueryRow(ctx, `
		delete from job_attachments a
		using jobs j
		where a.job_id = j.id
		  and a.id = $1
		  and j.id = $2
		  and j.client_id = $3
		  and j.status = $4
		returning a.id, a.file_name, a.content_type, a.storage_key, a.url, a.size_bytes
	`, attachmentID, jobID, clientID, domain.JobStatusOpen).Scan(&att.ID, &att.FileName, &att.ContentType, &att.StorageKey, &att.URL, &att.SizeBytes)
	if err != nil {
		if isNoRows(err) {
			return domain.Attachment{}, ErrNotFound
		}
		return domain.Attachment{}, err
	}
	return att, nil
}

func (r *JobRepo) getJobByQuery(ctx context.Context, where string, args ...any) (domain.Job, error) {
	query := fmt.Sprintf(`
		select id, client_id, title, description, required_skills, job_type,
			budget_fixed, hourly_rate, currency, budget_min, budget_max, visibility, experience_level,
			deadline, status, close_reason, created_at, updated_at, closed_at, paused_at, filled_at
		from jobs
		%s
	`, where)

	row := r.pool.QueryRow(ctx, query, args...)
	job, err := scanJob(row)
	if err != nil {
		if isNoRows(err) {
			return domain.Job{}, ErrNotFound
		}
		return domain.Job{}, err
	}

	attRows, err := r.pool.Query(ctx, `
		select id, file_name, content_type, storage_key, url, size_bytes
		from job_attachments
		where job_id = $1
		order by id asc
	`, job.ID)
	if err != nil {
		return domain.Job{}, err
	}
	defer attRows.Close()

	attachments := make([]domain.Attachment, 0)
	for attRows.Next() {
		var a domain.Attachment
		if err := attRows.Scan(&a.ID, &a.FileName, &a.ContentType, &a.StorageKey, &a.URL, &a.SizeBytes); err != nil {
			return domain.Job{}, err
		}
		attachments = append(attachments, a)
	}
	if attRows.Err() != nil {
		return domain.Job{}, attRows.Err()
	}
	job.Attachments = attachments
	return job, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanJob(scanner rowScanner) (domain.Job, error) {
	var job domain.Job
	var skillsRaw []byte
	err := scanner.Scan(
		&job.ID,
		&job.ClientID,
		&job.Title,
		&job.Description,
		&skillsRaw,
		&job.JobType,
		&job.BudgetFixed,
		&job.HourlyRate,
		&job.Currency,
		&job.BudgetMin,
		&job.BudgetMax,
		&job.Visibility,
		&job.ExperienceLevel,
		&job.Deadline,
		&job.Status,
		&job.CloseReason,
		&job.CreatedAt,
		&job.UpdatedAt,
		&job.ClosedAt,
		&job.PausedAt,
		&job.FilledAt,
	)
	if err != nil {
		return domain.Job{}, err
	}
	if len(skillsRaw) > 0 {
		_ = json.Unmarshal(skillsRaw, &job.RequiredSkills)
	}
	return job, nil
}

func (r *JobRepo) Update(ctx context.Context, job domain.Job) (domain.Job, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Job{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	skills, err := json.Marshal(job.RequiredSkills)
	if err != nil {
		return domain.Job{}, err
	}

	res, err := tx.Exec(ctx, `
		update jobs set
			title = $2, description = $3, required_skills = $4, job_type = $5,
			budget_fixed = $6, hourly_rate = $7, currency = $8, budget_min = $9, budget_max = $10,
			visibility = $11, experience_level = $12, deadline = $13, updated_at = $14
		where id = $1 and status = $15
	`, job.ID, job.Title, job.Description, skills, job.JobType,
		job.BudgetFixed, job.HourlyRate, job.Currency, job.BudgetMin, job.BudgetMax,
		job.Visibility, job.ExperienceLevel, job.Deadline, job.UpdatedAt,
		domain.JobStatusOpen)
	if err != nil {
		return domain.Job{}, err
	}
	if res.RowsAffected() == 0 {
		return domain.Job{}, ErrNotFound
	}

	// Replace attachments: delete old, insert new.
	_, err = tx.Exec(ctx, `delete from job_attachments where job_id = $1`, job.ID)
	if err != nil {
		return domain.Job{}, err
	}
	for _, att := range job.Attachments {
		_, err = tx.Exec(ctx, `
			insert into job_attachments (job_id, file_name, content_type, storage_key, url, size_bytes)
			values ($1,$2,$3,$4,$5,$6)
		`, job.ID, att.FileName, att.ContentType, att.StorageKey, att.URL, att.SizeBytes)
		if err != nil {
			return domain.Job{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Job{}, err
	}

	// Re-fetch to return the full persisted state.
	return r.GetByID(ctx, job.ID)
}

func (r *JobRepo) ListOpenFiltered(ctx context.Context, filter application.ListOpenFilter) ([]domain.Job, error) {
	where := []string{"status = 'open'", "visibility = 'public'"}
	args := []any{}
	argIdx := 1

	if filter.SearchQuery != "" {
		where = append(where, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+filter.SearchQuery+"%")
		argIdx++
	}
	if len(filter.Skills) > 0 {
		// JSONB containment: required_skills @> '["Go","Flutter"]'::jsonb
		skillsJSON, _ := json.Marshal(filter.Skills)
		where = append(where, fmt.Sprintf("required_skills @> $%d::jsonb", argIdx))
		args = append(args, string(skillsJSON))
		argIdx++
	}
	if filter.JobType != "" {
		where = append(where, fmt.Sprintf("job_type = $%d", argIdx))
		args = append(args, filter.JobType)
		argIdx++
	}
	if filter.Visibility != "" {
		where = append(where, fmt.Sprintf("visibility = $%d", argIdx))
		args = append(args, filter.Visibility)
		argIdx++
	}
	if filter.ExperienceLevel != "" {
		where = append(where, fmt.Sprintf("experience_level = $%d", argIdx))
		args = append(args, filter.ExperienceLevel)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")
	query := fmt.Sprintf(`
		select id, client_id, title, description, required_skills, job_type,
			budget_fixed, hourly_rate, currency, budget_min, budget_max, visibility, experience_level,
			deadline, status, close_reason, created_at, updated_at, closed_at, paused_at, filled_at
		from jobs
		where %s
		order by created_at desc
		limit $%d offset $%d
	`, whereClause, argIdx, argIdx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := make([]domain.Job, 0)
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	if err := r.hydrateAttachments(ctx, jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}

func (r *JobRepo) CountOpenFiltered(ctx context.Context, filter application.ListOpenFilter) (int64, error) {
	where := []string{"status = 'open'", "visibility = 'public'"}
	args := []any{}
	argIdx := 1

	if filter.SearchQuery != "" {
		where = append(where, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+filter.SearchQuery+"%")
		argIdx++
	}
	if len(filter.Skills) > 0 {
		skillsJSON, _ := json.Marshal(filter.Skills)
		where = append(where, fmt.Sprintf("required_skills @> $%d::jsonb", argIdx))
		args = append(args, string(skillsJSON))
		argIdx++
	}
	if filter.JobType != "" {
		where = append(where, fmt.Sprintf("job_type = $%d", argIdx))
		args = append(args, filter.JobType)
		argIdx++
	}
	if filter.Visibility != "" {
		where = append(where, fmt.Sprintf("visibility = $%d", argIdx))
		args = append(args, filter.Visibility)
		argIdx++
	}
	if filter.ExperienceLevel != "" {
		where = append(where, fmt.Sprintf("experience_level = $%d", argIdx))
		args = append(args, filter.ExperienceLevel)
		argIdx++
	}

	query := fmt.Sprintf(`select count(1) from jobs where %s`, strings.Join(where, " AND "))
	var total int64
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *JobRepo) SetVisibility(ctx context.Context, jobID int64, clientID uuid.UUID, visibility string, updatedAt time.Time) (domain.Job, error) {
	res, err := r.pool.Exec(ctx, `
		update jobs
		set visibility = $3, updated_at = $4
		where id = $1 and client_id = $2 and status in ('open','paused')
	`, jobID, clientID, visibility, updatedAt)
	if err != nil {
		return domain.Job{}, err
	}
	if res.RowsAffected() == 0 {
		return domain.Job{}, ErrNotFound
	}
	return r.GetByIDForClient(ctx, jobID, clientID)
}

func (r *JobRepo) SetBudgetRange(ctx context.Context, jobID int64, clientID uuid.UUID, budgetMin, budgetMax float64, updatedAt time.Time) (domain.Job, error) {
	res, err := r.pool.Exec(ctx, `
		update jobs
		set budget_min = $3, budget_max = $4, updated_at = $5
		where id = $1 and client_id = $2 and status in ('open','paused')
	`, jobID, clientID, budgetMin, budgetMax, updatedAt)
	if err != nil {
		return domain.Job{}, err
	}
	if res.RowsAffected() == 0 {
		return domain.Job{}, ErrNotFound
	}
	return r.GetByIDForClient(ctx, jobID, clientID)
}

func (r *JobRepo) SetExperienceLevel(ctx context.Context, jobID int64, clientID uuid.UUID, experienceLevel string, updatedAt time.Time) (domain.Job, error) {
	res, err := r.pool.Exec(ctx, `
		update jobs
		set experience_level = $3, updated_at = $4
		where id = $1 and client_id = $2 and status in ('open','paused')
	`, jobID, clientID, experienceLevel, updatedAt)
	if err != nil {
		return domain.Job{}, err
	}
	if res.RowsAffected() == 0 {
		return domain.Job{}, ErrNotFound
	}
	return r.GetByIDForClient(ctx, jobID, clientID)
}

func (r *JobRepo) Pause(ctx context.Context, jobID int64, clientID uuid.UUID, updatedAt time.Time) (domain.Job, error) {
	res, err := r.pool.Exec(ctx, `
		update jobs
		set status = $3, paused_at = $4, updated_at = $4
		where id = $1 and client_id = $2 and status = $5
	`, jobID, clientID, domain.JobStatusPaused, updatedAt, domain.JobStatusOpen)
	if err != nil {
		return domain.Job{}, err
	}
	if res.RowsAffected() == 0 {
		return domain.Job{}, ErrNotFound
	}
	return r.GetByIDForClient(ctx, jobID, clientID)
}

func (r *JobRepo) Reopen(ctx context.Context, jobID int64, clientID uuid.UUID, updatedAt time.Time) (domain.Job, error) {
	res, err := r.pool.Exec(ctx, `
		update jobs
		set status = $3, paused_at = null, updated_at = $4
		where id = $1 and client_id = $2 and status = $5
	`, jobID, clientID, domain.JobStatusOpen, updatedAt, domain.JobStatusPaused)
	if err != nil {
		return domain.Job{}, err
	}
	if res.RowsAffected() == 0 {
		return domain.Job{}, ErrNotFound
	}
	return r.GetByIDForClient(ctx, jobID, clientID)
}

func (r *JobRepo) MarkFilled(ctx context.Context, jobID int64, clientID uuid.UUID, updatedAt time.Time) (domain.Job, error) {
	res, err := r.pool.Exec(ctx, `
		update jobs
		set status = $3, filled_at = $4, updated_at = $4
		where id = $1 and client_id = $2 and status in ($5, $6)
	`, jobID, clientID, domain.JobStatusFilled, updatedAt, domain.JobStatusOpen, domain.JobStatusPaused)
	if err != nil {
		return domain.Job{}, err
	}
	if res.RowsAffected() == 0 {
		return domain.Job{}, ErrNotFound
	}
	return r.GetByIDForClient(ctx, jobID, clientID)
}

func (r *JobRepo) ListAttachments(ctx context.Context, jobID int64, clientID uuid.UUID) ([]domain.Attachment, error) {
	rows, err := r.pool.Query(ctx, `
		select a.id, a.file_name, a.content_type, a.storage_key, a.url, a.size_bytes
		from job_attachments a
		join jobs j on j.id = a.job_id
		where j.id = $1 and j.client_id = $2
		order by a.id asc
	`, jobID, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attachments := make([]domain.Attachment, 0)
	for rows.Next() {
		var a domain.Attachment
		if err := rows.Scan(&a.ID, &a.FileName, &a.ContentType, &a.StorageKey, &a.URL, &a.SizeBytes); err != nil {
			return nil, err
		}
		attachments = append(attachments, a)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return attachments, nil
}

func (r *JobRepo) GetAttachment(ctx context.Context, jobID int64, attachmentID int64, clientID uuid.UUID) (domain.Attachment, error) {
	var a domain.Attachment
	err := r.pool.QueryRow(ctx, `
		select a.id, a.file_name, a.content_type, a.storage_key, a.url, a.size_bytes
		from job_attachments a
		join jobs j on j.id = a.job_id
		where j.id = $1 and a.id = $2 and j.client_id = $3
	`, jobID, attachmentID, clientID).Scan(&a.ID, &a.FileName, &a.ContentType, &a.StorageKey, &a.URL, &a.SizeBytes)
	if err != nil {
		if isNoRows(err) {
			return domain.Attachment{}, ErrNotFound
		}
		return domain.Attachment{}, err
	}
	return a, nil
}

func (r *JobRepo) hydrateAttachments(ctx context.Context, jobs []domain.Job) error {
	if len(jobs) == 0 {
		return nil
	}

	jobIDArgs := make([]any, 0, len(jobs))
	placeholders := make([]string, 0, len(jobs))
	for i, j := range jobs {
		jobIDArgs = append(jobIDArgs, j.ID)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
	}

	query := fmt.Sprintf(`
		select job_id, id, file_name, content_type, storage_key, url, size_bytes
		from job_attachments
		where job_id in (%s)
		order by job_id asc, id asc
	`, strings.Join(placeholders, ","))

	rows, err := r.pool.Query(ctx, query, jobIDArgs...)
	if err != nil {
		return err
	}
	defer rows.Close()

	byJob := make(map[int64][]domain.Attachment, len(jobs))
	for rows.Next() {
		var jobID int64
		var a domain.Attachment
		if err := rows.Scan(&jobID, &a.ID, &a.FileName, &a.ContentType, &a.StorageKey, &a.URL, &a.SizeBytes); err != nil {
			return err
		}
		byJob[jobID] = append(byJob[jobID], a)
	}
	if rows.Err() != nil {
		return rows.Err()
	}

	for i := range jobs {
		jobs[i].Attachments = byJob[jobs[i].ID]
	}
	return nil
}
