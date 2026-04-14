package application

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type settingsRepoStub struct {
	settings UserSettings
	patch    PatchSettings
}

func (s *settingsRepoStub) GetSettingsByUserID(ctx context.Context, userID uuid.UUID) (UserSettings, error) {
	return s.settings, nil
}

func (s *settingsRepoStub) PatchSettingsByUserID(ctx context.Context, userID uuid.UUID, patch PatchSettings) (UserSettings, error) {
	s.patch = patch
	out := s.settings
	if patch.UILocale != nil {
		out.UILocale = *patch.UILocale
	}
	if patch.EmailNotificationsEnabled != nil {
		out.EmailNotificationsEnabled = *patch.EmailNotificationsEnabled
	}
	if patch.PushNotificationsEnabled != nil {
		out.PushNotificationsEnabled = *patch.PushNotificationsEnabled
	}
	return out, nil
}

func TestPatchSettingsTrimsLocale(t *testing.T) {
	repo := &settingsRepoStub{settings: UserSettings{UILocale: "en", EmailNotificationsEnabled: true, PushNotificationsEnabled: true}}
	uc := &PatchSettingsUseCase{Settings: repo}
	userID := uuid.New()
	raw := "  fr  "

	out, err := uc.Execute(context.Background(), PatchSettingsInput{UserID: userID, Patch: PatchSettings{UILocale: &raw}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Settings.UILocale != "fr" {
		t.Fatalf("expected trimmed locale 'fr', got %q", out.Settings.UILocale)
	}
}
