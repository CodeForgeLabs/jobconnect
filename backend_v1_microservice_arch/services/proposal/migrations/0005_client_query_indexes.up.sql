create index if not exists idx_proposals_client_created_at
    on proposals(client_id, created_at desc, id desc);

create index if not exists idx_proposals_client_status_created_at
    on proposals(client_id, status, created_at desc, id desc);

create index if not exists idx_proposals_client_job_created_at
    on proposals(client_id, job_id, created_at desc, id desc);

create index if not exists idx_proposals_client_job_status_created_at
    on proposals(client_id, job_id, status, created_at desc, id desc);

create index if not exists idx_proposals_client_freelancer_created_at
    on proposals(client_id, freelancer_id, created_at desc, id desc);

create index if not exists idx_proposals_client_freelancer_status_created_at
    on proposals(client_id, freelancer_id, status, created_at desc, id desc);