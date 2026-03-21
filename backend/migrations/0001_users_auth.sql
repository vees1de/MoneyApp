begin;

create table if not exists users (
  id uuid primary key,
  email text null,
  display_name text null,
  avatar_url text null,
  timezone text not null default 'Europe/Amsterdam',
  base_currency text not null default 'RUB',
  onboarding_completed boolean not null default false,
  weekly_review_weekday smallint not null default 1,
  weekly_review_hour smallint not null default 18,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create unique index if not exists users_email_lower_uidx
  on users (lower(email))
  where email is not null;

create table if not exists auth_identities (
  id uuid primary key,
  user_id uuid not null references users(id) on delete cascade,
  provider text not null,
  provider_user_id text not null,
  provider_email text null,
  access_meta jsonb not null default '{}'::jsonb,
  created_at timestamptz not null,
  unique (provider, provider_user_id)
);

create table if not exists sessions (
  id uuid primary key,
  user_id uuid not null references users(id) on delete cascade,
  refresh_token_hash text not null unique,
  user_agent text null,
  ip_address text null,
  expires_at timestamptz not null,
  created_at timestamptz not null,
  revoked_at timestamptz null
);

create index if not exists sessions_user_id_idx on sessions (user_id);
create index if not exists sessions_expires_at_idx on sessions (expires_at);

create table if not exists audit_logs (
  id uuid primary key,
  user_id uuid not null references users(id) on delete cascade,
  action text not null,
  entity_type text not null,
  entity_id uuid null,
  meta jsonb not null default '{}'::jsonb,
  created_at timestamptz not null
);

create index if not exists audit_logs_user_id_created_at_idx on audit_logs (user_id, created_at desc);

commit;
