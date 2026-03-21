begin;

create table if not exists finance_categories (
  id uuid primary key,
  user_id uuid null references users(id) on delete cascade,
  kind text not null,
  name text not null,
  color text null,
  icon text null,
  parent_id uuid null references finance_categories(id) on delete set null,
  is_system boolean not null default false,
  is_archived boolean not null default false,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  constraint finance_categories_kind_check check (kind in ('income', 'expense'))
);

alter table finance_transactions
  add constraint finance_transactions_category_id_fkey
  foreign key (category_id) references finance_categories(id) on delete set null;

create index if not exists finance_categories_user_id_idx on finance_categories (user_id);
create index if not exists finance_categories_kind_idx on finance_categories (kind);

insert into finance_categories (id, user_id, kind, name, color, icon, parent_id, is_system, is_archived, created_at, updated_at) values
  ('0c3518d6-9e34-4ab2-8e5c-4558b2fbf101', null, 'expense', 'Продукты', '#4F6F52', 'basket', null, true, false, now(), now()),
  ('0c3518d6-9e34-4ab2-8e5c-4558b2fbf102', null, 'expense', 'Кафе и рестораны', '#D07C54', 'coffee', null, true, false, now(), now()),
  ('0c3518d6-9e34-4ab2-8e5c-4558b2fbf103', null, 'expense', 'Транспорт', '#5F6CAF', 'bus', null, true, false, now(), now()),
  ('0c3518d6-9e34-4ab2-8e5c-4558b2fbf104', null, 'expense', 'Такси', '#6D597A', 'car', null, true, false, now(), now()),
  ('0c3518d6-9e34-4ab2-8e5c-4558b2fbf105', null, 'expense', 'Подписки', '#355070', 'repeat', null, true, false, now(), now()),
  ('0c3518d6-9e34-4ab2-8e5c-4558b2fbf106', null, 'expense', 'Жильё', '#3A5A40', 'home', null, true, false, now(), now()),
  ('0c3518d6-9e34-4ab2-8e5c-4558b2fbf107', null, 'expense', 'Коммунальные услуги', '#457B9D', 'bolt', null, true, false, now(), now()),
  ('0c3518d6-9e34-4ab2-8e5c-4558b2fbf108', null, 'expense', 'Связь / интернет', '#264653', 'wifi', null, true, false, now(), now()),
  ('0c3518d6-9e34-4ab2-8e5c-4558b2fbf109', null, 'expense', 'Здоровье', '#B56576', 'heart', null, true, false, now(), now()),
  ('0c3518d6-9e34-4ab2-8e5c-4558b2fbf110', null, 'expense', 'Одежда', '#A06CD5', 'shirt', null, true, false, now(), now()),
  ('0c3518d6-9e34-4ab2-8e5c-4558b2fbf111', null, 'expense', 'Образование', '#F4A261', 'book', null, true, false, now(), now()),
  ('0c3518d6-9e34-4ab2-8e5c-4558b2fbf112', null, 'expense', 'Подарки', '#E76F51', 'gift', null, true, false, now(), now()),
  ('0c3518d6-9e34-4ab2-8e5c-4558b2fbf113', null, 'expense', 'Развлечения', '#9C6644', 'sparkles', null, true, false, now(), now()),
  ('0c3518d6-9e34-4ab2-8e5c-4558b2fbf114', null, 'expense', 'Путешествия', '#2A9D8F', 'plane', null, true, false, now(), now()),
  ('0c3518d6-9e34-4ab2-8e5c-4558b2fbf115', null, 'expense', 'Техника', '#577590', 'laptop', null, true, false, now(), now()),
  ('0c3518d6-9e34-4ab2-8e5c-4558b2fbf116', null, 'expense', 'Дом', '#7F5539', 'sofa', null, true, false, now(), now()),
  ('0c3518d6-9e34-4ab2-8e5c-4558b2fbf117', null, 'expense', 'Наличные / снятие', '#8D99AE', 'wallet', null, true, false, now(), now()),
  ('0c3518d6-9e34-4ab2-8e5c-4558b2fbf118', null, 'expense', 'Прочее', '#6B705C', 'dots', null, true, false, now(), now()),
  ('0c3518d6-9e34-4ab2-8e5c-4558b2fbf119', null, 'income', 'Зарплата', '#2D6A4F', 'briefcase', null, true, false, now(), now()),
  ('0c3518d6-9e34-4ab2-8e5c-4558b2fbf120', null, 'income', 'Фриланс', '#1D3557', 'code', null, true, false, now(), now()),
  ('0c3518d6-9e34-4ab2-8e5c-4558b2fbf121', null, 'income', 'Подарки / перевод', '#E9C46A', 'banknote', null, true, false, now(), now()),
  ('0c3518d6-9e34-4ab2-8e5c-4558b2fbf122', null, 'income', 'Кэшбэк', '#43AA8B', 'coins', null, true, false, now(), now()),
  ('0c3518d6-9e34-4ab2-8e5c-4558b2fbf123', null, 'income', 'Продажи', '#F8961E', 'store', null, true, false, now(), now()),
  ('0c3518d6-9e34-4ab2-8e5c-4558b2fbf124', null, 'income', 'Прочий доход', '#277DA1', 'plus-circle', null, true, false, now(), now())
on conflict (id) do nothing;

commit;
