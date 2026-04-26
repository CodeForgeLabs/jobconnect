alter table contracts
    drop constraint if exists contracts_status_check;

alter table contracts
    add constraint contracts_status_check
        check (status in ('pending_acceptance', 'active', 'declined', 'paused', 'ended'));

alter table contracts
    drop column if exists revoked_at;
