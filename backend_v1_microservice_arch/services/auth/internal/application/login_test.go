package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"jobconnect/auth/internal/domain"

	"github.com/google/uuid"
)

// Fakes for Login dependencies.

type fakeUserRepo struct {
	user  domain.User
	found bool
	err   error
}

func (r *fakeUserRepo) Create(ctx context.Context, u domain.User) (domain.User, error) {
	return domain.User{}, errors.New("not implemented")
}

func (r *fakeUserRepo) GetByEmail(ctx context.Context, email string) (domain.User, bool, error) {
	return r.user, r.found, r.err
}

func (r *fakeUserRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.User, bool, error) {
	return domain.User{}, false, errors.New("not implemented")
}

func (r *fakeUserRepo) SetEmailVerified(ctx context.Context, userID uuid.UUID, at time.Time) error {
	return errors.New("not implemented")
}

func (r *fakeUserRepo) UpdateEmail(ctx context.Context, userID uuid.UUID, newEmail string, at time.Time) error {
	return errors.New("not implemented")
}

type fakeCredRepo struct {
	hash  string
	found bool
	err   error
}

func (r *fakeCredRepo) Create(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	return errors.New("not implemented")
}

func (r *fakeCredRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (string, bool, error) {
	return r.hash, r.found, r.err
}

func (r *fakeCredRepo) UpdatePasswordHash(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	return errors.New("not implemented")
}

type fakeSessionRepo struct{}

func (r *fakeSessionRepo) Create(ctx context.Context, userID uuid.UUID, refreshTokenHash string, expiresAt time.Time) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (r *fakeSessionRepo) GetByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (bool, uuid.UUID, uuid.UUID, time.Time, bool, error) {
	return false, uuid.Nil, uuid.Nil, time.Time{}, false, errors.New("not implemented")
}

func (r *fakeSessionRepo) GetByID(ctx context.Context, sessionID uuid.UUID) (uuid.UUID, time.Time, bool, error) {
	return uuid.Nil, time.Time{}, false, errors.New("not implemented")
}

func (r *fakeSessionRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]SessionSummary, error) {
	return nil, errors.New("not implemented")
}

func (r *fakeSessionRepo) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	return errors.New("not implemented")
}

func (r *fakeSessionRepo) RevokeByID(ctx context.Context, sessionID uuid.UUID) error {
	return errors.New("not implemented")
}

func (r *fakeSessionRepo) UpdateLastUsed(ctx context.Context, sessionID uuid.UUID, at time.Time) error {
	return errors.New("not implemented")
}

type fakeHasher struct {
	ok  bool
	err error
}

func (h *fakeHasher) Hash(password string) (string, error) {
	return "hash", nil
}

func (h *fakeHasher) Verify(password, hash string) (bool, error) {
	return h.ok, h.err
}

type fakeTokens struct{}

func (f *fakeTokens) IssueAccessToken(userID uuid.UUID, role string, expiresIn time.Duration) (string, error) {
	return "access", nil
}

func (f *fakeTokens) IssueRefreshToken() (string, string, error) {
	return "refresh", "refresh-hash", nil
}

func (f *fakeTokens) HashRefreshToken(token string) (string, error) {
	return "hash", nil
}

func (f *fakeTokens) ParseAccessToken(token string) (uuid.UUID, string, error) {
	return uuid.Nil, "", nil
}

type fakeClock struct{}

func (fakeClock) Now() time.Time { return time.Unix(0, 0).UTC() }

func TestLogin_Success(t *testing.T) {
	ctx := context.Background()
	user := domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  "client",
	}
	uc := &Login{
		Users:    &fakeUserRepo{user: user, found: true},
		Creds:    &fakeCredRepo{hash: "hash", found: true},
		Sessions: &fakeSessionRepo{},
		Hasher:   &fakeHasher{ok: true},
		Tokens:   &fakeTokens{},
		Clock:    fakeClock{},
		// TTLs don't matter for test, but avoid zero.
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
	}
	out, err := uc.Execute(ctx, LoginInput{
		Email:    "test@example.com",
		Password: "Password1!",
	})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if out.AccessToken == "" || out.RefreshToken == "" {
		t.Fatalf("expected non-empty tokens")
	}
	if out.ExpiresInSec <= 0 {
		t.Fatalf("expected positive ExpiresInSec")
	}
}
