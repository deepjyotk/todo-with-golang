package mocks

import (
	"github.com/deepjyotk/todo-with-golang/internal/models"

	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUserByEmail(email string) (*models.User, error) {

	args := m.Called(email)

	if args.Get(0) != nil {

		return args.Get(0).(*models.User), args.Error(1)

	}

	return nil, args.Error(1)

}

func (m *MockUserRepository) CreateUser(user *models.User) error {

	args := m.Called(user)

	return args.Error(0)

}
