begin;

create table if not exists file_attachments (
  id uuid primary key,
  storage_provider varchar(30) not null check (storage_provider in ('s3', 'local', 'minio')),
  storage_key text not null,
  original_name text not null,
  mime_type varchar(100) not null,
  size_bytes bigint not null,
  uploaded_by uuid null references users(id),
  created_at timestamptz not null
);

create table if not exists certificates (
  id uuid primary key,
  user_id uuid not null references users(id),
  course_id uuid null references courses(id),
  enrollment_id uuid null references enrollments(id),
  certificate_no varchar(255) null,
  issued_by varchar(255) null,
  issued_at date null,
  expires_at date null,
  status varchar(30) not null check (status in ('uploaded', 'under_review', 'verified', 'rejected', 'expired')),
  file_id uuid not null references file_attachments(id),
  uploaded_at timestamptz not null,
  verified_at timestamptz null,
  verified_by uuid null references users(id),
  notes text null
);

create table if not exists certificate_verifications (
  id uuid primary key,
  certificate_id uuid not null references certificates(id) on delete cascade,
  action varchar(30) not null check (action in ('submit', 'verify', 'reject', 'request_revision')),
  performed_by uuid not null references users(id),
  comment text null,
  created_at timestamptz not null
);

commit;
