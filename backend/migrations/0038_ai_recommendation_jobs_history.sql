begin;

create index if not exists ai_recommendation_jobs_user_created_idx
  on ai_recommendation_jobs (user_id, created_at desc, updated_at desc);

commit;
