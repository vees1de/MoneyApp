begin;

create table if not exists external_course_requests (
  id uuid primary key,
  request_no varchar(50) not null unique,
  employee_user_id uuid not null references users(id),
  department_id uuid null references departments(id),
  title varchar(500) not null,
  provider_id uuid null references providers(id),
  provider_name varchar(255) null,
  course_url text null,
  program_description text null,
  planned_start_date date null,
  planned_end_date date null,
  duration_hours numeric(8,2) null,
  cost_amount numeric(12,2) not null default 0,
  currency varchar(10) not null default 'RUB',
  business_goal text null,
  employee_comment text null,
  manager_comment text null,
  hr_comment text null,
  status varchar(40) not null check (status in (
    'draft', 'submitted', 'manager_approval', 'hr_approval', 'approved', 'rejected',
    'needs_revision', 'scheduled', 'in_training', 'completed', 'certificate_uploaded',
    'closed', 'canceled'
  )),
  calendar_conflict_status varchar(30) null check (calendar_conflict_status is null or calendar_conflict_status in ('not_checked', 'ok', 'conflict', 'warning', 'failed')),
  budget_check_status varchar(30) null check (budget_check_status is null or budget_check_status in ('not_checked', 'ok', 'exceeded', 'warning', 'failed')),
  current_approval_step_id uuid null,
  approved_at timestamptz null,
  rejected_at timestamptz null,
  sent_to_revision_at timestamptz null,
  training_started_at timestamptz null,
  training_completed_at timestamptz null,
  certificate_uploaded_at timestamptz null,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create table if not exists attached_documents (
  id uuid primary key,
  request_id uuid not null references external_course_requests(id) on delete cascade,
  file_id uuid not null references file_attachments(id) on delete cascade,
  document_type varchar(30) not null check (document_type in ('invoice', 'program', 'commercial_offer', 'certificate', 'other')),
  uploaded_by uuid not null references users(id),
  created_at timestamptz not null
);

create table if not exists budget_limits (
  id uuid primary key,
  scope_type varchar(30) not null check (scope_type in ('company', 'department', 'employee')),
  scope_id uuid null,
  period_year int not null,
  period_month int null check (period_month is null or (period_month between 1 and 12)),
  limit_amount numeric(12,2) not null,
  currency varchar(10) not null default 'RUB',
  is_active boolean not null default true,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create table if not exists budget_consumptions (
  id uuid primary key,
  budget_limit_id uuid not null references budget_limits(id) on delete cascade,
  request_id uuid not null references external_course_requests(id) on delete cascade,
  reserved_amount numeric(12,2) not null default 0,
  actual_amount numeric(12,2) not null default 0,
  status varchar(30) not null check (status in ('reserved', 'confirmed', 'released', 'spent')),
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create table if not exists approval_workflows (
  id uuid primary key,
  entity_type varchar(50) not null,
  name varchar(255) not null,
  is_active boolean not null default true,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create table if not exists approval_workflow_steps (
  id uuid primary key,
  workflow_id uuid not null references approval_workflows(id) on delete cascade,
  step_order int not null,
  role_code varchar(50) not null,
  approver_source varchar(30) not null check (approver_source in ('line_manager', 'specific_role', 'department_head', 'static_user')),
  approver_user_id uuid null references users(id),
  sla_hours int null,
  is_required boolean not null default true,
  unique (workflow_id, step_order)
);

create table if not exists approval_steps (
  id uuid primary key,
  entity_type varchar(50) not null,
  entity_id uuid not null,
  step_order int not null,
  approver_user_id uuid not null references users(id),
  role_code varchar(50) not null,
  status varchar(30) not null check (status in ('pending', 'approved', 'rejected', 'revision_requested', 'skipped')),
  comment text null,
  due_at timestamptz null,
  acted_at timestamptz null,
  created_at timestamptz not null
);

create table if not exists approval_history (
  id uuid primary key,
  entity_type varchar(50) not null,
  entity_id uuid not null,
  step_id uuid null references approval_steps(id) on delete set null,
  action varchar(30) not null check (action in ('submitted', 'approved', 'rejected', 'revision_requested', 'resubmitted', 'canceled')),
  from_status varchar(30) null,
  to_status varchar(30) null,
  performed_by uuid not null references users(id),
  comment text null,
  created_at timestamptz not null
);

alter table external_course_requests
  add constraint external_course_requests_current_step_fk
  foreign key (current_approval_step_id) references approval_steps(id);

commit;
