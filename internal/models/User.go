package models

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Link struct {
	ID          int       `json:"id"`
	UserID      int       `json:"userId" gorm:"index" validate:"required"`
	Title       string    `json:"title"`
	OriginalURL string    `json:"originalUrl" validate:"required,url"`
	ShortenURL  string    `json:"shortenUrl"`
	Clicks      int       `json:"clicks" sql:"default:0"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Links     []*Link   `json:"links" gorm:"foreignKey:UserID;references:ID"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// HashPassword takes a plain text password and returns a hashed password
func HashPassword(plainPassword string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

// PasswordMathes takes a plain text password and compares it to the hashed password
func (u *User) PasswordMathes(plainText string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plainText))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}
