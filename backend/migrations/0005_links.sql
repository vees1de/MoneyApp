begin;

create table if not exists entity_links (
  id uuid primary key,
  user_id uuid not null references users(id) on delete cascade,
  source_type text not null,
  source_id uuid not null,
  target_type text not null,
  target_id uuid not null,
  relation text not null,
  meta jsonb not null default '{}'::jsonb,
  created_at timestamptz not null
);

create index if not exists entity_links_source_idx on entity_links (user_id, source_type, source_id);
create index if not exists entity_links_target_idx on entity_links (user_id, target_type, target_id);

commit;
