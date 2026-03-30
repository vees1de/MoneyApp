package catalog

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

type Course struct {
	ID                 uuid.UUID  `json:"id"`
	Type               string     `json:"type"`
	SourceType         string     `json:"source_type"`
	Title              string     `json:"title"`
	Slug               *string    `json:"slug,omitempty"`
	ShortDescription   *string    `json:"short_description,omitempty"`
	Description        *string    `json:"description,omitempty"`
	ProviderID         *uuid.UUID `json:"provider_id,omitempty"`
	CategoryID         *uuid.UUID `json:"category_id,omitempty"`
	DirectionID        *uuid.UUID `json:"direction_id,omitempty"`
	Level              *string    `json:"level,omitempty"`
	DurationHours      *string    `json:"duration_hours,omitempty"`
	Language           *string    `json:"language,omitempty"`
	IsMandatoryDefault bool       `json:"is_mandatory_default"`
	Status             string     `json:"status"`
	ExternalURL        *string    `json:"external_url,omitempty"`
	Price              *string    `json:"price,omitempty"`
	PriceCurrency      *string    `json:"price_currency,omitempty"`
	NextStartDate      *time.Time `json:"next_start_date,omitempty"`
	ThumbnailFileID    *uuid.UUID `json:"thumbnail_file_id,omitempty"`
	CreatedBy          *uuid.UUID `json:"created_by,omitempty"`
	UpdatedBy          *uuid.UUID `json:"updated_by,omitempty"`
	PublishedAt        *time.Time `json:"published_at,omitempty"`
	ArchivedAt         *time.Time `json:"archived_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type CourseMaterial struct {
	ID          uuid.UUID  `json:"id"`
	CourseID    uuid.UUID  `json:"course_id"`
	Type        string     `json:"type"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	FileID      *uuid.UUID `json:"file_id,omitempty"`
	ExternalURL *string    `json:"external_url,omitempty"`
	SortOrder   int        `json:"sort_order"`
	IsRequired  bool       `json:"is_required"`
	CreatedAt   time.Time  `json:"created_at"`
}

type CreateCourseRequest struct {
	Type               string     `json:"type" validate:"required,oneof=internal external"`
	SourceType         string     `json:"source_type" validate:"required,oneof=catalog requested imported"`
	Title              string     `json:"title" validate:"required"`
	Slug               *string    `json:"slug,omitempty"`
	ShortDescription   *string    `json:"short_description,omitempty"`
	Description        *string    `json:"description,omitempty"`
	ProviderID         *uuid.UUID `json:"provider_id,omitempty"`
	CategoryID         *uuid.UUID `json:"category_id,omitempty"`
	DirectionID        *uuid.UUID `json:"direction_id,omitempty"`
	Level              *string    `json:"level,omitempty"`
	DurationHours      *string    `json:"duration_hours,omitempty"`
	Language           *string    `json:"language,omitempty"`
	IsMandatoryDefault bool       `json:"is_mandatory_default"`
	ExternalURL        *string    `json:"external_url,omitempty"`
	Price              *string    `json:"price,omitempty"`
	PriceCurrency      *string    `json:"price_currency,omitempty"`
	NextStartDate      *time.Time `json:"next_start_date,omitempty"`
	ThumbnailFileID    *uuid.UUID `json:"thumbnail_file_id,omitempty"`
}

type UpdateCourseRequest struct {
	Title              *string    `json:"title,omitempty"`
	Slug               *string    `json:"slug,omitempty"`
	ShortDescription   *string    `json:"short_description,omitempty"`
	Description        *string    `json:"description,omitempty"`
	ProviderID         *uuid.UUID `json:"provider_id,omitempty"`
	CategoryID         *uuid.UUID `json:"category_id,omitempty"`
	DirectionID        *uuid.UUID `json:"direction_id,omitempty"`
	Level              *string    `json:"level,omitempty"`
	DurationHours      *string    `json:"duration_hours,omitempty"`
	Language           *string    `json:"language,omitempty"`
	IsMandatoryDefault *bool      `json:"is_mandatory_default,omitempty"`
	ExternalURL        *string    `json:"external_url,omitempty"`
	Price              *string    `json:"price,omitempty"`
	PriceCurrency      *string    `json:"price_currency,omitempty"`
	NextStartDate      *time.Time `json:"next_start_date,omitempty"`
	ThumbnailFileID    *uuid.UUID `json:"thumbnail_file_id,omitempty"`
}

type CreateMaterialRequest struct {
	Type        string     `json:"type" validate:"required,oneof=file link video scorm pdf"`
	Title       string     `json:"title" validate:"required"`
	Description *string    `json:"description,omitempty"`
	FileID      *uuid.UUID `json:"file_id,omitempty"`
	ExternalURL *string    `json:"external_url,omitempty"`
	SortOrder   int        `json:"sort_order"`
	IsRequired  bool       `json:"is_required"`
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

func (r *Repository) CreateCourse(ctx context.Context, item Course, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into courses (
			id, type, source_type, title, slug, short_description, description,
			provider_id, category_id, direction_id, level, duration_hours, language,
			is_mandatory_default, status, external_url, price, price_currency, next_start_date,
			thumbnail_file_id, created_by, updated_by,
			published_at, archived_at, created_at, updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, nullif($12, '')::numeric, $13,
		        $14, $15, $16, nullif($17, '')::numeric, $18, $19::date,
		        $20, $21, $22, $23, $24, $25, $26)
	`, item.ID, item.Type, item.SourceType, item.Title, item.Slug, item.ShortDescription, item.Description,
		item.ProviderID, item.CategoryID, item.DirectionID, item.Level, item.DurationHours, item.Language,
		item.IsMandatoryDefault, item.Status, item.ExternalURL, item.Price, item.PriceCurrency, item.NextStartDate,
		item.ThumbnailFileID, item.CreatedBy, item.UpdatedBy,
		item.PublishedAt, item.ArchivedAt, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *Repository) GetCourse(ctx context.Context, id uuid.UUID, exec ...db.DBTX) (Course, error) {
	var item Course
	err := r.base(exec...).QueryRowContext(ctx, `
		select id, type, source_type, title, slug, short_description, description,
		       provider_id, category_id, direction_id, level, duration_hours::text, language,
		       is_mandatory_default, status, external_url, price::text, price_currency, next_start_date,
		       thumbnail_file_id, created_by, updated_by,
		       published_at, archived_at, created_at, updated_at
		from courses
		where id = $1
	`, id).Scan(
		&item.ID, &item.Type, &item.SourceType, &item.Title, &item.Slug, &item.ShortDescription, &item.Description,
		&item.ProviderID, &item.CategoryID, &item.DirectionID, &item.Level, &item.DurationHours, &item.Language,
		&item.IsMandatoryDefault, &item.Status, &item.ExternalURL, &item.Price, &item.PriceCurrency, &item.NextStartDate,
		&item.ThumbnailFileID, &item.CreatedBy, &item.UpdatedBy,
		&item.PublishedAt, &item.ArchivedAt, &item.CreatedAt, &item.UpdatedAt,
	)
	return item, err
}

func (r *Repository) ListCourses(ctx context.Context, exec ...db.DBTX) ([]Course, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select id, type, source_type, title, slug, short_description, description,
		       provider_id, category_id, direction_id, level, duration_hours::text, language,
		       is_mandatory_default, status, external_url, price::text, price_currency, next_start_date,
		       thumbnail_file_id, created_by, updated_by,
		       published_at, archived_at, created_at, updated_at
		from courses
		order by created_at desc
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Course
	for rows.Next() {
		var item Course
		if err := rows.Scan(
			&item.ID, &item.Type, &item.SourceType, &item.Title, &item.Slug, &item.ShortDescription, &item.Description,
			&item.ProviderID, &item.CategoryID, &item.DirectionID, &item.Level, &item.DurationHours, &item.Language,
			&item.IsMandatoryDefault, &item.Status, &item.ExternalURL, &item.Price, &item.PriceCurrency, &item.NextStartDate,
			&item.ThumbnailFileID, &item.CreatedBy, &item.UpdatedBy,
			&item.PublishedAt, &item.ArchivedAt, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) UpdateCourse(ctx context.Context, item Course, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update courses
		set title = $2,
		    slug = $3,
		    short_description = $4,
		    description = $5,
		    provider_id = $6,
		    category_id = $7,
		    direction_id = $8,
		    level = $9,
		    duration_hours = nullif($10, '')::numeric,
		    language = $11,
		    is_mandatory_default = $12,
		    status = $13,
		    external_url = $14,
		    price = nullif($15, '')::numeric,
		    price_currency = $16,
		    next_start_date = $17::date,
		    thumbnail_file_id = $18,
		    updated_by = $19,
		    published_at = $20,
		    archived_at = $21,
		    updated_at = $22
		where id = $1
	`, item.ID, item.Title, item.Slug, item.ShortDescription, item.Description, item.ProviderID, item.CategoryID,
		item.DirectionID, item.Level, item.DurationHours, item.Language, item.IsMandatoryDefault, item.Status,
		item.ExternalURL, item.Price, item.PriceCurrency, item.NextStartDate,
		item.ThumbnailFileID, item.UpdatedBy, item.PublishedAt, item.ArchivedAt, item.UpdatedAt)
	return err
}

func (r *Repository) CreateMaterial(ctx context.Context, item CourseMaterial, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into course_materials (
			id, course_id, type, title, description, file_id, external_url, sort_order, is_required, created_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, item.ID, item.CourseID, item.Type, item.Title, item.Description, item.FileID, item.ExternalURL, item.SortOrder, item.IsRequired, item.CreatedAt)
	return err
}

func (r *Repository) ListMaterials(ctx context.Context, courseID uuid.UUID, exec ...db.DBTX) ([]CourseMaterial, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select id, course_id, type, title, description, file_id, external_url, sort_order, is_required, created_at
		from course_materials
		where course_id = $1
		order by sort_order asc, created_at asc
	`, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []CourseMaterial
	for rows.Next() {
		var item CourseMaterial
		if err := rows.Scan(&item.ID, &item.CourseID, &item.Type, &item.Title, &item.Description, &item.FileID, &item.ExternalURL, &item.SortOrder, &item.IsRequired, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

type Service struct {
	repo  *Repository
	clock clock.Clock
}

func NewService(repo *Repository, appClock clock.Clock) *Service {
	return &Service{repo: repo, clock: appClock}
}

func (s *Service) ListCourses(ctx context.Context, principal platformauth.Principal, filters CourseListFilters) ([]Course, error) {
	return s.repo.ListCoursesFiltered(ctx, filters)
}

func (s *Service) CreateCourse(ctx context.Context, principal platformauth.Principal, req CreateCourseRequest) (Course, error) {
	now := s.clock.Now()
	item := Course{
		ID:                 uuid.New(),
		Type:               req.Type,
		SourceType:         req.SourceType,
		Title:              req.Title,
		Slug:               req.Slug,
		ShortDescription:   req.ShortDescription,
		Description:        req.Description,
		ProviderID:         req.ProviderID,
		CategoryID:         req.CategoryID,
		DirectionID:        req.DirectionID,
		Level:              req.Level,
		DurationHours:      req.DurationHours,
		Language:           req.Language,
		IsMandatoryDefault: req.IsMandatoryDefault,
		Status:             "draft",
		ExternalURL:        req.ExternalURL,
		Price:              req.Price,
		PriceCurrency:      req.PriceCurrency,
		NextStartDate:      req.NextStartDate,
		ThumbnailFileID:    req.ThumbnailFileID,
		CreatedBy:          &principal.UserID,
		UpdatedBy:          &principal.UserID,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	return item, s.repo.CreateCourse(ctx, item)
}

func (s *Service) GetCourse(ctx context.Context, id uuid.UUID) (Course, error) {
	item, err := s.repo.GetCourse(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Course{}, httpx.NotFound("course_not_found", "course not found")
		}
		return Course{}, err
	}
	return item, nil
}

func (s *Service) UpdateCourse(ctx context.Context, principal platformauth.Principal, id uuid.UUID, req UpdateCourseRequest) (Course, error) {
	item, err := s.GetCourse(ctx, id)
	if err != nil {
		return Course{}, err
	}

	if req.Title != nil {
		item.Title = *req.Title
	}
	if req.Slug != nil {
		item.Slug = req.Slug
	}
	if req.ShortDescription != nil {
		item.ShortDescription = req.ShortDescription
	}
	if req.Description != nil {
		item.Description = req.Description
	}
	if req.ProviderID != nil {
		item.ProviderID = req.ProviderID
	}
	if req.CategoryID != nil {
		item.CategoryID = req.CategoryID
	}
	if req.DirectionID != nil {
		item.DirectionID = req.DirectionID
	}
	if req.Level != nil {
		item.Level = req.Level
	}
	if req.DurationHours != nil {
		item.DurationHours = req.DurationHours
	}
	if req.Language != nil {
		item.Language = req.Language
	}
	if req.IsMandatoryDefault != nil {
		item.IsMandatoryDefault = *req.IsMandatoryDefault
	}
	if req.ExternalURL != nil {
		item.ExternalURL = req.ExternalURL
	}
	if req.Price != nil {
		item.Price = req.Price
	}
	if req.PriceCurrency != nil {
		item.PriceCurrency = req.PriceCurrency
	}
	if req.NextStartDate != nil {
		item.NextStartDate = req.NextStartDate
	}
	if req.ThumbnailFileID != nil {
		item.ThumbnailFileID = req.ThumbnailFileID
	}
	item.UpdatedBy = &principal.UserID
	item.UpdatedAt = s.clock.Now()

	return item, s.repo.UpdateCourse(ctx, item)
}

func (s *Service) PublishCourse(ctx context.Context, principal platformauth.Principal, id uuid.UUID) (Course, error) {
	item, err := s.GetCourse(ctx, id)
	if err != nil {
		return Course{}, err
	}
	now := s.clock.Now()
	item.Status = "published"
	item.PublishedAt = &now
	item.ArchivedAt = nil
	item.UpdatedBy = &principal.UserID
	item.UpdatedAt = now
	return item, s.repo.UpdateCourse(ctx, item)
}

func (s *Service) ArchiveCourse(ctx context.Context, principal platformauth.Principal, id uuid.UUID) (Course, error) {
	item, err := s.GetCourse(ctx, id)
	if err != nil {
		return Course{}, err
	}
	now := s.clock.Now()
	item.Status = "archived"
	item.ArchivedAt = &now
	item.UpdatedBy = &principal.UserID
	item.UpdatedAt = now
	return item, s.repo.UpdateCourse(ctx, item)
}

func (s *Service) ListMaterials(ctx context.Context, courseID uuid.UUID) ([]CourseMaterial, error) {
	return s.repo.ListMaterials(ctx, courseID)
}

func (s *Service) CreateMaterial(ctx context.Context, courseID uuid.UUID, req CreateMaterialRequest) (CourseMaterial, error) {
	if _, err := s.GetCourse(ctx, courseID); err != nil {
		return CourseMaterial{}, err
	}

	item := CourseMaterial{
		ID:          uuid.New(),
		CourseID:    courseID,
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		FileID:      req.FileID,
		ExternalURL: req.ExternalURL,
		SortOrder:   req.SortOrder,
		IsRequired:  req.IsRequired,
		CreatedAt:   s.clock.Now(),
	}
	return item, s.repo.CreateMaterial(ctx, item)
}

type Handler struct {
	service   *Service
	validator *validator.Validate
}

func NewHandler(service *Service, validator *validator.Validate) *Handler {
	return &Handler{service: service, validator: validator}
}

func principalFromContext(r *http.Request) (platformauth.Principal, error) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		return platformauth.Principal{}, httpx.Unauthorized("unauthorized", "authorization required")
	}
	return principal, nil
}

func (h *Handler) ListCourses(w http.ResponseWriter, r *http.Request) {
	principal, err := principalFromContext(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	filters, err := parseCourseListFilters(r, principal)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	items, err := h.service.ListCourses(r.Context(), principal, filters)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) CreateCourse(w http.ResponseWriter, r *http.Request) {
	principal, err := principalFromContext(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	var req CreateCourseRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}

	item, err := h.service.CreateCourse(r.Context(), principal, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) GetCourse(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_course_id", "invalid course id"))
		return
	}
	item, err := h.service.GetCourse(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) UpdateCourse(w http.ResponseWriter, r *http.Request) {
	principal, err := principalFromContext(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_course_id", "invalid course id"))
		return
	}
	var req UpdateCourseRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	item, err := h.service.UpdateCourse(r.Context(), principal, id, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) PublishCourse(w http.ResponseWriter, r *http.Request) {
	principal, err := principalFromContext(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_course_id", "invalid course id"))
		return
	}
	item, err := h.service.PublishCourse(r.Context(), principal, id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) ArchiveCourse(w http.ResponseWriter, r *http.Request) {
	principal, err := principalFromContext(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_course_id", "invalid course id"))
		return
	}
	item, err := h.service.ArchiveCourse(r.Context(), principal, id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) ListMaterials(w http.ResponseWriter, r *http.Request) {
	courseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_course_id", "invalid course id"))
		return
	}
	items, err := h.service.ListMaterials(r.Context(), courseID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) CreateMaterial(w http.ResponseWriter, r *http.Request) {
	courseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_course_id", "invalid course id"))
		return
	}
	var req CreateMaterialRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}
	item, err := h.service.CreateMaterial(r.Context(), courseID, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}
