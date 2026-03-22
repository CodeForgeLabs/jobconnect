package storage

import (
	"context"
	"strings"
	"testing"

	"jobconnect/user/internal/domain"

	"github.com/google/uuid"
)

func TestPutAvatarValidatesInputs(t *testing.T) {
	store := &AvatarStore{}

	err := store.PutAvatar(context.Background(), domain.AvatarObject{})
	if err == nil || !strings.Contains(err.Error(), "user_id") {
		t.Fatalf("expected user_id validation error, got %v", err)
	}

	err = store.PutAvatar(context.Background(), domain.AvatarObject{UserID: uuid.New()})
	if err == nil || !strings.Contains(err.Error(), "storage_key") {
		t.Fatalf("expected storage_key validation error, got %v", err)
	}

	err = store.PutAvatar(context.Background(), domain.AvatarObject{UserID: uuid.New(), StorageKey: "avatars/x.png"})
	if err == nil || !strings.Contains(err.Error(), "content") {
		t.Fatalf("expected content validation error, got %v", err)
	}
}

func TestGetAvatarValidatesStorageKey(t *testing.T) {
	store := &AvatarStore{}

	_, err := store.GetAvatar(context.Background(), uuid.New(), "")
	if err == nil || !strings.Contains(err.Error(), "storage_key") {
		t.Fatalf("expected storage_key validation error, got %v", err)
	}
}

func TestDeleteAvatarValidatesStorageKey(t *testing.T) {
	store := &AvatarStore{}

	err := store.DeleteAvatar(context.Background(), uuid.New(), "")
	if err == nil || !strings.Contains(err.Error(), "storage_key") {
		t.Fatalf("expected storage_key validation error, got %v", err)
	}
}
