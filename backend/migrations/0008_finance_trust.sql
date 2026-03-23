begin;

alter table accounts
  add column if not exists last_recalculated_at timestamptz null;

alter table finance_transactions
  add column if not exists posting_state text not null default 'posted',
  add column if not exists deleted_at timestamptz null,
  add column if not exists deleted_by uuid null references users(id) on delete set null,
  add column if not exists delete_reason text null,
  add column if not exists restored_at timestamptz null,
  add column if not exists source text not null default 'manual',
  add column if not exists source_ref_id uuid null,
  add column if not exists template_id uuid null,
  add column if not exists recurring_rule_id uuid null,
  add column if not exists planned_expense_id uuid null,
  add column if not exists title_normalized text null,
  add column if not exists is_mandatory boolean not null default false,
  add column if not exists is_subscription boolean not null default false,
  add column if not exists base_currency text null,
  add column if not exists base_amount numeric(14,2) null,
  add column if not exists fx_rate numeric(18,8) null;

do $$
begin
  if not exists (
    select 1
    from pg_constraint
    where conname = 'finance_transactions_posting_state_check'
  ) then
    alter table finance_transactions
      add constraint finance_transactions_posting_state_check
        check (posting_state in ('draft', 'posted'));
  end if;
end $$;

do $$
begin
  if not exists (
    select 1
    from pg_constraint
    where conname = 'finance_transactions_source_check'
  ) then
    alter table finance_transactions
      add constraint finance_transactions_source_check
        check (source in ('manual', 'recurring', 'review', 'system'));
  end if;
end $$;

create index if not exists finance_transactions_active_user_id_occurred_at_idx
  on finance_transactions (user_id, occurred_at desc)
  where deleted_at is null and posting_state = 'posted';

create index if not exists finance_transactions_active_user_id_category_id_occurred_at_idx
  on finance_transactions (user_id, category_id, occurred_at desc)
  where deleted_at is null and posting_state = 'posted' and type = 'expense';

create index if not exists finance_transactions_active_user_id_transfer_account_id_occurred_at_idx
  on finance_transactions (user_id, transfer_account_id, occurred_at desc)
  where deleted_at is null;

alter table audit_logs
  add column if not exists source text not null default 'manual',
  add column if not exists request_id text null,
  add column if not exists session_id uuid null references sessions(id) on delete set null,
  add column if not exists change_set jsonb not null default '{}'::jsonb,
  add column if not exists actor_type text not null default 'user',
  add column if not exists actor_id uuid null references users(id) on delete set null;

create index if not exists audit_logs_entity_created_at_idx
  on audit_logs (entity_type, entity_id, created_at desc);

commit;
