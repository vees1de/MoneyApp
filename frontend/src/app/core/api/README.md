# Core API Layer

## Base URL
- `API_BASE_URL = https://bims.su/api`
- All services use backend as source of truth.

## Added Services
- `AuthApiService` (in `core/auth`)
- `AnalyticsApiService`
- `ApprovalWorkflowsApiService`
- `AssignmentsApiService`
- `AuditLogsApiService`
- `BudgetLimitsApiService`
- `CalendarApiService`
- `CertificatesApiService`
- `CoursesApiService`
- `DashboardApiService`
- `DepartmentsApiService` (derived from `auth/me`, no public `/departments` endpoint)
- `EnrollmentsApiService`
- `ExternalRequestsApiService`
- `IntegrationsApiService` (Outlook + Jira summary)
- `LearningPlanApiService`
- `NotificationsApiService`
- `ProgramsApiService`
- `RecommendationsApiService`
- `ReportsApiService`
- `TestsApiService`
- `UsersApiService`
- `RolesApiService`

## Notes
- DTO typing is partial for now. Methods with unstable contract use `Record<string, unknown>`.
- Next step: replace placeholders with strict DTOs from `ENTITY_TS_REFERENCE.md`.