alter table jobs
    add column if not exists visibility text not null default 'public',
    add column if not exists experience_level text not null default 'intermediate',
    add column if not exists budget_min double precision not null default 0,
    add column if not exists budget_max double precision not null default 0,
    add column if not exists paused_at timestamptz null,
    add column if not exists filled_at timestamptz null;

alter table jobs
    drop constraint if exists jobs_status_check;

alter table jobs
    add constraint jobs_status_check check (status in ('open', 'paused', 'filled', 'closed'));

alter table jobs
    drop constraint if exists jobs_visibility_check;

alter table jobs
    add constraint jobs_visibility_check check (visibility in ('public', 'private', 'invite_only'));

alter table jobs
    drop constraint if exists jobs_experience_level_check;

alter table jobs
    add constraint jobs_experience_level_check check (experience_level in ('entry', 'intermediate', 'expert'));
