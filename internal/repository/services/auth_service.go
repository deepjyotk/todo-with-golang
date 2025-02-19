package services

import (
	"errors"
	"time"

	"github.com/deepjyotk/todo-with-golang/internal/models"
	"github.com/deepjyotk/todo-with-golang/internal/repository/postgres"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService defines the authentication operations.
type AuthService interface {
	Register(username, email, password string) (*models.User, error)
	Login(email, password string) (string, error)
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
	// Hash the password with bcrypt
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

// Login verifies user credentials and returns a JWT token string.
func (s *authService) Login(email, password string) (string, error) {
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("invalid email or password")
	}
	// Compare the hashed password with the provided one
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	// Create JWT token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // Token valid for 24 hours
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
