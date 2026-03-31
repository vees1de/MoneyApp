begin;

alter table users add column if not exists display_name varchar(150) null;
alter table users add column if not exists avatar_url text null;
alter table users add column if not exists timezone varchar(100) not null default 'Europe/Amsterdam';
alter table users add column if not exists base_currency varchar(3) not null default 'RUB';
alter table users add column if not exists onboarding_completed boolean not null default false;
alter table users add column if not exists weekly_review_weekday int not null default 1;
alter table users add column if not exists weekly_review_hour int not null default 18;

update users u
set display_name = nullif(trim(concat_ws(' ', ep.first_name, ep.last_name)), '')
from employee_profiles ep
where ep.user_id = u.id
  and (u.display_name is null or trim(u.display_name) = '');

create table if not exists profile_roles (
  id uuid primary key,
  code varchar(60) not null unique,
  name varchar(120) not null,
  description text null,
  sort_order int not null default 0,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create table if not exists user_profile_roles (
  user_id uuid not null references users(id) on delete cascade,
  profile_role_id uuid not null references profile_roles(id) on delete cascade,
  created_at timestamptz not null,
  primary key (user_id, profile_role_id)
);

alter table org_groups add column if not exists description text null;
alter table org_groups add column if not exists group_type varchar(30) not null default 'general';
alter table org_groups add column if not exists lead_user_id uuid null references users(id);
alter table org_groups add column if not exists created_by_user_id uuid null references users(id);

do $$
begin
  if not exists (
    select 1
    from pg_constraint
    where conname = 'org_groups_group_type_check'
  ) then
    alter table org_groups
      add constraint org_groups_group_type_check
      check (group_type in ('general', 'development_team'));
  end if;
end $$;

create index if not exists org_group_members_user_id_idx on org_group_members (user_id);
create index if not exists org_groups_group_type_idx on org_groups (group_type);
create index if not exists org_groups_lead_user_id_idx on org_groups (lead_user_id);
create index if not exists org_groups_created_by_user_id_idx on org_groups (created_by_user_id);

insert into profile_roles (id, code, name, description, sort_order, created_at, updated_at)
values
  ('dc65ab8e-4a33-4b61-93e4-f3e6c72a6c10', 'backend_developer', 'Backend-разработчик', 'Разрабатывает серверную логику, API и интеграции.', 10, now(), now()),
  ('8921cd47-9296-45f0-8ea8-aa84634a9352', 'frontend_developer', 'Frontend-разработчик', 'Отвечает за пользовательский интерфейс и клиентскую часть.', 20, now(), now()),
  ('d8dbffb0-9eb5-43bc-8d42-4a53753819bc', 'fullstack_developer', 'Fullstack-разработчик', 'Работает и с клиентской, и с серверной частью продукта.', 30, now(), now()),
  ('b9d3c768-4cf0-4149-b61e-f0dcd111d34b', 'mobile_developer', 'Мобильный разработчик', 'Разрабатывает мобильные приложения и связанные сервисы.', 40, now(), now()),
  ('c2d22f61-cf32-457a-b385-993c116f7e85', 'qa_engineer', 'QA-инженер', 'Отвечает за тестирование и качество продукта.', 50, now(), now()),
  ('181d576d-3158-497e-8f33-497f36d67036', 'devops_engineer', 'DevOps-инженер', 'Поддерживает инфраструктуру, CI/CD и эксплуатацию сервисов.', 60, now(), now()),
  ('d3053445-7e23-4bdf-bd17-6220e63ccad4', 'ui_ux_designer', 'UI/UX-дизайнер', 'Проектирует интерфейсы и пользовательский опыт.', 70, now(), now()),
  ('30f8ebf8-1708-4d72-98f6-cd9816faf2f3', 'product_manager', 'Product manager', 'Определяет продуктовую стратегию и приоритеты развития.', 80, now(), now()),
  ('7b50bbd5-a3cf-471e-a667-8efd4c080f67', 'project_manager', 'Project manager', 'Управляет сроками, ресурсами и delivery проекта.', 90, now(), now()),
  ('3d897fbc-4f33-44df-a96f-7d34b91d3207', 'business_analyst', 'Бизнес-аналитик', 'Собирает требования бизнеса и переводит их в задачи.', 100, now(), now()),
  ('3e4d75d0-7dad-4650-b773-5daff09734ff', 'system_analyst', 'Системный аналитик', 'Прорабатывает системные требования и интеграционные схемы.', 110, now(), now()),
  ('96dc4c5d-4a52-44bd-8d5c-c8eb041a45d4', 'tech_lead', 'Tech Lead', 'Отвечает за технические решения и инженерное качество команды.', 120, now(), now())
on conflict (code) do update
set name = excluded.name,
    description = excluded.description,
    sort_order = excluded.sort_order,
    updated_at = excluded.updated_at;

commit;
