CREATE TABLE IF NOT EXISTS user_settings (
    profile_id BIGINT PRIMARY KEY REFERENCES profiles(id) ON DELETE CASCADE,
    ui_locale VARCHAR NOT NULL DEFAULT 'en',
    email_notifications_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    push_notifications_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_user_settings_ui_locale_nonempty CHECK (btrim(ui_locale) <> '')
);

INSERT INTO user_settings (
    profile_id,
    ui_locale,
    email_notifications_enabled,
    push_notifications_enabled,
    created_at,
    updated_at
)
SELECT
    p.id,
    COALESCE(NULLIF(btrim(p.language), ''), 'en'),
    TRUE,
    TRUE,
    NOW(),
    NOW()
FROM profiles p
ON CONFLICT (profile_id) DO UPDATE SET
    ui_locale = EXCLUDED.ui_locale,
    updated_at = NOW();

ALTER TABLE profiles
    DROP COLUMN IF EXISTS language;
