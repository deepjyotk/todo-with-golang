package auth_service

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/deepjyotk/todo-with-golang/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5" // Updated import to jwt/v5
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository implements the postgres.UserRepository interface.
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	user := args.Get(0)
	if user == nil {
		return nil, args.Error(1)
	}
	return user.(*models.User), args.Error(1)
}

func (m *MockUserRepository) CreateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func TestRegister_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtSecret := []byte("testsecret")
	authSvc := NewAuthService(mockRepo, jwtSecret)

	username, email, password := "testuser", "test@example.com", "password123"

	// Expect that no user exists for this email.
	mockRepo.On("GetUserByEmail", email).Return(nil, nil)
	// Expect that CreateUser is called with a user pointer.
	mockRepo.On("CreateUser", mock.MatchedBy(func(user *models.User) bool {
		// Check that the user fields are properly set.
		return user.Username == username &&
			user.Email == email &&
			// The password is stored hashed, so verify it matches.
			bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) == nil &&
			!user.CreatedAt.IsZero() &&
			!user.UpdatedAt.IsZero()
	})).Return(nil)

	user, err := authSvc.Register(username, email, password)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, email, user.Email)
	// Verify that the hashed password corresponds to the plain password.
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)))

	mockRepo.AssertExpectations(t)
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtSecret := []byte("testsecret")
	authSvc := NewAuthService(mockRepo, jwtSecret)

	username, email, password := "testuser", "test@example.com", "password123"
	existingUser := &models.User{
		Username:  "existinguser",
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Simulate that a user is already registered with the given email.
	mockRepo.On("GetUserByEmail", email).Return(existingUser, nil)

	user, err := authSvc.Register(username, email, password)
	assert.Nil(t, user)
	assert.EqualError(t, err, "user already exists")

	mockRepo.AssertExpectations(t)
}

func TestRegister_GetUserByEmailError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtSecret := []byte("testsecret")
	authSvc := NewAuthService(mockRepo, jwtSecret)

	username, email, password := "testuser", "test@example.com", "password123"

	// Simulate an error while retrieving the user.
	mockRepo.On("GetUserByEmail", email).Return(nil, errors.New("db error"))

	user, err := authSvc.Register(username, email, password)
	assert.Nil(t, user)
	assert.EqualError(t, err, "db error")

	mockRepo.AssertExpectations(t)
}

func TestRegister_CreateUserError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtSecret := []byte("testsecret")
	authSvc := NewAuthService(mockRepo, jwtSecret)

	username, email, password := "testuser", "test@example.com", "password123"

	// No user exists with the given email.
	mockRepo.On("GetUserByEmail", email).Return(nil, nil)
	// Simulate an error during user creation.
	mockRepo.On("CreateUser", mock.AnythingOfType("*models.User")).Return(errors.New("create error"))

	user, err := authSvc.Register(username, email, password)
	assert.Nil(t, user)
	assert.EqualError(t, err, "create error")

	mockRepo.AssertExpectations(t)
}

//! *********

func TestGenerateToken_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	jwtSecret := []byte("testsecret")
	authSvc := NewAuthService(mockRepo, jwtSecret)

	user := &models.User{
		ID:       123,
		Username: "testuser",
		Email:    "test@example.com",
	}

	// Act
	tokenString, err := authSvc.GenerateToken(user)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Parse token to verify its claims.
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	assert.NoError(t, err)
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, float64(user.ID), claims["user_id"])

	// Check expiration is roughly 24 hours from now.
	expUnix, ok := claims["exp"].(float64)
	assert.True(t, ok)
	expTime := time.Unix(int64(expUnix), 0)
	assert.WithinDuration(t, time.Now().Add(24*time.Hour), expTime, time.Minute)
}

func TestLogin_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	jwtSecret := []byte("testsecret")
	authSvc := NewAuthService(mockRepo, jwtSecret)
	email, password := "test@example.com", "password123"

	mockRepo.On("GetUserByEmail", email).Return(nil, errors.New("db error"))

	// Act
	token, err := authSvc.Login(email, password)

	// Assert
	assert.Empty(t, token)
	assert.EqualError(t, err, "db error")
	mockRepo.AssertExpectations(t)
}

func TestLogin_UserNotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	jwtSecret := []byte("testsecret")
	authSvc := NewAuthService(mockRepo, jwtSecret)
	email, password := "nonexistent@example.com", "password123"

	mockRepo.On("GetUserByEmail", email).Return(nil, nil)

	// Act
	token, err := authSvc.Login(email, password)

	// Assert
	assert.Empty(t, token)
	assert.EqualError(t, err, "invalid email or password")
	mockRepo.AssertExpectations(t)
}

func TestLogin_IncorrectPassword(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	jwtSecret := []byte("testsecret")
	authSvc := NewAuthService(mockRepo, jwtSecret)
	email, password := "test@example.com", "wrongpassword"

	// Create a user with a known password hash.
	correctPassword := "correctpassword"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)
	assert.NoError(t, err)
	user := &models.User{
		ID:       456,
		Username: "testuser",
		Email:    email,
		Password: string(hashedPassword),
	}

	mockRepo.On("GetUserByEmail", email).Return(user, nil)

	// Act
	token, err := authSvc.Login(email, password)

	// Assert
	assert.Empty(t, token)
	assert.EqualError(t, err, "invalid email or password")
	mockRepo.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	jwtSecret := []byte("testsecret")
	authSvc := NewAuthService(mockRepo, jwtSecret)
	email, password := "test@example.com", "password123"

	// Create a user with a password hash that corresponds to password123.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)
	user := &models.User{
		ID:       789,
		Username: "testuser",
		Email:    email,
		Password: string(hashedPassword),
	}

	mockRepo.On("GetUserByEmail", email).Return(user, nil)

	// Act
	token, err := authSvc.Login(email, password)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify token structure by checking the token string (e.g., contains two dots).
	assert.True(t, strings.Count(token, ".") == 2, "token should be a JWT with two dots")

	// Optionally, parse the token to ensure the correct claims.
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	assert.NoError(t, err)
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, float64(user.ID), claims["user_id"])

	mockRepo.AssertExpectations(t)
}

//! *************

func createTestContext() *gin.Context {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	return ctx
}

func TestGetUserIDFromContext_MissingJWT(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	jwtSecret := []byte("testsecret")
	authSvc := NewAuthService(mockRepo, jwtSecret)
	ctx := createTestContext()
	// Do not set the "jwt" key

	// Act
	userID, err := authSvc.GetUserIDFromContext(ctx)

	// Assert
	assert.Equal(t, uint(0), userID)
	assert.EqualError(t, err, "user not authenticated")
}

func TestGetUserIDFromContext_InvalidClaimsType(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	jwtSecret := []byte("testsecret")
	authSvc := NewAuthService(mockRepo, jwtSecret)
	ctx := createTestContext()
	// Set the "jwt" key with an invalid type.
	ctx.Set("jwt", "not a valid jwt.MapClaims")

	// Act
	userID, err := authSvc.GetUserIDFromContext(ctx)

	// Assert
	assert.Equal(t, uint(0), userID)
	assert.EqualError(t, err, "invalid JWT claims")
}

func TestGetUserIDFromContext_MissingUserID(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	jwtSecret := []byte("testsecret")
	authSvc := NewAuthService(mockRepo, jwtSecret)
	ctx := createTestContext()
	// Set the "jwt" key with jwt.MapClaims that lacks user_id.
	claims := jwt.MapClaims{
		"some_other_key": "value",
	}
	ctx.Set("jwt", claims)

	// Act
	userID, err := authSvc.GetUserIDFromContext(ctx)

	// Assert
	assert.Equal(t, uint(0), userID)
	// Note: The error message for missing user_id should be "user_id not found in token"
	assert.EqualError(t, err, "user_id not found in token")
}

func TestGetUserIDFromContext_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	jwtSecret := []byte("testsecret")
	authSvc := NewAuthService(mockRepo, jwtSecret)
	ctx := createTestContext()

	// Ensure that the key is exactly "jwt" (all lowercase) as used in your function.
	var expectedUserID uint = 42
	claims := jwt.MapClaims{
		"user_id": float64(expectedUserID), // jwt/v5 unmarshals numbers as float64.
	}
	ctx.Set("jwt", claims)

	// Optional debug: verify the context actually has the "jwt" key.
	value, exists := ctx.Get("jwt")
	assert.True(t, exists, "jwt key should exist in context")
	assert.Equal(t, claims, value, "jwt claims in context should match the expected claims")

	// Act
	userID, err := authSvc.GetUserIDFromContext(ctx)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedUserID, userID)
}
