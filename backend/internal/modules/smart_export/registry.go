package smart_export

// ColumnDef describes a single exportable column for a given data source.
type ColumnDef struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Type     string `json:"type"` // string, number, date, currency
	Default  bool   `json:"default"`
	SQLExpr  string `json:"-"`
	SQLAlias string `json:"-"`
}

// SourceDef describes a data source available for export.
type SourceDef struct {
	Key     string      `json:"key"`
	Label   string      `json:"label"`
	Columns []ColumnDef `json:"columns"`
	// FromSQL is the base FROM + JOINs clause for this source.
	FromSQL string `json:"-"`
}

// allSources returns all registered data sources with their column definitions.
func allSources() []SourceDef {
	return []SourceDef{
		sourceApplications(),
		sourceIntakes(),
		sourceSuggestions(),
		sourceCourseRequests(),
		sourceEnrollments(),
		sourceCourses(),
		sourceEmployees(),
	}
}

func getSource(key string) *SourceDef {
	for _, s := range allSources() {
		if s.Key == key {
			return &s
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Source: applications (заявки на наборы)
// ---------------------------------------------------------------------------

func sourceApplications() SourceDef {
	return SourceDef{
		Key:   "applications",
		Label: "Заявки на курсы",
		FromSQL: `course_applications ca
			JOIN course_intakes ci ON ci.id = ca.intake_id
			JOIN users u ON u.id = ca.applicant_id
			LEFT JOIN employee_profiles ep ON ep.user_id = u.id
			LEFT JOIN departments d ON d.id = ep.department_id
			LEFT JOIN courses c ON c.id = ci.course_id
			LEFT JOIN providers p ON p.id = c.provider_id
			LEFT JOIN employee_profiles mgr ON mgr.user_id = ca.manager_approver_id
			LEFT JOIN employee_profiles hr ON hr.user_id = ca.hr_approver_id`,
		Columns: []ColumnDef{
			{Key: "application_id", Label: "ID заявки", Type: "string", Default: false,
				SQLExpr: "ca.id", SQLAlias: "application_id"},
			{Key: "intake_id", Label: "ID набора", Type: "string", Default: false,
				SQLExpr: "ca.intake_id", SQLAlias: "intake_id"},
			{Key: "employee_name", Label: "ФИО сотрудника", Type: "string", Default: true,
				SQLExpr: "coalesce(ep.last_name || ' ' || ep.first_name || coalesce(' ' || ep.middle_name, ''), u.email)", SQLAlias: "employee_name"},
			{Key: "employee_email", Label: "Email", Type: "string", Default: true,
				SQLExpr: "u.email", SQLAlias: "employee_email"},
			{Key: "department", Label: "Отдел", Type: "string", Default: true,
				SQLExpr: "d.name", SQLAlias: "department"},
			{Key: "position", Label: "Должность", Type: "string", Default: false,
				SQLExpr: "ep.position_title", SQLAlias: "position"},
			{Key: "intake_title", Label: "Набор", Type: "string", Default: true,
				SQLExpr: "ci.title", SQLAlias: "intake_title"},
			{Key: "course_title", Label: "Курс", Type: "string", Default: true,
				SQLExpr: "c.title", SQLAlias: "course_title"},
			{Key: "course_provider", Label: "Провайдер", Type: "string", Default: false,
				SQLExpr: "p.name", SQLAlias: "course_provider"},
			{Key: "course_price", Label: "Стоимость", Type: "currency", Default: true,
				SQLExpr: "c.price", SQLAlias: "course_price"},
			{Key: "price_currency", Label: "Валюта", Type: "string", Default: false,
				SQLExpr: "c.price_currency", SQLAlias: "price_currency"},
			{Key: "status", Label: "Статус заявки", Type: "string", Default: true,
				SQLExpr: "ca.status", SQLAlias: "status"},
			{Key: "payment_status", Label: "Статус оплаты", Type: "string", Default: false,
				SQLExpr: "ca.payment_status", SQLAlias: "payment_status"},
			{Key: "enrollment_status", Label: "Статус обучения", Type: "string", Default: false,
				SQLExpr: "coalesce((SELECT e.status FROM enrollments e WHERE e.id = ca.enrollment_id), '')", SQLAlias: "enrollment_status"},
			{Key: "certificate_status", Label: "Статус сертификата", Type: "string", Default: false,
				SQLExpr: "coalesce((SELECT c.status FROM certificates c WHERE c.enrollment_id = ca.enrollment_id ORDER BY c.uploaded_at DESC LIMIT 1), '')", SQLAlias: "certificate_status"},
			{Key: "certificate_uploaded_at", Label: "Загрузка сертификата", Type: "date", Default: false,
				SQLExpr: "(SELECT c.uploaded_at FROM certificates c WHERE c.enrollment_id = ca.enrollment_id ORDER BY c.uploaded_at DESC LIMIT 1)", SQLAlias: "certificate_uploaded_at"},
			{Key: "motivation", Label: "Мотивация", Type: "string", Default: false,
				SQLExpr: "ca.motivation", SQLAlias: "motivation"},
			{Key: "manager_approver", Label: "Руководитель", Type: "string", Default: false,
				SQLExpr: "coalesce(mgr.last_name || ' ' || mgr.first_name, '')", SQLAlias: "manager_approver"},
			{Key: "manager_comment", Label: "Комм. руководителя", Type: "string", Default: false,
				SQLExpr: "ca.manager_comment", SQLAlias: "manager_comment"},
			{Key: "manager_decided_at", Label: "Решение руководителя", Type: "date", Default: false,
				SQLExpr: "ca.manager_decided_at", SQLAlias: "manager_decided_at"},
			{Key: "hr_approver", Label: "HR", Type: "string", Default: false,
				SQLExpr: "coalesce(hr.last_name || ' ' || hr.first_name, '')", SQLAlias: "hr_approver"},
			{Key: "hr_comment", Label: "Комм. HR", Type: "string", Default: false,
				SQLExpr: "ca.hr_comment", SQLAlias: "hr_comment"},
			{Key: "hr_decided_at", Label: "Решение HR", Type: "date", Default: false,
				SQLExpr: "ca.hr_decided_at", SQLAlias: "hr_decided_at"},
			{Key: "applied_at", Label: "Дата подачи", Type: "date", Default: true,
				SQLExpr: "ca.created_at", SQLAlias: "applied_at"},
			{Key: "updated_at", Label: "Последнее обновление", Type: "date", Default: false,
				SQLExpr: "ca.updated_at", SQLAlias: "updated_at"},
			{Key: "duration_hours", Label: "Длительность (ч)", Type: "number", Default: false,
				SQLExpr: "c.duration_hours", SQLAlias: "duration_hours"},
			{Key: "course_level", Label: "Уровень курса", Type: "string", Default: false,
				SQLExpr: "c.level", SQLAlias: "course_level"},
		},
	}
}

// ---------------------------------------------------------------------------
// Source: intakes (наборы)
// ---------------------------------------------------------------------------

func sourceIntakes() SourceDef {
	return SourceDef{
		Key:   "intakes",
		Label: "Наборы на курсы",
		FromSQL: `course_intakes ci
			LEFT JOIN courses c ON c.id = ci.course_id
			LEFT JOIN providers p ON p.id = c.provider_id
			JOIN users u ON u.id = ci.opened_by
			LEFT JOIN employee_profiles ep ON ep.user_id = u.id
			LEFT JOIN employee_profiles apr ON apr.user_id = ci.approver_id`,
		Columns: []ColumnDef{
			{Key: "title", Label: "Название набора", Type: "string", Default: true,
				SQLExpr: "ci.title", SQLAlias: "title"},
			{Key: "course_title", Label: "Курс", Type: "string", Default: true,
				SQLExpr: "c.title", SQLAlias: "course_title"},
			{Key: "course_provider", Label: "Провайдер", Type: "string", Default: false,
				SQLExpr: "p.name", SQLAlias: "course_provider"},
			{Key: "course_price", Label: "Стоимость", Type: "currency", Default: true,
				SQLExpr: "c.price", SQLAlias: "course_price"},
			{Key: "status", Label: "Статус", Type: "string", Default: true,
				SQLExpr: "ci.status", SQLAlias: "status"},
			{Key: "opened_by", Label: "Открыл", Type: "string", Default: true,
				SQLExpr: "coalesce(ep.last_name || ' ' || ep.first_name, u.email)", SQLAlias: "opened_by"},
			{Key: "approver", Label: "Апрувер", Type: "string", Default: false,
				SQLExpr: "coalesce(apr.last_name || ' ' || apr.first_name, '')", SQLAlias: "approver"},
			{Key: "max_participants", Label: "Макс. участников", Type: "number", Default: false,
				SQLExpr: "ci.max_participants", SQLAlias: "max_participants"},
			{Key: "application_count", Label: "Кол-во заявок", Type: "number", Default: true,
				SQLExpr: "(SELECT count(*) FROM course_applications WHERE intake_id = ci.id)", SQLAlias: "application_count"},
			{Key: "approved_count", Label: "Одобрено", Type: "number", Default: true,
				SQLExpr: "(SELECT count(*) FROM course_applications WHERE intake_id = ci.id AND status IN ('approved','enrolled'))", SQLAlias: "approved_count"},
			{Key: "start_date", Label: "Дата начала", Type: "date", Default: true,
				SQLExpr: "ci.start_date", SQLAlias: "start_date"},
			{Key: "end_date", Label: "Дата окончания", Type: "date", Default: false,
				SQLExpr: "ci.end_date", SQLAlias: "end_date"},
			{Key: "application_deadline", Label: "Дедлайн подачи", Type: "date", Default: false,
				SQLExpr: "ci.application_deadline", SQLAlias: "application_deadline"},
			{Key: "created_at", Label: "Дата создания", Type: "date", Default: false,
				SQLExpr: "ci.created_at", SQLAlias: "created_at"},
		},
	}
}

// ---------------------------------------------------------------------------
// Source: suggestions (предложения курсов)
// ---------------------------------------------------------------------------

func sourceSuggestions() SourceDef {
	return SourceDef{
		Key:   "suggestions",
		Label: "Предложения курсов",
		FromSQL: `course_suggestions cs
			JOIN users u ON u.id = cs.suggested_by
			LEFT JOIN employee_profiles ep ON ep.user_id = u.id
			LEFT JOIN departments d ON d.id = ep.department_id
			LEFT JOIN employee_profiles rv ON rv.user_id = cs.reviewed_by`,
		Columns: []ColumnDef{
			{Key: "employee_name", Label: "Предложил", Type: "string", Default: true,
				SQLExpr: "coalesce(ep.last_name || ' ' || ep.first_name, u.email)", SQLAlias: "employee_name"},
			{Key: "department", Label: "Отдел", Type: "string", Default: true,
				SQLExpr: "d.name", SQLAlias: "department"},
			{Key: "title", Label: "Название курса", Type: "string", Default: true,
				SQLExpr: "cs.title", SQLAlias: "title"},
			{Key: "external_url", Label: "Ссылка", Type: "string", Default: false,
				SQLExpr: "cs.external_url", SQLAlias: "external_url"},
			{Key: "provider_name", Label: "Провайдер", Type: "string", Default: true,
				SQLExpr: "cs.provider_name", SQLAlias: "provider_name"},
			{Key: "price", Label: "Стоимость", Type: "currency", Default: true,
				SQLExpr: "cs.price", SQLAlias: "price"},
			{Key: "price_currency", Label: "Валюта", Type: "string", Default: false,
				SQLExpr: "cs.price_currency", SQLAlias: "price_currency"},
			{Key: "duration_hours", Label: "Длительность (ч)", Type: "number", Default: false,
				SQLExpr: "cs.duration_hours", SQLAlias: "duration_hours"},
			{Key: "status", Label: "Статус", Type: "string", Default: true,
				SQLExpr: "cs.status", SQLAlias: "status"},
			{Key: "reviewed_by", Label: "Рецензент", Type: "string", Default: false,
				SQLExpr: "coalesce(rv.last_name || ' ' || rv.first_name, '')", SQLAlias: "reviewed_by"},
			{Key: "review_comment", Label: "Комментарий", Type: "string", Default: false,
				SQLExpr: "cs.review_comment", SQLAlias: "review_comment"},
			{Key: "reviewed_at", Label: "Дата рецензии", Type: "date", Default: false,
				SQLExpr: "cs.reviewed_at", SQLAlias: "reviewed_at"},
			{Key: "created_at", Label: "Дата предложения", Type: "date", Default: true,
				SQLExpr: "cs.created_at", SQLAlias: "created_at"},
		},
	}
}

// ---------------------------------------------------------------------------
// Source: course_requests (заявки на курсы — старый флоу)
// ---------------------------------------------------------------------------

func sourceCourseRequests() SourceDef {
	return SourceDef{
		Key:   "course_requests",
		Label: "Заявки на курсы (legacy)",
		FromSQL: `course_requests cr
			LEFT JOIN courses c ON c.id = cr.course_id
			LEFT JOIN providers p ON p.id = c.provider_id`,
		Columns: []ColumnDef{
			{Key: "request_no", Label: "Номер заявки", Type: "string", Default: true,
				SQLExpr: "cr.request_no", SQLAlias: "request_no"},
			{Key: "employee_name", Label: "ФИО", Type: "string", Default: true,
				SQLExpr: "cr.employee_full_name", SQLAlias: "employee_name"},
			{Key: "employee_email", Label: "Email", Type: "string", Default: true,
				SQLExpr: "cr.employee_email", SQLAlias: "employee_email"},
			{Key: "course_title", Label: "Курс", Type: "string", Default: true,
				SQLExpr: "cr.course_title", SQLAlias: "course_title"},
			{Key: "course_provider", Label: "Провайдер", Type: "string", Default: false,
				SQLExpr: "p.name", SQLAlias: "course_provider"},
			{Key: "course_price", Label: "Стоимость", Type: "currency", Default: false,
				SQLExpr: "c.price", SQLAlias: "course_price"},
			{Key: "status", Label: "Статус", Type: "string", Default: true,
				SQLExpr: "cr.status", SQLAlias: "status"},
			{Key: "manager_name", Label: "Руководитель", Type: "string", Default: false,
				SQLExpr: "cr.manager_full_name", SQLAlias: "manager_name"},
			{Key: "manager_comment", Label: "Комм. руководителя", Type: "string", Default: false,
				SQLExpr: "cr.manager_comment", SQLAlias: "manager_comment"},
			{Key: "hr_name", Label: "HR", Type: "string", Default: false,
				SQLExpr: "cr.hr_full_name", SQLAlias: "hr_name"},
			{Key: "hr_comment", Label: "Комм. HR", Type: "string", Default: false,
				SQLExpr: "cr.hr_comment", SQLAlias: "hr_comment"},
			{Key: "rejection_reason", Label: "Причина отказа", Type: "string", Default: false,
				SQLExpr: "cr.rejection_reason", SQLAlias: "rejection_reason"},
			{Key: "deadline_at", Label: "Дедлайн", Type: "date", Default: false,
				SQLExpr: "cr.deadline_at", SQLAlias: "deadline_at"},
			{Key: "requested_at", Label: "Дата заявки", Type: "date", Default: true,
				SQLExpr: "cr.requested_at", SQLAlias: "requested_at"},
			{Key: "manager_approved_at", Label: "Апрув руководителя", Type: "date", Default: false,
				SQLExpr: "cr.manager_approved_at", SQLAlias: "manager_approved_at"},
			{Key: "hr_approved_at", Label: "Апрув HR", Type: "date", Default: false,
				SQLExpr: "cr.hr_approved_at", SQLAlias: "hr_approved_at"},
			{Key: "started_at", Label: "Старт курса", Type: "date", Default: false,
				SQLExpr: "cr.started_at", SQLAlias: "started_at"},
			{Key: "completed_at", Label: "Завершение", Type: "date", Default: false,
				SQLExpr: "cr.completed_at", SQLAlias: "completed_at"},
		},
	}
}

// ---------------------------------------------------------------------------
// Source: enrollments (записи на курсы)
// ---------------------------------------------------------------------------

func sourceEnrollments() SourceDef {
	return SourceDef{
		Key:   "enrollments",
		Label: "Записи на курсы",
		FromSQL: `enrollments e
			JOIN users u ON u.id = e.user_id
			LEFT JOIN employee_profiles ep ON ep.user_id = u.id
			LEFT JOIN departments d ON d.id = ep.department_id
			JOIN courses c ON c.id = e.course_id
			LEFT JOIN providers p ON p.id = c.provider_id
			LEFT JOIN course_categories cat ON cat.id = c.category_id`,
		Columns: []ColumnDef{
			{Key: "employee_name", Label: "ФИО сотрудника", Type: "string", Default: true,
				SQLExpr: "coalesce(ep.last_name || ' ' || ep.first_name || coalesce(' ' || ep.middle_name, ''), u.email)", SQLAlias: "employee_name"},
			{Key: "employee_email", Label: "Email", Type: "string", Default: false,
				SQLExpr: "u.email", SQLAlias: "employee_email"},
			{Key: "department", Label: "Отдел", Type: "string", Default: true,
				SQLExpr: "d.name", SQLAlias: "department"},
			{Key: "position", Label: "Должность", Type: "string", Default: false,
				SQLExpr: "ep.position_title", SQLAlias: "position"},
			{Key: "course_title", Label: "Курс", Type: "string", Default: true,
				SQLExpr: "c.title", SQLAlias: "course_title"},
			{Key: "course_category", Label: "Категория", Type: "string", Default: false,
				SQLExpr: "cat.name", SQLAlias: "course_category"},
			{Key: "course_provider", Label: "Провайдер", Type: "string", Default: false,
				SQLExpr: "p.name", SQLAlias: "course_provider"},
			{Key: "course_price", Label: "Стоимость", Type: "currency", Default: true,
				SQLExpr: "c.price", SQLAlias: "course_price"},
			{Key: "price_currency", Label: "Валюта", Type: "string", Default: false,
				SQLExpr: "c.price_currency", SQLAlias: "price_currency"},
			{Key: "source", Label: "Источник", Type: "string", Default: true,
				SQLExpr: "e.source", SQLAlias: "source"},
			{Key: "status", Label: "Статус", Type: "string", Default: true,
				SQLExpr: "e.status", SQLAlias: "status"},
			{Key: "completion_percent", Label: "Прогресс %", Type: "number", Default: true,
				SQLExpr: "e.completion_percent", SQLAlias: "completion_percent"},
			{Key: "is_mandatory", Label: "Обязательный", Type: "string", Default: false,
				SQLExpr: "CASE WHEN e.is_mandatory THEN 'Да' ELSE 'Нет' END", SQLAlias: "is_mandatory"},
			{Key: "enrolled_at", Label: "Дата записи", Type: "date", Default: true,
				SQLExpr: "e.enrolled_at", SQLAlias: "enrolled_at"},
			{Key: "started_at", Label: "Дата старта", Type: "date", Default: false,
				SQLExpr: "e.started_at", SQLAlias: "started_at"},
			{Key: "completed_at", Label: "Дата завершения", Type: "date", Default: false,
				SQLExpr: "e.completed_at", SQLAlias: "completed_at"},
			{Key: "deadline_at", Label: "Дедлайн", Type: "date", Default: false,
				SQLExpr: "e.deadline_at", SQLAlias: "deadline_at"},
			{Key: "last_activity_at", Label: "Посл. активность", Type: "date", Default: false,
				SQLExpr: "e.last_activity_at", SQLAlias: "last_activity_at"},
			{Key: "duration_hours", Label: "Длительность (ч)", Type: "number", Default: false,
				SQLExpr: "c.duration_hours", SQLAlias: "duration_hours"},
			{Key: "course_level", Label: "Уровень курса", Type: "string", Default: false,
				SQLExpr: "c.level", SQLAlias: "course_level"},
		},
	}
}

// ---------------------------------------------------------------------------
// Source: courses (каталог курсов)
// ---------------------------------------------------------------------------

func sourceCourses() SourceDef {
	return SourceDef{
		Key:   "courses",
		Label: "Каталог курсов",
		FromSQL: `courses c
			LEFT JOIN providers p ON p.id = c.provider_id
			LEFT JOIN course_categories cat ON cat.id = c.category_id
			LEFT JOIN course_directions dir ON dir.id = c.direction_id`,
		Columns: []ColumnDef{
			{Key: "title", Label: "Название", Type: "string", Default: true,
				SQLExpr: "c.title", SQLAlias: "title"},
			{Key: "type", Label: "Тип", Type: "string", Default: true,
				SQLExpr: "c.type", SQLAlias: "type"},
			{Key: "source_type", Label: "Источник", Type: "string", Default: false,
				SQLExpr: "c.source_type", SQLAlias: "source_type"},
			{Key: "provider", Label: "Провайдер", Type: "string", Default: true,
				SQLExpr: "p.name", SQLAlias: "provider"},
			{Key: "category", Label: "Категория", Type: "string", Default: true,
				SQLExpr: "cat.name", SQLAlias: "category"},
			{Key: "direction", Label: "Направление", Type: "string", Default: false,
				SQLExpr: "dir.name", SQLAlias: "direction"},
			{Key: "level", Label: "Уровень", Type: "string", Default: true,
				SQLExpr: "c.level", SQLAlias: "level"},
			{Key: "price", Label: "Стоимость", Type: "currency", Default: true,
				SQLExpr: "c.price", SQLAlias: "price"},
			{Key: "price_currency", Label: "Валюта", Type: "string", Default: false,
				SQLExpr: "c.price_currency", SQLAlias: "price_currency"},
			{Key: "duration_hours", Label: "Длительность (ч)", Type: "number", Default: true,
				SQLExpr: "c.duration_hours", SQLAlias: "duration_hours"},
			{Key: "language", Label: "Язык", Type: "string", Default: false,
				SQLExpr: "c.language", SQLAlias: "language"},
			{Key: "status", Label: "Статус", Type: "string", Default: true,
				SQLExpr: "c.status", SQLAlias: "status"},
			{Key: "is_mandatory", Label: "Обязательный", Type: "string", Default: false,
				SQLExpr: "CASE WHEN c.is_mandatory_default THEN 'Да' ELSE 'Нет' END", SQLAlias: "is_mandatory"},
			{Key: "enrollment_count", Label: "Кол-во записей", Type: "number", Default: true,
				SQLExpr: "(SELECT count(*) FROM enrollments WHERE course_id = c.id)", SQLAlias: "enrollment_count"},
			{Key: "next_start_date", Label: "Ближайший старт", Type: "date", Default: false,
				SQLExpr: "c.next_start_date", SQLAlias: "next_start_date"},
			{Key: "external_url", Label: "Ссылка", Type: "string", Default: false,
				SQLExpr: "c.external_url", SQLAlias: "external_url"},
			{Key: "published_at", Label: "Дата публикации", Type: "date", Default: false,
				SQLExpr: "c.published_at", SQLAlias: "published_at"},
			{Key: "created_at", Label: "Дата создания", Type: "date", Default: false,
				SQLExpr: "c.created_at", SQLAlias: "created_at"},
		},
	}
}

// ---------------------------------------------------------------------------
// Source: employees (сотрудники)
// ---------------------------------------------------------------------------

func sourceEmployees() SourceDef {
	return SourceDef{
		Key:   "employees",
		Label: "Сотрудники",
		FromSQL: `employee_profiles ep
			JOIN users u ON u.id = ep.user_id
			LEFT JOIN departments d ON d.id = ep.department_id`,
		Columns: []ColumnDef{
			{Key: "full_name", Label: "ФИО", Type: "string", Default: true,
				SQLExpr: "ep.last_name || ' ' || ep.first_name || coalesce(' ' || ep.middle_name, '')", SQLAlias: "full_name"},
			{Key: "email", Label: "Email", Type: "string", Default: true,
				SQLExpr: "u.email", SQLAlias: "email"},
			{Key: "employee_no", Label: "Табельный номер", Type: "string", Default: false,
				SQLExpr: "ep.employee_no", SQLAlias: "employee_no"},
			{Key: "position", Label: "Должность", Type: "string", Default: true,
				SQLExpr: "ep.position_title", SQLAlias: "position"},
			{Key: "department", Label: "Отдел", Type: "string", Default: true,
				SQLExpr: "d.name", SQLAlias: "department"},
			{Key: "employment_status", Label: "Статус", Type: "string", Default: true,
				SQLExpr: "ep.employment_status", SQLAlias: "employment_status"},
			{Key: "hire_date", Label: "Дата приёма", Type: "date", Default: true,
				SQLExpr: "ep.hire_date", SQLAlias: "hire_date"},
			{Key: "enrollment_count", Label: "Кол-во курсов", Type: "number", Default: true,
				SQLExpr: "(SELECT count(*) FROM enrollments WHERE user_id = ep.user_id)", SQLAlias: "enrollment_count"},
			{Key: "completed_count", Label: "Завершено курсов", Type: "number", Default: true,
				SQLExpr: "(SELECT count(*) FROM enrollments WHERE user_id = ep.user_id AND status = 'completed')", SQLAlias: "completed_count"},
			{Key: "in_progress_count", Label: "В процессе", Type: "number", Default: false,
				SQLExpr: "(SELECT count(*) FROM enrollments WHERE user_id = ep.user_id AND status = 'in_progress')", SQLAlias: "in_progress_count"},
			{Key: "total_spent", Label: "Потрачено на обучение", Type: "currency", Default: false,
				SQLExpr: "(SELECT coalesce(sum(c2.price), 0) FROM enrollments e2 JOIN courses c2 ON c2.id = e2.course_id WHERE e2.user_id = ep.user_id AND e2.status IN ('completed','in_progress'))", SQLAlias: "total_spent"},
			{Key: "outlook_email", Label: "Outlook Email", Type: "string", Default: false,
				SQLExpr: "ep.outlook_email", SQLAlias: "outlook_email"},
		},
	}
}
