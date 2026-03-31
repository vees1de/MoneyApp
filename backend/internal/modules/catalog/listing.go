package catalog

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"moneyapp/backend/internal/core/common"
	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/db"
	"moneyapp/backend/internal/platform/httpx"
)

const defaultCourseListLimit = 20

type CourseListFilters struct {
	Query              string
	Types              []string
	Statuses           []string
	DirectionIDs       []string
	CategoryIDs        []string
	Levels             []string
	Languages          []string
	SourceTypes        []string
	IsMandatoryDefault *bool
	Sort               string
	Pagination         common.Pagination
}

func parseCourseListFilters(r *http.Request, principal platformauth.Principal) (CourseListFilters, error) {
	query := r.URL.Query()
	filters := CourseListFilters{
		Query:        strings.TrimSpace(query.Get("q")),
		Types:        normalizeValues(query["type"]),
		Statuses:     normalizeValues(query["status"]),
		DirectionIDs: normalizeValues(query["direction_id"]),
		CategoryIDs:  normalizeValues(query["category_id"]),
		Levels:       normalizeValues(query["level"]),
		Languages:    normalizeValues(query["language"]),
		SourceTypes:  normalizeValues(query["source_type"]),
		Sort:         strings.TrimSpace(query.Get("sort")),
		Pagination:   common.PaginationFromRequest(r, defaultCourseListLimit),
	}

	if raw := strings.TrimSpace(query.Get("is_mandatory_default")); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return CourseListFilters{}, httpx.BadRequest("invalid_is_mandatory_default", "is_mandatory_default must be a boolean")
		}
		filters.IsMandatoryDefault = &value
	}

	if !principal.HasPermission("courses.write") {
		filters.Statuses = []string{"published"}
	}

	if filters.Sort == "" {
		if filters.Query != "" {
			filters.Sort = "relevance"
		} else {
			filters.Sort = "newest"
		}
	}

	switch filters.Sort {
	case "relevance", "newest", "updated", "title_asc", "title_desc", "duration_asc", "duration_desc":
	default:
		return CourseListFilters{}, httpx.BadRequest("invalid_sort", "unsupported sort value")
	}

	return filters, nil
}

func normalizeValues(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, raw := range values {
		for _, part := range strings.Split(raw, ",") {
			value := strings.TrimSpace(part)
			if value == "" {
				continue
			}
			if _, ok := seen[value]; ok {
				continue
			}
			seen[value] = struct{}{}
			result = append(result, value)
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

func (r *Repository) ListCoursesFiltered(ctx context.Context, filters CourseListFilters, exec ...db.DBTX) ([]Course, error) {
	args := make([]any, 0, 16)
	where := make([]string, 0, 8)

	searchVector := "to_tsvector('simple', coalesce(title, '') || ' ' || coalesce(short_description, '') || ' ' || coalesce(description, ''))"
	searchQuery := ""
	if filters.Query != "" {
		args = append(args, filters.Query)
		searchQuery = fmt.Sprintf("websearch_to_tsquery('simple', $%d)", len(args))
		where = append(where, fmt.Sprintf("%s @@ %s", searchVector, searchQuery))
	}

	addStringFilter := func(column string, values []string) {
		if len(values) == 0 {
			return
		}
		placeholders := make([]string, 0, len(values))
		for _, value := range values {
			args = append(args, value)
			placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)))
		}
		where = append(where, fmt.Sprintf("%s in (%s)", column, strings.Join(placeholders, ", ")))
	}

	addStringFilter("type", filters.Types)
	addStringFilter("status", filters.Statuses)
	addStringFilter("direction_id::text", filters.DirectionIDs)
	addStringFilter("category_id::text", filters.CategoryIDs)
	addStringFilter("level", filters.Levels)
	addStringFilter("language", filters.Languages)
	addStringFilter("source_type", filters.SourceTypes)

	if filters.IsMandatoryDefault != nil {
		args = append(args, *filters.IsMandatoryDefault)
		where = append(where, fmt.Sprintf("is_mandatory_default = $%d", len(args)))
	}

	query := strings.Builder{}
	query.WriteString(`
		select id, type, source_type, title, slug, short_description, description,
		       provider_id, category_id, direction_id, level, duration_hours::text, language,
		       is_mandatory_default, status, external_url, price::text, price_currency, next_start_date,
		       thumbnail_file_id, created_by, updated_by,
		       published_at, archived_at, created_at, updated_at
		from courses
	`)
	if len(where) > 0 {
		query.WriteString(" where ")
		query.WriteString(strings.Join(where, " and "))
	}

	switch filters.Sort {
	case "relevance":
		if searchQuery == "" {
			query.WriteString(" order by coalesce(published_at, created_at) desc, created_at desc")
		} else {
			query.WriteString(" order by ")
			query.WriteString(fmt.Sprintf("ts_rank(%s, %s) desc, ", searchVector, searchQuery))
			query.WriteString("coalesce(published_at, created_at) desc, created_at desc")
		}
	case "updated":
		query.WriteString(" order by updated_at desc")
	case "title_asc":
		query.WriteString(" order by title asc, created_at desc")
	case "title_desc":
		query.WriteString(" order by title desc, created_at desc")
	case "duration_asc":
		query.WriteString(" order by duration_hours asc nulls last, created_at desc")
	case "duration_desc":
		query.WriteString(" order by duration_hours desc nulls last, created_at desc")
	default:
		query.WriteString(" order by coalesce(published_at, created_at) desc, created_at desc")
	}

	args = append(args, filters.Pagination.Limit, filters.Pagination.Offset)
	query.WriteString(fmt.Sprintf(" limit $%d offset $%d", len(args)-1, len(args)))

	rows, err := r.base(exec...).QueryContext(ctx, query.String(), args...)
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
