alter table jobs
    drop constraint if exists jobs_experience_level_check;

alter table jobs
    drop constraint if exists jobs_visibility_check;

alter table jobs
    drop constraint if exists jobs_status_check;

alter table jobs
    add constraint jobs_status_check check (status in ('open', 'closed'));

alter table jobs
    drop column if exists filled_at,
    drop column if exists paused_at,
    drop column if exists budget_max,
    drop column if exists budget_min,
    drop column if exists experience_level,
    drop column if exists visibility;
