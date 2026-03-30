package course_requests

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"html"
	"io"
	"net/http"
	"strings"
	"time"

	"moneyapp/backend/internal/modules/catalog"
	"moneyapp/backend/internal/modules/certificates"
	"moneyapp/backend/internal/modules/identity"
	"moneyapp/backend/internal/modules/learning"
	"moneyapp/backend/internal/modules/org"
	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/clock"
	"moneyapp/backend/internal/platform/db"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CourseRequest struct {
	ID                           uuid.UUID  `json:"id"`
	RequestNo                    string     `json:"request_no"`
	CourseID                     uuid.UUID  `json:"course_id"`
	CourseTitle                  string     `json:"course_title"`
	EmployeeUserID               uuid.UUID  `json:"employee_user_id"`
	EmployeeFullName             string     `json:"employee_full_name"`
	EmployeeEmail                string     `json:"employee_email"`
	DepartmentID                 *uuid.UUID `json:"department_id,omitempty"`
	ManagerUserID                *uuid.UUID `json:"manager_user_id,omitempty"`
	ManagerFullName              *string    `json:"manager_full_name,omitempty"`
	HRUserID                     *uuid.UUID `json:"hr_user_id,omitempty"`
	HRFullName                   *string    `json:"hr_full_name,omitempty"`
	EnrollmentID                 *uuid.UUID `json:"enrollment_id,omitempty"`
	CertificateID                *uuid.UUID `json:"certificate_id,omitempty"`
	CertificateOriginalName      *string    `json:"certificate_original_name,omitempty"`
	Status                       string     `json:"status"`
	DisplayStatus                string     `json:"display_status"`
	StatusLabel                  string     `json:"status_label"`
	CertificateApprovalSummary   string     `json:"certificate_approval_summary"`
	EmployeeComment              *string    `json:"employee_comment,omitempty"`
	ManagerComment               *string    `json:"manager_comment,omitempty"`
	HRComment                    *string    `json:"hr_comment,omitempty"`
	RejectionReason              *string    `json:"rejection_reason,omitempty"`
	DeadlineAt                   *time.Time `json:"deadline_at,omitempty"`
	RequestedAt                  time.Time  `json:"requested_at"`
	ManagerApprovedAt            *time.Time `json:"manager_approved_at,omitempty"`
	HRApprovedAt                 *time.Time `json:"hr_approved_at,omitempty"`
	ApprovedAt                   *time.Time `json:"approved_at,omitempty"`
	StartedAt                    *time.Time `json:"started_at,omitempty"`
	CompletedAt                  *time.Time `json:"completed_at,omitempty"`
	CertificateUploadedAt        *time.Time `json:"certificate_uploaded_at,omitempty"`
	CertificateApprovedAt        *time.Time `json:"certificate_approved_at,omitempty"`
	CertificateManagerApprovedAt *time.Time `json:"certificate_manager_approved_at,omitempty"`
	CertificateManagerApprovedBy *uuid.UUID `json:"certificate_manager_approved_by,omitempty"`
	CertificateHRApprovedAt      *time.Time `json:"certificate_hr_approved_at,omitempty"`
	CertificateHRApprovedBy      *uuid.UUID `json:"certificate_hr_approved_by,omitempty"`
	CanceledAt                   *time.Time `json:"canceled_at,omitempty"`
	RejectedAt                   *time.Time `json:"rejected_at,omitempty"`
	RejectedBy                   *uuid.UUID `json:"rejected_by,omitempty"`
	CreatedAt                    time.Time  `json:"created_at"`
	UpdatedAt                    time.Time  `json:"updated_at"`
}

type CreateCourseRequestRequest struct {
	CourseID        uuid.UUID  `json:"course_id" validate:"required"`
	EmployeeComment *string    `json:"employee_comment,omitempty"`
	DeadlineAt      *time.Time `json:"deadline_at,omitempty"`
}

type ActionCommentRequest struct {
	Comment *string `json:"comment,omitempty"`
}

type UploadCertificateRequest struct {
	CertificateNo   *string    `json:"certificate_no,omitempty"`
	IssuedBy        *string    `json:"issued_by,omitempty"`
	IssuedAt        *time.Time `json:"issued_at,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	Notes           *string    `json:"notes,omitempty"`
	StorageProvider string     `json:"storage_provider" validate:"required,oneof=s3 local minio"`
	StorageKey      string     `json:"storage_key" validate:"required"`
	OriginalName    string     `json:"original_name" validate:"required"`
	MimeType        string     `json:"mime_type" validate:"required"`
	SizeBytes       int64      `json:"size_bytes" validate:"required,min=1"`
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

func (r *Repository) CreateRequest(ctx context.Context, item CourseRequest, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into course_requests (
			id, request_no, course_id, employee_user_id, department_id, manager_user_id, hr_user_id,
			enrollment_id, certificate_id, status, employee_comment, manager_comment, hr_comment,
			rejection_reason, deadline_at, requested_at, manager_approved_at, hr_approved_at, approved_at,
			started_at, completed_at, certificate_uploaded_at, certificate_approved_at,
			certificate_manager_approved_at, certificate_manager_approved_by,
			certificate_hr_approved_at, certificate_hr_approved_by,
			canceled_at, rejected_at, rejected_by, created_at, updated_at
		)
		values (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13,
			$14, $15, $16, $17, $18, $19,
			$20, $21, $22, $23,
			$24, $25,
			$26, $27,
			$28, $29, $30, $31, $32
		)
	`, item.ID, item.RequestNo, item.CourseID, item.EmployeeUserID, item.DepartmentID, item.ManagerUserID, item.HRUserID,
		item.EnrollmentID, item.CertificateID, item.Status, item.EmployeeComment, item.ManagerComment, item.HRComment,
		item.RejectionReason, item.DeadlineAt, item.RequestedAt, item.ManagerApprovedAt, item.HRApprovedAt, item.ApprovedAt,
		item.StartedAt, item.CompletedAt, item.CertificateUploadedAt, item.CertificateApprovedAt,
		item.CertificateManagerApprovedAt, item.CertificateManagerApprovedBy,
		item.CertificateHRApprovedAt, item.CertificateHRApprovedBy,
		item.CanceledAt, item.RejectedAt, item.RejectedBy, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *Repository) UpdateRequest(ctx context.Context, item CourseRequest, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update course_requests
		set manager_user_id = $2,
		    hr_user_id = $3,
		    enrollment_id = $4,
		    certificate_id = $5,
		    status = $6,
		    employee_comment = $7,
		    manager_comment = $8,
		    hr_comment = $9,
		    rejection_reason = $10,
		    deadline_at = $11,
		    manager_approved_at = $12,
		    hr_approved_at = $13,
		    approved_at = $14,
		    started_at = $15,
		    completed_at = $16,
		    certificate_uploaded_at = $17,
		    certificate_approved_at = $18,
		    certificate_manager_approved_at = $19,
		    certificate_manager_approved_by = $20,
		    certificate_hr_approved_at = $21,
		    certificate_hr_approved_by = $22,
		    canceled_at = $23,
		    rejected_at = $24,
		    rejected_by = $25,
		    updated_at = $26
		where id = $1
	`, item.ID, item.ManagerUserID, item.HRUserID, item.EnrollmentID, item.CertificateID, item.Status, item.EmployeeComment,
		item.ManagerComment, item.HRComment, item.RejectionReason, item.DeadlineAt, item.ManagerApprovedAt, item.HRApprovedAt,
		item.ApprovedAt, item.StartedAt, item.CompletedAt, item.CertificateUploadedAt, item.CertificateApprovedAt,
		item.CertificateManagerApprovedAt, item.CertificateManagerApprovedBy, item.CertificateHRApprovedAt, item.CertificateHRApprovedBy,
		item.CanceledAt, item.RejectedAt, item.RejectedBy, item.UpdatedAt)
	return err
}

func (r *Repository) CreateEvent(ctx context.Context, requestID, actorID uuid.UUID, action string, comment *string, createdAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into course_request_events (id, course_request_id, action, performed_by, comment, created_at)
		values ($1, $2, $3, $4, $5, $6)
	`, uuid.New(), requestID, action, actorID, comment, createdAt)
	return err
}

func (r *Repository) CreateNotification(ctx context.Context, userID uuid.UUID, typ, title, body, relatedEntityType string, relatedEntityID uuid.UUID, createdAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into notifications (
			id, user_id, channel, type, title, body, status, related_entity_type, related_entity_id, created_at
		)
		values ($1, $2, 'in_app', $3, $4, $5, 'pending', $6, $7, $8)
	`, uuid.New(), userID, typ, title, body, relatedEntityType, relatedEntityID, createdAt)
	return err
}

func (r *Repository) HasActiveRequest(ctx context.Context, employeeUserID, courseID uuid.UUID, exec ...db.DBTX) (bool, error) {
	var exists bool
	err := r.base(exec...).QueryRowContext(ctx, `
		select exists (
			select 1
			from course_requests
			where employee_user_id = $1
			  and course_id = $2
			  and status not in ('rejected_by_manager', 'rejected_by_hr', 'canceled_by_employee', 'certificate_approved')
		)
	`, employeeUserID, courseID).Scan(&exists)
	return exists, err
}

func (r *Repository) queryRequests(ctx context.Context, where string, args []any, exec ...db.DBTX) ([]CourseRequest, error) {
	query := `
		select
			cr.id,
			cr.request_no,
			cr.course_id,
			c.title,
			cr.employee_user_id,
			trim(concat_ws(' ', ep.last_name, ep.first_name, ep.middle_name)) as employee_full_name,
			u.email,
			cr.department_id,
			cr.manager_user_id,
			nullif(trim(concat_ws(' ', mp.last_name, mp.first_name, mp.middle_name)), '') as manager_full_name,
			cr.hr_user_id,
			nullif(trim(concat_ws(' ', hp.last_name, hp.first_name, hp.middle_name)), '') as hr_full_name,
			cr.enrollment_id,
			cr.certificate_id,
			fa.original_name,
			cr.status,
			cr.employee_comment,
			cr.manager_comment,
			cr.hr_comment,
			cr.rejection_reason,
			cr.deadline_at,
			cr.requested_at,
			cr.manager_approved_at,
			cr.hr_approved_at,
			cr.approved_at,
			cr.started_at,
			cr.completed_at,
			cr.certificate_uploaded_at,
			cr.certificate_approved_at,
			cr.certificate_manager_approved_at,
			cr.certificate_manager_approved_by,
			cr.certificate_hr_approved_at,
			cr.certificate_hr_approved_by,
			cr.canceled_at,
			cr.rejected_at,
			cr.rejected_by,
			cr.created_at,
			cr.updated_at
		from course_requests cr
		join courses c on c.id = cr.course_id
		join users u on u.id = cr.employee_user_id
		left join employee_profiles ep on ep.user_id = cr.employee_user_id
		left join employee_profiles mp on mp.user_id = cr.manager_user_id
		left join employee_profiles hp on hp.user_id = cr.hr_user_id
		left join certificates cert on cert.id = cr.certificate_id
		left join file_attachments fa on fa.id = cert.file_id
	` + where + `
		order by cr.created_at desc
	`

	rows, err := r.base(exec...).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []CourseRequest
	for rows.Next() {
		var item CourseRequest
		if err := rows.Scan(
			&item.ID,
			&item.RequestNo,
			&item.CourseID,
			&item.CourseTitle,
			&item.EmployeeUserID,
			&item.EmployeeFullName,
			&item.EmployeeEmail,
			&item.DepartmentID,
			&item.ManagerUserID,
			&item.ManagerFullName,
			&item.HRUserID,
			&item.HRFullName,
			&item.EnrollmentID,
			&item.CertificateID,
			&item.CertificateOriginalName,
			&item.Status,
			&item.EmployeeComment,
			&item.ManagerComment,
			&item.HRComment,
			&item.RejectionReason,
			&item.DeadlineAt,
			&item.RequestedAt,
			&item.ManagerApprovedAt,
			&item.HRApprovedAt,
			&item.ApprovedAt,
			&item.StartedAt,
			&item.CompletedAt,
			&item.CertificateUploadedAt,
			&item.CertificateApprovedAt,
			&item.CertificateManagerApprovedAt,
			&item.CertificateManagerApprovedBy,
			&item.CertificateHRApprovedAt,
			&item.CertificateHRApprovedBy,
			&item.CanceledAt,
			&item.RejectedAt,
			&item.RejectedBy,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) GetRequest(ctx context.Context, id uuid.UUID, exec ...db.DBTX) (CourseRequest, error) {
	items, err := r.queryRequests(ctx, ` where cr.id = $1 `, []any{id}, exec...)
	if err != nil {
		return CourseRequest{}, err
	}
	if len(items) == 0 {
		return CourseRequest{}, sql.ErrNoRows
	}
	return items[0], nil
}

func (r *Repository) ListAllRequests(ctx context.Context, exec ...db.DBTX) ([]CourseRequest, error) {
	return r.queryRequests(ctx, ``, nil, exec...)
}

func (r *Repository) ListRequestsByEmployee(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) ([]CourseRequest, error) {
	return r.queryRequests(ctx, ` where cr.employee_user_id = $1 `, []any{userID}, exec...)
}

func (r *Repository) ListRequestsByManager(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) ([]CourseRequest, error) {
	return r.queryRequests(ctx, ` where cr.manager_user_id = $1 `, []any{userID}, exec...)
}

type Service struct {
	db               *sql.DB
	repo             *Repository
	identityRepo     *identity.Repository
	orgService       *org.Service
	catalogService   *catalog.Service
	learningRepo     *learning.Repository
	certificatesRepo *certificates.Repository
	clock            clock.Clock
}

func NewService(
	database *sql.DB,
	repo *Repository,
	identityRepo *identity.Repository,
	orgService *org.Service,
	catalogService *catalog.Service,
	learningRepo *learning.Repository,
	certificatesRepo *certificates.Repository,
	appClock clock.Clock,
) *Service {
	return &Service{
		db:               database,
		repo:             repo,
		identityRepo:     identityRepo,
		orgService:       orgService,
		catalogService:   catalogService,
		learningRepo:     learningRepo,
		certificatesRepo: certificatesRepo,
		clock:            appClock,
	}
}

func (s *Service) Create(ctx context.Context, principal platformauth.Principal, req CreateCourseRequestRequest) (CourseRequest, error) {
	if _, err := s.catalogService.GetCourse(ctx, req.CourseID); err != nil {
		return CourseRequest{}, err
	}

	managerID, err := s.orgService.GetPrimaryManager(ctx, principal.UserID)
	if err != nil {
		return CourseRequest{}, err
	}
	if managerID == nil {
		return CourseRequest{}, httpx.Conflict("manager_missing", "employee has no primary manager")
	}

	hrID, err := s.identityRepo.FindUserIDByRoleCode(ctx, "hr")
	if err != nil {
		return CourseRequest{}, err
	}
	if hrID == nil {
		return CourseRequest{}, httpx.Conflict("hr_missing", "hr approver could not be resolved")
	}

	hasActive, err := s.repo.HasActiveRequest(ctx, principal.UserID, req.CourseID)
	if err != nil {
		return CourseRequest{}, err
	}
	if hasActive {
		return CourseRequest{}, httpx.Conflict("active_request_exists", "active request for this course already exists")
	}

	now := s.clock.Now()
	deadlineAt := req.DeadlineAt
	if deadlineAt == nil {
		defaultDeadline := now.AddDate(0, 1, 0)
		deadlineAt = &defaultDeadline
	}

	item := CourseRequest{
		ID:              uuid.New(),
		RequestNo:       "CR-" + now.Format("20060102") + "-" + uuid.NewString()[:8],
		CourseID:        req.CourseID,
		EmployeeUserID:  principal.UserID,
		DepartmentID:    principal.DepartmentID,
		ManagerUserID:   managerID,
		HRUserID:        hrID,
		Status:          "pending_manager_approval",
		EmployeeComment: req.EmployeeComment,
		DeadlineAt:      deadlineAt,
		RequestedAt:     now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	err = db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.repo.CreateRequest(ctx, item, tx); err != nil {
			return err
		}
		if err := s.repo.CreateEvent(ctx, item.ID, principal.UserID, "created", req.EmployeeComment, now, tx); err != nil {
			return err
		}
		return s.repo.CreateNotification(ctx, *managerID, "course_request_approval_required", "Новая заявка на курс", "Сотрудник отправил заявку на обучение", "course_request", item.ID, now, tx)
	})
	if err != nil {
		return CourseRequest{}, err
	}

	return s.Get(ctx, principal, item.ID)
}

func (s *Service) List(ctx context.Context, principal platformauth.Principal) ([]CourseRequest, error) {
	var (
		items []CourseRequest
		err   error
	)

	switch {
	case hasRole(principal, "admin"), hasRole(principal, "hr"):
		items, err = s.repo.ListAllRequests(ctx)
	case hasRole(principal, "manager"):
		items, err = s.repo.ListRequestsByManager(ctx, principal.UserID)
	default:
		items, err = s.repo.ListRequestsByEmployee(ctx, principal.UserID)
	}
	if err != nil {
		return nil, err
	}

	for i := range items {
		s.decorate(&items[i])
	}

	return items, nil
}

func (s *Service) Get(ctx context.Context, principal platformauth.Principal, id uuid.UUID) (CourseRequest, error) {
	item, err := s.repo.GetRequest(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return CourseRequest{}, httpx.NotFound("course_request_not_found", "course request not found")
		}
		return CourseRequest{}, err
	}
	if !s.canView(principal, item) {
		return CourseRequest{}, httpx.Forbidden("forbidden", "permission denied")
	}

	s.decorate(&item)
	return item, nil
}

func (s *Service) ApproveManager(ctx context.Context, principal platformauth.Principal, id uuid.UUID, comment *string) (CourseRequest, error) {
	item, err := s.Get(ctx, principal, id)
	if err != nil {
		return CourseRequest{}, err
	}
	if item.Status != "pending_manager_approval" {
		return CourseRequest{}, httpx.Conflict("invalid_status", "request is not waiting for manager approval")
	}
	if item.ManagerUserID == nil || (*item.ManagerUserID != principal.UserID && !hasRole(principal, "admin")) {
		return CourseRequest{}, httpx.Forbidden("forbidden", "only assigned manager can approve")
	}

	now := s.clock.Now()
	item.Status = "pending_hr_approval"
	item.ManagerApprovedAt = &now
	item.ManagerComment = comment
	item.UpdatedAt = now

	err = db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.repo.UpdateRequest(ctx, item, tx); err != nil {
			return err
		}
		if err := s.repo.CreateEvent(ctx, item.ID, principal.UserID, "manager_approved", comment, now, tx); err != nil {
			return err
		}
		if item.HRUserID != nil {
			return s.repo.CreateNotification(ctx, *item.HRUserID, "course_request_hr_approval_required", "Нужен HR-апрув", "Заявка на курс ждёт HR-апрува", "course_request", item.ID, now, tx)
		}
		return nil
	})
	if err != nil {
		return CourseRequest{}, err
	}

	return s.Get(ctx, principal, id)
}

func (s *Service) ApproveHR(ctx context.Context, principal platformauth.Principal, id uuid.UUID, comment *string) (CourseRequest, error) {
	item, err := s.Get(ctx, principal, id)
	if err != nil {
		return CourseRequest{}, err
	}
	if item.Status != "pending_hr_approval" {
		return CourseRequest{}, httpx.Conflict("invalid_status", "request is not waiting for hr approval")
	}
	if !s.canActAsHR(principal, item) {
		return CourseRequest{}, httpx.Forbidden("forbidden", "only hr can approve this step")
	}

	now := s.clock.Now()
	enrollment := learning.Enrollment{
		ID:                uuid.New(),
		CourseID:          item.CourseID,
		UserID:            item.EmployeeUserID,
		AssignmentID:      nil,
		Source:            "self",
		Status:            "not_started",
		EnrolledAt:        now,
		DeadlineAt:        item.DeadlineAt,
		CompletionPercent: "0",
		IsMandatory:       false,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	item.Status = "approved_waiting_start"
	item.HRApprovedAt = &now
	item.ApprovedAt = &now
	item.HRComment = comment
	item.EnrollmentID = &enrollment.ID
	item.UpdatedAt = now

	err = db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.learningRepo.CreateEnrollment(ctx, enrollment, tx); err != nil {
			return err
		}
		if err := s.repo.UpdateRequest(ctx, item, tx); err != nil {
			return err
		}
		if err := s.repo.CreateEvent(ctx, item.ID, principal.UserID, "hr_approved", comment, now, tx); err != nil {
			return err
		}
		return s.repo.CreateNotification(ctx, item.EmployeeUserID, "course_request_approved", "Заявка на курс одобрена", "Можно начинать обучение", "course_request", item.ID, now, tx)
	})
	if err != nil {
		return CourseRequest{}, err
	}

	return s.Get(ctx, principal, id)
}

func (s *Service) Reject(ctx context.Context, principal platformauth.Principal, id uuid.UUID, comment *string) (CourseRequest, error) {
	item, err := s.Get(ctx, principal, id)
	if err != nil {
		return CourseRequest{}, err
	}

	now := s.clock.Now()
	var action string
	switch item.Status {
	case "pending_manager_approval":
		if item.ManagerUserID == nil || (*item.ManagerUserID != principal.UserID && !hasRole(principal, "admin")) {
			return CourseRequest{}, httpx.Forbidden("forbidden", "only assigned manager can reject")
		}
		item.Status = "rejected_by_manager"
		item.ManagerComment = comment
		action = "rejected_by_manager"
	case "pending_hr_approval":
		if !s.canActAsHR(principal, item) {
			return CourseRequest{}, httpx.Forbidden("forbidden", "only hr can reject this step")
		}
		item.Status = "rejected_by_hr"
		item.HRComment = comment
		action = "rejected_by_hr"
	default:
		return CourseRequest{}, httpx.Conflict("invalid_status", "request cannot be rejected in current state")
	}

	item.RejectedAt = &now
	item.RejectedBy = &principal.UserID
	item.RejectionReason = comment
	item.UpdatedAt = now

	err = db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.repo.UpdateRequest(ctx, item, tx); err != nil {
			return err
		}
		if err := s.repo.CreateEvent(ctx, item.ID, principal.UserID, action, comment, now, tx); err != nil {
			return err
		}
		return s.repo.CreateNotification(ctx, item.EmployeeUserID, "course_request_rejected", "Заявка на курс отклонена", "Проверьте комментарий согласующего", "course_request", item.ID, now, tx)
	})
	if err != nil {
		return CourseRequest{}, err
	}

	return s.Get(ctx, principal, id)
}

func (s *Service) Cancel(ctx context.Context, principal platformauth.Principal, id uuid.UUID, comment *string) (CourseRequest, error) {
	item, err := s.Get(ctx, principal, id)
	if err != nil {
		return CourseRequest{}, err
	}
	if item.EmployeeUserID != principal.UserID && !hasRole(principal, "admin") {
		return CourseRequest{}, httpx.Forbidden("forbidden", "only employee can cancel request")
	}
	if isFinalRequestStatus(item.Status) {
		return CourseRequest{}, httpx.Conflict("invalid_status", "request is already closed")
	}

	now := s.clock.Now()
	item.Status = "canceled_by_employee"
	item.CanceledAt = &now
	item.UpdatedAt = now

	err = db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if item.EnrollmentID != nil {
			enrollment, err := s.learningRepo.GetEnrollment(ctx, *item.EnrollmentID, tx)
			if err == nil {
				enrollment.Status = "canceled"
				enrollment.UpdatedAt = now
				if err := s.learningRepo.UpdateEnrollment(ctx, enrollment, tx); err != nil {
					return err
				}
			} else if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
		}
		if err := s.repo.UpdateRequest(ctx, item, tx); err != nil {
			return err
		}
		return s.repo.CreateEvent(ctx, item.ID, principal.UserID, "canceled_by_employee", comment, now, tx)
	})
	if err != nil {
		return CourseRequest{}, err
	}

	return s.Get(ctx, principal, id)
}

func (s *Service) Start(ctx context.Context, principal platformauth.Principal, id uuid.UUID) (CourseRequest, error) {
	item, err := s.Get(ctx, principal, id)
	if err != nil {
		return CourseRequest{}, err
	}
	if item.EmployeeUserID != principal.UserID && !hasRole(principal, "admin") {
		return CourseRequest{}, httpx.Forbidden("forbidden", "only employee can start the course")
	}
	if item.Status != "approved_waiting_start" {
		return CourseRequest{}, httpx.Conflict("invalid_status", "request is not waiting to start")
	}
	if item.EnrollmentID == nil {
		return CourseRequest{}, httpx.Conflict("enrollment_missing", "approved request has no enrollment")
	}

	now := s.clock.Now()
	err = db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		enrollment, err := s.learningRepo.GetEnrollment(ctx, *item.EnrollmentID, tx)
		if err != nil {
			return err
		}
		enrollment.Status = "in_progress"
		if enrollment.StartedAt == nil {
			enrollment.StartedAt = &now
		}
		enrollment.LastActivityAt = &now
		enrollment.UpdatedAt = now
		if err := s.learningRepo.UpdateEnrollment(ctx, enrollment, tx); err != nil {
			return err
		}

		item.Status = "in_progress"
		item.StartedAt = &now
		item.UpdatedAt = now
		if err := s.repo.UpdateRequest(ctx, item, tx); err != nil {
			return err
		}
		return s.repo.CreateEvent(ctx, item.ID, principal.UserID, "started", nil, now, tx)
	})
	if err != nil {
		return CourseRequest{}, err
	}

	return s.Get(ctx, principal, id)
}

func (s *Service) Complete(ctx context.Context, principal platformauth.Principal, id uuid.UUID, comment *string) (CourseRequest, error) {
	item, err := s.Get(ctx, principal, id)
	if err != nil {
		return CourseRequest{}, err
	}
	if item.EmployeeUserID != principal.UserID && !hasRole(principal, "admin") {
		return CourseRequest{}, httpx.Forbidden("forbidden", "only employee can complete the course")
	}
	if item.Status != "in_progress" {
		return CourseRequest{}, httpx.Conflict("invalid_status", "request is not in progress")
	}
	if item.EnrollmentID == nil {
		return CourseRequest{}, httpx.Conflict("enrollment_missing", "request has no enrollment")
	}

	now := s.clock.Now()
	err = db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		enrollment, err := s.learningRepo.GetEnrollment(ctx, *item.EnrollmentID, tx)
		if err != nil {
			return err
		}
		enrollment.Status = "completed"
		enrollment.CompletedAt = &now
		enrollment.LastActivityAt = &now
		enrollment.CompletionPercent = "100"
		enrollment.UpdatedAt = now
		if err := s.learningRepo.UpdateEnrollment(ctx, enrollment, tx); err != nil {
			return err
		}
		if err := s.learningRepo.CreateCompletionRecord(ctx, enrollment.ID, principal.UserID, "manual", nil, comment, now, tx); err != nil {
			return err
		}

		item.Status = "completed_waiting_certificate"
		item.CompletedAt = &now
		item.UpdatedAt = now
		if err := s.repo.UpdateRequest(ctx, item, tx); err != nil {
			return err
		}
		return s.repo.CreateEvent(ctx, item.ID, principal.UserID, "completed", comment, now, tx)
	})
	if err != nil {
		return CourseRequest{}, err
	}

	return s.Get(ctx, principal, id)
}

func (s *Service) UploadCertificate(ctx context.Context, principal platformauth.Principal, id uuid.UUID, req UploadCertificateRequest) (CourseRequest, error) {
	item, err := s.Get(ctx, principal, id)
	if err != nil {
		return CourseRequest{}, err
	}
	if item.EmployeeUserID != principal.UserID && !hasRole(principal, "admin") {
		return CourseRequest{}, httpx.Forbidden("forbidden", "only employee can upload certificate")
	}
	if item.Status != "completed_waiting_certificate" && item.Status != "certificate_rejected" && item.Status != "certificate_under_review" {
		return CourseRequest{}, httpx.Conflict("invalid_status", "request is not waiting for certificate upload")
	}

	now := s.clock.Now()
	file := certificates.FileAttachment{
		ID:              uuid.New(),
		StorageProvider: req.StorageProvider,
		StorageKey:      req.StorageKey,
		OriginalName:    req.OriginalName,
		MimeType:        req.MimeType,
		SizeBytes:       req.SizeBytes,
		UploadedBy:      &principal.UserID,
		CreatedAt:       now,
	}
	certificate := certificates.Certificate{
		ID:            uuid.New(),
		UserID:        item.EmployeeUserID,
		CourseID:      &item.CourseID,
		EnrollmentID:  item.EnrollmentID,
		CertificateNo: req.CertificateNo,
		IssuedBy:      req.IssuedBy,
		IssuedAt:      req.IssuedAt,
		ExpiresAt:     req.ExpiresAt,
		Status:        "under_review",
		FileID:        file.ID,
		UploadedAt:    now,
		Notes:         req.Notes,
	}

	item.Status = "certificate_under_review"
	item.CertificateID = &certificate.ID
	item.CertificateUploadedAt = &now
	item.CertificateApprovedAt = nil
	item.CertificateManagerApprovedAt = nil
	item.CertificateManagerApprovedBy = nil
	item.CertificateHRApprovedAt = nil
	item.CertificateHRApprovedBy = nil
	item.UpdatedAt = now

	err = db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.certificatesRepo.CreateFile(ctx, file, tx); err != nil {
			return err
		}
		if err := s.certificatesRepo.CreateCertificate(ctx, certificate, tx); err != nil {
			return err
		}
		if err := s.certificatesRepo.CreateVerification(ctx, certificate.ID, principal.UserID, "submit", nil, now, tx); err != nil {
			return err
		}
		if err := s.repo.UpdateRequest(ctx, item, tx); err != nil {
			return err
		}
		if err := s.repo.CreateEvent(ctx, item.ID, principal.UserID, "certificate_uploaded", req.Notes, now, tx); err != nil {
			return err
		}
		if item.ManagerUserID != nil {
			if err := s.repo.CreateNotification(ctx, *item.ManagerUserID, "course_certificate_review_required", "Нужен апрув сертификата", "Сотрудник загрузил сертификат по курсу", "course_request", item.ID, now, tx); err != nil {
				return err
			}
		}
		if item.HRUserID != nil {
			if err := s.repo.CreateNotification(ctx, *item.HRUserID, "course_certificate_review_required", "Нужен апрув сертификата", "Сотрудник загрузил сертификат по курсу", "course_request", item.ID, now, tx); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return CourseRequest{}, err
	}

	return s.Get(ctx, principal, id)
}

func (s *Service) ApproveCertificate(ctx context.Context, principal platformauth.Principal, id uuid.UUID, comment *string) (CourseRequest, error) {
	item, err := s.Get(ctx, principal, id)
	if err != nil {
		return CourseRequest{}, err
	}
	if item.CertificateID == nil {
		return CourseRequest{}, httpx.Conflict("certificate_missing", "request has no uploaded certificate")
	}
	if item.Status != "certificate_under_review" && item.Status != "certificate_approved" {
		return CourseRequest{}, httpx.Conflict("invalid_status", "certificate is not under review")
	}

	role, allowed := s.certificateReviewerRole(principal, item)
	if !allowed {
		return CourseRequest{}, httpx.Forbidden("forbidden", "only manager or hr can approve certificate")
	}

	now := s.clock.Now()
	action := "certificate_approved_by_" + role
	item.Status = "certificate_approved"
	if item.CertificateApprovedAt == nil {
		item.CertificateApprovedAt = &now
	}
	switch role {
	case "manager":
		item.CertificateManagerApprovedAt = &now
		item.CertificateManagerApprovedBy = &principal.UserID
	case "hr":
		item.CertificateHRApprovedAt = &now
		item.CertificateHRApprovedBy = &principal.UserID
	}
	item.UpdatedAt = now

	err = db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		certificate, err := s.certificatesRepo.GetCertificate(ctx, *item.CertificateID, tx)
		if err != nil {
			return err
		}
		certificate.Status = "verified"
		certificate.VerifiedAt = &now
		certificate.VerifiedBy = &principal.UserID
		certificate.Notes = comment
		if err := s.certificatesRepo.UpdateCertificate(ctx, certificate, tx); err != nil {
			return err
		}
		if err := s.certificatesRepo.CreateVerification(ctx, certificate.ID, principal.UserID, "verify", comment, now, tx); err != nil {
			return err
		}
		if err := s.repo.UpdateRequest(ctx, item, tx); err != nil {
			return err
		}
		if err := s.repo.CreateEvent(ctx, item.ID, principal.UserID, action, comment, now, tx); err != nil {
			return err
		}
		return s.repo.CreateNotification(ctx, item.EmployeeUserID, "course_certificate_approved", "Сертификат одобрен", "Сертификат по курсу одобрен", "course_request", item.ID, now, tx)
	})
	if err != nil {
		return CourseRequest{}, err
	}

	return s.Get(ctx, principal, id)
}

func (s *Service) RejectCertificate(ctx context.Context, principal platformauth.Principal, id uuid.UUID, comment *string) (CourseRequest, error) {
	item, err := s.Get(ctx, principal, id)
	if err != nil {
		return CourseRequest{}, err
	}
	if item.CertificateID == nil {
		return CourseRequest{}, httpx.Conflict("certificate_missing", "request has no uploaded certificate")
	}
	if item.Status != "certificate_under_review" && item.Status != "certificate_approved" {
		return CourseRequest{}, httpx.Conflict("invalid_status", "certificate is not under review")
	}

	role, allowed := s.certificateReviewerRole(principal, item)
	if !allowed {
		return CourseRequest{}, httpx.Forbidden("forbidden", "only manager or hr can reject certificate")
	}

	now := s.clock.Now()
	action := "certificate_rejected_by_" + role
	item.Status = "certificate_rejected"
	item.UpdatedAt = now

	err = db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		certificate, err := s.certificatesRepo.GetCertificate(ctx, *item.CertificateID, tx)
		if err != nil {
			return err
		}
		certificate.Status = "rejected"
		certificate.VerifiedAt = nil
		certificate.VerifiedBy = nil
		certificate.Notes = comment
		if err := s.certificatesRepo.UpdateCertificate(ctx, certificate, tx); err != nil {
			return err
		}
		if err := s.certificatesRepo.CreateVerification(ctx, certificate.ID, principal.UserID, "reject", comment, now, tx); err != nil {
			return err
		}
		if err := s.repo.UpdateRequest(ctx, item, tx); err != nil {
			return err
		}
		if err := s.repo.CreateEvent(ctx, item.ID, principal.UserID, action, comment, now, tx); err != nil {
			return err
		}
		return s.repo.CreateNotification(ctx, item.EmployeeUserID, "course_certificate_rejected", "Сертификат отклонён", "Загрузите корректный сертификат заново", "course_request", item.ID, now, tx)
	})
	if err != nil {
		return CourseRequest{}, err
	}

	return s.Get(ctx, principal, id)
}

func (s *Service) ExportExcel(ctx context.Context, principal platformauth.Principal) ([]byte, string, error) {
	if !hasRole(principal, "hr") && !hasRole(principal, "admin") {
		return nil, "", httpx.Forbidden("forbidden", "only hr can export the report")
	}

	items, err := s.repo.ListAllRequests(ctx)
	if err != nil {
		return nil, "", err
	}
	for i := range items {
		s.decorate(&items[i])
	}

	var buf bytes.Buffer
	buf.WriteString("<html><head><meta charset=\"utf-8\"></head><body><table border=\"1\">")
	buf.WriteString("<tr>")
	for _, header := range []string{
		"Номер заявки",
		"ФИО",
		"Email",
		"Курс",
		"Статус",
		"Апрув руководителя",
		"Апрув HR",
		"Апрувы сертификата",
		"Дедлайн",
		"Дата заявки",
		"Старт курса",
		"Завершение курса",
		"Сертификат",
	} {
		buf.WriteString("<th>" + html.EscapeString(header) + "</th>")
	}
	buf.WriteString("</tr>")

	for _, item := range items {
		buf.WriteString("<tr>")
		values := []string{
			item.RequestNo,
			item.EmployeeFullName,
			item.EmployeeEmail,
			item.CourseTitle,
			item.StatusLabel,
			approvalCell(item.ManagerApprovedAt),
			approvalCell(item.HRApprovedAt),
			item.CertificateApprovalSummary,
			timeCell(item.DeadlineAt),
			item.RequestedAt.Format(time.RFC3339),
			timeCell(item.StartedAt),
			timeCell(item.CompletedAt),
			stringOrDash(item.CertificateOriginalName),
		}
		for _, value := range values {
			buf.WriteString("<td>" + html.EscapeString(value) + "</td>")
		}
		buf.WriteString("</tr>")
	}
	buf.WriteString("</table></body></html>")

	filename := "course-requests-report-" + s.clock.Now().Format("20060102-150405") + ".xls"
	return buf.Bytes(), filename, nil
}

func (s *Service) canView(principal platformauth.Principal, item CourseRequest) bool {
	if hasRole(principal, "admin") || hasRole(principal, "hr") {
		return true
	}
	if item.EmployeeUserID == principal.UserID {
		return true
	}
	return item.ManagerUserID != nil && *item.ManagerUserID == principal.UserID
}

func (s *Service) canActAsHR(principal platformauth.Principal, item CourseRequest) bool {
	if hasRole(principal, "admin") {
		return true
	}
	if !hasRole(principal, "hr") {
		return false
	}
	if item.HRUserID == nil {
		return true
	}
	return *item.HRUserID == principal.UserID
}

func (s *Service) certificateReviewerRole(principal platformauth.Principal, item CourseRequest) (string, bool) {
	if hasRole(principal, "admin") {
		return "hr", true
	}
	if item.ManagerUserID != nil && *item.ManagerUserID == principal.UserID {
		return "manager", true
	}
	if s.canActAsHR(principal, item) {
		return "hr", true
	}
	return "", false
}

func (s *Service) decorate(item *CourseRequest) {
	displayStatus := item.Status
	if item.DeadlineAt != nil && item.DeadlineAt.Before(s.clock.Now()) && item.Status != "certificate_approved" && item.Status != "canceled_by_employee" && item.Status != "rejected_by_manager" && item.Status != "rejected_by_hr" {
		displayStatus = "deadline_missed"
	}
	item.DisplayStatus = displayStatus
	item.StatusLabel = statusLabel(displayStatus)
	item.CertificateApprovalSummary = certificateApprovalSummary(item)
}

type Handler struct {
	service   *Service
	validator *validator.Validate
}

func NewHandler(service *Service, validator *validator.Validate) *Handler {
	return &Handler{service: service, validator: validator}
}

func courseRequestsPrincipal(r *http.Request) (platformauth.Principal, error) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		return platformauth.Principal{}, httpx.Unauthorized("unauthorized", "authorization required")
	}
	return principal, nil
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	principal, err := courseRequestsPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	var req CreateCourseRequestRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}
	item, err := h.service.Create(r.Context(), principal, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	principal, err := courseRequestsPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	items, err := h.service.List(r.Context(), principal)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	principal, err := courseRequestsPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_course_request_id", "invalid course request id"))
		return
	}
	item, err := h.service.Get(r.Context(), principal, id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) ApproveManager(w http.ResponseWriter, r *http.Request) {
	h.handleCommentAction(w, r, h.service.ApproveManager)
}

func (h *Handler) ApproveHR(w http.ResponseWriter, r *http.Request) {
	h.handleCommentAction(w, r, h.service.ApproveHR)
}

func (h *Handler) Reject(w http.ResponseWriter, r *http.Request) {
	h.handleCommentAction(w, r, h.service.Reject)
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	h.handleCommentAction(w, r, h.service.Cancel)
}

func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
	principal, err := courseRequestsPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_course_request_id", "invalid course request id"))
		return
	}
	item, err := h.service.Start(r.Context(), principal, id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	h.handleCommentAction(w, r, h.service.Complete)
}

func (h *Handler) UploadCertificate(w http.ResponseWriter, r *http.Request) {
	principal, err := courseRequestsPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_course_request_id", "invalid course request id"))
		return
	}
	var req UploadCertificateRequest
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

func (h *Handler) ApproveCertificate(w http.ResponseWriter, r *http.Request) {
	h.handleCommentAction(w, r, h.service.ApproveCertificate)
}

func (h *Handler) RejectCertificate(w http.ResponseWriter, r *http.Request) {
	h.handleCommentAction(w, r, h.service.RejectCertificate)
}

func (h *Handler) ExportExcel(w http.ResponseWriter, r *http.Request) {
	principal, err := courseRequestsPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	payload, filename, err := h.service.ExportExcel(r.Context(), principal)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/vnd.ms-excel; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}

func (h *Handler) handleCommentAction(
	w http.ResponseWriter,
	r *http.Request,
	action func(context.Context, platformauth.Principal, uuid.UUID, *string) (CourseRequest, error),
) {
	principal, err := courseRequestsPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_course_request_id", "invalid course request id"))
		return
	}
	var req ActionCommentRequest
	if err := httpx.DecodeJSON(r, &req); err != nil && !errors.Is(err, io.EOF) {
		httpx.WriteError(w, err)
		return
	}
	item, err := action(r.Context(), principal, id, req.Comment)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func hasRole(principal platformauth.Principal, roleCode string) bool {
	for _, role := range principal.RoleCodes {
		if role == roleCode {
			return true
		}
	}
	return false
}

func isFinalRequestStatus(status string) bool {
	switch status {
	case "certificate_approved", "rejected_by_manager", "rejected_by_hr", "canceled_by_employee":
		return true
	default:
		return false
	}
}

func statusLabel(status string) string {
	switch status {
	case "pending_manager_approval":
		return "Ожидает апрува руководителя"
	case "pending_hr_approval":
		return "Ожидает апрува HR"
	case "approved_waiting_start":
		return "Апрув получен, ожидает старта курса"
	case "in_progress":
		return "Курс начался, сотрудник учится"
	case "completed_waiting_certificate":
		return "Курс завершён, ожидает сертификат"
	case "certificate_under_review":
		return "Сертификат загружен, идёт проверка"
	case "certificate_approved":
		return "Сертификат одобрен"
	case "certificate_rejected":
		return "Сертификат отклонён"
	case "deadline_missed":
		return "Дедлайн курса пропущен"
	case "canceled_by_employee":
		return "Сотрудник отказался от курса"
	case "rejected_by_manager":
		return "Отклонено руководителем"
	case "rejected_by_hr":
		return "Отклонено HR"
	default:
		return status
	}
}

func certificateApprovalSummary(item *CourseRequest) string {
	parts := make([]string, 0, 2)
	if item.CertificateManagerApprovedAt != nil {
		parts = append(parts, "руководитель")
	}
	if item.CertificateHRApprovedAt != nil {
		parts = append(parts, "HR")
	}
	if len(parts) == 0 {
		if item.Status == "certificate_rejected" {
			return "Отклонён"
		}
		if item.CertificateID != nil {
			return "Ожидает апрува"
		}
		return "-"
	}
	return "Одобрили: " + strings.Join(parts, ", ")
}

func approvalCell(ts *time.Time) string {
	if ts == nil {
		return "Нет"
	}
	return ts.Format(time.RFC3339)
}

func timeCell(ts *time.Time) string {
	if ts == nil {
		return "-"
	}
	return ts.Format(time.RFC3339)
}

func stringOrDash(value *string) string {
	if value == nil || *value == "" {
		return "-"
	}
	return *value
}
