drop index if exists idx_job_attachments_storage_key;

alter table if exists job_attachments
    drop column if exists storage_key;
