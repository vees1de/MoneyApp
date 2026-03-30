begin;

create table if not exists learning_metrics_daily (
  id uuid primary key,
  metric_date date not null,
  department_id uuid null references departments(id),
  course_id uuid null references courses(id),
  program_id uuid null references internal_programs(id),
  mandatory_assigned_count int not null default 0,
  mandatory_completed_count int not null default 0,
  overdue_count int not null default 0,
  completion_rate numeric(7,2) not null default 0,
  budget_plan numeric(12,2) not null default 0,
  budget_actual numeric(12,2) not null default 0,
  calendar_conflict_count int not null default 0,
  avg_approval_time_hours numeric(10,2) not null default 0,
  created_at timestamptz not null,
  unique (metric_date, department_id, course_id, program_id)
);

commit;
