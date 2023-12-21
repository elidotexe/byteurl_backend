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

	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(authMiddleware.RequireAuth)

		mux.Patch("/users/{id}", handlers.Repo.UpdateUserName)

		mux.Get("/users/{id}/links", handlers.Repo.AllUserLinks)
	})

	apiRouter := chi.NewRouter()
	apiRouter.Mount("/api", mux)

	return apiRouter
}
