begin;

create table if not exists departments (
  id uuid primary key,
  name varchar(255) not null,
  code varchar(100) null unique,
  parent_id uuid null references departments(id),
  head_user_id uuid null references users(id),
  is_active boolean not null default true,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create table if not exists employee_profiles (
  id uuid primary key,
  user_id uuid not null unique references users(id) on delete cascade,
  employee_no varchar(100) null unique,
  first_name varchar(100) not null,
  last_name varchar(100) not null,
  middle_name varchar(100) null,
  position_title varchar(255) null,
  department_id uuid null references departments(id),
  hire_date date null,
  employment_status varchar(30) not null check (employment_status in ('active', 'on_leave', 'dismissed')),
  timezone varchar(100) null,
  outlook_email varchar(255) null,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create table if not exists manager_relations (
  id uuid primary key,
  employee_user_id uuid not null references users(id) on delete cascade,
  manager_user_id uuid not null references users(id) on delete cascade,
  relation_type varchar(30) not null check (relation_type in ('line_manager', 'functional_manager')),
  is_primary boolean not null default false,
  created_at timestamptz not null,
  unique (employee_user_id, manager_user_id, relation_type)
);

create table if not exists org_groups (
  id uuid primary key,
  name varchar(255) not null,
  code varchar(100) null unique,
  department_id uuid null references departments(id),
  is_active boolean not null default true,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create table if not exists org_group_members (
  group_id uuid not null references org_groups(id) on delete cascade,
  user_id uuid not null references users(id) on delete cascade,
  created_at timestamptz not null,
  primary key (group_id, user_id)
);

commit;
