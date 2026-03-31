package course_intakes

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/clock"
	"moneyapp/backend/internal/platform/db"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Models
// ---------------------------------------------------------------------------

type Intake struct {
	ID                  uuid.UUID  `json:"id"`
	CourseID            *uuid.UUID `json:"course_id,omitempty"`
	Title               string     `json:"title"`
	Description         *string    `json:"description,omitempty"`
	OpenedBy            uuid.UUID  `json:"opened_by"`
	ApproverID          *uuid.UUID `json:"approver_id,omitempty"`
	MaxParticipants     *int       `json:"max_participants,omitempty"`
	StartDate           *string    `json:"start_date,omitempty"`
	EndDate             *string    `json:"end_date,omitempty"`
	ApplicationDeadline *time.Time `json:"application_deadline,omitempty"`
	Status              string     `json:"status"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type Application struct {
	ID                uuid.UUID  `json:"id"`
	IntakeID          uuid.UUID  `json:"intake_id"`
	ApplicantID       uuid.UUID  `json:"applicant_id"`
	Motivation        *string    `json:"motivation,omitempty"`
	Status            string     `json:"status"`
	ManagerApproverID *uuid.UUID `json:"manager_approver_id,omitempty"`
	ManagerComment    *string    `json:"manager_comment,omitempty"`
	ManagerDecidedAt  *time.Time `json:"manager_decided_at,omitempty"`
	HRApproverID      *uuid.UUID `json:"hr_approver_id,omitempty"`
	HRComment         *string    `json:"hr_comment,omitempty"`
	HRDecidedAt       *time.Time `json:"hr_decided_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type Suggestion struct {
	ID            uuid.UUID  `json:"id"`
	SuggestedBy   uuid.UUID  `json:"suggested_by"`
	Title         string     `json:"title"`
	Description   *string    `json:"description,omitempty"`
	ExternalURL   *string    `json:"external_url,omitempty"`
	ProviderName  *string    `json:"provider_name,omitempty"`
	Price         *string    `json:"price,omitempty"`
	PriceCurrency string     `json:"price_currency"`
	DurationHours *string    `json:"duration_hours,omitempty"`
	ApproverID    *uuid.UUID `json:"approver_id,omitempty"`
	Status        string     `json:"status"`
	ReviewedBy    *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewComment *string    `json:"review_comment,omitempty"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	IntakeID      *uuid.UUID `json:"intake_id,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

type CreateIntakeRequest struct {
	CourseID            *uuid.UUID `json:"course_id"`
	Title               string     `json:"title" validate:"required,max=500"`
	Description         *string    `json:"description"`
	ApproverID          *uuid.UUID `json:"approver_id"`
	MaxParticipants     *int       `json:"max_participants"`
	StartDate           *string    `json:"start_date"`
	EndDate             *string    `json:"end_date"`
	ApplicationDeadline *time.Time `json:"application_deadline"`
}

type UpdateIntakeRequest struct {
	Title               *string    `json:"title" validate:"omitempty,max=500"`
	Description         *string    `json:"description"`
	ApproverID          *uuid.UUID `json:"approver_id"`
	MaxParticipants     *int       `json:"max_participants"`
	StartDate           *string    `json:"start_date"`
	EndDate             *string    `json:"end_date"`
	ApplicationDeadline *time.Time `json:"application_deadline"`
	Status              *string    `json:"status" validate:"omitempty,oneof=open closed canceled completed"`
}

type ApplyRequest struct {
	IntakeID   uuid.UUID `json:"intake_id" validate:"required"`
	Motivation *string   `json:"motivation"`
}

type ApproveRejectRequest struct {
	Comment *string `json:"comment"`
}

type CreateSuggestionRequest struct {
	Title         string     `json:"title" validate:"required,max=500"`
	Description   *string    `json:"description"`
	ExternalURL   *string    `json:"external_url"`
	ProviderName  *string    `json:"provider_name"`
	Price         *string    `json:"price"`
	PriceCurrency *string    `json:"price_currency"`
	DurationHours *string    `json:"duration_hours"`
	ApproverID    *uuid.UUID `json:"approver_id"`
}

type ReviewSuggestionRequest struct {
	Comment *string `json:"comment"`
}

// ---------------------------------------------------------------------------
// Repository
// ---------------------------------------------------------------------------

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

// --- Intakes ---

func (r *Repository) CreateIntake(ctx context.Context, item Intake, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		INSERT INTO course_intakes (id, course_id, title, description, opened_by, approver_id,
			max_participants, start_date, end_date, application_deadline, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`, item.ID, item.CourseID, item.Title, item.Description, item.OpenedBy, item.ApproverID,
		item.MaxParticipants, item.StartDate, item.EndDate, item.ApplicationDeadline,
		item.Status, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *Repository) GetIntake(ctx context.Context, id uuid.UUID) (*Intake, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, course_id, title, description, opened_by, approver_id,
			max_participants, start_date, end_date, application_deadline, status, created_at, updated_at
		FROM course_intakes WHERE id = $1
	`, id)
	return scanIntake(row)
}

func (r *Repository) ListIntakes(ctx context.Context, status string) ([]Intake, error) {
	q := `SELECT id, course_id, title, description, opened_by, approver_id,
		max_participants, start_date, end_date, application_deadline, status, created_at, updated_at
		FROM course_intakes`
	var args []any
	if status != "" {
		q += " WHERE status = $1"
		args = append(args, status)
	}
	q += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Intake
	for rows.Next() {
		var it Intake
		if err := rows.Scan(&it.ID, &it.CourseID, &it.Title, &it.Description, &it.OpenedBy, &it.ApproverID,
			&it.MaxParticipants, &it.StartDate, &it.EndDate, &it.ApplicationDeadline,
			&it.Status, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, it)
	}
	return list, rows.Err()
}

func (r *Repository) UpdateIntake(ctx context.Context, item Intake, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		UPDATE course_intakes
		SET title=$2, description=$3, approver_id=$4, max_participants=$5,
			start_date=$6, end_date=$7, application_deadline=$8, status=$9, updated_at=$10
		WHERE id=$1
	`, item.ID, item.Title, item.Description, item.ApproverID, item.MaxParticipants,
		item.StartDate, item.EndDate, item.ApplicationDeadline, item.Status, item.UpdatedAt)
	return err
}

// --- Applications ---

func (r *Repository) CreateApplication(ctx context.Context, item Application, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		INSERT INTO course_applications (id, intake_id, applicant_id, motivation, status,
			manager_approver_id, manager_comment, manager_decided_at,
			hr_approver_id, hr_comment, hr_decided_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`, item.ID, item.IntakeID, item.ApplicantID, item.Motivation, item.Status,
		item.ManagerApproverID, item.ManagerComment, item.ManagerDecidedAt,
		item.HRApproverID, item.HRComment, item.HRDecidedAt,
		item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *Repository) GetApplication(ctx context.Context, id uuid.UUID) (*Application, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, intake_id, applicant_id, motivation, status,
			manager_approver_id, manager_comment, manager_decided_at,
			hr_approver_id, hr_comment, hr_decided_at, created_at, updated_at
		FROM course_applications WHERE id = $1
	`, id)
	return scanApplication(row)
}

func (r *Repository) ListApplicationsByIntake(ctx context.Context, intakeID uuid.UUID) ([]Application, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, intake_id, applicant_id, motivation, status,
			manager_approver_id, manager_comment, manager_decided_at,
			hr_approver_id, hr_comment, hr_decided_at, created_at, updated_at
		FROM course_applications WHERE intake_id = $1 ORDER BY created_at DESC
	`, intakeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanApplications(rows)
}

func (r *Repository) ListMyApplications(ctx context.Context, userID uuid.UUID) ([]Application, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, intake_id, applicant_id, motivation, status,
			manager_approver_id, manager_comment, manager_decided_at,
			hr_approver_id, hr_comment, hr_decided_at, created_at, updated_at
		FROM course_applications WHERE applicant_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanApplications(rows)
}

func (r *Repository) ListPendingManagerApprovals(ctx context.Context, managerID uuid.UUID) ([]Application, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, intake_id, applicant_id, motivation, status,
			manager_approver_id, manager_comment, manager_decided_at,
			hr_approver_id, hr_comment, hr_decided_at, created_at, updated_at
		FROM course_applications
		WHERE manager_approver_id = $1 AND status = 'pending_manager'
		ORDER BY created_at DESC
	`, managerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanApplications(rows)
}

func (r *Repository) UpdateApplication(ctx context.Context, item Application, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		UPDATE course_applications
		SET status=$2, manager_approver_id=$3, manager_comment=$4, manager_decided_at=$5,
			hr_approver_id=$6, hr_comment=$7, hr_decided_at=$8, updated_at=$9
		WHERE id=$1
	`, item.ID, item.Status, item.ManagerApproverID, item.ManagerComment, item.ManagerDecidedAt,
		item.HRApproverID, item.HRComment, item.HRDecidedAt, item.UpdatedAt)
	return err
}

// --- Suggestions ---

func (r *Repository) CreateSuggestion(ctx context.Context, item Suggestion, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		INSERT INTO course_suggestions (id, suggested_by, title, description, external_url,
			provider_name, price, price_currency, duration_hours, approver_id, status,
			reviewed_by, review_comment, reviewed_at, intake_id, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
	`, item.ID, item.SuggestedBy, item.Title, item.Description, item.ExternalURL,
		item.ProviderName, item.Price, item.PriceCurrency, item.DurationHours,
		item.ApproverID, item.Status, item.ReviewedBy, item.ReviewComment, item.ReviewedAt,
		item.IntakeID, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *Repository) GetSuggestion(ctx context.Context, id uuid.UUID) (*Suggestion, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, suggested_by, title, description, external_url,
			provider_name, price, price_currency, duration_hours, approver_id, status,
			reviewed_by, review_comment, reviewed_at, intake_id, created_at, updated_at
		FROM course_suggestions WHERE id = $1
	`, id)
	return scanSuggestion(row)
}

func (r *Repository) ListSuggestions(ctx context.Context, status string) ([]Suggestion, error) {
	q := `SELECT id, suggested_by, title, description, external_url,
		provider_name, price, price_currency, duration_hours, approver_id, status,
		reviewed_by, review_comment, reviewed_at, intake_id, created_at, updated_at
		FROM course_suggestions`
	var args []any
	if status != "" {
		q += " WHERE status = $1"
		args = append(args, status)
	}
	q += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSuggestions(rows)
}

func (r *Repository) ListMySuggestions(ctx context.Context, userID uuid.UUID) ([]Suggestion, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, suggested_by, title, description, external_url,
			provider_name, price, price_currency, duration_hours, approver_id, status,
			reviewed_by, review_comment, reviewed_at, intake_id, created_at, updated_at
		FROM course_suggestions WHERE suggested_by = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSuggestions(rows)
}

func (r *Repository) UpdateSuggestion(ctx context.Context, item Suggestion, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		UPDATE course_suggestions
		SET status=$2, reviewed_by=$3, review_comment=$4, reviewed_at=$5, intake_id=$6, updated_at=$7
		WHERE id=$1
	`, item.ID, item.Status, item.ReviewedBy, item.ReviewComment, item.ReviewedAt, item.IntakeID, item.UpdatedAt)
	return err
}

// ---------------------------------------------------------------------------
// Row scanners
// ---------------------------------------------------------------------------

type rowScanner interface {
	Scan(dest ...any) error
}

func scanIntake(row rowScanner) (*Intake, error) {
	var it Intake
	err := row.Scan(&it.ID, &it.CourseID, &it.Title, &it.Description, &it.OpenedBy, &it.ApproverID,
		&it.MaxParticipants, &it.StartDate, &it.EndDate, &it.ApplicationDeadline,
		&it.Status, &it.CreatedAt, &it.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &it, err
}

func scanApplication(row rowScanner) (*Application, error) {
	var a Application
	err := row.Scan(&a.ID, &a.IntakeID, &a.ApplicantID, &a.Motivation, &a.Status,
		&a.ManagerApproverID, &a.ManagerComment, &a.ManagerDecidedAt,
		&a.HRApproverID, &a.HRComment, &a.HRDecidedAt,
		&a.CreatedAt, &a.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &a, err
}

func scanApplications(rows *sql.Rows) ([]Application, error) {
	var list []Application
	for rows.Next() {
		var a Application
		if err := rows.Scan(&a.ID, &a.IntakeID, &a.ApplicantID, &a.Motivation, &a.Status,
			&a.ManagerApproverID, &a.ManagerComment, &a.ManagerDecidedAt,
			&a.HRApproverID, &a.HRComment, &a.HRDecidedAt,
			&a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func scanSuggestion(row rowScanner) (*Suggestion, error) {
	var s Suggestion
	err := row.Scan(&s.ID, &s.SuggestedBy, &s.Title, &s.Description, &s.ExternalURL,
		&s.ProviderName, &s.Price, &s.PriceCurrency, &s.DurationHours,
		&s.ApproverID, &s.Status, &s.ReviewedBy, &s.ReviewComment, &s.ReviewedAt,
		&s.IntakeID, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func scanSuggestions(rows *sql.Rows) ([]Suggestion, error) {
	var list []Suggestion
	for rows.Next() {
		var s Suggestion
		if err := rows.Scan(&s.ID, &s.SuggestedBy, &s.Title, &s.Description, &s.ExternalURL,
			&s.ProviderName, &s.Price, &s.PriceCurrency, &s.DurationHours,
			&s.ApproverID, &s.Status, &s.ReviewedBy, &s.ReviewComment, &s.ReviewedAt,
			&s.IntakeID, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type Service struct {
	db    *sql.DB
	repo  *Repository
	clock clock.Clock
}

func NewService(database *sql.DB, repo *Repository, clk clock.Clock) *Service {
	return &Service{db: database, repo: repo, clock: clk}
}

// --- Intakes ---

func (s *Service) CreateIntake(ctx context.Context, principal platformauth.Principal, req CreateIntakeRequest) (*Intake, error) {
	now := s.clock.Now()
	intake := Intake{
		ID:                  uuid.New(),
		CourseID:            req.CourseID,
		Title:               req.Title,
		Description:         req.Description,
		OpenedBy:            principal.UserID,
		ApproverID:          req.ApproverID,
		MaxParticipants:     req.MaxParticipants,
		StartDate:           req.StartDate,
		EndDate:             req.EndDate,
		ApplicationDeadline: req.ApplicationDeadline,
		Status:              "open",
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := s.repo.CreateIntake(ctx, intake); err != nil {
		return nil, err
	}
	return &intake, nil
}

func (s *Service) GetIntake(ctx context.Context, id uuid.UUID) (*Intake, error) {
	return s.repo.GetIntake(ctx, id)
}

func (s *Service) ListIntakes(ctx context.Context, status string) ([]Intake, error) {
	return s.repo.ListIntakes(ctx, status)
}

func (s *Service) UpdateIntake(ctx context.Context, principal platformauth.Principal, id uuid.UUID, req UpdateIntakeRequest) (*Intake, error) {
	intake, err := s.repo.GetIntake(ctx, id)
	if err != nil {
		return nil, err
	}
	if intake == nil {
		return nil, httpx.NotFound("not_found", "intake not found")
	}

	if req.Title != nil {
		intake.Title = *req.Title
	}
	if req.Description != nil {
		intake.Description = req.Description
	}
	if req.ApproverID != nil {
		intake.ApproverID = req.ApproverID
	}
	if req.MaxParticipants != nil {
		intake.MaxParticipants = req.MaxParticipants
	}
	if req.StartDate != nil {
		intake.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		intake.EndDate = req.EndDate
	}
	if req.ApplicationDeadline != nil {
		intake.ApplicationDeadline = req.ApplicationDeadline
	}
	if req.Status != nil {
		intake.Status = *req.Status
	}
	intake.UpdatedAt = s.clock.Now()

	if err := s.repo.UpdateIntake(ctx, *intake); err != nil {
		return nil, err
	}
	return intake, nil
}

func (s *Service) CloseIntake(ctx context.Context, id uuid.UUID) (*Intake, error) {
	intake, err := s.repo.GetIntake(ctx, id)
	if err != nil {
		return nil, err
	}
	if intake == nil {
		return nil, httpx.NotFound("not_found", "intake not found")
	}
	if intake.Status != "open" {
		return nil, httpx.BadRequest("invalid_status", "intake is not open")
	}
	intake.Status = "closed"
	intake.UpdatedAt = s.clock.Now()
	if err := s.repo.UpdateIntake(ctx, *intake); err != nil {
		return nil, err
	}
	return intake, nil
}

// --- Applications ---

func (s *Service) Apply(ctx context.Context, principal platformauth.Principal, req ApplyRequest) (*Application, error) {
	intake, err := s.repo.GetIntake(ctx, req.IntakeID)
	if err != nil {
		return nil, err
	}
	if intake == nil {
		return nil, httpx.NotFound("not_found", "intake not found")
	}
	if intake.Status != "open" {
		return nil, httpx.BadRequest("intake_closed", "intake is not accepting applications")
	}
	if intake.ApplicationDeadline != nil && s.clock.Now().After(*intake.ApplicationDeadline) {
		return nil, httpx.BadRequest("deadline_passed", "application deadline has passed")
	}

	now := s.clock.Now()
	app := Application{
		ID:          uuid.New(),
		IntakeID:    req.IntakeID,
		ApplicantID: principal.UserID,
		Motivation:  req.Motivation,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Determine initial status based on whether intake has a manager approver
	if intake.ApproverID != nil {
		app.Status = "pending_manager"
		app.ManagerApproverID = intake.ApproverID
	} else {
		app.Status = "pending"
	}

	if err := s.repo.CreateApplication(ctx, app); err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, httpx.Conflict("already_applied", "you have already applied to this intake")
		}
		return nil, err
	}
	return &app, nil
}

func (s *Service) GetApplication(ctx context.Context, id uuid.UUID) (*Application, error) {
	return s.repo.GetApplication(ctx, id)
}

func (s *Service) ListApplicationsByIntake(ctx context.Context, intakeID uuid.UUID) ([]Application, error) {
	return s.repo.ListApplicationsByIntake(ctx, intakeID)
}

func (s *Service) ListMyApplications(ctx context.Context, principal platformauth.Principal) ([]Application, error) {
	return s.repo.ListMyApplications(ctx, principal.UserID)
}

func (s *Service) ListPendingManagerApprovals(ctx context.Context, principal platformauth.Principal) ([]Application, error) {
	return s.repo.ListPendingManagerApprovals(ctx, principal.UserID)
}

func (s *Service) ApproveByManager(ctx context.Context, principal platformauth.Principal, appID uuid.UUID, req ApproveRejectRequest) (*Application, error) {
	app, err := s.repo.GetApplication(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, httpx.NotFound("not_found", "application not found")
	}
	if app.Status != "pending_manager" {
		return nil, httpx.BadRequest("invalid_status", "application is not pending manager approval")
	}
	if app.ManagerApproverID == nil || *app.ManagerApproverID != principal.UserID {
		return nil, httpx.Forbidden("not_approver", "you are not the designated manager approver")
	}

	now := s.clock.Now()
	app.Status = "approved_by_manager"
	app.ManagerComment = req.Comment
	app.ManagerDecidedAt = &now
	app.UpdatedAt = now

	if err := s.repo.UpdateApplication(ctx, *app); err != nil {
		return nil, err
	}
	return app, nil
}

func (s *Service) RejectByManager(ctx context.Context, principal platformauth.Principal, appID uuid.UUID, req ApproveRejectRequest) (*Application, error) {
	app, err := s.repo.GetApplication(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, httpx.NotFound("not_found", "application not found")
	}
	if app.Status != "pending_manager" {
		return nil, httpx.BadRequest("invalid_status", "application is not pending manager approval")
	}
	if app.ManagerApproverID == nil || *app.ManagerApproverID != principal.UserID {
		return nil, httpx.Forbidden("not_approver", "you are not the designated manager approver")
	}

	now := s.clock.Now()
	app.Status = "rejected_by_manager"
	app.ManagerComment = req.Comment
	app.ManagerDecidedAt = &now
	app.UpdatedAt = now

	if err := s.repo.UpdateApplication(ctx, *app); err != nil {
		return nil, err
	}
	return app, nil
}

func (s *Service) ApproveByHR(ctx context.Context, principal platformauth.Principal, appID uuid.UUID, req ApproveRejectRequest) (*Application, error) {
	app, err := s.repo.GetApplication(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, httpx.NotFound("not_found", "application not found")
	}

	// HR can approve if status is "pending" (no manager) or "approved_by_manager"
	if app.Status != "pending" && app.Status != "approved_by_manager" {
		return nil, httpx.BadRequest("invalid_status", "application is not pending HR approval")
	}

	now := s.clock.Now()
	app.Status = "approved"
	app.HRApproverID = &principal.UserID
	app.HRComment = req.Comment
	app.HRDecidedAt = &now
	app.UpdatedAt = now

	if err := s.repo.UpdateApplication(ctx, *app); err != nil {
		return nil, err
	}
	return app, nil
}

func (s *Service) RejectByHR(ctx context.Context, principal platformauth.Principal, appID uuid.UUID, req ApproveRejectRequest) (*Application, error) {
	app, err := s.repo.GetApplication(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, httpx.NotFound("not_found", "application not found")
	}
	if app.Status != "pending" && app.Status != "approved_by_manager" {
		return nil, httpx.BadRequest("invalid_status", "application is not pending HR approval")
	}

	now := s.clock.Now()
	app.Status = "rejected_by_hr"
	app.HRApproverID = &principal.UserID
	app.HRComment = req.Comment
	app.HRDecidedAt = &now
	app.UpdatedAt = now

	if err := s.repo.UpdateApplication(ctx, *app); err != nil {
		return nil, err
	}
	return app, nil
}

func (s *Service) WithdrawApplication(ctx context.Context, principal platformauth.Principal, appID uuid.UUID) (*Application, error) {
	app, err := s.repo.GetApplication(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, httpx.NotFound("not_found", "application not found")
	}
	if app.ApplicantID != principal.UserID {
		return nil, httpx.Forbidden("not_owner", "you can only withdraw your own application")
	}
	if app.Status == "enrolled" || app.Status == "withdrawn" {
		return nil, httpx.BadRequest("invalid_status", "cannot withdraw application in current status")
	}

	app.Status = "withdrawn"
	app.UpdatedAt = s.clock.Now()

	if err := s.repo.UpdateApplication(ctx, *app); err != nil {
		return nil, err
	}
	return app, nil
}

func (s *Service) EnrollApplication(ctx context.Context, principal platformauth.Principal, appID uuid.UUID) (*Application, error) {
	app, err := s.repo.GetApplication(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, httpx.NotFound("not_found", "application not found")
	}
	if app.Status != "approved" {
		return nil, httpx.BadRequest("invalid_status", "application must be approved before enrollment")
	}

	app.Status = "enrolled"
	app.UpdatedAt = s.clock.Now()

	if err := s.repo.UpdateApplication(ctx, *app); err != nil {
		return nil, err
	}
	return app, nil
}

// --- Suggestions ---

func (s *Service) CreateSuggestion(ctx context.Context, principal platformauth.Principal, req CreateSuggestionRequest) (*Suggestion, error) {
	now := s.clock.Now()
	currency := "RUB"
	if req.PriceCurrency != nil {
		currency = *req.PriceCurrency
	}

	sug := Suggestion{
		ID:            uuid.New(),
		SuggestedBy:   principal.UserID,
		Title:         req.Title,
		Description:   req.Description,
		ExternalURL:   req.ExternalURL,
		ProviderName:  req.ProviderName,
		Price:         req.Price,
		PriceCurrency: currency,
		DurationHours: req.DurationHours,
		ApproverID:    req.ApproverID,
		Status:        "pending",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.CreateSuggestion(ctx, sug); err != nil {
		return nil, err
	}
	return &sug, nil
}

func (s *Service) GetSuggestion(ctx context.Context, id uuid.UUID) (*Suggestion, error) {
	return s.repo.GetSuggestion(ctx, id)
}

func (s *Service) ListSuggestions(ctx context.Context, status string) ([]Suggestion, error) {
	return s.repo.ListSuggestions(ctx, status)
}

func (s *Service) ListMySuggestions(ctx context.Context, principal platformauth.Principal) ([]Suggestion, error) {
	return s.repo.ListMySuggestions(ctx, principal.UserID)
}

func (s *Service) ApproveSuggestion(ctx context.Context, principal platformauth.Principal, id uuid.UUID, req ReviewSuggestionRequest) (*Suggestion, error) {
	sug, err := s.repo.GetSuggestion(ctx, id)
	if err != nil {
		return nil, err
	}
	if sug == nil {
		return nil, httpx.NotFound("not_found", "suggestion not found")
	}
	if sug.Status != "pending" {
		return nil, httpx.BadRequest("invalid_status", "suggestion is not pending review")
	}

	now := s.clock.Now()
	sug.Status = "approved"
	sug.ReviewedBy = &principal.UserID
	sug.ReviewComment = req.Comment
	sug.ReviewedAt = &now
	sug.UpdatedAt = now

	if err := s.repo.UpdateSuggestion(ctx, *sug); err != nil {
		return nil, err
	}
	return sug, nil
}

func (s *Service) RejectSuggestion(ctx context.Context, principal platformauth.Principal, id uuid.UUID, req ReviewSuggestionRequest) (*Suggestion, error) {
	sug, err := s.repo.GetSuggestion(ctx, id)
	if err != nil {
		return nil, err
	}
	if sug == nil {
		return nil, httpx.NotFound("not_found", "suggestion not found")
	}
	if sug.Status != "pending" {
		return nil, httpx.BadRequest("invalid_status", "suggestion is not pending review")
	}

	now := s.clock.Now()
	sug.Status = "rejected"
	sug.ReviewedBy = &principal.UserID
	sug.ReviewComment = req.Comment
	sug.ReviewedAt = &now
	sug.UpdatedAt = now

	if err := s.repo.UpdateSuggestion(ctx, *sug); err != nil {
		return nil, err
	}
	return sug, nil
}

// OpenIntakeFromSuggestion approves the suggestion and creates a new intake from it.
func (s *Service) OpenIntakeFromSuggestion(ctx context.Context, principal platformauth.Principal, sugID uuid.UUID, intakeReq CreateIntakeRequest) (*Suggestion, *Intake, error) {
	sug, err := s.repo.GetSuggestion(ctx, sugID)
	if err != nil {
		return nil, nil, err
	}
	if sug == nil {
		return nil, nil, httpx.NotFound("not_found", "suggestion not found")
	}
	if sug.Status != "pending" && sug.Status != "approved" {
		return nil, nil, httpx.BadRequest("invalid_status", "suggestion must be pending or approved")
	}

	var intake *Intake
	err = db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		now := s.clock.Now()

		// Use suggestion title if intake title not provided
		title := intakeReq.Title
		if title == "" {
			title = sug.Title
		}

		i := Intake{
			ID:                  uuid.New(),
			CourseID:            intakeReq.CourseID,
			Title:               title,
			Description:         intakeReq.Description,
			OpenedBy:            principal.UserID,
			ApproverID:          intakeReq.ApproverID,
			MaxParticipants:     intakeReq.MaxParticipants,
			StartDate:           intakeReq.StartDate,
			EndDate:             intakeReq.EndDate,
			ApplicationDeadline: intakeReq.ApplicationDeadline,
			Status:              "open",
			CreatedAt:           now,
			UpdatedAt:           now,
		}

		if err := s.repo.CreateIntake(ctx, i, tx); err != nil {
			return err
		}

		sug.Status = "intake_opened"
		sug.ReviewedBy = &principal.UserID
		sug.ReviewedAt = &now
		sug.IntakeID = &i.ID
		sug.UpdatedAt = now

		if err := s.repo.UpdateSuggestion(ctx, *sug, tx); err != nil {
			return err
		}

		intake = &i
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return sug, intake, nil
}

// ---------------------------------------------------------------------------
// Handler
// ---------------------------------------------------------------------------

type Handler struct {
	service  *Service
	validate *validator.Validate
}

func NewHandler(service *Service, validate *validator.Validate) *Handler {
	return &Handler{service: service, validate: validate}
}

func (h *Handler) principalOrError(w http.ResponseWriter, r *http.Request) (platformauth.Principal, bool) {
	p, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authentication required"))
		return p, false
	}
	return p, true
}

func (h *Handler) uuidParam(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	raw := chi.URLParam(r, name)
	id, err := uuid.Parse(raw)
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_id", "invalid "+name))
		return uuid.Nil, false
	}
	return id, true
}

// --- Intake handlers ---

func (h *Handler) CreateIntake(w http.ResponseWriter, r *http.Request) {
	p, ok := h.principalOrError(w, r)
	if !ok {
		return
	}
	var req CreateIntakeRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}

	intake, err := h.service.CreateIntake(r.Context(), p, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, intake)
}

func (h *Handler) GetIntake(w http.ResponseWriter, r *http.Request) {
	id, ok := h.uuidParam(w, r, "id")
	if !ok {
		return
	}
	intake, err := h.service.GetIntake(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if intake == nil {
		httpx.WriteError(w, httpx.NotFound("not_found", "intake not found"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, intake)
}

func (h *Handler) ListIntakes(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	list, err := h.service.ListIntakes(r.Context(), status)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, list)
}

func (h *Handler) UpdateIntake(w http.ResponseWriter, r *http.Request) {
	p, ok := h.principalOrError(w, r)
	if !ok {
		return
	}
	id, ok := h.uuidParam(w, r, "id")
	if !ok {
		return
	}
	var req UpdateIntakeRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}

	intake, err := h.service.UpdateIntake(r.Context(), p, id, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, intake)
}

func (h *Handler) CloseIntake(w http.ResponseWriter, r *http.Request) {
	id, ok := h.uuidParam(w, r, "id")
	if !ok {
		return
	}
	intake, err := h.service.CloseIntake(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, intake)
}

// --- Application handlers ---

func (h *Handler) Apply(w http.ResponseWriter, r *http.Request) {
	p, ok := h.principalOrError(w, r)
	if !ok {
		return
	}
	var req ApplyRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}

	app, err := h.service.Apply(r.Context(), p, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, app)
}

func (h *Handler) GetApplication(w http.ResponseWriter, r *http.Request) {
	id, ok := h.uuidParam(w, r, "id")
	if !ok {
		return
	}
	app, err := h.service.GetApplication(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if app == nil {
		httpx.WriteError(w, httpx.NotFound("not_found", "application not found"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, app)
}

func (h *Handler) ListApplicationsByIntake(w http.ResponseWriter, r *http.Request) {
	id, ok := h.uuidParam(w, r, "intakeId")
	if !ok {
		return
	}
	list, err := h.service.ListApplicationsByIntake(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, list)
}

func (h *Handler) ListMyApplications(w http.ResponseWriter, r *http.Request) {
	p, ok := h.principalOrError(w, r)
	if !ok {
		return
	}
	list, err := h.service.ListMyApplications(r.Context(), p)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, list)
}

func (h *Handler) ListPendingManagerApprovals(w http.ResponseWriter, r *http.Request) {
	p, ok := h.principalOrError(w, r)
	if !ok {
		return
	}
	list, err := h.service.ListPendingManagerApprovals(r.Context(), p)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, list)
}

func (h *Handler) ApproveByManager(w http.ResponseWriter, r *http.Request) {
	p, ok := h.principalOrError(w, r)
	if !ok {
		return
	}
	id, ok := h.uuidParam(w, r, "id")
	if !ok {
		return
	}
	var req ApproveRejectRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	app, err := h.service.ApproveByManager(r.Context(), p, id, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, app)
}

func (h *Handler) RejectByManager(w http.ResponseWriter, r *http.Request) {
	p, ok := h.principalOrError(w, r)
	if !ok {
		return
	}
	id, ok := h.uuidParam(w, r, "id")
	if !ok {
		return
	}
	var req ApproveRejectRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	app, err := h.service.RejectByManager(r.Context(), p, id, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, app)
}

func (h *Handler) ApproveByHR(w http.ResponseWriter, r *http.Request) {
	p, ok := h.principalOrError(w, r)
	if !ok {
		return
	}
	id, ok := h.uuidParam(w, r, "id")
	if !ok {
		return
	}
	var req ApproveRejectRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	app, err := h.service.ApproveByHR(r.Context(), p, id, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, app)
}

func (h *Handler) RejectByHR(w http.ResponseWriter, r *http.Request) {
	p, ok := h.principalOrError(w, r)
	if !ok {
		return
	}
	id, ok := h.uuidParam(w, r, "id")
	if !ok {
		return
	}
	var req ApproveRejectRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	app, err := h.service.RejectByHR(r.Context(), p, id, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, app)
}

func (h *Handler) WithdrawApplication(w http.ResponseWriter, r *http.Request) {
	p, ok := h.principalOrError(w, r)
	if !ok {
		return
	}
	id, ok := h.uuidParam(w, r, "id")
	if !ok {
		return
	}

	app, err := h.service.WithdrawApplication(r.Context(), p, id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, app)
}

func (h *Handler) EnrollApplication(w http.ResponseWriter, r *http.Request) {
	p, ok := h.principalOrError(w, r)
	if !ok {
		return
	}
	id, ok := h.uuidParam(w, r, "id")
	if !ok {
		return
	}

	app, err := h.service.EnrollApplication(r.Context(), p, id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, app)
}

// --- Suggestion handlers ---

func (h *Handler) CreateSuggestion(w http.ResponseWriter, r *http.Request) {
	p, ok := h.principalOrError(w, r)
	if !ok {
		return
	}
	var req CreateSuggestionRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}

	sug, err := h.service.CreateSuggestion(r.Context(), p, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, sug)
}

func (h *Handler) GetSuggestion(w http.ResponseWriter, r *http.Request) {
	id, ok := h.uuidParam(w, r, "id")
	if !ok {
		return
	}
	sug, err := h.service.GetSuggestion(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if sug == nil {
		httpx.WriteError(w, httpx.NotFound("not_found", "suggestion not found"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, sug)
}

func (h *Handler) ListSuggestions(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	list, err := h.service.ListSuggestions(r.Context(), status)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, list)
}

func (h *Handler) ListMySuggestions(w http.ResponseWriter, r *http.Request) {
	p, ok := h.principalOrError(w, r)
	if !ok {
		return
	}
	list, err := h.service.ListMySuggestions(r.Context(), p)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, list)
}

func (h *Handler) ApproveSuggestion(w http.ResponseWriter, r *http.Request) {
	p, ok := h.principalOrError(w, r)
	if !ok {
		return
	}
	id, ok := h.uuidParam(w, r, "id")
	if !ok {
		return
	}
	var req ReviewSuggestionRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	sug, err := h.service.ApproveSuggestion(r.Context(), p, id, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, sug)
}

func (h *Handler) RejectSuggestion(w http.ResponseWriter, r *http.Request) {
	p, ok := h.principalOrError(w, r)
	if !ok {
		return
	}
	id, ok := h.uuidParam(w, r, "id")
	if !ok {
		return
	}
	var req ReviewSuggestionRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	sug, err := h.service.RejectSuggestion(r.Context(), p, id, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, sug)
}

func (h *Handler) OpenIntakeFromSuggestion(w http.ResponseWriter, r *http.Request) {
	p, ok := h.principalOrError(w, r)
	if !ok {
		return
	}
	id, ok := h.uuidParam(w, r, "id")
	if !ok {
		return
	}
	var req CreateIntakeRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	sug, intake, err := h.service.OpenIntakeFromSuggestion(r.Context(), p, id, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, map[string]any{
		"suggestion": sug,
		"intake":     intake,
	})
}
