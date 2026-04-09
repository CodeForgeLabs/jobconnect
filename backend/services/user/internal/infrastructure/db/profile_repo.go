package db

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"jobconnect/user/internal/application"
	"jobconnect/user/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProfileRepo implements application.ProfileRepository.
type ProfileRepo struct {
	pool *pgxpool.Pool
}

func NewProfileRepo(pool *pgxpool.Pool) *ProfileRepo {
	return &ProfileRepo{pool: pool}
}

func (r *ProfileRepo) Create(ctx context.Context, profile domain.Profile, client *domain.ClientProfile, freelancer *domain.FreelancerProfile) (int64, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var profileID int64
	var taxID *string
	if strings.TrimSpace(profile.TaxID) != "" {
		v := strings.TrimSpace(profile.TaxID)
		taxID = &v
	}
	var verificationStatus *string
	if strings.TrimSpace(profile.VerificationStatus) != "" {
		v := strings.TrimSpace(profile.VerificationStatus)
		verificationStatus = &v
	}

	err = tx.QueryRow(ctx, `
		insert into profiles (user_id, role, first_name, last_name, display_name, avatar_url, contact_email, contact_phone, bio, location, account_status, suspension_reason, tax_id, verification_status, last_active_at, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		returning id
	`, profile.UserID, profile.Role, profile.FirstName, profile.LastName, profile.DisplayName, profile.AvatarURL, profile.ContactEmail, profile.ContactPhone, profile.Bio, profile.Location, profile.AccountStatus, profile.SuspensionReason, taxID, verificationStatus, profile.LastActiveAt, profile.CreatedAt, profile.UpdatedAt).Scan(&profileID)
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(ctx, `
		insert into user_settings (profile_id, ui_locale, email_notifications_enabled, push_notifications_enabled)
		values ($1, 'en', true, true)
		on conflict (profile_id) do nothing
	`, profileID)
	if err != nil {
		return 0, err
	}

	if client != nil {
		_, err = tx.Exec(ctx, `
			insert into client_profiles (profile_id, company_name)
			values ($1, $2)
		`, profileID, client.CompanyName)
		if err != nil {
			return 0, err
		}
	}

	if freelancer != nil {
		skillsJSON, err := json.Marshal(freelancer.Skills)
		if err != nil {
			return 0, err
		}
		_, err = tx.Exec(ctx, `
			insert into freelancer_profiles (
				profile_id,
				headline,
				skills,
				rating,
				job_success_score,
				total_reviews,
				total_jobs,
				total_earnings,
				hourly_rate,
				availability
			)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, profileID, freelancer.Headline, skillsJSON, freelancer.Rating, freelancer.Reputation.JobSuccessScore, freelancer.Reputation.TotalReviews, freelancer.Reputation.TotalJobs, freelancer.Reputation.TotalEarningsUSD, freelancer.HourlyRate, freelancer.Availability)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return profileID, nil
}

func (r *ProfileRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (domain.Profile, *domain.ClientProfile, *domain.FreelancerProfile, error) {
	var profile domain.Profile
	var deletedAt *time.Time
	var profileTaxID string
	var profileVerificationStatus string
	err := r.pool.QueryRow(ctx, `
		select id,
			user_id,
			role,
			coalesce(first_name, ''),
			coalesce(last_name, ''),
			coalesce(display_name, ''),
			coalesce(avatar_url, ''),
			coalesce(contact_email, ''),
			coalesce(contact_phone, ''),
			coalesce(bio, ''),
			coalesce(location, ''),
			coalesce(account_status, 'ACTIVE'),
			coalesce(suspension_reason, ''),
			coalesce(tax_id, ''),
			coalesce(verification_status, ''),
			last_active_at,
			created_at,
			updated_at,
			deleted_at
		from profiles
		where user_id = $1
	`, userID).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.Role,
		&profile.FirstName,
		&profile.LastName,
		&profile.DisplayName,
		&profile.AvatarURL,
		&profile.ContactEmail,
		&profile.ContactPhone,
		&profile.Bio,
		&profile.Location,
		&profile.AccountStatus,
		&profile.SuspensionReason,
		&profileTaxID,
		&profileVerificationStatus,
		&profile.LastActiveAt,
		&profile.CreatedAt,
		&profile.UpdatedAt,
		&deletedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return domain.Profile{}, nil, nil, ErrNotFound
		}
		return domain.Profile{}, nil, nil, err
	}
	profile.DeletedAt = deletedAt
	profile.TaxID = profileTaxID
	profile.VerificationStatus = profileVerificationStatus

	var client *domain.ClientProfile
	if profile.Role == domain.RoleClient {
		cp := &domain.ClientProfile{}
		err = r.pool.QueryRow(ctx, `
			select coalesce(company_name, '')
			from client_profiles
			where profile_id = $1
		`, profile.ID).Scan(&cp.CompanyName)
		if err != nil {
			if !isNoRows(err) {
				return domain.Profile{}, nil, nil, err
			}
		} else {
			cp.TaxID = profile.TaxID
			cp.VerificationStatus = profile.VerificationStatus
			client = cp
		}
	}

	var freelancer *domain.FreelancerProfile
	if profile.Role == domain.RoleFreelancer {
		fp := &domain.FreelancerProfile{}
		var skillsRaw []byte
		err = r.pool.QueryRow(ctx, `
			select coalesce(headline, ''),
				coalesce(skills, '[]'::jsonb),
				coalesce(rating, 0),
				coalesce(job_success_score, 0),
				coalesce(total_reviews, 0),
				coalesce(total_jobs, 0),
				coalesce(total_earnings, 0),
				coalesce(hourly_rate, 0),
				coalesce(availability, 'AS_NEEDED')
			from freelancer_profiles
			where profile_id = $1
		`, profile.ID).Scan(&fp.Headline, &skillsRaw, &fp.Rating, &fp.Reputation.JobSuccessScore, &fp.Reputation.TotalReviews, &fp.Reputation.TotalJobs, &fp.Reputation.TotalEarningsUSD, &fp.HourlyRate, &fp.Availability)
		if err != nil {
			if !isNoRows(err) {
				return domain.Profile{}, nil, nil, err
			}
		} else {
			if len(skillsRaw) > 0 {
				_ = json.Unmarshal(skillsRaw, &fp.Skills)
			}
			fp.VerificationStatus = profile.VerificationStatus
			fp.Reputation.AvgRating = fp.Rating
			freelancer = fp
		}
	}

	return profile, client, freelancer, nil
}

func (r *ProfileRepo) Update(ctx context.Context, profile domain.Profile, client *domain.ClientProfile, freelancer *domain.FreelancerProfile) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	profile.UpdatedAt = time.Now().UTC()
	taxID := &profile.TaxID
	verificationStatus := &profile.VerificationStatus
	_, err = tx.Exec(ctx, `
		update profiles
		set first_name = $2,
			last_name = $3,
			display_name = $4,
			avatar_url = $5,
			contact_email = $6,
			contact_phone = $7,
			bio = $8,
			location = $9,
			account_status = $10,
			suspension_reason = $11,
			last_active_at = $12,
			tax_id = coalesce($13, tax_id),
			verification_status = coalesce($14, verification_status),
			updated_at = $15
		where user_id = $1
	`, profile.UserID, profile.FirstName, profile.LastName, profile.DisplayName, profile.AvatarURL, profile.ContactEmail, profile.ContactPhone, profile.Bio, profile.Location, profile.AccountStatus, profile.SuspensionReason, profile.LastActiveAt, taxID, verificationStatus, profile.UpdatedAt)
	if err != nil {
		return err
	}

	if profile.Role == domain.RoleClient && client != nil {
		_, err = tx.Exec(ctx, `
			insert into client_profiles (profile_id, company_name)
			values ($1, $2)
			on conflict (profile_id) do update set
				company_name = excluded.company_name
		`, profile.ID, client.CompanyName)
		if err != nil {
			return err
		}
	}

	if profile.Role == domain.RoleFreelancer && freelancer != nil {
		skillsJSON, err := json.Marshal(freelancer.Skills)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
			insert into freelancer_profiles (
				profile_id,
				headline,
				skills,
				rating,
				job_success_score,
				total_reviews,
				total_jobs,
				total_earnings,
				hourly_rate,
				availability
			)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			on conflict (profile_id) do update set
				headline = excluded.headline,
				skills = excluded.skills,
				rating = excluded.rating,
				job_success_score = excluded.job_success_score,
				total_reviews = excluded.total_reviews,
				total_jobs = excluded.total_jobs,
				total_earnings = excluded.total_earnings,
				hourly_rate = excluded.hourly_rate,
				availability = excluded.availability
		`, profile.ID, freelancer.Headline, skillsJSON, freelancer.Rating, freelancer.Reputation.JobSuccessScore, freelancer.Reputation.TotalReviews, freelancer.Reputation.TotalJobs, freelancer.Reputation.TotalEarningsUSD, freelancer.HourlyRate, freelancer.Availability)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r *ProfileRepo) Delete(ctx context.Context, userID uuid.UUID, hardDelete bool, deletedAt time.Time) error {
	if hardDelete {
		res, err := r.pool.Exec(ctx, `delete from profiles where user_id = $1`, userID)
		if err != nil {
			return err
		}
		if res.RowsAffected() == 0 {
			return ErrNotFound
		}
		_, _ = r.pool.Exec(ctx, `delete from profile_avatars where user_id = $1`, userID)
		return nil
	}

	res, err := r.pool.Exec(ctx, `
		update profiles
		set deleted_at = $2, updated_at = $2
		where user_id = $1
	`, userID, deletedAt)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ProfileRepo) SaveAvatar(ctx context.Context, avatar domain.Avatar) error {
	_, err := r.pool.Exec(ctx, `
		insert into profile_avatars (user_id, file_name, content_type, storage_key, width, height, size_bytes, updated_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8)
		on conflict (user_id) do update set
			file_name = excluded.file_name,
			content_type = excluded.content_type,
			storage_key = excluded.storage_key,
			width = excluded.width,
			height = excluded.height,
			size_bytes = excluded.size_bytes,
			updated_at = excluded.updated_at
	`, avatar.UserID, avatar.FileName, avatar.ContentType, avatar.StorageKey, avatar.Width, avatar.Height, avatar.SizeBytes, avatar.UpdatedAt)
	return err
}

func (r *ProfileRepo) GetAvatar(ctx context.Context, userID uuid.UUID) (domain.Avatar, error) {
	var avatar domain.Avatar
	err := r.pool.QueryRow(ctx, `
		select user_id, file_name, content_type, storage_key, width, height, size_bytes, updated_at
		from profile_avatars
		where user_id = $1
	`, userID).Scan(&avatar.UserID, &avatar.FileName, &avatar.ContentType, &avatar.StorageKey, &avatar.Width, &avatar.Height, &avatar.SizeBytes, &avatar.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.Avatar{}, ErrNotFound
		}
		return domain.Avatar{}, err
	}
	return avatar, nil
}

func (r *ProfileRepo) RemoveAvatar(ctx context.Context, userID uuid.UUID) error {
	res, err := r.pool.Exec(ctx, `delete from profile_avatars where user_id = $1`, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ProfileRepo) SaveCV(ctx context.Context, cv application.CV) error {
	_, err := r.pool.Exec(ctx, `
		insert into profile_cvs (user_id, file_name, content_type, storage_key, size_bytes, updated_at)
		values ($1, $2, $3, $4, $5, $6)
		on conflict (user_id) do update set
			file_name = excluded.file_name,
			content_type = excluded.content_type,
			storage_key = excluded.storage_key,
			size_bytes = excluded.size_bytes,
			updated_at = excluded.updated_at
	`, cv.UserID, cv.FileName, cv.ContentType, cv.StorageKey, cv.SizeBytes, cv.UpdatedAt)
	return err
}

func (r *ProfileRepo) GetCV(ctx context.Context, userID uuid.UUID) (application.CV, error) {
	var cv application.CV
	err := r.pool.QueryRow(ctx, `
		select user_id, file_name, content_type, storage_key, size_bytes, updated_at
		from profile_cvs
		where user_id = $1
	`, userID).Scan(&cv.UserID, &cv.FileName, &cv.ContentType, &cv.StorageKey, &cv.SizeBytes, &cv.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return application.CV{}, ErrNotFound
		}
		return application.CV{}, err
	}
	return cv, nil
}

func (r *ProfileRepo) RemoveCV(ctx context.Context, userID uuid.UUID) error {
	res, err := r.pool.Exec(ctx, `delete from profile_cvs where user_id = $1`, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ProfileRepo) GetSettingsByUserID(ctx context.Context, userID uuid.UUID) (application.UserSettings, error) {
	var out application.UserSettings
	err := r.pool.QueryRow(ctx, `
		select
			coalesce(us.ui_locale, 'en'),
			coalesce(us.email_notifications_enabled, true),
			coalesce(us.push_notifications_enabled, true)
		from profiles p
		left join user_settings us on us.profile_id = p.id
		where p.user_id = $1
	`, userID).Scan(&out.UILocale, &out.EmailNotificationsEnabled, &out.PushNotificationsEnabled)
	if err != nil {
		if isNoRows(err) {
			return application.UserSettings{}, ErrNotFound
		}
		return application.UserSettings{}, err
	}

	if strings.TrimSpace(out.UILocale) == "" {
		out.UILocale = "en"
	}
	return out, nil
}

func (r *ProfileRepo) PatchSettingsByUserID(ctx context.Context, userID uuid.UUID, patch application.PatchSettings) (application.UserSettings, error) {
	current, err := r.GetSettingsByUserID(ctx, userID)
	if err != nil {
		return application.UserSettings{}, err
	}

	if patch.UILocale != nil {
		current.UILocale = strings.TrimSpace(*patch.UILocale)
	}
	if patch.EmailNotificationsEnabled != nil {
		current.EmailNotificationsEnabled = *patch.EmailNotificationsEnabled
	}
	if patch.PushNotificationsEnabled != nil {
		current.PushNotificationsEnabled = *patch.PushNotificationsEnabled
	}

	if strings.TrimSpace(current.UILocale) == "" {
		current.UILocale = "en"
	}

	_, err = r.pool.Exec(ctx, `
		insert into user_settings (profile_id, ui_locale, email_notifications_enabled, push_notifications_enabled, created_at, updated_at)
		select p.id, $2, $3, $4, now(), now()
		from profiles p
		where p.user_id = $1
		on conflict (profile_id) do update set
			ui_locale = excluded.ui_locale,
			email_notifications_enabled = excluded.email_notifications_enabled,
			push_notifications_enabled = excluded.push_notifications_enabled,
			updated_at = now()
	`, userID, current.UILocale, current.EmailNotificationsEnabled, current.PushNotificationsEnabled)
	if err != nil {
		return application.UserSettings{}, err
	}

	return current, nil
}
