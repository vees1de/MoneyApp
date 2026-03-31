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
			r.Use(middleware.RBAC("settings.manage"))

			r.Route("/integrations/yougile", func(r chi.Router) {
				r.Post("/connections", container.YougileHandler.CreateConnection)
				r.Post("/connections/connect", container.YougileHandler.ConnectConnection)
				r.Post("/connections/test-key", container.YougileHandler.TestKey)
				r.Post("/connections/create-key", container.YougileHandler.CreateKey)
				r.Post("/discover-companies", container.YougileHandler.DiscoverCompanies)
				r.Get("/connections", container.YougileHandler.ListConnections)
				r.Get("/connections/{id}", container.YougileHandler.GetConnection)
				r.Patch("/connections/{id}", container.YougileHandler.UpdateConnection)
				r.Delete("/connections/{id}", container.YougileHandler.DeleteConnection)
				r.Post("/connections/{id}/import/users", container.YougileHandler.ImportUsers)
				r.Post("/connections/{id}/import/structure", container.YougileHandler.ImportStructure)
				r.Get("/connections/{id}/users", container.YougileHandler.ListUsers)
				r.Get("/connections/{id}/projects", container.YougileHandler.ListProjects)
				r.Get("/connections/{id}/boards", container.YougileHandler.ListBoards)
				r.Get("/connections/{id}/columns", container.YougileHandler.ListColumns)
				r.Get("/connections/{id}/tasks", container.YougileHandler.ListTasks)
				r.Post("/connections/{id}/mappings/auto-match", container.YougileHandler.AutoMatch)
				r.Get("/connections/{id}/mappings", container.YougileHandler.ListMappings)
				r.Post("/connections/{id}/mappings", container.YougileHandler.CreateMapping)
				r.Delete("/connections/{id}/mappings/{mappingId}", container.YougileHandler.DeleteMapping)
				r.Post("/connections/{id}/sync", container.YougileHandler.StartSync)
				r.Get("/sync-jobs/{jobId}", container.YougileHandler.GetSyncJob)
				r.Post("/connections/{id}/sync/backfill", container.YougileHandler.Backfill)
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthRequired(container.JWT))
			r.Use(middleware.RBAC("settings.manage"))

			r.Route("/integrations/github", func(r chi.Router) {
				r.Get("/connections", container.GitHubHandler.ListConnections)
				r.Post("/connections", container.GitHubHandler.CreateConnection)
				r.Post("/connections/test", container.GitHubHandler.TestConnection)
				r.Get("/connections/{connectionId}", container.GitHubHandler.GetConnection)
				r.Patch("/connections/{connectionId}", container.GitHubHandler.UpdateConnection)
				r.Delete("/connections/{connectionId}", container.GitHubHandler.DeleteConnection)

				r.Post("/connections/{connectionId}/import/users", container.GitHubHandler.ImportUsers)
				r.Post("/connections/{connectionId}/import/repos", container.GitHubHandler.ImportRepos)
				r.Post("/connections/{connectionId}/import/languages", container.GitHubHandler.ImportLanguages)
				r.Post("/connections/{connectionId}/sync", container.GitHubHandler.StartSync)
				r.Get("/sync-jobs/{jobId}", container.GitHubHandler.GetSyncJob)

				r.Get("/connections/{connectionId}/mappings", container.GitHubHandler.ListMappings)
				r.Post("/connections/{connectionId}/mappings", container.GitHubHandler.CreateMapping)
				r.Post("/connections/{connectionId}/mappings/auto-match", container.GitHubHandler.AutoMatchMappings)
				r.Delete("/connections/{connectionId}/mappings/{mappingId}", container.GitHubHandler.DeleteMapping)

				r.Get("/users", container.GitHubHandler.ListGithubUsers)
				r.Get("/repositories", container.GitHubHandler.ListRepositories)
				r.Get("/repositories/{repoId}", container.GitHubHandler.GetRepository)
				r.Get("/repositories/{repoId}/languages", container.GitHubHandler.GetRepositoryLanguages)
				r.Get("/repositories/{repoId}/contributors", container.GitHubHandler.GetRepositoryContributors)

				r.Get("/employees/{employeeUserId}/profile", container.GitHubHandler.GetEmployeeProfile)
				r.Get("/employees/{employeeUserId}/languages", container.GitHubHandler.GetEmployeeLanguages)
				r.Get("/employees/{employeeUserId}/stats", container.GitHubHandler.GetEmployeeStats)
				r.Get("/employees/{employeeUserId}/activity", container.GitHubHandler.GetEmployeeActivity)

				r.Get("/analytics/team", container.GitHubHandler.GetTeamAnalytics)
				r.Get("/analytics/languages", container.GitHubHandler.GetLanguageAnalytics)
				r.Get("/analytics/top-languages", container.GitHubHandler.GetTopLanguages)
				r.Get("/analytics/repository-health", container.GitHubHandler.GetRepositoryHealth)
				r.Get("/analytics/repository-ownership", container.GitHubHandler.GetRepositoryOwnership)
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

			r.Route("/learning-plan", func(r chi.Router) {
				r.Get("/my", container.LearningPlanHandler.MyPlan)
			})

			r.Route("/recommendations", func(r chi.Router) {
				r.Get("/courses", container.LearningPlanHandler.RecommendedCourses)
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

			r.Route("/intakes", func(r chi.Router) {
				r.With(middleware.RBAC("intakes.manage")).Post("/", container.CourseIntakesHandler.CreateIntake)
				r.Get("/", container.CourseIntakesHandler.ListIntakes)
				r.Get("/{id}", container.CourseIntakesHandler.GetIntake)
				r.With(middleware.RBAC("intakes.manage")).Patch("/{id}", container.CourseIntakesHandler.UpdateIntake)
				r.With(middleware.RBAC("intakes.manage")).Post("/{id}/close", container.CourseIntakesHandler.CloseIntake)
				r.With(middleware.RBAC("intakes.manage")).Post("/{id}/payment-status", container.CourseIntakesHandler.UpdatePaymentStatusByIntake)
				r.Get("/{intakeId}/applications", container.CourseIntakesHandler.ListApplicationsByIntake)
			})

			r.Route("/applications", func(r chi.Router) {
				r.Post("/", container.CourseIntakesHandler.Apply)
				r.Get("/my", container.CourseIntakesHandler.ListMyApplications)
				r.Get("/pending-manager", container.CourseIntakesHandler.ListPendingManagerApprovals)
				r.Get("/{id}", container.CourseIntakesHandler.GetApplication)
				r.Post("/{id}/approve-manager", container.CourseIntakesHandler.ApproveByManager)
				r.Post("/{id}/reject-manager", container.CourseIntakesHandler.RejectByManager)
				r.With(middleware.RBAC("intakes.manage")).Post("/{id}/approve-hr", container.CourseIntakesHandler.ApproveByHR)
				r.With(middleware.RBAC("intakes.manage")).Post("/{id}/reject-hr", container.CourseIntakesHandler.RejectByHR)
				r.Post("/{id}/withdraw", container.CourseIntakesHandler.WithdrawApplication)
				r.With(middleware.RBAC("intakes.manage")).Post("/{id}/enroll", container.CourseIntakesHandler.EnrollApplication)
			})

			r.Route("/suggestions", func(r chi.Router) {
				r.Post("/", container.CourseIntakesHandler.CreateSuggestion)
				r.Get("/", container.CourseIntakesHandler.ListSuggestions)
				r.Get("/my", container.CourseIntakesHandler.ListMySuggestions)
				r.Get("/{id}", container.CourseIntakesHandler.GetSuggestion)
				r.With(middleware.RBAC("intakes.manage")).Post("/{id}/approve", container.CourseIntakesHandler.ApproveSuggestion)
				r.With(middleware.RBAC("intakes.manage")).Post("/{id}/reject", container.CourseIntakesHandler.RejectSuggestion)
				r.With(middleware.RBAC("intakes.manage")).Post("/{id}/open-intake", container.CourseIntakesHandler.OpenIntakeFromSuggestion)
			})

			r.Route("/course-requests", func(r chi.Router) {
				r.Get("/", container.CourseRequestsHandler.List)
				r.Get("/export/excel", container.CourseRequestsHandler.ExportExcel)
				r.Post("/", container.CourseRequestsHandler.Create)
				r.Get("/{id}", container.CourseRequestsHandler.Get)
				r.Post("/{id}/approve-manager", container.CourseRequestsHandler.ApproveManager)
				r.Post("/{id}/approve-hr", container.CourseRequestsHandler.ApproveHR)
				r.Post("/{id}/reject", container.CourseRequestsHandler.Reject)
				r.Post("/{id}/cancel", container.CourseRequestsHandler.Cancel)
				r.Post("/{id}/start", container.CourseRequestsHandler.Start)
				r.Post("/{id}/complete", container.CourseRequestsHandler.Complete)
				r.Post("/{id}/certificate/upload", container.CourseRequestsHandler.UploadCertificate)
				r.Post("/{id}/certificate/approve", container.CourseRequestsHandler.ApproveCertificate)
				r.Post("/{id}/certificate/reject", container.CourseRequestsHandler.RejectCertificate)
			})

			r.Route("/external-requests", func(r chi.Router) {
				r.Post("/", container.ExternalTrainingHandler.CreateRequest)
				r.Get("/", container.ExternalTrainingHandler.List)
				r.Get("/my", container.ExternalTrainingHandler.ListMine)
				r.Get("/pending-approvals", container.ExternalTrainingHandler.PendingApprovals)
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

			r.Route("/calendar", func(r chi.Router) {
				r.Get("/events/upcoming", container.CalendarHandler.Upcoming)
			})

			r.Route("/dashboard", func(r chi.Router) {
				r.Get("/employee", container.DashboardAPIHandler.Employee)
				r.Get("/manager", container.DashboardAPIHandler.Manager)
			})

			r.Route("/jira", func(r chi.Router) {
				r.Get("/board-summary", container.BoardSummaryHandler.Summary)
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

			r.Route("/reports/sources", func(r chi.Router) {
				r.Get("/", container.SmartExportHandler.Sources)
			})
			r.Route("/reports/smart-export", func(r chi.Router) {
				r.Post("/", container.SmartExportHandler.SmartExport)
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
