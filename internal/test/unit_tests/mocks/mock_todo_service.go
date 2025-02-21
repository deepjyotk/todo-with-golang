package mocks

import (
	"github.com/deepjyotk/todo-with-golang/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockTodoService is a mock implementation of the TodoService interface.
type MockTodoService struct {
	mock.Mock
}

func (m *MockTodoService) CreateTodo(todo *models.Todo) error {
	args := m.Called(todo)
	return args.Error(0)
}

func (m *MockTodoService) AddAttachment(attachment *models.Attachment) error {
	args := m.Called(attachment)
	return args.Error(0)
}

func (m *MockTodoService) DeleteItemWithItsAttachments(todoID uint) error {
	args := m.Called(todoID)
	return args.Error(0)
}

func (m *MockTodoService) GeneratePresignedS3Url(userID uint, fileNameOrURL string, presignRequestType string) (string, error) {
	args := m.Called(userID, fileNameOrURL, presignRequestType)
	return args.String(0), args.Error(1)
}

func (m *MockTodoService) GetTodoByIDAndUserID(id uint, userID uint) (*models.Todo, error) {
	args := m.Called(id, userID)
	todo, _ := args.Get(0).(*models.Todo)
	return todo, args.Error(1)
}

func (m *MockTodoService) GetTodosByUser(userID uint) ([]models.Todo, error) {
	args := m.Called(userID)
	todos, _ := args.Get(0).([]models.Todo)
	return todos, args.Error(1)
}

func (m *MockTodoService) UpdateTodo(todo *models.Todo) error {
	args := m.Called(todo)
	return args.Error(0)
}
