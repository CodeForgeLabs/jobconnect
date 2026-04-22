alter table proposals
    drop constraint if exists proposals_status_check;

alter table proposals
    add column if not exists offer_sent_at timestamptz null;

alter table proposals
    add constraint proposals_status_check
    check (status in ('sent', 'shortlisted', 'rejected', 'offer_sent', 'hired', 'withdrawn'));

drop index if exists uq_proposals_active_per_job_freelancer;

create unique index if not exists uq_proposals_active_per_job_freelancer
    on proposals(job_id, freelancer_id)
    where status in ('sent', 'shortlisted', 'offer_sent', 'hired');
