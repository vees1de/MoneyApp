package app

import (
	"database/sql"
	"log/slog"

	"github.com/go-playground/validator/v10"

	"moneyapp/backend/internal/config"
	coreaudit "moneyapp/backend/internal/core/audit"
	coreauth "moneyapp/backend/internal/core/auth"
	corehealth "moneyapp/backend/internal/core/health"
	corejobs "moneyapp/backend/internal/core/jobs"
	corelinks "moneyapp/backend/internal/core/links"
	coreusers "moneyapp/backend/internal/core/users"
	dashboardmodule "moneyapp/backend/internal/modules/dashboard"
	financeaccounts "moneyapp/backend/internal/modules/finance/accounts"
	financecategories "moneyapp/backend/internal/modules/finance/categories"
	financesummary "moneyapp/backend/internal/modules/finance/summary"
	financetransactions "moneyapp/backend/internal/modules/finance/transactions"
	financetransfers "moneyapp/backend/internal/modules/finance/transfers"
	reviewmodule "moneyapp/backend/internal/modules/review"
	savingsmodule "moneyapp/backend/internal/modules/savings"
	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/clock"
	platformjobs "moneyapp/backend/internal/platform/jobs"
)

type Container struct {
	Config     *config.Config
	Logger     *slog.Logger
	DB         *sql.DB
	Clock      clock.Clock
	JWT        *platformauth.JWTManager
	Dispatcher *platformjobs.Dispatcher
	Scheduler  *platformjobs.Scheduler
	Validator  *validator.Validate

	HealthService      *corehealth.Service
	UserService        *coreusers.Service
	AuthService        *coreauth.Service
	AuditService       *coreaudit.Service
	LinksService       *corelinks.Service
	JobService         *corejobs.Service
	AccountService     *financeaccounts.Service
	CategoryService    *financecategories.Service
	TransactionService *financetransactions.Service
	TransferService    *financetransfers.Service
	SummaryService     *financesummary.Service
	SavingsService     *savingsmodule.Service
	ReviewService      *reviewmodule.Service
	DashboardService   *dashboardmodule.Service

	HealthHandler      *corehealth.Handler
	AuthHandler        *coreauth.Handler
	UserHandler        *coreusers.Handler
	LinksHandler       *corelinks.Handler
	AccountHandler     *financeaccounts.Handler
	CategoryHandler    *financecategories.Handler
	TransactionHandler *financetransactions.Handler
	TransferHandler    *financetransfers.Handler
	SavingsHandler     *savingsmodule.Handler
	ReviewHandler      *reviewmodule.Handler
	DashboardHandler   *dashboardmodule.Handler
}
