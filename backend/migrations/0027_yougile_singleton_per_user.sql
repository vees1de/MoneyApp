begin;

update integration_yougile_connections as connection
set status = 'revoked',
    updated_at = now()
from (
  select id,
         row_number() over (
           partition by created_by
           order by
             case when status = 'active' then 0 else 1 end,
             updated_at desc,
             created_at desc,
             id desc
         ) as row_no
  from integration_yougile_connections
  where status <> 'revoked'
) ranked
where connection.id = ranked.id
  and ranked.row_no > 1;

alter table integration_yougile_connections
  drop constraint if exists integration_yougile_connections_company_id_yougile_company_id_key;

alter table integration_yougile_connections
  drop constraint if exists integration_yougile_connections_created_by_yougile_company_id_key;

create unique index if not exists idx_yougile_connections_singleton_per_user
  on integration_yougile_connections (created_by)
  where status <> 'revoked';

commit;
