begin;

with desired_tags (id, name, slug) as (
  values
    ('9bfecb5b-1105-45da-8a59-b3072f788fe8'::uuid, 'Язык Go', 'go'),
    ('1a567d31-d311-44ad-9868-99fb35705cc6'::uuid, 'PRO', 'pro'),
    ('c9d5a1ef-3735-4397-8fb3-1d4251a1e0a6'::uuid, 'Программирование', 'programming'),
    ('2d4e93c2-4610-4f63-a79f-91ae7ab1360e'::uuid, 'Бэкенд', 'backend'),
    ('65ce6794-9dbb-4bbf-ae82-e22d9267a2d2'::uuid, 'С трудоустройством', 's-trudoustrojstvom'),
    ('bee31657-5918-4944-9260-57fc1ea5b11c'::uuid, 'SQL', 'sql'),
    ('8a4c4e6a-a660-4a86-b76f-00ea216b336a'::uuid, 'Linux', 'linux'),
    ('32ea1669-be79-41ac-88a1-b3a4123fe9d1'::uuid, 'Базы данных', 'bazy-dannykh'),
    ('bfc7a748-2882-400e-9457-eaed8fa86cc6'::uuid, 'С нуля', 'start'),
    ('907ac6b3-e5c6-4e49-baf5-31dcbbcc680c'::uuid, 'Веб-разработка', 'web-razrabotka'),
    ('44040d8a-e19c-4974-82ce-8a9dbe37d13b'::uuid, 'Backend-разработка', 'backend-razrabotka'),
    ('07b0cac5-3626-4d23-8104-8c441da0ca98'::uuid, 'REST API', 'rest-api'),
    ('129def95-7339-43c7-ac14-d7031b992d21'::uuid, 'Алгоритмы', 'algoritms-courses'),
    ('63518b16-a396-47f4-aa4e-280c9000ac8f'::uuid, 'Маркетинг', 'marketing'),
    ('9b49b11b-2a59-4dbc-a36e-7d6ffc9ea433'::uuid, 'Digital-маркетинг', 'digital-marketing')
)
insert into skill_tags (id, name, slug)
select id, name, slug
from desired_tags
on conflict (slug) do update
set name = excluded.name;

insert into courses (
  id,
  type,
  source_type,
  title,
  slug,
  short_description,
  description,
  level,
  duration_hours,
  language,
  is_mandatory_default,
  status,
  created_at,
  updated_at,
  published_at
)
values
  (
    'c73788c9-8656-42cf-b290-9ad1d6cc4f74',
    'external',
    'imported',
    'Основы Go',
    'go-basics',
    'Курс для тех, у кого есть опыт разработки на других языках.',
    'Курс для тех, у кого есть опыт разработки на других языках. Освоите базу, чтобы дальше развиваться в разработке на Go.',
    'beginner',
    null,
    'ru',
    false,
    'published',
    now(),
    now(),
    now()
  ),
  (
    '9752fad5-3e3d-4b18-a8f5-32321468b5bd',
    'external',
    'imported',
    'Go-разработчик с нуля',
    'go-developer-basic',
    'Курс по разработке на языке Go для начинающих.',
    'Курс по разработке на языке Go для начинающих.',
    'beginner',
    null,
    'ru',
    false,
    'published',
    now(),
    now(),
    now()
  ),
  (
    '9663b609-0da2-47e0-94ef-1675ecf03c0f',
    'external',
    'imported',
    'Продвинутый Go-разработчик',
    'go-advanced',
    'За 6 месяцев выйдете на новый уровень разработки на Golang.',
    'За 6 месяцев выйдете на новый уровень разработки на Golang.',
    'advanced',
    null,
    'ru',
    false,
    'published',
    now(),
    now(),
    now()
  ),
  (
    '35ed6b5e-d5e9-43e8-bbc0-425a25293b39',
    'external',
    'imported',
    'Продуктовый маркетолог',
    'product-marketing-manager',
    'Онлайн-курс по обучению на продуктового маркетолога.',
    'Онлайн-курс «Product marketing manager»: обучение на продуктового маркетолога.',
    'beginner',
    null,
    'ru',
    false,
    'published',
    now(),
    now(),
    now()
  )
on conflict (id) do update
set type = excluded.type,
    source_type = excluded.source_type,
    title = excluded.title,
    slug = excluded.slug,
    short_description = excluded.short_description,
    description = excluded.description,
    level = excluded.level,
    duration_hours = excluded.duration_hours,
    language = excluded.language,
    is_mandatory_default = excluded.is_mandatory_default,
    status = excluded.status,
    updated_at = now(),
    published_at = excluded.published_at,
    archived_at = null;

with desired_links (course_id, tag_slug) as (
  values
    ('c73788c9-8656-42cf-b290-9ad1d6cc4f74'::uuid, 'go'),
    ('c73788c9-8656-42cf-b290-9ad1d6cc4f74'::uuid, 'pro'),
    ('c73788c9-8656-42cf-b290-9ad1d6cc4f74'::uuid, 'programming'),
    ('9752fad5-3e3d-4b18-a8f5-32321468b5bd'::uuid, 'backend'),
    ('9752fad5-3e3d-4b18-a8f5-32321468b5bd'::uuid, 's-trudoustrojstvom'),
    ('9752fad5-3e3d-4b18-a8f5-32321468b5bd'::uuid, 'go'),
    ('9752fad5-3e3d-4b18-a8f5-32321468b5bd'::uuid, 'programming'),
    ('9752fad5-3e3d-4b18-a8f5-32321468b5bd'::uuid, 'sql'),
    ('9752fad5-3e3d-4b18-a8f5-32321468b5bd'::uuid, 'linux'),
    ('9752fad5-3e3d-4b18-a8f5-32321468b5bd'::uuid, 'bazy-dannykh'),
    ('9752fad5-3e3d-4b18-a8f5-32321468b5bd'::uuid, 'start'),
    ('9752fad5-3e3d-4b18-a8f5-32321468b5bd'::uuid, 'web-razrabotka'),
    ('9663b609-0da2-47e0-94ef-1675ecf03c0f'::uuid, 'backend-razrabotka'),
    ('9663b609-0da2-47e0-94ef-1675ecf03c0f'::uuid, 'rest-api'),
    ('9663b609-0da2-47e0-94ef-1675ecf03c0f'::uuid, 'algoritms-courses'),
    ('9663b609-0da2-47e0-94ef-1675ecf03c0f'::uuid, 'go'),
    ('9663b609-0da2-47e0-94ef-1675ecf03c0f'::uuid, 'pro'),
    ('9663b609-0da2-47e0-94ef-1675ecf03c0f'::uuid, 'programming'),
    ('9663b609-0da2-47e0-94ef-1675ecf03c0f'::uuid, 'bazy-dannykh'),
    ('9663b609-0da2-47e0-94ef-1675ecf03c0f'::uuid, 'backend'),
    ('35ed6b5e-d5e9-43e8-bbc0-425a25293b39'::uuid, 'marketing'),
    ('35ed6b5e-d5e9-43e8-bbc0-425a25293b39'::uuid, 'start'),
    ('35ed6b5e-d5e9-43e8-bbc0-425a25293b39'::uuid, 's-trudoustrojstvom'),
    ('35ed6b5e-d5e9-43e8-bbc0-425a25293b39'::uuid, 'digital-marketing')
)
insert into course_skill_tags (course_id, skill_tag_id)
select l.course_id, t.id
from desired_links l
join skill_tags t on t.slug = l.tag_slug
on conflict do nothing;

commit;
