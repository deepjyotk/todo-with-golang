package auth_handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deepjyotk/todo-with-golang/internal/models"
	"github.com/deepjyotk/todo-with-golang/internal/test/unit_tests/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRegister_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthService := new(mocks.MockAuthService)
	authHandler := NewAuthHandler(mockAuthService)
	router.POST("/api/v1/auth/register", authHandler.Register)

	// Prepare a valid registration request payload.
	reqBody := RegisterRequest{
		Username: "johndoe",
		Email:    "john@example.com",
		Password: "strongpassword",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Return a pointer to a models.User (not a value) for the successful registration.
	expectedUser := &models.User{
		ID:       1,
		Username: "johndoe",
		Email:    "john@example.com",
	}
	dummyToken := "dummy.jwt.token"

	mockAuthService.
		On("Register", reqBody.Username, reqBody.Email, reqBody.Password).
		Return(expectedUser, nil)
	mockAuthService.
		On("GenerateToken", expectedUser).
		Return(dummyToken, nil)

	respRecorder := httptest.NewRecorder()
	router.ServeHTTP(respRecorder, req)

	assert.Equal(t, http.StatusCreated, respRecorder.Code)
	var resp RegisterResponse
	err := json.Unmarshal(respRecorder.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser.ID, resp.ID)
	assert.Equal(t, expectedUser.Username, resp.Username)
	assert.Equal(t, expectedUser.Email, resp.Email)

	// Check that the JWT cookie was set correctly.
	cookieHeader := respRecorder.Header().Get("Set-Cookie")
	assert.Contains(t, cookieHeader, "jwt="+dummyToken)

	mockAuthService.AssertExpectations(t)
}

func TestRegister_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthService := new(mocks.MockAuthService)
	authHandler := NewAuthHandler(mockAuthService)
	router.POST("/api/v1/auth/register", authHandler.Register)

	// Send an invalid JSON payload.
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	respRecorder := httptest.NewRecorder()
	router.ServeHTTP(respRecorder, req)

	// Expect a 400 Bad Request response.
	assert.Equal(t, http.StatusBadRequest, respRecorder.Code)
	// The error message should indicate a JSON binding error.
	assert.Contains(t, respRecorder.Body.String(), "invalid character")
}

func TestRegister_RegisterFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthService := new(mocks.MockAuthService)
	authHandler := NewAuthHandler(mockAuthService)
	router.POST("/api/v1/auth/register", authHandler.Register)

	reqBody := RegisterRequest{
		Username: "johndoe",
		Email:    "john@example.com",
		Password: "strongpassword",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Simulate a failure during user registration.
	mockAuthService.
		On("Register", reqBody.Username, reqBody.Email, reqBody.Password).
		Return((*models.User)(nil), errors.New("registration error"))

	respRecorder := httptest.NewRecorder()
	router.ServeHTTP(respRecorder, req)

	// Expect a 500 Internal Server Error with the error message.
	assert.Equal(t, http.StatusInternalServerError, respRecorder.Code)
	assert.Contains(t, respRecorder.Body.String(), "registration error")

	mockAuthService.AssertExpectations(t)
}

func TestRegister_GenerateTokenFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthService := new(mocks.MockAuthService)
	authHandler := NewAuthHandler(mockAuthService)
	router.POST("/api/v1/auth/register", authHandler.Register)

	reqBody := RegisterRequest{
		Username: "johndoe",
		Email:    "john@example.com",
		Password: "strongpassword",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	expectedUser := &models.User{
		ID:       1,
		Username: "johndoe",
		Email:    "john@example.com",
	}
	// Simulate a successful registration.
	mockAuthService.
		On("Register", reqBody.Username, reqBody.Email, reqBody.Password).
		Return(expectedUser, nil)
	// Simulate failure during token generation.
	mockAuthService.
		On("GenerateToken", expectedUser).
		Return("", errors.New("token generation failed"))

	respRecorder := httptest.NewRecorder()
	router.ServeHTTP(respRecorder, req)

	// Expect a 500 Internal Server Error with the token error message.
	assert.Equal(t, http.StatusInternalServerError, respRecorder.Code)
	assert.Contains(t, respRecorder.Body.String(), "token generation failed")

	mockAuthService.AssertExpectations(t)
}

//! ******** Login Test ********

func TestLogin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a new mock auth service.
	mockAuthService := new(mocks.MockAuthService)
	// Initialize AuthHandler with the mock.
	authHandler := NewAuthHandler(mockAuthService)
	router.POST("/api/v1/auth/login", authHandler.Login)

	// Prepare a valid login request payload.
	reqBody := LoginRequest{
		Email:    "john@example.com",
		Password: "strongpassword",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Set expected behavior:
	// On successful authentication, the Login method returns a dummy token.
	dummyToken := "dummy.jwt.token"
	mockAuthService.
		On("Login", reqBody.Email, reqBody.Password).
		Return(dummyToken, nil)

	// Execute the request.
	respRecorder := httptest.NewRecorder()
	router.ServeHTTP(respRecorder, req)

	// Assert HTTP status.
	assert.Equal(t, http.StatusOK, respRecorder.Code)

	// Parse and assert the response body.
	var resp LoginResponse
	err := json.Unmarshal(respRecorder.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "login successful", resp.Message)

	// Check that the JWT cookie was set correctly.
	cookieHeader := respRecorder.Header().Get("Set-Cookie")
	assert.Contains(t, cookieHeader, "jwt="+dummyToken)

	mockAuthService.AssertExpectations(t)
}

func TestLogin_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthService := new(mocks.MockAuthService)
	authHandler := NewAuthHandler(mockAuthService)
	router.POST("/api/v1/auth/login", authHandler.Login)

	// Send an invalid JSON payload.
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	respRecorder := httptest.NewRecorder()
	router.ServeHTTP(respRecorder, req)

	// Expect a 400 Bad Request response.
	assert.Equal(t, http.StatusBadRequest, respRecorder.Code)
	// The error message should mention an issue with JSON parsing.
	assert.Contains(t, respRecorder.Body.String(), "invalid character")
}

func TestLogin_AuthenticationFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthService := new(mocks.MockAuthService)
	authHandler := NewAuthHandler(mockAuthService)
	router.POST("/api/v1/auth/login", authHandler.Login)

	// Prepare a login request payload with wrong credentials.
	reqBody := LoginRequest{
		Email:    "john@example.com",
		Password: "wrongpassword",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Simulate authentication failure.
	mockAuthService.
		On("Login", reqBody.Email, reqBody.Password).
		Return("", errors.New("invalid credentials"))

	respRecorder := httptest.NewRecorder()
	router.ServeHTTP(respRecorder, req)

	// Expect a 401 Unauthorized response with the error message.
	assert.Equal(t, http.StatusUnauthorized, respRecorder.Code)
	assert.Contains(t, respRecorder.Body.String(), "invalid credentials")

	mockAuthService.AssertExpectations(t)
}
