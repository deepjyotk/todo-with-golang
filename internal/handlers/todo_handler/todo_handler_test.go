package todo_handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deepjyotk/todo-with-golang/internal/dto"

	"github.com/deepjyotk/todo-with-golang/internal/models"
	"github.com/deepjyotk/todo-with-golang/internal/test/unit_tests/mocks"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateTodo_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	//! Arrange
	mockService := new(mocks.MockTodoService)
	mockValidator := new(mocks.MockTodoValidator)
	mockAuthService := new(mocks.MockAuthService)

	// Prepare test data
	reqBody := dto.CreateTodoRequest{
		Title:       "Test Todo",
		Description: "Test Description",
	}

	// Set expectations on mocks
	mockValidator.
		On("ValidateCreateTodoRequest", &reqBody).
		Return(nil)
	mockService.
		On("CreateTodo", mock.Anything).
		Return(nil)

	mockAuthService.
		On("GetUserIDFromContext", mock.AnythingOfType("*gin.Context")).
		Return(uint(1), nil)

	// Inject a valid JWT claim into the Gin context
	router.Use(func(c *gin.Context) {
		claims := jwt.MapClaims{"user_id": float64(1)}
		c.Set("jwt", claims)
		c.Next()
	})

	//! Act
	todoHandler := NewTodoHandler(mockService, mockValidator, mockAuthService)
	router.POST("/todos", todoHandler.CreateTodo)

	// Create the HTTP request
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record the response
	respRecorder := httptest.NewRecorder()
	router.ServeHTTP(respRecorder, req)

	//! Assert
	assert.Equal(t, http.StatusCreated, respRecorder.Code)
	mockValidator.AssertExpectations(t)
	mockService.AssertExpectations(t)
}

func TestCreateTodoHandler_InvalidJSON(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create mocks
	mockService := new(mocks.MockTodoService)
	mockValidator := new(mocks.MockTodoValidator)
	mockAuthService := new(mocks.MockAuthService)
	todoHandler := NewTodoHandler(mockService, mockValidator, mockAuthService)
	router.POST("/todos", todoHandler.CreateTodo)

	// Send an invalid JSON payload
	req, _ := http.NewRequest(http.MethodPost, "/todos", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	respRecorder := httptest.NewRecorder()
	router.ServeHTTP(respRecorder, req)

	// Expect a 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, respRecorder.Code)
}

func TestCreateTodo_UnauthorizedUser_ReturnsHttpUnauthorized(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create mocks
	mockService := new(mocks.MockTodoService)
	mockValidator := new(mocks.MockTodoValidator)
	mockAuthService := new(mocks.MockAuthService)

	// Set up mock expectation for an **invalid token scenario**
	mockAuthService.
		On("GetUserIDFromContext", mock.AnythingOfType("*gin.Context")).
		Return(uint(0), errors.New("invalid token")) // Simulating token failure

	// Initialize handler and route
	todoHandler := NewTodoHandler(mockService, mockValidator, mockAuthService)
	router.POST("/todos", todoHandler.CreateTodo)

	// Create a valid request body
	reqBody := dto.CreateTodoRequest{
		Title:       "Test Todo",
		Description: "Test Description",
	}
	body, _ := json.Marshal(reqBody)

	// Create HTTP request
	req, _ := http.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record the response
	respRecorder := httptest.NewRecorder()
	router.ServeHTTP(respRecorder, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, respRecorder.Code)     // Expect 401 Unauthorized
	assert.Contains(t, respRecorder.Body.String(), "invalid token") // Error message should match

	mockAuthService.AssertExpectations(t) // Verify that the mock was called
}

func TestCreateTodo_InvalidRequest_ReturnsHttpBadRequest(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	//! Arrange
	mockService := new(mocks.MockTodoService)
	mockValidator := new(mocks.MockTodoValidator)
	mockAuthService := new(mocks.MockAuthService)

	// Valid JSON, but an invalid request for validation
	reqBody := dto.CreateTodoRequest{
		Title:       "Lorem ipsum dolor sit  sapien tempor at. Cras vitae hendrerit magna. Maecenas orci eros, porta nec finibus quis, iaculis sed justo. Vestibulum auctor rutrum erat id venenatis. In eu volutpat justo. Etiam facilisis tellus est, sit amet euismod tellus rutrum posuere. Ut nec ", // Invalid: Title is empty (assuming it's required)
		Description: "Test Description",
	}

	// Ensure the request is parsed correctly but fails validation
	mockAuthService.
		On("GetUserIDFromContext", mock.AnythingOfType("*gin.Context")).
		Return(uint(1), nil)

	mockValidator.
		On("ValidateCreateTodoRequest", mock.Anything).
		Return(errors.New("Too long")) // Simulating validation failure

		// Inject a valid JWT claim into the Gin context
	router.Use(func(c *gin.Context) {
		claims := jwt.MapClaims{"user_id": float64(1)}
		c.Set("jwt", claims)
		c.Next()
	})

	// Initialize handler and route
	//! Act
	todoHandler := NewTodoHandler(mockService, mockValidator, mockAuthService)
	router.POST("/todos", todoHandler.CreateTodo)

	// Create the HTTP request (valid JSON format but logically incorrect data)
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record the response
	respRecorder := httptest.NewRecorder()
	router.ServeHTTP(respRecorder, req)

	//! Assert
	assert.Equal(t, http.StatusBadRequest, respRecorder.Code)  // Expect 400 Bad Request
	assert.Contains(t, respRecorder.Body.String(), "Too long") // Error message should match

	mockValidator.AssertExpectations(t)   // Ensure validation was called
	mockAuthService.AssertExpectations(t) // Ensure auth service was called
}

//! ***************** eneratePresignedS3UrlPutRequest *****************

// TestGeneratePresignedS3UrlPutRequest_Success verifies that a valid request returns a presigned URL.
func TestGeneratePresignedS3UrlPutRequest_Success(t *testing.T) {
	// Setup Gin in test mode and create a new router.
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create mocks.
	mockService := new(mocks.MockTodoService)
	mockValidator := new(mocks.MockTodoValidator)
	mockAuthService := new(mocks.MockAuthService)

	// Prepare test data.
	fileName := "testfile.txt"
	presignedURL := "https://s3.amazonaws.com/bucket/testfile.txt?signature=xyz"
	userID := uint(1) // Correct type: uint

	// Set expectations on the mocks.
	mockAuthService.
		On("GetUserIDFromContext", mock.AnythingOfType("*gin.Context")).
		Return(userID, nil)

	mockValidator.
		On("ValidFileName", fileName).
		Return(nil)

	mockService.
		On("GeneratePresignedS3Url", userID, fileName, "PUT").
		Return(presignedURL, nil)

	// Optionally inject a valid JWT claim into the context.
	router.Use(func(c *gin.Context) {
		c.Set("jwt", map[string]interface{}{"user_id": float64(userID)})
		c.Next()
	})

	// Initialize the handler and register the route.
	todoHandler := NewTodoHandler(mockService, mockValidator, mockAuthService)
	router.GET("/presigned-url", todoHandler.GeneratePresignedS3UrlPutRequest)

	// Create an HTTP GET request with the filename query parameter.
	req, _ := http.NewRequest(http.MethodGet, "/presigned-url?filename="+fileName, nil)
	respRecorder := httptest.NewRecorder()

	// Execute the request.
	router.ServeHTTP(respRecorder, req)

	// Assert that the status code is 200 and the response contains the presigned URL.
	assert.Equal(t, http.StatusOK, respRecorder.Code)
	expectedBody := `{"url":"` + presignedURL + `"}`
	assert.JSONEq(t, expectedBody, respRecorder.Body.String())

	// Verify that all expectations were met.
	mockAuthService.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
	mockService.AssertExpectations(t)
}

// TestGeneratePresignedS3UrlPutRequest_MissingFilename ensures that requests missing the 'filename' query parameter are rejected.
func TestGeneratePresignedS3UrlPutRequest_MissingFilename(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := new(mocks.MockTodoService)
	mockValidator := new(mocks.MockTodoValidator)
	mockAuthService := new(mocks.MockAuthService)

	userID := uint(1)
	// Set the auth service to return a valid user ID.
	mockAuthService.
		On("GetUserIDFromContext", mock.AnythingOfType("*gin.Context")).
		Return(userID, nil)

	todoHandler := NewTodoHandler(mockService, mockValidator, mockAuthService)
	router.GET("/presigned-url", todoHandler.GeneratePresignedS3UrlPutRequest)

	// Create a request without the filename parameter.
	req, _ := http.NewRequest(http.MethodGet, "/presigned-url", nil)
	respRecorder := httptest.NewRecorder()

	router.ServeHTTP(respRecorder, req)

	// Expect a 400 Bad Request with an appropriate error message.
	assert.Equal(t, http.StatusBadRequest, respRecorder.Code)
	assert.Contains(t, respRecorder.Body.String(), "filename is required")

	mockAuthService.AssertExpectations(t)
}

// TestGeneratePresignedS3UrlPutRequest_InvalidFilename tests that an invalid filename returns a 400 Bad Request.
func TestGeneratePresignedS3UrlPutRequest_InvalidFilename(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := new(mocks.MockTodoService)
	mockValidator := new(mocks.MockTodoValidator)
	mockAuthService := new(mocks.MockAuthService)

	userID := uint(1)
	fileName := "invalid|file.txt" // Example of an invalid filename.
	validationErr := errors.New("invalid filename format")

	mockAuthService.
		On("GetUserIDFromContext", mock.AnythingOfType("*gin.Context")).
		Return(userID, nil)

	mockValidator.
		On("ValidFileName", fileName).
		Return(validationErr)

	todoHandler := NewTodoHandler(mockService, mockValidator, mockAuthService)
	router.GET("/presigned-url", todoHandler.GeneratePresignedS3UrlPutRequest)

	req, _ := http.NewRequest(http.MethodGet, "/presigned-url?filename="+fileName, nil)
	respRecorder := httptest.NewRecorder()

	router.ServeHTTP(respRecorder, req)

	// Expect a 400 Bad Request with the validation error message.
	assert.Equal(t, http.StatusBadRequest, respRecorder.Code)
	assert.Contains(t, respRecorder.Body.String(), validationErr.Error())

	mockAuthService.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
}

// TestGeneratePresignedS3UrlPutRequest_Unauthorized verifies that when the user is not authenticated, a 401 Unauthorized is returned.
func TestGeneratePresignedS3UrlPutRequest_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := new(mocks.MockTodoService)
	mockValidator := new(mocks.MockTodoValidator)
	mockAuthService := new(mocks.MockAuthService)

	authErr := errors.New("invalid token")
	mockAuthService.
		On("GetUserIDFromContext", mock.AnythingOfType("*gin.Context")).
		Return(0, authErr)

	todoHandler := NewTodoHandler(mockService, mockValidator, mockAuthService)
	router.GET("/presigned-url", todoHandler.GeneratePresignedS3UrlPutRequest)

	req, _ := http.NewRequest(http.MethodGet, "/presigned-url?filename=testfile.txt", nil)
	respRecorder := httptest.NewRecorder()

	router.ServeHTTP(respRecorder, req)

	// Expect a 401 Unauthorized and a corresponding error message.
	assert.Equal(t, http.StatusUnauthorized, respRecorder.Code)
	assert.Contains(t, respRecorder.Body.String(), "unauthorized")

	mockAuthService.AssertExpectations(t)
}

// TestGeneratePresignedS3UrlPutRequest_ServiceError ensures that if the URL generation service fails, a 500 Internal Server Error is returned.
func TestGeneratePresignedS3UrlPutRequest_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := new(mocks.MockTodoService)
	mockValidator := new(mocks.MockTodoValidator)
	mockAuthService := new(mocks.MockAuthService)

	userID := uint(1)
	fileName := "testfile.txt"
	serviceErr := errors.New("s3 error")

	mockAuthService.
		On("GetUserIDFromContext", mock.AnythingOfType("*gin.Context")).
		Return(userID, nil)

	mockValidator.
		On("ValidFileName", fileName).
		Return(nil)

	mockService.
		On("GeneratePresignedS3Url", userID, fileName, "PUT").
		Return("", serviceErr)

	todoHandler := NewTodoHandler(mockService, mockValidator, mockAuthService)
	router.GET("/presigned-url", todoHandler.GeneratePresignedS3UrlPutRequest)

	req, _ := http.NewRequest(http.MethodGet, "/presigned-url?filename="+fileName, nil)
	respRecorder := httptest.NewRecorder()

	router.ServeHTTP(respRecorder, req)

	// Expect a 500 Internal Server Error with an appropriate error message.
	assert.Equal(t, http.StatusInternalServerError, respRecorder.Code)
	assert.Contains(t, respRecorder.Body.String(), "failed to generate presigned URL")

	mockAuthService.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
	mockService.AssertExpectations(t)
}

// ! ********************* GetSpecificTodo Tests *********************
func TestGetSpecificTodo_Success(t *testing.T) {
	// Set Gin to test mode and initialize router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create mocks for the service and auth dependencies.
	mockTodoService := new(mocks.MockTodoService)
	mockAuthService := new(mocks.MockAuthService)

	// Prepare a sample todo that will be returned by the service.
	testTodo := models.Todo{
		ID:          uint(1),
		Title:       "Test Todo",
		Description: "Test Description",
		UserID:      uint(1),
	}

	// Set expectations on the auth mock: return a valid user ID.
	mockAuthService.
		On("GetUserIDFromContext", mock.AnythingOfType("*gin.Context")).
		Return(uint(1), nil)

	// Set up the todo service mock using loose matchers.
	mockTodoService.
		On("GetTodoByIDAndUserID", mock.Anything, mock.Anything).
		Return(&testTodo, nil)

	// Initialize the handler with the mocks.
	todoHandler := NewTodoHandler(mockTodoService, nil, mockAuthService)
	// Register the GET route.
	router.GET("/todos/:id", todoHandler.GetSpecificTodo)

	// Create a GET request for todo id "1"
	req, _ := http.NewRequest(http.MethodGet, "/todos/1", nil)
	respRecorder := httptest.NewRecorder()

	// Execute the request.
	router.ServeHTTP(respRecorder, req)

	// Assert that the status is 200 OK.
	assert.Equal(t, http.StatusOK, respRecorder.Code)

	// Unmarshal the response and compare the returned todo.
	var returnedTodo models.Todo
	err := json.Unmarshal(respRecorder.Body.Bytes(), &returnedTodo)
	assert.NoError(t, err)
	assert.Equal(t, testTodo, returnedTodo)

	// Explicitly assert that GetTodoByIDAndUserID was called with the expected arguments.
	mockTodoService.AssertCalled(t, "GetTodoByIDAndUserID", uint(1), uint(1))
	mockAuthService.AssertExpectations(t)
}

func TestGetSpecificTodo_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// We still need to create the mocks, even though they won't be invoked.
	mockTodoService := new(mocks.MockTodoService)
	mockAuthService := new(mocks.MockAuthService)

	todoHandler := NewTodoHandler(mockTodoService, nil, mockAuthService)
	router.GET("/todos/:id", todoHandler.GetSpecificTodo)

	// Use an invalid (non-numeric) id value.
	req, _ := http.NewRequest(http.MethodGet, "/todos/abc", nil)
	respRecorder := httptest.NewRecorder()

	router.ServeHTTP(respRecorder, req)

	// Expect a 400 Bad Request with an "invalid todo id" error message.
	assert.Equal(t, http.StatusBadRequest, respRecorder.Code)
	assert.Contains(t, respRecorder.Body.String(), "invalid todo id")
}

func TestGetSpecificTodo_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockTodoService := new(mocks.MockTodoService)
	mockAuthService := new(mocks.MockAuthService)

	// Simulate an error when retrieving the user ID.
	mockAuthService.
		On("GetUserIDFromContext", mock.AnythingOfType("*gin.Context")).
		Return(uint(0), errors.New("unauthorized"))

	todoHandler := NewTodoHandler(mockTodoService, nil, mockAuthService)
	router.GET("/todos/:id", todoHandler.GetSpecificTodo)

	req, _ := http.NewRequest(http.MethodGet, "/todos/1", nil)
	respRecorder := httptest.NewRecorder()

	router.ServeHTTP(respRecorder, req)

	// Expect a 401 Unauthorized with an "unauthorized" error message.
	assert.Equal(t, http.StatusUnauthorized, respRecorder.Code)
	assert.Contains(t, respRecorder.Body.String(), "unauthorized")

	mockAuthService.AssertExpectations(t)
}

func TestGetSpecificTodo_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockTodoService := new(mocks.MockTodoService)
	mockAuthService := new(mocks.MockAuthService)

	// Set the auth expectation to return a valid user.
	mockAuthService.
		On("GetUserIDFromContext", mock.AnythingOfType("*gin.Context")).
		Return(uint(1), nil)
	// Simulate the case where no todo is found.
	mockTodoService.
		On("GetTodoByIDAndUserID", uint(1), uint(1)).
		Return(models.Todo{}, errors.New("not found"))

	todoHandler := NewTodoHandler(mockTodoService, nil, mockAuthService)
	router.GET("/todos/:id", todoHandler.GetSpecificTodo)

	req, _ := http.NewRequest(http.MethodGet, "/todos/1", nil)
	respRecorder := httptest.NewRecorder()

	router.ServeHTTP(respRecorder, req)

	// Expect a 404 Not Found with a "todo not found" message.
	assert.Equal(t, http.StatusNotFound, respRecorder.Code)
	assert.Contains(t, respRecorder.Body.String(), "todo not found")

	mockAuthService.AssertExpectations(t)
	mockTodoService.AssertExpectations(t)
}

//! ***************** UpdateTodo *****************

func TestUpdateTodo_Success(t *testing.T) {
	// Setup Gin in test mode.
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create mocks.
	mockService := new(mocks.MockTodoService)
	mockValidator := new(mocks.MockTodoValidator)
	mockAuthService := new(mocks.MockAuthService)

	// Prepare test request.
	reqBody := dto.UpdateTodoRequest{
		ID:          1,
		Title:       "Updated Title",
		Description: "Updated Description",
	}

	// Set expectations.
	mockAuthService.
		On("GetUserIDFromContext", mock.AnythingOfType("*gin.Context")).
		Return(uint(1), nil)

	mockValidator.
		On("ValidateUpdateTodoRequest", &reqBody).
		Return(nil)

	// Simulate existing todo retrieval.
	mockService.
		On("GetTodoByIDAndUserID", uint(reqBody.ID), uint(1)).
		Return(&models.Todo{ID: 1, Title: "Old Title", Description: "Old Desc", UserID: 1}, nil)

	// Expect the update call to succeed.
	mockService.
		On("UpdateTodo", mock.AnythingOfType("*models.Todo")).
		Return(nil)

	// Initialize the handler.
	todoHandler := NewTodoHandler(mockService, mockValidator, mockAuthService)
	// Using PUT method on a sample URL; the actual path value is not used by the handler.
	router.PUT("/todos/1", func(c *gin.Context) {
		// Inject a valid JWT claim into the context.
		claims := jwt.MapClaims{"user_id": float64(1)}
		c.Set("jwt", claims)
		c.Next()
	}, todoHandler.UpdateTodo)

	// Create the HTTP request.
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, "/todos/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record the response.
	respRecorder := httptest.NewRecorder()
	router.ServeHTTP(respRecorder, req)

	// Assert.
	assert.Equal(t, http.StatusOK, respRecorder.Code)
	mockAuthService.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
	mockService.AssertExpectations(t)
}

func TestUpdateTodo_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create mocks.
	mockService := new(mocks.MockTodoService)
	mockValidator := new(mocks.MockTodoValidator)
	mockAuthService := new(mocks.MockAuthService)

	todoHandler := NewTodoHandler(mockService, mockValidator, mockAuthService)
	router.PUT("/todos/1", todoHandler.UpdateTodo)

	// Send an invalid JSON payload.
	req, _ := http.NewRequest(http.MethodPut, "/todos/1", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	respRecorder := httptest.NewRecorder()
	router.ServeHTTP(respRecorder, req)

	// Expect 400 Bad Request.
	assert.Equal(t, http.StatusBadRequest, respRecorder.Code)
}

func TestUpdateTodo_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create mocks.
	mockService := new(mocks.MockTodoService)
	mockValidator := new(mocks.MockTodoValidator)
	mockAuthService := new(mocks.MockAuthService)

	// Simulate an unauthorized scenario.
	mockAuthService.
		On("GetUserIDFromContext", mock.AnythingOfType("*gin.Context")).
		Return(uint(0), errors.New("invalid token"))

	todoHandler := NewTodoHandler(mockService, mockValidator, mockAuthService)
	router.PUT("/todos/1", todoHandler.UpdateTodo)

	// Create a valid JSON payload.
	reqBody := dto.UpdateTodoRequest{
		ID:          1,
		Title:       "Updated Title",
		Description: "Updated Description",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, "/todos/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	respRecorder := httptest.NewRecorder()
	router.ServeHTTP(respRecorder, req)

	// Expect 401 Unauthorized.
	assert.Equal(t, http.StatusUnauthorized, respRecorder.Code)
	assert.Contains(t, respRecorder.Body.String(), "invalid token")
	mockAuthService.AssertExpectations(t)
}

func TestUpdateTodo_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := new(mocks.MockTodoService)
	mockValidator := new(mocks.MockTodoValidator)
	mockAuthService := new(mocks.MockAuthService)

	reqBody := dto.UpdateTodoRequest{
		ID:          1,
		Title:       "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.", // assuming title is required so this is invalid
		Description: "Updated Description",
	}

	// Setup mocks.
	mockAuthService.
		On("GetUserIDFromContext", mock.AnythingOfType("*gin.Context")).
		Return(uint(1), nil)
	mockValidator.
		On("ValidateUpdateTodoRequest", &reqBody).
		Return(errors.New("title too long"))

	// Inject JWT claim.
	router.Use(func(c *gin.Context) {
		claims := jwt.MapClaims{"user_id": float64(1)}
		c.Set("jwt", claims)
		c.Next()
	})

	todoHandler := NewTodoHandler(mockService, mockValidator, mockAuthService)
	router.PUT("/todos/1", todoHandler.UpdateTodo)

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, "/todos/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	respRecorder := httptest.NewRecorder()
	router.ServeHTTP(respRecorder, req)

	// Expect 400 Bad Request.
	assert.Equal(t, http.StatusBadRequest, respRecorder.Code)
	assert.Contains(t, respRecorder.Body.String(), "title too long")
	mockValidator.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestUpdateTodo_TodoNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := new(mocks.MockTodoService)
	mockValidator := new(mocks.MockTodoValidator)
	mockAuthService := new(mocks.MockAuthService)

	reqBody := dto.UpdateTodoRequest{
		ID:          99, // assuming this ID does not exist
		Title:       "Updated Title",
		Description: "Updated Description",
	}

	mockAuthService.
		On("GetUserIDFromContext", mock.AnythingOfType("*gin.Context")).
		Return(uint(1), nil)

	mockValidator.
		On("ValidateUpdateTodoRequest", &reqBody).
		Return(nil)

	// Simulate that the todo is not found.
	mockService.
		On("GetTodoByIDAndUserID", uint(reqBody.ID), uint(1)).
		Return((*models.Todo)(nil), errors.New("not found"))

	router.Use(func(c *gin.Context) {
		claims := jwt.MapClaims{"user_id": float64(1)}
		c.Set("jwt", claims)
		c.Next()
	})

	todoHandler := NewTodoHandler(mockService, mockValidator, mockAuthService)
	router.PUT("/todos/99", todoHandler.UpdateTodo)

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, "/todos/99", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	respRecorder := httptest.NewRecorder()
	router.ServeHTTP(respRecorder, req)

	// Expect 404 Not Found.
	assert.Equal(t, http.StatusNotFound, respRecorder.Code)
	assert.Contains(t, respRecorder.Body.String(), "todo not found")
	mockValidator.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
	mockService.AssertExpectations(t)
}

func TestUpdateTodo_UpdateFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := new(mocks.MockTodoService)
	mockValidator := new(mocks.MockTodoValidator)
	mockAuthService := new(mocks.MockAuthService)

	reqBody := dto.UpdateTodoRequest{
		ID:          1,
		Title:       "Updated Title",
		Description: "Updated Description",
	}

	mockAuthService.
		On("GetUserIDFromContext", mock.AnythingOfType("*gin.Context")).
		Return(uint(1), nil)

	mockValidator.
		On("ValidateUpdateTodoRequest", &reqBody).
		Return(nil)

	// Simulate that the todo exists.
	mockService.
		On("GetTodoByIDAndUserID", uint(reqBody.ID), uint(1)).
		Return(&models.Todo{ID: 1, Title: "Old Title", Description: "Old Desc", UserID: 1}, nil)

	// Simulate an update failure.
	mockService.
		On("UpdateTodo", mock.AnythingOfType("*models.Todo")).
		Return(errors.New("update failed"))

	router.Use(func(c *gin.Context) {
		claims := jwt.MapClaims{"user_id": float64(1)}
		c.Set("jwt", claims)
		c.Next()
	})

	todoHandler := NewTodoHandler(mockService, mockValidator, mockAuthService)
	router.PUT("/todos/1", todoHandler.UpdateTodo)

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, "/todos/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	respRecorder := httptest.NewRecorder()
	router.ServeHTTP(respRecorder, req)

	// Expect 500 Internal Server Error.
	assert.Equal(t, http.StatusInternalServerError, respRecorder.Code)
	assert.Contains(t, respRecorder.Body.String(), "failed to update todo")
	mockAuthService.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
	mockService.AssertExpectations(t)
}

//! ****************** DeleteTodo *****************

func TestDeleteTodo_Success(t *testing.T) {
	// Setup Gin in test mode
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create mocks for service and auth service.
	mockService := new(mocks.MockTodoService)
	mockAuthService := new(mocks.MockAuthService)

	// Prepare a dummy Todo that belongs to userID 1.
	dummyTodo := models.Todo{
		ID:          uint(1),
		Title:       "Test Todo",
		Description: "Test Description",
		UserID:      uint(1),
	}

	// Set expectations.
	mockAuthService.
		On("GetUserIDFromContext", mock.AnythingOfType("*gin.Context")).
		Return(uint(1), nil)
	mockService.
		On("GetTodoByIDAndUserID", uint(1), uint(1)).
		Return(&dummyTodo, nil)
	mockService.
		On("DeleteItemWithItsAttachments", dummyTodo.ID).
		Return(nil)

	// Initialize the handler.
	todoHandler := NewTodoHandler(mockService, nil, mockAuthService)
	router.DELETE("/todos/:id", todoHandler.DeleteTodo)

	// Create a DELETE request.
	req, _ := http.NewRequest(http.MethodDelete, "/todos/1", nil)
	respRecorder := httptest.NewRecorder()

	// Execute the request.
	router.ServeHTTP(respRecorder, req)

	// Assert the response.
	assert.Equal(t, http.StatusNoContent, respRecorder.Code)
	mockAuthService.AssertExpectations(t)
	mockService.AssertExpectations(t)
}

func TestDeleteTodo_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := new(mocks.MockTodoService)
	mockAuthService := new(mocks.MockAuthService)

	// Simulate an invalid token scenario.
	mockAuthService.
		On("GetUserIDFromContext", mock.AnythingOfType("*gin.Context")).
		Return(uint(0), errors.New("unauthorized"))

	todoHandler := NewTodoHandler(mockService, nil, mockAuthService)
	router.DELETE("/todos/:id", todoHandler.DeleteTodo)

	req, _ := http.NewRequest(http.MethodDelete, "/todos/1", nil)
	respRecorder := httptest.NewRecorder()
	router.ServeHTTP(respRecorder, req)

	assert.Equal(t, http.StatusUnauthorized, respRecorder.Code)
	assert.Contains(t, respRecorder.Body.String(), "unauthorized")
	mockAuthService.AssertExpectations(t)
}

func TestDeleteTodo_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := new(mocks.MockTodoService)
	mockAuthService := new(mocks.MockAuthService)

	// Even if the auth is valid, the invalid todo id should be caught.
	mockAuthService.
		On("GetUserIDFromContext", mock.AnythingOfType("*gin.Context")).
		Return(uint(1), nil)

	todoHandler := NewTodoHandler(mockService, nil, mockAuthService)
	router.DELETE("/todos/:id", todoHandler.DeleteTodo)

	// Pass a non-numeric id.
	req, _ := http.NewRequest(http.MethodDelete, "/todos/abc", nil)
	respRecorder := httptest.NewRecorder()
	router.ServeHTTP(respRecorder, req)

	assert.Equal(t, http.StatusBadRequest, respRecorder.Code)
	assert.Contains(t, respRecorder.Body.String(), "invalid todo id")
	mockAuthService.AssertExpectations(t)
}

func TestDeleteTodo_TodoNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := new(mocks.MockTodoService)
	mockAuthService := new(mocks.MockAuthService)

	// Valid auth.
	mockAuthService.
		On("GetUserIDFromContext", mock.AnythingOfType("*gin.Context")).
		Return(uint(1), nil)
	// Simulate todo not found.
	mockService.
		On("GetTodoByIDAndUserID", uint(1), uint(1)).
		Return(models.Todo{}, errors.New("todo not found"))

	todoHandler := NewTodoHandler(mockService, nil, mockAuthService)
	router.DELETE("/todos/:id", todoHandler.DeleteTodo)

	req, _ := http.NewRequest(http.MethodDelete, "/todos/1", nil)
	respRecorder := httptest.NewRecorder()
	router.ServeHTTP(respRecorder, req)

	assert.Equal(t, http.StatusNotFound, respRecorder.Code)
	assert.Contains(t, respRecorder.Body.String(), "todo not found")
	mockAuthService.AssertExpectations(t)
	mockService.AssertExpectations(t)
}

func TestDeleteTodo_DeleteError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := new(mocks.MockTodoService)
	mockAuthService := new(mocks.MockAuthService)

	// Dummy todo for deletion.
	dummyTodo := models.Todo{
		ID:          uint(1),
		Title:       "Test Todo",
		Description: "Test Description",
		UserID:      uint(1),
	}

	mockAuthService.
		On("GetUserIDFromContext", mock.AnythingOfType("*gin.Context")).
		Return(uint(1), nil)
	mockService.
		On("GetTodoByIDAndUserID", uint(1), uint(1)).
		Return(&dummyTodo, nil)
	// Simulate error during deletion.
	mockService.
		On("DeleteItemWithItsAttachments", dummyTodo.ID).
		Return(errors.New("failed to delete todo"))

	todoHandler := NewTodoHandler(mockService, nil, mockAuthService)
	router.DELETE("/todos/:id", todoHandler.DeleteTodo)

	req, _ := http.NewRequest(http.MethodDelete, "/todos/1", nil)
	respRecorder := httptest.NewRecorder()
	router.ServeHTTP(respRecorder, req)

	assert.Equal(t, http.StatusInternalServerError, respRecorder.Code)
	assert.Contains(t, respRecorder.Body.String(), "failed to delete todo")
	mockAuthService.AssertExpectations(t)
	mockService.AssertExpectations(t)
}
