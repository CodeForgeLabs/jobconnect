package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"jobconnect/user/internal/domain"

	"github.com/google/uuid"
)

type portfolioRoleRepoMock struct {
	profile domain.Profile
	err     error
}

func (m *portfolioRoleRepoMock) GetByUserID(context.Context, uuid.UUID) (domain.Profile, *domain.ClientProfile, *domain.FreelancerProfile, error) {
	if m.err != nil {
		return domain.Profile{}, nil, nil, m.err
	}
	if m.profile.UserID == uuid.Nil {
		m.profile.UserID = uuid.New()
	}
	if m.profile.Role == "" {
		m.profile.Role = domain.RoleFreelancer
	}
	return m.profile, nil, nil, nil
}

type portfolioObjectStoreMock struct {
	presignPutKey         string
	presignPutContentType string
	presignPutTTL         time.Duration
	presignPutURL         string
	presignPutErr         error
}

func (m *portfolioObjectStoreMock) PutObject(context.Context, string, []byte, string) error {
	return nil
}

func (m *portfolioObjectStoreMock) GetObject(context.Context, string) ([]byte, error) {
	return nil, nil
}

func (m *portfolioObjectStoreMock) DeleteObject(context.Context, string) error {
	return nil
}

func (m *portfolioObjectStoreMock) PresignGetObject(context.Context, string, time.Duration) (string, error) {
	return "", nil
}

func (m *portfolioObjectStoreMock) PresignPutObject(_ context.Context, storageKey string, contentType string, ttl time.Duration) (string, error) {
	if m.presignPutErr != nil {
		return "", m.presignPutErr
	}
	m.presignPutKey = storageKey
	m.presignPutContentType = contentType
	m.presignPutTTL = ttl
	if m.presignPutURL == "" {
		return "https://example.test/upload", nil
	}
	return m.presignPutURL, nil
}

func TestGetPortfolioMediaUploadURLSuccess(t *testing.T) {
	userID := uuid.New()
	store := &portfolioObjectStoreMock{}
	uc := &GetPortfolioMediaUploadURL{Store: store, RoleProfiles: &portfolioRoleRepoMock{profile: domain.Profile{UserID: userID, Role: domain.RoleFreelancer}}}

	out, err := uc.Execute(context.Background(), GetPortfolioMediaUploadURLInput{
		UserID:      userID,
		FileName:    "design.png",
		ContentType: "IMAGE/PNG",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if out.StorageKey == "" || !strings.HasPrefix(out.StorageKey, "portfolio/"+userID.String()+"/") {
		t.Fatalf("expected portfolio storage key prefix, got %q", out.StorageKey)
	}
	if !strings.HasSuffix(out.StorageKey, ".png") {
		t.Fatalf("expected storage key with png extension, got %q", out.StorageKey)
	}
	if store.presignPutKey != out.StorageKey {
		t.Fatalf("expected presign key %q, got %q", out.StorageKey, store.presignPutKey)
	}
	if store.presignPutContentType != "image/png" {
		t.Fatalf("expected normalized content-type, got %q", store.presignPutContentType)
	}
	if store.presignPutTTL != 15*time.Minute {
		t.Fatalf("expected default ttl 15m, got %s", store.presignPutTTL)
	}
}

func TestGetPortfolioMediaUploadURLValidation(t *testing.T) {
	uc := &GetPortfolioMediaUploadURL{Store: &portfolioObjectStoreMock{}, RoleProfiles: &portfolioRoleRepoMock{profile: domain.Profile{Role: domain.RoleFreelancer}}}

	_, err := uc.Execute(context.Background(), GetPortfolioMediaUploadURLInput{})
	if err == nil || !strings.Contains(err.Error(), "user_id is required") {
		t.Fatalf("expected user_id error, got %v", err)
	}

	_, err = uc.Execute(context.Background(), GetPortfolioMediaUploadURLInput{UserID: uuid.New(), ContentType: "image/png"})
	if err == nil || !strings.Contains(err.Error(), "file_name is required") {
		t.Fatalf("expected file_name error, got %v", err)
	}

	_, err = uc.Execute(context.Background(), GetPortfolioMediaUploadURLInput{UserID: uuid.New(), FileName: "a.png"})
	if err == nil || !strings.Contains(err.Error(), "content_type is required") {
		t.Fatalf("expected content_type error, got %v", err)
	}

	_, err = uc.Execute(context.Background(), GetPortfolioMediaUploadURLInput{UserID: uuid.New(), FileName: "a.exe", ContentType: "application/x-msdownload"})
	if err == nil || !strings.Contains(err.Error(), "unsupported portfolio content_type") {
		t.Fatalf("expected portfolio content_type error, got %v", err)
	}
}

func TestGetPortfolioMediaUploadURLPropagatesStoreError(t *testing.T) {
	uc := &GetPortfolioMediaUploadURL{
		Store:        &portfolioObjectStoreMock{presignPutErr: errors.New("boom")},
		RoleProfiles: &portfolioRoleRepoMock{profile: domain.Profile{Role: domain.RoleFreelancer}},
	}

	_, err := uc.Execute(context.Background(), GetPortfolioMediaUploadURLInput{
		UserID:      uuid.New(),
		FileName:    "design.png",
		ContentType: "image/png",
	})
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected propagated store error, got %v", err)
	}
}

func TestGetPortfolioMediaUploadURLRequiresFreelancerRole(t *testing.T) {
	uc := &GetPortfolioMediaUploadURL{
		Store:        &portfolioObjectStoreMock{},
		RoleProfiles: &portfolioRoleRepoMock{profile: domain.Profile{Role: domain.RoleClient}},
	}

	_, err := uc.Execute(context.Background(), GetPortfolioMediaUploadURLInput{
		UserID:      uuid.New(),
		FileName:    "design.png",
		ContentType: "image/png",
	})
	if err == nil || !strings.Contains(err.Error(), "freelancer role required") {
		t.Fatalf("expected freelancer role required error, got %v", err)
	}
}
