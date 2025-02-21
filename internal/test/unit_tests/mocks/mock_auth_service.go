package mocks

import (
	"fmt"

	"github.com/deepjyotk/todo-with-golang/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

// MockAuthService is a mock implementation of services.AuthService.
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(username, email, password string) (*models.User, error) {
	args := m.Called(username, email, password)
	if user, ok := args.Get(0).(*models.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthService) Login(email, password string) (string, error) {
	args := m.Called(email, password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) GenerateToken(user *models.User) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

//	func (m *MockAuthService) GetUserIDFromContext(c *gin.Context) (uint, error) {
//		args := m.Called(c)
//		// if args.Get(0) == nil {
//		// 	return 0, args.Error(1) // Ensure it doesn't panic if nil
//		// }
//		return (args.Int(0)), args.Error(1)
//	}
func (m *MockAuthService) GetUserIDFromContext(c *gin.Context) (uint, error) {
	args := m.Called(c)
	if args.Get(0) == nil {
		return 0, args.Error(1)
	}
	userID, ok := args.Get(0).(uint)
	if !ok {
		return 0, fmt.Errorf("expected uint but got %T", args.Get(0))
	}
	return userID, args.Error(1)
}
