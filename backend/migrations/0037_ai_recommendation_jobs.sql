begin;

create table if not exists ai_recommendation_jobs (
  id uuid primary key,
  user_id uuid not null references users(id) on delete cascade,
  status varchar(30) not null check (status in ('pending', 'processing', 'retry', 'done', 'failed')),
  attempt int not null default 0,
  result jsonb null,
  last_error text null,
  started_at timestamptz null,
  finished_at timestamptz null,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create index if not exists ai_recommendation_jobs_user_status_idx
  on ai_recommendation_jobs (user_id, status, created_at desc);

commit;
