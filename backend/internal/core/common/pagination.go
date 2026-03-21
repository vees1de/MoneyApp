package common

import (
	"net/http"
	"strconv"
)

type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func PaginationFromRequest(r *http.Request, defaultLimit int) Pagination {
	query := r.URL.Query()

	limit := defaultLimit
	if raw := query.Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	offset := 0
	if raw := query.Get("offset"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return Pagination{
		Limit:  limit,
		Offset: offset,
	}
}
