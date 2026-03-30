begin;

-- Add missing columns expected by identity module
alter table users add column if not exists password_hash text null;
alter table users add column if not exists status text null;
alter table users add column if not exists is_email_verified boolean not null default false;
alter table users add column if not exists last_login_at timestamptz null;
alter table users add column if not exists deleted_at timestamptz null;

-- Set default status for existing rows
update users set status = 'active' where status is null;

-- Now apply NOT NULL + check constraint
alter table users alter column status set not null;
alter table users add constraint users_status_check
  check (status in ('active', 'invited', 'blocked', 'deleted'));

-- Fill null emails with placeholder based on display_name or id
update users
  set email = coalesce(
    lower(replace(display_name, ' ', '.')) || '@placeholder.local',
    id::text || '@placeholder.local'
  )
where email is null or email = '';

-- Convert email to citext if extension is available
do $$
begin
  if exists (select 1 from pg_extension where extname = 'citext') then
    execute 'alter table users alter column email type citext using email::citext';
  end if;
end $$;

alter table users alter column email set not null;

-- Add unique constraint on email if missing
do $$
begin
  if not exists (
    select 1 from pg_indexes where tablename = 'users' and indexname = 'users_email_key'
  ) then
    execute 'alter table users add constraint users_email_key unique (email)';
  end if;
exception when others then
  raise notice 'email unique constraint already exists or cannot be added: %', sqlerrm;
end $$;

commit;
