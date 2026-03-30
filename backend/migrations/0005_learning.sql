begin;

create table if not exists course_assignments (
  id uuid primary key,
  course_id uuid not null references courses(id),
  assignment_type varchar(30) not null check (assignment_type in ('individual', 'department', 'group', 'role_based')),
  target_type varchar(30) not null check (target_type in ('user', 'department', 'group')),
  target_id uuid not null,
  assigned_by uuid not null references users(id),
  priority varchar(30) not null check (priority in ('mandatory', 'recommended')),
  reason text null,
  start_at timestamptz null,
  deadline_at timestamptz null,
  status varchar(30) not null check (status in ('active', 'canceled', 'completed', 'expired')),
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create table if not exists enrollments (
  id uuid primary key,
  course_id uuid not null references courses(id),
  user_id uuid not null references users(id),
  assignment_id uuid null references course_assignments(id),
  source varchar(30) not null check (source in ('self', 'assignment', 'external_request', 'program')),
  status varchar(30) not null check (status in ('not_started', 'in_progress', 'completed', 'failed', 'canceled')),
  enrolled_at timestamptz not null,
  started_at timestamptz null,
  completed_at timestamptz null,
  deadline_at timestamptz null,
  last_activity_at timestamptz null,
  completion_percent numeric(5,2) not null default 0,
  is_mandatory boolean not null default false,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  unique (course_id, user_id, source, assignment_id)
);

create table if not exists module_progress (
  id uuid primary key,
  enrollment_id uuid not null references enrollments(id) on delete cascade,
  course_module_id uuid not null references course_modules(id) on delete cascade,
  status varchar(30) not null check (status in ('not_started', 'in_progress', 'completed')),
  progress_percent numeric(5,2) not null default 0,
  started_at timestamptz null,
  completed_at timestamptz null,
  updated_at timestamptz not null,
  unique (enrollment_id, course_module_id)
);

create table if not exists completion_records (
  id uuid primary key,
  enrollment_id uuid not null unique references enrollments(id) on delete cascade,
  completion_type varchar(30) not null check (completion_type in ('auto', 'manual', 'certificate_verified', 'trainer_confirmed')),
  completed_by uuid null references users(id),
  score numeric(6,2) null,
  completed_at timestamptz not null,
  notes text null
);

commit;
