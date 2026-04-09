package application

import (
	"context"
	"testing"
	"time"

	"jobconnect/user/internal/domain"

	"github.com/google/uuid"
)

type cvRepoMock struct {
	storedCV     CV
	saveCalled   bool
	getCalled    bool
	removeCalled bool
	saveErr      error
	getErr       error
	removeErr    error
}

func (m *cvRepoMock) SaveCV(_ context.Context, cv CV) error {
	m.saveCalled = true
	if m.saveErr != nil {
		return m.saveErr
	}
	m.storedCV = cv
	return nil
}

func (m *cvRepoMock) GetCV(_ context.Context, _ uuid.UUID) (CV, error) {
	m.getCalled = true
	if m.getErr != nil {
		return CV{}, m.getErr
	}
	return m.storedCV, nil
}

func (m *cvRepoMock) RemoveCV(_ context.Context, _ uuid.UUID) error {
	m.removeCalled = true
	if m.removeErr != nil {
		return m.removeErr
	}
	m.storedCV = CV{}
	return nil
}

type cvRoleRepoMock struct {
	profile domain.Profile
	err     error
}

func (m *cvRoleRepoMock) GetByUserID(_ context.Context, userID uuid.UUID) (domain.Profile, *domain.ClientProfile, *domain.FreelancerProfile, error) {
	if m.err != nil {
		return domain.Profile{}, nil, nil, m.err
	}
	p := m.profile
	if p.UserID == uuid.Nil {
		p.UserID = userID
	}
	return p, nil, nil, nil
}

type cvStoreMock struct {
	putCalled    bool
	deleteCalled bool
	deletedKey   string
	putErr       error
	deleteErr    error
	presignErr   error
	presignURL   string
}

func (m *cvStoreMock) PutCV(_ context.Context, _ domain.CVObject) error {
	m.putCalled = true
	if m.putErr != nil {
		return m.putErr
	}
	return nil
}

func (m *cvStoreMock) DeleteCV(_ context.Context, _ uuid.UUID, storageKey string) error {
	m.deleteCalled = true
	m.deletedKey = storageKey
	if m.deleteErr != nil {
		return m.deleteErr
	}
	return nil
}

func (m *cvStoreMock) PresignGetObject(_ context.Context, _ string, _ time.Duration) (string, error) {
	if m.presignErr != nil {
		return "", m.presignErr
	}
	if m.presignURL == "" {
		return "https://example.test/cv", nil
	}
	return m.presignURL, nil
}

func TestUpsertCVRejectsNonFreelancer(t *testing.T) {
	userID := uuid.New()
	repo := &cvRepoMock{}
	store := &cvStoreMock{}
	roles := &cvRoleRepoMock{profile: domain.Profile{UserID: userID, Role: domain.RoleClient}}
	uc := &UpsertCV{Profiles: repo, RoleProfiles: roles, Store: store}

	_, err := uc.Execute(context.Background(), UpsertCVInput{
		UserID:      userID,
		FileName:    "resume.pdf",
		ContentType: "application/pdf",
		Content:     []byte("pdf-content"),
	})
	if err == nil || err.Error() != "freelancer role required" {
		t.Fatalf("expected freelancer role error, got %v", err)
	}
	if store.putCalled {
		t.Fatalf("expected no object store writes")
	}
	if repo.saveCalled {
		t.Fatalf("expected no cv metadata writes")
	}
}

func TestGetCVRejectsNonFreelancer(t *testing.T) {
	userID := uuid.New()
	repo := &cvRepoMock{}
	store := &cvStoreMock{}
	roles := &cvRoleRepoMock{profile: domain.Profile{UserID: userID, Role: domain.RoleClient}}
	uc := &GetCV{Profiles: repo, RoleProfiles: roles, Store: store}

	_, err := uc.Execute(context.Background(), GetCVInput{UserID: userID})
	if err == nil || err.Error() != "freelancer role required" {
		t.Fatalf("expected freelancer role error, got %v", err)
	}
	if repo.getCalled {
		t.Fatalf("expected no cv metadata reads")
	}
}

func TestRemoveCVRejectsNonFreelancer(t *testing.T) {
	userID := uuid.New()
	repo := &cvRepoMock{}
	store := &cvStoreMock{}
	roles := &cvRoleRepoMock{profile: domain.Profile{UserID: userID, Role: domain.RoleClient}}
	uc := &RemoveCV{Profiles: repo, RoleProfiles: roles, Store: store}

	_, err := uc.Execute(context.Background(), RemoveCVInput{UserID: userID})
	if err == nil || err.Error() != "freelancer role required" {
		t.Fatalf("expected freelancer role error, got %v", err)
	}
	if repo.getCalled || repo.removeCalled || store.deleteCalled {
		t.Fatalf("expected no cv delete flow for non-freelancer")
	}
}

func TestUpsertCVAllowsFreelancer(t *testing.T) {
	userID := uuid.New()
	repo := &cvRepoMock{}
	store := &cvStoreMock{}
	roles := &cvRoleRepoMock{profile: domain.Profile{UserID: userID, Role: domain.RoleFreelancer}}
	uc := &UpsertCV{Profiles: repo, RoleProfiles: roles, Store: store}

	out, err := uc.Execute(context.Background(), UpsertCVInput{
		UserID:      userID,
		FileName:    "resume",
		ContentType: "application/pdf",
		Content:     []byte("pdf-content"),
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !store.putCalled || !repo.saveCalled {
		t.Fatalf("expected cv object and metadata to be stored")
	}
	if out.CV.FileName != "resume.pdf" {
		t.Fatalf("unexpected file name: %q", out.CV.FileName)
	}
	if out.DownloadURL == "" {
		t.Fatalf("expected download URL")
	}
}
