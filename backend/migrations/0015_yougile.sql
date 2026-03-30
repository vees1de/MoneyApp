begin;

create table if not exists integration_yougile_connections (
  id uuid primary key,
  company_id uuid null,
  title text null,
  api_base_url text not null default 'https://yougile.com',
  yougile_company_id text not null,
  api_key_encrypted text not null,
  api_key_last4 text null,
  status varchar(30) not null check (status in ('active', 'invalid', 'revoked', 'sync_error')),
  created_by uuid not null references users(id),
  last_sync_at timestamptz null,
  last_success_sync_at timestamptz null,
  last_error text null,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  unique (company_id, yougile_company_id),
  unique (created_by, yougile_company_id)
);

create table if not exists yougile_users (
  id uuid primary key,
  connection_id uuid not null references integration_yougile_connections(id) on delete cascade,
  yougile_user_id text not null,
  email text null,
  real_name text null,
  is_admin boolean not null default false,
  status text null,
  last_activity_at timestamptz null,
  raw_payload jsonb not null default '{}'::jsonb,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  unique (connection_id, yougile_user_id)
);

create table if not exists yougile_employee_mappings (
  id uuid primary key,
  connection_id uuid not null references integration_yougile_connections(id) on delete cascade,
  employee_user_id uuid not null references users(id) on delete cascade,
  yougile_user_id text not null,
  match_source varchar(30) not null check (match_source in ('manual', 'email', 'imported')),
  is_active boolean not null default true,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  unique (connection_id, employee_user_id),
  unique (connection_id, yougile_user_id)
);

create table if not exists yougile_projects (
  id uuid primary key,
  connection_id uuid not null references integration_yougile_connections(id) on delete cascade,
  yougile_project_id text not null,
  title text not null,
  deleted boolean not null default false,
  raw_payload jsonb not null default '{}'::jsonb,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  unique (connection_id, yougile_project_id)
);

create table if not exists yougile_boards (
  id uuid primary key,
  connection_id uuid not null references integration_yougile_connections(id) on delete cascade,
  yougile_board_id text not null,
  yougile_project_id text not null,
  title text not null,
  deleted boolean not null default false,
  raw_payload jsonb not null default '{}'::jsonb,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  unique (connection_id, yougile_board_id)
);

create table if not exists yougile_columns (
  id uuid primary key,
  connection_id uuid not null references integration_yougile_connections(id) on delete cascade,
  yougile_column_id text not null,
  yougile_board_id text not null,
  title text not null,
  color int null,
  deleted boolean not null default false,
  raw_payload jsonb not null default '{}'::jsonb,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  unique (connection_id, yougile_column_id)
);

create table if not exists yougile_tasks (
  id uuid primary key,
  connection_id uuid not null references integration_yougile_connections(id) on delete cascade,
  yougile_task_id text not null,
  title text not null,
  description text null,
  yougile_column_id text null,
  yougile_board_id text null,
  yougile_project_id text null,
  created_by_yougile_user_id text null,
  created_at_remote timestamptz null,
  updated_at_remote timestamptz null,
  completed boolean not null default false,
  completed_at_remote timestamptz null,
  archived boolean not null default false,
  archived_at_remote timestamptz null,
  deadline_at timestamptz null,
  deadline_start_at timestamptz null,
  deadline_with_time boolean null,
  time_plan_hours numeric(10,2) null,
  time_work_hours numeric(10,2) null,
  stopwatch_seconds bigint null,
  timer_seconds bigint null,
  id_task_common text null,
  id_task_project text null,
  type varchar(20) null,
  raw_payload jsonb not null default '{}'::jsonb,
  synced_at timestamptz not null,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  unique (connection_id, yougile_task_id)
);

create table if not exists yougile_task_assignees (
  task_id uuid not null references yougile_tasks(id) on delete cascade,
  yougile_user_id text not null,
  primary key (task_id, yougile_user_id)
);

create table if not exists yougile_sync_jobs (
  id uuid primary key,
  connection_id uuid not null references integration_yougile_connections(id) on delete cascade,
  job_type varchar(30) not null check (job_type in ('full_sync', 'users_sync', 'tasks_sync', 'backfill', 'webhook_reconcile', 'structure_sync')),
  status varchar(30) not null check (status in ('pending', 'processing', 'done', 'failed', 'retry')),
  cursor jsonb null,
  progress jsonb not null default '{}'::jsonb,
  started_at timestamptz null,
  finished_at timestamptz null,
  attempt int not null default 0,
  next_retry_at timestamptz null,
  error_text text null,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create table if not exists yougile_webhook_subscriptions (
  id uuid primary key,
  connection_id uuid not null references integration_yougile_connections(id) on delete cascade,
  yougile_webhook_id text not null,
  url text not null,
  event_pattern text not null,
  filters jsonb not null default '{}'::jsonb,
  status varchar(30) not null check (status in ('active', 'paused', 'broken')),
  created_at timestamptz not null,
  updated_at timestamptz not null,
  unique (connection_id, yougile_webhook_id)
);

create table if not exists yougile_webhook_events (
  id uuid primary key,
  connection_id uuid not null references integration_yougile_connections(id) on delete cascade,
  event_id text null,
  event_type text not null,
  payload jsonb not null default '{}'::jsonb,
  received_at timestamptz not null,
  processed_at timestamptz null,
  status varchar(30) not null check (status in ('pending', 'processed', 'failed', 'ignored')),
  error_text text null,
  dedupe_hash text not null unique
);

create table if not exists yougile_employee_metrics_daily (
  id uuid primary key,
  connection_id uuid not null references integration_yougile_connections(id) on delete cascade,
  employee_user_id uuid not null references users(id) on delete cascade,
  metric_date date not null,
  assigned_total int not null default 0,
  active_total int not null default 0,
  completed_total int not null default 0,
  overdue_total int not null default 0,
  completed_on_time_total int not null default 0,
  completion_rate numeric(7,2) not null default 0,
  on_time_rate numeric(7,2) not null default 0,
  avg_cycle_time_hours numeric(10,2) null,
  avg_delay_hours numeric(10,2) null,
  planned_hours numeric(10,2) null,
  worked_hours numeric(10,2) null,
  workload_score numeric(7,2) null,
  execution_score numeric(7,2) null,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  unique (connection_id, employee_user_id, metric_date)
);

create index if not exists idx_yougile_connections_status on integration_yougile_connections(status, updated_at desc);
create index if not exists idx_yougile_users_connection_email on yougile_users(connection_id, email);
create index if not exists idx_yougile_mappings_connection_employee on yougile_employee_mappings(connection_id, employee_user_id);
create index if not exists idx_yougile_projects_connection on yougile_projects(connection_id);
create index if not exists idx_yougile_boards_connection on yougile_boards(connection_id);
create index if not exists idx_yougile_columns_connection on yougile_columns(connection_id);
create index if not exists idx_yougile_tasks_connection_sync on yougile_tasks(connection_id, synced_at desc);
create index if not exists idx_yougile_tasks_connection_completed on yougile_tasks(connection_id, completed, archived);
create index if not exists idx_yougile_sync_jobs_connection_status on yougile_sync_jobs(connection_id, status, created_at desc);
create index if not exists idx_yougile_webhook_events_connection_status on yougile_webhook_events(connection_id, status, received_at desc);
create index if not exists idx_yougile_employee_metrics_daily_lookup on yougile_employee_metrics_daily(connection_id, employee_user_id, metric_date desc);

commit;
