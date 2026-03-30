package catalog

import (
	"net/http/httptest"
	"testing"

	platformauth "moneyapp/backend/internal/platform/auth"

	"github.com/google/uuid"
)

func TestParseCourseListFiltersForcesPublishedForReadOnlyUsers(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/courses?status=draft&sort=updated", nil)
	principal := platformauth.Principal{UserID: uuid.New(), PermissionCodes: []string{"courses.read"}}

	filters, err := parseCourseListFilters(req, principal)
	if err != nil {
		t.Fatalf("parseCourseListFilters() error = %v", err)
	}
	if len(filters.Statuses) != 1 || filters.Statuses[0] != "published" {
		t.Fatalf("expected status to be forced to published, got %+v", filters.Statuses)
	}
	if filters.Sort != "updated" {
		t.Fatalf("expected explicit sort to be preserved, got %q", filters.Sort)
	}
}

func TestParseCourseListFiltersDefaultsToRelevanceWhenSearchPresent(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/courses?q=go", nil)
	principal := platformauth.Principal{UserID: uuid.New(), PermissionCodes: []string{"courses.read", "courses.write"}}

	filters, err := parseCourseListFilters(req, principal)
	if err != nil {
		t.Fatalf("parseCourseListFilters() error = %v", err)
	}
	if filters.Sort != "relevance" {
		t.Fatalf("expected relevance sort, got %q", filters.Sort)
	}
}
