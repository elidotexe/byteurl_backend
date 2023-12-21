package driver

import (
	"fmt"
	"time"

	"github.com/elidotexe/backend_byteurl/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB holds the database connection pool
type DB struct {
	Gorm *gorm.DB
}

var dbConn = &DB{}

const maxOpenDBConn = 10
const maxIdleDBConn = 5
const maxDBLifetime = 5 * time.Minute
const maxRetries = 3
const retryInterval = 5 * time.Second

// ConnectSQL creates a connection to the database
func ConnectGORM(dsn string) (*DB, error) {
	var db *gorm.DB
	var err error

	// Retry connection to the database if it fails for maxRetries times
	for retries := 0; retries < maxRetries; retries++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			fmt.Printf("Failed to connect to the database (attempt %d/%d): %v\n", retries+1, maxRetries, err)
			time.Sleep(retryInterval)
		} else {
			break
		}
	}

	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(maxOpenDBConn)
	sqlDB.SetMaxIdleConns(maxIdleDBConn)
	sqlDB.SetConnMaxLifetime(maxDBLifetime)

	dbConn.Gorm = db

	err = runMigrations(dbConn.Gorm)
	if err != nil {
		return nil, err
	}

	return dbConn, nil
}

// runMigrations runs the database migrations for the models
func runMigrations(db *gorm.DB) error {
	err := db.AutoMigrate(&models.User{}, &models.Link{})
	if err != nil {
		fmt.Printf("Cannot migrate user table: %v\n", err)
		return err
	}

	return nil
}
