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

