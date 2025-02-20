// internal/handlers/todo_handler_test.go
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deepjyotk/todo-with-golang/internal/dto"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

// MockTodoService implements the TodoService interface for testing.
type MockTodoService struct {
	mock.Mock
}

func (m *MockTodoService) CreateTodo(todo *dto.CreateTodoRequest) error {
	args := m.Called(todo)
	return args.Error(0)
}

// (Define other methods as needed for your tests.)

// MockTodoValidator implements the TodoValidator interface for testing.
type MockTodoValidator struct {
	mock.Mock
}

func (m *MockTodoValidator) ValidateCreateTodoRequest(req *dto.CreateTodoRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

// --- Unit Test ---

func TestCreateTodoHandler_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create mocks
	mockService := new(MockTodoService)
	mockValidator := new(MockTodoValidator)

	// Prepare test data
	reqBody := dto.CreateTodoRequest{
		Title:       "Test Todo",
		Description: "Test Description",
	}
	// Assuming that in your real implementation, after mapping, a Todo object is created.
	// Set expectations on mocks.
	mockValidator.
		On("ValidateCreateTodoRequest", &reqBody).
		Return(nil)
	mockService.
		On("CreateTodo", mock.Anything).
		Return(nil)

	// Inject a dummy user ID into the Gin context using middleware or directly.
	router.Use(func(c *gin.Context) {
		// Simulate userID extraction from context.
		c.Set("userID", uint(1))
		c.Next()
	})

	// Initialize handler and route.
	todoHandler := NewTodoHandler(mockService, mockValidator)
	router.POST("/todos", todoHandler.CreateTodo)

	// Create the HTTP request.
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record the response.
	respRecorder := httptest.NewRecorder()
	router.ServeHTTP(respRecorder, req)

	// Assertions.
	assert.Equal(t, http.StatusCreated, respRecorder.Code)
	mockValidator.AssertExpectations(t)
	mockService.AssertExpectations(t)
}

func TestCreateTodoHandler_InvalidJSON(t *testing.T) {
	// Setup a minimal router with the handler.
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mocks for service and validator.
	mockService := new(MockTodoService)
	mockValidator := new(MockTodoValidator)
	todoHandler := NewTodoHandler(mockService, mockValidator)
	router.POST("/todos", todoHandler.CreateTodo)

	// Send an invalid JSON payload.
	req, _ := http.NewRequest(http.MethodPost, "/todos", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	respRecorder := httptest.NewRecorder()
	router.ServeHTTP(respRecorder, req)

	// Expect a 400 Bad Request.
	assert.Equal(t, http.StatusBadRequest, respRecorder.Code)
}
