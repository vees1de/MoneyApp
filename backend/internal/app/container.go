package app

import (
	"database/sql"
	"log/slog"

	"github.com/go-playground/validator/v10"

	"moneyapp/backend/internal/config"
	corehealth "moneyapp/backend/internal/core/health"
	coreusers "moneyapp/backend/internal/core/users"
	adminmodule "moneyapp/backend/internal/modules/admin"
	analyticsmodule "moneyapp/backend/internal/modules/analytics"
	auditmodule "moneyapp/backend/internal/modules/audit"
	boardsummarymodule "moneyapp/backend/internal/modules/board_summary"
	calendarmodule "moneyapp/backend/internal/modules/calendar"
	catalogmodule "moneyapp/backend/internal/modules/catalog"
	certificatesmodule "moneyapp/backend/internal/modules/certificates"
	courseintakesmodule "moneyapp/backend/internal/modules/course_intakes"
	courserequestsmodule "moneyapp/backend/internal/modules/course_requests"
	dashboardapimodule "moneyapp/backend/internal/modules/dashboard_api"
	externaltrainingmodule "moneyapp/backend/internal/modules/external_training"
	githubmodule "moneyapp/backend/internal/modules/github_integration"
	identitymodule "moneyapp/backend/internal/modules/identity"
	learningmodule "moneyapp/backend/internal/modules/learning"
	learningplanmodule "moneyapp/backend/internal/modules/learning_plan"
	notificationsmodule "moneyapp/backend/internal/modules/notifications"
	orgmodule "moneyapp/backend/internal/modules/org"
	outlookmodule "moneyapp/backend/internal/modules/outlook"
	smartexportmodule "moneyapp/backend/internal/modules/smart_export"
	testingmodule "moneyapp/backend/internal/modules/testing"
	universitymodule "moneyapp/backend/internal/modules/university"
	employeesstatsmodule "moneyapp/backend/internal/modules/employees_stats"
	yougilemodule "moneyapp/backend/internal/modules/yougile"
	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/clock"
	"moneyapp/backend/internal/platform/outbox"
	platformworker "moneyapp/backend/internal/platform/worker"
)

type Container struct {
	Config    *config.Config
	Logger    *slog.Logger
	DB        *sql.DB
	Clock     clock.Clock
	JWT       *platformauth.JWTManager
	Outbox    *outbox.Service
	Queue     *platformworker.Queue
	Validator *validator.Validate

	HealthService *corehealth.Service
	UsersService  *coreusers.Service

	OrgService              *orgmodule.Service
	IdentityService         *identitymodule.Service
	AdminService            *adminmodule.Service
	CatalogService          *catalogmodule.Service
	LearningService         *learningmodule.Service
	TestingService          *testingmodule.Service
	CertificatesService     *certificatesmodule.Service
	CourseIntakesService    *courseintakesmodule.Service
	CourseRequestsService   *courserequestsmodule.Service
	ExternalTrainingService *externaltrainingmodule.Service
	CalendarService         *calendarmodule.Service
	LearningPlanService     *learningplanmodule.Service
	BoardSummaryService     *boardsummarymodule.Service
	DashboardAPIService     *dashboardapimodule.Service
	OutlookService          *outlookmodule.Service
	NotificationsService    *notificationsmodule.Service
	UniversityService       *universitymodule.Service
	SmartExportService      *smartexportmodule.Service
	AnalyticsService        *analyticsmodule.Service
	AuditService            *auditmodule.Service
	YougileService          *yougilemodule.Service
	GitHubService           *githubmodule.Service
	EmployeesStatsService   *employeesstatsmodule.Service

	HealthHandler           *corehealth.Handler
	IdentityHandler         *identitymodule.Handler
	AdminHandler            *adminmodule.Handler
	CatalogHandler          *catalogmodule.Handler
	LearningHandler         *learningmodule.Handler
	TestingHandler          *testingmodule.Handler
	CertificatesHandler     *certificatesmodule.Handler
	CourseIntakesHandler    *courseintakesmodule.Handler
	CourseRequestsHandler   *courserequestsmodule.Handler
	ExternalTrainingHandler *externaltrainingmodule.Handler
	CalendarHandler         *calendarmodule.Handler
	LearningPlanHandler     *learningplanmodule.Handler
	BoardSummaryHandler     *boardsummarymodule.Handler
	DashboardAPIHandler     *dashboardapimodule.Handler
	OutlookHandler          *outlookmodule.Handler
	NotificationsHandler    *notificationsmodule.Handler
	UniversityHandler       *universitymodule.Handler
	SmartExportHandler      *smartexportmodule.Handler
	AnalyticsHandler        *analyticsmodule.Handler
	AuditHandler            *auditmodule.Handler
	YougileHandler          *yougilemodule.Handler
	GitHubHandler           *githubmodule.Handler
	EmployeesStatsHandler   *employeesstatsmodule.Handler
	UsersHandler            *coreusers.Handler
}
