begin;

create table if not exists tests (
  id uuid primary key,
  course_id uuid null references courses(id),
  title varchar(255) not null,
  description text null,
  attempts_limit int null,
  passing_score numeric(5,2) not null,
  shuffle_questions boolean not null default false,
  shuffle_answers boolean not null default false,
  status varchar(30) not null check (status in ('draft', 'published', 'archived')),
  created_by uuid not null references users(id),
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create table if not exists questions (
  id uuid primary key,
  test_id uuid not null references tests(id) on delete cascade,
  type varchar(30) not null check (type in ('single_choice', 'multiple_choice', 'text', 'number', 'true_false')),
  text text not null,
  explanation text null,
  sort_order int not null default 0,
  points numeric(6,2) not null default 1,
  is_required boolean not null default true
);

create table if not exists answer_options (
  id uuid primary key,
  question_id uuid not null references questions(id) on delete cascade,
  text text not null,
  is_correct boolean not null default false,
  sort_order int not null default 0
);

create table if not exists test_attempts (
  id uuid primary key,
  test_id uuid not null references tests(id) on delete cascade,
  user_id uuid not null references users(id) on delete cascade,
  enrollment_id uuid null references enrollments(id) on delete set null,
  attempt_no int not null,
  status varchar(30) not null check (status in ('started', 'submitted', 'checked', 'canceled')),
  started_at timestamptz not null,
  submitted_at timestamptz null,
  checked_at timestamptz null,
  score numeric(6,2) null,
  passed boolean null,
  unique (test_id, user_id, attempt_no)
);

create table if not exists test_answers (
  id uuid primary key,
  attempt_id uuid not null references test_attempts(id) on delete cascade,
  question_id uuid not null references questions(id) on delete cascade,
  answer_text text null,
  selected_option_id uuid null references answer_options(id),
  selected_option_ids jsonb null,
  is_correct boolean null,
  awarded_points numeric(6,2) null,
  unique (attempt_id, question_id)
);

create table if not exists test_results (
  id uuid primary key,
  test_id uuid not null references tests(id) on delete cascade,
  user_id uuid not null references users(id) on delete cascade,
  best_attempt_id uuid not null references test_attempts(id),
  best_score numeric(6,2) not null,
  passed boolean not null,
  completed_at timestamptz not null,
  unique (test_id, user_id)
);

commit;
