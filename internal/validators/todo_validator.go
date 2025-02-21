// internal/validators/todo_validator.go
package validators

import (
	"errors"

	"github.com/deepjyotk/todo-with-golang/configs"
	"github.com/deepjyotk/todo-with-golang/internal/dto"
)

// TodoValidatorInterface defines the contract for validating todo requests.
type TodoValidatorInterface interface {
	ValidateCreateTodoRequest(req *dto.CreateTodoRequest) error
	ValidateUpdateTodoRequest(req *dto.UpdateTodoRequest) error
	ValidFileName(fileName string) error
}

// TodoValidator holds any dependencies (like configuration) required for validation.
type TodoValidator struct {
	Config *configs.Config
}

// NewTodoValidator is a constructor for TodoValidator that returns the interface.
func NewTodoValidator(cfg *configs.Config) TodoValidatorInterface {
	return &TodoValidator{Config: cfg}
}

// ValidateCreateTodoRequest validates the CreateTodoRequest payload.
func (v *TodoValidator) ValidateCreateTodoRequest(req *dto.CreateTodoRequest) error {
	// Basic non-empty checks.
	if req.Title == "" {
		return errors.New("title is required")
	}
	if req.Description == "" {
		return errors.New("description is required")
	}

	// Validate title with helper: len > 2 and < 100.
	if err := ValidateTitle(req.Title); err != nil {
		return err
	}

	// Validate description: len < 10000.
	if err := ValidateDescription(req.Description); err != nil {
		return err
	}

	// Validate each attachment URL: len < 100.
	for _, url := range req.AttachmentURLs {
		if err := ValidateAttachmentURL(url, v.Config.S3.Bucket, v.Config.S3.Region); err != nil {
			return err
		}
	}
	return nil
}

// ValidateUpdateTodoRequest validates the UpdateTodoRequest payload.
func (v *TodoValidator) ValidateUpdateTodoRequest(req *dto.UpdateTodoRequest) error {
	// Check if ID is non-zero.
	if req.ID == 0 {
		return errors.New("id is required and must be a non-zero unsigned integer")
	}

	// Basic non-empty checks.
	if req.Title == "" {
		return errors.New("title is required")
	}
	if req.Description == "" {
		return errors.New("description is required")
	}

	// Validate title with helper: len > 2 and < 100.
	if err := ValidateTitle(req.Title); err != nil {
		return err
	}

	// Validate description: len < 10000.
	if err := ValidateDescription(req.Description); err != nil {
		return err
	}

	// Validate each attachment URL: len < 100.
	for _, url := range req.AttachmentURLs {
		if err := ValidateAttachmentURL(url, v.Config.S3.Bucket, v.Config.S3.Region); err != nil {
			return err
		}
	}
	return nil
}

// ValidFileName validates the given file name.
func (v *TodoValidator) ValidFileName(fileName string) error {
	// Basic non-empty checks using the helper.
	if err := ValidFileWithExtension(fileName); err != nil {
		return err
	}
	return nil
}
