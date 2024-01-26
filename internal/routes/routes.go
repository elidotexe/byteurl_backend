package routes

import (
	"net/http"

	"github.com/elidotexe/backend_byteurl/internal/auth"
	"github.com/elidotexe/backend_byteurl/internal/config"
	"github.com/elidotexe/backend_byteurl/internal/handlers"
	"github.com/elidotexe/backend_byteurl/internal/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func SetupRoutes(app *config.AppConfig, authInstance *auth.Auth) http.Handler {
	mux := chi.NewRouter()

	authMiddleware := middlewares.NewAuthMiddleware(app, authInstance)
	mux.Use(authMiddleware.EnableCORS)
	mux.Use(middleware.Recoverer)

	mux.Get("/", handlers.Repo.Home)

	mux.Post("/login", handlers.Repo.Login)
	mux.Post("/signup", handlers.Repo.Signup)

	mux.Get("/redirect/{short}", handlers.Repo.RedirectToOriginalURL)
	mux.Post("/redirect/{short}", handlers.Repo.CreateRedirectHistory)

	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(authMiddleware.RequireAuth)

		mux.Get("/users/{id}", handlers.Repo.GetUserName)
		mux.Patch("/users/{id}", handlers.Repo.UpdateUserName)

		mux.Get("/users/{id}/links", handlers.Repo.AllLinks)
		mux.Put("/users/{id}/links/0", handlers.Repo.CreateLink)
		mux.Get("/users/{id}/links/{linkID}", handlers.Repo.SingleLink)
		mux.Patch("/users/{id}/links/{linkID}", handlers.Repo.UpdateLink)
		mux.Delete("/users/{id}/links/{linkID}", handlers.Repo.DeleteLink)
	})

	apiRouter := chi.NewRouter()
	apiRouter.Mount("/api", mux)

	return apiRouter
}
