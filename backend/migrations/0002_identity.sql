begin;

create table if not exists users (
  id uuid primary key,
  email citext not null unique,
  password_hash text null,
  status text not null check (status in ('active', 'invited', 'blocked', 'deleted')),
  is_email_verified boolean not null default false,
  last_login_at timestamptz null,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  deleted_at timestamptz null
);

create table if not exists roles (
  id uuid primary key,
  code varchar(50) not null unique,
  name varchar(100) not null,
  description text null,
  is_system boolean not null default false
);

create table if not exists permissions (
  id uuid primary key,
  code varchar(100) not null unique,
  module varchar(50) not null,
  action varchar(50) not null,
  description text null
);

create table if not exists role_permissions (
  role_id uuid not null references roles(id) on delete cascade,
  permission_id uuid not null references permissions(id) on delete cascade,
  primary key (role_id, permission_id)
);

create table if not exists user_roles (
  user_id uuid not null references users(id) on delete cascade,
  role_id uuid not null references roles(id) on delete cascade,
  scope_type varchar(30) null check (scope_type is null or scope_type in ('global', 'department', 'program')),
  scope_id uuid null,
  created_at timestamptz not null,
  primary key (user_id, role_id, scope_type, scope_id)
);

create table if not exists sessions (
  id uuid primary key,
  user_id uuid not null references users(id) on delete cascade,
  refresh_token_hash text not null unique,
  user_agent text null,
  ip inet null,
  expires_at timestamptz not null,
  revoked_at timestamptz null,
  created_at timestamptz not null
);

create table if not exists password_resets (
  id uuid primary key,
  user_id uuid not null references users(id) on delete cascade,
  token_hash text not null unique,
  expires_at timestamptz not null,
  used_at timestamptz null,
  created_at timestamptz not null
);

commit;
