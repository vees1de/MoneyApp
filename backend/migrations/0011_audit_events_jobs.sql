begin;

create table if not exists audit_logs (
  id uuid primary key,
  actor_user_id uuid null references users(id),
  user_id uuid null references users(id),
  entity_type varchar(50) not null,
  entity_id uuid null,
  action varchar(50) not null,
  old_values jsonb null,
  new_values jsonb null,
  meta jsonb not null default '{}'::jsonb,
  source text not null default 'system',
  request_id text null,
  session_id uuid null,
  change_set jsonb not null default '{}'::jsonb,
  actor_type text not null default 'user',
  actor_id uuid null,
  ip inet null,
  user_agent text null,
  created_at timestamptz not null
);

create table if not exists domain_events (
  id uuid primary key,
  event_type varchar(100) not null,
  entity_type varchar(50) not null,
  entity_id uuid not null,
  payload jsonb not null,
  occurred_at timestamptz not null,
  processed_at timestamptz null
);

create table if not exists outbox_messages (
  id uuid primary key,
  topic varchar(100) not null,
  event_type varchar(100) not null,
  entity_type varchar(50) not null,
  entity_id uuid not null,
  payload jsonb not null,
  status varchar(30) not null check (status in ('pending', 'processing', 'processed', 'failed')),
  available_at timestamptz not null,
  processed_at timestamptz null,
  attempt int not null default 0,
  last_error text null,
  created_at timestamptz not null
);

create table if not exists background_jobs (
  id uuid primary key,
  queue varchar(100) not null,
  job_type varchar(100) not null,
  payload jsonb not null,
  idempotency_key varchar(255) null,
  status varchar(30) not null check (status in ('pending', 'processing', 'completed', 'failed', 'retry')),
  run_at timestamptz not null,
  attempt int not null default 0,
  max_attempts int not null default 5,
  locked_at timestamptz null,
  locked_by varchar(100) null,
  completed_at timestamptz null,
  last_error text null,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  unique nulls not distinct (queue, idempotency_key)
);

commit;
