create table if not exists email_change_requests (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null unique references users(id) on delete cascade,
  new_email text not null,
  otp_hash text not null,
  expires_at timestamptz not null,
  attempts int not null default 0,
  confirmed_at timestamptz,
  created_at timestamptz not null default now()
);

create index if not exists email_change_requests_expires_idx on email_change_requests(expires_at);

create table if not exists oauth_identities (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references users(id) on delete cascade,
  provider text not null,
  provider_user_id text not null,
  email text not null,
  created_at timestamptz not null default now(),
  unique(provider, provider_user_id)
);

create index if not exists oauth_identities_user_id_idx on oauth_identities(user_id);
