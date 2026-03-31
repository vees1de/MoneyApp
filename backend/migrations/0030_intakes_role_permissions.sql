-- 0030_intakes_role_permissions.sql
-- Привязка пермишенов intakes.manage / intakes.apply к ролям
-- (миграция 0029 создала пермишены, но не привязала к ролям)

BEGIN;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM (
  VALUES
    ('hr',      'intakes.manage'),
    ('hr',      'intakes.apply'),
    ('admin',   'intakes.manage'),
    ('admin',   'intakes.apply'),
    ('manager', 'intakes.apply'),
    ('trainer', 'intakes.apply')
) mapping(role_code, permission_code)
JOIN roles r ON r.code = mapping.role_code
JOIN permissions p ON p.code = mapping.permission_code
LEFT JOIN role_permissions rp ON rp.role_id = r.id AND rp.permission_id = p.id
WHERE rp.role_id IS NULL;

COMMIT;
