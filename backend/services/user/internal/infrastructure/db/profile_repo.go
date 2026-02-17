package db

import (
	"context"
	"encoding/json"

	"jobconnect/user/internal/domain"

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
		insert into profiles (user_id, role, first_name, last_name, display_name, avatar_url, created_at)
		values ($1, $2, $3, $4, $5, $6, $7)
		returning id
	`, profile.UserID, profile.Role, profile.FirstName, profile.LastName, profile.DisplayName, profile.AvatarURL, profile.CreatedAt).Scan(&profileID)
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
