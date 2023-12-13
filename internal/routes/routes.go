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
	mux.Get("/refresh", handlers.Repo.RefreshToken)
	mux.Post("/users", handlers.Repo.UserForEdit)

	// mux.Route("/admin", func(r chi.Router) {
	// 	// TODO: Implement admin routes
	// }

	apiRouter := chi.NewRouter()
	apiRouter.Mount("/api", mux)

	return apiRouter
}
