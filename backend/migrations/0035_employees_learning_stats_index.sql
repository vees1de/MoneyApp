begin;

create index if not exists enrollments_deadline_status_idx on enrollments (deadline_at, status)
  where deadline_at is not null and status <> 'completed';

commit;
