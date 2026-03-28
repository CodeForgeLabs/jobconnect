alter table if exists job_attachments
    add column if not exists storage_key text not null default '';

create index if not exists idx_job_attachments_storage_key
    on job_attachments(storage_key);
