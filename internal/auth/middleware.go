package auth

import (
	"net/http"

	"github.com/elidotexe/backend_byteurl/internal/config"
)

type AuthMiddleware struct {
	app *config.AppConfig
}

func NewAuthMiddleware(app *config.AppConfig) *AuthMiddleware {
	return &AuthMiddleware{
		app: app,
	}
}

func (a *AuthMiddleware) EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			w.WriteHeader(http.StatusNoContent)
			return
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

// func (a *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		_, _, err := a.Aut.GetTokenFromHeaderAndVerify(w, r)
// 		if err != nil {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			return
// 		}
//
// 		next.ServeHTTP(w, r)
// 	})
// }
