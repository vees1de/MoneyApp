begin;

alter table outlook_accounts
  add column if not exists auth_mode varchar(30) not null default 'oauth',
  add column if not exists system_email_enabled boolean not null default false,
  add column if not exists last_mail_sync_at timestamptz null,
  add column if not exists last_calendar_sync_at timestamptz null,
  add column if not exists last_error text null;

alter table outlook_accounts
  drop constraint if exists outlook_accounts_auth_mode_check;

alter table outlook_accounts
  add constraint outlook_accounts_auth_mode_check
    check (auth_mode in ('oauth', 'access_token'));

create table if not exists outlook_messages (
  id uuid primary key,
  user_id uuid not null references users(id) on delete cascade,
  account_id uuid not null references outlook_accounts(id) on delete cascade,
  external_message_id varchar(255) not null,
  conversation_id varchar(255) null,
  subject text not null,
  sender_email varchar(255) null,
  sender_name varchar(255) null,
  received_at timestamptz not null,
  is_read boolean not null default false,
  body_preview text null,
  web_link text null,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  unique (user_id, external_message_id)
);

create index if not exists idx_outlook_messages_user_received_at
  on outlook_messages (user_id, received_at desc);

alter table calendar_events
  drop constraint if exists calendar_events_source_type_check;

alter table calendar_events
  add constraint calendar_events_source_type_check
    check (source_type in ('external_request', 'internal_session', 'deadline_reminder', 'outlook_remote'));

create index if not exists idx_calendar_events_user_provider_start_at
  on calendar_events (user_id, provider, start_at);

commit;
