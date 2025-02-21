package auth_handler

import (
	"net/http"

	"github.com/deepjyotk/todo-with-golang/internal/services/auth_service"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles user authentication endpoints.
type AuthHandler struct {
	authService auth_service.AuthService
}

// NewAuthHandler returns a new instance of AuthHandler.
func NewAuthHandler(authService auth_service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RegisterRequest defines the payload for user registration.
type RegisterRequest struct {
	Username string `json:"username" binding:"required" example:"johndoe"`
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"strongpassword"`
}

// RegisterResponse defines the response payload after registration.
type RegisterResponse struct {
	ID       uint   `json:"id" example:"1"`
	Username string `json:"username" example:"johndoe"`
	Email    string `json:"email" example:"john@example.com"`
}

// LoginRequest defines the payload for user login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required" example:"strongpassword"`
}

// LoginResponse defines the response payload after a successful login.
type LoginResponse struct {
	Message string `json:"message" example:"login successful"`
}

// ErrorResponse defines a standard error response.
type ErrorResponse struct {
	Message string `json:"message" example:"error message"`
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with username, email, and password.
// @Tags auth
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "User registration data"
// @Success 201 {object} RegisterResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
		return
	}

	// Create the user.
	user, err := h.authService.Register(req.Username, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		return
	}

	// Generate a JWT token for the new user.
	token, err := h.authService.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		return
	}

	// Store the token in an HTTP-only cookie (valid for 24 hours).
	c.SetCookie("jwt", token, 3600*24, "/", "", false, true)

	// Respond with the created user details.
	c.JSON(http.StatusCreated, RegisterResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	})
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return JWT token in an HTTP-only cookie.
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "User credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
		return
	}

	// Authenticate the user and generate a JWT token.
	token, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: err.Error()})
		return
	}

	// Set the JWT token in an HTTP-only cookie (valid for 24 hours).
	c.SetCookie("jwt", token, 3600*24, "/", "", false, true)
	c.JSON(http.StatusOK, LoginResponse{Message: "login successful"})
}
