begin;

create table if not exists providers (
  id uuid primary key,
  type varchar(30) not null check (type in ('internal', 'external')),
  name varchar(255) not null,
  website_url text null,
  contact_email varchar(255) null,
  is_active boolean not null default true,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create table if not exists course_categories (
  id uuid primary key,
  name varchar(255) not null,
  code varchar(100) not null unique,
  parent_id uuid null references course_categories(id),
  is_active boolean not null default true
);

create table if not exists course_directions (
  id uuid primary key,
  code varchar(100) not null unique,
  name varchar(255) not null,
  sort_order int not null default 0,
  is_active boolean not null default true
);

create table if not exists skill_tags (
  id uuid primary key,
  name varchar(100) not null unique,
  slug varchar(120) not null unique
);

create table if not exists courses (
  id uuid primary key,
  type varchar(30) not null check (type in ('internal', 'external')),
  source_type varchar(30) not null check (source_type in ('catalog', 'requested', 'imported')),
  title varchar(500) not null,
  slug varchar(550) null unique,
  short_description text null,
  description text null,
  provider_id uuid null references providers(id),
  category_id uuid null references course_categories(id),
  direction_id uuid null references course_directions(id),
  level varchar(30) null check (level is null or level in ('beginner', 'intermediate', 'advanced')),
  duration_hours numeric(8,2) null,
  language varchar(30) null,
  is_mandatory_default boolean not null default false,
  status varchar(30) not null check (status in ('draft', 'published', 'archived')),
  thumbnail_file_id uuid null,
  created_by uuid null references users(id),
  updated_by uuid null references users(id),
  published_at timestamptz null,
  archived_at timestamptz null,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create table if not exists course_skill_tags (
  course_id uuid not null references courses(id) on delete cascade,
  skill_tag_id uuid not null references skill_tags(id) on delete cascade,
  primary key (course_id, skill_tag_id)
);

create table if not exists course_materials (
  id uuid primary key,
  course_id uuid not null references courses(id) on delete cascade,
  type varchar(30) not null check (type in ('file', 'link', 'video', 'scorm', 'pdf')),
  title varchar(255) not null,
  description text null,
  file_id uuid null,
  external_url text null,
  sort_order int not null default 0,
  is_required boolean not null default true,
  created_at timestamptz not null
);

create table if not exists course_modules (
  id uuid primary key,
  course_id uuid not null references courses(id) on delete cascade,
  title varchar(255) not null,
  description text null,
  sort_order int not null default 0,
  estimated_minutes int null,
  is_required boolean not null default true,
  created_at timestamptz not null
);

commit;
