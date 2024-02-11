package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
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

	response := map[string]string{"message": "success"}

	_ = utils.WriteJSON(w, http.StatusOK, response)
}

func (m *Repository) GetUserName(w http.ResponseWriter, r *http.Request) {
	id, _ := utils.GetIDFromURL(r.URL.Path)

	userID, err := strconv.Atoi(id)
	if err != nil {
		utils.ErrorJSON(w, errors.New("invalid user id"), http.StatusBadRequest)
		return
	}

	user, err := m.DB.GetUserByID(userID)
	if err != nil {
		utils.ErrorJSON(w, errors.New("invalid user"), http.StatusBadRequest)
		return
	}

	response := map[string]string{"name": user.Name}

	utils.WriteJSON(w, http.StatusOK, response)
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

	var user *models.User
	user, err = m.DB.GetUserByID(userID)
	if err != nil {
		utils.ErrorJSON(w, errors.New("failed to retrieve updated user"), http.StatusInternalServerError)
		return
	}

	if user.Name == payload.Name {
		utils.ErrorJSON(w, errors.New("name is the same"), http.StatusBadRequest)
		return
	}

	updatedUser := &models.User{
		Name:      payload.Name,
		UpdatedAt: time.Now(),
	}

	err = m.DB.UpdateUserNameByID(userID, updatedUser)
	if err != nil {
		utils.ErrorJSON(w, errors.New("failed to update user"), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"name": updatedUser.Name}

	utils.WriteJSON(w, http.StatusOK, response)
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

	randHash, err := utils.GenerateRandomHash(10)
	if err != nil {
		utils.ErrorJSON(w, errors.New("failed to generate random hash"), http.StatusInternalServerError)
		return
	}

	newLink := models.Link{
		UserID:      userID,
		Title:       payload.Title,
		OriginalURL: payload.OriginalURL,
		ShortenURL:  randHash,
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

func (m *Repository) RedirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
	hashURLPattern := regexp.MustCompile(`/([a-zA-Z0-9-]+)$`)
	matches := hashURLPattern.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		utils.ErrorJSON(w, errors.New("invalid short url"), http.StatusBadRequest)
		return
	}

	hash := matches[1]

	link, err := m.DB.GetLinkByShortenURL(hash)
	if err != nil {
		utils.ErrorJSON(w, errors.New("failed to retrieve link"), http.StatusInternalServerError)
		return
	}

	link.Clicks++

	_, err = m.DB.UpdateRedirectDetails(link)
	if err != nil {
		utils.ErrorJSON(w, errors.New("failed to update link"), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"originalUrl": link.OriginalURL}

	utils.WriteJSON(w, http.StatusOK, response)
}

func (m *Repository) CreateRedirectHistory(w http.ResponseWriter, r *http.Request) {
	hashURLPattern := regexp.MustCompile(`/([a-zA-Z0-9-]+)$`)
	matches := hashURLPattern.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		utils.ErrorJSON(w, errors.New("invalid short url"), http.StatusBadRequest)
		return
	}

	hash := matches[1]

	link, err := m.DB.GetLinkByShortenURL(hash)
	if err != nil {
		utils.ErrorJSON(w, errors.New("failed to retrieve link"), http.StatusInternalServerError)
		return
	}

	var payload struct {
		Device    string `json:"device"`
		Browser   string `json:"browser"`
		IPAddress string `json:"ipAddress"`
		Location  string `json:"location"`
	}

	err = utils.ReadJSON(w, r, &payload)
	if err != nil {
		utils.ErrorJSON(w, errors.New("invalid request payload"), http.StatusBadRequest)
		return
	}

	fmt.Println(payload.Location)
	fmt.Println(link.ID)

	redirectHistory := models.RedirectHistory{
		LinkID:    link.ID,
		Device:    payload.Device,
		Browser:   payload.Browser,
		IPAddress: payload.IPAddress,
		Location:  payload.Location,
		CreatedAt: time.Now(),
	}

	fmt.Println(redirectHistory)

	_, err = m.DB.InsertRedirectHistory(&redirectHistory)
	if err != nil {
		utils.ErrorJSON(w, errors.New("failed to insert redirect history"), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"message": "success"}

	utils.WriteJSON(w, http.StatusOK, response)
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

func (m *Repository) LinksWithRedirectHistory(w http.ResponseWriter, r *http.Request) {
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

	link, err := m.DB.GetLinksWithRedirectHistory(userID)
	if err != nil {
		utils.ErrorJSON(w, errors.New("failed to retrieve link"), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, link)
}
