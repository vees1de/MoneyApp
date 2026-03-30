begin;

-- Rename ip_address -> ip and change type to inet (as expected by identity module)
-- Strip port suffix (e.g. "172.18.0.1:51734" -> "172.18.0.1")
update sessions set ip_address = split_part(ip_address, ':', 1)
  where ip_address is not null and ip_address like '%:%';

-- Null out anything still not valid inet
update sessions set ip_address = null
  where ip_address is not null
    and ip_address !~ '^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$';

alter table sessions rename column ip_address to ip;
alter table sessions alter column ip type inet using ip::inet;

commit;
