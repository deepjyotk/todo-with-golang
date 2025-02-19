package postgres

import (
	"errors"

	"github.com/deepjyotk/todo-with-golang/internal/models"
	"gorm.io/gorm"
)

// UserRepository defines the methods to interact with the User storage.
type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository returns a new instance of UserRepository.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// CreateUser persists a new user in the database.
func (r *userRepository) CreateUser(user *models.User) error {
	result := r.db.Create(user)
	return result.Error
}

// GetUserByEmail fetches a user by email.
func (r *userRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := r.db.Where("email = ?", email).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, result.Error
}
