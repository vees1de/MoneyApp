package external_training

import (
	"net/http/httptest"
	"testing"

	platformauth "moneyapp/backend/internal/platform/auth"

	"github.com/google/uuid"
)

func TestCanListExternalScope(t *testing.T) {
	manager := platformauth.Principal{
		UserID:    uuid.New(),
		RoleCodes: []string{"manager"},
	}
	hr := platformauth.Principal{
		UserID:    uuid.New(),
		RoleCodes: []string{"hr"},
	}
	employee := platformauth.Principal{
		UserID:          uuid.New(),
		RoleCodes:       []string{"employee"},
		PermissionCodes: []string{"external_requests.read_own"},
	}

	if !canListExternalScope(manager, "team") {
		t.Fatalf("manager should be able to list team scope")
	}
	if canListExternalScope(manager, "all") {
		t.Fatalf("manager should not be able to list all scope")
	}
	if !canListExternalScope(hr, "all") {
		t.Fatalf("hr should be able to list all scope")
	}
	if canListExternalScope(employee, "team") {
		t.Fatalf("employee should not be able to list team scope")
	}
}

func TestSelfServiceRoleHelpers(t *testing.T) {
	for _, roleCode := range []string{"manager", "hr", "trainer", "admin"} {
		principal := platformauth.Principal{
			UserID:    uuid.New(),
			RoleCodes: []string{roleCode},
		}
		if !canCreateOwnExternalRequest(principal) {
			t.Fatalf("%s should be able to create own external request", roleCode)
		}
		if !canReadOwnExternalRequests(principal) {
			t.Fatalf("%s should be able to read own external requests", roleCode)
		}
	}
}

func TestSubmissionHistoryAction(t *testing.T) {
	if action := submissionHistoryAction("draft"); action != "submitted" {
		t.Fatalf("expected submitted action for draft, got %q", action)
	}
	if action := submissionHistoryAction("needs_revision"); action != "resubmitted" {
		t.Fatalf("expected resubmitted action for needs_revision, got %q", action)
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
