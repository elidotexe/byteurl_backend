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

func (m *postgresDBRepo) GetAllLinks(userID int) ([]models.Link, error) {
	var links []models.Link

	if err := m.DB.Where("user_id = ?", userID).Find(&links).Error; err != nil {
		return nil, err
	}

	return links, nil
}

func (m *postgresDBRepo) InsertLink(link *models.Link) (*models.Link, error) {
	var maxLinkID int

	if err := m.DB.Model(&models.Link{}).Where("user_id = ?", link.UserID).Select("COALESCE(MAX(id), 0)").Row().Scan(&maxLinkID); err != nil {
		return nil, err
	}

	link.ID = maxLinkID + 1

	if err := m.DB.Create(link).Error; err != nil {
		return nil, err
	}

	return link, nil
}

func (m *postgresDBRepo) GetLink(userID, linkID int) (*models.Link, error) {
	var link models.Link

	result := m.DB.Where("user_id = ? AND id = ?", userID, linkID).First(&link)
	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, errors.New("link not found")
	}

	return &link, nil
}

func (m *postgresDBRepo) UpdateLink(link *models.Link) (*models.Link, error) {
	result := m.DB.Model(&models.Link{}).Where("user_id = ? AND id = ?", link.UserID, link.ID).Updates(models.Link{Title: link.Title, OriginalURL: link.OriginalURL, UpdatedAt: link.UpdatedAt})
	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, errors.New("link not found")
	}

	return link, nil
}

func (m *postgresDBRepo) DeleteLink(userID int, linkID int) error {
	result := m.DB.Where("user_id = ? AND id = ?", userID, linkID).Delete(&models.Link{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("link not found")
	}

	return nil
}
