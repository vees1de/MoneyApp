package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"moneyapp/backend/internal/config"
	corehealth "moneyapp/backend/internal/core/health"
	adminmodule "moneyapp/backend/internal/modules/admin"
	analyticsmodule "moneyapp/backend/internal/modules/analytics"
	auditmodule "moneyapp/backend/internal/modules/audit"
	catalogmodule "moneyapp/backend/internal/modules/catalog"
	certificatesmodule "moneyapp/backend/internal/modules/certificates"
	courserequestsmodule "moneyapp/backend/internal/modules/course_requests"
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
	"moneyapp/backend/internal/platform/db"
	"moneyapp/backend/internal/platform/logger"
	"moneyapp/backend/internal/platform/outbox"
	"moneyapp/backend/internal/platform/validation"
	platformworker "moneyapp/backend/internal/platform/worker"
)

type App struct {
	container *Container
	server    *http.Server
}

func New(cfg *config.Config) (*App, error) {
	container, err := NewContainer(cfg)
	if err != nil {
		return nil, err
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

func NewContainer(cfg *config.Config) (*Container, error) {
	log := logger.New(cfg.Environment)
	database, err := db.Open(context.Background(), cfg.Database)
	if err != nil {
		return nil, err
	}

	appClock := clock.RealClock{}
	jwtManager := platformauth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.JWTIssuer, cfg.Auth.AccessTokenTTL)
	validate := validation.New()
	outboxService := outbox.NewService()
	queue := platformworker.NewQueue(database, log)

	orgRepo := orgmodule.NewRepository(database)
	identityRepo := identitymodule.NewRepository(database)
	catalogRepo := catalogmodule.NewRepository(database)
	learningRepo := learningmodule.NewRepository(database)
	testingRepo := testingmodule.NewRepository(database)
	certificatesRepo := certificatesmodule.NewRepository(database)
	courseRequestsRepo := courserequestsmodule.NewRepository(database)
	externalTrainingRepo := externaltrainingmodule.NewRepository(database)
	outlookRepo := outlookmodule.NewRepository(database)
	notificationsRepo := notificationsmodule.NewRepository(database)
	universityRepo := universitymodule.NewRepository(database)
	yougileRepo := yougilemodule.NewRepository(database)
	githubRepo := githubmodule.NewRepository(database)

	orgService := orgmodule.NewService(orgRepo)
	identityService := identitymodule.NewService(database, identityRepo, orgService, outboxService, queue, jwtManager, appClock, cfg.Auth.AccessTokenTTL, cfg.Auth.RefreshTokenTTL)
	adminService := adminmodule.NewService(database, identityRepo, orgService, appClock)
	catalogService := catalogmodule.NewService(catalogRepo, appClock)
	learningService := learningmodule.NewService(database, learningRepo, orgService, catalogService, outboxService, appClock)
	testingService := testingmodule.NewService(database, testingRepo, appClock)
	certificatesService := certificatesmodule.NewService(database, certificatesRepo, outboxService, appClock)
	courseRequestsService := courserequestsmodule.NewService(database, courseRequestsRepo, identityRepo, orgService, catalogService, learningRepo, certificatesRepo, appClock)
	externalTrainingService := externaltrainingmodule.NewService(database, externalTrainingRepo, identityRepo, orgService, outboxService, queue, appClock)
	outlookService := outlookmodule.NewService(database, outlookRepo, queue, appClock)
	notificationsService := notificationsmodule.NewService(notificationsRepo, appClock)
	universityService := universitymodule.NewService(universityRepo, appClock)
	analyticsService := analyticsmodule.NewService(database, queue)
	auditService := auditmodule.NewService(database)
	yougileService := yougilemodule.NewService(database, yougileRepo, queue, appClock)
	githubService := githubmodule.NewService(database, githubRepo, queue, appClock)
	healthService := corehealth.NewService(map[string]corehealth.CheckFunc{
		"postgres": database.PingContext,
	})

	registerWorkerHandlers(queue, log, yougileService, githubService)

	container := &Container{
		Config:                  cfg,
		Logger:                  log,
		DB:                      database,
		Clock:                   appClock,
		JWT:                     jwtManager,
		Outbox:                  outboxService,
		Queue:                   queue,
		Validator:               validate,
		HealthService:           healthService,
		OrgService:              orgService,
		IdentityService:         identityService,
		AdminService:            adminService,
		CatalogService:          catalogService,
		LearningService:         learningService,
		TestingService:          testingService,
		CertificatesService:     certificatesService,
		CourseRequestsService:   courseRequestsService,
		ExternalTrainingService: externalTrainingService,
		OutlookService:          outlookService,
		NotificationsService:    notificationsService,
		UniversityService:       universityService,
		AnalyticsService:        analyticsService,
		AuditService:            auditService,
		YougileService:          yougileService,
		GitHubService:           githubService,
		HealthHandler:           corehealth.NewHandler(healthService),
		IdentityHandler:         identitymodule.NewHandler(identityService, validate),
		AdminHandler:            adminmodule.NewHandler(adminService, validate),
		CatalogHandler:          catalogmodule.NewHandler(catalogService, validate),
		LearningHandler:         learningmodule.NewHandler(learningService, validate),
		TestingHandler:          testingmodule.NewHandler(testingService, validate),
		CertificatesHandler:     certificatesmodule.NewHandler(certificatesService, validate),
		CourseRequestsHandler:   courserequestsmodule.NewHandler(courseRequestsService, validate),
		ExternalTrainingHandler: externaltrainingmodule.NewHandler(externalTrainingService, validate),
		OutlookHandler:          outlookmodule.NewHandler(outlookService),
		NotificationsHandler:    notificationsmodule.NewHandler(notificationsService),
		UniversityHandler:       universitymodule.NewHandler(universityService, validate),
		AnalyticsHandler:        analyticsmodule.NewHandler(analyticsService),
		AuditHandler:            auditmodule.NewHandler(auditService),
		YougileHandler:          yougilemodule.NewHandler(yougileService, validate),
		GitHubHandler:           githubmodule.NewHandler(githubService, validate),
	}

	return container, nil
}

func (a *App) Run(ctx context.Context) error {
	serverErr := make(chan error, 1)
	go func() {
		a.container.Logger.Info("http server started", "addr", a.server.Addr)
		serverErr <- a.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			_ = a.container.DB.Close()
			return err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		_ = a.container.DB.Close()
		return err
	}

	_ = a.container.DB.Close()
	return nil
}

func registerWorkerHandlers(queue *platformworker.Queue, logger *slog.Logger, yougileService *yougilemodule.Service, githubService *githubmodule.Service) {
	queue.Register("send_password_reset", func(ctx context.Context, job platformworker.Job) error {
		logger.Info("process job", "job_type", job.JobType, "queue", job.Queue, "payload", string(job.Payload))
		return nil
	})
	queue.Register("outlook_sync", func(ctx context.Context, job platformworker.Job) error {
		logger.Info("process job", "job_type", job.JobType, "queue", job.Queue, "payload", string(job.Payload))
		return nil
	})
	queue.Register("outlook_create_event", func(ctx context.Context, job platformworker.Job) error {
		logger.Info("process job", "job_type", job.JobType, "queue", job.Queue, "payload", string(job.Payload))
		return nil
	})
	queue.Register("export_excel", func(ctx context.Context, job platformworker.Job) error {
		logger.Info("process job", "job_type", job.JobType, "queue", job.Queue, "payload", string(job.Payload))
		return nil
	})
	queue.Register("export_pdf", func(ctx context.Context, job platformworker.Job) error {
		logger.Info("process job", "job_type", job.JobType, "queue", job.Queue, "payload", string(job.Payload))
		return nil
	})
	queue.Register("yougile_sync", func(ctx context.Context, job platformworker.Job) error {
		return yougileService.ProcessSyncJob(ctx, job)
	})
	queue.Register("github_sync", func(ctx context.Context, job platformworker.Job) error {
		return githubService.ProcessSyncJob(ctx, job)
	})
}
