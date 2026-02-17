-- Add first_name and last_name to users.

alter table users add column if not exists first_name text;
alter table users add column if not exists last_name text;

-- Backfill existing rows.
update users set first_name = display_name where (first_name is null or first_name = '');
update users set last_name = '' where last_name is null;

alter table users alter column first_name set not null;
alter table users alter column last_name set not null;
