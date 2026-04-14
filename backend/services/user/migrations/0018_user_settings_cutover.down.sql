ALTER TABLE profiles
    ADD COLUMN IF NOT EXISTS language VARCHAR;

UPDATE profiles p
SET language = us.ui_locale
FROM user_settings us
WHERE us.profile_id = p.id
  AND us.ui_locale IS NOT NULL
  AND btrim(us.ui_locale) <> '';

DROP TABLE IF EXISTS user_settings;
