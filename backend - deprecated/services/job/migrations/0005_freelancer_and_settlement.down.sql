alter table jobs
    drop column if exists cancellation_reason,
    drop column if exists settlement_policy,
    drop column if exists canceled_at,
    drop column if exists completed_at;

drop table if exists saved_jobs;

alter table job_invites
    drop constraint if exists job_invites_response_status_check;

alter table job_invites
    drop column if exists responded_at,
    drop column if exists response_status;
