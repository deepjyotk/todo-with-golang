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
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/deepjyotk/todo-with-golang/docs" // Swagger docs import
)

func main() {
	// Load configuration
	cfg, err := configs.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize database connection
	db, err := gorm.Open(pgdriver.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate the models: User, Todo, and Attachment
	err = db.AutoMigrate(&models.User{}, &models.Todo{}, &models.Attachment{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	todoRepo := postgres.NewTodoRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, []byte(cfg.JWT.Secret))
	todoService := services.NewTodoService(todoRepo, cfg.S3)

	// Create validator with injected configuration
	todoValidator := validators.NewTodoValidator(cfg)

	// Now, when initializing your handlers, pass the validator if needed.
	// For example, if the TodoHandler requires validation:
	todoHandler := handlers.NewTodoHandler(todoService, todoValidator)
	authHandler := handlers.NewAuthHandler(authService)

	// Setup routes with dependency injection of both handlers
	router := routes.SetupRouter(authHandler, todoHandler, cfg.JWT.Secret)

	// Start the server on the specified port
	err = router.Run(":" + cfg.Server.Port)
	if err != nil {
		log.Fatal("Failed to run server:", err)
	}
}
