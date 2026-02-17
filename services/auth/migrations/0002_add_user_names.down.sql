-- Remove first_name and last_name from users.

alter table users drop column if exists first_name;
alter table users drop column if exists last_name;
