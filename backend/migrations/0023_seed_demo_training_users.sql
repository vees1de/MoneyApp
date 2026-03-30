begin;

insert into users (
  id, email, password_hash, status, is_email_verified, created_at, updated_at
)
values
  (
    '11111111-1111-1111-1111-111111111111',
    'employee.demo@moneyapp.local',
    '$2a$10$OELjNz/M0uGcJfoWwNywFOcsDD4aop6Ryz9emIV6ig8O5ffjNDYoa',
    'active',
    true,
    now(),
    now()
  ),
  (
    '22222222-2222-2222-2222-222222222222',
    'manager.demo@moneyapp.local',
    '$2a$10$EsmQq0AWyQAjTXLvDFIKaOHho6YU5..5a1cE6SQ.IhCEznQIMmr/2',
    'active',
    true,
    now(),
    now()
  ),
  (
    '33333333-3333-3333-3333-333333333333',
    'hr.demo@moneyapp.local',
    '$2a$10$PAuDzk0r6k7yJsCUS7glUOJJkR4NxnFLG.N3YLDWoBNAJErbXDTCu',
    'active',
    true,
    now(),
    now()
  )
on conflict (email) do update
set password_hash = excluded.password_hash,
    status = excluded.status,
    is_email_verified = excluded.is_email_verified,
    updated_at = now();

insert into departments (
  id, name, code, parent_id, head_user_id, is_active, created_at, updated_at
)
values (
  '44444444-4444-4444-4444-444444444444',
  'Engineering Demo',
  'ENG-DEMO',
  null,
  '22222222-2222-2222-2222-222222222222',
  true,
  now(),
  now()
)
on conflict (code) do update
set name = excluded.name,
    head_user_id = excluded.head_user_id,
    is_active = true,
    updated_at = now();

insert into employee_profiles (
  id, user_id, employee_no, first_name, last_name, middle_name, position_title,
  department_id, hire_date, employment_status, timezone, outlook_email, created_at, updated_at
)
values
  (
    '51111111-1111-1111-1111-111111111111',
    '11111111-1111-1111-1111-111111111111',
    'EMP-DEMO-001',
    'Иван',
    'Сотрудников',
    null,
    'Backend Engineer',
    '44444444-4444-4444-4444-444444444444',
    current_date,
    'active',
    'Asia/Yakutsk',
    'employee.demo@moneyapp.local',
    now(),
    now()
  ),
  (
    '52222222-2222-2222-2222-222222222222',
    '22222222-2222-2222-2222-222222222222',
    'MGR-DEMO-001',
    'Мария',
    'Руководитель',
    null,
    'Engineering Manager',
    '44444444-4444-4444-4444-444444444444',
    current_date,
    'active',
    'Asia/Yakutsk',
    'manager.demo@moneyapp.local',
    now(),
    now()
  ),
  (
    '53333333-3333-3333-3333-333333333333',
    '33333333-3333-3333-3333-333333333333',
    'HR-DEMO-001',
    'Анна',
    'HR',
    null,
    'HR Business Partner',
    '44444444-4444-4444-4444-444444444444',
    current_date,
    'active',
    'Asia/Yakutsk',
    'hr.demo@moneyapp.local',
    now(),
    now()
  )
on conflict (user_id) do update
set employee_no = excluded.employee_no,
    first_name = excluded.first_name,
    last_name = excluded.last_name,
    middle_name = excluded.middle_name,
    position_title = excluded.position_title,
    department_id = excluded.department_id,
    hire_date = excluded.hire_date,
    employment_status = excluded.employment_status,
    timezone = excluded.timezone,
    outlook_email = excluded.outlook_email,
    updated_at = now();

insert into manager_relations (
  id, employee_user_id, manager_user_id, relation_type, is_primary, created_at
)
values (
  '61111111-1111-1111-1111-111111111111',
  '11111111-1111-1111-1111-111111111111',
  '22222222-2222-2222-2222-222222222222',
  'line_manager',
  true,
  now()
)
on conflict (employee_user_id, manager_user_id, relation_type) do update
set is_primary = excluded.is_primary;

insert into user_roles (user_id, role_id, scope_type, scope_id, created_at)
select mapping.user_id, roles.id, 'global', null, now()
from (
  values
    ('11111111-1111-1111-1111-111111111111'::uuid, 'employee'),
    ('22222222-2222-2222-2222-222222222222'::uuid, 'manager'),
    ('33333333-3333-3333-3333-333333333333'::uuid, 'hr')
) as mapping(user_id, role_code)
join roles on roles.code = mapping.role_code
on conflict do nothing;

commit;
