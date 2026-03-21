begin;

create table if not exists app_jobs (
  id uuid primary key,
  user_id uuid null references users(id) on delete cascade,
  job_type text not null,
  status text not null default 'pending',
  payload jsonb not null default '{}'::jsonb,
  run_at timestamptz not null,
  locked_at timestamptz null,
  completed_at timestamptz null,
  created_at timestamptz not null
);

create index if not exists app_jobs_status_run_at_idx on app_jobs (status, run_at);

commit;
