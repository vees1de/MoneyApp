begin;

create table if not exists weekly_reviews (
  id uuid primary key,
  user_id uuid not null references users(id) on delete cascade,
  account_id uuid null references accounts(id) on delete cascade,
  period_start date not null,
  period_end date not null,
  expected_balance numeric(14,2) not null,
  actual_balance numeric(14,2) null,
  delta numeric(14,2) null,
  status text not null,
  resolution_note text null,
  created_at timestamptz not null,
  completed_at timestamptz null,
  constraint weekly_reviews_status_check check (status in ('pending', 'matched', 'discrepancy_found', 'resolved', 'skipped'))
);

create unique index if not exists weekly_reviews_user_period_global_uidx
  on weekly_reviews (user_id, period_start, period_end)
  where account_id is null;

create unique index if not exists weekly_reviews_user_account_period_uidx
  on weekly_reviews (user_id, account_id, period_start, period_end)
  where account_id is not null;

commit;
