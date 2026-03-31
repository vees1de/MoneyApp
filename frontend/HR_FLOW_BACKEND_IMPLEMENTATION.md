# HR Flow Backend Implementation Guide

Документ описывает, как реализовать user-flow с диаграммы, используя текущие backend API.

## 1. Роли и доступ

- HR/админ: управление курсами, назначениями, лимитами, аналитикой.
- Manager: согласование заявок команды, просмотр manager dashboard.
- Employee: создание заявок, прохождение обучения, загрузка сертификатов.

Ключевая проверка прав: `GET /api/v1/auth/me` (`user.roles`, `user.permissions`).

## 2. Ветка "Каталог заявок" (внешнее обучение)

### 2.1 Создание и просмотр заявок

- Создать заявку: `POST /api/v1/external-requests`
- Список заявок с фильтрами:
  - `GET /api/v1/external-requests?scope=my|team|all&status=...&assignee=<uuid>`
- Список только своих: `GET /api/v1/external-requests/my`
- Карточка заявки: `GET /api/v1/external-requests/{id}`
- Виджет "мои задачи на согласование": `GET /api/v1/external-requests/pending-approvals`

### 2.2 Согласование

- Отправка в процесс согласования: `POST /api/v1/external-requests/{id}/submit`
- Approve: `POST /api/v1/external-requests/{id}/approve`
- Reject: `POST /api/v1/external-requests/{id}/reject`
- Request revision: `POST /api/v1/external-requests/{id}/request-revision`

### 2.3 Бюджет и календарь

При финальном approve backend уже выполняет:

- проверку/резерв лимита (если настроен `budget-limits`);
- создание calendar event;
- постановку integration job для Outlook.

Для UI:

- показывать `calendar_conflict_status`, `budget_check_status`, `status` в карточке заявки;
- при reject/revision предлагать "рекомендованные курсы" через `GET /api/v1/recommendations/courses`.

## 3. Ветка "Создать обязательный/назначенный курс"

### 3.1 Подготовка курса

- Создать курс: `POST /api/v1/courses`
- Обновить: `PATCH /api/v1/courses/{id}`
- Опубликовать: `POST /api/v1/courses/{id}/publish`
- Материалы: `POST /api/v1/courses/{id}/materials`

### 3.2 Назначение по отделам/командам

- Назначить курс: `POST /api/v1/assignments`
  - использовать `target_type=department|group|user`
  - задавать `deadline_at`, `priority`, `reason`
- Контроль назначений: `GET /api/v1/assignments`

### 3.3 Прохождение и сертификат

- Список моих enrollments: `GET /api/v1/enrollments/my`
- Start/progress/complete:
  - `POST /api/v1/enrollments/{id}/start`
  - `POST /api/v1/enrollments/{id}/progress`
  - `POST /api/v1/enrollments/{id}/complete`
- Загрузка сертификата:
  - общий кейс: `POST /api/v1/certificates/upload`
  - для external request: `POST /api/v1/external-requests/{id}/upload-certificate`

## 4. Ветка "Календарь и интеграция с Outlook"

- Инициация OAuth: `GET /api/v1/integrations/outlook/connect`
- Статус интеграции: `GET /api/v1/integrations/outlook/status`
- Ручная синхронизация: `POST /api/v1/integrations/outlook/sync`
- Отключение: `POST /api/v1/integrations/outlook/disconnect`
- Виджет "Ближайшие события": `GET /api/v1/calendar/events/upcoming?limit=5`

## 5. Ветка "Создание отдела / сотрудники / бюджет"

### Что уже есть

- У пользователя и профиля есть `department_id`, это возвращается в `auth/me`.
- Лимиты бюджета по отделу:
  - `POST /api/v1/budget-limits` (`scope_type=department`, `scope_id=<department_uuid>`)
  - `GET /api/v1/budget-limits`
- Пользователей можно заводить/обновлять через admin endpoints (`department_id` в payload).

### Что отсутствует как публичный API

- Нет отдельного REST набора для CRUD департаментов (`/departments`).
- Если нужен UI "создание отдела/описание отдела", нужен новый модуль API:
  - `POST /api/v1/departments`
  - `GET /api/v1/departments`
  - `PATCH /api/v1/departments/{id}`
  - `GET /api/v1/departments/{id}/employees`

## 6. Ветка "Дашборд и аналитика"

### Typed dashboard payloads

- Employee dashboard: `GET /api/v1/dashboard/employee`
  - тип: `dashboard_api.EmployeeDashboard`
  - включает: stats, upcoming events, recommended courses, learning plan, preview external requests
- Manager dashboard: `GET /api/v1/dashboard/manager`
  - тип: `dashboard_api.ManagerDashboard`
  - включает: stats, team preview, pending approvals, team requests

### Дополнительная аналитика/экспорт

- `GET /api/v1/analytics/dashboard/hr`
- `GET /api/v1/analytics/dashboard/manager`
- `GET /api/v1/analytics/compliance`
- `GET /api/v1/analytics/external-requests`
- `GET /api/v1/analytics/budget`
- `GET /api/v1/analytics/trainers`
- Экспорт: `GET /api/v1/reports/export/excel`, `GET /api/v1/reports/export/pdf`

## 7. Ветка "Рекомендации / план развития / Jira-виджет"

- Рекомендации курсов: `GET /api/v1/recommendations/courses`
- Мой план развития: `GET /api/v1/learning-plan/my`
- Jira/board summary (в текущем backend источник Yougile): `GET /api/v1/jira/board-summary`

## 8. Рекомендуемый frontend orchestration

### После логина

1. `POST /api/v1/auth/login`
2. `GET /api/v1/auth/me`
3. Параллельно:
   - `GET /api/v1/dashboard/employee` или `/dashboard/manager` (по permissions)
   - `GET /api/v1/notifications`
   - `GET /api/v1/external-requests/pending-approvals` (для manager/HR)

### Для HR workspace

1. Каталог заявок: `GET /external-requests?scope=all`
2. Фильтры: `status`, `assignee`, `scope`
3. Карточка: `GET /external-requests/{id}`
4. Решение: approve/reject/request-revision
5. Мониторинг: dashboard + analytics + export

## 9. Что добавить в API для полного соответствия схеме

- CRUD департаментов как публичные endpoints.
- Явный endpoint "проверка календарных конфликтов до submit" (если нужен pre-check в UI до фактического approve).
- Typed DTO для legacy analytics endpoints (сейчас часть из них `Record<string, unknown>` в референсе).

