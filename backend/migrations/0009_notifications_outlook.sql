begin;

create table if not exists outlook_accounts (
  id uuid primary key,
  user_id uuid not null unique references users(id) on delete cascade,
  external_account_id varchar(255) not null,
  email varchar(255) not null,
  access_token_encrypted text not null,
  refresh_token_encrypted text not null,
  token_expires_at timestamptz not null,
  scope text null,
  status varchar(30) not null check (status in ('active', 'expired', 'revoked', 'error')),
  last_sync_at timestamptz null,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create table if not exists calendar_events (
  id uuid primary key,
  user_id uuid not null references users(id) on delete cascade,
  source_type varchar(50) not null check (source_type in ('external_request', 'internal_session', 'deadline_reminder')),
  source_id uuid not null,
  provider varchar(30) not null check (provider in ('outlook', 'system')),
  external_event_id varchar(255) null,
  title varchar(255) not null,
  start_at timestamptz not null,
  end_at timestamptz not null,
  timezone varchar(100) null,
  status varchar(30) not null check (status in ('scheduled', 'updated', 'canceled', 'sync_error')),
  meeting_url text null,
  location text null,
  payload jsonb null,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create table if not exists integration_jobs (
  id uuid primary key,
  integration_type varchar(50) not null check (integration_type in ('outlook_calendar_sync')),
  entity_type varchar(50) not null,
  entity_id uuid not null,
  job_type varchar(50) not null check (job_type in ('create_event', 'update_event', 'delete_event', 'refresh_token', 'pull_changes')),
  status varchar(30) not null check (status in ('pending', 'processing', 'done', 'failed', 'retry')),
  attempt int not null default 0,
  max_attempts int not null default 5,
  next_retry_at timestamptz null,
  last_error text null,
  payload jsonb null,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create table if not exists calendar_conflicts (
  id uuid primary key,
  user_id uuid not null references users(id) on delete cascade,
  source_type varchar(50) not null,
  source_id uuid not null,
  conflict_type varchar(30) not null check (conflict_type in ('overlap', 'ooo', 'tentative', 'busy_slot')),
  conflict_with jsonb not null,
  detected_at timestamptz not null,
  resolved_at timestamptz null,
  status varchar(30) not null check (status in ('active', 'ignored', 'resolved'))
);

create table if not exists notification_templates (
  id uuid primary key,
  code varchar(100) not null unique,
  channel varchar(30) not null check (channel in ('email', 'in_app')),
  subject_template text null,
  body_template text not null,
  is_active boolean not null default true,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create table if not exists notifications (
  id uuid primary key,
  user_id uuid not null references users(id) on delete cascade,
  channel varchar(30) not null check (channel in ('email', 'in_app')),
  type varchar(50) not null,
  title varchar(255) not null,
  body text not null,
  status varchar(30) not null check (status in ('pending', 'sent', 'delivered', 'read', 'failed')),
  related_entity_type varchar(50) null,
  related_entity_id uuid null,
  scheduled_at timestamptz null,
  sent_at timestamptz null,
  read_at timestamptz null,
  created_at timestamptz not null
);

create table if not exists notification_logs (
  id uuid primary key,
  notification_id uuid not null references notifications(id) on delete cascade,
  provider varchar(30) not null check (provider in ('smtp', 'internal', 'outlook')),
  status varchar(30) not null,
  response_payload jsonb null,
  error_message text null,
  created_at timestamptz not null
);

commit;
