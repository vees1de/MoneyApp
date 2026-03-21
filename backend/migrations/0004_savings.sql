begin;

create table if not exists savings_goals (
  id uuid primary key,
  user_id uuid not null references users(id) on delete cascade,
  title text not null,
  target_amount numeric(14,2) not null,
  current_amount numeric(14,2) not null default 0,
  currency text not null,
  target_date date null,
  priority text not null default 'medium',
  status text not null default 'active',
  created_at timestamptz not null,
  updated_at timestamptz not null,
  constraint savings_goals_priority_check check (priority in ('low', 'medium', 'high')),
  constraint savings_goals_status_check check (status in ('active', 'paused', 'completed', 'archived'))
);

create index if not exists savings_goals_user_id_idx on savings_goals (user_id);

commit;
