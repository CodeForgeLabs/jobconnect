package db

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"jobconnect/user/internal/application"
	"jobconnect/user/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func encodeSavedFreelancersCursor(createdAt time.Time, freelancerUserID uuid.UUID) string {
	raw := fmt.Sprintf("%d|%s", createdAt.UTC().UnixNano(), freelancerUserID.String())
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeSavedFreelancersCursor(token string) (time.Time, uuid.UUID, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return time.Time{}, uuid.Nil, fmt.Errorf("invalid page_token")
	}
	parts := strings.SplitN(string(decoded), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, uuid.Nil, fmt.Errorf("invalid page_token")
	}
	nanos, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return time.Time{}, uuid.Nil, fmt.Errorf("invalid page_token")
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return time.Time{}, uuid.Nil, fmt.Errorf("invalid page_token")
	}
	return time.Unix(0, nanos).UTC(), id, nil
}

func (r *ProfileRepo) freelancerProfileID(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `
		select id
		from profiles
		where user_id = $1 and role = $2 and deleted_at is null
	`
	args := []any{userID, domain.RoleFreelancer}

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
	profileID, err := r.freelancerProfileID(ctx, userID)
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
		insert into portfolio_items (profile_id, title, description, project_url, role_in_project, completed_at)
		values ($1,$2,$3,$4,$5,$6)
		returning id, title, description, project_url, role_in_project, completed_at, created_at, updated_at
	`, profileID, in.Title, in.Description, in.ProjectURL, in.RoleInProject, in.CompletedAt).Scan(
		&out.ID, &out.Title, &out.Description, &out.ProjectURL, &out.RoleInProject, &completedAt, &out.CreatedAt, &out.UpdatedAt,
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
	for _, media := range in.Media {
		_, err := tx.Exec(ctx, `
			insert into portfolio_media (portfolio_item_id, media_type, storage_key, external_url, file_name, content_type, size_bytes, width, height)
			values ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		`, out.ID, strings.ToUpper(strings.TrimPrefix(strings.TrimSpace(media.MediaType), "PORTFOLIO_MEDIA_TYPE_")), media.StorageKey, media.ExternalURL, media.FileName, media.ContentType, media.SizeBytes, media.Width, media.Height)
		if err != nil {
			return application.PortfolioItem{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return application.PortfolioItem{}, err
	}
	return r.getPortfolioItem(ctx, userID, out.ID)
}

func (r *ProfileRepo) GetPortfolioItem(ctx context.Context, userID uuid.UUID, itemID int64) (application.PortfolioItem, error) {
	return r.getPortfolioItem(ctx, userID, itemID)
}

func (r *ProfileRepo) UpdatePortfolioItem(ctx context.Context, userID uuid.UUID, itemID int64, in application.PortfolioItem) (application.PortfolioItem, error) {
	profileID, err := r.freelancerProfileID(ctx, userID)
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
		set title=$3, description=$4, project_url=$5, role_in_project=$6, completed_at=$7, updated_at=now()
		where id=$1 and profile_id=$2 and deleted_at is null
	`, itemID, profileID, in.Title, in.Description, in.ProjectURL, in.RoleInProject, in.CompletedAt)
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
	for _, media := range in.Media {
		if _, err := tx.Exec(ctx, `
			insert into portfolio_media (portfolio_item_id, media_type, storage_key, external_url, file_name, content_type, size_bytes, width, height)
			values ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		`, itemID, strings.ToUpper(strings.TrimPrefix(strings.TrimSpace(media.MediaType), "PORTFOLIO_MEDIA_TYPE_")), media.StorageKey, media.ExternalURL, media.FileName, media.ContentType, media.SizeBytes, media.Width, media.Height); err != nil {
			return application.PortfolioItem{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return application.PortfolioItem{}, err
	}
	return r.getPortfolioItem(ctx, userID, itemID)
}

func (r *ProfileRepo) DeletePortfolioItem(ctx context.Context, userID uuid.UUID, itemID int64) (bool, error) {
	profileID, err := r.freelancerProfileID(ctx, userID)
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
	return r.listPortfolioItems(ctx, userID, pageSize, pageToken)
}

func (r *ProfileRepo) listPortfolioItems(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string) (application.ListResult[application.PortfolioItem], error) {
	profileID, err := r.freelancerProfileID(ctx, userID)
	if err != nil {
		return application.ListResult[application.PortfolioItem]{}, err
	}
	limit, offset, err := parsePage(pageSize, pageToken)
	if err != nil {
		return application.ListResult[application.PortfolioItem]{}, err
	}

	query := `
		select id, title, description, project_url, role_in_project, completed_at, created_at, updated_at
		from portfolio_items
		where profile_id = $1 and deleted_at is null
	`
	args := []any{profileID, limit, offset}
	query += ` order by created_at desc, id desc limit $2 offset $3`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return application.ListResult[application.PortfolioItem]{}, err
	}
	defer rows.Close()

	items := make([]application.PortfolioItem, 0, limit)
	for rows.Next() {
		var it application.PortfolioItem
		var completedAt *time.Time
		if err := rows.Scan(&it.ID, &it.Title, &it.Description, &it.ProjectURL, &it.RoleInProject, &completedAt, &it.CreatedAt, &it.UpdatedAt); err != nil {
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
		select id, media_type, coalesce(storage_key, ''), coalesce(external_url, ''), coalesce(file_name, ''), coalesce(content_type, ''), coalesce(size_bytes, 0), coalesce(width, 0), coalesce(height, 0)
		from portfolio_media where portfolio_item_id = $1 order by id asc
	`, itemID)
	if err != nil {
		return nil, nil, err
	}
	defer mediaRows.Close()
	media := make([]application.PortfolioMedia, 0)
	for mediaRows.Next() {
		var m application.PortfolioMedia
		if err := mediaRows.Scan(&m.ID, &m.MediaType, &m.StorageKey, &m.ExternalURL, &m.FileName, &m.ContentType, &m.SizeBytes, &m.Width, &m.Height); err != nil {
			return nil, nil, err
		}
		media = append(media, m)
	}
	return tags, media, nil
}

func (r *ProfileRepo) getPortfolioItem(ctx context.Context, userID uuid.UUID, itemID int64) (application.PortfolioItem, error) {
	profileID, err := r.freelancerProfileID(ctx, userID)
	if err != nil {
		return application.PortfolioItem{}, err
	}
	query := `
		select id, title, description, project_url, role_in_project, completed_at, created_at, updated_at
		from portfolio_items
		where id = $1 and profile_id = $2 and deleted_at is null
	`
	var out application.PortfolioItem
	var completedAt *time.Time
	if err := r.pool.QueryRow(ctx, query, itemID, profileID).Scan(&out.ID, &out.Title, &out.Description, &out.ProjectURL, &out.RoleInProject, &completedAt, &out.CreatedAt, &out.UpdatedAt); err != nil {
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

func (r *ProfileRepo) SetWorkPreferences(ctx context.Context, userID uuid.UUID, in application.WorkPreferences) (application.WorkPreferences, error) {
	profileID, err := r.freelancerProfileID(ctx, userID)
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
		insert into freelancer_work_preferences (profile_id, preferred_project_length, min_budget, max_budget, contract_types, weekly_capacity_hours)
		values ($1, $2, $3, $4, $5::jsonb, $6)
		on conflict (profile_id) do update set
			preferred_project_length = excluded.preferred_project_length,
			min_budget = excluded.min_budget,
			max_budget = excluded.max_budget,
			contract_types = excluded.contract_types,
			weekly_capacity_hours = excluded.weekly_capacity_hours,
			updated_at = now()
	`, profileID, application.CanonicalProjectLengthOrUnspecified(in.PreferredProjectLength), in.MinBudgetUSD, in.MaxBudgetUSD, string(rawContractTypes), in.WeeklyCapacityHours); err != nil {
		return application.WorkPreferences{}, err
	}

	return r.GetWorkPreferences(ctx, userID)
}

func (r *ProfileRepo) GetWorkPreferences(ctx context.Context, userID uuid.UUID) (application.WorkPreferences, error) {
	profileID, err := r.freelancerProfileID(ctx, userID)
	if err != nil {
		return application.WorkPreferences{}, err
	}

	out := application.WorkPreferences{}
	var rawContractTypes []byte
	err = r.pool.QueryRow(ctx, `
		select
			coalesce(preferred_project_length, 'PROJECT_LENGTH_UNSPECIFIED'),
			coalesce(min_budget, 0),
			coalesce(max_budget, 0),
			coalesce(contract_types, '[]'::jsonb),
			coalesce(weekly_capacity_hours, 0)
		from freelancer_work_preferences
		where profile_id = $1
	`, profileID).Scan(&out.PreferredProjectLength, &out.MinBudgetUSD, &out.MaxBudgetUSD, &rawContractTypes, &out.WeeklyCapacityHours)
	if err != nil {
		if isNoRows(err) {
			return out, nil
		}
		return application.WorkPreferences{}, err
	}

	if len(rawContractTypes) > 0 {
		_ = json.Unmarshal(rawContractTypes, &out.ContractTypes)
	}
	out.PreferredProjectLength = application.CanonicalProjectLengthOrUnspecified(out.PreferredProjectLength)

	return out, nil
}

func (r *ProfileRepo) GetHiringPreferences(ctx context.Context, userID uuid.UUID) (application.HiringPreferences, error) {
	profileID, err := r.clientProfileID(ctx, userID)
	if err != nil {
		return application.HiringPreferences{}, err
	}

	out := application.HiringPreferences{}
	var rawLocations []byte
	err = r.pool.QueryRow(ctx, `
		select
			coalesce(min_hourly_rate, 0),
			coalesce(max_hourly_rate, 0),
			coalesce(preferred_locations, '[]'::jsonb)
		from client_hiring_preferences
		where profile_id = $1
	`, profileID).Scan(&out.MinHourlyRate, &out.MaxHourlyRate, &rawLocations)
	if err != nil {
		if isNoRows(err) {
			return out, nil
		}
		return application.HiringPreferences{}, err
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

	locations := make([]string, 0, len(in.PreferredLocations))
	seenLocations := make(map[string]struct{}, len(in.PreferredLocations))
	for _, location := range in.PreferredLocations {
		trimmed := strings.TrimSpace(location)
		if trimmed == "" {
			continue
		}
		if _, exists := seenLocations[trimmed]; exists {
			continue
		}
		seenLocations[trimmed] = struct{}{}
		locations = append(locations, trimmed)
	}

	rawLocations, err := json.Marshal(locations)
	if err != nil {
		return application.HiringPreferences{}, err
	}

	if _, err := r.pool.Exec(ctx, `
		insert into client_hiring_preferences (profile_id, min_hourly_rate, max_hourly_rate, preferred_locations)
		values ($1, $2, $3, $4::jsonb)
		on conflict (profile_id) do update set
			min_hourly_rate = excluded.min_hourly_rate,
			max_hourly_rate = excluded.max_hourly_rate,
			preferred_locations = excluded.preferred_locations,
			updated_at = now()
	`, profileID, in.MinHourlyRate, in.MaxHourlyRate, string(rawLocations)); err != nil {
		return application.HiringPreferences{}, err
	}

	return r.GetHiringPreferences(ctx, userID)
}

func (r *ProfileRepo) SaveFreelancer(ctx context.Context, userID uuid.UUID, freelancerUserID uuid.UUID) (application.SavedFreelancer, error) {
	profileID, err := r.clientProfileID(ctx, userID)
	if err != nil {
		return application.SavedFreelancer{}, err
	}
	if _, err := r.freelancerProfileID(ctx, freelancerUserID); err != nil {
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

	if pageSize == 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	limit := int(pageSize)

	token := strings.TrimSpace(pageToken)
	legacyOffsetMode := false
	offset := 0
	var cursorCreatedAt time.Time
	var cursorFreelancerID uuid.UUID
	if token != "" {
		if parsed, parseErr := strconv.Atoi(token); parseErr == nil && parsed >= 0 {
			legacyOffsetMode = true
			offset = parsed
		} else {
			cursorCreatedAt, cursorFreelancerID, err = decodeSavedFreelancersCursor(token)
			if err != nil {
				return application.ListResult[application.SavedFreelancer]{}, err
			}
		}
	}

	queryLimit := limit + 1
	var rows pgx.Rows
	if legacyOffsetMode {
		rows, err = r.pool.Query(ctx, `
			select freelancer_user_id, created_at
			from client_saved_freelancers
			where profile_id = $1
			order by created_at desc, freelancer_user_id asc
			limit $2 offset $3
		`, profileID, queryLimit, offset)
	} else if token == "" {
		rows, err = r.pool.Query(ctx, `
			select freelancer_user_id, created_at
			from client_saved_freelancers
			where profile_id = $1
			order by created_at desc, freelancer_user_id asc
			limit $2
		`, profileID, queryLimit)
	} else {
		rows, err = r.pool.Query(ctx, `
			select freelancer_user_id, created_at
			from client_saved_freelancers
			where profile_id = $1
			  and (
				created_at < $2
				or (created_at = $2 and freelancer_user_id > $3)
			  )
			order by created_at desc, freelancer_user_id asc
			limit $4
		`, profileID, cursorCreatedAt, cursorFreelancerID, queryLimit)
	}
	if err != nil {
		return application.ListResult[application.SavedFreelancer]{}, err
	}
	defer rows.Close()

	items := make([]application.SavedFreelancer, 0, queryLimit)
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

	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}

	nextPageToken := ""
	if hasMore {
		if legacyOffsetMode {
			nextPageToken = strconv.Itoa(offset + limit)
		} else {
			last := items[len(items)-1]
			nextPageToken = encodeSavedFreelancersCursor(last.SavedAt, last.FreelancerUserID)
		}
	}

	return application.ListResult[application.SavedFreelancer]{Items: items, NextPageToken: nextPageToken}, nil
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
		return false, nil
	}
	return true, nil
}

func (r *ProfileRepo) UpsertFreelancerNote(ctx context.Context, userID uuid.UUID, freelancerUserID uuid.UUID, note string) (application.FreelancerNote, error) {
	profileID, err := r.clientProfileID(ctx, userID)
	if err != nil {
		return application.FreelancerNote{}, err
	}
	if _, err := r.freelancerProfileID(ctx, freelancerUserID); err != nil {
		return application.FreelancerNote{}, err
	}
	normalizedNote := strings.TrimSpace(note)

	var updatedAt time.Time
	err = r.pool.QueryRow(ctx, `
		with saved as (
			select 1
			from client_saved_freelancers
			where profile_id = $1 and freelancer_user_id = $2
		), upserted as (
			insert into client_freelancer_notes (profile_id, freelancer_user_id, note)
			select $1, $2, $3
			from saved
			on conflict (profile_id, freelancer_user_id) do update set
				note = excluded.note,
				updated_at = now()
			returning updated_at
		)
		select updated_at from upserted
	`, profileID, freelancerUserID, normalizedNote).Scan(&updatedAt)
	if err != nil {
		if isNoRows(err) {
			return application.FreelancerNote{}, ErrNotFound
		}
		return application.FreelancerNote{}, err
	}

	return application.FreelancerNote{FreelancerUserID: freelancerUserID, Note: normalizedNote, UpdatedAt: updatedAt}, nil
}

func (r *ProfileRepo) ListDiscoverableFreelancers(ctx context.Context, filter application.DiscoverableFreelancerFilter, pageSize uint32, pageToken string) (application.ListResult[application.DiscoverableFreelancer], error) {
	limit, offset, err := parsePage(pageSize, pageToken)
	if err != nil {
		return application.ListResult[application.DiscoverableFreelancer]{}, err
	}

	whereClauses := []string{
		"p.role = 'freelancer'",
		"p.deleted_at is null",
	}
	args := []any{}
	argIdx := 1

	if filter.RequireActiveAccount {
		whereClauses = append(whereClauses, "coalesce(p.account_status, 'ACTIVE') = 'ACTIVE'")
	}
	if filter.RequireHeadline {
		whereClauses = append(whereClauses, "coalesce(fp.headline, '') <> ''")
	}
	if filter.MinSkills > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("coalesce(jsonb_array_length(fp.skills), 0) >= $%d", argIdx))
		args = append(args, filter.MinSkills)
		argIdx++
	}

	normalizedSkills := make([]string, 0, len(filter.Skills))
	seen := make(map[string]struct{}, len(filter.Skills))
	for _, skill := range filter.Skills {
		s := strings.ToLower(strings.TrimSpace(skill))
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		normalizedSkills = append(normalizedSkills, s)
	}
	if len(normalizedSkills) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf(
			"exists (select 1 from jsonb_array_elements_text(fp.skills) s where lower(s) = any($%d::text[]))",
			argIdx,
		))
		args = append(args, normalizedSkills)
		argIdx++
	}

	args = append(args, limit+1, offset)
	limitPos := argIdx
	offsetPos := argIdx + 1

	query := fmt.Sprintf(`
		select
			p.user_id,
			coalesce(fp.headline, ''),
			coalesce(p.bio, ''),
			coalesce(fp.skills, '[]'::jsonb),
			coalesce(fp.hourly_rate, 0),
			coalesce(fp.availability, 'AS_NEEDED'),
			coalesce(fp.rating, 0),
			coalesce(fp.total_reviews, 0),
			coalesce(p.location, ''),
			p.last_active_at
		from profiles p
		join freelancer_profiles fp on fp.profile_id = p.id
		where %s
		order by coalesce(fp.rating, 0) desc, p.updated_at desc, p.user_id asc
		limit $%d offset $%d
	`, strings.Join(whereClauses, " and "), limitPos, offsetPos)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return application.ListResult[application.DiscoverableFreelancer]{}, err
	}
	defer rows.Close()

	items := make([]application.DiscoverableFreelancer, 0, limit)
	for rows.Next() {
		var row application.DiscoverableFreelancer
		var skillsRaw []byte
		var lastActive *time.Time
		if err := rows.Scan(
			&row.UserID,
			&row.Headline,
			&row.Bio,
			&skillsRaw,
			&row.HourlyRate,
			&row.Availability,
			&row.Rating,
			&row.TotalReviews,
			&row.Location,
			&lastActive,
		); err != nil {
			return application.ListResult[application.DiscoverableFreelancer]{}, err
		}
		if len(skillsRaw) > 0 {
			_ = json.Unmarshal(skillsRaw, &row.Skills)
		}
		if lastActive != nil {
			utc := lastActive.UTC()
			row.LastActiveAt = &utc
		}
		items = append(items, row)
	}
	if err := rows.Err(); err != nil {
		return application.ListResult[application.DiscoverableFreelancer]{}, err
	}

	var nextToken string
	if len(items) > limit {
		items = items[:limit]
		nextToken = strconv.Itoa(offset + limit)
	}
	return application.ListResult[application.DiscoverableFreelancer]{Items: items, NextPageToken: nextToken}, nil
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
