package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/elidotexe/backend_byteurl/internal/auth"
	"github.com/elidotexe/backend_byteurl/internal/config"
	"github.com/elidotexe/backend_byteurl/internal/driver"
	"github.com/elidotexe/backend_byteurl/internal/models"
	"github.com/elidotexe/backend_byteurl/internal/repository"
	"github.com/elidotexe/backend_byteurl/internal/repository/dbrepo"
	"github.com/elidotexe/backend_byteurl/internal/utils"
	"github.com/golang-jwt/jwt/v4"
)

var Repo *Repository

type Repository struct {
	App  *config.AppConfig
	DB   repository.DatabaseRepo
	Auth *auth.Auth
}

func NewRepo(a *config.AppConfig, db *driver.DB, authInstance *auth.Auth) *Repository {
	return &Repository{
		App:  a,
		DB:   dbrepo.NewPostgresRepo(db.Gorm, a),
		Auth: authInstance,
	}
}

func NewHandlers(r *Repository) {
	Repo = r
}

func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	var payload = struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Version string `json:"version"`
	}{
		Status:  "active",
		Message: "Welcome to the ByteURL API 🛸",
		Version: "1.0.0",
	}

	_ = utils.WriteJSON(w, http.StatusOK, payload)
}

func (m *Repository) Login(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := utils.ReadJSON(w, r, &payload)
	if err != nil {
		utils.ErrorJSON(w, err)
		return
	}

	if !utils.IsValidEmail(payload.Email) {
		utils.ErrorJSON(w, errors.New("invalid email address"), http.StatusBadRequest)
		return
	}

	userExists, err := m.DB.UserExists(payload.Email)
	if err != nil {
		utils.ErrorJSON(w, err)
		return
	}
	if !userExists {
		utils.ErrorJSON(w, errors.New("user does not exist"), http.StatusBadRequest)
		return
	}

	user, err := m.DB.GetUserByEmail(payload.Email)
	if err != nil {
		utils.ErrorJSON(w, errors.New("invalid email or password"), http.StatusBadRequest)
		return
	}

	valid, err := user.PasswordMathes(payload.Password)
	if !valid || err != nil {
		utils.ErrorJSON(w, errors.New("invalid email or password"), http.StatusBadRequest)
		return
	}

	u := auth.JWTUser{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	token, err := m.Auth.GenerateTokenPair(&u)
	if err != nil {
		utils.ErrorJSON(w, err)
		return
	}

	u.Token = token

	refreshCookie := m.Auth.GetRefreshCookie(u.Token)
	http.SetCookie(w, refreshCookie)

	response := struct {
		User auth.JWTUser `json:"user"`
	}{
		User: u,
	}

	_ = utils.WriteJSON(w, http.StatusOK, response)
}

func (m *Repository) Signup(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := utils.ReadJSON(w, r, &payload)
	if err != nil {
		utils.ErrorJSON(w, err)
		return
	}

	if len(payload.Name) < 3 {
		utils.ErrorJSON(w, errors.New("name must be at least 3 characters"), http.StatusBadRequest)
		return
	}

	if !utils.IsValidEmail(payload.Email) {
		utils.ErrorJSON(w, errors.New("invalid email address"), http.StatusBadRequest)
		return
	}

	if len(payload.Password) < 8 {
		utils.ErrorJSON(w, errors.New("password must be at least 8 characters"), http.StatusBadRequest)
		return
	}

	hashedPassword, err := models.HashPassword(payload.Password)
	if err != nil {
		utils.ErrorJSON(w, err)
		return
	}

	newUser := &models.User{
		Name:      payload.Name,
		Email:     payload.Email,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	var userExists bool

	userExists, err = m.DB.UserExists(newUser.Email)
	if err != nil {
		utils.ErrorJSON(w, err)
		return
	}
	if userExists {
		utils.ErrorJSON(w, errors.New("user already exists"), http.StatusConflict)
		return
	}

	err = m.DB.CreateUser(newUser)
	if err != nil {
		utils.ErrorJSON(w, err)
		return
	}

	u := auth.JWTUser{
		ID:    newUser.ID,
		Name:  newUser.Name,
		Email: newUser.Email,
	}

	token, err := m.Auth.GenerateTokenPair(&u)
	if err != nil {
		utils.ErrorJSON(w, err)
		return
	}

	u.Token = token

	refreshCookie := m.Auth.GetRefreshCookie(token)
	http.SetCookie(w, refreshCookie)

	response := struct {
		User auth.JWTUser `json:"user"`
	}{
		User: u,
	}

	_ = utils.WriteJSON(w, http.StatusOK, response)
}

func (m *Repository) RefreshToken(w http.ResponseWriter, r *http.Request) {
	for _, cookie := range r.Cookies() {
		if cookie.Name == m.Auth.CookieName {
			claims := &auth.Claims{}
			refreshToken := cookie.Value

			// parse the token to get the claims
			_, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(m.Auth.Secret), nil
			})
			if err != nil {
				utils.ErrorJSON(w, errors.New("unauthorized"), http.StatusUnauthorized)
				return
			}

			// get the user id from the token claims
			userID, err := strconv.Atoi(claims.Subject)
			if err != nil {
				utils.ErrorJSON(w, errors.New("unknown user"), http.StatusUnauthorized)
				return
			}

			user, err := m.DB.GetUserByID(userID)
			if err != nil {
				utils.ErrorJSON(w, errors.New("unknown user"), http.StatusUnauthorized)
				return
			}

			u := auth.JWTUser{
				ID:    user.ID,
				Email: user.Email,
			}

			token, err := m.Auth.GenerateTokenPair(&u)
			if err != nil {
				utils.ErrorJSON(w, errors.New("error generating token"), http.StatusUnauthorized)
				return
			}

			http.SetCookie(w, m.Auth.GetRefreshCookie(token))

			utils.WriteJSON(w, http.StatusOK, token)
		}
	}
}

func (m *Repository) UpdateUserName(w http.ResponseWriter, r *http.Request) {
	// Get the user id from the url
	path := strings.Split(r.URL.Path, "/")
	id := path[len(path)-1]

	userID, err := strconv.Atoi(id)
	if err != nil {
		utils.ErrorJSON(w, errors.New("invalid user id"), http.StatusBadRequest)
		return
	}

	var payload struct {
		Name string `json:"name"`
	}

	err = utils.ReadJSON(w, r, &payload)
	if err != nil {
		utils.ErrorJSON(w, errors.New("invalid request payload"), http.StatusBadRequest)
		return
	}

	if payload.Name == "" {
		utils.ErrorJSON(w, errors.New("name cannot be empty"), http.StatusBadRequest)
		return
	}

	if len(payload.Name) < 3 || len(payload.Name) > 32 {
		utils.ErrorJSON(w, errors.New("name must be between 3 and 32 characters"), http.StatusBadRequest)
		return
	}

	updatedUser := &models.User{ID: userID, Name: payload.Name}
	updatedUser, err = m.DB.GetUserByID(userID)
	if err != nil {
		utils.ErrorJSON(w, errors.New("failed to retrieve updated user"), http.StatusInternalServerError)
		return
	}

	if updatedUser.Name == payload.Name {
		utils.ErrorJSON(w, errors.New("name is the same"), http.StatusBadRequest)
		return
	}

	err = m.DB.UpdateUserNameByID(userID, updatedUser)
	if err != nil {
		utils.ErrorJSON(w, errors.New("failed to update user"), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, updatedUser.Name)
}
