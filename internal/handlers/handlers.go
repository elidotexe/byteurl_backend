package handlers

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/elidotexe/backend_byteurl/internal/auth"
	"github.com/elidotexe/backend_byteurl/internal/config"
	"github.com/elidotexe/backend_byteurl/internal/driver"
	"github.com/elidotexe/backend_byteurl/internal/models"
	"github.com/elidotexe/backend_byteurl/internal/repository"
	"github.com/elidotexe/backend_byteurl/internal/repository/dbrepo"
	"github.com/elidotexe/backend_byteurl/internal/utils"
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

func (m *Repository) UpdateUserName(w http.ResponseWriter, r *http.Request) {
	id, _ := utils.GetIDFromURL(r.URL.Path)

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

func (m *Repository) AllLinks(w http.ResponseWriter, r *http.Request) {
	id, _ := utils.GetIDFromURL(r.URL.Path)
	userID, err := strconv.Atoi(id)
	if err != nil {
		utils.ErrorJSON(w, errors.New("invalid user id"), http.StatusBadRequest)
		return
	}

	links, err := m.DB.GetAllLinks(userID)
	if err != nil {
		utils.ErrorJSON(w, errors.New("failed to retrieve user links"), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, links)
}

func (m *Repository) CreateLink(w http.ResponseWriter, r *http.Request) {
	pathUserID, _ := utils.GetIDFromURL(r.URL.Path)
	if pathUserID == "" {
		utils.ErrorJSON(w, errors.New("pathUserID is empty"), http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(pathUserID)
	if err != nil {
		utils.ErrorJSON(w, errors.New("invalid user id"), http.StatusBadRequest)
		return
	}

	var payload struct {
		Title       string `json:"title"`
		OriginalURL string `json:"originalUrl"`
	}

	err = utils.ReadJSON(w, r, &payload)
	if err != nil {
		utils.ErrorJSON(w, errors.New("invalid request payload"), http.StatusBadRequest)
		return
	}

	if len(payload.Title) < 3 {
		utils.ErrorJSON(w, errors.New("title must be at least 3 characters"), http.StatusBadRequest)
		return
	}

	if payload.OriginalURL == "" {
		utils.ErrorJSON(w, errors.New("originalUrl cannot be empty"), http.StatusBadRequest)
		return
	}
	_, err = url.Parse(payload.OriginalURL)
	if err != nil {
		utils.ErrorJSON(w, errors.New("invalid originalUrl"), http.StatusBadRequest)
		return
	}

	randString := utils.GenerateRandomString(5)

	newLink := models.Link{
		UserID:      userID,
		Title:       payload.Title,
		OriginalURL: payload.OriginalURL,
		ShortenURL:  "https://byteurl.io/" + randString,
		Clicks:      0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	insertLink, err := m.DB.InsertLink(&newLink)
	if err != nil {
		utils.ErrorJSON(w, errors.New("failed to insert link"), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, insertLink)
}

func (m *Repository) UpdateLink(w http.ResponseWriter, r *http.Request) {
	pathUserID, pathLinkID := utils.GetIDFromURL(r.URL.Path)
	if pathUserID == "" {
		utils.ErrorJSON(w, errors.New("pathUserID is empty"), http.StatusBadRequest)
		return
	}

	if pathLinkID == "" {
		utils.ErrorJSON(w, errors.New("pathLinkID is empty"), http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(pathUserID)
	if err != nil {
		utils.ErrorJSON(w, errors.New("invalid user id"), http.StatusBadRequest)
		return
	}

	pathID, err := strconv.Atoi(pathLinkID)
	if err != nil {
		utils.ErrorJSON(w, errors.New("invalid user id"), http.StatusBadRequest)
		return
	}

	var payload struct {
		Title       string `json:"title"`
		OriginalURL string `json:"originalUrl"`
	}

	err = utils.ReadJSON(w, r, &payload)
	if err != nil {
		utils.ErrorJSON(w, errors.New("invalid request payload"), http.StatusBadRequest)
		return
	}

	if len(payload.Title) < 3 {
		utils.ErrorJSON(w, errors.New("title must be at least 3 characters"), http.StatusBadRequest)
		return
	}

	if payload.OriginalURL == "" {
		utils.ErrorJSON(w, errors.New("originalUrl cannot be empty"), http.StatusBadRequest)
		return
	}
	_, err = url.Parse(payload.OriginalURL)
	if err != nil {
		utils.ErrorJSON(w, errors.New("invalid originalUrl"), http.StatusBadRequest)
		return
	}

	link, err := m.DB.GetLink(userID, pathID)
	if err != nil {
		utils.ErrorJSON(w, errors.New("failed to retrieve link"), http.StatusInternalServerError)
		return
	}

	link.Title = payload.Title
	link.OriginalURL = payload.OriginalURL
	link.UpdatedAt = time.Now()

	updatedLink, err := m.DB.UpdateLink(link)
	if err != nil {
		utils.ErrorJSON(w, errors.New("failed to update link"), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, updatedLink)
}

func (m *Repository) SingleLink(w http.ResponseWriter, r *http.Request) {
	pathUserID, pathLinkID := utils.GetIDFromURL(r.URL.Path)
	if pathUserID == "" || pathLinkID == "" {
		utils.ErrorJSON(w, errors.New("pathUserID is empty"), http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(pathUserID)
	if err != nil {
		utils.ErrorJSON(w, errors.New("invalid user id"), http.StatusBadRequest)
		return
	}

	linkID, err := strconv.Atoi(pathLinkID)
	if err != nil {
		utils.ErrorJSON(w, errors.New("invalid user id"), http.StatusBadRequest)
		return
	}

	link, err := m.DB.GetLink(userID, linkID)
	if err != nil {
		utils.ErrorJSON(w, errors.New("failed to retrieve link"), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, link)
}

func (m *Repository) DeleteLink(w http.ResponseWriter, r *http.Request) {
	pathUserID, pathLinkID := utils.GetIDFromURL(r.URL.Path)
	if pathUserID == "" || pathLinkID == "" {
		utils.ErrorJSON(w, errors.New("pathUserID is empty"), http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(pathUserID)
	if err != nil {
		utils.ErrorJSON(w, errors.New("invalid user id"), http.StatusBadRequest)
		return
	}

	linkID, err := strconv.Atoi(pathLinkID)
	if err != nil {
		utils.ErrorJSON(w, errors.New("invalid user id"), http.StatusBadRequest)
		return
	}

	err = m.DB.DeleteLink(userID, linkID)
	if err != nil {
		utils.ErrorJSON(w, errors.New("failed to delete link"), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, "Link successfully deleted!")
}
