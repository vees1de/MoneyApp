begin;

insert into roles (id, code, name, description, is_system)
values
  ('00000000-0000-0000-0000-000000000101', 'employee', 'Employee', 'Employee self-service role', true),
  ('00000000-0000-0000-0000-000000000102', 'manager', 'Manager', 'Manager approval and team visibility role', true),
  ('00000000-0000-0000-0000-000000000103', 'hr', 'HR/L&D', 'Learning and development role', true),
  ('00000000-0000-0000-0000-000000000104', 'trainer', 'Trainer', 'Internal trainer role', true),
  ('00000000-0000-0000-0000-000000000105', 'admin', 'Administrator', 'System administrator role', true)
on conflict (code) do nothing;

insert into permissions (id, code, module, action, description)
values
  ('00000000-0000-0000-0000-000000000201', 'users.read', 'users', 'read', 'Read users'),
  ('00000000-0000-0000-0000-000000000202', 'users.write', 'users', 'write', 'Manage users'),
  ('00000000-0000-0000-0000-000000000203', 'roles.manage', 'roles', 'manage', 'Manage roles'),
  ('00000000-0000-0000-0000-000000000204', 'courses.read', 'courses', 'read', 'Read courses'),
  ('00000000-0000-0000-0000-000000000205', 'courses.write', 'courses', 'write', 'Manage courses'),
  ('00000000-0000-0000-0000-000000000206', 'courses.assign', 'courses', 'assign', 'Assign courses'),
  ('00000000-0000-0000-0000-000000000207', 'enrollments.read', 'enrollments', 'read', 'Read enrollments'),
  ('00000000-0000-0000-0000-000000000208', 'enrollments.manage', 'enrollments', 'manage', 'Manage enrollments'),
  ('00000000-0000-0000-0000-000000000209', 'external_requests.create', 'external_requests', 'create', 'Create external requests'),
  ('00000000-0000-0000-0000-000000000210', 'external_requests.read_own', 'external_requests', 'read_own', 'Read own external requests'),
  ('00000000-0000-0000-0000-000000000211', 'external_requests.read_all', 'external_requests', 'read_all', 'Read all external requests'),
  ('00000000-0000-0000-0000-000000000212', 'external_requests.approve_manager', 'external_requests', 'approve_manager', 'Manager approval'),
  ('00000000-0000-0000-0000-000000000213', 'external_requests.approve_hr', 'external_requests', 'approve_hr', 'HR approval'),
  ('00000000-0000-0000-0000-000000000214', 'certificates.verify', 'certificates', 'verify', 'Verify certificates'),
  ('00000000-0000-0000-0000-000000000215', 'programs.manage', 'programs', 'manage', 'Manage programs'),
  ('00000000-0000-0000-0000-000000000216', 'analytics.read_hr', 'analytics', 'read_hr', 'Read HR analytics'),
  ('00000000-0000-0000-0000-000000000217', 'analytics.read_manager', 'analytics', 'read_manager', 'Read manager analytics'),
  ('00000000-0000-0000-0000-000000000218', 'notifications.manage', 'notifications', 'manage', 'Manage notifications'),
  ('00000000-0000-0000-0000-000000000219', 'settings.manage', 'settings', 'manage', 'Manage settings'),
  ('00000000-0000-0000-0000-000000000220', 'audit.read', 'audit', 'read', 'Read audit logs')
on conflict (code) do nothing;

insert into role_permissions (role_id, permission_id)
select role_id, permission_id
from (
  values
    ('employee', 'courses.read'),
    ('employee', 'enrollments.read'),
    ('employee', 'external_requests.create'),
    ('employee', 'external_requests.read_own'),
    ('manager', 'courses.read'),
    ('manager', 'enrollments.read'),
    ('manager', 'external_requests.read_all'),
    ('manager', 'external_requests.approve_manager'),
    ('manager', 'analytics.read_manager'),
    ('hr', 'users.read'),
    ('hr', 'courses.read'),
    ('hr', 'courses.write'),
    ('hr', 'courses.assign'),
    ('hr', 'enrollments.read'),
    ('hr', 'enrollments.manage'),
    ('hr', 'external_requests.read_all'),
    ('hr', 'external_requests.approve_hr'),
    ('hr', 'certificates.verify'),
    ('hr', 'analytics.read_hr'),
    ('hr', 'notifications.manage'),
    ('trainer', 'courses.read'),
    ('trainer', 'programs.manage'),
    ('admin', 'users.read'),
    ('admin', 'users.write'),
    ('admin', 'roles.manage'),
    ('admin', 'courses.read'),
    ('admin', 'courses.write'),
    ('admin', 'courses.assign'),
    ('admin', 'enrollments.read'),
    ('admin', 'enrollments.manage'),
    ('admin', 'external_requests.read_all'),
    ('admin', 'external_requests.approve_manager'),
    ('admin', 'external_requests.approve_hr'),
    ('admin', 'certificates.verify'),
    ('admin', 'programs.manage'),
    ('admin', 'analytics.read_hr'),
    ('admin', 'analytics.read_manager'),
    ('admin', 'notifications.manage'),
    ('admin', 'settings.manage'),
    ('admin', 'audit.read')
) mapping(role_code, permission_code)
join roles r on r.code = mapping.role_code
join permissions p on p.code = mapping.permission_code
left join role_permissions rp on rp.role_id = r.id and rp.permission_id = p.id
where rp.role_id is null;

commit;
