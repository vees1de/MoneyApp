# Unified Frontend API + Types Reference

Единый референс, собранный из:
- FRONTEND_API_GUIDE.md (методы)
- ENTITY_TS_REFERENCE.md (типы)

## 1) Typed API Contract

| Method | Path | Request Type | Response Type | Notes |
|---|---|---|---|---|
| GET | /api/v1/analytics/budget | never | Record<string, unknown> | Бюджетная аналитика (точный DTO не выделен). |
| GET | /api/v1/analytics/compliance | never | Record<string, unknown> | Compliance аналитика (точный DTO не выделен). |
| GET | /api/v1/analytics/dashboard/hr | never | Record<string, unknown> | HR dashboard (точный DTO не выделен в ENTITY_TS_REFERENCE.md). |
| GET | /api/v1/analytics/dashboard/manager | never | Record<string, unknown> | Manager dashboard (точный DTO не выделен). |
| GET | /api/v1/analytics/external-requests | never | Record<string, unknown> | Аналитика внешнего обучения (точный DTO не выделен). |
| GET | /api/v1/analytics/trainers | never | Record<string, unknown> | Аналитика тренеров (точный DTO не выделен). |
| GET | /api/v1/approval-workflows | never | external_training.ApprovalWorkflow[] | Список workflow. |
| GET | /api/v1/assignments | never | learning.Assignment[] | Список назначений. |
| GET | /api/v1/audit-logs | never | audit.LogEntry[] | Журнал аудита. |
| GET | /api/v1/auth/me | never | identity.MeResponse | Основной bootstrap endpoint для ролей и permissions. |
| GET | /api/v1/budget-limits | never | external_training.BudgetLimit[] | Список лимитов бюджета. |
| GET | /api/v1/certificates/my | never | certificates.Certificate[] | Мои сертификаты. |
| GET | /api/v1/courses | never | catalog.Course[] | Список курсов (с фильтрами через query params). |
| GET | /api/v1/courses/{id} | never | catalog.Course | Детали курса. |
| GET | /api/v1/courses/{id}/materials | never | catalog.CourseMaterial[] | Материалы курса. |
| GET | /api/v1/enrollments/{id} | never | learning.Enrollment | Детали enrollment. |
| GET | /api/v1/enrollments/my | never | learning.Enrollment[] | Обучения текущего пользователя. |
| GET | /api/v1/external-requests/{id} | never | external_training.ExternalRequest | Детали заявки. |
| GET | /api/v1/external-requests/my | never | external_training.ExternalRequest[] | Мои заявки. |
| GET | /api/v1/integrations/outlook/connect | never | outlook.ConnectResponse | Ссылка и state для OAuth подключения. |
| GET | /api/v1/integrations/outlook/status | never | outlook.IntegrationStatus | Текущий статус интеграции. |
| GET | /api/v1/notifications | never | notifications.Notification[] | Список уведомлений. |
| GET | /api/v1/programs | never | university.Program[] | Список программ. |
| GET | /api/v1/programs/{id} | never | university.Program | Детали программы. |
| GET | /api/v1/reports/export/excel | never | Blob | Скачивание Excel отчёта. |
| GET | /api/v1/reports/export/pdf | never | Blob | Скачивание PDF отчёта. |
| GET | /api/v1/tests/{id} | never | testing.Test | Тест с вопросами/опциями. |
| GET | /api/v1/tests/{id}/results | never | testing.TestResult[] | История/результаты по тесту. |
| PATCH | /api/v1/courses/{id} | catalog.UpdateCourseRequest | catalog.Course | Частичное обновление курса. |
| PATCH | /api/v1/external-requests/{id} | external_training.UpdateExternalRequestRequest | external_training.ExternalRequest | Обновление заявки. |
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
| POST | /api/v1/integrations/outlook/disconnect | never | outlook.IntegrationStatus | Отключение интеграции. |
| POST | /api/v1/integrations/outlook/sync | never | outlook.IntegrationStatus | Ручная синхронизация. |
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
  'GET /api/v1/certificates/my': { request: never; response: certificates.Certificate[] };
  'GET /api/v1/courses': { request: never; response: catalog.Course[] };
  'GET /api/v1/courses/{id}': { request: never; response: catalog.Course };
  'GET /api/v1/courses/{id}/materials': { request: never; response: catalog.CourseMaterial[] };
  'GET /api/v1/enrollments/{id}': { request: never; response: learning.Enrollment };
  'GET /api/v1/enrollments/my': { request: never; response: learning.Enrollment[] };
  'GET /api/v1/external-requests/{id}': { request: never; response: external_training.ExternalRequest };
  'GET /api/v1/external-requests/my': { request: never; response: external_training.ExternalRequest[] };
  'GET /api/v1/integrations/outlook/connect': { request: never; response: outlook.ConnectResponse };
  'GET /api/v1/integrations/outlook/status': { request: never; response: outlook.IntegrationStatus };
  'GET /api/v1/notifications': { request: never; response: notifications.Notification[] };
  'GET /api/v1/programs': { request: never; response: university.Program[] };
  'GET /api/v1/programs/{id}': { request: never; response: university.Program };
  'GET /api/v1/reports/export/excel': { request: never; response: Blob };
  'GET /api/v1/reports/export/pdf': { request: never; response: Blob };
  'GET /api/v1/tests/{id}': { request: never; response: testing.Test };
  'GET /api/v1/tests/{id}/results': { request: never; response: testing.TestResult[] };
  'PATCH /api/v1/courses/{id}': { request: catalog.UpdateCourseRequest; response: catalog.Course };
  'PATCH /api/v1/external-requests/{id}': { request: external_training.UpdateExternalRequestRequest; response: external_training.ExternalRequest };
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
  'POST /api/v1/integrations/outlook/disconnect': { request: never; response: outlook.IntegrationStatus };
  'POST /api/v1/integrations/outlook/sync': { request: never; response: outlook.IntegrationStatus };
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

