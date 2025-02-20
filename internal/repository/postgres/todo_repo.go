// internal/repository/postgres/todo_repo.go
package postgres

import (
	"github.com/deepjyotk/todo-with-golang/internal/models"
	"gorm.io/gorm"
)

// TodoRepository defines the interface for todo data operations.
type TodoRepository interface {
	Create(todo *models.Todo) error
	GetByID(id uint) (*models.Todo, error)
	GetByUser(userID uint) ([]models.Todo, error)
	Update(todo *models.Todo) error
	Delete(todo *models.Todo) error
	CreateAttachment(attachment *models.Attachment) error
	DeleteAttachmentsByItemID(itemID uint) error
	GetAttachmentsByTodo(todoID uint) ([]models.Attachment, error)
	DeleteByID(id uint) error
}

type todoRepository struct {
	db *gorm.DB
}

// NewTodoRepository returns a new instance of TodoRepository.
func NewTodoRepository(db *gorm.DB) TodoRepository {
	return &todoRepository{db: db}
}

func (r *todoRepository) Create(todo *models.Todo) error {
	return r.db.Create(todo).Error
}

func (r *todoRepository) GetByID(id uint) (*models.Todo, error) {
	var todo models.Todo
	if err := r.db.Preload("Attachments").First(&todo, id).Error; err != nil {
		return nil, err
	}
	return &todo, nil
}

func (r *todoRepository) GetByUser(userID uint) ([]models.Todo, error) {
	var todos []models.Todo
	if err := r.db.Preload("Attachments").Where("user_id = ?", userID).Find(&todos).Error; err != nil {
		return nil, err
	}
	return todos, nil
}

func (r *todoRepository) DeleteAttachmentsByItemID(itemID uint) error {
	// This will delete all attachments where item_id = itemID
	if err := r.db.Where("item_id = ?", itemID).Delete(&models.Attachment{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *todoRepository) Update(todo *models.Todo) error {
	return r.db.Save(todo).Error
}

func (r *todoRepository) Delete(todo *models.Todo) error {
	return r.db.Delete(todo).Error
}

func (r *todoRepository) CreateAttachment(attachment *models.Attachment) error {
	return r.db.Create(attachment).Error
}

func (r *todoRepository) GetAttachmentsByTodo(todoID uint) ([]models.Attachment, error) {
	var attachments []models.Attachment
	if err := r.db.Where("item_id = ?", todoID).Find(&attachments).Error; err != nil {
		return nil, err
	}
	return attachments, nil
}

// DeleteByID removes a todo item from the database by its ID.
func (r *todoRepository) DeleteByID(id uint) error {
	if err := r.db.Where("id = ?", id).Delete(&models.Todo{}).Error; err != nil {
		return err
	}
	return nil
}
