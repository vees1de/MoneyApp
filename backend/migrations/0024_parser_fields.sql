begin;

alter table courses
  add column if not exists external_url text null,
  add column if not exists price numeric(12,2) null,
  add column if not exists price_currency varchar(10) null default 'RUB',
  add column if not exists next_start_date date null;

commit;
