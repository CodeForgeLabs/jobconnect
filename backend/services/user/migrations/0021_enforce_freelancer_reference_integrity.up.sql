-- Clean orphaned or invalid freelancer references before adding foreign keys.
DELETE FROM client_freelancer_notes n
WHERE NOT EXISTS (
    SELECT 1
    FROM profiles p
    WHERE p.user_id = n.freelancer_user_id
      AND p.deleted_at IS NULL
      AND p.role = 'freelancer'
);

DELETE FROM client_saved_freelancers s
WHERE NOT EXISTS (
    SELECT 1
    FROM profiles p
    WHERE p.user_id = s.freelancer_user_id
      AND p.deleted_at IS NULL
      AND p.role = 'freelancer'
);

CREATE INDEX IF NOT EXISTS idx_client_saved_freelancers_freelancer_user_id
    ON client_saved_freelancers(freelancer_user_id);

CREATE INDEX IF NOT EXISTS idx_client_freelancer_notes_freelancer_user_id
    ON client_freelancer_notes(freelancer_user_id);

ALTER TABLE client_saved_freelancers
    DROP CONSTRAINT IF EXISTS fk_client_saved_freelancers_freelancer_user_id,
    ADD CONSTRAINT fk_client_saved_freelancers_freelancer_user_id
        FOREIGN KEY (freelancer_user_id) REFERENCES profiles(user_id) ON DELETE CASCADE;

ALTER TABLE client_freelancer_notes
    DROP CONSTRAINT IF EXISTS fk_client_freelancer_notes_freelancer_user_id,
    ADD CONSTRAINT fk_client_freelancer_notes_freelancer_user_id
        FOREIGN KEY (freelancer_user_id) REFERENCES profiles(user_id) ON DELETE CASCADE;
