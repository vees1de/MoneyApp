package smart_export

import (
	"database/sql"
	"net/http"
	"time"

	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/go-playground/validator/v10"
)

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

// SmartExportRequest is the POST body for /reports/smart-export.
type SmartExportRequest struct {
	Source  string         `json:"source" validate:"required,oneof=applications intakes suggestions course_requests enrollments courses employees"`
	Columns []string       `json:"columns"`
	Filters *ExportFilters `json:"filters"`
	SortBy  *string        `json:"sort_by"`
	SortDir *string        `json:"sort_dir" validate:"omitempty,oneof=asc desc"`
	Format  string         `json:"format" validate:"omitempty,oneof=xlsx csv"`
}

// SourcesResponse is returned by GET /reports/sources.
type SourcesResponse struct {
	Sources []SourceMeta `json:"sources"`
}

// SourceMeta is the public metadata for a single source (no SQL internals).
type SourceMeta struct {
	Key     string       `json:"key"`
	Label   string       `json:"label"`
	Columns []ColumnMeta `json:"columns"`
}

// ColumnMeta is public column info sent to the frontend.
type ColumnMeta struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Type    string `json:"type"`
	Default bool   `json:"default"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type Service struct {
	db *sql.DB
}

func NewService(database *sql.DB) *Service {
	return &Service{db: database}
}

// Sources returns metadata for all available export sources.
func (s *Service) Sources() SourcesResponse {
	sources := allSources()
	metas := make([]SourceMeta, len(sources))
	for i, src := range sources {
		cols := make([]ColumnMeta, len(src.Columns))
		for j, c := range src.Columns {
			cols[j] = ColumnMeta{Key: c.Key, Label: c.Label, Type: c.Type, Default: c.Default}
		}
		metas[i] = SourceMeta{Key: src.Key, Label: src.Label, Columns: cols}
	}
	return SourcesResponse{Sources: metas}
}

// Export builds the query, runs it, and returns Excel bytes + filename.
func (s *Service) Export(r *http.Request, req SmartExportRequest) ([]byte, string, error) {
	src := getSource(req.Source)
	if src == nil {
		return nil, "", httpx.BadRequest("invalid_source", "unknown export source: "+req.Source)
	}

	// Resolve selected columns
	selected := resolveColumns(src.Columns, req.Columns)
	if len(selected) == 0 {
		return nil, "", httpx.BadRequest("no_columns", "no columns selected for export")
	}

	// Build and execute query
	qr, err := BuildQuery(r.Context(), s.db, src, selected, req.Filters, req.SortBy, req.SortDir)
	if err != nil {
		return nil, "", err
	}

	// Translate enum values to Russian
	LocalizeRows(qr)

	// Generate XLSX
	sheetName := src.Label
	data, err := GenerateXLSX(qr, sheetName)
	if err != nil {
		return nil, "", err
	}

	filename := req.Source + "-export-" + time.Now().Format("20060102-150405") + ".xlsx"
	return data, filename, nil
}

// resolveColumns picks columns by keys, or returns defaults if keys is empty.
func resolveColumns(all []ColumnDef, keys []string) []ColumnDef {
	if len(keys) == 0 {
		var defaults []ColumnDef
		for _, c := range all {
			if c.Default {
				defaults = append(defaults, c)
			}
		}
		return defaults
	}

	keySet := make(map[string]int, len(keys))
	for i, k := range keys {
		keySet[k] = i
	}

	result := make([]ColumnDef, 0, len(keys))
	// Preserve the order from the request
	ordered := make([]ColumnDef, len(keys))
	found := 0
	for _, c := range all {
		if idx, ok := keySet[c.Key]; ok {
			ordered[idx] = c
			found++
		}
	}
	for _, c := range ordered {
		if c.Key != "" {
			result = append(result, c)
		}
	}
	return result
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

// Sources returns available data sources and their columns.
// GET /api/v1/reports/sources
func (h *Handler) Sources(w http.ResponseWriter, r *http.Request) {
	_, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authentication required"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, h.service.Sources())
}

// SmartExport generates and returns the Excel file.
// POST /api/v1/reports/smart-export
func (h *Handler) SmartExport(w http.ResponseWriter, r *http.Request) {
	_, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authentication required"))
		return
	}

	var req SmartExportRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}

	data, filename, err := h.service.Export(r, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
