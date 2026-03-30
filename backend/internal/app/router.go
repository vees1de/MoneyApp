package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"moneyapp/backend/internal/docs"
	"moneyapp/backend/internal/middleware"
)

func NewRouter(container *Container) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.CORSLocalhost4200)
	router.Use(middleware.Recovery(container.Logger))
	router.Use(middleware.Logging(container.Logger))

	router.Get("/healthz", container.HealthHandler.Live)
	router.Get("/readyz", container.HealthHandler.Ready)
	router.Get("/openapi.yaml", docs.OpenAPI)
	router.Get("/swagger.json", docs.SwaggerJSON)
	router.Get("/swagger", docs.SwaggerUI("/openapi.yaml"))
	router.Get("/swagger/", docs.SwaggerUI("/openapi.yaml"))

	router.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", container.IdentityHandler.Register)
			r.Post("/login", container.IdentityHandler.Login)
			r.Post("/refresh", container.IdentityHandler.Refresh)
			r.Post("/logout", container.IdentityHandler.Logout)
			r.Post("/forgot-password", container.IdentityHandler.ForgotPassword)
			r.Post("/reset-password", container.IdentityHandler.ResetPassword)

			r.Group(func(r chi.Router) {
				r.Use(middleware.AuthRequired(container.JWT))
				r.Get("/me", container.IdentityHandler.Me)
			})
		})

		r.Route("/integrations/outlook", func(r chi.Router) {
			r.Get("/callback", container.OutlookHandler.Callback)

			r.Group(func(r chi.Router) {
				r.Use(middleware.AuthRequired(container.JWT))
				r.Get("/connect", container.OutlookHandler.Connect)
				r.Post("/disconnect", container.OutlookHandler.Disconnect)
				r.Get("/status", container.OutlookHandler.Status)
				r.Post("/sync", container.OutlookHandler.Sync)
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthRequired(container.JWT))

			r.Route("/admin", func(r chi.Router) {
				r.With(middleware.RBAC("users.read")).Get("/users", container.AdminHandler.ListUsers)
				r.With(middleware.RBAC("users.write")).Post("/users", container.AdminHandler.CreateUser)
				r.With(middleware.RBAC("users.read")).Get("/users/{id}", container.AdminHandler.GetUser)
				r.With(middleware.RBAC("users.write")).Patch("/users/{id}", container.AdminHandler.UpdateUser)
				r.With(middleware.RBAC("roles.manage")).Post("/users/{id}/roles", container.AdminHandler.AssignRole)
				r.With(middleware.RBAC("roles.manage")).Delete("/users/{id}/roles/{roleId}", container.AdminHandler.RemoveRole)
				r.With(middleware.RBAC("roles.manage")).Get("/roles", container.AdminHandler.ListRoles)
				r.With(middleware.RBAC("roles.manage")).Get("/permissions", container.AdminHandler.ListPermissions)
			})

			r.Route("/courses", func(r chi.Router) {
				r.Get("/", container.CatalogHandler.ListCourses)
				r.With(middleware.RBAC("courses.write")).Post("/", container.CatalogHandler.CreateCourse)
				r.Get("/{id}", container.CatalogHandler.GetCourse)
				r.With(middleware.RBAC("courses.write")).Patch("/{id}", container.CatalogHandler.UpdateCourse)
				r.With(middleware.RBAC("courses.write")).Post("/{id}/publish", container.CatalogHandler.PublishCourse)
				r.With(middleware.RBAC("courses.write")).Post("/{id}/archive", container.CatalogHandler.ArchiveCourse)
				r.Get("/{id}/materials", container.CatalogHandler.ListMaterials)
				r.With(middleware.RBAC("courses.write")).Post("/{id}/materials", container.CatalogHandler.CreateMaterial)
			})

			r.Route("/assignments", func(r chi.Router) {
				r.With(middleware.RBAC("courses.assign")).Post("/", container.LearningHandler.CreateAssignment)
				r.With(middleware.RBAC("courses.assign")).Get("/", container.LearningHandler.ListAssignments)
			})

			r.Route("/enrollments", func(r chi.Router) {
				r.Get("/my", container.LearningHandler.ListMyEnrollments)
				r.Get("/{id}", container.LearningHandler.GetEnrollment)
				r.Post("/{id}/start", container.LearningHandler.StartEnrollment)
				r.Post("/{id}/progress", container.LearningHandler.ProgressEnrollment)
				r.Post("/{id}/complete", container.LearningHandler.CompleteEnrollment)
			})

			r.Route("/tests", func(r chi.Router) {
				r.With(middleware.RBAC("courses.write")).Post("/", container.TestingHandler.CreateTest)
				r.Get("/{id}", container.TestingHandler.GetTest)
				r.Post("/{id}/attempts", container.TestingHandler.StartAttempt)
				r.Get("/{id}/results", container.TestingHandler.ListResults)
			})
			r.Route("/test-attempts", func(r chi.Router) {
				r.Post("/{id}/answers", container.TestingHandler.SaveAnswers)
				r.Post("/{id}/submit", container.TestingHandler.SubmitAttempt)
			})

			r.Route("/certificates", func(r chi.Router) {
				r.Post("/upload", container.CertificatesHandler.Upload)
				r.Get("/my", container.CertificatesHandler.ListMine)
				r.With(middleware.RBAC("certificates.verify")).Post("/{id}/verify", container.CertificatesHandler.Verify)
				r.With(middleware.RBAC("certificates.verify")).Post("/{id}/reject", container.CertificatesHandler.Reject)
			})

			r.Route("/external-requests", func(r chi.Router) {
				r.Post("/", container.ExternalTrainingHandler.CreateRequest)
				r.Get("/my", container.ExternalTrainingHandler.ListMine)
				r.Get("/{id}", container.ExternalTrainingHandler.GetRequest)
				r.Patch("/{id}", container.ExternalTrainingHandler.UpdateRequest)
				r.Post("/{id}/submit", container.ExternalTrainingHandler.Submit)
				r.Post("/{id}/approve", container.ExternalTrainingHandler.Approve)
				r.Post("/{id}/reject", container.ExternalTrainingHandler.Reject)
				r.Post("/{id}/request-revision", container.ExternalTrainingHandler.RequestRevision)
				r.Post("/{id}/upload-certificate", container.ExternalTrainingHandler.UploadCertificate)
			})

			r.Route("/approval-workflows", func(r chi.Router) {
				r.With(middleware.RBAC("settings.manage")).Post("/", container.ExternalTrainingHandler.CreateWorkflow)
				r.With(middleware.RBAC("settings.manage")).Get("/", container.ExternalTrainingHandler.ListWorkflows)
			})

			r.Route("/budget-limits", func(r chi.Router) {
				r.With(middleware.RBAC("settings.manage")).Post("/", container.ExternalTrainingHandler.CreateBudgetLimit)
				r.With(middleware.RBAC("settings.manage")).Get("/", container.ExternalTrainingHandler.ListBudgetLimits)
			})

			r.Route("/notifications", func(r chi.Router) {
				r.Get("/", container.NotificationsHandler.List)
				r.Post("/{id}/read", container.NotificationsHandler.MarkRead)
				r.Post("/read-all", container.NotificationsHandler.MarkAllRead)
			})

			r.Route("/programs", func(r chi.Router) {
				r.With(middleware.RBAC("programs.manage")).Post("/", container.UniversityHandler.CreateProgram)
				r.Get("/", container.UniversityHandler.ListPrograms)
				r.Get("/{id}", container.UniversityHandler.GetProgram)
				r.With(middleware.RBAC("programs.manage")).Post("/{id}/groups", container.UniversityHandler.CreateGroup)
			})

			r.Route("/groups", func(r chi.Router) {
				r.With(middleware.RBAC("programs.manage")).Post("/{id}/participants", container.UniversityHandler.AddParticipant)
				r.With(middleware.RBAC("programs.manage")).Post("/{id}/sessions", container.UniversityHandler.CreateSession)
			})

			r.Route("/sessions", func(r chi.Router) {
				r.With(middleware.RBAC("programs.manage")).Post("/{id}/trainer-feedback", container.UniversityHandler.TrainerFeedback)
				r.Post("/{id}/participant-feedback", container.UniversityHandler.ParticipantFeedback)
			})

			r.Route("/analytics", func(r chi.Router) {
				r.With(middleware.RBAC("analytics.read_hr")).Get("/dashboard/hr", container.AnalyticsHandler.DashboardHR)
				r.Get("/dashboard/manager", container.AnalyticsHandler.DashboardManager)
				r.Get("/compliance", container.AnalyticsHandler.Compliance)
				r.Get("/external-requests", container.AnalyticsHandler.ExternalRequests)
				r.Get("/budget", container.AnalyticsHandler.Budget)
				r.Get("/trainers", container.AnalyticsHandler.Trainers)
			})

			r.Route("/reports/export", func(r chi.Router) {
				r.Get("/excel", container.AnalyticsHandler.ExportExcel)
				r.Get("/pdf", container.AnalyticsHandler.ExportPDF)
			})

			r.With(middleware.RBAC("audit.read")).Get("/audit-logs", container.AuditHandler.List)
		})
	})

	if frontend := newFrontendHandler(container.Config.HTTP.FrontendDistDir); frontend != nil {
		router.NotFound(frontend.ServeHTTP)
	}

	return router
}
