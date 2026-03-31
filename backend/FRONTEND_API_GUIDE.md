# Frontend API Guide

Этот гайд нужен фронтенду как короткая карта интеграции с LMS backend.

Source of truth:
- OpenAPI YAML: `/openapi.yaml`
- Swagger UI: `/swagger`
- API base: `/api/v1`

## Общие правила

- Все защищенные методы требуют заголовок `Authorization: Bearer <access_token>`.
- Все даты и timestamps приходят в RFC3339, например `2026-03-30T12:00:00Z`.
- Все `numeric` поля сериализуются строками, например `"12.50"`, `"0"`, `"87.25"`.
- Проверку ролей и permissions лучше строить по ответу `GET /api/v1/auth/me`.
- Если пользователь не авторизован, backend возвращает `401`.
- Если роль есть, но permission не хватает, backend возвращает `403`.
- Ошибка приходит в виде:

```json
{
  "error": {
    "code": "forbidden",
    "message": "missing permission"
  }
}
```

## Что фронт должен загрузить сразу после логина

Базовый bootstrap приложения:

1. `POST /api/v1/auth/login`
2. Сохранить `access_token` и `refresh_token`
3. `GET /api/v1/auth/me`
4. По `roles` и `permissions` решить, какие разделы показывать в меню
5. Параллельно можно загрузить:
   - `GET /api/v1/notifications`
   - `GET /api/v1/enrollments/my`
   - `GET /api/v1/external-requests/my`
   - `GET /api/v1/certificates/my`

`/auth/me` для фронта главный метод. Он нужен не только для профиля, но и для:
- построения меню
- route guards
- определения HR / manager / trainer / admin сценариев
- показа employee profile и department scope

## 1. Аутентификация

### Регистрация

Экран: регистрация нового сотрудника

- `POST /api/v1/auth/register`
- После успешной регистрации backend сразу возвращает токены
- Затем фронт вызывает `GET /api/v1/auth/me`

### Логин

Экран: форма логина

- `POST /api/v1/auth/login`
- В body отправляем `email` и `password`
- В ответе получаем `access_token`, `refresh_token`, `expires_in`

### Обновление токена

- Когда access token истек, фронт вызывает `POST /api/v1/auth/refresh`
- Если refresh token тоже невалиден, нужно разлогинить пользователя

### Выход

- `POST /api/v1/auth/logout`
- После этого фронт очищает local storage / cookie storage и редиректит на login

### Сброс пароля

Flow:
- `POST /api/v1/auth/forgot-password`
- пользователь получает reset flow
- `POST /api/v1/auth/reset-password`

## 2. Навигация по ролям

UI лучше строить от permissions, а не от hardcoded ролей.

Примеры:
- `users.read`, `users.write`, `roles.manage` показывают admin users
- `courses.write` показывает редактирование каталога
- `courses.assign` показывает экран назначения обучения
- `certificates.verify` показывает модерацию сертификатов
- `programs.manage` показывает corporate university management
- `analytics.read_hr` показывает HR dashboard
- `audit.read` показывает аудит

## 3. Личный кабинет сотрудника

Экран "Мое обучение":

- `GET /api/v1/enrollments/my`
- Использовать для вкладок:
  - все активные
  - обязательные
  - завершенные
  - просроченные

Экран "Мои сертификаты":

- `GET /api/v1/certificates/my`

Экран "Мои заявки на внешнее обучение":

- `GET /api/v1/external-requests/my`

Экран "Уведомления":

- `GET /api/v1/notifications`
- `POST /api/v1/notifications/{id}/read`
- `POST /api/v1/notifications/read-all`

## 4. Каталог курсов

### Список курсов

Экран: catalog / search / filters

- `GET /api/v1/courses`

Что удобно строить на фронте:
- фильтры по `type`, `status`, `direction_id`, `category_id`
- карточки курса
- переход в course details

### Детали курса

- `GET /api/v1/courses/{id}`
- `GET /api/v1/courses/{id}/materials`

### Управление курсом

Только для HR/L&D или admin с нужными permissions:

- `POST /api/v1/courses`
- `PATCH /api/v1/courses/{id}`
- `POST /api/v1/courses/{id}/publish`
- `POST /api/v1/courses/{id}/archive`
- `POST /api/v1/courses/{id}/materials`

Ожидаемый UI flow:
- создать draft
- заполнить описание
- добавить материалы
- опубликовать

## 5. Назначение обучения

Экран: assign training

- `POST /api/v1/assignments`
- `GET /api/v1/assignments`

Важные поля:
- `course_id`
- `assignment_type`: `individual | department | group | role_based`
- `target_type`: `user | department | group`
- `target_id`
- `priority`: `mandatory | recommended`
- `deadline_at`

Что должен делать фронт:
- дать HR выбрать target
- показать предупреждение, если назначение обязательное и есть deadline
- после успеха обновить список assignments

Backend на этом шаге создает enrollments автоматически.

## 6. Прохождение курса

Экран: enrollment details

- `GET /api/v1/enrollments/{id}`

Flow прохождения:

1. Пользователь открывает курс
2. Фронт вызывает `POST /api/v1/enrollments/{id}/start`
3. По мере прохождения модулей отправляет `POST /api/v1/enrollments/{id}/progress`
4. Когда условия выполнены, вызывает `POST /api/v1/enrollments/{id}/complete`

Фронту важно:
- не пытаться завершать курс, если тест или обязательные материалы еще не пройдены
- обновлять локальный progress после каждого ответа backend

## 7. Тестирование

Экран: test details

- `GET /api/v1/tests/{id}`

Flow попытки:

1. `POST /api/v1/tests/{id}/attempts`
2. Фронт рендерит вопросы
3. Сохраняет ответы через `POST /api/v1/test-attempts/{id}/answers`
4. Отправляет попытку через `POST /api/v1/test-attempts/{id}/submit`
5. Для истории и аналитики можно вызвать `GET /api/v1/tests/{id}/results`

UI рекомендации:
- для автосохранения вызывать `answers` батчами
- после `submit` блокировать редактирование
- показывать `passed`, `score`, `attempt_no`

## 8. Сертификаты

### Загрузка сотрудником

- `POST /api/v1/certificates/upload`
- `GET /api/v1/certificates/my`

### Модерация HR/L&D

- `POST /api/v1/certificates/{id}/verify`
- `POST /api/v1/certificates/{id}/reject`

Экран для фронта обычно делится на:
- мои сертификаты
- очередь на проверку
- карточка сертификата с историей статусов

## 9. Внешнее обучение

Это отдельный бизнес-flow и его лучше делать как wizard.

### Шаг 1. Создание черновика

- `POST /api/v1/external-requests`

Экран собирает:
- название курса
- провайдера
- ссылку
- стоимость
- сроки
- описание программы
- бизнес-цель
- комментарий сотрудника

### Шаг 2. Редактирование черновика

- `PATCH /api/v1/external-requests/{id}`
- `GET /api/v1/external-requests/{id}`

### Шаг 3. Отправка на согласование

- `POST /api/v1/external-requests/{id}/submit`

После этого фронт должен:
- обновить карточку заявки
- показать текущий статус
- отобразить, кому ушло согласование

### Шаг 4. Действия согласующего

- `POST /api/v1/external-requests/{id}/approve`
- `POST /api/v1/external-requests/{id}/reject`
- `POST /api/v1/external-requests/{id}/request-revision`

Для списка и карточки заявки backend теперь дополнительно отдает:
- `employee_full_name`
- `employee_email`
- `department_name`
- `current_approval_status`
- `current_approval_role_code`
- `current_approval_due_at`
- `current_approver_user_id`
- `current_approver_full_name`

### Шаг 5. Сертификат по внешнему курсу

- `POST /api/v1/external-requests/{id}/upload-certificate`

Главные статусы для UI:
- `draft`
- `submitted`
- `manager_approval`
- `hr_approval`
- `approved`
- `needs_revision`
- `rejected`
- `in_training`
- `completed`
- `certificate_uploaded`
- `closed`
- `canceled`

Практично сделать на фронте status map:
- label
- color
- allowed actions
- next step hint

Видимость:
- сотрудник видит только свои заявки
- руководитель видит заявки своей команды и свои собственные
- HR/L&D видит все заявки и свои собственные
- очередь согласования: `GET /api/v1/external-requests/pending-approvals`

## 10. Approval workflows и бюджеты

Это обычно admin settings screens.

Approval workflows:
- `GET /api/v1/approval-workflows`
- `POST /api/v1/approval-workflows`

Budget limits:
- `GET /api/v1/budget-limits`
- `POST /api/v1/budget-limits`

Для фронта это CRUD-настройки, но они влияют на flow внешнего обучения.

## 11. Outlook интеграция

Текущий backend flow упрощенный. Для фронта сценарий такой:

1. Пользователь открывает settings integrations
2. `GET /api/v1/integrations/outlook/status`
3. Если не подключено, `GET /api/v1/integrations/outlook/connect`
4. Фронт открывает `auth_url`
5. После callback backend связывает аккаунт
6. Пользователь может вызвать `POST /api/v1/integrations/outlook/sync`
7. Отключение через `POST /api/v1/integrations/outlook/disconnect`

Важно:
- сейчас это backend-level scaffold, а не полная Microsoft Graph интеграция
- поэтому на фронте лучше пометить фичу как beta/integration in progress

## 12. Корпоративный университет

### Программы

- `GET /api/v1/programs`
- `POST /api/v1/programs`
- `GET /api/v1/programs/{id}`

### Группы

- `POST /api/v1/programs/{id}/groups`
- `POST /api/v1/groups/{id}/participants`

### Сессии

- `POST /api/v1/groups/{id}/sessions`
- `POST /api/v1/sessions/{id}/trainer-feedback`
- `POST /api/v1/sessions/{id}/participant-feedback`

Ожидаемый UI flow:
- HR создает программу
- внутри программы создает группы
- в группе создает сессии
- добавляет участников
- тренер отмечает attendance
- участники отправляют feedback

## 13. YouGile интеграция

Для YouGile лучше делать отдельный settings flow для HR/admin.

### Шаг 1. Подключение

- `POST /api/v1/integrations/yougile/connections/test-key`
- `POST /api/v1/integrations/yougile/connections`

Если ключа еще нет:
- `POST /api/v1/integrations/yougile/discover-companies`
- `POST /api/v1/integrations/yougile/connections/create-key`

### Шаг 2. Импорт справочников

- `POST /api/v1/integrations/yougile/connections/{id}/import/users`
- `POST /api/v1/integrations/yougile/connections/{id}/import/structure`

Локальные данные потом читать отсюда:
- `GET /api/v1/integrations/yougile/connections/{id}/users`
- `GET /api/v1/integrations/yougile/connections/{id}/projects`
- `GET /api/v1/integrations/yougile/connections/{id}/boards`
- `GET /api/v1/integrations/yougile/connections/{id}/columns`

### Шаг 3. Сопоставление сотрудников

- `POST /api/v1/integrations/yougile/connections/{id}/mappings/auto-match`
- `GET /api/v1/integrations/yougile/connections/{id}/mappings`
- `POST /api/v1/integrations/yougile/connections/{id}/mappings`
- `DELETE /api/v1/integrations/yougile/connections/{id}/mappings/{mappingId}`

Практически:
- сначала попытаться `auto-match` по email
- потом на фронте показать unmatched пользователей
- для unmatched дать ручной mapping

### Шаг 4. Sync

- `POST /api/v1/integrations/yougile/connections/{id}/sync`
- `GET /api/v1/integrations/yougile/sync-jobs/{jobId}`
- `POST /api/v1/integrations/yougile/connections/{id}/sync/backfill`

Что важно сейчас:
- users и structure sync уже поддержаны
- tasks sync в текущем backend пока scaffolded, но job-контур и storage уже есть

### UI рекомендации

- сделать отдельный экран connection details
- там показывать status, last sync, last error
- import и sync actions лучше запускать явными кнопками
- progress экрана строить по `sync-jobs/{jobId}`

## 14. GitHub интеграция

Это отдельный engineering dashboard flow. Его лучше не смешивать с HR-оценкой сотрудника.

### Подключение

- `POST /api/v1/integrations/github/connections/test`
- `POST /api/v1/integrations/github/connections`
- `GET /api/v1/integrations/github/connections`
- `GET /api/v1/integrations/github/connections/{connectionId}`
- `PATCH /api/v1/integrations/github/connections/{connectionId}`
- `DELETE /api/v1/integrations/github/connections/{connectionId}`

Текущий MVP auth mode:
- `pat`

### Mapping сотрудников

- `GET /api/v1/integrations/github/connections/{connectionId}/mappings`
- `POST /api/v1/integrations/github/connections/{connectionId}/mappings`
- `POST /api/v1/integrations/github/connections/{connectionId}/mappings/auto-match`
- `DELETE /api/v1/integrations/github/connections/{connectionId}/mappings/{mappingId}`

Стратегии автосопоставления:
- `email`
- `login`
- `domain`

### Импорт и sync

- `POST /api/v1/integrations/github/connections/{connectionId}/import/users`
- `POST /api/v1/integrations/github/connections/{connectionId}/import/repos`
- `POST /api/v1/integrations/github/connections/{connectionId}/import/languages`
- `POST /api/v1/integrations/github/connections/{connectionId}/sync`
- `GET /api/v1/integrations/github/sync-jobs/{jobId}`

### Данные GitHub

- `GET /api/v1/integrations/github/users`
- `GET /api/v1/integrations/github/repositories`
- `GET /api/v1/integrations/github/repositories/{repoId}`
- `GET /api/v1/integrations/github/repositories/{repoId}/languages`
- `GET /api/v1/integrations/github/repositories/{repoId}/contributors`

### Метрики по сотруднику

- `GET /api/v1/integrations/github/employees/{employeeUserId}/profile?connectionId=...`
- `GET /api/v1/integrations/github/employees/{employeeUserId}/languages?connectionId=...`
- `GET /api/v1/integrations/github/employees/{employeeUserId}/stats?connectionId=...`
- `GET /api/v1/integrations/github/employees/{employeeUserId}/activity?connectionId=...`

Что реально можно использовать уже сейчас:
- language profile
- repositories count
- active repositories count
- stars / forks
- avg repo freshness
- primary languages
- `engineeringActivityScore`

Важно:
- `engineeringActivityScore` это не performance review
- это engineering activity footprint, собранный из свежести repo, активных repo и delivery signals

### Командная аналитика

- `GET /api/v1/integrations/github/analytics/team?connectionId=...`
- `GET /api/v1/integrations/github/analytics/languages?connectionId=...`
- `GET /api/v1/integrations/github/analytics/top-languages?connectionId=...`
- `GET /api/v1/integrations/github/analytics/repository-health?connectionId=...`
- `GET /api/v1/integrations/github/analytics/repository-ownership?connectionId=...`

## 15. Аналитика и отчеты

HR dashboard:
- `GET /api/v1/analytics/dashboard/hr`

Manager dashboard:
- `GET /api/v1/analytics/dashboard/manager`

Отчеты и срезы:
- `GET /api/v1/analytics/compliance`
- `GET /api/v1/analytics/external-requests`
- `GET /api/v1/analytics/budget`
- `GET /api/v1/analytics/trainers`

Exports:
- `GET /api/v1/reports/export/excel`
- `GET /api/v1/reports/export/pdf`

Практически:
- dashboard endpoints лучше грузить лениво по вкладкам
- export endpoints лучше вызывать через явную кнопку "Сформировать"

## 16. Аудит

Экран для admin/audit:

- `GET /api/v1/audit-logs`

Полезно для:
- timeline по изменениям
- разборов кто и когда одобрил заявку
- compliance и traceability screens

## 17. Рекомендуемый frontend bootstrap по экранам

### После входа сотрудника

Загрузить:
- `GET /api/v1/auth/me`
- `GET /api/v1/enrollments/my`
- `GET /api/v1/external-requests/my`
- `GET /api/v1/notifications`

### После входа HR/L&D

Загрузить:
- `GET /api/v1/auth/me`
- `GET /api/v1/courses`
- `GET /api/v1/assignments`
- `GET /api/v1/analytics/dashboard/hr`
- `GET /api/v1/notifications`

### После входа руководителя

Загрузить:
- `GET /api/v1/auth/me`
- `GET /api/v1/analytics/dashboard/manager`
- заявки, доступные на согласование
- `GET /api/v1/notifications`

## 18. Что уже можно интегрировать, а что пока считать beta

Можно интегрировать как основной контракт:
- auth
- admin users / roles / permissions
- courses
- assignments
- enrollments
- tests
- certificates
- external requests
- notifications
- programs / groups / sessions
- analytics endpoints
- audit logs
- yougile connection / import / mapping / sync API
- github connection / mapping / repos / languages / employee stats API

Пока лучше считать beta/scaffold:
- Outlook connect/sync
- фоновые job handlers
- outbox processing вне самой транзакционной записи
- YouGile tasks ingestion и webhook processing
- GitHub PR/review/commit daily enrichment через GraphQL или extended REST backfill

## 19. Где смотреть точные payloads

Точные схемы request/response уже описаны в Swagger:

- UI: `/swagger`
- YAML: `/openapi.yaml`

Если фронту нужен автоген клиента:
- брать лучше `openapi.yaml`
- использовать его как основной контракт вместо чтения Go handler-ов
