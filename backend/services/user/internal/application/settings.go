package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type UserSettings struct {
	UILocale                  string
	EmailNotificationsEnabled bool
	PushNotificationsEnabled  bool
}

type PatchSettings struct {
	UILocale                  *string
	EmailNotificationsEnabled *bool
	PushNotificationsEnabled  *bool
}

type GetSettingsInput struct {
	UserID uuid.UUID
}

type GetSettingsOutput struct {
	Settings UserSettings
}

type GetSettings struct {
	Settings SettingsRepository
}

func (uc *GetSettings) Execute(ctx context.Context, in GetSettingsInput) (GetSettingsOutput, error) {
	if in.UserID == uuid.Nil {
		return GetSettingsOutput{}, fmt.Errorf("user_id is required")
	}
	settings, err := uc.Settings.GetSettingsByUserID(ctx, in.UserID)
	if err != nil {
		return GetSettingsOutput{}, err
	}
	return GetSettingsOutput{Settings: settings}, nil
}

type PatchSettingsInput struct {
	UserID uuid.UUID
	Patch  PatchSettings
}

type PatchSettingsOutput struct {
	Settings UserSettings
}

type PatchSettingsUseCase struct {
	Settings SettingsRepository
}

func (uc *PatchSettingsUseCase) Execute(ctx context.Context, in PatchSettingsInput) (PatchSettingsOutput, error) {
	if in.UserID == uuid.Nil {
		return PatchSettingsOutput{}, fmt.Errorf("user_id is required")
	}
	if in.Patch.UILocale == nil && in.Patch.EmailNotificationsEnabled == nil && in.Patch.PushNotificationsEnabled == nil {
		return PatchSettingsOutput{}, fmt.Errorf("at least one updatable setting is required")
	}
	if in.Patch.UILocale != nil {
		locale := strings.TrimSpace(*in.Patch.UILocale)
		if locale == "" {
			return PatchSettingsOutput{}, fmt.Errorf("ui_locale cannot be empty")
		}
		in.Patch.UILocale = &locale
	}

	settings, err := uc.Settings.PatchSettingsByUserID(ctx, in.UserID, in.Patch)
	if err != nil {
		return PatchSettingsOutput{}, err
	}
	return PatchSettingsOutput{Settings: settings}, nil
}
