package mocks

import (
	"github.com/deepjyotk/todo-with-golang/internal/dto"
	"github.com/stretchr/testify/mock"
)

// MockTodoValidator is a mock implementation of the TodoValidatorInterface.
type MockTodoValidator struct {
	mock.Mock
}

func (m *MockTodoValidator) ValidateCreateTodoRequest(req *dto.CreateTodoRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockTodoValidator) ValidFileName(fileName string) error {
	args := m.Called(fileName)
	return args.Error(0)
}

func (m *MockTodoValidator) ValidateUpdateTodoRequest(req *dto.UpdateTodoRequest) error {
	args := m.Called(req)
	return args.Error(0)
}
