package repository

import "github.com/elidotexe/backend_byteurl/internal/models"

type DatabaseRepo interface {
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id int) (*models.User, error)
	UserExists(email string) (bool, error)
	CreateUser(user *models.User) error
}
