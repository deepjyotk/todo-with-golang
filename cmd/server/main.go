// cmd/server/main.go
package main

import (
	"log"

	"github.com/deepjyotk/todo-with-golang/configs"
	"github.com/deepjyotk/todo-with-golang/internal/handlers"
	"github.com/deepjyotk/todo-with-golang/internal/models"
	"github.com/deepjyotk/todo-with-golang/internal/repository/postgres"
	"github.com/deepjyotk/todo-with-golang/internal/routes"
	"github.com/deepjyotk/todo-with-golang/internal/services"
	"github.com/deepjyotk/todo-with-golang/internal/validators"
	"github.com/joho/godotenv"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/deepjyotk/todo-with-golang/docs" // Swagger docs import
)

func main() {
	// Load environment variables from .env file (optional).
	// Make sure your .env actually contains all DB_ / SERVER_ / JWT_ / S3_ vars.
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: no .env file found (continuing if Docker sets env).")
	}

	// Load configuration from environment (via Viper).
	cfg, err := configs.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Check that our DSN is non-empty
	log.Println("Database DSN:", cfg.Database.DSN)

	// Connect to Postgres using the DSN
	db, err := gorm.Open(pgdriver.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate the models
	if err := db.AutoMigrate(&models.User{}, &models.Todo{}, &models.Attachment{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	todoRepo := postgres.NewTodoRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, []byte(cfg.JWT.Secret))
	todoService := services.NewTodoService(todoRepo, cfg.S3)

	// Validator (optional, if you have S3 constraints or other needs)
	todoValidator := validators.NewTodoValidator(cfg)

	// Initialize handlers
	todoHandler := handlers.NewTodoHandler(todoService, todoValidator)
	authHandler := handlers.NewAuthHandler(authService)

	// Set up routes
	router := routes.SetupRouter(authHandler, todoHandler, cfg.JWT.Secret)

	// Start the server
	err = router.Run(":" + cfg.Server.Port)
	if err != nil {
		log.Fatal("Failed to run server:", err)
	}
}
