package repository

import "github.com/elidotexe/backend_byteurl/internal/models"

type DatabaseRepo interface {
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id int) (*models.User, error)
	UserExists(email string) (bool, error)
	CreateUser(user *models.User) error
	UpdateUserNameByID(userID int, user *models.User) error

	GetAllLinks(userID int) ([]models.Link, error)
	InsertLink(link *models.Link) (*models.Link, error)
	GetLink(userID, linkID int) (*models.Link, error)
	GetLinkByShortenURL(shortenURL string) (*models.Link, error)
	UpdateLink(link *models.Link) (*models.Link, error)
	DeleteLink(userID int, linkID int) error
}
