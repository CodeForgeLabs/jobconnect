package grpcadapter

import (
	"context"
	"testing"
	"time"

	authv1 "jobconnect/auth/gen/auth/v1"
	"jobconnect/auth/internal/application"
	"jobconnect/auth/internal/domain"

	"github.com/google/uuid"
)

// Fakes shared across endpoint tests.

type fakeUserRepo struct {
	userByEmail  domain.User
	foundByEmail bool
	errByEmail   error
}

func (r *fakeUserRepo) Create(ctx context.Context, u domain.User) (domain.User, error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return u, nil
}

func (r *fakeUserRepo) GetByEmail(ctx context.Context, email string) (domain.User, bool, error) {
	return r.userByEmail, r.foundByEmail, r.errByEmail
}

func (r *fakeUserRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.User, bool, error) {
	return r.userByEmail, r.foundByEmail, r.errByEmail
}

func (r *fakeUserRepo) SetEmailVerified(ctx context.Context, userID uuid.UUID, at time.Time) error {
	return nil
}

type fakeCredRepo struct {
	hash  string
	found bool
	err   error
}

func (r *fakeCredRepo) Create(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	r.hash = passwordHash
	r.found = true
	return nil
}

func (r *fakeCredRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (string, bool, error) {
	return r.hash, r.found, r.err
}

func (r *fakeCredRepo) UpdatePasswordHash(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	r.hash = passwordHash
	r.found = true
	return nil
}

type fakeOTPRepo struct {
	consumeOK  bool
	consumeErr error
}

func (r *fakeOTPRepo) Create(ctx context.Context, email, purpose, otpHash string, expiresAt time.Time) error {
	return nil
}

func (r *fakeOTPRepo) Consume(ctx context.Context, email, purpose, otpPlain string, hasher domain.PasswordHasher) (bool, error) {
	return r.consumeOK, r.consumeErr
}

func (r *fakeOTPRepo) IncrementAttempts(ctx context.Context, email, purpose string) error {
	return nil
}

type fakeSessionRepo struct{}

func (r *fakeSessionRepo) Create(ctx context.Context, userID uuid.UUID, refreshTokenHash string, expiresAt time.Time) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (r *fakeSessionRepo) GetByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (bool, uuid.UUID, uuid.UUID, time.Time, bool, error) {
	return false, uuid.Nil, uuid.Nil, time.Time{}, false, nil
}

func (r *fakeSessionRepo) GetByID(ctx context.Context, sessionID uuid.UUID) (uuid.UUID, time.Time, bool, error) {
	return uuid.Nil, time.Time{}, false, nil
}

func (r *fakeSessionRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]application.SessionSummary, error) {
	return nil, nil
}

func (r *fakeSessionRepo) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (r *fakeSessionRepo) RevokeByID(ctx context.Context, sessionID uuid.UUID) error {
	return nil
}

func (r *fakeSessionRepo) UpdateLastUsed(ctx context.Context, sessionID uuid.UUID, at time.Time) error {
	return nil
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
	return "refresh-hash", nil
}

func (f *fakeTokens) ParseAccessToken(token string) (uuid.UUID, string, error) {
	return uuid.Nil, "", nil
}

type fakeClock struct{}

func (fakeClock) Now() time.Time { return time.Unix(0, 0).UTC() }

type fakeTOSRepo struct{}

func (fakeTOSRepo) Create(ctx context.Context, userID uuid.UUID, termsVersion, privacyVersion string) error {
	return nil
}

type fakeEmailSender struct{}

func (fakeEmailSender) SendVerifyEmailOTP(ctx context.Context, email, otp string) error {
	return nil
}

func (fakeEmailSender) SendPasswordResetOTP(ctx context.Context, email, otp string) error {
	return nil
}

// --- Tests ---

func TestAuthServer_Login_Success(t *testing.T) {
	user := domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  domain.RoleClient,
	}
	userRepo := &fakeUserRepo{userByEmail: user, foundByEmail: true}
	credRepo := &fakeCredRepo{hash: "hash", found: true}
	hasher := &fakeHasher{ok: true}
	sessions := &fakeSessionRepo{}
	tokens := &fakeTokens{}
	clock := fakeClock{}

	loginUC := &application.Login{
		Users:      userRepo,
		Creds:      credRepo,
		Sessions:   sessions,
		Hasher:     hasher,
		Tokens:     tokens,
		Clock:      clock,
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
	}

	srv := NewAuthServer(
		&application.RegisterUser{}, // not used here
		&application.VerifyEmailOTP{},
		loginUC,
		&application.Refresh{},
		&application.Logout{},
	)

	resp, err := srv.Login(context.Background(), &authv1.LoginRequest{
		Email:    "test@example.com",
		Password: "Password1!",
	})
	if err != nil {
		t.Fatalf("Login error: %v", err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Fatalf("expected non-empty tokens")
	}
	if resp.AccessTokenExpiresInSeconds <= 0 {
		t.Fatalf("expected positive expiry seconds")
	}
}

func TestAuthServer_Register_Success(t *testing.T) {
	userRepo := &fakeUserRepo{foundByEmail: false}
	credRepo := &fakeCredRepo{}
	otpRepo := &fakeOTPRepo{consumeOK: true}
	hasher := &fakeHasher{ok: true}
	clock := fakeClock{}
	tosRepo := &fakeTOSRepo{}
	emailSender := &fakeEmailSender{}

	registerUC := &application.RegisterUser{
		Users:          userRepo,
		Creds:          credRepo,
		OTPs:           otpRepo,
		TOS:            tosRepo,
		Hasher:         hasher,
		Clock:          clock,
		EmailSend:      emailSender,
		OTPTTL:         time.Minute,
		TOSVersion:     "1.0",
		PrivacyVersion: "1.0",
	}

	srv := NewAuthServer(
		registerUC,
		&application.VerifyEmailOTP{},
		&application.Login{},
		&application.Refresh{},
		&application.Logout{},
	)

	resp, err := srv.Register(context.Background(), &authv1.RegisterRequest{
		Email:       "test@example.com",
		Password:    "Password1!",
		FirstName:   "Test",
		LastName:    "User",
		Role:        domain.RoleClient,
		AcceptTerms: true,
	})
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}
	if resp.UserId == "" {
		t.Fatalf("expected non-empty user_id")
	}
}
