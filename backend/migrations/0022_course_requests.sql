begin;

create table if not exists course_requests (
  id uuid primary key,
  request_no varchar(50) not null unique,
  course_id uuid not null references courses(id),
  employee_user_id uuid not null references users(id),
  department_id uuid null references departments(id),
  manager_user_id uuid null references users(id),
  hr_user_id uuid null references users(id),
  enrollment_id uuid null references enrollments(id),
  certificate_id uuid null references certificates(id),
  status varchar(40) not null check (status in (
    'pending_manager_approval',
    'pending_hr_approval',
    'approved_waiting_start',
    'in_progress',
    'completed_waiting_certificate',
    'certificate_under_review',
    'certificate_approved',
    'certificate_rejected',
    'rejected_by_manager',
    'rejected_by_hr',
    'canceled_by_employee'
  )),
  employee_comment text null,
  manager_comment text null,
  hr_comment text null,
  rejection_reason text null,
  deadline_at timestamptz null,
  requested_at timestamptz not null,
  manager_approved_at timestamptz null,
  hr_approved_at timestamptz null,
  approved_at timestamptz null,
  started_at timestamptz null,
  completed_at timestamptz null,
  certificate_uploaded_at timestamptz null,
  certificate_approved_at timestamptz null,
  certificate_manager_approved_at timestamptz null,
  certificate_manager_approved_by uuid null references users(id),
  certificate_hr_approved_at timestamptz null,
  certificate_hr_approved_by uuid null references users(id),
  canceled_at timestamptz null,
  rejected_at timestamptz null,
  rejected_by uuid null references users(id),
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create table if not exists course_request_events (
  id uuid primary key,
  course_request_id uuid not null references course_requests(id) on delete cascade,
  action varchar(50) not null check (action in (
    'created',
    'manager_approved',
    'hr_approved',
    'rejected_by_manager',
    'rejected_by_hr',
    'canceled_by_employee',
    'started',
    'completed',
    'certificate_uploaded',
    'certificate_approved_by_manager',
    'certificate_approved_by_hr',
    'certificate_rejected_by_manager',
    'certificate_rejected_by_hr'
  )),
  performed_by uuid not null references users(id),
  comment text null,
  created_at timestamptz not null
);

create index if not exists course_requests_employee_idx on course_requests (employee_user_id, created_at desc);
create index if not exists course_requests_manager_idx on course_requests (manager_user_id, created_at desc);
create index if not exists course_requests_hr_idx on course_requests (hr_user_id, created_at desc);
create index if not exists course_requests_status_idx on course_requests (status, created_at desc);
create index if not exists course_request_events_request_idx on course_request_events (course_request_id, created_at asc);

commit;
