package certificates

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/clock"
	"moneyapp/backend/internal/platform/db"
	"moneyapp/backend/internal/platform/events"
	"moneyapp/backend/internal/platform/httpx"
	"moneyapp/backend/internal/platform/outbox"
	"moneyapp/backend/internal/platform/uploads"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type FileAttachment struct {
	ID              uuid.UUID  `json:"id"`
	StorageProvider string     `json:"storage_provider"`
	StorageKey      string     `json:"storage_key"`
	OriginalName    string     `json:"original_name"`
	MimeType        string     `json:"mime_type"`
	SizeBytes       int64      `json:"size_bytes"`
	UploadedBy      *uuid.UUID `json:"uploaded_by,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type Certificate struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	CourseID         *uuid.UUID `json:"course_id,omitempty"`
	EnrollmentID     *uuid.UUID `json:"enrollment_id,omitempty"`
	CertificateNo    *string    `json:"certificate_no,omitempty"`
	IssuedBy         *string    `json:"issued_by,omitempty"`
	IssuedAt         *time.Time `json:"issued_at,omitempty"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	Status           string     `json:"status"`
	FileID           uuid.UUID  `json:"file_id"`
	FileStorageKey   string     `json:"file_storage_key,omitempty"`
	FileOriginalName string     `json:"file_original_name,omitempty"`
	UploadedAt       time.Time  `json:"uploaded_at"`
	VerifiedAt       *time.Time `json:"verified_at,omitempty"`
	VerifiedBy       *uuid.UUID `json:"verified_by,omitempty"`
	Notes            *string    `json:"notes,omitempty"`
}

type UploadCertificateRequest struct {
	CourseID        *uuid.UUID `json:"course_id,omitempty"`
	EnrollmentID    *uuid.UUID `json:"enrollment_id,omitempty"`
	IssuedAt        *time.Time `json:"issued_at,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	StorageProvider string     `json:"storage_provider" validate:"required,oneof=s3 local minio"`
	StorageKey      string     `json:"storage_key" validate:"required"`
	OriginalName    string     `json:"original_name" validate:"required"`
	MimeType        string     `json:"mime_type" validate:"required"`
	SizeBytes       int64      `json:"size_bytes" validate:"required,min=1"`
	FileContent     []byte     `json:"-"`
}

const maxCertificateUploadSizeBytes = 25 << 20

type EnrollmentCompletionHook interface {
	CompleteEnrollmentAfterCertificate(
		ctx context.Context,
		enrollmentID uuid.UUID,
		actorID uuid.UUID,
		notes *string,
		completedAt time.Time,
		exec db.DBTX,
	) error
}

type ReviewRequest struct {
	Comment *string `json:"comment,omitempty"`
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

func (r *Repository) CreateFile(ctx context.Context, item FileAttachment, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into file_attachments (
			id, storage_provider, storage_key, original_name, mime_type, size_bytes, uploaded_by, created_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8)
	`, item.ID, item.StorageProvider, item.StorageKey, item.OriginalName, item.MimeType, item.SizeBytes, item.UploadedBy, item.CreatedAt)
	return err
}

func (r *Repository) CreateCertificate(ctx context.Context, item Certificate, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into certificates (
			id, user_id, course_id, enrollment_id, certificate_no, issued_by, issued_at, expires_at,
			status, file_id, uploaded_at, verified_at, verified_by, notes
		)
		values ($1, $2, $3, $4, $5, $6, $7::date, $8::date, $9, $10, $11, $12, $13, $14)
	`, item.ID, item.UserID, item.CourseID, item.EnrollmentID, item.CertificateNo, item.IssuedBy, item.IssuedAt, item.ExpiresAt, item.Status, item.FileID, item.UploadedAt, item.VerifiedAt, item.VerifiedBy, item.Notes)
	return err
}

func (r *Repository) CreateVerification(ctx context.Context, certificateID, performedBy uuid.UUID, action string, comment *string, createdAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into certificate_verifications (id, certificate_id, action, performed_by, comment, created_at)
		values ($1, $2, $3, $4, $5, $6)
	`, uuid.New(), certificateID, action, performedBy, comment, createdAt)
	return err
}

func (r *Repository) GetCertificate(ctx context.Context, id uuid.UUID, exec ...db.DBTX) (Certificate, error) {
	var item Certificate
	err := r.base(exec...).QueryRowContext(ctx, `
		select id, user_id, course_id, enrollment_id, certificate_no, issued_by, issued_at::timestamptz,
		       expires_at::timestamptz, status, file_id, uploaded_at, verified_at, verified_by, notes,
		       fa.storage_key, fa.original_name
		from certificates
		join file_attachments fa on fa.id = certificates.file_id
		where id = $1
	`, id).Scan(&item.ID, &item.UserID, &item.CourseID, &item.EnrollmentID, &item.CertificateNo, &item.IssuedBy,
		&item.IssuedAt, &item.ExpiresAt, &item.Status, &item.FileID, &item.UploadedAt, &item.VerifiedAt, &item.VerifiedBy, &item.Notes,
		&item.FileStorageKey, &item.FileOriginalName)
	return item, err
}

func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) ([]Certificate, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select id, user_id, course_id, enrollment_id, certificate_no, issued_by, issued_at::timestamptz,
		       expires_at::timestamptz, status, file_id, uploaded_at, verified_at, verified_by, notes,
		       fa.storage_key, fa.original_name
		from certificates
		join file_attachments fa on fa.id = certificates.file_id
		where user_id = $1
		order by uploaded_at desc
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Certificate
	for rows.Next() {
		var item Certificate
		if err := rows.Scan(&item.ID, &item.UserID, &item.CourseID, &item.EnrollmentID, &item.CertificateNo, &item.IssuedBy,
			&item.IssuedAt, &item.ExpiresAt, &item.Status, &item.FileID, &item.UploadedAt, &item.VerifiedAt, &item.VerifiedBy, &item.Notes,
			&item.FileStorageKey, &item.FileOriginalName); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) GetLatestByEnrollment(ctx context.Context, enrollmentID uuid.UUID, exec ...db.DBTX) (Certificate, error) {
	var item Certificate
	err := r.base(exec...).QueryRowContext(ctx, `
		select id, user_id, course_id, enrollment_id, certificate_no, issued_by, issued_at::timestamptz,
		       expires_at::timestamptz, status, file_id, uploaded_at, verified_at, verified_by, notes,
		       fa.storage_key, fa.original_name
		from certificates
		join file_attachments fa on fa.id = certificates.file_id
		where enrollment_id = $1
		order by uploaded_at desc
		limit 1
	`, enrollmentID).Scan(&item.ID, &item.UserID, &item.CourseID, &item.EnrollmentID, &item.CertificateNo, &item.IssuedBy,
		&item.IssuedAt, &item.ExpiresAt, &item.Status, &item.FileID, &item.UploadedAt, &item.VerifiedAt, &item.VerifiedBy, &item.Notes,
		&item.FileStorageKey, &item.FileOriginalName)
	return item, err
}

func (r *Repository) UpdateCertificate(ctx context.Context, item Certificate, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update certificates
		set status = $2,
		    verified_at = $3,
		    verified_by = $4,
		    notes = $5
		where id = $1
	`, item.ID, item.Status, item.VerifiedAt, item.VerifiedBy, item.Notes)
	return err
}

type Service struct {
	db                       *sql.DB
	repo                     *Repository
	outbox                   *outbox.Service
	clock                    clock.Clock
	uploadsDir               string
	enrollmentCompletionHook EnrollmentCompletionHook
}

func NewService(
	database *sql.DB,
	repo *Repository,
	outboxService *outbox.Service,
	appClock clock.Clock,
	uploadsDir string,
	enrollmentCompletionHook EnrollmentCompletionHook,
) *Service {
	return &Service{
		db:                       database,
		repo:                     repo,
		outbox:                   outboxService,
		clock:                    appClock,
		uploadsDir:               uploadsDir,
		enrollmentCompletionHook: enrollmentCompletionHook,
	}
}

func (s *Service) Upload(ctx context.Context, principal platformauth.Principal, req UploadCertificateRequest) (Certificate, error) {
	now := s.clock.Now()
	file := FileAttachment{
		ID:              uuid.New(),
		StorageProvider: req.StorageProvider,
		StorageKey:      req.StorageKey,
		OriginalName:    req.OriginalName,
		MimeType:        req.MimeType,
		SizeBytes:       req.SizeBytes,
		UploadedBy:      &principal.UserID,
		CreatedAt:       now,
	}
	item := Certificate{
		ID:               uuid.New(),
		UserID:           principal.UserID,
		CourseID:         req.CourseID,
		EnrollmentID:     req.EnrollmentID,
		IssuedAt:         req.IssuedAt,
		ExpiresAt:        req.ExpiresAt,
		Status:           "uploaded",
		FileID:           file.ID,
		FileStorageKey:   file.StorageKey,
		FileOriginalName: file.OriginalName,
		UploadedAt:       now,
	}

	if req.StorageProvider == "local" {
		if strings.TrimSpace(s.uploadsDir) == "" {
			return Certificate{}, httpx.Internal("uploads_dir_not_configured")
		}
		if len(req.FileContent) == 0 {
			return Certificate{}, httpx.BadRequest("file_required", "certificate file is required")
		}
		if err := uploads.Save(s.uploadsDir, req.StorageKey, req.FileContent); err != nil {
			return Certificate{}, err
		}
	}

	err := db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.repo.CreateFile(ctx, file, tx); err != nil {
			return err
		}
		if err := s.repo.CreateCertificate(ctx, item, tx); err != nil {
			return err
		}
		if err := s.repo.CreateVerification(ctx, item.ID, principal.UserID, "submit", nil, now, tx); err != nil {
			return err
		}
		return s.outbox.Publish(ctx, tx, events.Message{
			Topic:      "certificates",
			EventType:  "certificate.uploaded",
			EntityType: "certificate",
			EntityID:   item.ID,
			Payload: map[string]any{
				"user_id":       item.UserID,
				"enrollment_id": item.EnrollmentID,
			},
			OccurredAt: now,
		})
	})
	if err != nil && req.StorageProvider == "local" && len(req.FileContent) > 0 {
		_ = uploads.Remove(s.uploadsDir, req.StorageKey)
	}
	return item, err
}

func (s *Service) ListMine(ctx context.Context, principal platformauth.Principal) ([]Certificate, error) {
	return s.repo.ListByUser(ctx, principal.UserID)
}

func (s *Service) Verify(ctx context.Context, principal platformauth.Principal, id uuid.UUID, comment *string) (Certificate, error) {
	item, err := s.repo.GetCertificate(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Certificate{}, httpx.NotFound("certificate_not_found", "certificate not found")
		}
		return Certificate{}, err
	}
	now := s.clock.Now()
	item.Status = "verified"
	item.VerifiedAt = &now
	item.VerifiedBy = &principal.UserID
	item.Notes = comment
	err = db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.repo.UpdateCertificate(ctx, item, tx); err != nil {
			return err
		}
		if err := s.repo.CreateVerification(ctx, item.ID, principal.UserID, "verify", comment, now, tx); err != nil {
			return err
		}
		if item.EnrollmentID != nil && s.enrollmentCompletionHook != nil {
			if err := s.enrollmentCompletionHook.CompleteEnrollmentAfterCertificate(ctx, *item.EnrollmentID, principal.UserID, comment, now, tx); err != nil {
				return err
			}
		}
		return s.outbox.Publish(ctx, tx, events.Message{
			Topic:      "certificates",
			EventType:  "certificate.verified",
			EntityType: "certificate",
			EntityID:   item.ID,
			Payload: map[string]any{
				"user_id": item.UserID,
			},
			OccurredAt: now,
		})
	})
	return item, err
}

func (s *Service) Reject(ctx context.Context, principal platformauth.Principal, id uuid.UUID, comment *string) (Certificate, error) {
	item, err := s.repo.GetCertificate(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Certificate{}, httpx.NotFound("certificate_not_found", "certificate not found")
		}
		return Certificate{}, err
	}
	item.Status = "rejected"
	item.VerifiedAt = nil
	item.VerifiedBy = nil
	item.Notes = comment
	err = db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.repo.UpdateCertificate(ctx, item, tx); err != nil {
			return err
		}
		return s.repo.CreateVerification(ctx, item.ID, principal.UserID, "reject", comment, s.clock.Now(), tx)
	})
	return item, err
}

type Handler struct {
	service   *Service
	validator *validator.Validate
}

func NewHandler(service *Service, validator *validator.Validate) *Handler {
	return &Handler{service: service, validator: validator}
}

func certificatesPrincipal(r *http.Request) (platformauth.Principal, error) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		return platformauth.Principal{}, httpx.Unauthorized("unauthorized", "authorization required")
	}
	return principal, nil
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	principal, err := certificatesPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	req, err := decodeCertificateUploadRequest(w, r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}
	item, err := h.service.Upload(r.Context(), principal, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func decodeCertificateUploadRequest(w http.ResponseWriter, r *http.Request) (UploadCertificateRequest, error) {
	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	if strings.HasPrefix(contentType, "multipart/form-data") {
		r.Body = http.MaxBytesReader(w, r.Body, maxCertificateUploadSizeBytes+1024)
		if err := r.ParseMultipartForm(maxCertificateUploadSizeBytes + 1024); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "too large") {
				return UploadCertificateRequest{}, httpx.BadRequest("file_too_large", "certificate size must be within 25MB")
			}
			return UploadCertificateRequest{}, httpx.BadRequest("invalid_multipart", err.Error())
		}

		req := UploadCertificateRequest{
			StorageProvider: strings.TrimSpace(r.FormValue("storage_provider")),
			StorageKey:      strings.TrimSpace(r.FormValue("storage_key")),
			OriginalName:    strings.TrimSpace(r.FormValue("original_name")),
			MimeType:        strings.TrimSpace(r.FormValue("mime_type")),
		}

		if value := strings.TrimSpace(r.FormValue("course_id")); value != "" {
			courseID, err := uuid.Parse(value)
			if err != nil {
				return UploadCertificateRequest{}, httpx.BadRequest("invalid_course_id", "invalid course id")
			}
			req.CourseID = &courseID
		}
		if value := strings.TrimSpace(r.FormValue("enrollment_id")); value != "" {
			enrollmentID, err := uuid.Parse(value)
			if err != nil {
				return UploadCertificateRequest{}, httpx.BadRequest("invalid_enrollment_id", "invalid enrollment id")
			}
			req.EnrollmentID = &enrollmentID
		}
		if value := strings.TrimSpace(r.FormValue("issued_at")); value != "" {
			issuedAt, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return UploadCertificateRequest{}, httpx.BadRequest("invalid_issued_at", "invalid issued at value")
			}
			req.IssuedAt = &issuedAt
		}
		if value := strings.TrimSpace(r.FormValue("expires_at")); value != "" {
			expiresAt, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return UploadCertificateRequest{}, httpx.BadRequest("invalid_expires_at", "invalid expires at value")
			}
			req.ExpiresAt = &expiresAt
		}

		if value := strings.TrimSpace(r.FormValue("size_bytes")); value != "" {
			sizeBytes, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return UploadCertificateRequest{}, httpx.BadRequest("invalid_size_bytes", "invalid file size")
			}
			req.SizeBytes = sizeBytes
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			if errors.Is(err, http.ErrMissingFile) {
				return UploadCertificateRequest{}, httpx.BadRequest("file_required", "certificate file is required")
			}
			return UploadCertificateRequest{}, httpx.BadRequest("invalid_multipart", err.Error())
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			return UploadCertificateRequest{}, err
		}
		if len(content) == 0 {
			return UploadCertificateRequest{}, httpx.BadRequest("file_required", "certificate file is required")
		}

		if req.OriginalName == "" && header != nil {
			req.OriginalName = strings.TrimSpace(header.Filename)
		}
		if req.MimeType == "" && header != nil {
			req.MimeType = strings.TrimSpace(header.Header.Get("Content-Type"))
		}
		if req.MimeType == "" {
			req.MimeType = http.DetectContentType(content)
		}
		if req.SizeBytes <= 0 {
			req.SizeBytes = int64(len(content))
		}
		req.FileContent = content

		return req, nil
	}

	var req UploadCertificateRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		return UploadCertificateRequest{}, err
	}
	return req, nil
}

func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	principal, err := certificatesPrincipal(r)
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

func (h *Handler) Verify(w http.ResponseWriter, r *http.Request) {
	principal, err := certificatesPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_certificate_id", "invalid certificate id"))
		return
	}
	var req ReviewRequest
	if err := httpx.DecodeJSON(r, &req); err != nil && !errors.Is(err, io.EOF) {
		httpx.WriteError(w, err)
		return
	}
	item, err := h.service.Verify(r.Context(), principal, id, req.Comment)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) Reject(w http.ResponseWriter, r *http.Request) {
	principal, err := certificatesPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_certificate_id", "invalid certificate id"))
		return
	}
	var req ReviewRequest
	if err := httpx.DecodeJSON(r, &req); err != nil && !errors.Is(err, io.EOF) {
		httpx.WriteError(w, err)
		return
	}
	item, err := h.service.Reject(r.Context(), principal, id, req.Comment)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}
