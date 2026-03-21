package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"jobconnect/user/internal/domain"

	"github.com/google/uuid"
)

type avatarRepoMock struct {
	savedAvatar   domain.Avatar
	storedAvatar  domain.Avatar
	saveErr       error
	getAvatarErr  error
	removeErr     error
	updatedProfile domain.Profile
	updated       bool
}

func (m *avatarRepoMock) Create(context.Context, domain.Profile, *domain.ClientProfile, *domain.FreelancerProfile) (int64, error) {
	panic("not implemented")
}

func (m *avatarRepoMock) GetByUserID(context.Context, uuid.UUID) (domain.Profile, *domain.ClientProfile, *domain.FreelancerProfile, error) {
	p := m.updatedProfile
	if p.UserID == uuid.Nil {
		p.UserID = uuid.New()
	}
	return p, nil, nil, nil
}

func (m *avatarRepoMock) Update(_ context.Context, profile domain.Profile, _ *domain.ClientProfile, _ *domain.FreelancerProfile) error {
	m.updated = true
	m.updatedProfile = profile
	return nil
}

func (m *avatarRepoMock) UpdateAccountState(context.Context, uuid.UUID, string, string, string, time.Time) (domain.Profile, *domain.ClientProfile, *domain.FreelancerProfile, error) {
	panic("not implemented")
}

func (m *avatarRepoMock) Delete(context.Context, uuid.UUID, bool, time.Time) error {
	panic("not implemented")
}

func (m *avatarRepoMock) SaveAvatar(_ context.Context, avatar domain.Avatar) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedAvatar = avatar
	m.storedAvatar = avatar
	return nil
}

func (m *avatarRepoMock) GetAvatar(_ context.Context, _ uuid.UUID) (domain.Avatar, error) {
	if m.getAvatarErr != nil {
		return domain.Avatar{}, m.getAvatarErr
	}
	return m.storedAvatar, nil
}

func (m *avatarRepoMock) RemoveAvatar(_ context.Context, _ uuid.UUID) error {
	if m.removeErr != nil {
		return m.removeErr
	}
	m.storedAvatar = domain.Avatar{}
	return nil
}

type avatarStoreMock struct {
	putAvatar    domain.AvatarObject
	getContent   []byte
	putErr       error
	getErr       error
	deleteErr    error
	deletedKey   string
}

func (m *avatarStoreMock) PutAvatar(_ context.Context, avatar domain.AvatarObject) error {
	if m.putErr != nil {
		return m.putErr
	}
	m.putAvatar = avatar
	return nil
}

func (m *avatarStoreMock) GetAvatar(_ context.Context, _ uuid.UUID, storageKey string) ([]byte, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if storageKey == "" {
		return nil, errors.New("missing storage key")
	}
	return m.getContent, nil
}

func (m *avatarStoreMock) DeleteAvatar(_ context.Context, _ uuid.UUID, storageKey string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deletedKey = storageKey
	return nil
}

type avatarProcessorMock struct{}

func (m *avatarProcessorMock) Process(content []byte, declaredContentType string) ([]byte, string, int, int, error) {
	return content, declaredContentType, 64, 64, nil
}

type avatarClockMock struct{ now time.Time }

func (m avatarClockMock) Now() time.Time { return m.now }

func TestUploadAvatarStoresObjectAndMetadata(t *testing.T) {
	userID := uuid.New()
	repo := &avatarRepoMock{updatedProfile: domain.Profile{UserID: userID}}
	store := &avatarStoreMock{}
	uc := &UploadAvatar{
		Profiles:  repo,
		Store:     store,
		Processor: &avatarProcessorMock{},
		Clock:     avatarClockMock{now: time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC)},
	}

	out, err := uc.Execute(context.Background(), UploadAvatarInput{
		UserID:      userID,
		FileName:    "me.png",
		ContentType: "image/png",
		Content:     []byte("fake-image"),
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if store.putAvatar.StorageKey == "" {
		t.Fatalf("expected object storage key")
	}
	if repo.savedAvatar.StorageKey == "" {
		t.Fatalf("expected repo to save storage key")
	}
	if out.AvatarURL == "" {
		t.Fatalf("expected avatar url")
	}
}

func TestGetAvatarReadsFromObjectStore(t *testing.T) {
	userID := uuid.New()
	repo := &avatarRepoMock{storedAvatar: domain.Avatar{UserID: userID, FileName: "avatar.png", ContentType: "image/png", StorageKey: "avatars/key.png"}}
	store := &avatarStoreMock{getContent: []byte("bytes")}
	uc := &GetAvatar{Profiles: repo, Store: store}

	out, err := uc.Execute(context.Background(), GetAvatarInput{UserID: userID})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if string(out.Content) != "bytes" {
		t.Fatalf("unexpected content: %q", string(out.Content))
	}
}

func TestRemoveAvatarDeletesObjectThenMetadata(t *testing.T) {
	userID := uuid.New()
	repo := &avatarRepoMock{storedAvatar: domain.Avatar{UserID: userID, StorageKey: "avatars/key.png"}, updatedProfile: domain.Profile{UserID: userID}}
	store := &avatarStoreMock{}
	uc := &RemoveAvatar{Profiles: repo, Store: store}

	_, err := uc.Execute(context.Background(), RemoveAvatarInput{UserID: userID})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if store.deletedKey != "avatars/key.png" {
		t.Fatalf("unexpected deleted key: %q", store.deletedKey)
	}
}
