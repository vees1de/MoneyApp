begin;

create table if not exists github_connections (
  id uuid primary key,
  company_id uuid null,
  title text not null,
  provider varchar(20) not null default 'github',
  auth_mode varchar(20) not null check (auth_mode in ('pat', 'github_app', 'oauth')),
  base_url text not null default 'https://api.github.com',
  status varchar(30) not null check (status in ('active', 'invalid', 'revoked', 'sync_error')),
  token_encrypted text null,
  token_last4 text null,
  github_app_id text null,
  github_installation_id text null,
  created_by uuid not null references users(id),
  last_sync_at timestamptz null,
  last_success_sync_at timestamptz null,
  last_error text null,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create table if not exists github_user_mappings (
  id uuid primary key,
  connection_id uuid not null references github_connections(id) on delete cascade,
  employee_user_id uuid not null references users(id) on delete cascade,
  github_login text not null,
  github_user_id bigint null,
  profile_url text null,
  match_source varchar(30) not null check (match_source in ('manual', 'email', 'imported', 'login', 'domain')),
  is_active boolean not null default true,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  unique(connection_id, employee_user_id),
  unique(connection_id, github_login)
);

create table if not exists github_users (
  id uuid primary key,
  connection_id uuid not null references github_connections(id) on delete cascade,
  github_user_id bigint not null,
  login text not null,
  name text null,
  email text null,
  avatar_url text null,
  html_url text null,
  company text null,
  location text null,
  bio text null,
  followers int null,
  following int null,
  public_repos int null,
  public_gists int null,
  created_at_remote timestamptz null,
  updated_at_remote timestamptz null,
  raw_payload jsonb not null default '{}'::jsonb,
  synced_at timestamptz not null,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  unique(connection_id, github_user_id),
  unique(connection_id, login)
);

create table if not exists github_repositories (
  id uuid primary key,
  connection_id uuid not null references github_connections(id) on delete cascade,
  github_repo_id bigint not null,
  owner_login text not null,
  name text not null,
  full_name text not null,
  private boolean not null default false,
  archived boolean not null default false,
  fork boolean not null default false,
  default_branch text null,
  language text null,
  size_kb int null,
  stargazers_count int null,
  watchers_count int null,
  forks_count int null,
  open_issues_count int null,
  subscribers_count int null,
  network_count int null,
  pushed_at timestamptz null,
  created_at_remote timestamptz null,
  updated_at_remote timestamptz null,
  html_url text null,
  raw_payload jsonb not null default '{}'::jsonb,
  synced_at timestamptz not null,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  unique(connection_id, github_repo_id)
);

create table if not exists github_repository_languages (
  id uuid primary key,
  repository_id uuid not null references github_repositories(id) on delete cascade,
  language_name text not null,
  bytes bigint not null,
  percent numeric(7,2) null,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  unique(repository_id, language_name)
);

create table if not exists github_repository_contributors (
  id uuid primary key,
  repository_id uuid not null references github_repositories(id) on delete cascade,
  github_user_id bigint null,
  github_login text not null,
  contributions int not null default 0,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  unique(repository_id, github_login)
);

create table if not exists github_pull_request_stats (
  id uuid primary key,
  connection_id uuid not null references github_connections(id) on delete cascade,
  employee_user_id uuid not null references users(id) on delete cascade,
  metric_date date not null,
  opened_prs int not null default 0,
  merged_prs int not null default 0,
  reviewed_prs int not null default 0,
  closed_prs int not null default 0,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  unique(connection_id, employee_user_id, metric_date)
);

create table if not exists github_commit_stats_daily (
  id uuid primary key,
  connection_id uuid not null references github_connections(id) on delete cascade,
  employee_user_id uuid not null references users(id) on delete cascade,
  metric_date date not null,
  commit_count int not null default 0,
  additions int null,
  deletions int null,
  changed_files int null,
  active_repos_count int not null default 0,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  unique(connection_id, employee_user_id, metric_date)
);

create table if not exists github_employee_language_profiles (
  id uuid primary key,
  connection_id uuid not null references github_connections(id) on delete cascade,
  employee_user_id uuid not null references users(id) on delete cascade,
  language_name text not null,
  bytes bigint not null default 0,
  percent numeric(7,2) not null default 0,
  repos_count int not null default 0,
  last_used_at timestamptz null,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  unique(connection_id, employee_user_id, language_name)
);

create table if not exists github_sync_jobs (
  id uuid primary key,
  connection_id uuid not null references github_connections(id) on delete cascade,
  job_type varchar(30) not null check (job_type in ('users_sync', 'repos_sync', 'languages_sync', 'activity_sync', 'backfill', 'full_sync')),
  status varchar(30) not null check (status in ('pending', 'processing', 'done', 'failed', 'retry')),
  cursor jsonb null,
  progress jsonb not null default '{}'::jsonb,
  attempt int not null default 0,
  started_at timestamptz null,
  finished_at timestamptz null,
  next_retry_at timestamptz null,
  error_text text null,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create index if not exists idx_github_connections_status on github_connections(status, updated_at desc);
create index if not exists idx_github_users_connection_login on github_users(connection_id, login);
create index if not exists idx_github_repo_connection_owner on github_repositories(connection_id, owner_login);
create index if not exists idx_github_repo_connection_pushed on github_repositories(connection_id, pushed_at desc);
create index if not exists idx_github_language_profiles_lookup on github_employee_language_profiles(connection_id, employee_user_id, percent desc);
create index if not exists idx_github_sync_jobs_lookup on github_sync_jobs(connection_id, status, created_at desc);

commit;
