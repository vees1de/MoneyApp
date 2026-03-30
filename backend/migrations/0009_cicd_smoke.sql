begin;

create table if not exists cicd_smoke_checks (
  id uuid primary key,
  user_id uuid not null references users(id) on delete cascade,
  session_id uuid not null references sessions(id) on delete cascade,
  request_id text not null,
  trigger text not null,
  created_at timestamptz not null
);

create index if not exists cicd_smoke_checks_user_id_created_at_idx
  on cicd_smoke_checks (user_id, created_at desc);

commit;
