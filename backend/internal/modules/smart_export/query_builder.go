package smart_export

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ExportFilters holds all optional filter criteria.
type ExportFilters struct {
	EmployeeIDs   []uuid.UUID `json:"employee_ids"`
	DepartmentIDs []uuid.UUID `json:"department_ids"`
	CourseIDs     []uuid.UUID `json:"course_ids"`
	CategoryIDs   []uuid.UUID `json:"category_ids"`
	Levels        []string    `json:"levels"`
	Statuses      []string    `json:"statuses"`
	PriceMin      *float64    `json:"price_min"`
	PriceMax      *float64    `json:"price_max"`
	PriceCurrency *string     `json:"price_currency"`
	DateFrom      *time.Time  `json:"date_from"`
	DateTo        *time.Time  `json:"date_to"`
	DateField     *string     `json:"date_field"`
	Search        *string     `json:"search"`
}

// QueryResult holds the rows returned by the dynamic query.
type QueryResult struct {
	Columns []ColumnDef
	Rows    [][]any
}

// filterMappings maps source keys to table aliases used in WHERE clauses.
type filterMapping struct {
	employeeUserID string // e.g. "ca.applicant_id" or "ep.user_id"
	departmentID   string // e.g. "ep.department_id"
	courseID       string // e.g. "ci.course_id" or "c.id"
	categoryID    string // e.g. "c.category_id"
	level          string // e.g. "c.level"
	status         string // e.g. "ca.status"
	priceExpr      string // e.g. "c.price"
	priceCurrency  string // e.g. "c.price_currency"
	defaultDateCol string // e.g. "ca.created_at"
	searchColumns  []string
}

var sourceMappings = map[string]filterMapping{
	"applications": {
		employeeUserID: "ca.applicant_id",
		departmentID:   "ep.department_id",
		courseID:        "ci.course_id",
		categoryID:     "c.category_id",
		level:          "c.level",
		status:         "ca.status",
		priceExpr:      "c.price",
		priceCurrency:  "c.price_currency",
		defaultDateCol: "ca.created_at",
		searchColumns:  []string{"ci.title", "c.title", "u.email"},
	},
	"intakes": {
		departmentID:   "",
		courseID:        "ci.course_id",
		categoryID:     "c.category_id",
		level:          "c.level",
		status:         "ci.status",
		priceExpr:      "c.price",
		priceCurrency:  "c.price_currency",
		defaultDateCol: "ci.created_at",
		searchColumns:  []string{"ci.title", "c.title"},
	},
	"suggestions": {
		employeeUserID: "cs.suggested_by",
		departmentID:   "ep.department_id",
		status:         "cs.status",
		priceExpr:      "cs.price",
		priceCurrency:  "cs.price_currency",
		defaultDateCol: "cs.created_at",
		searchColumns:  []string{"cs.title", "cs.provider_name"},
	},
	"course_requests": {
		employeeUserID: "cr.employee_user_id",
		departmentID:   "cr.department_id",
		courseID:        "cr.course_id",
		categoryID:     "c.category_id",
		level:          "c.level",
		status:         "cr.status",
		priceExpr:      "c.price",
		priceCurrency:  "c.price_currency",
		defaultDateCol: "cr.requested_at",
		searchColumns:  []string{"cr.employee_full_name", "cr.course_title"},
	},
	"enrollments": {
		employeeUserID: "e.user_id",
		departmentID:   "ep.department_id",
		courseID:        "e.course_id",
		categoryID:     "c.category_id",
		level:          "c.level",
		status:         "e.status",
		priceExpr:      "c.price",
		priceCurrency:  "c.price_currency",
		defaultDateCol: "e.enrolled_at",
		searchColumns:  []string{"c.title", "u.email"},
	},
	"courses": {
		courseID:        "c.id",
		categoryID:     "c.category_id",
		level:          "c.level",
		status:         "c.status",
		priceExpr:      "c.price",
		priceCurrency:  "c.price_currency",
		defaultDateCol: "c.created_at",
		searchColumns:  []string{"c.title", "p.name"},
	},
	"employees": {
		employeeUserID: "ep.user_id",
		departmentID:   "ep.department_id",
		status:         "ep.employment_status",
		defaultDateCol: "ep.hire_date",
		searchColumns:  []string{"ep.last_name", "ep.first_name", "u.email", "ep.position_title"},
	},
}

// BuildQuery constructs and executes the dynamic SQL, returning typed rows.
func BuildQuery(ctx context.Context, db *sql.DB, source *SourceDef, columns []ColumnDef, filters *ExportFilters, sortBy *string, sortDir *string) (*QueryResult, error) {
	// SELECT
	selectParts := make([]string, len(columns))
	for i, col := range columns {
		selectParts[i] = col.SQLExpr + " AS " + col.SQLAlias
	}

	// WHERE
	mapping := sourceMappings[source.Key]
	whereClauses, args := buildWhere(filters, mapping)

	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(selectParts, ", "), source.FromSQL)
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// ORDER BY
	if sortBy != nil && *sortBy != "" {
		sortAlias := ""
		for _, col := range columns {
			if col.Key == *sortBy {
				sortAlias = col.SQLAlias
				break
			}
		}
		if sortAlias != "" {
			dir := "ASC"
			if sortDir != nil && strings.ToUpper(*sortDir) == "DESC" {
				dir = "DESC"
			}
			query += fmt.Sprintf(" ORDER BY %s %s NULLS LAST", sortAlias, dir)
		}
	} else {
		// Default: sort by first date column or created_at
		if mapping.defaultDateCol != "" {
			query += " ORDER BY " + mapping.defaultDateCol + " DESC NULLS LAST"
		}
	}

	query += " LIMIT 50000"

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("export query failed: %w", err)
	}
	defer rows.Close()

	var result [][]any
	for rows.Next() {
		vals := make([]any, len(columns))
		ptrs := make([]any, len(columns))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		result = append(result, vals)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &QueryResult{Columns: columns, Rows: result}, nil
}

func buildWhere(filters *ExportFilters, m filterMapping) ([]string, []any) {
	if filters == nil {
		return nil, nil
	}

	var clauses []string
	var args []any
	paramIdx := 1

	if len(filters.EmployeeIDs) > 0 && m.employeeUserID != "" {
		clauses = append(clauses, m.employeeUserID+" = ANY("+placeholder(&paramIdx)+")")
		args = append(args, uuidArray(filters.EmployeeIDs))
	}
	if len(filters.DepartmentIDs) > 0 && m.departmentID != "" {
		clauses = append(clauses, m.departmentID+" = ANY("+placeholder(&paramIdx)+")")
		args = append(args, uuidArray(filters.DepartmentIDs))
	}
	if len(filters.CourseIDs) > 0 && m.courseID != "" {
		clauses = append(clauses, m.courseID+" = ANY("+placeholder(&paramIdx)+")")
		args = append(args, uuidArray(filters.CourseIDs))
	}
	if len(filters.CategoryIDs) > 0 && m.categoryID != "" {
		clauses = append(clauses, m.categoryID+" = ANY("+placeholder(&paramIdx)+")")
		args = append(args, uuidArray(filters.CategoryIDs))
	}
	if len(filters.Levels) > 0 && m.level != "" {
		clauses = append(clauses, m.level+" = ANY("+placeholder(&paramIdx)+")")
		args = append(args, stringArray(filters.Levels))
	}
	if len(filters.Statuses) > 0 && m.status != "" {
		clauses = append(clauses, m.status+" = ANY("+placeholder(&paramIdx)+")")
		args = append(args, stringArray(filters.Statuses))
	}
	if filters.PriceMin != nil && m.priceExpr != "" {
		clauses = append(clauses, m.priceExpr+" >= "+placeholder(&paramIdx))
		args = append(args, *filters.PriceMin)
	}
	if filters.PriceMax != nil && m.priceExpr != "" {
		clauses = append(clauses, m.priceExpr+" <= "+placeholder(&paramIdx))
		args = append(args, *filters.PriceMax)
	}
	if filters.PriceCurrency != nil && m.priceCurrency != "" {
		clauses = append(clauses, m.priceCurrency+" = "+placeholder(&paramIdx))
		args = append(args, *filters.PriceCurrency)
	}

	dateCol := m.defaultDateCol
	if filters.DateField != nil && *filters.DateField != "" {
		dateCol = *filters.DateField
	}
	if filters.DateFrom != nil && dateCol != "" {
		clauses = append(clauses, dateCol+" >= "+placeholder(&paramIdx))
		args = append(args, *filters.DateFrom)
	}
	if filters.DateTo != nil && dateCol != "" {
		clauses = append(clauses, dateCol+" <= "+placeholder(&paramIdx))
		args = append(args, *filters.DateTo)
	}

	if filters.Search != nil && *filters.Search != "" && len(m.searchColumns) > 0 {
		searchParts := make([]string, len(m.searchColumns))
		p := placeholder(&paramIdx)
		for i, col := range m.searchColumns {
			searchParts[i] = "coalesce(" + col + ",'') ILIKE " + p
		}
		clauses = append(clauses, "("+strings.Join(searchParts, " OR ")+")")
		args = append(args, "%"+*filters.Search+"%")
	}

	return clauses, args
}

func placeholder(idx *int) string {
	s := fmt.Sprintf("$%d", *idx)
	*idx++
	return s
}

func uuidArray(ids []uuid.UUID) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = id.String()
	}
	return out
}

func stringArray(vals []string) []string {
	return vals
}
