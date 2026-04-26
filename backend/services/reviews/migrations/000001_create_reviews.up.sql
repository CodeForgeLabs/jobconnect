create table if not exists reviews (
    id bigserial primary key,

    contract_id bigint not null,

    client_id uuid not null,
    freelancer_id uuid not null,

    reviewer_role text not null check (reviewer_role in ('client','freelancer')),

    rating integer not null check (rating between 1 and 5),

    title text not null,
    comment text not null default '',


    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),

    constraint uq_review_per_contract_role unique (contract_id, reviewer_role)
);

create index if not exists idx_reviews_client on reviews(client_id);
create index if not exists idx_reviews_freelancer on reviews(freelancer_id);
create index if not exists idx_reviews_contract on reviews(contract_id);