alter table contracts
    add column if not exists revoked_at timestamptz null;

do $$
declare
    status_constraint text;
begin
    select con.conname
    into status_constraint
    from pg_constraint con
    join pg_class rel on rel.oid = con.conrelid
    where rel.relname = 'contracts'
      and con.contype = 'c'
      and pg_get_constraintdef(con.oid) like '%status in (''pending_acceptance'', ''active'', ''declined'', ''paused'', ''ended'')%';

    if status_constraint is not null then
        execute format('alter table contracts drop constraint %I', status_constraint);
    end if;
end $$;

alter table contracts
    add constraint contracts_status_check
        check (status in ('pending_acceptance', 'active', 'declined', 'revoked', 'paused', 'ended'));
