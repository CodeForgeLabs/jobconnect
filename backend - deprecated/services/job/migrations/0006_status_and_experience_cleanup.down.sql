alter table jobs
    add column if not exists experience_level text not null default 'intermediate';

alter table jobs
    drop constraint if exists jobs_experience_level_check;

alter table jobs
    add constraint jobs_experience_level_check
    check (experience_level in ('entry', 'intermediate', 'expert'));

alter table jobs
    drop constraint if exists jobs_status_check;

alter table jobs
    add constraint jobs_status_check
    check (status in ('open', 'paused', 'filled', 'closed'));
