ALTER TABLE client_freelancer_notes
    DROP CONSTRAINT IF EXISTS fk_client_freelancer_notes_freelancer_user_id;

ALTER TABLE client_saved_freelancers
    DROP CONSTRAINT IF EXISTS fk_client_saved_freelancers_freelancer_user_id;

DROP INDEX IF EXISTS idx_client_freelancer_notes_freelancer_user_id;
DROP INDEX IF EXISTS idx_client_saved_freelancers_freelancer_user_id;
