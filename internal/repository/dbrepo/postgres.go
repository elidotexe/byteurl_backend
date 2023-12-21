package dbrepo

import (
	"errors"

	"github.com/elidotexe/backend_byteurl/internal/models"
	"gorm.io/gorm"
)

func (m *postgresDBRepo) GetUserByEmail(email string) (*models.User, error) {
	var user models.User

	if err := m.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

func (m *postgresDBRepo) GetUserByID(userID int) (*models.User, error) {
	var user models.User

	if err := m.DB.Select("id, name, email").First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		return nil, err
	}

	return &user, nil
}

func (m *postgresDBRepo) UserExists(email string) (bool, error) {
	var existingUser models.User

	if err := m.DB.Where("email = ?", email).First(&existingUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (m *postgresDBRepo) CreateUser(user *models.User) error {
	if err := m.DB.Create(user).Error; err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) UpdateUserNameByID(userID int, user *models.User) error {
	if err := m.DB.Model(&models.User{}).Where("id = ?", userID).Update("name", user.Name).Error; err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) GetAllUserLinks(userID int) ([]models.Link, error) {
	var links []models.Link

	if err := m.DB.Where("user_id = ?", userID).Find(&links).Error; err != nil {
		return nil, err
	}

	return links, nil
}
