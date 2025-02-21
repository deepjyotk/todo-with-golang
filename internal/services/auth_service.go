package services

import (
	"errors"
	"time"

	"github.com/deepjyotk/todo-with-golang/internal/models"
	"github.com/deepjyotk/todo-with-golang/internal/repository/postgres"
	"github.com/gin-gonic/gin"     // <-- new import for context
	"github.com/golang-jwt/jwt/v5" // <-- needed for jwt.MapClaims
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(username, email, password string) (*models.User, error)
	Login(email, password string) (string, error)
	GenerateToken(user *models.User) (string, error)

	// New method for extracting user ID from context:
	GetUserIDFromContext(c *gin.Context) (uint, error)
}

type authService struct {
	userRepo  postgres.UserRepository
	jwtSecret []byte
}

// NewAuthService creates a new instance of AuthService.
func NewAuthService(userRepo postgres.UserRepository, jwtSecret []byte) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// Register creates a new user after hashing the password.
func (s *authService) Register(username, email, password string) (*models.User, error) {
	existingUser, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:  username,
		Email:     email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.userRepo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GenerateToken creates a JWT token for the given user.
func (s *authService) GenerateToken(user *models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // Token valid for 24 hours.
	})
	return token.SignedString(s.jwtSecret)
}

// Login verifies user credentials and returns a JWT token string.
func (s *authService) Login(email, password string) (string, error) {
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	return s.GenerateToken(user)
}

// GetUserIDFromContext extracts the user ID from the JWT claims stored in the Gin context.
// It returns the user ID as a uint or an error if it cannot be extracted.
func (s *authService) GetUserIDFromContext(c *gin.Context) (uint, error) {
	// Retrieve the JWT claims from the context.
	jwtClaimsVal, exists := c.Get("jwt")
	if !exists {
		return 0, errors.New("user not authenticated")
	}

	// Assert that the claims are of type jwt.MapClaims.
	claims, ok := jwtClaimsVal.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid JWT claims")
	}

	// Extract the user_id from the claims.
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("user_id not found in token")
	}

	return uint(userIDFloat), nil
}
