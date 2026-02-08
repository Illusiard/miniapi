create table if not exists notes (
  id bigserial primary key,
  title text not null,
  content text not null,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists idx_notes_created_at on notes (created_at desc);
