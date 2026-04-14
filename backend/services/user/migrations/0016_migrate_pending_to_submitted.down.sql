-- Roll back submitted status values to pending for compatibility.
UPDATE profiles
SET verification_status = 'PENDING'
WHERE upper(trim(coalesce(verification_status, ''))) IN (
    'SUBMITTED',
    'VERIFICATION_STATUS_SUBMITTED'
);
