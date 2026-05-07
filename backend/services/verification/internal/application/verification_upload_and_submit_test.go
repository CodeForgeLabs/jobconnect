package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"jobconnect/verification/internal/domain"

	"github.com/google/uuid"
)

type testClock struct {
	now time.Time
}

func (c testClock) Now() time.Time { return c.now }

type verificationRepoStub struct {
	latest   domain.VerificationRequest
	hasLatest bool
	created  domain.VerificationRequest
}

func (s *verificationRepoStub) CreateSubmission(_ context.Context, req domain.VerificationRequest) (domain.VerificationRequest, error) {
	req.ID = 1
	s.created = req
	return req, nil
}
func (s *verificationRepoStub) GetLatestByUserID(_ context.Context, _ uuid.UUID) (domain.VerificationRequest, error) {
	if s.hasLatest {
		return s.latest, nil
	}
	return domain.VerificationRequest{}, errors.New("not found")
}
func (s *verificationRepoStub) GetByID(context.Context, int64) (domain.VerificationRequest, error) {
	return domain.VerificationRequest{}, errors.New("not found")
}
func (s *verificationRepoStub) ListPending(context.Context, int32, int32) ([]domain.VerificationRequest, error) {
	return nil, nil
}
func (s *verificationRepoStub) Review(context.Context, int64, uuid.UUID, string, string, string, time.Time) (domain.VerificationRequest, error) {
	return domain.VerificationRequest{}, nil
}
func (s *verificationRepoStub) MarkReverificationRequired(context.Context, uuid.UUID, uuid.UUID, string, time.Time, time.Time) (domain.VerificationRequest, error) {
	return domain.VerificationRequest{}, nil
}
func (s *verificationRepoStub) AppendEvent(context.Context, domain.VerificationEvent) error { return nil }

type evidenceStoreStub struct {
	key string
	url string
}

func (s *evidenceStoreStub) BuildObjectKey(_ uuid.UUID, _ string) string { return s.key }
func (s *evidenceStoreStub) PresignPutObject(context.Context, string, time.Duration) (string, error) {
	return s.url, nil
}

func TestSubmitVerification_RequiresEvidenceURL(t *testing.T) {
	uc := &SubmitVerification{
		Repo:  &verificationRepoStub{},
		Clock: testClock{now: time.Unix(1700000000, 0).UTC()},
	}
	_, err := uc.Execute(context.Background(), SubmitVerificationInput{
		UserID:               uuid.New(),
		LegalName:            "Test User",
		CountryCode:          "US",
		DocumentType:         "ID_CARD",
		DocumentNumberMasked: "****1234",
		EvidenceURL:          "",
	})
	if err == nil || !strings.Contains(err.Error(), "evidence_url is required") {
		t.Fatalf("expected evidence_url validation error, got %v", err)
	}
}

func TestSubmitVerification_PersistsEvidenceURL(t *testing.T) {
	repo := &verificationRepoStub{}
	uc := &SubmitVerification{
		Repo:  repo,
		Clock: testClock{now: time.Unix(1700000000, 0).UTC()},
	}
	evidenceURL := "https://example.test/verification/id-front.jpg"
	_, err := uc.Execute(context.Background(), SubmitVerificationInput{
		UserID:               uuid.New(),
		LegalName:            "Test User",
		CountryCode:          "US",
		DocumentType:         "ID_CARD",
		DocumentNumberMasked: "****1234",
		EvidenceURL:          evidenceURL,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.created.EvidenceURL != evidenceURL {
		t.Fatalf("expected evidence_url %q, got %q", evidenceURL, repo.created.EvidenceURL)
	}
}

func TestGetVerificationEvidenceUploadURL_ReturnsPresignedURL(t *testing.T) {
	store := &evidenceStoreStub{
		key: "verifications/user/file.jpg",
		url: "https://upload.test/file.jpg",
	}
	uc := &GetVerificationEvidenceUploadURL{
		Store:  store,
		PutTTL: 15 * time.Minute,
	}
	out, err := uc.Execute(context.Background(), GetVerificationEvidenceUploadURLInput{
		UserID:      uuid.New(),
		FileName:    "file.jpg",
		ContentType: "image/jpeg",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.StorageKey != store.key || out.UploadURL != store.url {
		t.Fatalf("unexpected output: %+v", out)
	}
}
