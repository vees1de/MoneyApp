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
			r.Post("/telegram", container.AuthHandler.TelegramLogin)
			r.Post("/yandex", container.AuthHandler.YandexLogin)
			r.Post("/refresh", container.AuthHandler.Refresh)
			r.Post("/logout", container.AuthHandler.Logout)

			r.Group(func(r chi.Router) {
				r.Use(middleware.AuthRequired(container.JWT))
				r.Get("/me", container.AuthHandler.Me)
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthRequired(container.JWT))

			r.Get("/users/me", container.UserHandler.Me)
			r.Patch("/users/preferences", container.UserHandler.UpdatePreferences)

			r.Route("/accounts", func(r chi.Router) {
				r.Get("/", container.AccountHandler.List)
				r.Post("/", container.AccountHandler.Create)
				r.Get("/{id}", container.AccountHandler.Get)
				r.Patch("/{id}", container.AccountHandler.Update)
			})

			r.Route("/finance/categories", func(r chi.Router) {
				r.Get("/", container.CategoryHandler.List)
				r.Post("/", container.CategoryHandler.Create)
				r.Patch("/{id}", container.CategoryHandler.Update)
				r.Delete("/{id}", container.CategoryHandler.Delete)
			})

			r.Route("/finance/transactions", func(r chi.Router) {
				r.Get("/", container.TransactionHandler.List)
				r.Post("/", container.TransactionHandler.Create)
				r.Get("/{id}", container.TransactionHandler.Get)
				r.Patch("/{id}", container.TransactionHandler.Update)
				r.Delete("/{id}", container.TransactionHandler.Delete)
				r.Post("/{id}/restore", container.TransactionHandler.Restore)
			})

			r.Route("/finance/transfers", func(r chi.Router) {
				r.Post("/", container.TransferHandler.Create)
				r.Patch("/{id}", container.TransferHandler.Update)
				r.Delete("/{id}", container.TransferHandler.Delete)
				r.Post("/{id}/restore", container.TransferHandler.Restore)
			})

			r.Route("/savings", func(r chi.Router) {
				r.Get("/goals", container.SavingsHandler.ListGoals)
				r.Post("/goals", container.SavingsHandler.CreateGoal)
				r.Patch("/goals/{id}", container.SavingsHandler.UpdateGoal)
				r.Get("/summary", container.SavingsHandler.Summary)
			})

			r.Route("/reviews/weekly", func(r chi.Router) {
				r.Get("/current", container.ReviewHandler.Current)
				r.Post("/{id}/submit-balance", container.ReviewHandler.SubmitBalance)
				r.Post("/{id}/resolve", container.ReviewHandler.Resolve)
				r.Post("/{id}/skip", container.ReviewHandler.Skip)
			})

			r.Get("/dashboard/finance", container.DashboardHandler.Finance)

			r.Route("/links", func(r chi.Router) {
				r.Post("/", container.LinksHandler.Create)
				r.Get("/by-entity", container.LinksHandler.ByEntity)
			})
		})
	})

	if frontend := newFrontendHandler(container.Config.HTTP.FrontendDistDir); frontend != nil {
		router.NotFound(frontend.ServeHTTP)
	}

	return router
}
