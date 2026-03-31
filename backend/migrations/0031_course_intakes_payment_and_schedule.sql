-- 0031_course_intakes_payment_and_schedule.sql
-- Цена и длительность наборов, статус оплаты заявок

ALTER TABLE course_intakes
    ADD COLUMN IF NOT EXISTS duration_weeks int,
    ADD COLUMN IF NOT EXISTS price numeric,
    ADD COLUMN IF NOT EXISTS price_currency varchar(3);

ALTER TABLE course_intakes
    ALTER COLUMN price_currency SET DEFAULT 'RUB';

ALTER TABLE course_intakes
    DROP CONSTRAINT IF EXISTS course_intakes_duration_weeks_check;

ALTER TABLE course_intakes
    ADD CONSTRAINT course_intakes_duration_weeks_check
        CHECK (duration_weeks IS NULL OR duration_weeks > 0);

UPDATE course_intakes ci
SET
    price = c.price,
    price_currency = COALESCE(c.price_currency, 'RUB')
FROM courses c
WHERE ci.course_id = c.id
  AND ci.price IS NULL;

UPDATE course_intakes
SET price_currency = 'RUB'
WHERE price_currency IS NULL;

ALTER TABLE course_applications
    ADD COLUMN IF NOT EXISTS payment_status varchar(20) NOT NULL DEFAULT 'unpaid';

ALTER TABLE course_applications
    DROP CONSTRAINT IF EXISTS course_applications_payment_status_check;

ALTER TABLE course_applications
    ADD CONSTRAINT course_applications_payment_status_check
        CHECK (payment_status IN ('paid', 'unpaid'));

CREATE INDEX IF NOT EXISTS idx_course_applications_payment_status
    ON course_applications (payment_status);
