alter table jobs
    drop constraint if exists jobs_status_check;

alter table jobs
    add constraint jobs_status_check
    check (status in ('open', 'paused', 'filled', 'closed', 'completed', 'canceled'));

alter table jobs
    drop constraint if exists jobs_experience_level_check;

alter table jobs
    drop column if exists experience_level;
