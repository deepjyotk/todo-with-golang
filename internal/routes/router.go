package routes

import (
	_ "github.com/deepjyotk/todo-with-golang/docs" // Import the generated Swagger docs
	_ "github.com/deepjyotk/todo-with-golang/docs" // Import the generated Swagger docs
	"github.com/deepjyotk/todo-with-golang/internal/handlers"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter initializes the Gin engine, middleware, and routes.
func SetupRouter(authHandler *handlers.AuthHandler) *gin.Engine {
	router := gin.Default()

	// Global middleware (e.g., error handler) can be attached here.
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API versioning
	api := router.Group("/api/v1")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			// Future: authGroup.POST("/forgot-password", authHandler.ForgotPassword)
		}
	}

	return router
}
