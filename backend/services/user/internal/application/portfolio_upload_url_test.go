package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

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
	uc := &GetPortfolioMediaUploadURL{Store: store}

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
	uc := &GetPortfolioMediaUploadURL{Store: &portfolioObjectStoreMock{}}

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
}

func TestGetPortfolioMediaUploadURLPropagatesStoreError(t *testing.T) {
	uc := &GetPortfolioMediaUploadURL{
		Store: &portfolioObjectStoreMock{presignPutErr: errors.New("boom")},
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
