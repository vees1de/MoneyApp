begin;

alter table integration_yougile_connections
  drop constraint if exists integration_yougile_connections_company_id_yougile_company_id_key;

alter table integration_yougile_connections
  drop constraint if exists integration_yougile_connections_created_by_yougile_company_id_key;

alter table integration_yougile_connections
  drop constraint if exists integration_yougile_connectio_company_id_yougile_company_id_key;

alter table integration_yougile_connections
  drop constraint if exists integration_yougile_connectio_created_by_yougile_company_id_key;

drop index if exists idx_yougile_connections_singleton_per_user;

create unique index if not exists idx_yougile_connections_singleton_per_user
  on integration_yougile_connections (created_by)
  where status <> 'revoked';

commit;
