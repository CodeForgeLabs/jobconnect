alter table proposal_attachments
    add column if not exists storage_key text not null default '';

create index if not exists idx_proposal_attachments_storage_key on proposal_attachments(storage_key);
