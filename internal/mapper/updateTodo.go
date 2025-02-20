package mapper

import (
	"time"

	"github.com/deepjyotk/todo-with-golang/internal/dto"
	"github.com/deepjyotk/todo-with-golang/internal/models"
)

// MapCreateTodoRequestToTodo transforms a CreateTodoRequest DTO into a Todo domain model.
// It takes the incoming DTO and the authenticated user's ID, then maps the fields accordingly.
func MapUpdateTodoRequestToTodo(req dto.UpdateTodoRequest, userID uint) models.Todo {
	// Map the attachment URLs from the DTO to the models.Attachment slice.
	var attachments []models.Attachment
	for _, url := range req.AttachmentURLs {
		attachments = append(attachments, models.Attachment{
			AttachmentURL: url, // Corrected field name
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			// ItemID will be set by the database or later in the service logic after the Todo is created.
		})
	}

	return models.Todo{
		ID:          req.ID,
		Title:       req.Title,
		Description: req.Description,
		UserID:      userID,
		Attachments: attachments,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}
