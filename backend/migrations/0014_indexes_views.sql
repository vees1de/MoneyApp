begin;

create index if not exists users_email_idx on users (email);
create index if not exists sessions_user_id_idx on sessions (user_id);
create index if not exists sessions_expires_at_idx on sessions (expires_at);
create index if not exists employee_profiles_department_id_idx on employee_profiles (department_id);
create index if not exists courses_search_idx on courses using gin (to_tsvector('simple', coalesce(title, '') || ' ' || coalesce(short_description, '') || ' ' || coalesce(description, '')));
create index if not exists courses_filter_idx on courses (type, status, direction_id, category_id);
create index if not exists enrollments_user_status_idx on enrollments (user_id, status);
create index if not exists enrollments_course_user_idx on enrollments (course_id, user_id);
create index if not exists external_course_requests_employee_status_created_idx on external_course_requests (employee_user_id, status, created_at desc);
create index if not exists approval_steps_approver_status_due_idx on approval_steps (approver_user_id, status, due_at);
create index if not exists notifications_user_status_created_idx on notifications (user_id, status, created_at desc);
create index if not exists audit_logs_entity_idx on audit_logs (entity_type, entity_id, created_at desc);
create index if not exists background_jobs_pending_idx on background_jobs (queue, status, run_at);
create index if not exists outbox_messages_pending_idx on outbox_messages (status, available_at);

create or replace view reporting_external_request_funnel as
select
  status,
  count(*)::bigint as total
from external_course_requests
group by status;

create or replace view reporting_mandatory_training_compliance as
select
  ep.department_id,
  count(*) filter (where e.is_mandatory) as mandatory_assigned_count,
  count(*) filter (where e.is_mandatory and e.status = 'completed') as mandatory_completed_count,
  count(*) filter (where e.is_mandatory and e.deadline_at is not null and e.deadline_at < now() and e.status <> 'completed') as overdue_count
from enrollments e
join employee_profiles ep on ep.user_id = e.user_id
group by ep.department_id;

commit;
