package db

import (
	"context"

	"jobconnect/auth/internal/application"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OAuthIdentityRepo struct {
	pool *pgxpool.Pool
}

func NewOAuthIdentityRepo(pool *pgxpool.Pool) *OAuthIdentityRepo {
	return &OAuthIdentityRepo{pool: pool}
}

func (r *OAuthIdentityRepo) GetByProviderUserID(ctx context.Context, provider, providerUserID string) (application.OAuthIdentity, bool, error) {
	var userID uuid.UUID
	var email string
	err := r.pool.QueryRow(ctx, `
		select user_id, email
		from oauth_identities
		where provider = $1 and provider_user_id = $2
	`, provider, providerUserID).Scan(&userID, &email)
	if err != nil {
		if isNoRows(err) {
			return application.OAuthIdentity{}, false, nil
		}
		return application.OAuthIdentity{}, false, err
	}
	return application.OAuthIdentity{
		UserID:         userID,
		Provider:       provider,
		ProviderUserID: providerUserID,
		Email:          email,
	}, true, nil
}

func (r *OAuthIdentityRepo) Create(ctx context.Context, identity application.OAuthIdentity) error {
	_, err := r.pool.Exec(ctx, `
		insert into oauth_identities (user_id, provider, provider_user_id, email)
		values ($1, $2, $3, $4)
		on conflict (provider, provider_user_id) do update
		set user_id = excluded.user_id,
		    email = excluded.email
	`, identity.UserID, identity.Provider, identity.ProviderUserID, identity.Email)
	return err
}
