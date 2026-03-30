begin;

create index if not exists manager_relations_manager_primary_employee_idx
  on manager_relations (manager_user_id, is_primary, employee_user_id);

create index if not exists calendar_events_user_status_start_idx
  on calendar_events (user_id, status, start_at);

commit;
