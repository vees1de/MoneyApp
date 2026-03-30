package external_training

import (
	"net/http/httptest"
	"testing"

	platformauth "moneyapp/backend/internal/platform/auth"

	"github.com/google/uuid"
)

func TestCanListExternalScope(t *testing.T) {
	manager := platformauth.Principal{
		UserID: uuid.New(),
		PermissionCodes: []string{
			"external_requests.read_all",
			"external_requests.approve_manager",
		},
	}
	employee := platformauth.Principal{
		UserID:          uuid.New(),
		PermissionCodes: []string{"external_requests.read_own"},
	}

	if !canListExternalScope(manager, "team") {
		t.Fatalf("manager should be able to list team scope")
	}
	if !canListExternalScope(manager, "all") {
		t.Fatalf("manager with read_all should be able to list all scope")
	}
	if canListExternalScope(employee, "team") {
		t.Fatalf("employee should not be able to list team scope")
	}
}

func TestParseRequestListFiltersDefaultsScopeAndParsesAssignee(t *testing.T) {
	assigneeID := uuid.New()
	req := httptest.NewRequest("GET", "/api/v1/external-requests?status=draft&status=approved&assignee="+assigneeID.String(), nil)

	filters, err := parseRequestListFilters(req)
	if err != nil {
		t.Fatalf("parseRequestListFilters() error = %v", err)
	}
	if filters.Scope != "my" {
		t.Fatalf("expected default scope my, got %q", filters.Scope)
	}
	if filters.AssigneeID == nil || *filters.AssigneeID != assigneeID {
		t.Fatalf("expected assignee %s, got %+v", assigneeID, filters.AssigneeID)
	}
	if len(filters.Statuses) != 2 {
		t.Fatalf("expected two statuses, got %+v", filters.Statuses)
	}
}
