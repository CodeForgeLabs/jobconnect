-- Align legacy profile verification status values to the new submitted state.
UPDATE profiles
SET verification_status = 'SUBMITTED'
WHERE upper(trim(coalesce(verification_status, ''))) IN (
    'PENDING',
    'VERIFICATION_STATUS_PENDING'
);
