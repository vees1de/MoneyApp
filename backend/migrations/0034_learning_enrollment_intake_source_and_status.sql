-- 0034_learning_enrollment_intake_source_and_status.sql
-- Приводим ограничения enrollments к фактической модели обучения через intake.

ALTER TABLE enrollments
    DROP CONSTRAINT IF EXISTS enrollments_source_check;

ALTER TABLE enrollments
    ADD CONSTRAINT enrollments_source_check
    CHECK (source IN ('self', 'assignment', 'external_request', 'program', 'intake'));

ALTER TABLE enrollments
    DROP CONSTRAINT IF EXISTS enrollments_status_check;

ALTER TABLE enrollments
    ADD CONSTRAINT enrollments_status_check
    CHECK (status IN ('not_started', 'enrolled', 'in_progress', 'completed', 'failed', 'canceled'));
