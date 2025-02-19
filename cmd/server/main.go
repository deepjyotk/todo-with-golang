package main

import (
	"log"

	"github.com/deepjyotk/todo-with-golang/configs"
	"github.com/deepjyotk/todo-with-golang/internal/handlers"
	"github.com/deepjyotk/todo-with-golang/internal/models"
	"github.com/deepjyotk/todo-with-golang/internal/repository/postgres"
	"github.com/deepjyotk/todo-with-golang/internal/routes"
	"github.com/deepjyotk/todo-with-golang/internal/services"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/deepjyotk/todo-with-golang/docs"
)

// Import docs for Swagger

// main is the entry point of the application. It performs the following steps:
// 1. Loads the application configuration from a YAML file.
// 2. Initializes a connection to the PostgreSQL database using the DSN from the configuration.
// 3. Automatically migrates the User model schema in the database.
// 4. Instantiates the user repository and authentication service with the JWT secret.
// 5. Sets up HTTP handlers for authentication routes.
// 6. Configures and starts the HTTP server using the specified port from the configuration.

func main() {
	// Load configuration (ensure configs/config.yaml exists and is correct)
	cfg, err := configs.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize database connection using DSN from config
	db, err := gorm.Open(pgdriver.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate the User model
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)

	// Initialize services with JWT secret from configuration
	authService := services.NewAuthService(userRepo, []byte(cfg.JWT.Secret))

	// Initialize HTTP handlers
	authHandler := handlers.NewAuthHandler(authService)

	// Setup routes
	router := routes.SetupRouter(authHandler)

	// Start the server (port from config)
	err = router.Run(":" + cfg.Server.Port)
	if err != nil {
		log.Fatal("Failed to run server:", err)
	}
}
