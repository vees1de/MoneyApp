package external_training

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"moneyapp/backend/internal/modules/identity"
	"moneyapp/backend/internal/modules/org"
	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/clock"
	"moneyapp/backend/internal/platform/db"
	"moneyapp/backend/internal/platform/events"
	"moneyapp/backend/internal/platform/httpx"
	"moneyapp/backend/internal/platform/notificationsx"
	"moneyapp/backend/internal/platform/outbox"
	"moneyapp/backend/internal/platform/worker"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ExternalRequest struct {
	ID                     uuid.UUID  `json:"id"`
	RequestNo              string     `json:"request_no"`
	EmployeeUserID         uuid.UUID  `json:"employee_user_id"`
	EmployeeFullName       string     `json:"employee_full_name,omitempty"`
	EmployeeEmail          string     `json:"employee_email,omitempty"`
	DepartmentID           *uuid.UUID `json:"department_id,omitempty"`
	DepartmentName         string     `json:"department_name,omitempty"`
	Title                  string     `json:"title"`
	ProviderID             *uuid.UUID `json:"provider_id,omitempty"`
	ProviderName           *string    `json:"provider_name,omitempty"`
	CourseURL              *string    `json:"course_url,omitempty"`
	ProgramDescription     *string    `json:"program_description,omitempty"`
	PlannedStartDate       *time.Time `json:"planned_start_date,omitempty"`
	PlannedEndDate         *time.Time `json:"planned_end_date,omitempty"`
	DurationHours          *string    `json:"duration_hours,omitempty"`
	CostAmount             string     `json:"cost_amount"`
	Currency               string     `json:"currency"`
	BusinessGoal           *string    `json:"business_goal,omitempty"`
	EmployeeComment        *string    `json:"employee_comment,omitempty"`
	ManagerComment         *string    `json:"manager_comment,omitempty"`
	HRComment              *string    `json:"hr_comment,omitempty"`
	Status                 string     `json:"status"`
	CalendarConflictStatus *string    `json:"calendar_conflict_status,omitempty"`
	BudgetCheckStatus      *string    `json:"budget_check_status,omitempty"`
	CurrentApprovalStepID  *uuid.UUID `json:"current_approval_step_id,omitempty"`
	CurrentApprovalStatus  string     `json:"current_approval_status,omitempty"`
	CurrentApprovalRole    string     `json:"current_approval_role_code,omitempty"`
	CurrentApprovalDueAt   *time.Time `json:"current_approval_due_at,omitempty"`
	CurrentApproverUserID  *uuid.UUID `json:"current_approver_user_id,omitempty"`
	CurrentApproverName    string     `json:"current_approver_full_name,omitempty"`
	ApprovedAt             *time.Time `json:"approved_at,omitempty"`
	RejectedAt             *time.Time `json:"rejected_at,omitempty"`
	SentToRevisionAt       *time.Time `json:"sent_to_revision_at,omitempty"`
	TrainingStartedAt      *time.Time `json:"training_started_at,omitempty"`
	TrainingCompletedAt    *time.Time `json:"training_completed_at,omitempty"`
	CertificateUploadedAt  *time.Time `json:"certificate_uploaded_at,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

type BudgetLimit struct {
	ID          uuid.UUID  `json:"id"`
	ScopeType   string     `json:"scope_type"`
	ScopeID     *uuid.UUID `json:"scope_id,omitempty"`
	PeriodYear  int        `json:"period_year"`
	PeriodMonth *int       `json:"period_month,omitempty"`
	LimitAmount string     `json:"limit_amount"`
	Currency    string     `json:"currency"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type ApprovalWorkflow struct {
	ID         uuid.UUID              `json:"id"`
	EntityType string                 `json:"entity_type"`
	Name       string                 `json:"name"`
	IsActive   bool                   `json:"is_active"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Steps      []ApprovalWorkflowStep `json:"steps,omitempty"`
}

type ApprovalWorkflowStep struct {
	ID             uuid.UUID  `json:"id"`
	WorkflowID     uuid.UUID  `json:"workflow_id"`
	StepOrder      int        `json:"step_order"`
	RoleCode       string     `json:"role_code"`
	ApproverSource string     `json:"approver_source"`
	ApproverUserID *uuid.UUID `json:"approver_user_id,omitempty"`
	SLAHours       *int       `json:"sla_hours,omitempty"`
	IsRequired     bool       `json:"is_required"`
}

type ApprovalStep struct {
	ID             uuid.UUID  `json:"id"`
	EntityType     string     `json:"entity_type"`
	EntityID       uuid.UUID  `json:"entity_id"`
	StepOrder      int        `json:"step_order"`
	ApproverUserID uuid.UUID  `json:"approver_user_id"`
	RoleCode       string     `json:"role_code"`
	Status         string     `json:"status"`
	Comment        *string    `json:"comment,omitempty"`
	DueAt          *time.Time `json:"due_at,omitempty"`
	ActedAt        *time.Time `json:"acted_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type CreateExternalRequestRequest struct {
	Title              string     `json:"title" validate:"required"`
	ProviderID         *uuid.UUID `json:"provider_id,omitempty"`
	ProviderName       *string    `json:"provider_name,omitempty"`
	CourseURL          *string    `json:"course_url,omitempty"`
	ProgramDescription *string    `json:"program_description,omitempty"`
	PlannedStartDate   *time.Time `json:"planned_start_date,omitempty"`
	PlannedEndDate     *time.Time `json:"planned_end_date,omitempty"`
	DurationHours      *string    `json:"duration_hours,omitempty"`
	CostAmount         string     `json:"cost_amount" validate:"required"`
	Currency           string     `json:"currency" validate:"required"`
	BusinessGoal       *string    `json:"business_goal,omitempty"`
	EmployeeComment    *string    `json:"employee_comment,omitempty"`
}

type UpdateExternalRequestRequest struct {
	Title              *string    `json:"title,omitempty"`
	ProviderID         *uuid.UUID `json:"provider_id,omitempty"`
	ProviderName       *string    `json:"provider_name,omitempty"`
	CourseURL          *string    `json:"course_url,omitempty"`
	ProgramDescription *string    `json:"program_description,omitempty"`
	PlannedStartDate   *time.Time `json:"planned_start_date,omitempty"`
	PlannedEndDate     *time.Time `json:"planned_end_date,omitempty"`
	DurationHours      *string    `json:"duration_hours,omitempty"`
	CostAmount         *string    `json:"cost_amount,omitempty"`
	Currency           *string    `json:"currency,omitempty"`
	BusinessGoal       *string    `json:"business_goal,omitempty"`
	EmployeeComment    *string    `json:"employee_comment,omitempty"`
	ManagerComment     *string    `json:"manager_comment,omitempty"`
	HRComment          *string    `json:"hr_comment,omitempty"`
}

type ActionCommentRequest struct {
	Comment *string `json:"comment,omitempty"`
}

type UploadRequestCertificateRequest struct {
	StorageProvider string `json:"storage_provider" validate:"required,oneof=s3 local minio"`
	StorageKey      string `json:"storage_key" validate:"required"`
	OriginalName    string `json:"original_name" validate:"required"`
	MimeType        string `json:"mime_type" validate:"required"`
	SizeBytes       int64  `json:"size_bytes" validate:"required,min=1"`
}

type CreateBudgetLimitRequest struct {
	ScopeType   string     `json:"scope_type" validate:"required,oneof=company department employee"`
	ScopeID     *uuid.UUID `json:"scope_id,omitempty"`
	PeriodYear  int        `json:"period_year" validate:"required"`
	PeriodMonth *int       `json:"period_month,omitempty"`
	LimitAmount string     `json:"limit_amount" validate:"required"`
	Currency    string     `json:"currency" validate:"required"`
	IsActive    bool       `json:"is_active"`
}

type CreateWorkflowStepRequest struct {
	StepOrder      int        `json:"step_order" validate:"required"`
	RoleCode       string     `json:"role_code" validate:"required"`
	ApproverSource string     `json:"approver_source" validate:"required,oneof=line_manager specific_role department_head static_user"`
	ApproverUserID *uuid.UUID `json:"approver_user_id,omitempty"`
	SLAHours       *int       `json:"sla_hours,omitempty"`
	IsRequired     bool       `json:"is_required"`
}

type CreateWorkflowRequest struct {
	EntityType string                      `json:"entity_type" validate:"required"`
	Name       string                      `json:"name" validate:"required"`
	IsActive   bool                        `json:"is_active"`
	Steps      []CreateWorkflowStepRequest `json:"steps" validate:"required,min=1,dive"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(database *sql.DB) *Repository {
	return &Repository{db: database}
}

func (r *Repository) base(exec ...db.DBTX) db.DBTX {
	if len(exec) > 0 && exec[0] != nil {
		return exec[0]
	}
	return r.db
}

func (r *Repository) CreateRequest(ctx context.Context, item ExternalRequest, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into external_course_requests (
			id, request_no, employee_user_id, department_id, title, provider_id, provider_name, course_url,
			program_description, planned_start_date, planned_end_date, duration_hours, cost_amount, currency,
			business_goal, employee_comment, manager_comment, hr_comment, status, calendar_conflict_status,
			budget_check_status, current_approval_step_id, approved_at, rejected_at, sent_to_revision_at,
			training_started_at, training_completed_at, certificate_uploaded_at, created_at, updated_at
		)
		values (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10::date, $11::date,
			nullif($12, '')::numeric, nullif($13, '')::numeric, $14, $15, $16, $17, $18, $19, $20, $21,
			$22, $23, $24, $25, $26, $27, $28, $29, $30
		)
	`, item.ID, item.RequestNo, item.EmployeeUserID, item.DepartmentID, item.Title, item.ProviderID, item.ProviderName, item.CourseURL,
		item.ProgramDescription, item.PlannedStartDate, item.PlannedEndDate, item.DurationHours, item.CostAmount, item.Currency,
		item.BusinessGoal, item.EmployeeComment, item.ManagerComment, item.HRComment, item.Status, item.CalendarConflictStatus,
		item.BudgetCheckStatus, item.CurrentApprovalStepID, item.ApprovedAt, item.RejectedAt, item.SentToRevisionAt,
		item.TrainingStartedAt, item.TrainingCompletedAt, item.CertificateUploadedAt, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *Repository) UpdateRequest(ctx context.Context, item ExternalRequest, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update external_course_requests
		set title = $2,
		    provider_id = $3,
		    provider_name = $4,
		    course_url = $5,
		    program_description = $6,
		    planned_start_date = $7::date,
		    planned_end_date = $8::date,
		    duration_hours = nullif($9, '')::numeric,
		    cost_amount = nullif($10, '')::numeric,
		    currency = $11,
		    business_goal = $12,
		    employee_comment = $13,
		    manager_comment = $14,
		    hr_comment = $15,
		    status = $16,
		    calendar_conflict_status = $17,
		    budget_check_status = $18,
		    current_approval_step_id = $19,
		    approved_at = $20,
		    rejected_at = $21,
		    sent_to_revision_at = $22,
		    training_started_at = $23,
		    training_completed_at = $24,
		    certificate_uploaded_at = $25,
		    updated_at = $26
		where id = $1
	`, item.ID, item.Title, item.ProviderID, item.ProviderName, item.CourseURL, item.ProgramDescription,
		item.PlannedStartDate, item.PlannedEndDate, item.DurationHours, item.CostAmount, item.Currency, item.BusinessGoal,
		item.EmployeeComment, item.ManagerComment, item.HRComment, item.Status, item.CalendarConflictStatus, item.BudgetCheckStatus,
		item.CurrentApprovalStepID, item.ApprovedAt, item.RejectedAt, item.SentToRevisionAt, item.TrainingStartedAt,
		item.TrainingCompletedAt, item.CertificateUploadedAt, item.UpdatedAt)
	return err
}

func (r *Repository) GetRequest(ctx context.Context, id uuid.UUID, exec ...db.DBTX) (ExternalRequest, error) {
	row := r.base(exec...).QueryRowContext(ctx, externalRequestReadColumns+externalRequestReadJoins+`
		where r.id = $1
	`, id)
	return scanExternalRequest(row)
}

func (r *Repository) ListRequestsByUser(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) ([]ExternalRequest, error) {
	rows, err := r.base(exec...).QueryContext(ctx, externalRequestReadColumns+externalRequestReadJoins+`
		where r.employee_user_id = $1
		order by r.created_at desc
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ExternalRequest
	for rows.Next() {
		item, err := scanExternalRequest(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateBudgetLimit(ctx context.Context, item BudgetLimit, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into budget_limits (id, scope_type, scope_id, period_year, period_month, limit_amount, currency, is_active, created_at, updated_at)
		values ($1, $2, $3, $4, $5, nullif($6, '')::numeric, $7, $8, $9, $10)
	`, item.ID, item.ScopeType, item.ScopeID, item.PeriodYear, item.PeriodMonth, item.LimitAmount, item.Currency, item.IsActive, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *Repository) ListBudgetLimits(ctx context.Context, exec ...db.DBTX) ([]BudgetLimit, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select id, scope_type, scope_id, period_year, period_month, limit_amount::text, currency, is_active, created_at, updated_at
		from budget_limits
		order by created_at desc
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []BudgetLimit
	for rows.Next() {
		var item BudgetLimit
		if err := rows.Scan(&item.ID, &item.ScopeType, &item.ScopeID, &item.PeriodYear, &item.PeriodMonth, &item.LimitAmount, &item.Currency, &item.IsActive, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) FindBudgetLimit(ctx context.Context, departmentID *uuid.UUID, year int, exec ...db.DBTX) (*BudgetLimit, error) {
	query := `
		select id, scope_type, scope_id, period_year, period_month, limit_amount::text, currency, is_active, created_at, updated_at
		from budget_limits
		where is_active = true and period_year = $1
		  and (
		    (scope_type = 'department' and scope_id = $2)
		    or scope_type = 'company'
		  )
		order by case when scope_type = 'department' then 0 else 1 end asc, created_at asc
		limit 1
	`
	var item BudgetLimit
	err := r.db.QueryRowContext(ctx, query, year, departmentID).Scan(&item.ID, &item.ScopeType, &item.ScopeID, &item.PeriodYear, &item.PeriodMonth, &item.LimitAmount, &item.Currency, &item.IsActive, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *Repository) ReserveBudget(ctx context.Context, limitID, requestID uuid.UUID, amount string, createdAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into budget_consumptions (
			id, budget_limit_id, request_id, reserved_amount, actual_amount, status, created_at, updated_at
		)
		values ($1, $2, $3, nullif($4, '')::numeric, 0, 'reserved', $5, $5)
		on conflict do nothing
	`, uuid.New(), limitID, requestID, amount, createdAt)
	return err
}

func (r *Repository) CreateWorkflow(ctx context.Context, item ApprovalWorkflow, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into approval_workflows (id, entity_type, name, is_active, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6)
	`, item.ID, item.EntityType, item.Name, item.IsActive, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *Repository) CreateWorkflowStep(ctx context.Context, item ApprovalWorkflowStep, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into approval_workflow_steps (
			id, workflow_id, step_order, role_code, approver_source, approver_user_id, sla_hours, is_required
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8)
	`, item.ID, item.WorkflowID, item.StepOrder, item.RoleCode, item.ApproverSource, item.ApproverUserID, item.SLAHours, item.IsRequired)
	return err
}

func (r *Repository) ListWorkflows(ctx context.Context, exec ...db.DBTX) ([]ApprovalWorkflow, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select id, entity_type, name, is_active, created_at, updated_at
		from approval_workflows
		order by created_at desc
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ApprovalWorkflow
	for rows.Next() {
		var item ApprovalWorkflow
		if err := rows.Scan(&item.ID, &item.EntityType, &item.Name, &item.IsActive, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ListWorkflowSteps(ctx context.Context, workflowID uuid.UUID, exec ...db.DBTX) ([]ApprovalWorkflowStep, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select id, workflow_id, step_order, role_code, approver_source, approver_user_id, sla_hours, is_required
		from approval_workflow_steps
		where workflow_id = $1
		order by step_order asc
	`, workflowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ApprovalWorkflowStep
	for rows.Next() {
		var item ApprovalWorkflowStep
		if err := rows.Scan(&item.ID, &item.WorkflowID, &item.StepOrder, &item.RoleCode, &item.ApproverSource, &item.ApproverUserID, &item.SLAHours, &item.IsRequired); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) GetActiveWorkflow(ctx context.Context, entityType string, exec ...db.DBTX) (*ApprovalWorkflow, error) {
	var item ApprovalWorkflow
	err := r.base(exec...).QueryRowContext(ctx, `
		select id, entity_type, name, is_active, created_at, updated_at
		from approval_workflows
		where entity_type = $1 and is_active = true
		order by created_at desc
		limit 1
	`, entityType).Scan(&item.ID, &item.EntityType, &item.Name, &item.IsActive, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *Repository) CreateApprovalStep(ctx context.Context, item ApprovalStep, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into approval_steps (
			id, entity_type, entity_id, step_order, approver_user_id, role_code, status, comment, due_at, acted_at, created_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, item.ID, item.EntityType, item.EntityID, item.StepOrder, item.ApproverUserID, item.RoleCode, item.Status, item.Comment, item.DueAt, item.ActedAt, item.CreatedAt)
	return err
}

func (r *Repository) UpdateApprovalStep(ctx context.Context, item ApprovalStep, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update approval_steps
		set status = $2, comment = $3, due_at = $4, acted_at = $5
		where id = $1
	`, item.ID, item.Status, item.Comment, item.DueAt, item.ActedAt)
	return err
}

func (r *Repository) GetApprovalStep(ctx context.Context, id uuid.UUID, exec ...db.DBTX) (ApprovalStep, error) {
	var item ApprovalStep
	err := r.base(exec...).QueryRowContext(ctx, `
		select id, entity_type, entity_id, step_order, approver_user_id, role_code, status, comment, due_at, acted_at, created_at
		from approval_steps
		where id = $1
	`, id).Scan(&item.ID, &item.EntityType, &item.EntityID, &item.StepOrder, &item.ApproverUserID, &item.RoleCode, &item.Status, &item.Comment, &item.DueAt, &item.ActedAt, &item.CreatedAt)
	return item, err
}

func (r *Repository) FindNextPendingApprovalStep(ctx context.Context, entityID uuid.UUID, afterOrder int, exec ...db.DBTX) (*ApprovalStep, error) {
	var item ApprovalStep
	err := r.base(exec...).QueryRowContext(ctx, `
		select id, entity_type, entity_id, step_order, approver_user_id, role_code, status, comment, due_at, acted_at, created_at
		from approval_steps
		where entity_type = 'external_course_request' and entity_id = $1 and status = 'pending' and step_order > $2
		order by step_order asc
		limit 1
	`, entityID, afterOrder).Scan(&item.ID, &item.EntityType, &item.EntityID, &item.StepOrder, &item.ApproverUserID, &item.RoleCode, &item.Status, &item.Comment, &item.DueAt, &item.ActedAt, &item.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *Repository) CreateApprovalHistory(ctx context.Context, entityID uuid.UUID, stepID *uuid.UUID, action string, fromStatus, toStatus *string, performedBy uuid.UUID, comment *string, createdAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into approval_history (
			id, entity_type, entity_id, step_id, action, from_status, to_status, performed_by, comment, created_at
		)
		values ($1, 'external_course_request', $2, $3, $4, $5, $6, $7, $8, $9)
	`, uuid.New(), entityID, stepID, action, fromStatus, toStatus, performedBy, comment, createdAt)
	return err
}

func (r *Repository) CreateNotification(ctx context.Context, userID uuid.UUID, typ, title, body string, relatedEntityID uuid.UUID, createdAt time.Time, exec ...db.DBTX) error {
	return notificationsx.CreateInAppWithLinkedEmailMirror(
		ctx,
		r.base(exec...),
		userID,
		typ,
		title,
		body,
		"external_course_request",
		relatedEntityID,
		createdAt,
	)
}

func (r *Repository) ResolveDepartmentHead(ctx context.Context, departmentID uuid.UUID, exec ...db.DBTX) (*uuid.UUID, error) {
	var userID uuid.UUID
	err := r.base(exec...).QueryRowContext(ctx, `select head_user_id from departments where id = $1`, departmentID).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &userID, nil
}

func (r *Repository) IsPrimaryManagerOf(ctx context.Context, managerUserID, employeeUserID uuid.UUID, exec ...db.DBTX) (bool, error) {
	var exists bool
	err := r.base(exec...).QueryRowContext(ctx, `
		select exists (
			select 1
			from manager_relations
			where manager_user_id = $1
			  and employee_user_id = $2
			  and is_primary = true
		)
	`, managerUserID, employeeUserID).Scan(&exists)
	return exists, err
}

func (r *Repository) CreateCalendarEvent(ctx context.Context, userID, sourceID uuid.UUID, title string, startAt, endAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into calendar_events (
			id, user_id, source_type, source_id, provider, external_event_id, title, start_at, end_at,
			timezone, status, meeting_url, location, payload, created_at, updated_at
		)
		values ($1, $2, 'external_request', $3, 'system', null, $4, $5, $6, 'UTC', 'scheduled', null, null, '{}'::jsonb, $7, $7)
	`, uuid.New(), userID, sourceID, title, startAt, endAt, time.Now().UTC())
	return err
}

func (r *Repository) CreateIntegrationJob(ctx context.Context, entityID uuid.UUID, payload string, createdAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into integration_jobs (
			id, integration_type, entity_type, entity_id, job_type, status, attempt, max_attempts, next_retry_at, last_error, payload, created_at, updated_at
		)
		values ($1, 'outlook_calendar_sync', 'external_course_request', $2, 'create_event', 'pending', 0, 5, null, null, $3::jsonb, $4, $4)
	`, uuid.New(), entityID, payload, createdAt)
	return err
}

func (r *Repository) CreateFileAttachment(ctx context.Context, id, uploadedBy uuid.UUID, provider, key, originalName, mimeType string, sizeBytes int64, createdAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into file_attachments (id, storage_provider, storage_key, original_name, mime_type, size_bytes, uploaded_by, created_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8)
	`, id, provider, key, originalName, mimeType, sizeBytes, uploadedBy, createdAt)
	return err
}

func (r *Repository) CreateAttachedDocument(ctx context.Context, requestID, fileID, uploadedBy uuid.UUID, documentType string, createdAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into attached_documents (id, request_id, file_id, document_type, uploaded_by, created_at)
		values ($1, $2, $3, $4, $5, $6)
	`, uuid.New(), requestID, fileID, documentType, uploadedBy, createdAt)
	return err
}

type Service struct {
	db           *sql.DB
	repo         *Repository
	identityRepo *identity.Repository
	orgService   *org.Service
	outbox       *outbox.Service
	queue        *worker.Queue
	clock        clock.Clock
}

func NewService(database *sql.DB, repo *Repository, identityRepo *identity.Repository, orgService *org.Service, outboxService *outbox.Service, queue *worker.Queue, appClock clock.Clock) *Service {
	return &Service{
		db:           database,
		repo:         repo,
		identityRepo: identityRepo,
		orgService:   orgService,
		outbox:       outboxService,
		queue:        queue,
		clock:        appClock,
	}
}

func hasRole(principal platformauth.Principal, roleCode string) bool {
	for _, role := range principal.RoleCodes {
		if role == roleCode {
			return true
		}
	}
	return false
}

func canCreateOwnExternalRequest(principal platformauth.Principal) bool {
	return principal.HasPermission("external_requests.create") ||
		hasRole(principal, "manager") ||
		hasRole(principal, "hr") ||
		hasRole(principal, "trainer") ||
		hasRole(principal, "admin")
}

func canReadOwnExternalRequests(principal platformauth.Principal) bool {
	return principal.HasPermission("external_requests.read_own") ||
		canCreateOwnExternalRequest(principal) ||
		canViewTeamExternalRequests(principal) ||
		canViewAllExternalRequests(principal)
}

func canViewTeamExternalRequests(principal platformauth.Principal) bool {
	return hasRole(principal, "manager") || hasRole(principal, "admin")
}

func canViewAllExternalRequests(principal platformauth.Principal) bool {
	return hasRole(principal, "hr") || hasRole(principal, "admin")
}

func validateExternalRequestSchedule(item ExternalRequest) error {
	if item.PlannedStartDate != nil && item.PlannedEndDate != nil && item.PlannedEndDate.Before(*item.PlannedStartDate) {
		return httpx.BadRequest("invalid_schedule", "planned_end_date must be on or after planned_start_date")
	}
	return nil
}

func submissionHistoryAction(status string) string {
	if status == "needs_revision" {
		return "resubmitted"
	}
	return "submitted"
}

func (s *Service) canViewRequest(ctx context.Context, principal platformauth.Principal, item ExternalRequest) (bool, error) {
	if item.EmployeeUserID == principal.UserID {
		return true, nil
	}
	if canViewAllExternalRequests(principal) {
		return true, nil
	}
	if item.CurrentApproverUserID != nil && *item.CurrentApproverUserID == principal.UserID {
		return true, nil
	}
	if canViewTeamExternalRequests(principal) {
		return s.repo.IsPrimaryManagerOf(ctx, principal.UserID, item.EmployeeUserID)
	}
	return false, nil
}

func (s *Service) CreateRequest(ctx context.Context, principal platformauth.Principal, req CreateExternalRequestRequest) (ExternalRequest, error) {
	if !canCreateOwnExternalRequest(principal) {
		return ExternalRequest{}, httpx.Forbidden("forbidden", "permission denied")
	}

	now := s.clock.Now()
	item := ExternalRequest{
		ID:                 uuid.New(),
		RequestNo:          "EXT-" + now.Format("20060102") + "-" + uuid.NewString()[:8],
		EmployeeUserID:     principal.UserID,
		DepartmentID:       principal.DepartmentID,
		Title:              req.Title,
		ProviderID:         req.ProviderID,
		ProviderName:       req.ProviderName,
		CourseURL:          req.CourseURL,
		ProgramDescription: req.ProgramDescription,
		PlannedStartDate:   req.PlannedStartDate,
		PlannedEndDate:     req.PlannedEndDate,
		DurationHours:      req.DurationHours,
		CostAmount:         req.CostAmount,
		Currency:           req.Currency,
		BusinessGoal:       req.BusinessGoal,
		EmployeeComment:    req.EmployeeComment,
		Status:             "draft",
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := validateExternalRequestSchedule(item); err != nil {
		return ExternalRequest{}, err
	}
	if err := s.repo.CreateRequest(ctx, item); err != nil {
		return ExternalRequest{}, err
	}
	return s.GetRequest(ctx, principal, item.ID)
}

func (s *Service) ListMine(ctx context.Context, principal platformauth.Principal) ([]ExternalRequest, error) {
	if !canReadOwnExternalRequests(principal) {
		return nil, httpx.Forbidden("forbidden", "permission denied")
	}
	return s.repo.ListRequestsByUser(ctx, principal.UserID)
}

func (s *Service) GetRequest(ctx context.Context, principal platformauth.Principal, id uuid.UUID) (ExternalRequest, error) {
	item, err := s.repo.GetRequest(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ExternalRequest{}, httpx.NotFound("external_request_not_found", "external request not found")
		}
		return ExternalRequest{}, err
	}
	allowed, err := s.canViewRequest(ctx, principal, item)
	if err != nil {
		return ExternalRequest{}, err
	}
	if !allowed {
		return ExternalRequest{}, httpx.Forbidden("forbidden", "permission denied")
	}
	return item, nil
}

func (s *Service) UpdateRequest(ctx context.Context, principal platformauth.Principal, id uuid.UUID, req UpdateExternalRequestRequest) (ExternalRequest, error) {
	item, err := s.GetRequest(ctx, principal, id)
	if err != nil {
		return ExternalRequest{}, err
	}
	if item.EmployeeUserID != principal.UserID && !hasRole(principal, "admin") {
		return ExternalRequest{}, httpx.Forbidden("forbidden", "only owner can edit request")
	}
	if item.Status != "draft" && item.Status != "needs_revision" {
		return ExternalRequest{}, httpx.Conflict("request_locked", "request cannot be edited in current state")
	}
	if req.ManagerComment != nil || req.HRComment != nil {
		return ExternalRequest{}, httpx.BadRequest("readonly_fields", "manager_comment and hr_comment can only be changed in approval actions")
	}
	if req.Title != nil {
		item.Title = *req.Title
	}
	if req.ProviderID != nil {
		item.ProviderID = req.ProviderID
	}
	if req.ProviderName != nil {
		item.ProviderName = req.ProviderName
	}
	if req.CourseURL != nil {
		item.CourseURL = req.CourseURL
	}
	if req.ProgramDescription != nil {
		item.ProgramDescription = req.ProgramDescription
	}
	if req.PlannedStartDate != nil {
		item.PlannedStartDate = req.PlannedStartDate
	}
	if req.PlannedEndDate != nil {
		item.PlannedEndDate = req.PlannedEndDate
	}
	if req.DurationHours != nil {
		item.DurationHours = req.DurationHours
	}
	if req.CostAmount != nil {
		item.CostAmount = *req.CostAmount
	}
	if req.Currency != nil {
		item.Currency = *req.Currency
	}
	if req.BusinessGoal != nil {
		item.BusinessGoal = req.BusinessGoal
	}
	if req.EmployeeComment != nil {
		item.EmployeeComment = req.EmployeeComment
	}
	if err := validateExternalRequestSchedule(item); err != nil {
		return ExternalRequest{}, err
	}
	item.UpdatedAt = s.clock.Now()
	if err := s.repo.UpdateRequest(ctx, item); err != nil {
		return ExternalRequest{}, err
	}
	return s.GetRequest(ctx, principal, item.ID)
}

func (s *Service) Submit(ctx context.Context, principal platformauth.Principal, id uuid.UUID) (ExternalRequest, error) {
	item, err := s.GetRequest(ctx, principal, id)
	if err != nil {
		return ExternalRequest{}, err
	}
	if item.EmployeeUserID != principal.UserID {
		return ExternalRequest{}, httpx.Forbidden("forbidden", "only owner can submit request")
	}
	if item.Status != "draft" && item.Status != "needs_revision" {
		return ExternalRequest{}, httpx.Conflict("invalid_status", "request cannot be submitted from current state")
	}
	if err := validateExternalRequestSchedule(item); err != nil {
		return ExternalRequest{}, err
	}

	fromStatus := item.Status
	if err := db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		steps, err := s.buildApprovalSteps(ctx, item, tx)
		if err != nil {
			return err
		}
		if len(steps) == 0 {
			return httpx.Conflict("approval_workflow_missing", "approval workflow could not be resolved")
		}
		for _, step := range steps {
			if err := s.repo.CreateApprovalStep(ctx, step, tx); err != nil {
				return err
			}
		}

		now := s.clock.Now()
		item.Status = approvalStatusFromRole(steps[0].RoleCode)
		item.CurrentApprovalStepID = &steps[0].ID
		notChecked := "not_checked"
		item.CalendarConflictStatus = &notChecked
		item.BudgetCheckStatus = &notChecked
		item.UpdatedAt = now
		if err := s.repo.UpdateRequest(ctx, item, tx); err != nil {
			return err
		}
		toStatus := item.Status
		if err := s.repo.CreateApprovalHistory(ctx, item.ID, nil, submissionHistoryAction(fromStatus), &fromStatus, &toStatus, principal.UserID, nil, now, tx); err != nil {
			return err
		}
		if err := s.repo.CreateNotification(ctx, steps[0].ApproverUserID, "approval_required", "Approval required", "External training request requires your approval", item.ID, now, tx); err != nil {
			return err
		}
		return s.outbox.Publish(ctx, tx, events.Message{
			Topic:      "external_training",
			EventType:  "external_request.submitted",
			EntityType: "external_course_request",
			EntityID:   item.ID,
			Payload: map[string]any{
				"employee_user_id": item.EmployeeUserID,
				"status":           item.Status,
			},
			OccurredAt: now,
		})
	}); err != nil {
		return ExternalRequest{}, err
	}
	return s.GetRequest(ctx, principal, item.ID)
}

func (s *Service) Approve(ctx context.Context, principal platformauth.Principal, id uuid.UUID, comment *string) (ExternalRequest, error) {
	item, err := s.GetRequest(ctx, principal, id)
	if err != nil {
		return ExternalRequest{}, err
	}
	if item.CurrentApprovalStepID == nil {
		return ExternalRequest{}, httpx.Conflict("approval_missing", "request has no current approval step")
	}

	if err := db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		step, err := s.repo.GetApprovalStep(ctx, *item.CurrentApprovalStepID, tx)
		if err != nil {
			return err
		}
		if step.ApproverUserID != principal.UserID && !principal.HasPermission("settings.manage") {
			return httpx.Forbidden("forbidden", "only current approver can act")
		}

		now := s.clock.Now()
		fromStatus := item.Status
		step.Status = "approved"
		step.Comment = comment
		step.ActedAt = &now
		if err := s.repo.UpdateApprovalStep(ctx, step, tx); err != nil {
			return err
		}

		nextStep, err := s.repo.FindNextPendingApprovalStep(ctx, item.ID, step.StepOrder, tx)
		if err != nil {
			return err
		}

		if nextStep != nil {
			item.CurrentApprovalStepID = &nextStep.ID
			item.Status = approvalStatusFromRole(nextStep.RoleCode)
			item.UpdatedAt = now
			if err := s.repo.UpdateRequest(ctx, item, tx); err != nil {
				return err
			}
			if err := s.repo.CreateApprovalHistory(ctx, item.ID, &step.ID, "approved", &fromStatus, &item.Status, principal.UserID, comment, now, tx); err != nil {
				return err
			}
			return s.repo.CreateNotification(ctx, nextStep.ApproverUserID, "approval_required", "Approval required", "External training request requires your approval", item.ID, now, tx)
		}

		item.Status = "approved"
		item.ApprovedAt = &now
		item.CurrentApprovalStepID = nil
		ok := "ok"
		item.BudgetCheckStatus = &ok
		item.CalendarConflictStatus = &ok
		item.UpdatedAt = now
		if err := s.repo.UpdateRequest(ctx, item, tx); err != nil {
			return err
		}

		if limit, err := s.repo.FindBudgetLimit(ctx, item.DepartmentID, now.Year(), tx); err != nil {
			return err
		} else if limit != nil {
			if err := s.repo.ReserveBudget(ctx, limit.ID, item.ID, item.CostAmount, now, tx); err != nil {
				return err
			}
		}

		startAt, endAt := requestSchedule(item, now)
		if err := s.repo.CreateCalendarEvent(ctx, item.EmployeeUserID, item.ID, item.Title, startAt, endAt, tx); err != nil {
			return err
		}
		if err := s.repo.CreateIntegrationJob(ctx, item.ID, `{"source":"external_request_approved"}`, now, tx); err != nil {
			return err
		}
		if err := s.queue.Enqueue(ctx, tx, "integrations", "outlook_create_event", map[string]any{
			"external_request_id": item.ID,
			"user_id":             item.EmployeeUserID,
		}, nil, now); err != nil {
			return err
		}
		if err := s.repo.CreateNotification(ctx, item.EmployeeUserID, "external_request_approved", "Request approved", "Your external training request has been approved", item.ID, now, tx); err != nil {
			return err
		}
		if err := s.repo.CreateApprovalHistory(ctx, item.ID, &step.ID, "approved", &fromStatus, &item.Status, principal.UserID, comment, now, tx); err != nil {
			return err
		}
		return s.outbox.Publish(ctx, tx, events.Message{
			Topic:      "external_training",
			EventType:  "external_request.approved",
			EntityType: "external_course_request",
			EntityID:   item.ID,
			Payload: map[string]any{
				"employee_user_id": item.EmployeeUserID,
			},
			OccurredAt: now,
		})
	}); err != nil {
		return ExternalRequest{}, err
	}
	return s.GetRequest(ctx, principal, item.ID)
}

func (s *Service) Reject(ctx context.Context, principal platformauth.Principal, id uuid.UUID, comment *string) (ExternalRequest, error) {
	item, err := s.GetRequest(ctx, principal, id)
	if err != nil {
		return ExternalRequest{}, err
	}
	if item.CurrentApprovalStepID == nil {
		return ExternalRequest{}, httpx.Conflict("approval_missing", "request has no current approval step")
	}
	if err := s.transitionApproval(ctx, principal, item, "rejected", "rejected", comment); err != nil {
		return ExternalRequest{}, err
	}
	return s.GetRequest(ctx, principal, item.ID)
}

func (s *Service) RequestRevision(ctx context.Context, principal platformauth.Principal, id uuid.UUID, comment *string) (ExternalRequest, error) {
	item, err := s.GetRequest(ctx, principal, id)
	if err != nil {
		return ExternalRequest{}, err
	}
	if item.CurrentApprovalStepID == nil {
		return ExternalRequest{}, httpx.Conflict("approval_missing", "request has no current approval step")
	}
	if err := s.transitionApproval(ctx, principal, item, "revision_requested", "needs_revision", comment); err != nil {
		return ExternalRequest{}, err
	}
	return s.GetRequest(ctx, principal, item.ID)
}

func (s *Service) UploadCertificate(ctx context.Context, principal platformauth.Principal, id uuid.UUID, req UploadRequestCertificateRequest) (ExternalRequest, error) {
	item, err := s.GetRequest(ctx, principal, id)
	if err != nil {
		return ExternalRequest{}, err
	}
	if item.EmployeeUserID != principal.UserID && !hasRole(principal, "admin") {
		return ExternalRequest{}, httpx.Forbidden("forbidden", "only owner can upload certificate")
	}
	switch item.Status {
	case "approved", "in_training", "completed":
	default:
		return ExternalRequest{}, httpx.Conflict("invalid_status", "certificate can only be uploaded after approval")
	}

	now := s.clock.Now()
	item.Status = "certificate_uploaded"
	item.CertificateUploadedAt = &now
	item.UpdatedAt = now

	err = db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		fileID := uuid.New()
		if err := s.repo.CreateFileAttachment(ctx, fileID, principal.UserID, req.StorageProvider, req.StorageKey, req.OriginalName, req.MimeType, req.SizeBytes, now, tx); err != nil {
			return err
		}
		if err := s.repo.CreateAttachedDocument(ctx, item.ID, fileID, principal.UserID, "certificate", now, tx); err != nil {
			return err
		}
		if err := s.repo.UpdateRequest(ctx, item, tx); err != nil {
			return err
		}
		return s.outbox.Publish(ctx, tx, events.Message{
			Topic:      "external_training",
			EventType:  "external_request.certificate_uploaded",
			EntityType: "external_course_request",
			EntityID:   item.ID,
			Payload: map[string]any{
				"employee_user_id": item.EmployeeUserID,
			},
			OccurredAt: now,
		})
	})
	if err != nil {
		return ExternalRequest{}, err
	}
	return s.GetRequest(ctx, principal, item.ID)
}

func (s *Service) CreateBudgetLimit(ctx context.Context, req CreateBudgetLimitRequest) (BudgetLimit, error) {
	now := s.clock.Now()
	item := BudgetLimit{
		ID:          uuid.New(),
		ScopeType:   req.ScopeType,
		ScopeID:     req.ScopeID,
		PeriodYear:  req.PeriodYear,
		PeriodMonth: req.PeriodMonth,
		LimitAmount: req.LimitAmount,
		Currency:    req.Currency,
		IsActive:    req.IsActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return item, s.repo.CreateBudgetLimit(ctx, item)
}

func (s *Service) ListBudgetLimits(ctx context.Context) ([]BudgetLimit, error) {
	return s.repo.ListBudgetLimits(ctx)
}

func (s *Service) CreateWorkflow(ctx context.Context, req CreateWorkflowRequest) (ApprovalWorkflow, error) {
	now := s.clock.Now()
	item := ApprovalWorkflow{
		ID:         uuid.New(),
		EntityType: req.EntityType,
		Name:       req.Name,
		IsActive:   req.IsActive,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	err := db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.repo.CreateWorkflow(ctx, item, tx); err != nil {
			return err
		}
		for _, stepReq := range req.Steps {
			if err := s.repo.CreateWorkflowStep(ctx, ApprovalWorkflowStep{
				ID:             uuid.New(),
				WorkflowID:     item.ID,
				StepOrder:      stepReq.StepOrder,
				RoleCode:       stepReq.RoleCode,
				ApproverSource: stepReq.ApproverSource,
				ApproverUserID: stepReq.ApproverUserID,
				SLAHours:       stepReq.SLAHours,
				IsRequired:     stepReq.IsRequired,
			}, tx); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return ApprovalWorkflow{}, err
	}
	item.Steps, _ = s.repo.ListWorkflowSteps(ctx, item.ID)
	return item, nil
}

func (s *Service) ListWorkflows(ctx context.Context) ([]ApprovalWorkflow, error) {
	items, err := s.repo.ListWorkflows(ctx)
	if err != nil {
		return nil, err
	}
	for i := range items {
		steps, err := s.repo.ListWorkflowSteps(ctx, items[i].ID)
		if err != nil {
			return nil, err
		}
		items[i].Steps = steps
	}
	return items, nil
}

func (s *Service) transitionApproval(ctx context.Context, principal platformauth.Principal, item ExternalRequest, stepStatus string, targetStatus string, comment *string) error {
	return db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		step, err := s.repo.GetApprovalStep(ctx, *item.CurrentApprovalStepID, tx)
		if err != nil {
			return err
		}
		if step.ApproverUserID != principal.UserID && !principal.HasPermission("settings.manage") {
			return httpx.Forbidden("forbidden", "only current approver can act")
		}

		now := s.clock.Now()
		fromStatus := item.Status
		step.Status = stepStatus
		step.Comment = comment
		step.ActedAt = &now
		if err := s.repo.UpdateApprovalStep(ctx, step, tx); err != nil {
			return err
		}
		item.Status = targetStatus
		item.CurrentApprovalStepID = nil
		item.UpdatedAt = now
		if targetStatus == "rejected" {
			item.RejectedAt = &now
		}
		if targetStatus == "needs_revision" {
			item.SentToRevisionAt = &now
		}
		if err := s.repo.UpdateRequest(ctx, item, tx); err != nil {
			return err
		}
		if err := s.repo.CreateApprovalHistory(ctx, item.ID, &step.ID, stepStatus, &fromStatus, &targetStatus, principal.UserID, comment, now, tx); err != nil {
			return err
		}

		notificationType := "external_request_rejected"
		title := "Request rejected"
		body := "Your external training request was rejected"
		if targetStatus == "needs_revision" {
			notificationType = "external_request_revision_requested"
			title = "Revision requested"
			body = "Your external training request was sent back for revision"
		}
		return s.repo.CreateNotification(ctx, item.EmployeeUserID, notificationType, title, body, item.ID, now, tx)
	})
}

func (s *Service) buildApprovalSteps(ctx context.Context, item ExternalRequest, exec ...db.DBTX) ([]ApprovalStep, error) {
	var workflowSteps []ApprovalWorkflowStep
	if workflow, err := s.repo.GetActiveWorkflow(ctx, "external_course_request", exec...); err != nil {
		return nil, err
	} else if workflow != nil {
		steps, err := s.repo.ListWorkflowSteps(ctx, workflow.ID, exec...)
		if err != nil {
			return nil, err
		}
		workflowSteps = steps
	}

	if len(workflowSteps) == 0 {
		workflowSteps = []ApprovalWorkflowStep{
			{StepOrder: 1, RoleCode: "manager", ApproverSource: "line_manager", IsRequired: true},
			{StepOrder: 2, RoleCode: "hr", ApproverSource: "specific_role", IsRequired: true},
		}
	}

	now := s.clock.Now()
	result := make([]ApprovalStep, 0, len(workflowSteps))
	for _, configured := range workflowSteps {
		approverUserID, err := s.resolveApprover(ctx, item, configured, exec...)
		if err != nil {
			return nil, err
		}
		if approverUserID == nil {
			if configured.IsRequired {
				return nil, httpx.Conflict("approver_not_found", "required approver could not be resolved")
			}
			continue
		}
		var dueAt *time.Time
		if configured.SLAHours != nil {
			value := now.Add(time.Duration(*configured.SLAHours) * time.Hour)
			dueAt = &value
		}
		result = append(result, ApprovalStep{
			ID:             uuid.New(),
			EntityType:     "external_course_request",
			EntityID:       item.ID,
			StepOrder:      configured.StepOrder,
			ApproverUserID: *approverUserID,
			RoleCode:       configured.RoleCode,
			Status:         "pending",
			DueAt:          dueAt,
			CreatedAt:      now,
		})
	}
	return result, nil
}

func (s *Service) resolveApprover(ctx context.Context, item ExternalRequest, step ApprovalWorkflowStep, exec ...db.DBTX) (*uuid.UUID, error) {
	switch step.ApproverSource {
	case "line_manager":
		return s.orgService.GetPrimaryManager(ctx, item.EmployeeUserID, exec...)
	case "specific_role":
		if step.ApproverUserID != nil {
			return step.ApproverUserID, nil
		}
		return s.identityRepo.FindUserIDByRoleCode(ctx, step.RoleCode, exec...)
	case "static_user":
		return step.ApproverUserID, nil
	case "department_head":
		if item.DepartmentID == nil {
			return nil, nil
		}
		return s.repo.ResolveDepartmentHead(ctx, *item.DepartmentID, exec...)
	default:
		return nil, nil
	}
}

func approvalStatusFromRole(roleCode string) string {
	switch roleCode {
	case "hr":
		return "hr_approval"
	default:
		return "manager_approval"
	}
}

func stringPtr(value string) *string {
	return &value
}

func requestSchedule(item ExternalRequest, fallback time.Time) (time.Time, time.Time) {
	startAt := fallback
	endAt := fallback.Add(2 * time.Hour)
	if item.PlannedStartDate != nil {
		startAt = time.Date(item.PlannedStartDate.Year(), item.PlannedStartDate.Month(), item.PlannedStartDate.Day(), 9, 0, 0, 0, time.UTC)
		endAt = startAt.Add(8 * time.Hour)
	}
	if item.PlannedEndDate != nil {
		endAt = time.Date(item.PlannedEndDate.Year(), item.PlannedEndDate.Month(), item.PlannedEndDate.Day(), 18, 0, 0, 0, time.UTC)
	}
	if !endAt.After(startAt) {
		endAt = startAt.Add(1 * time.Hour)
	}
	return startAt, endAt
}

type Handler struct {
	service   *Service
	validator *validator.Validate
}

func NewHandler(service *Service, validator *validator.Validate) *Handler {
	return &Handler{service: service, validator: validator}
}

func externalPrincipal(r *http.Request) (platformauth.Principal, error) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		return platformauth.Principal{}, httpx.Unauthorized("unauthorized", "authorization required")
	}
	return principal, nil
}

func (h *Handler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	principal, err := externalPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	var req CreateExternalRequestRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}
	item, err := h.service.CreateRequest(r.Context(), principal, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	principal, err := externalPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	items, err := h.service.ListMine(r.Context(), principal)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) GetRequest(w http.ResponseWriter, r *http.Request) {
	principal, err := externalPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_request_id", "invalid request id"))
		return
	}
	item, err := h.service.GetRequest(r.Context(), principal, id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) UpdateRequest(w http.ResponseWriter, r *http.Request) {
	principal, err := externalPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_request_id", "invalid request id"))
		return
	}
	var req UpdateExternalRequestRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	item, err := h.service.UpdateRequest(r.Context(), principal, id, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	principal, err := externalPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_request_id", "invalid request id"))
		return
	}
	item, err := h.service.Submit(r.Context(), principal, id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) Approve(w http.ResponseWriter, r *http.Request) {
	h.handleAction(w, r, (*Service).Approve)
}

func (h *Handler) Reject(w http.ResponseWriter, r *http.Request) {
	h.handleAction(w, r, (*Service).Reject)
}

func (h *Handler) RequestRevision(w http.ResponseWriter, r *http.Request) {
	h.handleAction(w, r, (*Service).RequestRevision)
}

func (h *Handler) handleAction(w http.ResponseWriter, r *http.Request, fn func(*Service, context.Context, platformauth.Principal, uuid.UUID, *string) (ExternalRequest, error)) {
	principal, err := externalPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_request_id", "invalid request id"))
		return
	}
	var req ActionCommentRequest
	if err := httpx.DecodeJSON(r, &req); err != nil && !errors.Is(err, http.ErrBodyNotAllowed) && !strings.Contains(err.Error(), "EOF") {
		httpx.WriteError(w, err)
		return
	}
	item, err := fn(h.service, r.Context(), principal, id, req.Comment)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) UploadCertificate(w http.ResponseWriter, r *http.Request) {
	principal, err := externalPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_request_id", "invalid request id"))
		return
	}
	var req UploadRequestCertificateRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}
	item, err := h.service.UploadCertificate(r.Context(), principal, id, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) CreateBudgetLimit(w http.ResponseWriter, r *http.Request) {
	var req CreateBudgetLimitRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}
	item, err := h.service.CreateBudgetLimit(r.Context(), req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) ListBudgetLimits(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListBudgetLimits(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) CreateWorkflow(w http.ResponseWriter, r *http.Request) {
	var req CreateWorkflowRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}
	item, err := h.service.CreateWorkflow(r.Context(), req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) ListWorkflows(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListWorkflows(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}
