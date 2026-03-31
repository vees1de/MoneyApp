# Unified Frontend API + Types Reference

Единый референс, собранный из:
- FRONTEND_API_GUIDE.md (базовые методы)
- internal/app/router.go (добавленные фактические методы)
- ENTITY_TS_REFERENCE.md (типы)

## 1) Typed API Contract

| Method | Path | Request Type | Response Type | Notes |
|---|---|---|---|---|
| DELETE | /api/v1/integrations/github/connections/{connectionId} | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| DELETE | /api/v1/integrations/github/connections/{connectionId}/mappings/{mappingId} | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| DELETE | /api/v1/integrations/yougile/connections/{id}/mappings/{mappingId} | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| GET | /api/v1/analytics/budget | never | Record<string, unknown> | Бюджетная аналитика. |
| GET | /api/v1/analytics/compliance | never | Record<string, unknown> | Compliance аналитика. |
| GET | /api/v1/analytics/dashboard/hr | never | Record<string, unknown> | HR dashboard (legacy analytics endpoint). |
| GET | /api/v1/analytics/dashboard/manager | never | Record<string, unknown> | Manager dashboard (legacy analytics endpoint). |
| GET | /api/v1/analytics/external-requests | never | Record<string, unknown> | Аналитика внешнего обучения. |
| GET | /api/v1/analytics/trainers | never | Record<string, unknown> | Аналитика тренеров. |
| GET | /api/v1/approval-workflows | never | external_training.ApprovalWorkflow[] | Список workflow. |
| GET | /api/v1/assignments | never | learning.Assignment[] | Список назначений. |
| GET | /api/v1/audit-logs | never | audit.LogEntry[] | Журнал аудита. |
| GET | /api/v1/auth/me | never | identity.MeResponse | Основной bootstrap endpoint для ролей и permissions. |
| GET | /api/v1/budget-limits | never | external_training.BudgetLimit[] | Список лимитов бюджета. |
| GET | /api/v1/calendar/events/upcoming | never | calendar.UpcomingCalendarEvent[] | Ближайшие события. Query: limit (1..20). |
| GET | /api/v1/certificates/my | never | certificates.Certificate[] | Мои сертификаты. |
| GET | /api/v1/courses | never | catalog.Course[] | Список курсов (с фильтрами через query params). |
| GET | /api/v1/courses/{id} | never | catalog.Course | Детали курса. |
| GET | /api/v1/courses/{id}/materials | never | catalog.CourseMaterial[] | Материалы курса. |
| GET | /api/v1/dashboard/employee | never | dashboard_api.EmployeeDashboard | Typed employee dashboard payload. |
| GET | /api/v1/dashboard/manager | never | dashboard_api.ManagerDashboard | Typed manager dashboard payload. |
| GET | /api/v1/enrollments/{id} | never | learning.Enrollment | Детали enrollment. |
| GET | /api/v1/enrollments/my | never | learning.Enrollment[] | Обучения текущего пользователя. |
| GET | /api/v1/external-requests | never | external_training.ExternalRequest[] | Список заявок. Query: scope=my\\|team\\|all, status (repeated), assignee. |
| GET | /api/v1/external-requests/{id} | never | external_training.ExternalRequest | Детали заявки. |
| GET | /api/v1/external-requests/my | never | external_training.ExternalRequest[] | Мои заявки. |
| GET | /api/v1/external-requests/pending-approvals | never | external_training.PendingApprovalItem[] | Текущие задачи согласования для пользователя. |
| GET | /api/v1/integrations/github/analytics/languages | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| GET | /api/v1/integrations/github/analytics/repository-health | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| GET | /api/v1/integrations/github/analytics/repository-ownership | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| GET | /api/v1/integrations/github/analytics/team | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| GET | /api/v1/integrations/github/analytics/top-languages | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| GET | /api/v1/integrations/github/connections | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| GET | /api/v1/integrations/github/connections/{connectionId} | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| GET | /api/v1/integrations/github/connections/{connectionId}/mappings | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| GET | /api/v1/integrations/github/employees/{employeeUserId}/activity | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| GET | /api/v1/integrations/github/employees/{employeeUserId}/languages | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| GET | /api/v1/integrations/github/employees/{employeeUserId}/profile | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| GET | /api/v1/integrations/github/employees/{employeeUserId}/stats | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| GET | /api/v1/integrations/github/repositories | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| GET | /api/v1/integrations/github/repositories/{repoId} | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| GET | /api/v1/integrations/github/repositories/{repoId}/contributors | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| GET | /api/v1/integrations/github/repositories/{repoId}/languages | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| GET | /api/v1/integrations/github/sync-jobs/{jobId} | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| GET | /api/v1/integrations/github/users | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| GET | /api/v1/integrations/outlook/connect | never | outlook.ConnectResponse | Ссылка и state для OAuth подключения. |
| GET | /api/v1/integrations/outlook/status | never | outlook.IntegrationStatus | Текущий статус интеграции. |
| GET | /api/v1/integrations/yougile/connections/{id}/boards | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| GET | /api/v1/integrations/yougile/connections/{id}/columns | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| GET | /api/v1/integrations/yougile/connections/{id}/mappings | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| GET | /api/v1/integrations/yougile/connections/{id}/projects | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| GET | /api/v1/integrations/yougile/connections/{id}/users | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| GET | /api/v1/integrations/yougile/sync-jobs/{jobId} | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| GET | /api/v1/jira/board-summary | never | board_summary.BoardSummary | Сводка по борду (в текущем коде источник yougile). Query: connection_id, board_id. |
| GET | /api/v1/learning-plan/my | never | learning_plan.MyLearningPlan | Мой план развития. |
| GET | /api/v1/notifications | never | notifications.Notification[] | Список уведомлений. |
| GET | /api/v1/programs | never | university.Program[] | Список программ. |
| GET | /api/v1/programs/{id} | never | university.Program | Детали программы. |
| GET | /api/v1/recommendations/courses | never | learning_plan.RecommendedCourseItem[] | Рекомендованные курсы. Query: limit, offset. |
| GET | /api/v1/reports/export/excel | never | Blob | Скачивание Excel отчёта. |
| GET | /api/v1/reports/export/pdf | never | Blob | Скачивание PDF отчёта. |
| GET | /api/v1/tests/{id} | never | testing.Test | Тест с вопросами/опциями. |
| GET | /api/v1/tests/{id}/results | never | testing.TestResult[] | История/результаты по тесту. |
| PATCH | /api/v1/courses/{id} | catalog.UpdateCourseRequest | catalog.Course | Частичное обновление курса. |
| PATCH | /api/v1/external-requests/{id} | external_training.UpdateExternalRequestRequest | external_training.ExternalRequest | Обновление заявки. |
| PATCH | /api/v1/integrations/github/connections/{connectionId} | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| POST | /api/v1/approval-workflows | external_training.CreateWorkflowRequest | external_training.ApprovalWorkflow | Создание workflow. |
| POST | /api/v1/assignments | learning.CreateAssignmentRequest | learning.Assignment | Назначить обучение. |
| POST | /api/v1/auth/forgot-password | identity.ForgotPasswordRequest | Record<string, unknown> | Запуск flow восстановления. |
| POST | /api/v1/auth/login | identity.LoginRequest | identity.AuthResponse | Логин по email/password. |
| POST | /api/v1/auth/logout | identity.LogoutRequest | Record<string, unknown> | Разлогин/инвалидация refresh token. |
| POST | /api/v1/auth/refresh | identity.RefreshRequest | identity.AuthResponse | Обновление access/refresh токенов. |
| POST | /api/v1/auth/register | identity.RegisterRequest | identity.AuthResponse | Регистрация + выдача токенов. |
| POST | /api/v1/auth/reset-password | identity.ResetPasswordRequest | Record<string, unknown> | Подтверждение сброса пароля. |
| POST | /api/v1/budget-limits | external_training.CreateBudgetLimitRequest | external_training.BudgetLimit | Создание лимита бюджета. |
| POST | /api/v1/certificates/{id}/reject | certificates.ReviewRequest | certificates.Certificate | Отклонение сертификата. |
| POST | /api/v1/certificates/{id}/verify | certificates.ReviewRequest | certificates.Certificate | Подтверждение сертификата. |
| POST | /api/v1/certificates/upload | certificates.UploadCertificateRequest | certificates.Certificate | Загрузка сертификата пользователем. |
| POST | /api/v1/courses | catalog.CreateCourseRequest | catalog.Course | Создание курса. |
| POST | /api/v1/courses/{id}/archive | never | catalog.Course | Архивация курса. |
| POST | /api/v1/courses/{id}/materials | catalog.CreateMaterialRequest | catalog.CourseMaterial | Добавление материала. |
| POST | /api/v1/courses/{id}/publish | never | catalog.Course | Публикация курса. |
| POST | /api/v1/enrollments/{id}/complete | learning.CompleteRequest | learning.Enrollment | Завершение обучения. |
| POST | /api/v1/enrollments/{id}/progress | learning.ProgressRequest | learning.Enrollment | Обновление прогресса. |
| POST | /api/v1/enrollments/{id}/start | never | learning.Enrollment | Старт обучения. |
| POST | /api/v1/external-requests | external_training.CreateExternalRequestRequest | external_training.ExternalRequest | Создание заявки на внешнее обучение. |
| POST | /api/v1/external-requests/{id}/approve | external_training.ActionCommentRequest | external_training.ExternalRequest | Согласование шага. |
| POST | /api/v1/external-requests/{id}/reject | external_training.ActionCommentRequest | external_training.ExternalRequest | Отклонение шага. |
| POST | /api/v1/external-requests/{id}/request-revision | external_training.ActionCommentRequest | external_training.ExternalRequest | Возврат на доработку. |
| POST | /api/v1/external-requests/{id}/submit | never | external_training.ExternalRequest | Отправить заявку в workflow. |
| POST | /api/v1/external-requests/{id}/upload-certificate | external_training.UploadRequestCertificateRequest | external_training.ExternalRequest | Загрузка сертификата по заявке. |
| POST | /api/v1/groups/{id}/participants | university.AddParticipantRequest | university.TrainingGroup | Добавление участника в группу. |
| POST | /api/v1/groups/{id}/sessions | university.CreateSessionRequest | university.Session | Создание сессии/занятия. |
| POST | /api/v1/integrations/github/connections | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| POST | /api/v1/integrations/github/connections/{connectionId}/import/languages | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| POST | /api/v1/integrations/github/connections/{connectionId}/import/repos | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| POST | /api/v1/integrations/github/connections/{connectionId}/import/users | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| POST | /api/v1/integrations/github/connections/{connectionId}/mappings | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| POST | /api/v1/integrations/github/connections/{connectionId}/mappings/auto-match | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| POST | /api/v1/integrations/github/connections/{connectionId}/sync | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| POST | /api/v1/integrations/github/connections/test | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| POST | /api/v1/integrations/outlook/disconnect | never | outlook.IntegrationStatus | Отключение интеграции. |
| POST | /api/v1/integrations/outlook/sync | never | outlook.IntegrationStatus | Ручная синхронизация. |
| POST | /api/v1/integrations/yougile/connections | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| POST | /api/v1/integrations/yougile/connections/{id}/import/structure | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| POST | /api/v1/integrations/yougile/connections/{id}/import/users | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| POST | /api/v1/integrations/yougile/connections/{id}/mappings | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| POST | /api/v1/integrations/yougile/connections/{id}/mappings/auto-match | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| POST | /api/v1/integrations/yougile/connections/{id}/sync | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| POST | /api/v1/integrations/yougile/connections/{id}/sync/backfill | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| POST | /api/v1/integrations/yougile/connections/create-key | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| POST | /api/v1/integrations/yougile/connections/test-key | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| POST | /api/v1/integrations/yougile/discover-companies | unknown | unknown | Не найдено точного соответствия в текущем typed map |
| POST | /api/v1/notifications/{id}/read | never | notifications.Notification | Пометить уведомление прочитанным. |
| POST | /api/v1/notifications/read-all | never | Record<string, unknown> | Пометить все как прочитанные. |
| POST | /api/v1/programs | university.CreateProgramRequest | university.Program | Создание программы. |
| POST | /api/v1/programs/{id}/groups | university.CreateGroupRequest | university.TrainingGroup | Создание группы программы. |
| POST | /api/v1/sessions/{id}/participant-feedback | university.ParticipantFeedbackRequest | Record<string, unknown> | Фидбек участника. |
| POST | /api/v1/sessions/{id}/trainer-feedback | university.TrainerFeedbackRequest | Record<string, unknown> | Фидбек тренера по участнику. |
| POST | /api/v1/test-attempts/{id}/answers | testing.SubmitAnswersRequest | testing.TestAttempt | Сохранение ответов. |
| POST | /api/v1/test-attempts/{id}/submit | never | testing.TestAttempt | Финальная отправка попытки. |
| POST | /api/v1/tests/{id}/attempts | testing.StartAttemptRequest | testing.TestAttempt | Старт попытки. |

### TS endpoint map

```ts
export type ApiContract = {
  'DELETE /api/v1/integrations/github/connections/{connectionId}': { request: unknown; response: unknown };
  'DELETE /api/v1/integrations/github/connections/{connectionId}/mappings/{mappingId}': { request: unknown; response: unknown };
  'DELETE /api/v1/integrations/yougile/connections/{id}/mappings/{mappingId}': { request: unknown; response: unknown };
  'GET /api/v1/analytics/budget': { request: never; response: Record<string, unknown> };
  'GET /api/v1/analytics/compliance': { request: never; response: Record<string, unknown> };
  'GET /api/v1/analytics/dashboard/hr': { request: never; response: Record<string, unknown> };
  'GET /api/v1/analytics/dashboard/manager': { request: never; response: Record<string, unknown> };
  'GET /api/v1/analytics/external-requests': { request: never; response: Record<string, unknown> };
  'GET /api/v1/analytics/trainers': { request: never; response: Record<string, unknown> };
  'GET /api/v1/approval-workflows': { request: never; response: external_training.ApprovalWorkflow[] };
  'GET /api/v1/assignments': { request: never; response: learning.Assignment[] };
  'GET /api/v1/audit-logs': { request: never; response: audit.LogEntry[] };
  'GET /api/v1/auth/me': { request: never; response: identity.MeResponse };
  'GET /api/v1/budget-limits': { request: never; response: external_training.BudgetLimit[] };
  'GET /api/v1/calendar/events/upcoming': { request: never; response: calendar.UpcomingCalendarEvent[] };
  'GET /api/v1/certificates/my': { request: never; response: certificates.Certificate[] };
  'GET /api/v1/courses': { request: never; response: catalog.Course[] };
  'GET /api/v1/courses/{id}': { request: never; response: catalog.Course };
  'GET /api/v1/courses/{id}/materials': { request: never; response: catalog.CourseMaterial[] };
  'GET /api/v1/dashboard/employee': { request: never; response: dashboard_api.EmployeeDashboard };
  'GET /api/v1/dashboard/manager': { request: never; response: dashboard_api.ManagerDashboard };
  'GET /api/v1/enrollments/{id}': { request: never; response: learning.Enrollment };
  'GET /api/v1/enrollments/my': { request: never; response: learning.Enrollment[] };
  'GET /api/v1/external-requests': { request: never; response: external_training.ExternalRequest[] };
  'GET /api/v1/external-requests/{id}': { request: never; response: external_training.ExternalRequest };
  'GET /api/v1/external-requests/my': { request: never; response: external_training.ExternalRequest[] };
  'GET /api/v1/external-requests/pending-approvals': { request: never; response: external_training.PendingApprovalItem[] };
  'GET /api/v1/integrations/github/analytics/languages': { request: unknown; response: unknown };
  'GET /api/v1/integrations/github/analytics/repository-health': { request: unknown; response: unknown };
  'GET /api/v1/integrations/github/analytics/repository-ownership': { request: unknown; response: unknown };
  'GET /api/v1/integrations/github/analytics/team': { request: unknown; response: unknown };
  'GET /api/v1/integrations/github/analytics/top-languages': { request: unknown; response: unknown };
  'GET /api/v1/integrations/github/connections': { request: unknown; response: unknown };
  'GET /api/v1/integrations/github/connections/{connectionId}': { request: unknown; response: unknown };
  'GET /api/v1/integrations/github/connections/{connectionId}/mappings': { request: unknown; response: unknown };
  'GET /api/v1/integrations/github/employees/{employeeUserId}/activity': { request: unknown; response: unknown };
  'GET /api/v1/integrations/github/employees/{employeeUserId}/languages': { request: unknown; response: unknown };
  'GET /api/v1/integrations/github/employees/{employeeUserId}/profile': { request: unknown; response: unknown };
  'GET /api/v1/integrations/github/employees/{employeeUserId}/stats': { request: unknown; response: unknown };
  'GET /api/v1/integrations/github/repositories': { request: unknown; response: unknown };
  'GET /api/v1/integrations/github/repositories/{repoId}': { request: unknown; response: unknown };
  'GET /api/v1/integrations/github/repositories/{repoId}/contributors': { request: unknown; response: unknown };
  'GET /api/v1/integrations/github/repositories/{repoId}/languages': { request: unknown; response: unknown };
  'GET /api/v1/integrations/github/sync-jobs/{jobId}': { request: unknown; response: unknown };
  'GET /api/v1/integrations/github/users': { request: unknown; response: unknown };
  'GET /api/v1/integrations/outlook/connect': { request: never; response: outlook.ConnectResponse };
  'GET /api/v1/integrations/outlook/status': { request: never; response: outlook.IntegrationStatus };
  'GET /api/v1/integrations/yougile/connections/{id}/boards': { request: unknown; response: unknown };
  'GET /api/v1/integrations/yougile/connections/{id}/columns': { request: unknown; response: unknown };
  'GET /api/v1/integrations/yougile/connections/{id}/mappings': { request: unknown; response: unknown };
  'GET /api/v1/integrations/yougile/connections/{id}/projects': { request: unknown; response: unknown };
  'GET /api/v1/integrations/yougile/connections/{id}/users': { request: unknown; response: unknown };
  'GET /api/v1/integrations/yougile/sync-jobs/{jobId}': { request: unknown; response: unknown };
  'GET /api/v1/jira/board-summary': { request: never; response: board_summary.BoardSummary };
  'GET /api/v1/learning-plan/my': { request: never; response: learning_plan.MyLearningPlan };
  'GET /api/v1/notifications': { request: never; response: notifications.Notification[] };
  'GET /api/v1/programs': { request: never; response: university.Program[] };
  'GET /api/v1/programs/{id}': { request: never; response: university.Program };
  'GET /api/v1/recommendations/courses': { request: never; response: learning_plan.RecommendedCourseItem[] };
  'GET /api/v1/reports/export/excel': { request: never; response: Blob };
  'GET /api/v1/reports/export/pdf': { request: never; response: Blob };
  'GET /api/v1/tests/{id}': { request: never; response: testing.Test };
  'GET /api/v1/tests/{id}/results': { request: never; response: testing.TestResult[] };
  'PATCH /api/v1/courses/{id}': { request: catalog.UpdateCourseRequest; response: catalog.Course };
  'PATCH /api/v1/external-requests/{id}': { request: external_training.UpdateExternalRequestRequest; response: external_training.ExternalRequest };
  'PATCH /api/v1/integrations/github/connections/{connectionId}': { request: unknown; response: unknown };
  'POST /api/v1/approval-workflows': { request: external_training.CreateWorkflowRequest; response: external_training.ApprovalWorkflow };
  'POST /api/v1/assignments': { request: learning.CreateAssignmentRequest; response: learning.Assignment };
  'POST /api/v1/auth/forgot-password': { request: identity.ForgotPasswordRequest; response: Record<string, unknown> };
  'POST /api/v1/auth/login': { request: identity.LoginRequest; response: identity.AuthResponse };
  'POST /api/v1/auth/logout': { request: identity.LogoutRequest; response: Record<string, unknown> };
  'POST /api/v1/auth/refresh': { request: identity.RefreshRequest; response: identity.AuthResponse };
  'POST /api/v1/auth/register': { request: identity.RegisterRequest; response: identity.AuthResponse };
  'POST /api/v1/auth/reset-password': { request: identity.ResetPasswordRequest; response: Record<string, unknown> };
  'POST /api/v1/budget-limits': { request: external_training.CreateBudgetLimitRequest; response: external_training.BudgetLimit };
  'POST /api/v1/certificates/{id}/reject': { request: certificates.ReviewRequest; response: certificates.Certificate };
  'POST /api/v1/certificates/{id}/verify': { request: certificates.ReviewRequest; response: certificates.Certificate };
  'POST /api/v1/certificates/upload': { request: certificates.UploadCertificateRequest; response: certificates.Certificate };
  'POST /api/v1/courses': { request: catalog.CreateCourseRequest; response: catalog.Course };
  'POST /api/v1/courses/{id}/archive': { request: never; response: catalog.Course };
  'POST /api/v1/courses/{id}/materials': { request: catalog.CreateMaterialRequest; response: catalog.CourseMaterial };
  'POST /api/v1/courses/{id}/publish': { request: never; response: catalog.Course };
  'POST /api/v1/enrollments/{id}/complete': { request: learning.CompleteRequest; response: learning.Enrollment };
  'POST /api/v1/enrollments/{id}/progress': { request: learning.ProgressRequest; response: learning.Enrollment };
  'POST /api/v1/enrollments/{id}/start': { request: never; response: learning.Enrollment };
  'POST /api/v1/external-requests': { request: external_training.CreateExternalRequestRequest; response: external_training.ExternalRequest };
  'POST /api/v1/external-requests/{id}/approve': { request: external_training.ActionCommentRequest; response: external_training.ExternalRequest };
  'POST /api/v1/external-requests/{id}/reject': { request: external_training.ActionCommentRequest; response: external_training.ExternalRequest };
  'POST /api/v1/external-requests/{id}/request-revision': { request: external_training.ActionCommentRequest; response: external_training.ExternalRequest };
  'POST /api/v1/external-requests/{id}/submit': { request: never; response: external_training.ExternalRequest };
  'POST /api/v1/external-requests/{id}/upload-certificate': { request: external_training.UploadRequestCertificateRequest; response: external_training.ExternalRequest };
  'POST /api/v1/groups/{id}/participants': { request: university.AddParticipantRequest; response: university.TrainingGroup };
  'POST /api/v1/groups/{id}/sessions': { request: university.CreateSessionRequest; response: university.Session };
  'POST /api/v1/integrations/github/connections': { request: unknown; response: unknown };
  'POST /api/v1/integrations/github/connections/{connectionId}/import/languages': { request: unknown; response: unknown };
  'POST /api/v1/integrations/github/connections/{connectionId}/import/repos': { request: unknown; response: unknown };
  'POST /api/v1/integrations/github/connections/{connectionId}/import/users': { request: unknown; response: unknown };
  'POST /api/v1/integrations/github/connections/{connectionId}/mappings': { request: unknown; response: unknown };
  'POST /api/v1/integrations/github/connections/{connectionId}/mappings/auto-match': { request: unknown; response: unknown };
  'POST /api/v1/integrations/github/connections/{connectionId}/sync': { request: unknown; response: unknown };
  'POST /api/v1/integrations/github/connections/test': { request: unknown; response: unknown };
  'POST /api/v1/integrations/outlook/disconnect': { request: never; response: outlook.IntegrationStatus };
  'POST /api/v1/integrations/outlook/sync': { request: never; response: outlook.IntegrationStatus };
  'POST /api/v1/integrations/yougile/connections': { request: unknown; response: unknown };
  'POST /api/v1/integrations/yougile/connections/{id}/import/structure': { request: unknown; response: unknown };
  'POST /api/v1/integrations/yougile/connections/{id}/import/users': { request: unknown; response: unknown };
  'POST /api/v1/integrations/yougile/connections/{id}/mappings': { request: unknown; response: unknown };
  'POST /api/v1/integrations/yougile/connections/{id}/mappings/auto-match': { request: unknown; response: unknown };
  'POST /api/v1/integrations/yougile/connections/{id}/sync': { request: unknown; response: unknown };
  'POST /api/v1/integrations/yougile/connections/{id}/sync/backfill': { request: unknown; response: unknown };
  'POST /api/v1/integrations/yougile/connections/create-key': { request: unknown; response: unknown };
  'POST /api/v1/integrations/yougile/connections/test-key': { request: unknown; response: unknown };
  'POST /api/v1/integrations/yougile/discover-companies': { request: unknown; response: unknown };
  'POST /api/v1/notifications/{id}/read': { request: never; response: notifications.Notification };
  'POST /api/v1/notifications/read-all': { request: never; response: Record<string, unknown> };
  'POST /api/v1/programs': { request: university.CreateProgramRequest; response: university.Program };
  'POST /api/v1/programs/{id}/groups': { request: university.CreateGroupRequest; response: university.TrainingGroup };
  'POST /api/v1/sessions/{id}/participant-feedback': { request: university.ParticipantFeedbackRequest; response: Record<string, unknown> };
  'POST /api/v1/sessions/{id}/trainer-feedback': { request: university.TrainerFeedbackRequest; response: Record<string, unknown> };
  'POST /api/v1/test-attempts/{id}/answers': { request: testing.SubmitAnswersRequest; response: testing.TestAttempt };
  'POST /api/v1/test-attempts/{id}/submit': { request: never; response: testing.TestAttempt };
  'POST /api/v1/tests/{id}/attempts': { request: testing.StartAttemptRequest; response: testing.TestAttempt };
};
```

### Endpoints with unknown typing

- DELETE /api/v1/integrations/github/connections/{connectionId}
- DELETE /api/v1/integrations/github/connections/{connectionId}/mappings/{mappingId}
- DELETE /api/v1/integrations/yougile/connections/{id}/mappings/{mappingId}
- GET /api/v1/integrations/github/analytics/languages
- GET /api/v1/integrations/github/analytics/repository-health
- GET /api/v1/integrations/github/analytics/repository-ownership
- GET /api/v1/integrations/github/analytics/team
- GET /api/v1/integrations/github/analytics/top-languages
- GET /api/v1/integrations/github/connections
- GET /api/v1/integrations/github/connections/{connectionId}
- GET /api/v1/integrations/github/connections/{connectionId}/mappings
- GET /api/v1/integrations/github/employees/{employeeUserId}/activity
- GET /api/v1/integrations/github/employees/{employeeUserId}/languages
- GET /api/v1/integrations/github/employees/{employeeUserId}/profile
- GET /api/v1/integrations/github/employees/{employeeUserId}/stats
- GET /api/v1/integrations/github/repositories
- GET /api/v1/integrations/github/repositories/{repoId}
- GET /api/v1/integrations/github/repositories/{repoId}/contributors
- GET /api/v1/integrations/github/repositories/{repoId}/languages
- GET /api/v1/integrations/github/sync-jobs/{jobId}
- GET /api/v1/integrations/github/users
- GET /api/v1/integrations/yougile/connections/{id}/boards
- GET /api/v1/integrations/yougile/connections/{id}/columns
- GET /api/v1/integrations/yougile/connections/{id}/mappings
- GET /api/v1/integrations/yougile/connections/{id}/projects
- GET /api/v1/integrations/yougile/connections/{id}/users
- GET /api/v1/integrations/yougile/sync-jobs/{jobId}
- PATCH /api/v1/integrations/github/connections/{connectionId}
- POST /api/v1/integrations/github/connections
- POST /api/v1/integrations/github/connections/{connectionId}/import/languages
- POST /api/v1/integrations/github/connections/{connectionId}/import/repos
- POST /api/v1/integrations/github/connections/{connectionId}/import/users
- POST /api/v1/integrations/github/connections/{connectionId}/mappings
- POST /api/v1/integrations/github/connections/{connectionId}/mappings/auto-match
- POST /api/v1/integrations/github/connections/{connectionId}/sync
- POST /api/v1/integrations/github/connections/test
- POST /api/v1/integrations/yougile/connections
- POST /api/v1/integrations/yougile/connections/{id}/import/structure
- POST /api/v1/integrations/yougile/connections/{id}/import/users
- POST /api/v1/integrations/yougile/connections/{id}/mappings
- POST /api/v1/integrations/yougile/connections/{id}/mappings/auto-match
- POST /api/v1/integrations/yougile/connections/{id}/sync
- POST /api/v1/integrations/yougile/connections/{id}/sync/backfill
- POST /api/v1/integrations/yougile/connections/create-key
- POST /api/v1/integrations/yougile/connections/test-key
- POST /api/v1/integrations/yougile/discover-companies

## 2) Full Type Definitions

Ниже типы из ENTITY_TS_REFERENCE.md (без изменений).

# Backend Entities -> TypeScript Interfaces

Автоген из `internal/core` + `internal/modules`: экспортируемые структуры с `json`-тегами.

## Shared Scalar Types

```ts
export type UUID = string;
export type Money = string;
export type DateTimeString = string;
```

## Namespace: accounts

```ts
export namespace accounts {
  export type Kind = 'bank_account' | 'bank_card' | 'cash' | 'savings' | 'virtual';

  export interface Account {
    id: UUID;
    user_id: UUID;
    name: string;
    kind: Kind;
    currency: string;
    opening_balance: Money;
    current_balance: Money;
    is_archived: boolean;
    last_recalculated_at?: DateTimeString | null;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface CreateAccountRequest {
    name: string;
    kind: Kind;
    currency: string;
    opening_balance: Money;
  }

  export interface UpdateAccountRequest {
    name: string | null;
    kind: Kind | null;
    is_archived: boolean | null;
  }
}
```

## Namespace: admin

```ts
export namespace admin {
  export interface AssignRoleRequest {
    role_code?: string;
  }

  export interface CreateUserRequest {
    email: string;
    status: string;
    first_name: string;
    last_name: string;
    middle_name?: string | null;
    position_title?: string | null;
    department_id?: UUID | null;
    timezone?: string | null;
    role_codes?: string[];
    password?: string | null;
  }

  export interface UpdateUserRequest {
    email?: string | null;
    status?: string | null;
    first_name?: string | null;
    last_name?: string | null;
    middle_name?: string | null;
    position_title?: string | null;
    department_id?: UUID | null;
    timezone?: string | null;
    outlook_email?: string | null;
  }
}
```

## Namespace: audit

```ts
export namespace audit {
  export interface Event {
    id: UUID;
    user_id: UUID;
    action: string;
    entity_type: string;
    entity_id?: UUID | null;
    meta?: Record<string, unknown>;
    source: string;
    request_id?: string | null;
    session_id?: UUID | null;
    change_set?: Record<string, unknown>;
    actor_type: string;
    actor_id?: UUID | null;
    created_at: DateTimeString;
  }

  export interface LogEntry {
    id: string;
    actor_user_id?: string | null;
    user_id?: string | null;
    entity_type: string;
    entity_id?: string | null;
    action: string;
    meta?: Record<string, any>;
    created_at: DateTimeString;
  }
}
```

## Namespace: auth

```ts
export namespace auth {
  export type Provider = 'telegram' | 'yandex';

  export interface AuthResponse {
    user: User;
    tokens: Tokens;
    meta?: Record<string, any>;
  }

  export interface Identity {
    id: UUID;
    user_id: UUID;
    provider: Provider;
    provider_user_id: string;
    provider_email?: string | null;
    created_at: DateTimeString;
  }

  export interface LoginResult {
    user: User;
    tokens: Record<string, any>;
    meta?: Record<string, any>;
  }

  export interface LogoutRequest {
    refresh_token: string;
  }

  export interface RefreshRequest {
    refresh_token: string;
  }

  export interface TelegramLoginRequest {
    provider_user_id: string;
    username: string | null;
    first_name: string | null;
    last_name: string | null;
    photo_url: string | null;
    auth_date: number;
    hash: string | null;
  }

  export interface YandexLoginRequest {
    code: string | null;
    id_token: string | null;
    provider_user_id: string | null;
    email: string | null;
    display_name: string | null;
    avatar_url: string | null;
  }
}
```

## Namespace: board_summary

```ts
export namespace board_summary {
  export interface BoardItem {
    board_id: string;
    title: string;
    tasks_total: number;
    active_total: number;
    completed_total: number;
    overdue_total: number;
  }

  export interface BoardSummary {
    source: string;
    status: string;
    summary: Summary;
    boards: BoardItem[];
    overdue_tasks: OverdueTask[];
  }

  export interface OverdueTask {
    task_id: string;
    board_id?: string | null;
    board_title?: string | null;
    title: string;
    deadline_at?: DateTimeString | null;
    completed: boolean;
    archived: boolean;
  }

  export interface Summary {
    boards_total: number;
    tasks_total: number;
    active_total: number;
    completed_total: number;
    overdue_total: number;
  }
}
```

## Namespace: calendar

```ts
export namespace calendar {
  export interface UpcomingCalendarEvent {
    id: UUID;
    source_type: string;
    source_id: UUID;
    provider: string;
    external_event_id?: string | null;
    title: string;
    start_at: DateTimeString;
    end_at: DateTimeString;
    timezone?: string | null;
    status: string;
    meeting_url?: string | null;
    location?: string | null;
  }
}
```

## Namespace: catalog

```ts
export namespace catalog {
  export interface Course {
    id: UUID;
    type: string;
    source_type: string;
    title: string;
    slug?: string | null;
    short_description?: string | null;
    description?: string | null;
    provider_id?: UUID | null;
    category_id?: UUID | null;
    direction_id?: UUID | null;
    level?: string | null;
    duration_hours?: string | null;
    language?: string | null;
    is_mandatory_default: boolean;
    status: string;
    external_url?: string | null;
    price?: string | null;
    price_currency?: string | null;
    next_start_date?: DateTimeString | null;
    thumbnail_file_id?: UUID | null;
    created_by?: UUID | null;
    updated_by?: UUID | null;
    published_at?: DateTimeString | null;
    archived_at?: DateTimeString | null;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface CourseMaterial {
    id: UUID;
    course_id: UUID;
    type: string;
    title: string;
    description?: string | null;
    file_id?: UUID | null;
    external_url?: string | null;
    sort_order: number;
    is_required: boolean;
    created_at: DateTimeString;
  }

  export interface CreateCourseRequest {
    type: string;
    source_type: string;
    title: string;
    slug?: string | null;
    short_description?: string | null;
    description?: string | null;
    provider_id?: UUID | null;
    category_id?: UUID | null;
    direction_id?: UUID | null;
    level?: string | null;
    duration_hours?: string | null;
    language?: string | null;
    is_mandatory_default: boolean;
    external_url?: string | null;
    price?: string | null;
    price_currency?: string | null;
    next_start_date?: DateTimeString | null;
    thumbnail_file_id?: UUID | null;
  }

  export interface CreateMaterialRequest {
    type: string;
    title: string;
    description?: string | null;
    file_id?: UUID | null;
    external_url?: string | null;
    sort_order: number;
    is_required: boolean;
  }

  export interface UpdateCourseRequest {
    title?: string | null;
    slug?: string | null;
    short_description?: string | null;
    description?: string | null;
    provider_id?: UUID | null;
    category_id?: UUID | null;
    direction_id?: UUID | null;
    level?: string | null;
    duration_hours?: string | null;
    language?: string | null;
    is_mandatory_default?: boolean | null;
    external_url?: string | null;
    price?: string | null;
    price_currency?: string | null;
    next_start_date?: DateTimeString | null;
    thumbnail_file_id?: UUID | null;
  }
}
```

## Namespace: categories

```ts
export namespace categories {
  export type Kind = 'expense' | 'income';

  export interface Category {
    id: UUID;
    user_id?: UUID | null;
    kind: Kind;
    name: string;
    color?: string | null;
    icon?: string | null;
    parent_id?: UUID | null;
    is_system: boolean;
    is_archived: boolean;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface CreateCategoryRequest {
    kind: Kind;
    name: string;
    color: string | null;
    icon: string | null;
    parent_id: UUID | null;
  }

  export interface UpdateCategoryRequest {
    name: string | null;
    color: string | null;
    icon: string | null;
    parent_id: UUID | null;
    is_archived: boolean | null;
  }
}
```

## Namespace: certificates

```ts
export namespace certificates {
  export interface Certificate {
    id: UUID;
    user_id: UUID;
    course_id?: UUID | null;
    enrollment_id?: UUID | null;
    certificate_no?: string | null;
    issued_by?: string | null;
    issued_at?: DateTimeString | null;
    expires_at?: DateTimeString | null;
    status: string;
    file_id: UUID;
    uploaded_at: DateTimeString;
    verified_at?: DateTimeString | null;
    verified_by?: UUID | null;
    notes?: string | null;
  }

  export interface FileAttachment {
    id: UUID;
    storage_provider: string;
    storage_key: string;
    original_name: string;
    mime_type: string;
    size_bytes: number;
    uploaded_by?: UUID | null;
    created_at: DateTimeString;
  }

  export interface ReviewRequest {
    comment?: string | null;
  }

  export interface UploadCertificateRequest {
    course_id?: UUID | null;
    enrollment_id?: UUID | null;
    certificate_no?: string | null;
    issued_by?: string | null;
    issued_at?: DateTimeString | null;
    expires_at?: DateTimeString | null;
    notes?: string | null;
    storage_provider: string;
    storage_key: string;
    original_name: string;
    mime_type: string;
    size_bytes: number;
  }
}
```

## Namespace: cicd

```ts
export namespace cicd {
  export interface SmokeCheck {
    id: UUID;
    user_id: UUID;
    session_id: UUID;
    request_id: string;
    trigger: string;
    created_at: DateTimeString;
  }

  export interface SmokeCheckResult {
    check: SmokeCheck;
    total_runs: number;
  }
}
```

## Namespace: common

```ts
export namespace common {
  export interface EntityRef {
    type: string;
    id: UUID;
  }

  export interface Pagination {
    limit: number;
    offset: number;
  }
}
```

## Namespace: course_requests

```ts
export namespace course_requests {
  export interface ActionCommentRequest {
    comment?: string | null;
  }

  export interface CourseRequest {
    id: UUID;
    request_no: string;
    course_id: UUID;
    course_title: string;
    employee_user_id: UUID;
    employee_full_name: string;
    employee_email: string;
    department_id?: UUID | null;
    manager_user_id?: UUID | null;
    manager_full_name?: string | null;
    hr_user_id?: UUID | null;
    hr_full_name?: string | null;
    enrollment_id?: UUID | null;
    certificate_id?: UUID | null;
    certificate_original_name?: string | null;
    status: string;
    display_status: string;
    status_label: string;
    certificate_approval_summary: string;
    employee_comment?: string | null;
    manager_comment?: string | null;
    hr_comment?: string | null;
    rejection_reason?: string | null;
    deadline_at?: DateTimeString | null;
    requested_at: DateTimeString;
    manager_approved_at?: DateTimeString | null;
    hr_approved_at?: DateTimeString | null;
    approved_at?: DateTimeString | null;
    started_at?: DateTimeString | null;
    completed_at?: DateTimeString | null;
    certificate_uploaded_at?: DateTimeString | null;
    certificate_approved_at?: DateTimeString | null;
    certificate_manager_approved_at?: DateTimeString | null;
    certificate_manager_approved_by?: UUID | null;
    certificate_hr_approved_at?: DateTimeString | null;
    certificate_hr_approved_by?: UUID | null;
    canceled_at?: DateTimeString | null;
    rejected_at?: DateTimeString | null;
    rejected_by?: UUID | null;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface CreateCourseRequestRequest {
    course_id: UUID;
    employee_comment?: string | null;
    deadline_at?: DateTimeString | null;
  }

  export interface UploadCertificateRequest {
    certificate_no?: string | null;
    issued_by?: string | null;
    issued_at?: DateTimeString | null;
    expires_at?: DateTimeString | null;
    notes?: string | null;
    storage_provider: string;
    storage_key: string;
    original_name: string;
    mime_type: string;
    size_bytes: number;
  }
}
```

## Namespace: dashboard

```ts
export namespace dashboard {
  export interface FinanceDashboard {
    current_balance: Money;
    monthly_income: Money;
    monthly_expense: Money;
    saved_this_month: Money;
    safe_to_spend: Money;
    top_categories: TopCategory[];
    savings: GoalProgress[];
    weekly_review: WeeklyReview;
    insights: string[];
  }
}
```

## Namespace: dashboard_api

```ts
export namespace dashboard_api {
  export interface EmployeeDashboard {
    stats: EmployeeStats;
    upcoming_events: UpcomingCalendarEvent[];
    recommended_courses: RecommendedCourseItem[];
    learning_plan: MyLearningPlan;
    external_requests_preview: ExternalRequest[];
  }

  export interface EmployeeStats {
    active_enrollments: number;
    recommended_courses: number;
    open_external_requests: number;
  }

  export interface ManagerDashboard {
    stats: ManagerStats;
    team_preview: TeamPreviewItem[];
    pending_external_approvals: PendingApprovalItem[];
    team_external_requests: ExternalRequest[];
    team_course_requests: CourseRequest[];
  }

  export interface ManagerStats {
    team_size: number;
    pending_external_approvals: number;
    team_external_requests: number;
    team_course_requests: number;
  }

  export interface TeamPreviewItem {
    user_id: UUID;
    first_name: string;
    last_name: string;
    position_title?: string | null;
    department_id?: UUID | null;
  }
}
```

## Namespace: external_training

```ts
export namespace external_training {
  export interface ActionCommentRequest {
    comment?: string | null;
  }

  export interface ApprovalStep {
    id: UUID;
    entity_type: string;
    entity_id: UUID;
    step_order: number;
    approver_user_id: UUID;
    role_code: string;
    status: string;
    comment?: string | null;
    due_at?: DateTimeString | null;
    acted_at?: DateTimeString | null;
    created_at: DateTimeString;
  }

  export interface ApprovalWorkflow {
    id: UUID;
    entity_type: string;
    name: string;
    is_active: boolean;
    created_at: DateTimeString;
    updated_at: DateTimeString;
    steps?: ApprovalWorkflowStep[];
  }

  export interface ApprovalWorkflowStep {
    id: UUID;
    workflow_id: UUID;
    step_order: number;
    role_code: string;
    approver_source: string;
    approver_user_id?: UUID | null;
    sla_hours?: number | null;
    is_required: boolean;
  }

  export interface BudgetLimit {
    id: UUID;
    scope_type: string;
    scope_id?: UUID | null;
    period_year: number;
    period_month?: number | null;
    limit_amount: string;
    currency: string;
    is_active: boolean;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface CreateBudgetLimitRequest {
    scope_type: string;
    scope_id?: UUID | null;
    period_year: number;
    period_month?: number | null;
    limit_amount: string;
    currency: string;
    is_active: boolean;
  }

  export interface CreateExternalRequestRequest {
    title: string;
    provider_id?: UUID | null;
    provider_name?: string | null;
    course_url?: string | null;
    program_description?: string | null;
    planned_start_date?: DateTimeString | null;
    planned_end_date?: DateTimeString | null;
    duration_hours?: string | null;
    cost_amount: string;
    currency: string;
    business_goal?: string | null;
    employee_comment?: string | null;
  }

  export interface CreateWorkflowRequest {
    entity_type: string;
    name: string;
    is_active: boolean;
    steps: CreateWorkflowStepRequest[];
  }

  export interface CreateWorkflowStepRequest {
    step_order: number;
    role_code: string;
    approver_source: string;
    approver_user_id?: UUID | null;
    sla_hours?: number | null;
    is_required: boolean;
  }

  export interface ExternalRequest {
    id: UUID;
    request_no: string;
    employee_user_id: UUID;
    department_id?: UUID | null;
    title: string;
    provider_id?: UUID | null;
    provider_name?: string | null;
    course_url?: string | null;
    program_description?: string | null;
    planned_start_date?: DateTimeString | null;
    planned_end_date?: DateTimeString | null;
    duration_hours?: string | null;
    cost_amount: string;
    currency: string;
    business_goal?: string | null;
    employee_comment?: string | null;
    manager_comment?: string | null;
    hr_comment?: string | null;
    status: string;
    calendar_conflict_status?: string | null;
    budget_check_status?: string | null;
    current_approval_step_id?: UUID | null;
    approved_at?: DateTimeString | null;
    rejected_at?: DateTimeString | null;
    sent_to_revision_at?: DateTimeString | null;
    training_started_at?: DateTimeString | null;
    training_completed_at?: DateTimeString | null;
    certificate_uploaded_at?: DateTimeString | null;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface PendingApprovalItem {
    request: ExternalRequest;
    current_step: PendingApprovalStepDTO;
  }

  export interface PendingApprovalStepDTO {
    step_id: UUID;
    role_code: string;
    due_at?: DateTimeString | null;
    approver_user_id: UUID;
  }

  export interface UpdateExternalRequestRequest {
    title?: string | null;
    provider_id?: UUID | null;
    provider_name?: string | null;
    course_url?: string | null;
    program_description?: string | null;
    planned_start_date?: DateTimeString | null;
    planned_end_date?: DateTimeString | null;
    duration_hours?: string | null;
    cost_amount?: string | null;
    currency?: string | null;
    business_goal?: string | null;
    employee_comment?: string | null;
    manager_comment?: string | null;
    hr_comment?: string | null;
  }

  export interface UploadRequestCertificateRequest {
    storage_provider: string;
    storage_key: string;
    original_name: string;
    mime_type: string;
    size_bytes: number;
  }
}
```

## Namespace: github_integration

```ts
export namespace github_integration {
  export interface AutoMatchRequest {
    strategy: string;
  }

  export interface Connection {
    id: UUID;
    company_id?: UUID | null;
    title: string;
    provider: string;
    auth_mode: string;
    base_url: string;
    status: string;
    token_last4?: string | null;
    github_app_id?: string | null;
    github_installation_id?: string | null;
    created_by: UUID;
    last_sync_at?: DateTimeString | null;
    last_success_sync_at?: DateTimeString | null;
    last_error?: string | null;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface CreateConnectionRequest {
    title: string;
    authMode: string;
    token?: string | null;
    baseUrl: string;
  }

  export interface CreateMappingRequest {
    employeeUserId: UUID;
    githubLogin: string;
  }

  export interface EmployeeActivityPoint {
    date: string;
    commitCount: number;
    openedPRs: number;
    mergedPRs: number;
    reviewedPRs: number;
  }

  export interface EmployeeLanguage {
    name: string;
    bytes: number;
    percent: string;
    reposCount: number;
    lastUsedAt?: DateTimeString | null;
  }

  export interface EmployeeStats {
    repositoriesCount: number;
    activeRepositoriesCount: number;
    commits: number;
    openedPRs: number;
    mergedPRs: number;
    reviewedPRs: number;
    starsReceived: number;
    forksReceived: number;
    avgRepoFreshnessDays: string;
    primaryLanguages: string[];
    engineeringActivityScore: string;
    dataScope: string;
  }

  export interface GitHubUser {
    id: UUID;
    connection_id: UUID;
    github_user_id: number;
    login: string;
    name?: string | null;
    email?: string | null;
    avatar_url?: string | null;
    html_url?: string | null;
    company?: string | null;
    location?: string | null;
    bio?: string | null;
    followers?: number | null;
    following?: number | null;
    public_repos?: number | null;
    public_gists?: number | null;
    created_at_remote?: DateTimeString | null;
    updated_at_remote?: DateTimeString | null;
    synced_at: DateTimeString;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface ImportResponse {
    imported: number;
    updated: number;
    failed: number;
  }

  export interface LanguageAnalyticsItem {
    name: string;
    percent: string;
  }

  export interface MappingView {
    id: UUID;
    employee_user_id: UUID;
    employee_name: string;
    employee_email: string;
    github_login: string;
    github_user_id?: number | null;
    profile_url?: string | null;
    match_source: string;
    is_active: boolean;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface RepositoryContributor {
    id: UUID;
    repository_id: UUID;
    github_user_id?: number | null;
    github_login: string;
    contributions: number;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface RepositoryDetail {
    languages: RepositoryLanguage[];
    contributors: RepositoryContributor[];
  }

  export interface RepositoryHealthItem {
    repositoryId: UUID;
    fullName: string;
    archived: boolean;
    openIssues: number;
    freshnessDays: string;
  }

  export interface RepositoryLanguage {
    id: UUID;
    repository_id: UUID;
    language_name: string;
    bytes: number;
    percent: string;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface RepositoryOwnershipItem {
    employeeUserId: UUID;
    employeeName: string;
    repositories: number;
    activeRepositories: number;
  }

  export interface RepoSummary {
    repositoriesCount: number;
    privateCount: number;
    publicCount: number;
  }

  export interface SyncJob {
    id: UUID;
    connection_id: UUID;
    job_type: string;
    status: string;
    cursor?: Record<string, unknown>;
    progress: Record<string, unknown>;
    attempt: number;
    started_at?: DateTimeString | null;
    finished_at?: DateTimeString | null;
    next_retry_at?: DateTimeString | null;
    error_text?: string | null;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface SyncRequest {
    mode: string;
    includeUsers: boolean;
    includeRepos: boolean;
    includeLanguages: boolean;
    includeContributors: boolean;
    includeActivity: boolean;
  }

  export interface TeamAnalyticsEmployee {
    employeeUserId: UUID;
    name: string;
    primaryLanguage: string;
    activeRepositories: number;
    commits: number;
    mergedPRs: number;
  }

  export interface TestConnectionRequest {
    token: string;
    baseUrl: string;
  }

  export interface TestConnectionResponse {
    ok: boolean;
    scopes: string;
    rateLimit: string;
    message: string;
  }

  export interface UpdateConnectionRequest {
    title?: string | null;
    status?: string | null;
  }
}
```

## Namespace: identity

```ts
export namespace identity {
  export interface AuthResponse {
    user: UserView;
    tokens: Tokens;
    meta?: Record<string, any>;
  }

  export interface ForgotPasswordRequest {
    email: string;
  }

  export interface LoginRequest {
    email: string;
    password: string;
  }

  export interface LogoutRequest {
    refresh_token: string;
  }

  export interface MeResponse {
    user: UserView;
  }

  export interface Permission {
    id: UUID;
    code: string;
    module: string;
    action: string;
    description?: string | null;
  }

  export interface RefreshRequest {
    refresh_token: string;
  }

  export interface RegisterRequest {
    email: string;
    password: string;
    first_name: string;
    last_name: string;
    middle_name?: string | null;
    position_title?: string | null;
  }

  export interface ResetPasswordRequest {
    token: string;
    new_password: string;
  }

  export interface Role {
    id: UUID;
    code: string;
    name: string;
    description?: string | null;
    is_system: boolean;
  }

  export interface Tokens {
    access_token: string;
    refresh_token: string;
    expires_in: number;
  }

  export interface User {
    id: UUID;
    email: string;
    status: string;
    is_email_verified: boolean;
    last_login_at?: DateTimeString | null;
    created_at: DateTimeString;
    updated_at: DateTimeString;
    deleted_at?: DateTimeString | null;
  }

  export interface UserView {
    roles: string[];
    permissions: string[];
    employee_profile?: EmployeeProfile | null;
  }
}
```

## Namespace: jobs

```ts
export namespace jobs {
  export type JobType = 'rebuild_summary' | 'savings_alert' | 'weekly_review';
}
```

## Namespace: learning

```ts
export namespace learning {
  export interface Assignment {
    id: UUID;
    course_id: UUID;
    assignment_type: string;
    target_type: string;
    target_id: UUID;
    assigned_by: UUID;
    priority: string;
    reason?: string | null;
    start_at?: DateTimeString | null;
    deadline_at?: DateTimeString | null;
    status: string;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface CompleteRequest {
    completion_type: string;
    score?: string | null;
    notes?: string | null;
  }

  export interface CreateAssignmentRequest {
    course_id: UUID;
    assignment_type: string;
    target_type: string;
    target_id: UUID;
    priority: string;
    reason?: string | null;
    start_at?: DateTimeString | null;
    deadline_at?: DateTimeString | null;
  }

  export interface Enrollment {
    id: UUID;
    course_id: UUID;
    user_id: UUID;
    assignment_id?: UUID | null;
    source: string;
    status: string;
    enrolled_at: DateTimeString;
    started_at?: DateTimeString | null;
    completed_at?: DateTimeString | null;
    deadline_at?: DateTimeString | null;
    last_activity_at?: DateTimeString | null;
    completion_percent: string;
    is_mandatory: boolean;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface ProgressRequest {
    course_module_id: UUID;
    status: string;
    progress_percent: string;
    completed_at?: DateTimeString | null;
  }
}
```

## Namespace: learning_plan

```ts
export namespace learning_plan {
  export interface LearningPlanItem {
    enrollment_id: UUID;
    course_id: UUID;
    assignment_id?: UUID | null;
    source: string;
    status: string;
    title: string;
    short_description?: string | null;
    deadline_at?: DateTimeString | null;
    started_at?: DateTimeString | null;
    completed_at?: DateTimeString | null;
    completion_percent: string;
    is_mandatory: boolean;
    reason?: string | null;
    enrollment_created_at: DateTimeString;
  }

  export interface LearningPlanSummary {
    total: number;
    in_progress: number;
    upcoming: number;
    completed_recently: number;
    recommended: number;
  }

  export interface MyLearningPlan {
    summary: LearningPlanSummary;
    in_progress: LearningPlanItem[];
    upcoming: LearningPlanItem[];
    completed_recently: LearningPlanItem[];
    recommended: RecommendedCourseItem[];
  }

  export interface RecommendedCourseItem {
    course_id: UUID;
    enrollment_id: UUID;
    assignment_id?: UUID | null;
    title: string;
    short_description?: string | null;
    status: string;
    deadline_at?: DateTimeString | null;
    completion_percent: string;
    reason?: string | null;
    enrollment_created_at: DateTimeString;
  }
}
```

## Namespace: links

```ts
export namespace links {
  export interface CreateLinkRequest {
    source_type: string;
    source_id: UUID;
    target_type: string;
    target_id: UUID;
    relation: string;
    meta: Record<string, any>;
  }

  export interface EntityLink {
    id: UUID;
    user_id: UUID;
    source_type: string;
    source_id: UUID;
    target_type: string;
    target_id: UUID;
    relation: string;
    meta?: Record<string, unknown>;
    created_at: DateTimeString;
  }

  export interface ListByEntityQuery {
    entity_type: string;
    entity_id: UUID;
  }
}
```

## Namespace: notifications

```ts
export namespace notifications {
  export interface Notification {
    id: UUID;
    user_id: UUID;
    channel: string;
    type: string;
    title: string;
    body: string;
    status: string;
    related_entity_type?: string | null;
    related_entity_id?: UUID | null;
    scheduled_at?: DateTimeString | null;
    sent_at?: DateTimeString | null;
    read_at?: DateTimeString | null;
    created_at: DateTimeString;
  }
}
```

## Namespace: org

```ts
export namespace org {
  export interface Department {
    id: UUID;
    name: string;
    code?: string | null;
    parent_id?: UUID | null;
    head_user_id?: UUID | null;
    is_active: boolean;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface EmployeeProfile {
    id: UUID;
    user_id: UUID;
    employee_no?: string | null;
    first_name: string;
    last_name: string;
    middle_name?: string | null;
    position_title?: string | null;
    department_id?: UUID | null;
    hire_date?: DateTimeString | null;
    employment_status: string;
    timezone?: string | null;
    outlook_email?: string | null;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }
}
```

## Namespace: outlook

```ts
export namespace outlook {
  export interface Account {
    id: UUID;
    user_id: UUID;
    external_account_id: string;
    email: string;
    token_expires_at: DateTimeString;
    scope?: string | null;
    status: string;
    last_sync_at?: DateTimeString | null;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface ConnectResponse {
    auth_url: string;
    state: string;
  }

  export interface IntegrationStatus {
    connected: boolean;
    account?: Account | null;
  }
}
```

## Namespace: review

```ts
export namespace review {
  export type Status = 'discrepancy_found' | 'matched' | 'pending' | 'resolved' | 'skipped';

  export interface ResolveRequest {
    resolution_note: string | null;
  }

  export interface SubmitBalanceRequest {
    actual_balance: Money;
  }

  export interface WeeklyReview {
    id: UUID;
    user_id: UUID;
    account_id?: UUID | null;
    period_start: DateTimeString;
    period_end: DateTimeString;
    expected_balance: Money;
    actual_balance?: Money | null;
    delta?: Money | null;
    status: Status;
    resolution_note?: string | null;
    created_at: DateTimeString;
    completed_at?: DateTimeString | null;
  }
}
```

## Namespace: savings

```ts
export namespace savings {
  export type Priority = 'high' | 'low' | 'medium';
  export type Status = 'active' | 'archived' | 'completed' | 'paused';

  export interface CreateGoalRequest {
    title: string;
    target_amount: Money;
    current_amount: Money;
    currency: string;
    target_date: DateTimeString | null;
    priority: Priority;
  }

  export interface Goal {
    id: UUID;
    user_id: UUID;
    title: string;
    target_amount: Money;
    current_amount: Money;
    currency: string;
    target_date?: DateTimeString | null;
    priority: Priority;
    status: Status;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface GoalProgress {
    goal: Goal;
    progress_percent: string;
    recommended_monthly_contribution: Money;
  }

  export interface Summary {
    total_target: Money;
    total_current: Money;
    reserved_this_month: Money;
    safe_to_spend: Money;
    goals: GoalProgress[];
  }

  export interface UpdateGoalRequest {
    title: string | null;
    target_amount: Money | null;
    current_amount: Money | null;
    target_date: DateTimeString | null;
    priority: Priority | null;
    status: Status | null;
  }
}
```

## Namespace: sessions

```ts
export namespace sessions {
  export interface Session {
    id: UUID;
    user_id: UUID;
    user_agent?: string | null;
    ip_address?: string | null;
    expires_at: DateTimeString;
    created_at: DateTimeString;
    revoked_at?: DateTimeString | null;
  }

  export interface Tokens {
    access_token: string;
    refresh_token: string;
    expires_in: number;
  }
}
```

## Namespace: summary

```ts
export namespace summary {
  export interface MonthlySummary {
    current_balance: Money;
    income_total: Money;
    expense_total: Money;
    top_categories: TopCategory[];
  }

  export interface TopCategory {
    category_id?: string | null;
    name: string;
    amount: Money;
  }
}
```

## Namespace: testing

```ts
export namespace testing {
  export interface AnswerInput {
    question_id: UUID;
    answer_text?: string | null;
    selected_option_id?: UUID | null;
    selected_option_ids?: UUID[];
  }

  export interface AnswerOption {
    id: UUID;
    question_id: UUID;
    text: string;
    is_correct: boolean;
    sort_order: number;
  }

  export interface CreateAnswerOptionRequest {
    text: string;
    is_correct: boolean;
    sort_order: number;
  }

  export interface CreateQuestionRequest {
    type: string;
    text: string;
    explanation?: string | null;
    sort_order: number;
    points: string;
    is_required: boolean;
    options?: CreateAnswerOptionRequest[];
  }

  export interface CreateTestRequest {
    course_id?: UUID | null;
    title: string;
    description?: string | null;
    attempts_limit?: number | null;
    passing_score: string;
    shuffle_questions: boolean;
    shuffle_answers: boolean;
    status: string;
    questions?: CreateQuestionRequest[];
  }

  export interface Question {
    id: UUID;
    test_id: UUID;
    type: string;
    text: string;
    explanation?: string | null;
    sort_order: number;
    points: string;
    is_required: boolean;
    options?: AnswerOption[];
  }

  export interface StartAttemptRequest {
    enrollment_id?: UUID | null;
  }

  export interface SubmitAnswersRequest {
    answers: AnswerInput[];
  }

  export interface Test {
    id: UUID;
    course_id?: UUID | null;
    title: string;
    description?: string | null;
    attempts_limit?: number | null;
    passing_score: string;
    shuffle_questions: boolean;
    shuffle_answers: boolean;
    status: string;
    created_by: UUID;
    created_at: DateTimeString;
    updated_at: DateTimeString;
    questions?: Question[];
  }

  export interface TestAttempt {
    id: UUID;
    test_id: UUID;
    user_id: UUID;
    enrollment_id?: UUID | null;
    attempt_no: number;
    status: string;
    started_at: DateTimeString;
    submitted_at?: DateTimeString | null;
    checked_at?: DateTimeString | null;
    score?: string | null;
    passed?: boolean | null;
  }

  export interface TestResult {
    id: UUID;
    test_id: UUID;
    user_id: UUID;
    best_attempt_id: UUID;
    best_score: string;
    passed: boolean;
    completed_at: DateTimeString;
  }
}
```

## Namespace: transactions

```ts
export namespace transactions {
  export type Direction = 'inflow' | 'internal' | 'outflow';
  export type PostingState = 'draft' | 'posted';
  export type Source = 'manual' | 'recurring' | 'review' | 'system';
  export type Type = 'correction' | 'expense' | 'income' | 'transfer';

  export interface CreateTransactionRequest {
    account_id: UUID;
    transfer_account_id: UUID | null;
    type: Type;
    category_id: UUID | null;
    amount: Money;
    currency: string;
    direction: Direction | null;
    title: string | null;
    note: string | null;
    is_mandatory: boolean;
    is_subscription: boolean;
    occurred_at: DateTimeString;
  }

  export interface Transaction {
    id: UUID;
    user_id: UUID;
    account_id: UUID;
    transfer_account_id?: UUID | null;
    type: Type;
    category_id?: UUID | null;
    amount: Money;
    currency: string;
    direction: Direction;
    posting_state: PostingState;
    source: Source;
    source_ref_id?: UUID | null;
    template_id?: UUID | null;
    recurring_rule_id?: UUID | null;
    planned_expense_id?: UUID | null;
    title?: string | null;
    title_normalized?: string | null;
    note?: string | null;
    is_mandatory: boolean;
    is_subscription: boolean;
    base_currency?: string | null;
    base_amount?: Money | null;
    fx_rate?: string | null;
    deleted_at?: DateTimeString | null;
    deleted_by?: UUID | null;
    delete_reason?: string | null;
    restored_at?: DateTimeString | null;
    occurred_at: DateTimeString;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface UpdateTransactionRequest {
    account_id: UUID | null;
    transfer_account_id: UUID | null;
    type: Type | null;
    category_id: UUID | null;
    amount: Money | null;
    currency: string | null;
    direction: Direction | null;
    title: string | null;
    note: string | null;
    is_mandatory: boolean | null;
    is_subscription: boolean | null;
    occurred_at: DateTimeString | null;
  }
}
```

## Namespace: transfers

```ts
export namespace transfers {
  export interface CreateTransferRequest {
    from_account_id: UUID;
    to_account_id: UUID;
    amount: Money;
    currency: string;
    title: string | null;
    note: string | null;
    occurred_at: DateTimeString;
  }

  export interface UpdateTransferRequest {
    from_account_id: UUID | null;
    to_account_id: UUID | null;
    amount: Money | null;
    currency: string | null;
    title: string | null;
    note: string | null;
    occurred_at: DateTimeString | null;
  }
}
```

## Namespace: university

```ts
export namespace university {
  export interface AddParticipantRequest {
    user_id: UUID;
    status: string;
  }

  export interface CreateGroupRequest {
    name: string;
    capacity?: number | null;
    status: string;
    enrollment_open_at?: DateTimeString | null;
    enrollment_close_at?: DateTimeString | null;
  }

  export interface CreateProgramRequest {
    title: string;
    description?: string | null;
    direction_id?: UUID | null;
    status: string;
  }

  export interface CreateSessionRequest {
    trainer_user_id: UUID;
    title: string;
    description?: string | null;
    start_at: DateTimeString;
    end_at: DateTimeString;
    location?: string | null;
    meeting_url?: string | null;
    status: string;
  }

  export interface ParticipantFeedbackRequest {
    program_id?: UUID | null;
    rating: number;
    comment?: string | null;
  }

  export interface Program {
    id: UUID;
    title: string;
    description?: string | null;
    direction_id?: UUID | null;
    status: string;
    created_by: UUID;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface Session {
    id: UUID;
    group_id: UUID;
    trainer_user_id: UUID;
    title: string;
    description?: string | null;
    start_at: DateTimeString;
    end_at: DateTimeString;
    location?: string | null;
    meeting_url?: string | null;
    status: string;
    calendar_event_id?: UUID | null;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface TrainerFeedbackRequest {
    participant_user_id: UUID;
    attendance_status: string;
    score?: string | null;
    comment?: string | null;
  }

  export interface TrainingGroup {
    id: UUID;
    program_id: UUID;
    name: string;
    capacity?: number | null;
    status: string;
    enrollment_open_at?: DateTimeString | null;
    enrollment_close_at?: DateTimeString | null;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }
}
```

## Namespace: users

```ts
export namespace users {
  export interface MeResponse {
    user: User;
  }

  export interface UpdatePreferencesRequest {
    timezone: string | null;
    base_currency: string | null;
    onboarding_completed: boolean | null;
    weekly_review_weekday: number | null;
    weekly_review_hour: number | null;
  }

  export interface User {
    id: UUID;
    email?: string | null;
    display_name?: string | null;
    avatar_url?: string | null;
    timezone: string;
    base_currency: string;
    onboarding_completed: boolean;
    weekly_review_weekday: number;
    weekly_review_hour: number;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }
}
```

## Namespace: yougile

```ts
export namespace yougile {
  export interface AutoMatchRequest {
    strategy: string;
  }

  export interface AutoMatchResponse {
    matched: number;
    unmatchedInternal: number;
    unmatchedYougile: number;
  }

  export interface BackfillRequest {
    from: string;
    to: string;
    employees?: UUID[];
  }

  export interface Board {
    id: UUID;
    connection_id: UUID;
    yougile_board_id: string;
    yougile_project_id: string;
    title: string;
    deleted: boolean;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface Column {
    id: UUID;
    connection_id: UUID;
    yougile_column_id: string;
    yougile_board_id: string;
    title: string;
    color?: number | null;
    deleted: boolean;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface Connection {
    id: UUID;
    company_id?: UUID | null;
    title?: string | null;
    api_base_url: string;
    yougile_company_id: string;
    api_key_last4?: string | null;
    status: string;
    created_by: UUID;
    last_sync_at?: DateTimeString | null;
    last_success_sync_at?: DateTimeString | null;
    last_error?: string | null;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface CreateConnectionRequest {
    title?: string | null;
    apiBaseUrl: string;
    authMode: string;
    yougileCompanyId: string;
    apiKey: string;
  }

  export interface CreateKeyRequest {
    login: string;
    password: string;
    companyId: string;
  }

  export interface CreateKeyResponse {
    companyId: string;
    apiKey: string;
    warning: string;
  }

  export interface CreateMappingRequest {
    employeeUserId: UUID;
    yougileUserId: string;
  }

  export interface DiscoverCompaniesRequest {
    login: string;
    password: string;
  }

  export interface ImportedUser {
    id: UUID;
    connection_id: UUID;
    yougile_user_id: string;
    email?: string | null;
    real_name?: string | null;
    is_admin: boolean;
    status?: string | null;
    last_activity_at?: DateTimeString | null;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface ImportStructureResponse {
    projectsImported: number;
    boardsImported: number;
    columnsImported: number;
  }

  export interface ImportUsersResponse {
    imported: number;
    updated: number;
    failed: number;
  }

  export interface MappingView {
    id: UUID;
    employee_user_id: UUID;
    employee_name: string;
    employee_email: string;
    yougile_user_id: string;
    yougile_real_name?: string | null;
    yougile_email?: string | null;
    match_source: string;
    is_active: boolean;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface Project {
    id: UUID;
    connection_id: UUID;
    yougile_project_id: string;
    title: string;
    deleted: boolean;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface SyncJob {
    id: UUID;
    connection_id: UUID;
    job_type: string;
    status: string;
    cursor?: Record<string, unknown>;
    progress: Record<string, unknown>;
    started_at?: DateTimeString | null;
    finished_at?: DateTimeString | null;
    attempt: number;
    next_retry_at?: DateTimeString | null;
    error_text?: string | null;
    created_at: DateTimeString;
    updated_at: DateTimeString;
  }

  export interface SyncRequest {
    mode: string;
    includeUsers: boolean;
    includeStructure: boolean;
    includeTasks: boolean;
    taskFilters?: Record<string, any>;
  }

  export interface TestConnectionResponse {
    ok: boolean;
    companyAccessible: boolean;
    rateLimitMode: string;
    message: string;
  }

  export interface TestKeyRequest {
    apiBaseUrl: string;
    apiKey: string;
  }

  export interface UpdateConnectionRequest {
    title?: string | null;
    status?: string | null;
  }
}
```

