begin;

insert into role_permissions (role_id, permission_id)
select r.id, p.id
from (
  values
    ('manager', 'external_requests.create'),
    ('manager', 'external_requests.read_own'),
    ('hr', 'external_requests.create'),
    ('hr', 'external_requests.read_own'),
    ('trainer', 'external_requests.create'),
    ('trainer', 'external_requests.read_own'),
    ('admin', 'external_requests.create'),
    ('admin', 'external_requests.read_own')
) mapping(role_code, permission_code)
join roles r on r.code = mapping.role_code
join permissions p on p.code = mapping.permission_code
left join role_permissions rp on rp.role_id = r.id and rp.permission_id = p.id
where rp.role_id is null;

commit;
