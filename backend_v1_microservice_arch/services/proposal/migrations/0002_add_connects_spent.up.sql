alter table proposals
    add column if not exists connects_spent integer not null default 0;
