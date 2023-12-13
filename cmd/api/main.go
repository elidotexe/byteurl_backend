package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/elidotexe/backend_byteurl/internal/auth"
	"github.com/elidotexe/backend_byteurl/internal/config"
	"github.com/elidotexe/backend_byteurl/internal/driver"
	"github.com/elidotexe/backend_byteurl/internal/handlers"
	"github.com/elidotexe/backend_byteurl/internal/routes"
)

var app config.AppConfig
var authInstance *auth.Auth

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	app = *config
	app.DSN = fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		config.DB_HOST, config.DB_PORT, config.DB_NAME, config.DB_USER, config.DB_PASSWORD)

	fmt.Println(app.JWT_SECRET)

	authInstance = &auth.Auth{
		Secret:       config.JWT_SECRET,
		Issuer:       config.JWT_ISSUER,
		Audience:     config.JWT_AUDIENCE,
		CookieDomain: config.COOKIE_DOMAIN,
	}

	_, err = run()
	if err != nil {
		log.Fatal(err)
	}

	src := &http.Server{
		Addr:    ":" + app.PORT,
		Handler: routes.SetupRoutes(&app),
	}

	log.Println("Starting server on port", app.PORT)
	err = src.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func run() (*driver.DB, error) {
	log.Println("Connecting to database...")
	db, err := driver.ConnectGORM(app.DSN)
	if err != nil {
		log.Fatal("Cannot connect to database! Dying...", err)
		return nil, err
	}

	log.Println("Connected to database!")

	repo := handlers.NewRepo(&app, db, authInstance)
	handlers.NewHandlers(repo)

	return db, nil
}
