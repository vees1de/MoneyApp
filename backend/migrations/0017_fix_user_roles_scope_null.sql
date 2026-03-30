begin;

do $$
begin
  if exists (
    select 1
    from pg_constraint
    where conname = 'user_roles_pkey'
  ) then
    alter table user_roles drop constraint user_roles_pkey;
  end if;
end $$;

do $$
begin
  if not exists (
    select 1
    from pg_constraint
    where conname = 'user_roles_unique_scope'
  ) then
    alter table user_roles
      add constraint user_roles_unique_scope
      unique nulls not distinct (user_id, role_id, scope_type, scope_id);
  end if;
end $$;

commit;
