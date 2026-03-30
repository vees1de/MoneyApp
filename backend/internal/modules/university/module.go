package university

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/clock"
	"moneyapp/backend/internal/platform/db"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Program struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	DirectionID *uuid.UUID `json:"direction_id,omitempty"`
	Status      string     `json:"status"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type TrainingGroup struct {
	ID                uuid.UUID  `json:"id"`
	ProgramID         uuid.UUID  `json:"program_id"`
	Name              string     `json:"name"`
	Capacity          *int       `json:"capacity,omitempty"`
	Status            string     `json:"status"`
	EnrollmentOpenAt  *time.Time `json:"enrollment_open_at,omitempty"`
	EnrollmentCloseAt *time.Time `json:"enrollment_close_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type Session struct {
	ID              uuid.UUID  `json:"id"`
	GroupID         uuid.UUID  `json:"group_id"`
	TrainerUserID   uuid.UUID  `json:"trainer_user_id"`
	Title           string     `json:"title"`
	Description     *string    `json:"description,omitempty"`
	StartAt         time.Time  `json:"start_at"`
	EndAt           time.Time  `json:"end_at"`
	Location        *string    `json:"location,omitempty"`
	MeetingURL      *string    `json:"meeting_url,omitempty"`
	Status          string     `json:"status"`
	CalendarEventID *uuid.UUID `json:"calendar_event_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CreateProgramRequest struct {
	Title       string     `json:"title" validate:"required"`
	Description *string    `json:"description,omitempty"`
	DirectionID *uuid.UUID `json:"direction_id,omitempty"`
	Status      string     `json:"status" validate:"omitempty,oneof=draft published archived"`
}

type CreateGroupRequest struct {
	Name              string     `json:"name" validate:"required"`
	Capacity          *int       `json:"capacity,omitempty"`
	Status            string     `json:"status" validate:"omitempty,oneof=planned open full in_progress completed canceled"`
	EnrollmentOpenAt  *time.Time `json:"enrollment_open_at,omitempty"`
	EnrollmentCloseAt *time.Time `json:"enrollment_close_at,omitempty"`
}

type AddParticipantRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	Status string    `json:"status" validate:"omitempty,oneof=enrolled waitlisted attended completed canceled"`
}

type CreateSessionRequest struct {
	TrainerUserID uuid.UUID `json:"trainer_user_id" validate:"required"`
	Title         string    `json:"title" validate:"required"`
	Description   *string   `json:"description,omitempty"`
	StartAt       time.Time `json:"start_at" validate:"required"`
	EndAt         time.Time `json:"end_at" validate:"required"`
	Location      *string   `json:"location,omitempty"`
	MeetingURL    *string   `json:"meeting_url,omitempty"`
	Status        string    `json:"status" validate:"omitempty,oneof=planned held canceled rescheduled"`
}

type TrainerFeedbackRequest struct {
	ParticipantUserID uuid.UUID `json:"participant_user_id" validate:"required"`
	AttendanceStatus  string    `json:"attendance_status" validate:"required,oneof=attended absent excused"`
	Score             *string   `json:"score,omitempty"`
	Comment           *string   `json:"comment,omitempty"`
}

type ParticipantFeedbackRequest struct {
	ProgramID *uuid.UUID `json:"program_id,omitempty"`
	Rating    int        `json:"rating" validate:"required,min=1,max=5"`
	Comment   *string    `json:"comment,omitempty"`
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

func (r *Repository) CreateProgram(ctx context.Context, item Program, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into internal_programs (id, title, description, direction_id, status, created_by, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8)
	`, item.ID, item.Title, item.Description, item.DirectionID, item.Status, item.CreatedBy, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *Repository) ListPrograms(ctx context.Context, exec ...db.DBTX) ([]Program, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select id, title, description, direction_id, status, created_by, created_at, updated_at
		from internal_programs
		order by created_at desc
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Program
	for rows.Next() {
		var item Program
		if err := rows.Scan(&item.ID, &item.Title, &item.Description, &item.DirectionID, &item.Status, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) GetProgram(ctx context.Context, id uuid.UUID, exec ...db.DBTX) (Program, error) {
	var item Program
	err := r.base(exec...).QueryRowContext(ctx, `
		select id, title, description, direction_id, status, created_by, created_at, updated_at
		from internal_programs
		where id = $1
	`, id).Scan(&item.ID, &item.Title, &item.Description, &item.DirectionID, &item.Status, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *Repository) CreateGroup(ctx context.Context, item TrainingGroup, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into training_groups (
			id, program_id, name, capacity, status, enrollment_open_at, enrollment_close_at, created_at, updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, item.ID, item.ProgramID, item.Name, item.Capacity, item.Status, item.EnrollmentOpenAt, item.EnrollmentCloseAt, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *Repository) AddParticipant(ctx context.Context, groupID, userID uuid.UUID, status string, enrolledAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into group_participants (id, group_id, user_id, status, enrolled_at, completed_at)
		values ($1, $2, $3, $4, $5, null)
		on conflict (group_id, user_id) do update
		set status = excluded.status
	`, uuid.New(), groupID, userID, status, enrolledAt)
	return err
}

func (r *Repository) CreateSession(ctx context.Context, item Session, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into sessions_university (
			id, group_id, trainer_user_id, title, description, start_at, end_at, location,
			meeting_url, status, calendar_event_id, created_at, updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, item.ID, item.GroupID, item.TrainerUserID, item.Title, item.Description, item.StartAt, item.EndAt,
		item.Location, item.MeetingURL, item.Status, item.CalendarEventID, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *Repository) CreateTrainerFeedback(ctx context.Context, sessionID, participantUserID, trainerUserID uuid.UUID, attendanceStatus string, score *string, comment *string, createdAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into trainer_feedback (
			id, session_id, participant_user_id, trainer_user_id, attendance_status, score, comment, created_at
		)
		values ($1, $2, $3, $4, $5, nullif($6, '')::numeric, $7, $8)
	`, uuid.New(), sessionID, participantUserID, trainerUserID, attendanceStatus, score, comment, createdAt)
	return err
}

func (r *Repository) CreateParticipantFeedback(ctx context.Context, sessionID uuid.UUID, programID *uuid.UUID, participantUserID uuid.UUID, rating int, comment *string, createdAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into participant_feedback (id, program_id, session_id, participant_user_id, rating, comment, created_at)
		values ($1, $2, $3, $4, $5, $6, $7)
	`, uuid.New(), programID, sessionID, participantUserID, rating, comment, createdAt)
	return err
}

type Service struct {
	repo  *Repository
	clock clock.Clock
}

func NewService(repo *Repository, appClock clock.Clock) *Service {
	return &Service{repo: repo, clock: appClock}
}

func (s *Service) CreateProgram(ctx context.Context, principal platformauth.Principal, req CreateProgramRequest) (Program, error) {
	status := req.Status
	if status == "" {
		status = "draft"
	}
	item := Program{
		ID:          uuid.New(),
		Title:       req.Title,
		Description: req.Description,
		DirectionID: req.DirectionID,
		Status:      status,
		CreatedBy:   principal.UserID,
		CreatedAt:   s.clock.Now(),
		UpdatedAt:   s.clock.Now(),
	}
	return item, s.repo.CreateProgram(ctx, item)
}

func (s *Service) ListPrograms(ctx context.Context) ([]Program, error) {
	return s.repo.ListPrograms(ctx)
}

func (s *Service) GetProgram(ctx context.Context, id uuid.UUID) (Program, error) {
	item, err := s.repo.GetProgram(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Program{}, httpx.NotFound("program_not_found", "program not found")
		}
		return Program{}, err
	}
	return item, nil
}

func (s *Service) CreateGroup(ctx context.Context, programID uuid.UUID, req CreateGroupRequest) (TrainingGroup, error) {
	status := req.Status
	if status == "" {
		status = "planned"
	}
	item := TrainingGroup{
		ID:                uuid.New(),
		ProgramID:         programID,
		Name:              req.Name,
		Capacity:          req.Capacity,
		Status:            status,
		EnrollmentOpenAt:  req.EnrollmentOpenAt,
		EnrollmentCloseAt: req.EnrollmentCloseAt,
		CreatedAt:         s.clock.Now(),
		UpdatedAt:         s.clock.Now(),
	}
	return item, s.repo.CreateGroup(ctx, item)
}

func (s *Service) AddParticipant(ctx context.Context, groupID uuid.UUID, req AddParticipantRequest) error {
	status := req.Status
	if status == "" {
		status = "enrolled"
	}
	return s.repo.AddParticipant(ctx, groupID, req.UserID, status, s.clock.Now())
}

func (s *Service) CreateSession(ctx context.Context, groupID uuid.UUID, req CreateSessionRequest) (Session, error) {
	status := req.Status
	if status == "" {
		status = "planned"
	}
	item := Session{
		ID:            uuid.New(),
		GroupID:       groupID,
		TrainerUserID: req.TrainerUserID,
		Title:         req.Title,
		Description:   req.Description,
		StartAt:       req.StartAt,
		EndAt:         req.EndAt,
		Location:      req.Location,
		MeetingURL:    req.MeetingURL,
		Status:        status,
		CreatedAt:     s.clock.Now(),
		UpdatedAt:     s.clock.Now(),
	}
	return item, s.repo.CreateSession(ctx, item)
}

func (s *Service) CreateTrainerFeedback(ctx context.Context, principal platformauth.Principal, sessionID uuid.UUID, req TrainerFeedbackRequest) error {
	return s.repo.CreateTrainerFeedback(ctx, sessionID, req.ParticipantUserID, principal.UserID, req.AttendanceStatus, req.Score, req.Comment, s.clock.Now())
}

func (s *Service) CreateParticipantFeedback(ctx context.Context, principal platformauth.Principal, sessionID uuid.UUID, req ParticipantFeedbackRequest) error {
	return s.repo.CreateParticipantFeedback(ctx, sessionID, req.ProgramID, principal.UserID, req.Rating, req.Comment, s.clock.Now())
}

type Handler struct {
	service   *Service
	validator *validator.Validate
}

func NewHandler(service *Service, validator *validator.Validate) *Handler {
	return &Handler{service: service, validator: validator}
}

func universityPrincipal(r *http.Request) (platformauth.Principal, error) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		return platformauth.Principal{}, httpx.Unauthorized("unauthorized", "authorization required")
	}
	return principal, nil
}

func (h *Handler) CreateProgram(w http.ResponseWriter, r *http.Request) {
	principal, err := universityPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	var req CreateProgramRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}
	item, err := h.service.CreateProgram(r.Context(), principal, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) ListPrograms(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListPrograms(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) GetProgram(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_program_id", "invalid program id"))
		return
	}
	item, err := h.service.GetProgram(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	programID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_program_id", "invalid program id"))
		return
	}
	var req CreateGroupRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}
	item, err := h.service.CreateGroup(r.Context(), programID, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) AddParticipant(w http.ResponseWriter, r *http.Request) {
	groupID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_group_id", "invalid group id"))
		return
	}
	var req AddParticipantRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}
	if err := h.service.AddParticipant(r.Context(), groupID, req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	groupID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_group_id", "invalid group id"))
		return
	}
	var req CreateSessionRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}
	item, err := h.service.CreateSession(r.Context(), groupID, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) TrainerFeedback(w http.ResponseWriter, r *http.Request) {
	principal, err := universityPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_session_id", "invalid session id"))
		return
	}
	var req TrainerFeedbackRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}
	if err := h.service.CreateTrainerFeedback(r.Context(), principal, sessionID, req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) ParticipantFeedback(w http.ResponseWriter, r *http.Request) {
	principal, err := universityPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_session_id", "invalid session id"))
		return
	}
	var req ParticipantFeedbackRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}
	if err := h.service.CreateParticipantFeedback(r.Context(), principal, sessionID, req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteNoContent(w)
}
