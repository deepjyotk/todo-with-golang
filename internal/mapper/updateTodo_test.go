package mapper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/deepjyotk/todo-with-golang/internal/dto"
)

func TestMapUpdateTodoRequestToTodo(t *testing.T) {
	// Define test input
	userID := uint(2)
	req := dto.UpdateTodoRequest{
		ID:             1,
		Title:          "Updated Todo",
		Description:    "This is an updated test todo item",
		AttachmentURLs: []string{"https://example.com/updated_attachment1.png", "https://example.com/updated_attachment2.png"},
	}

	// Call the function
	result := MapUpdateTodoRequestToTodo(req, userID)

	// Assert that fields are correctly mapped
	assert.Equal(t, req.ID, result.ID)
	assert.Equal(t, req.Title, result.Title)
	assert.Equal(t, req.Description, result.Description)
	assert.Equal(t, userID, result.UserID)
	assert.Len(t, result.Attachments, len(req.AttachmentURLs))

	// Check attachment URLs
	for i, attachment := range result.Attachments {
		assert.Equal(t, req.AttachmentURLs[i], attachment.AttachmentURL)
	}

	// Ensure timestamps are correctly assigned
	assert.WithinDuration(t, time.Now(), result.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), result.UpdatedAt, time.Second)
}
