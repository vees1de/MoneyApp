begin;

create table if not exists internal_programs (
  id uuid primary key,
  title varchar(255) not null,
  description text null,
  direction_id uuid null references course_directions(id),
  status varchar(30) not null check (status in ('draft', 'published', 'archived')),
  created_by uuid not null references users(id),
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create table if not exists training_groups (
  id uuid primary key,
  program_id uuid not null references internal_programs(id) on delete cascade,
  name varchar(255) not null,
  capacity int null,
  status varchar(30) not null check (status in ('planned', 'open', 'full', 'in_progress', 'completed', 'canceled')),
  enrollment_open_at timestamptz null,
  enrollment_close_at timestamptz null,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create table if not exists sessions_university (
  id uuid primary key,
  group_id uuid not null references training_groups(id) on delete cascade,
  trainer_user_id uuid not null references users(id),
  title varchar(255) not null,
  description text null,
  start_at timestamptz not null,
  end_at timestamptz not null,
  location text null,
  meeting_url text null,
  status varchar(30) not null check (status in ('planned', 'held', 'canceled', 'rescheduled')),
  calendar_event_id uuid null references calendar_events(id),
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create table if not exists group_participants (
  id uuid primary key,
  group_id uuid not null references training_groups(id) on delete cascade,
  user_id uuid not null references users(id) on delete cascade,
  status varchar(30) not null check (status in ('enrolled', 'waitlisted', 'attended', 'completed', 'canceled')),
  enrolled_at timestamptz not null,
  completed_at timestamptz null,
  unique (group_id, user_id)
);

create table if not exists trainer_feedback (
  id uuid primary key,
  session_id uuid not null references sessions_university(id) on delete cascade,
  participant_user_id uuid not null references users(id),
  trainer_user_id uuid not null references users(id),
  attendance_status varchar(30) not null check (attendance_status in ('attended', 'absent', 'excused')),
  score numeric(5,2) null,
  comment text null,
  created_at timestamptz not null
);

create table if not exists participant_feedback (
  id uuid primary key,
  program_id uuid null references internal_programs(id) on delete cascade,
  session_id uuid null references sessions_university(id) on delete cascade,
  participant_user_id uuid not null references users(id),
  rating int not null check (rating between 1 and 5),
  comment text null,
  created_at timestamptz not null
);

create table if not exists internal_certificates (
  id uuid primary key,
  program_id uuid not null references internal_programs(id) on delete cascade,
  group_id uuid null references training_groups(id) on delete set null,
  user_id uuid not null references users(id),
  file_id uuid null references file_attachments(id),
  issued_at timestamptz not null,
  issued_by uuid not null references users(id)
);

commit;
