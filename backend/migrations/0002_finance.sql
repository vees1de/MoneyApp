begin;

create table if not exists accounts (
  id uuid primary key,
  user_id uuid not null references users(id) on delete cascade,
  name text not null,
  kind text not null,
  currency text not null,
  opening_balance numeric(14,2) not null default 0,
  current_balance numeric(14,2) not null default 0,
  is_archived boolean not null default false,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  constraint accounts_kind_check check (kind in ('cash', 'bank_card', 'bank_account', 'savings', 'virtual'))
);

create index if not exists accounts_user_id_idx on accounts (user_id);

create table if not exists finance_transactions (
  id uuid primary key,
  user_id uuid not null references users(id) on delete cascade,
  account_id uuid not null references accounts(id) on delete restrict,
  transfer_account_id uuid null references accounts(id) on delete restrict,
  type text not null,
  category_id uuid null,
  amount numeric(14,2) not null,
  currency text not null,
  direction text not null,
  title text null,
  note text null,
  occurred_at timestamptz not null,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  constraint finance_transactions_type_check check (type in ('income', 'expense', 'transfer', 'correction')),
  constraint finance_transactions_direction_check check (direction in ('inflow', 'outflow', 'internal'))
);

create index if not exists finance_transactions_user_id_occurred_at_idx on finance_transactions (user_id, occurred_at desc);
create index if not exists finance_transactions_account_id_idx on finance_transactions (account_id);
create index if not exists finance_transactions_transfer_account_id_idx on finance_transactions (transfer_account_id);

commit;
