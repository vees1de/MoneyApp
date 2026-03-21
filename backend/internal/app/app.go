package app

import (
	"context"
	"net/http"
	"time"

	"moneyapp/backend/internal/config"
	coreaudit "moneyapp/backend/internal/core/audit"
	coreauth "moneyapp/backend/internal/core/auth"
	corehealth "moneyapp/backend/internal/core/health"
	corejobs "moneyapp/backend/internal/core/jobs"
	corelinks "moneyapp/backend/internal/core/links"
	coresessions "moneyapp/backend/internal/core/sessions"
	coreusers "moneyapp/backend/internal/core/users"
	"moneyapp/backend/internal/integrations/telegram"
	"moneyapp/backend/internal/integrations/yandex"
	dashboardmodule "moneyapp/backend/internal/modules/dashboard"
	financeaccounts "moneyapp/backend/internal/modules/finance/accounts"
	financecategories "moneyapp/backend/internal/modules/finance/categories"
	financesummary "moneyapp/backend/internal/modules/finance/summary"
	financetransactions "moneyapp/backend/internal/modules/finance/transactions"
	reviewmodule "moneyapp/backend/internal/modules/review"
	savingsmodule "moneyapp/backend/internal/modules/savings"
	platformauth "moneyapp/backend/internal/platform/auth"
	platformcache "moneyapp/backend/internal/platform/cache"
	"moneyapp/backend/internal/platform/clock"
	"moneyapp/backend/internal/platform/db"
	platformevents "moneyapp/backend/internal/platform/events"
	"moneyapp/backend/internal/platform/jobs"
	"moneyapp/backend/internal/platform/logger"
	"moneyapp/backend/internal/platform/validation"
)

type App struct {
	container *Container
	server    *http.Server
}

func New(cfg *config.Config) (*App, error) {
	log := logger.New(cfg.Environment)
	database, err := db.Open(context.Background(), cfg.Database)
	if err != nil {
		return nil, err
	}

	redisStore, err := platformcache.NewRedisStore(context.Background(), cfg.Redis)
	if err != nil {
		_ = database.Close()
		return nil, err
	}

	kafkaPublisher, err := platformevents.NewKafkaPublisher(context.Background(), cfg.Kafka)
	if err != nil {
		_ = redisStore.Close()
		_ = database.Close()
		return nil, err
	}

	dispatcher := jobs.NewDispatcher(log)
	scheduler := jobs.NewScheduler(log, dispatcher)
	appClock := clock.RealClock{}
	jwtManager := platformauth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.JWTIssuer, cfg.Auth.AccessTokenTTL)
	validate := validation.New()

	userRepo := coreusers.NewRepository(database)
	sessionRepo := coresessions.NewRepository(database)
	auditRepo := coreaudit.NewRepository(database)
	authRepo := coreauth.NewRepository(database)
	linksRepo := corelinks.NewRepository(database)
	accountRepo := financeaccounts.NewRepository(database)
	categoryRepo := financecategories.NewRepository(database)
	transactionRepo := financetransactions.NewRepository(database)
	summaryRepo := financesummary.NewRepository(database)
	savingsRepo := savingsmodule.NewRepository(database)
	reviewRepo := reviewmodule.NewRepository(database)
	dashboardRepo := dashboardmodule.NewRepository(database)

	auditService := coreaudit.NewService(auditRepo, appClock, log, kafkaPublisher, cfg.Kafka.AuditTopic)
	userService := coreusers.NewService(database, userRepo)
	sessionService := coresessions.NewService(database, sessionRepo, jwtManager, appClock, cfg.Auth.AccessTokenTTL, cfg.Auth.RefreshTokenTTL)
	authService := coreauth.NewService(
		database,
		cfg.Auth,
		appClock,
		authRepo,
		userRepo,
		sessionService,
		auditService,
		telegram.NewVerifier(cfg.Integrations.Telegram.BotToken, cfg.Auth.AllowInsecureDevAuth),
		yandex.NewVerifier(cfg.Integrations.Yandex.ClientID, cfg.Auth.AllowInsecureDevAuth),
	)
	linksService := corelinks.NewService(linksRepo, appClock)
	accountService := financeaccounts.NewService(database, accountRepo, auditService, appClock)
	categoryService := financecategories.NewService(categoryRepo, auditService, appClock)
	transactionService := financetransactions.NewService(database, transactionRepo, accountRepo, categoryRepo, auditService, appClock)
	summaryService := financesummary.NewService(summaryRepo)
	savingsService := savingsmodule.NewService(savingsRepo, summaryRepo, auditService, appClock)
	reviewService := reviewmodule.NewService(reviewRepo, accountRepo, transactionRepo, auditService, appClock)
	dashboardService := dashboardmodule.NewService(dashboardRepo, summaryService, savingsService, reviewService, appClock, redisStore, cfg.Redis.DashboardTTL)
	jobService := corejobs.NewService(scheduler)
	healthService := corehealth.NewService(map[string]corehealth.CheckFunc{
		"postgres": database.PingContext,
		"redis":    redisStore.Ping,
		"kafka":    kafkaPublisher.Ping,
	})

	container := &Container{
		Config:             cfg,
		Logger:             log,
		DB:                 database,
		Cache:              redisStore,
		Events:             kafkaPublisher,
		Clock:              appClock,
		JWT:                jwtManager,
		Dispatcher:         dispatcher,
		Scheduler:          scheduler,
		Validator:          validate,
		HealthService:      healthService,
		UserService:        userService,
		AuthService:        authService,
		AuditService:       auditService,
		LinksService:       linksService,
		JobService:         jobService,
		AccountService:     accountService,
		CategoryService:    categoryService,
		TransactionService: transactionService,
		SummaryService:     summaryService,
		SavingsService:     savingsService,
		ReviewService:      reviewService,
		DashboardService:   dashboardService,
		HealthHandler:      corehealth.NewHandler(healthService),
		AuthHandler:        coreauth.NewHandler(authService, validate),
		UserHandler:        coreusers.NewHandler(userService, validate),
		LinksHandler:       corelinks.NewHandler(linksService, validate),
		AccountHandler:     financeaccounts.NewHandler(accountService, validate),
		CategoryHandler:    financecategories.NewHandler(categoryService, validate),
		TransactionHandler: financetransactions.NewHandler(transactionService, validate),
		SavingsHandler:     savingsmodule.NewHandler(savingsService, validate),
		ReviewHandler:      reviewmodule.NewHandler(reviewService, validate),
		DashboardHandler:   dashboardmodule.NewHandler(dashboardService),
	}

	server := &http.Server{
		Addr:         cfg.HTTP.Address,
		Handler:      NewRouter(container),
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	return &App{
		container: container,
		server:    server,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	a.container.Scheduler.Start(ctx)

	serverErr := make(chan error, 1)
	go func() {
		a.container.Logger.Info("http server started", "addr", a.server.Addr)
		serverErr <- a.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			_ = a.container.Events.Close()
			_ = a.container.Cache.Close()
			_ = a.container.DB.Close()
			return err
		}
	}

	a.container.Scheduler.Stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		_ = a.container.Events.Close()
		_ = a.container.Cache.Close()
		_ = a.container.DB.Close()
		return err
	}

	_ = a.container.Events.Close()
	_ = a.container.Cache.Close()
	_ = a.container.DB.Close()
	return nil
}
