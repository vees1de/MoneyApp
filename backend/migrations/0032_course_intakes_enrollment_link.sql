-- 0032_course_intakes_enrollment_link.sql
-- Связываем intake applications с enrollments для старта/завершения/сертификатов

ALTER TABLE course_applications
    ADD COLUMN IF NOT EXISTS enrollment_id uuid REFERENCES enrollments(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_course_applications_enrollment_id
    ON course_applications (enrollment_id);
