begin;

-- scope_type and scope_id must be nullable for global roles (no scope)
alter table user_roles alter column scope_type drop not null;
alter table user_roles alter column scope_id drop not null;

commit;
