package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"jobconnect/user/internal/application"
	"jobconnect/user/internal/domain"

	"github.com/google/uuid"
)

func parsePage(pageSize uint32, pageToken string) (limit int, offset int, err error) {
	if pageSize == 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset = 0
	if strings.TrimSpace(pageToken) != "" {
		offset, err = strconv.Atoi(strings.TrimSpace(pageToken))
		if err != nil || offset < 0 {
			return 0, 0, fmt.Errorf("invalid page_token")
		}
	}
	return int(pageSize), offset, nil
}

func nextToken(limit, offset, fetched int) string {
	if fetched < limit {
		return ""
	}
	return strconv.Itoa(offset + fetched)
}

func (r *ProfileRepo) freelancerProfileID(ctx context.Context, userID uuid.UUID, publicOnly bool) (int64, error) {
	query := `
		select id
		from profiles
		where user_id = $1 and role = $2 and deleted_at is null
	`
	args := []any{userID, domain.RoleFreelancer}
	if publicOnly {
		query += ` and account_status = 'ACTIVE'`
	}

	var profileID int64
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&profileID); err != nil {
		if isNoRows(err) {
			return 0, ErrNotFound
		}
		return 0, err
	}
	return profileID, nil
}

func (r *ProfileRepo) clientProfileID(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `
		select id
		from profiles
		where user_id = $1 and role = $2 and deleted_at is null
	`
	var profileID int64
	if err := r.pool.QueryRow(ctx, query, userID, domain.RoleClient).Scan(&profileID); err != nil {
		if isNoRows(err) {
			return 0, ErrNotFound
		}
		return 0, err
	}
	return profileID, nil
}

func (r *ProfileRepo) CreatePortfolioItem(ctx context.Context, userID uuid.UUID, in application.PortfolioItem) (application.PortfolioItem, error) {
	profileID, err := r.freelancerProfileID(ctx, userID, false)
	if err != nil {
		return application.PortfolioItem{}, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return application.PortfolioItem{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var out application.PortfolioItem
	var completedAt *time.Time
	err = tx.QueryRow(ctx, `
		insert into portfolio_items (profile_id, title, description, project_url, role_in_project, completed_at, sort_order)
		values ($1,$2,$3,$4,$5,$6,$7)
		returning id, title, description, project_url, role_in_project, completed_at, sort_order, created_at, updated_at
	`, profileID, in.Title, in.Description, in.ProjectURL, in.RoleInProject, in.CompletedAt, in.SortOrder).Scan(
		&out.ID, &out.Title, &out.Description, &out.ProjectURL, &out.RoleInProject, &completedAt, &out.SortOrder, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return application.PortfolioItem{}, err
	}
	out.UserID = userID
	out.CompletedAt = completedAt

	for _, tag := range in.Tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, err := tx.Exec(ctx, `insert into portfolio_tags (portfolio_item_id, tag) values ($1,$2) on conflict do nothing`, out.ID, tag); err != nil {
			return application.PortfolioItem{}, err
		}
	}
	for i, media := range in.Media {
		_, err := tx.Exec(ctx, `
			insert into portfolio_media (portfolio_item_id, media_type, storage_key, external_url, file_name, content_type, size_bytes, width, height, sort_order)
			values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		`, out.ID, strings.ToUpper(strings.TrimPrefix(strings.TrimSpace(media.MediaType), "PORTFOLIO_MEDIA_TYPE_")), media.StorageKey, media.ExternalURL, media.FileName, media.ContentType, media.SizeBytes, media.Width, media.Height, i)
		if err != nil {
			return application.PortfolioItem{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return application.PortfolioItem{}, err
	}
	return r.getPortfolioItem(ctx, userID, out.ID, false)
}

func (r *ProfileRepo) GetPortfolioItem(ctx context.Context, userID uuid.UUID, itemID int64) (application.PortfolioItem, error) {
	return r.getPortfolioItem(ctx, userID, itemID, false)
}

func (r *ProfileRepo) UpdatePortfolioItem(ctx context.Context, userID uuid.UUID, itemID int64, in application.PortfolioItem) (application.PortfolioItem, error) {
	profileID, err := r.freelancerProfileID(ctx, userID, false)
	if err != nil {
		return application.PortfolioItem{}, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return application.PortfolioItem{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	res, err := tx.Exec(ctx, `
		update portfolio_items
		set title=$3, description=$4, project_url=$5, role_in_project=$6, completed_at=$7, sort_order=$8, updated_at=now()
		where id=$1 and profile_id=$2 and deleted_at is null
	`, itemID, profileID, in.Title, in.Description, in.ProjectURL, in.RoleInProject, in.CompletedAt, in.SortOrder)
	if err != nil {
		return application.PortfolioItem{}, err
	}
	if res.RowsAffected() == 0 {
		return application.PortfolioItem{}, ErrNotFound
	}

	if _, err := tx.Exec(ctx, `delete from portfolio_tags where portfolio_item_id = $1`, itemID); err != nil {
		return application.PortfolioItem{}, err
	}
	for _, tag := range in.Tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, err := tx.Exec(ctx, `insert into portfolio_tags (portfolio_item_id, tag) values ($1,$2) on conflict do nothing`, itemID, tag); err != nil {
			return application.PortfolioItem{}, err
		}
	}
	if _, err := tx.Exec(ctx, `delete from portfolio_media where portfolio_item_id = $1`, itemID); err != nil {
		return application.PortfolioItem{}, err
	}
	for i, media := range in.Media {
		if _, err := tx.Exec(ctx, `
			insert into portfolio_media (portfolio_item_id, media_type, storage_key, external_url, file_name, content_type, size_bytes, width, height, sort_order)
			values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		`, itemID, strings.ToUpper(strings.TrimPrefix(strings.TrimSpace(media.MediaType), "PORTFOLIO_MEDIA_TYPE_")), media.StorageKey, media.ExternalURL, media.FileName, media.ContentType, media.SizeBytes, media.Width, media.Height, i); err != nil {
			return application.PortfolioItem{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return application.PortfolioItem{}, err
	}
	return r.getPortfolioItem(ctx, userID, itemID, false)
}

func (r *ProfileRepo) DeletePortfolioItem(ctx context.Context, userID uuid.UUID, itemID int64) (bool, error) {
	profileID, err := r.freelancerProfileID(ctx, userID, false)
	if err != nil {
		return false, err
	}
	res, err := r.pool.Exec(ctx, `update portfolio_items set deleted_at = now(), updated_at = now() where id = $1 and profile_id = $2 and deleted_at is null`, itemID, profileID)
	if err != nil {
		return false, err
	}
	if res.RowsAffected() == 0 {
		return false, ErrNotFound
	}
	return true, nil
}

func (r *ProfileRepo) ListMyPortfolioItems(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string) (application.ListResult[application.PortfolioItem], error) {
	return r.listPortfolioItems(ctx, userID, pageSize, pageToken, false)
}

func (r *ProfileRepo) ListPublicPortfolioItems(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string) (application.ListResult[application.PortfolioItem], error) {
	return r.listPortfolioItems(ctx, userID, pageSize, pageToken, true)
}

func (r *ProfileRepo) listPortfolioItems(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string, publicOnly bool) (application.ListResult[application.PortfolioItem], error) {
	profileID, err := r.freelancerProfileID(ctx, userID, publicOnly)
	if err != nil {
		return application.ListResult[application.PortfolioItem]{}, err
	}
	limit, offset, err := parsePage(pageSize, pageToken)
	if err != nil {
		return application.ListResult[application.PortfolioItem]{}, err
	}

	query := `
		select id, title, description, project_url, role_in_project, completed_at, sort_order, created_at, updated_at
		from portfolio_items
		where profile_id = $1 and deleted_at is null
	`
	args := []any{profileID, limit, offset}
	query += ` order by sort_order asc, id asc limit $2 offset $3`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return application.ListResult[application.PortfolioItem]{}, err
	}
	defer rows.Close()

	items := make([]application.PortfolioItem, 0, limit)
	for rows.Next() {
		var it application.PortfolioItem
		var completedAt *time.Time
		if err := rows.Scan(&it.ID, &it.Title, &it.Description, &it.ProjectURL, &it.RoleInProject, &completedAt, &it.SortOrder, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return application.ListResult[application.PortfolioItem]{}, err
		}
		it.UserID = userID
		it.CompletedAt = completedAt
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return application.ListResult[application.PortfolioItem]{}, err
	}

	for i := range items {
		tags, media, err := r.loadPortfolioItemDetails(ctx, items[i].ID)
		if err != nil {
			return application.ListResult[application.PortfolioItem]{}, err
		}
		items[i].Tags = tags
		items[i].Media = media
	}

	return application.ListResult[application.PortfolioItem]{Items: items, NextPageToken: nextToken(limit, offset, len(items))}, nil
}

func (r *ProfileRepo) loadPortfolioItemDetails(ctx context.Context, itemID int64) ([]string, []application.PortfolioMedia, error) {
	tagsRows, err := r.pool.Query(ctx, `select tag from portfolio_tags where portfolio_item_id = $1 order by tag asc`, itemID)
	if err != nil {
		return nil, nil, err
	}
	defer tagsRows.Close()
	tags := make([]string, 0)
	for tagsRows.Next() {
		var tag string
		if err := tagsRows.Scan(&tag); err != nil {
			return nil, nil, err
		}
		tags = append(tags, tag)
	}

	mediaRows, err := r.pool.Query(ctx, `
		select id, media_type, coalesce(storage_key, ''), coalesce(external_url, ''), coalesce(file_name, ''), coalesce(content_type, ''), coalesce(size_bytes, 0), coalesce(width, 0), coalesce(height, 0), coalesce(sort_order, 0)
		from portfolio_media where portfolio_item_id = $1 order by sort_order asc, id asc
	`, itemID)
	if err != nil {
		return nil, nil, err
	}
	defer mediaRows.Close()
	media := make([]application.PortfolioMedia, 0)
	for mediaRows.Next() {
		var m application.PortfolioMedia
		if err := mediaRows.Scan(&m.ID, &m.MediaType, &m.StorageKey, &m.ExternalURL, &m.FileName, &m.ContentType, &m.SizeBytes, &m.Width, &m.Height, &m.SortOrder); err != nil {
			return nil, nil, err
		}
		media = append(media, m)
	}
	return tags, media, nil
}

func (r *ProfileRepo) getPortfolioItem(ctx context.Context, userID uuid.UUID, itemID int64, publicOnly bool) (application.PortfolioItem, error) {
	profileID, err := r.freelancerProfileID(ctx, userID, publicOnly)
	if err != nil {
		return application.PortfolioItem{}, err
	}
	query := `
		select id, title, description, project_url, role_in_project, completed_at, sort_order, created_at, updated_at
		from portfolio_items
		where id = $1 and profile_id = $2 and deleted_at is null
	`
	var out application.PortfolioItem
	var completedAt *time.Time
	if err := r.pool.QueryRow(ctx, query, itemID, profileID).Scan(&out.ID, &out.Title, &out.Description, &out.ProjectURL, &out.RoleInProject, &completedAt, &out.SortOrder, &out.CreatedAt, &out.UpdatedAt); err != nil {
		if isNoRows(err) {
			return application.PortfolioItem{}, ErrNotFound
		}
		return application.PortfolioItem{}, err
	}
	out.UserID = userID
	out.CompletedAt = completedAt
	tags, media, err := r.loadPortfolioItemDetails(ctx, out.ID)
	if err != nil {
		return application.PortfolioItem{}, err
	}
	out.Tags = tags
	out.Media = media
	return out, nil
}

func (r *ProfileRepo) ReorderPortfolioItems(ctx context.Context, userID uuid.UUID, itemIDs []int64) ([]application.PortfolioItem, error) {
	profileID, err := r.freelancerProfileID(ctx, userID, false)
	if err != nil {
		return nil, err
	}
	if len(itemIDs) == 0 {
		return []application.PortfolioItem{}, nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for i, id := range itemIDs {
		res, err := tx.Exec(ctx, `
			update portfolio_items
			set sort_order = $3, updated_at = now()
			where id = $1 and profile_id = $2 and deleted_at is null
		`, id, profileID, i)
		if err != nil {
			return nil, err
		}
		if res.RowsAffected() == 0 {
			return nil, ErrNotFound
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	result := make([]application.PortfolioItem, 0, len(itemIDs))
	for _, id := range itemIDs {
		it, err := r.getPortfolioItem(ctx, userID, id, false)
		if err != nil {
			return nil, err
		}
		result = append(result, it)
	}
	return result, nil
}

func (r *ProfileRepo) CreateEmployment(ctx context.Context, userID uuid.UUID, in application.Employment) (application.Employment, error) {
	profileID, err := r.freelancerProfileID(ctx, userID, false)
	if err != nil {
		return application.Employment{}, err
	}
	var out application.Employment
	err = r.pool.QueryRow(ctx, `
		insert into profile_employment (profile_id, company_name, title, employment_type, location, is_current, start_date, end_date, description, sort_order)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		returning id, company_name, title, coalesce(employment_type, ''), coalesce(location, ''), is_current, start_date, end_date, coalesce(description, ''), sort_order, created_at, updated_at
	`, profileID, in.CompanyName, in.Title, in.EmploymentType, in.Location, in.IsCurrent, in.StartDate, in.EndDate, in.Description, in.SortOrder).Scan(
		&out.ID, &out.CompanyName, &out.Title, &out.EmploymentType, &out.Location, &out.IsCurrent, &out.StartDate, &out.EndDate, &out.Description, &out.SortOrder, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return application.Employment{}, err
	}
	out.UserID = userID
	return out, nil
}

func (r *ProfileRepo) GetEmployment(ctx context.Context, userID uuid.UUID, employmentID int64) (application.Employment, error) {
	return r.getEmployment(ctx, userID, employmentID, false)
}

func (r *ProfileRepo) UpdateEmployment(ctx context.Context, userID uuid.UUID, employmentID int64, in application.Employment) (application.Employment, error) {
	profileID, err := r.freelancerProfileID(ctx, userID, false)
	if err != nil {
		return application.Employment{}, err
	}
	res, err := r.pool.Exec(ctx, `
		update profile_employment
		set company_name=$3, title=$4, employment_type=$5, location=$6, is_current=$7, start_date=$8, end_date=$9, description=$10, sort_order=$11, updated_at=now()
		where id=$1 and profile_id=$2 and deleted_at is null
	`, employmentID, profileID, in.CompanyName, in.Title, in.EmploymentType, in.Location, in.IsCurrent, in.StartDate, in.EndDate, in.Description, in.SortOrder)
	if err != nil {
		return application.Employment{}, err
	}
	if res.RowsAffected() == 0 {
		return application.Employment{}, ErrNotFound
	}
	return r.getEmployment(ctx, userID, employmentID, false)
}

func (r *ProfileRepo) DeleteEmployment(ctx context.Context, userID uuid.UUID, employmentID int64) (bool, error) {
	profileID, err := r.freelancerProfileID(ctx, userID, false)
	if err != nil {
		return false, err
	}
	res, err := r.pool.Exec(ctx, `update profile_employment set deleted_at = now(), updated_at = now() where id = $1 and profile_id = $2 and deleted_at is null`, employmentID, profileID)
	if err != nil {
		return false, err
	}
	if res.RowsAffected() == 0 {
		return false, ErrNotFound
	}
	return true, nil
}

func (r *ProfileRepo) ListMyEmployment(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string) (application.ListResult[application.Employment], error) {
	return r.listEmployment(ctx, userID, pageSize, pageToken, false)
}

func (r *ProfileRepo) ListPublicEmployment(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string) (application.ListResult[application.Employment], error) {
	return r.listEmployment(ctx, userID, pageSize, pageToken, true)
}

func (r *ProfileRepo) listEmployment(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string, publicOnly bool) (application.ListResult[application.Employment], error) {
	profileID, err := r.freelancerProfileID(ctx, userID, publicOnly)
	if err != nil {
		return application.ListResult[application.Employment]{}, err
	}
	limit, offset, err := parsePage(pageSize, pageToken)
	if err != nil {
		return application.ListResult[application.Employment]{}, err
	}
	query := `
		select id, company_name, title, coalesce(employment_type, ''), coalesce(location, ''), is_current, start_date, end_date, coalesce(description, ''), sort_order, created_at, updated_at
		from profile_employment
		where profile_id = $1 and deleted_at is null
	`
	query += ` order by sort_order asc, start_date desc nulls last, id desc limit $2 offset $3`
	rows, err := r.pool.Query(ctx, query, profileID, limit, offset)
	if err != nil {
		return application.ListResult[application.Employment]{}, err
	}
	defer rows.Close()
	items := make([]application.Employment, 0, limit)
	for rows.Next() {
		var it application.Employment
		if err := rows.Scan(&it.ID, &it.CompanyName, &it.Title, &it.EmploymentType, &it.Location, &it.IsCurrent, &it.StartDate, &it.EndDate, &it.Description, &it.SortOrder, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return application.ListResult[application.Employment]{}, err
		}
		it.UserID = userID
		items = append(items, it)
	}
	return application.ListResult[application.Employment]{Items: items, NextPageToken: nextToken(limit, offset, len(items))}, nil
}

func (r *ProfileRepo) getEmployment(ctx context.Context, userID uuid.UUID, employmentID int64, publicOnly bool) (application.Employment, error) {
	profileID, err := r.freelancerProfileID(ctx, userID, publicOnly)
	if err != nil {
		return application.Employment{}, err
	}
	query := `
		select id, company_name, title, coalesce(employment_type, ''), coalesce(location, ''), is_current, start_date, end_date, coalesce(description, ''), sort_order, created_at, updated_at
		from profile_employment
		where id = $1 and profile_id = $2 and deleted_at is null
	`
	var out application.Employment
	if err := r.pool.QueryRow(ctx, query, employmentID, profileID).Scan(&out.ID, &out.CompanyName, &out.Title, &out.EmploymentType, &out.Location, &out.IsCurrent, &out.StartDate, &out.EndDate, &out.Description, &out.SortOrder, &out.CreatedAt, &out.UpdatedAt); err != nil {
		if isNoRows(err) {
			return application.Employment{}, ErrNotFound
		}
		return application.Employment{}, err
	}
	out.UserID = userID
	return out, nil
}

func (r *ProfileRepo) CreateEducation(ctx context.Context, userID uuid.UUID, in application.Education) (application.Education, error) {
	profileID, err := r.freelancerProfileID(ctx, userID, false)
	if err != nil {
		return application.Education{}, err
	}
	var out application.Education
	err = r.pool.QueryRow(ctx, `
		insert into profile_education (profile_id, school_name, degree, field_of_study, start_date, end_date, is_current, grade, description, sort_order)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		returning id, school_name, coalesce(degree, ''), coalesce(field_of_study, ''), is_current, start_date, end_date, coalesce(grade, ''), coalesce(description, ''), sort_order, created_at, updated_at
	`, profileID, in.SchoolName, in.Degree, in.FieldOfStudy, in.StartDate, in.EndDate, in.IsCurrent, in.Grade, in.Description, in.SortOrder).Scan(
		&out.ID, &out.SchoolName, &out.Degree, &out.FieldOfStudy, &out.IsCurrent, &out.StartDate, &out.EndDate, &out.Grade, &out.Description, &out.SortOrder, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return application.Education{}, err
	}
	out.UserID = userID
	return out, nil
}

func (r *ProfileRepo) GetEducation(ctx context.Context, userID uuid.UUID, educationID int64) (application.Education, error) {
	return r.getEducation(ctx, userID, educationID, false)
}

func (r *ProfileRepo) UpdateEducation(ctx context.Context, userID uuid.UUID, educationID int64, in application.Education) (application.Education, error) {
	profileID, err := r.freelancerProfileID(ctx, userID, false)
	if err != nil {
		return application.Education{}, err
	}
	res, err := r.pool.Exec(ctx, `
		update profile_education
		set school_name=$3, degree=$4, field_of_study=$5, is_current=$6, start_date=$7, end_date=$8, grade=$9, description=$10, sort_order=$11, updated_at=now()
		where id=$1 and profile_id=$2 and deleted_at is null
	`, educationID, profileID, in.SchoolName, in.Degree, in.FieldOfStudy, in.IsCurrent, in.StartDate, in.EndDate, in.Grade, in.Description, in.SortOrder)
	if err != nil {
		return application.Education{}, err
	}
	if res.RowsAffected() == 0 {
		return application.Education{}, ErrNotFound
	}
	return r.getEducation(ctx, userID, educationID, false)
}

func (r *ProfileRepo) DeleteEducation(ctx context.Context, userID uuid.UUID, educationID int64) (bool, error) {
	profileID, err := r.freelancerProfileID(ctx, userID, false)
	if err != nil {
		return false, err
	}
	res, err := r.pool.Exec(ctx, `update profile_education set deleted_at = now(), updated_at = now() where id = $1 and profile_id = $2 and deleted_at is null`, educationID, profileID)
	if err != nil {
		return false, err
	}
	if res.RowsAffected() == 0 {
		return false, ErrNotFound
	}
	return true, nil
}

func (r *ProfileRepo) ListMyEducation(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string) (application.ListResult[application.Education], error) {
	return r.listEducation(ctx, userID, pageSize, pageToken, false)
}

func (r *ProfileRepo) ListPublicEducation(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string) (application.ListResult[application.Education], error) {
	return r.listEducation(ctx, userID, pageSize, pageToken, true)
}

func (r *ProfileRepo) listEducation(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string, publicOnly bool) (application.ListResult[application.Education], error) {
	profileID, err := r.freelancerProfileID(ctx, userID, publicOnly)
	if err != nil {
		return application.ListResult[application.Education]{}, err
	}
	limit, offset, err := parsePage(pageSize, pageToken)
	if err != nil {
		return application.ListResult[application.Education]{}, err
	}
	query := `
		select id, school_name, coalesce(degree, ''), coalesce(field_of_study, ''), is_current, start_date, end_date, coalesce(grade, ''), coalesce(description, ''), sort_order, created_at, updated_at
		from profile_education
		where profile_id = $1 and deleted_at is null
	`
	query += ` order by sort_order asc, start_date desc nulls last, id desc limit $2 offset $3`
	rows, err := r.pool.Query(ctx, query, profileID, limit, offset)
	if err != nil {
		return application.ListResult[application.Education]{}, err
	}
	defer rows.Close()
	items := make([]application.Education, 0, limit)
	for rows.Next() {
		var it application.Education
		if err := rows.Scan(&it.ID, &it.SchoolName, &it.Degree, &it.FieldOfStudy, &it.IsCurrent, &it.StartDate, &it.EndDate, &it.Grade, &it.Description, &it.SortOrder, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return application.ListResult[application.Education]{}, err
		}
		it.UserID = userID
		items = append(items, it)
	}
	return application.ListResult[application.Education]{Items: items, NextPageToken: nextToken(limit, offset, len(items))}, nil
}

func (r *ProfileRepo) getEducation(ctx context.Context, userID uuid.UUID, educationID int64, publicOnly bool) (application.Education, error) {
	profileID, err := r.freelancerProfileID(ctx, userID, publicOnly)
	if err != nil {
		return application.Education{}, err
	}
	query := `
		select id, school_name, coalesce(degree, ''), coalesce(field_of_study, ''), is_current, start_date, end_date, coalesce(grade, ''), coalesce(description, ''), sort_order, created_at, updated_at
		from profile_education
		where id = $1 and profile_id = $2 and deleted_at is null
	`
	var out application.Education
	if err := r.pool.QueryRow(ctx, query, educationID, profileID).Scan(&out.ID, &out.SchoolName, &out.Degree, &out.FieldOfStudy, &out.IsCurrent, &out.StartDate, &out.EndDate, &out.Grade, &out.Description, &out.SortOrder, &out.CreatedAt, &out.UpdatedAt); err != nil {
		if isNoRows(err) {
			return application.Education{}, ErrNotFound
		}
		return application.Education{}, err
	}
	out.UserID = userID
	return out, nil
}

func (r *ProfileRepo) CreateCertification(ctx context.Context, userID uuid.UUID, in application.Certification) (application.Certification, error) {
	profileID, err := r.freelancerProfileID(ctx, userID, false)
	if err != nil {
		return application.Certification{}, err
	}
	var out application.Certification
	err = r.pool.QueryRow(ctx, `
		insert into profile_certifications (profile_id, name, issuing_organization, credential_id, credential_url, issue_date, expiration_date, does_not_expire)
		values ($1,$2,$3,$4,$5,$6,$7,$8)
		returning id, name, issuing_organization, coalesce(credential_id, ''), coalesce(credential_url, ''), issue_date, expiration_date, does_not_expire, created_at, updated_at
	`, profileID, in.Name, in.IssuingOrganization, in.CredentialID, in.CredentialURL, in.IssueDate, in.ExpirationDate, in.DoesNotExpire).Scan(
		&out.ID, &out.Name, &out.IssuingOrganization, &out.CredentialID, &out.CredentialURL, &out.IssueDate, &out.ExpirationDate, &out.DoesNotExpire, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return application.Certification{}, err
	}
	out.UserID = userID
	return out, nil
}

func (r *ProfileRepo) GetCertification(ctx context.Context, userID uuid.UUID, certificationID int64) (application.Certification, error) {
	return r.getCertification(ctx, userID, certificationID, false)
}

func (r *ProfileRepo) UpdateCertification(ctx context.Context, userID uuid.UUID, certificationID int64, in application.Certification) (application.Certification, error) {
	profileID, err := r.freelancerProfileID(ctx, userID, false)
	if err != nil {
		return application.Certification{}, err
	}
	res, err := r.pool.Exec(ctx, `
		update profile_certifications
		set name=$3, issuing_organization=$4, credential_id=$5, credential_url=$6, issue_date=$7, expiration_date=$8, does_not_expire=$9, updated_at=now()
		where id=$1 and profile_id=$2 and deleted_at is null
	`, certificationID, profileID, in.Name, in.IssuingOrganization, in.CredentialID, in.CredentialURL, in.IssueDate, in.ExpirationDate, in.DoesNotExpire)
	if err != nil {
		return application.Certification{}, err
	}
	if res.RowsAffected() == 0 {
		return application.Certification{}, ErrNotFound
	}
	return r.getCertification(ctx, userID, certificationID, false)
}

func (r *ProfileRepo) DeleteCertification(ctx context.Context, userID uuid.UUID, certificationID int64) (bool, error) {
	profileID, err := r.freelancerProfileID(ctx, userID, false)
	if err != nil {
		return false, err
	}
	res, err := r.pool.Exec(ctx, `update profile_certifications set deleted_at = now(), updated_at = now() where id = $1 and profile_id = $2 and deleted_at is null`, certificationID, profileID)
	if err != nil {
		return false, err
	}
	if res.RowsAffected() == 0 {
		return false, ErrNotFound
	}
	return true, nil
}

func (r *ProfileRepo) ListMyCertifications(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string) (application.ListResult[application.Certification], error) {
	return r.listCertifications(ctx, userID, pageSize, pageToken, false)
}

func (r *ProfileRepo) ListPublicCertifications(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string) (application.ListResult[application.Certification], error) {
	return r.listCertifications(ctx, userID, pageSize, pageToken, true)
}

func (r *ProfileRepo) listCertifications(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string, publicOnly bool) (application.ListResult[application.Certification], error) {
	profileID, err := r.freelancerProfileID(ctx, userID, publicOnly)
	if err != nil {
		return application.ListResult[application.Certification]{}, err
	}
	limit, offset, err := parsePage(pageSize, pageToken)
	if err != nil {
		return application.ListResult[application.Certification]{}, err
	}
	query := `
		select id, name, issuing_organization, coalesce(credential_id, ''), coalesce(credential_url, ''), issue_date, expiration_date, does_not_expire, created_at, updated_at
		from profile_certifications
		where profile_id = $1 and deleted_at is null
	`
	query += ` order by issue_date desc nulls last, id desc limit $2 offset $3`
	rows, err := r.pool.Query(ctx, query, profileID, limit, offset)
	if err != nil {
		return application.ListResult[application.Certification]{}, err
	}
	defer rows.Close()
	items := make([]application.Certification, 0, limit)
	for rows.Next() {
		var it application.Certification
		if err := rows.Scan(&it.ID, &it.Name, &it.IssuingOrganization, &it.CredentialID, &it.CredentialURL, &it.IssueDate, &it.ExpirationDate, &it.DoesNotExpire, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return application.ListResult[application.Certification]{}, err
		}
		it.UserID = userID
		items = append(items, it)
	}
	return application.ListResult[application.Certification]{Items: items, NextPageToken: nextToken(limit, offset, len(items))}, nil
}

func (r *ProfileRepo) getCertification(ctx context.Context, userID uuid.UUID, certificationID int64, publicOnly bool) (application.Certification, error) {
	profileID, err := r.freelancerProfileID(ctx, userID, publicOnly)
	if err != nil {
		return application.Certification{}, err
	}
	query := `
		select id, name, issuing_organization, coalesce(credential_id, ''), coalesce(credential_url, ''), issue_date, expiration_date, does_not_expire, created_at, updated_at
		from profile_certifications
		where id = $1 and profile_id = $2 and deleted_at is null
	`
	var out application.Certification
	if err := r.pool.QueryRow(ctx, query, certificationID, profileID).Scan(&out.ID, &out.Name, &out.IssuingOrganization, &out.CredentialID, &out.CredentialURL, &out.IssueDate, &out.ExpirationDate, &out.DoesNotExpire, &out.CreatedAt, &out.UpdatedAt); err != nil {
		if isNoRows(err) {
			return application.Certification{}, ErrNotFound
		}
		return application.Certification{}, err
	}
	out.UserID = userID
	return out, nil
}

func (r *ProfileRepo) UpsertLanguages(ctx context.Context, userID uuid.UUID, languages []application.LanguageProficiency) ([]application.LanguageProficiency, error) {
	profileID, err := r.freelancerProfileID(ctx, userID, false)
	if err != nil {
		return nil, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `delete from profile_languages where profile_id = $1`, profileID); err != nil {
		return nil, err
	}
	for _, lang := range languages {
		code := strings.TrimSpace(strings.ToLower(lang.LanguageCode))
		if code == "" {
			continue
		}
		proficiency := strings.ToUpper(strings.TrimSpace(lang.Proficiency))
		if _, err := tx.Exec(ctx, `
			insert into profile_languages (profile_id, language_code, proficiency, created_at, updated_at)
			values ($1,$2,$3, now(), now())
		`, profileID, code, proficiency); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetMyLanguages(ctx, userID)
}

func (r *ProfileRepo) GetMyLanguages(ctx context.Context, userID uuid.UUID) ([]application.LanguageProficiency, error) {
	return r.getLanguages(ctx, userID, false)
}

func (r *ProfileRepo) GetPublicLanguages(ctx context.Context, userID uuid.UUID) ([]application.LanguageProficiency, error) {
	return r.getLanguages(ctx, userID, true)
}

func (r *ProfileRepo) getLanguages(ctx context.Context, userID uuid.UUID, publicOnly bool) ([]application.LanguageProficiency, error) {
	profileID, err := r.freelancerProfileID(ctx, userID, publicOnly)
	if err != nil {
		return nil, err
	}
	query := `select language_code, proficiency from profile_languages where profile_id = $1`
	query += ` order by language_code asc`
	rows, err := r.pool.Query(ctx, query, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]application.LanguageProficiency, 0)
	for rows.Next() {
		var lp application.LanguageProficiency
		if err := rows.Scan(&lp.LanguageCode, &lp.Proficiency); err != nil {
			return nil, err
		}
		out = append(out, lp)
	}
	return out, rows.Err()
}

func (r *ProfileRepo) SetAvailability(ctx context.Context, userID uuid.UUID, in application.AvailabilitySettings) (application.AvailabilitySettings, error) {
	profileID, err := r.freelancerProfileID(ctx, userID, false)
	if err != nil {
		return application.AvailabilitySettings{}, err
	}

	availability := strings.ToUpper(strings.TrimPrefix(strings.TrimSpace(in.Availability), "AVAILABILITY_"))
	if availability == "" {
		availability = domain.AvailabilityAsNeeded
	}

	if _, err := r.pool.Exec(ctx, `
		update freelancer_profiles
		set availability = $2
		where profile_id = $1
	`, profileID, availability); err != nil {
		return application.AvailabilitySettings{}, err
	}

	if _, err := r.pool.Exec(ctx, `
		insert into freelancer_work_preferences (profile_id, weekly_capacity_hours)
		values ($1, $2)
		on conflict (profile_id) do update set
			weekly_capacity_hours = excluded.weekly_capacity_hours,
			updated_at = now()
	`, profileID, in.WeeklyCapacityHours); err != nil {
		return application.AvailabilitySettings{}, err
	}

	return r.GetAvailability(ctx, userID)
}

func (r *ProfileRepo) GetAvailability(ctx context.Context, userID uuid.UUID) (application.AvailabilitySettings, error) {
	profileID, err := r.freelancerProfileID(ctx, userID, false)
	if err != nil {
		return application.AvailabilitySettings{}, err
	}

	out := application.AvailabilitySettings{}
	if err := r.pool.QueryRow(ctx, `
		select coalesce(availability, 'AS_NEEDED')
		from freelancer_profiles
		where profile_id = $1
	`, profileID).Scan(&out.Availability); err != nil {
		if isNoRows(err) {
			return application.AvailabilitySettings{}, ErrNotFound
		}
		return application.AvailabilitySettings{}, err
	}

	if err := r.pool.QueryRow(ctx, `
		select weekly_capacity_hours
		from freelancer_work_preferences
		where profile_id = $1
	`, profileID).Scan(&out.WeeklyCapacityHours); err != nil && !isNoRows(err) {
		return application.AvailabilitySettings{}, err
	}

	return out, nil
}

func (r *ProfileRepo) SetRates(ctx context.Context, userID uuid.UUID, in application.RateSettings) (application.RateSettings, error) {
	profileID, err := r.freelancerProfileID(ctx, userID, false)
	if err != nil {
		return application.RateSettings{}, err
	}

	if _, err := r.pool.Exec(ctx, `
		update freelancer_profiles
		set hourly_rate = $2
		where profile_id = $1
	`, profileID, in.HourlyRate); err != nil {
		return application.RateSettings{}, err
	}

	return r.GetRates(ctx, userID)
}

func (r *ProfileRepo) GetRates(ctx context.Context, userID uuid.UUID) (application.RateSettings, error) {
	profileID, err := r.freelancerProfileID(ctx, userID, false)
	if err != nil {
		return application.RateSettings{}, err
	}

	out := application.RateSettings{Currency: "USD"}
	if err := r.pool.QueryRow(ctx, `
		select coalesce(hourly_rate, 0)
		from freelancer_profiles
		where profile_id = $1
	`, profileID).Scan(&out.HourlyRate); err != nil {
		if isNoRows(err) {
			return application.RateSettings{}, ErrNotFound
		}
		return application.RateSettings{}, err
	}

	return out, nil
}

func (r *ProfileRepo) SetWorkPreferences(ctx context.Context, userID uuid.UUID, in application.WorkPreferences) (application.WorkPreferences, error) {
	profileID, err := r.freelancerProfileID(ctx, userID, false)
	if err != nil {
		return application.WorkPreferences{}, err
	}

	contractTypes := make([]string, 0, len(in.ContractTypes))
	for _, t := range in.ContractTypes {
		trimmed := strings.TrimSpace(t)
		if trimmed != "" {
			contractTypes = append(contractTypes, trimmed)
		}
	}
	rawContractTypes, err := json.Marshal(contractTypes)
	if err != nil {
		return application.WorkPreferences{}, err
	}

	if _, err := r.pool.Exec(ctx, `
		insert into freelancer_work_preferences (profile_id, preferred_project_length, min_budget_usd, max_budget_usd, contract_types)
		values ($1, $2, $3, $4, $5::jsonb)
		on conflict (profile_id) do update set
			preferred_project_length = excluded.preferred_project_length,
			min_budget_usd = excluded.min_budget_usd,
			max_budget_usd = excluded.max_budget_usd,
			contract_types = excluded.contract_types,
			updated_at = now()
	`, profileID, strings.TrimSpace(in.PreferredProjectLength), in.MinBudgetUSD, in.MaxBudgetUSD, string(rawContractTypes)); err != nil {
		return application.WorkPreferences{}, err
	}

	return r.GetWorkPreferences(ctx, userID)
}

func (r *ProfileRepo) GetWorkPreferences(ctx context.Context, userID uuid.UUID) (application.WorkPreferences, error) {
	profileID, err := r.freelancerProfileID(ctx, userID, false)
	if err != nil {
		return application.WorkPreferences{}, err
	}

	out := application.WorkPreferences{}
	var rawContractTypes []byte
	err = r.pool.QueryRow(ctx, `
		select
			coalesce(preferred_project_length, ''),
			coalesce(min_budget_usd, 0),
			coalesce(max_budget_usd, 0),
			coalesce(contract_types, '[]'::jsonb)
		from freelancer_work_preferences
		where profile_id = $1
	`, profileID).Scan(&out.PreferredProjectLength, &out.MinBudgetUSD, &out.MaxBudgetUSD, &rawContractTypes)
	if err != nil {
		if isNoRows(err) {
			return out, nil
		}
		return application.WorkPreferences{}, err
	}

	if len(rawContractTypes) > 0 {
		_ = json.Unmarshal(rawContractTypes, &out.ContractTypes)
	}

	return out, nil
}

func (r *ProfileRepo) GetCompany(ctx context.Context, userID uuid.UUID) (application.CompanySettings, error) {
	profileID, err := r.clientProfileID(ctx, userID)
	if err != nil {
		return application.CompanySettings{}, err
	}

	out := application.CompanySettings{}
	err = r.pool.QueryRow(ctx, `
		select
			coalesce(company_name, ''),
			coalesce(p.tax_id, '')
		from client_profiles cp
		join profiles p on p.id = cp.profile_id
		where cp.profile_id = $1
	`, profileID).Scan(&out.CompanyName, &out.TaxID)
	if err != nil {
		if isNoRows(err) {
			return out, nil
		}
		return application.CompanySettings{}, err
	}
	return out, nil
}

func (r *ProfileRepo) UpdateCompany(ctx context.Context, userID uuid.UUID, in application.CompanySettings) (application.CompanySettings, error) {
	profileID, err := r.clientProfileID(ctx, userID)
	if err != nil {
		return application.CompanySettings{}, err
	}

	if _, err := r.pool.Exec(ctx, `
		insert into client_profiles (profile_id, company_name)
		values ($1, $2)
		on conflict (profile_id) do update set
			company_name = excluded.company_name
	`, profileID, strings.TrimSpace(in.CompanyName)); err != nil {
		return application.CompanySettings{}, err
	}

	if _, err := r.pool.Exec(ctx, `
		update profiles
		set tax_id = $2,
			updated_at = now()
		where id = $1
	`, profileID, strings.TrimSpace(in.TaxID)); err != nil {
		return application.CompanySettings{}, err
	}

	updated, err := r.GetCompany(ctx, userID)
	if err != nil {
		return application.CompanySettings{}, err
	}
	return updated, nil
}

func (r *ProfileRepo) GetHiringPreferences(ctx context.Context, userID uuid.UUID) (application.HiringPreferences, error) {
	profileID, err := r.clientProfileID(ctx, userID)
	if err != nil {
		return application.HiringPreferences{}, err
	}

	out := application.HiringPreferences{}
	var rawLevels []byte
	var rawLocations []byte
	err = r.pool.QueryRow(ctx, `
		select
			coalesce(min_hourly_rate, 0),
			coalesce(max_hourly_rate, 0),
			coalesce(preferred_experience_levels, '[]'::jsonb),
			coalesce(preferred_locations, '[]'::jsonb)
		from client_hiring_preferences
		where profile_id = $1
	`, profileID).Scan(&out.MinHourlyRate, &out.MaxHourlyRate, &rawLevels, &rawLocations)
	if err != nil {
		if isNoRows(err) {
			return out, nil
		}
		return application.HiringPreferences{}, err
	}
	if len(rawLevels) > 0 {
		_ = json.Unmarshal(rawLevels, &out.PreferredExperienceLevels)
	}
	if len(rawLocations) > 0 {
		_ = json.Unmarshal(rawLocations, &out.PreferredLocations)
	}
	return out, nil
}

func (r *ProfileRepo) UpdateHiringPreferences(ctx context.Context, userID uuid.UUID, in application.HiringPreferences) (application.HiringPreferences, error) {
	profileID, err := r.clientProfileID(ctx, userID)
	if err != nil {
		return application.HiringPreferences{}, err
	}

	rawLevels, err := json.Marshal(in.PreferredExperienceLevels)
	if err != nil {
		return application.HiringPreferences{}, err
	}
	rawLocations, err := json.Marshal(in.PreferredLocations)
	if err != nil {
		return application.HiringPreferences{}, err
	}

	if _, err := r.pool.Exec(ctx, `
		insert into client_hiring_preferences (profile_id, min_hourly_rate, max_hourly_rate, preferred_experience_levels, preferred_locations)
		values ($1, $2, $3, $4::jsonb, $5::jsonb)
		on conflict (profile_id) do update set
			min_hourly_rate = excluded.min_hourly_rate,
			max_hourly_rate = excluded.max_hourly_rate,
			preferred_experience_levels = excluded.preferred_experience_levels,
			preferred_locations = excluded.preferred_locations,
			updated_at = now()
	`, profileID, in.MinHourlyRate, in.MaxHourlyRate, string(rawLevels), string(rawLocations)); err != nil {
		return application.HiringPreferences{}, err
	}

	return r.GetHiringPreferences(ctx, userID)
}

func (r *ProfileRepo) SaveFreelancer(ctx context.Context, userID uuid.UUID, freelancerUserID uuid.UUID) (application.SavedFreelancer, error) {
	profileID, err := r.clientProfileID(ctx, userID)
	if err != nil {
		return application.SavedFreelancer{}, err
	}

	var savedAt time.Time
	err = r.pool.QueryRow(ctx, `
		insert into client_saved_freelancers (profile_id, freelancer_user_id)
		values ($1, $2)
		on conflict (profile_id, freelancer_user_id) do update set
			created_at = client_saved_freelancers.created_at
		returning created_at
	`, profileID, freelancerUserID).Scan(&savedAt)
	if err != nil {
		return application.SavedFreelancer{}, err
	}

	return application.SavedFreelancer{FreelancerUserID: freelancerUserID, SavedAt: savedAt}, nil
}

func (r *ProfileRepo) ListSavedFreelancers(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string) (application.ListResult[application.SavedFreelancer], error) {
	profileID, err := r.clientProfileID(ctx, userID)
	if err != nil {
		return application.ListResult[application.SavedFreelancer]{}, err
	}
	limit, offset, err := parsePage(pageSize, pageToken)
	if err != nil {
		return application.ListResult[application.SavedFreelancer]{}, err
	}

	rows, err := r.pool.Query(ctx, `
		select freelancer_user_id, created_at
		from client_saved_freelancers
		where profile_id = $1
		order by created_at desc, freelancer_user_id asc
		limit $2 offset $3
	`, profileID, limit, offset)
	if err != nil {
		return application.ListResult[application.SavedFreelancer]{}, err
	}
	defer rows.Close()

	items := make([]application.SavedFreelancer, 0, limit)
	for rows.Next() {
		var item application.SavedFreelancer
		if err := rows.Scan(&item.FreelancerUserID, &item.SavedAt); err != nil {
			return application.ListResult[application.SavedFreelancer]{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return application.ListResult[application.SavedFreelancer]{}, err
	}

	return application.ListResult[application.SavedFreelancer]{Items: items, NextPageToken: nextToken(limit, offset, len(items))}, nil
}

func (r *ProfileRepo) RemoveSavedFreelancer(ctx context.Context, userID uuid.UUID, freelancerUserID uuid.UUID) (bool, error) {
	profileID, err := r.clientProfileID(ctx, userID)
	if err != nil {
		return false, err
	}

	res, err := r.pool.Exec(ctx, `
		delete from client_saved_freelancers
		where profile_id = $1 and freelancer_user_id = $2
	`, profileID, freelancerUserID)
	if err != nil {
		return false, err
	}
	if res.RowsAffected() == 0 {
		return false, ErrNotFound
	}
	return true, nil
}

func (r *ProfileRepo) UpsertFreelancerNote(ctx context.Context, userID uuid.UUID, freelancerUserID uuid.UUID, note string) (application.FreelancerNote, error) {
	profileID, err := r.clientProfileID(ctx, userID)
	if err != nil {
		return application.FreelancerNote{}, err
	}

	var updatedAt time.Time
	err = r.pool.QueryRow(ctx, `
		insert into client_freelancer_notes (profile_id, freelancer_user_id, note)
		values ($1, $2, $3)
		on conflict (profile_id, freelancer_user_id) do update set
			note = excluded.note,
			updated_at = now()
		returning updated_at
	`, profileID, freelancerUserID, strings.TrimSpace(note)).Scan(&updatedAt)
	if err != nil {
		return application.FreelancerNote{}, err
	}

	return application.FreelancerNote{FreelancerUserID: freelancerUserID, Note: strings.TrimSpace(note), UpdatedAt: updatedAt}, nil
}

func (r *ProfileRepo) GetFreelancerNote(ctx context.Context, userID uuid.UUID, freelancerUserID uuid.UUID) (application.FreelancerNote, error) {
	profileID, err := r.clientProfileID(ctx, userID)
	if err != nil {
		return application.FreelancerNote{}, err
	}

	var out application.FreelancerNote
	err = r.pool.QueryRow(ctx, `
		select note, updated_at
		from client_freelancer_notes
		where profile_id = $1 and freelancer_user_id = $2
	`, profileID, freelancerUserID).Scan(&out.Note, &out.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return application.FreelancerNote{}, ErrNotFound
		}
		return application.FreelancerNote{}, err
	}
	out.FreelancerUserID = freelancerUserID
	return out, nil
}

func (r *ProfileRepo) ensureAdminRequester(ctx context.Context, requesterUserID uuid.UUID) error {
	var role string
	err := r.pool.QueryRow(ctx, `
		select role
		from profiles
		where user_id = $1 and deleted_at is null
	`, requesterUserID).Scan(&role)
	if err != nil {
		if isNoRows(err) {
			return ErrNotFound
		}
		return err
	}
	if strings.TrimSpace(strings.ToLower(role)) != domain.RoleAdmin {
		return fmt.Errorf("admin role required")
	}
	return nil
}

func (r *ProfileRepo) ListUsers(ctx context.Context, requesterUserID uuid.UUID, filter application.ListUsersFilter) (application.ListResult[application.UserSummary], error) {
	if err := r.ensureAdminRequester(ctx, requesterUserID); err != nil {
		return application.ListResult[application.UserSummary]{}, err
	}

	limit, offset, err := parsePage(filter.PageSize, filter.PageToken)
	if err != nil {
		return application.ListResult[application.UserSummary]{}, err
	}

	query := `
		select
			user_id,
			role,
			coalesce(account_status, 'ACTIVE'),
			coalesce(first_name, ''),
			coalesce(last_name, ''),
			coalesce(display_name, ''),
			coalesce(avatar_url, ''),
			created_at,
			updated_at
		from profiles
		where deleted_at is null
	`
	args := []any{}
	argPos := 1

	if v := strings.TrimSpace(filter.Role); v != "" {
		query += fmt.Sprintf(" and role = $%d", argPos)
		args = append(args, strings.ToLower(v))
		argPos++
	}
	if v := strings.TrimSpace(filter.Status); v != "" {
		query += fmt.Sprintf(" and upper(account_status) = upper($%d)", argPos)
		args = append(args, strings.TrimPrefix(strings.ToUpper(v), "ACCOUNT_STATUS_"))
		argPos++
	}
	if v := strings.TrimSpace(filter.Q); v != "" {
		query += fmt.Sprintf(" and (user_id::text ilike $%d or first_name ilike $%d or last_name ilike $%d or display_name ilike $%d)", argPos, argPos, argPos, argPos)
		args = append(args, "%"+v+"%")
		argPos++
	}

	query += fmt.Sprintf(" order by created_at desc, user_id asc limit $%d offset $%d", argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return application.ListResult[application.UserSummary]{}, err
	}
	defer rows.Close()

	items := make([]application.UserSummary, 0, limit)
	for rows.Next() {
		var item application.UserSummary
		if err := rows.Scan(&item.UserID, &item.Role, &item.Status, &item.FirstName, &item.LastName, &item.DisplayName, &item.AvatarURL, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return application.ListResult[application.UserSummary]{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return application.ListResult[application.UserSummary]{}, err
	}

	return application.ListResult[application.UserSummary]{Items: items, NextPageToken: nextToken(limit, offset, len(items))}, nil
}

func (r *ProfileRepo) CreateImpersonationToken(ctx context.Context, requesterUserID uuid.UUID, targetUserID uuid.UUID, reason string, ttlSeconds uint32) (application.ImpersonationToken, error) {
	if err := r.ensureAdminRequester(ctx, requesterUserID); err != nil {
		return application.ImpersonationToken{}, err
	}

	ttl := time.Duration(ttlSeconds) * time.Second
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	if ttl > 24*time.Hour {
		ttl = 24 * time.Hour
	}

	tokenID := uuid.New()
	expiresAt := time.Now().UTC().Add(ttl)

	if _, err := r.pool.Exec(ctx, `
		insert into admin_impersonation_tokens (token_id, admin_user_id, target_user_id, reason, expires_at)
		values ($1, $2, $3, $4, $5)
	`, tokenID, requesterUserID, targetUserID, strings.TrimSpace(reason), expiresAt); err != nil {
		return application.ImpersonationToken{}, err
	}

	return application.ImpersonationToken{Token: tokenID.String(), ExpiresAt: expiresAt}, nil
}

func (r *ProfileRepo) GetUserAuditSummary(ctx context.Context, requesterUserID uuid.UUID, targetUserID uuid.UUID) (application.UserAuditSummary, error) {
	if err := r.ensureAdminRequester(ctx, requesterUserID); err != nil {
		return application.UserAuditSummary{}, err
	}

	var out application.UserAuditSummary
	out.UserID = targetUserID

	err := r.pool.QueryRow(ctx, `
		select
			coalesce(account_status, 'ACTIVE'),
			updated_at
		from profiles
		where user_id = $1 and deleted_at is null
	`, targetUserID).Scan(&out.Status, &out.ProfileUpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return application.UserAuditSummary{}, ErrNotFound
		}
		return application.UserAuditSummary{}, err
	}

	_ = r.pool.QueryRow(ctx, `
		select updated_at
		from profile_avatars
		where user_id = $1
	`, targetUserID).Scan(&out.AvatarUpdatedAt)

	_ = r.pool.QueryRow(ctx, `
		select count(*)
		from client_saved_freelancers sf
		join profiles p on p.id = sf.profile_id
		where p.user_id = $1 and p.deleted_at is null
	`, targetUserID).Scan(&out.SavedFreelancersCount)

	_ = r.pool.QueryRow(ctx, `
		select count(*)
		from portfolio_items pi
		join profiles p on p.id = pi.profile_id
		where p.user_id = $1 and p.deleted_at is null and pi.deleted_at is null
	`, targetUserID).Scan(&out.PortfolioItemsCount)

	return out, nil
}
