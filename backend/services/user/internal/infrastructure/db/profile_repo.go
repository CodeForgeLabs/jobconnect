package db

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"jobconnect/user/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	err = tx.QueryRow(ctx, `
		insert into profiles (user_id, role, first_name, last_name, display_name, avatar_url, language, contact_email, contact_phone, bio, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		returning id
	`, profile.UserID, profile.Role, profile.FirstName, profile.LastName, profile.DisplayName, profile.AvatarURL, profile.Language, profile.ContactEmail, profile.ContactPhone, profile.Bio, profile.CreatedAt, profile.UpdatedAt).Scan(&profileID)
	if err != nil {
		return 0, err
	}
	if client != nil {
		_, err = tx.Exec(ctx, `
			insert into client_profiles (profile_id, company_name, billing_address, tax_id, verification_status)
			values ($1, $2, $3, $4, $5)
		`, profileID, client.CompanyName, client.BillingAddress, client.TaxID, client.VerificationStatus)
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
			insert into freelancer_profiles (profile_id, headline, bio, skills, experience_level, rating, verification_status)
			values ($1, $2, $3, $4, $5, $6, $7)
		`, profileID, freelancer.Headline, freelancer.Bio, skillsJSON, freelancer.ExperienceLevel, freelancer.Rating, freelancer.VerificationStatus)
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
	err := r.pool.QueryRow(ctx, `
		select id,
			user_id,
			role,
			coalesce(first_name, ''),
			coalesce(last_name, ''),
			coalesce(display_name, ''),
			coalesce(avatar_url, ''),
			coalesce(language, ''),
			coalesce(contact_email, ''),
			coalesce(contact_phone, ''),
			coalesce(bio, ''),
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
		&profile.Language,
		&profile.ContactEmail,
		&profile.ContactPhone,
		&profile.Bio,
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

	var client *domain.ClientProfile
	if profile.Role == domain.RoleClient {
		cp := &domain.ClientProfile{}
		err = r.pool.QueryRow(ctx, `
			select coalesce(company_name, ''),
				coalesce(billing_address, ''),
				coalesce(tax_id, ''),
				coalesce(verification_status, '')
			from client_profiles
			where profile_id = $1
		`, profile.ID).Scan(&cp.CompanyName, &cp.BillingAddress, &cp.TaxID, &cp.VerificationStatus)
		if err != nil {
			if !isNoRows(err) {
				return domain.Profile{}, nil, nil, err
			}
		} else {
			client = cp
		}
	}

	var freelancer *domain.FreelancerProfile
	if profile.Role == domain.RoleFreelancer {
		fp := &domain.FreelancerProfile{}
		var skillsRaw []byte
		err = r.pool.QueryRow(ctx, `
			select coalesce(headline, ''),
				coalesce(bio, ''),
				coalesce(skills, '[]'::jsonb),
				coalesce(experience_level, ''),
				coalesce(rating, 0),
				coalesce(verification_status, '')
			from freelancer_profiles
			where profile_id = $1
		`, profile.ID).Scan(&fp.Headline, &fp.Bio, &skillsRaw, &fp.ExperienceLevel, &fp.Rating, &fp.VerificationStatus)
		if err != nil {
			if !isNoRows(err) {
				return domain.Profile{}, nil, nil, err
			}
		} else {
			if len(skillsRaw) > 0 {
				_ = json.Unmarshal(skillsRaw, &fp.Skills)
			}
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
	_, err = tx.Exec(ctx, `
		update profiles
		set first_name = $2,
			last_name = $3,
			display_name = $4,
			avatar_url = $5,
			language = $6,
			contact_email = $7,
			contact_phone = $8,
			bio = $9,
			updated_at = $10
		where user_id = $1
	`, profile.UserID, profile.FirstName, profile.LastName, profile.DisplayName, profile.AvatarURL, profile.Language, profile.ContactEmail, profile.ContactPhone, profile.Bio, profile.UpdatedAt)
	if err != nil {
		return err
	}

	if profile.Role == domain.RoleClient && client != nil {
		_, err = tx.Exec(ctx, `
			insert into client_profiles (profile_id, company_name, billing_address, tax_id, verification_status)
			values ($1, $2, $3, $4, $5)
			on conflict (profile_id) do update set
				company_name = excluded.company_name,
				billing_address = excluded.billing_address,
				tax_id = excluded.tax_id,
				verification_status = excluded.verification_status
		`, profile.ID, client.CompanyName, client.BillingAddress, client.TaxID, client.VerificationStatus)
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
			insert into freelancer_profiles (profile_id, headline, bio, skills, experience_level, rating, verification_status)
			values ($1, $2, $3, $4, $5, $6, $7)
			on conflict (profile_id) do update set
				headline = excluded.headline,
				bio = excluded.bio,
				skills = excluded.skills,
				experience_level = excluded.experience_level,
				rating = excluded.rating,
				verification_status = excluded.verification_status
		`, profile.ID, freelancer.Headline, freelancer.Bio, skillsJSON, freelancer.ExperienceLevel, freelancer.Rating, freelancer.VerificationStatus)
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
		insert into profile_avatars (user_id, file_name, content_type, content, width, height, size_bytes, updated_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8)
		on conflict (user_id) do update set
			file_name = excluded.file_name,
			content_type = excluded.content_type,
			content = excluded.content,
			width = excluded.width,
			height = excluded.height,
			size_bytes = excluded.size_bytes,
			updated_at = excluded.updated_at
	`, avatar.UserID, avatar.FileName, avatar.ContentType, avatar.Content, avatar.Width, avatar.Height, avatar.SizeBytes, avatar.UpdatedAt)
	return err
}

func (r *ProfileRepo) GetAvatar(ctx context.Context, userID uuid.UUID) (domain.Avatar, error) {
	var avatar domain.Avatar
	err := r.pool.QueryRow(ctx, `
		select user_id, file_name, content_type, content, width, height, size_bytes, updated_at
		from profile_avatars
		where user_id = $1
	`, userID).Scan(&avatar.UserID, &avatar.FileName, &avatar.ContentType, &avatar.Content, &avatar.Width, &avatar.Height, &avatar.SizeBytes, &avatar.UpdatedAt)
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

func isNoRowsOrNotFound(err error) bool {
	return isNoRows(err) || errors.Is(err, pgx.ErrNoRows)
}
