package app

import (
	"database/sql"
	"log/slog"

	"github.com/go-playground/validator/v10"

	"moneyapp/backend/internal/config"
	corehealth "moneyapp/backend/internal/core/health"
	adminmodule "moneyapp/backend/internal/modules/admin"
	analyticsmodule "moneyapp/backend/internal/modules/analytics"
	auditmodule "moneyapp/backend/internal/modules/audit"
	catalogmodule "moneyapp/backend/internal/modules/catalog"
	certificatesmodule "moneyapp/backend/internal/modules/certificates"
	externaltrainingmodule "moneyapp/backend/internal/modules/external_training"
	githubmodule "moneyapp/backend/internal/modules/github_integration"
	identitymodule "moneyapp/backend/internal/modules/identity"
	learningmodule "moneyapp/backend/internal/modules/learning"
	notificationsmodule "moneyapp/backend/internal/modules/notifications"
	orgmodule "moneyapp/backend/internal/modules/org"
	outlookmodule "moneyapp/backend/internal/modules/outlook"
	testingmodule "moneyapp/backend/internal/modules/testing"
	universitymodule "moneyapp/backend/internal/modules/university"
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

	OrgService              *orgmodule.Service
	IdentityService         *identitymodule.Service
	AdminService            *adminmodule.Service
	CatalogService          *catalogmodule.Service
	LearningService         *learningmodule.Service
	TestingService          *testingmodule.Service
	CertificatesService     *certificatesmodule.Service
	ExternalTrainingService *externaltrainingmodule.Service
	OutlookService          *outlookmodule.Service
	NotificationsService    *notificationsmodule.Service
	UniversityService       *universitymodule.Service
	AnalyticsService        *analyticsmodule.Service
	AuditService            *auditmodule.Service
	YougileService          *yougilemodule.Service
	GitHubService           *githubmodule.Service

	HealthHandler           *corehealth.Handler
	IdentityHandler         *identitymodule.Handler
	AdminHandler            *adminmodule.Handler
	CatalogHandler          *catalogmodule.Handler
	LearningHandler         *learningmodule.Handler
	TestingHandler          *testingmodule.Handler
	CertificatesHandler     *certificatesmodule.Handler
	ExternalTrainingHandler *externaltrainingmodule.Handler
	OutlookHandler          *outlookmodule.Handler
	NotificationsHandler    *notificationsmodule.Handler
	UniversityHandler       *universitymodule.Handler
	AnalyticsHandler        *analyticsmodule.Handler
	AuditHandler            *auditmodule.Handler
	YougileHandler          *yougilemodule.Handler
	GitHubHandler           *githubmodule.Handler
}
