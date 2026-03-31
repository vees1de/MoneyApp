-- 0029_course_intakes.sql
-- Наборы на курсы, заявки сотрудников, предложения внешних курсов

-- ============================================================
-- 1. Наборы (HR открывает набор на курс)
-- ============================================================
CREATE TABLE IF NOT EXISTS course_intakes (
    id              uuid        PRIMARY KEY,
    course_id       uuid        REFERENCES courses(id) ON DELETE SET NULL,
    title           varchar(500) NOT NULL,
    description     text,
    opened_by       uuid        NOT NULL REFERENCES users(id),
    approver_id     uuid        REFERENCES users(id),            -- опциональный менеджер-апрувер
    max_participants int,
    start_date      date,
    end_date        date,
    application_deadline timestamptz,
    status          varchar(30) NOT NULL DEFAULT 'open'
                    CHECK (status IN ('open','closed','canceled','completed')),
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_course_intakes_status     ON course_intakes (status);
CREATE INDEX idx_course_intakes_course_id  ON course_intakes (course_id);
CREATE INDEX idx_course_intakes_opened_by  ON course_intakes (opened_by);

-- ============================================================
-- 2. Заявки сотрудников на набор
-- ============================================================
CREATE TABLE IF NOT EXISTS course_applications (
    id                  uuid        PRIMARY KEY,
    intake_id           uuid        NOT NULL REFERENCES course_intakes(id) ON DELETE CASCADE,
    applicant_id        uuid        NOT NULL REFERENCES users(id),
    motivation          text,

    status              varchar(30) NOT NULL DEFAULT 'pending'
                        CHECK (status IN (
                            'pending',
                            'pending_manager',
                            'approved_by_manager',
                            'approved',
                            'rejected_by_manager',
                            'rejected_by_hr',
                            'withdrawn',
                            'enrolled'
                        )),

    manager_approver_id uuid        REFERENCES users(id),
    manager_comment     text,
    manager_decided_at  timestamptz,

    hr_approver_id      uuid        REFERENCES users(id),
    hr_comment          text,
    hr_decided_at       timestamptz,

    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),

    UNIQUE (intake_id, applicant_id)
);

CREATE INDEX idx_course_applications_intake     ON course_applications (intake_id);
CREATE INDEX idx_course_applications_applicant  ON course_applications (applicant_id);
CREATE INDEX idx_course_applications_status     ON course_applications (status);

-- ============================================================
-- 3. Предложения курсов от сотрудников
-- ============================================================
CREATE TABLE IF NOT EXISTS course_suggestions (
    id              uuid        PRIMARY KEY,
    suggested_by    uuid        NOT NULL REFERENCES users(id),
    title           varchar(500) NOT NULL,
    description     text,
    external_url    text,
    provider_name   varchar(255),
    price           numeric,
    price_currency  varchar(3)  NOT NULL DEFAULT 'RUB',
    duration_hours  numeric(8,2),

    approver_id     uuid        REFERENCES users(id),            -- менеджер-апрувер (опционально)
    status          varchar(30) NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','approved','rejected','intake_opened')),

    reviewed_by     uuid        REFERENCES users(id),
    review_comment  text,
    reviewed_at     timestamptz,

    intake_id       uuid        REFERENCES course_intakes(id),   -- если из предложения создан набор

    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_course_suggestions_suggested_by ON course_suggestions (suggested_by);
CREATE INDEX idx_course_suggestions_status       ON course_suggestions (status);

-- ============================================================
-- 4. Permissions
-- ============================================================
INSERT INTO permissions (id, code, module, action, description) VALUES
    (gen_random_uuid(), 'intakes.manage', 'intakes', 'manage', 'Управление наборами на курсы (HR)'),
    (gen_random_uuid(), 'intakes.apply',  'intakes', 'apply',  'Подача заявок на курсы (сотрудники)')
ON CONFLICT (code) DO NOTHING;
