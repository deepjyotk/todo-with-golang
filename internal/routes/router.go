// internal/routes/router.go
package routes

import (
	_ "github.com/deepjyotk/todo-with-golang/docs" // Import the generated Swagger docs
	"github.com/deepjyotk/todo-with-golang/internal/handlers"
	"github.com/deepjyotk/todo-with-golang/internal/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter initializes the Gin engine, middleware, and routes.
// It now accepts jwtSecret as an argument for use in protected routes.
func SetupRouter(authHandler *handlers.AuthHandler, todoHandler *handlers.TodoHandler, jwtSecret string) *gin.Engine {
	router := gin.Default()

	// Swagger documentation route (publicly accessible)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API versioning group
	api := router.Group("/api/v1")
	{
		// Authentication routes (publicly accessible)
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			// Future endpoints:
			// authGroup.POST("/forgot-password", authHandler.ForgotPassword)
		}

		// Todo routes (protected by JWT authentication middleware)
		todoGroup := api.Group("/todos")
		// Apply the auth middleware only to TODO endpoints.
		todoGroup.Use(middleware.AuthMiddleware(jwtSecret))
		{
			// Create a new Todo item
			todoGroup.POST("", todoHandler.CreateTodo)
			// Retrieve a single Todo item by its ID
			todoGroup.GET("/:id", todoHandler.GetSpecificTodo)
			// Update an existing Todo item
			todoGroup.PUT("/:id", todoHandler.UpdateTodo)
			// Delete a Todo item
			todoGroup.DELETE("/:id", todoHandler.DeleteTodo)

			// Add an attachment to a Todo item
			// todoGroup.POST("/:id/attachments", todoHandler.AddAttachment) --> Future scope

			// Generate a presigned URL route for S3 operations.
			// @Router /api/v1/todos/presigned-url [get]
			todoGroup.GET("/presigned-url", todoHandler.GeneratePresignedS3UrlPutRequest)

			// api.GET("/users/:user_id/todos", middleware.AuthMiddleware(jwtSecret), todoHandler.GetTodos)
			todoGroup.GET("/get-all", todoHandler.GetAllTodosForUser)

		}

		// Get all Todos for a user by user ID.
		// This route is now authenticated by applying the middleware inline.

	}

	return router
}
