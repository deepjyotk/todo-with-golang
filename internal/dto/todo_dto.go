// internal/dto/todo_dto.go

package dto

// CreateTodoRequest represents the payload for creating a TODO item.
type CreateTodoRequest struct {
	Title          string   `json:"title" binding:"required"`
	Description    string   `json:"description" binding:"required"`
	AttachmentURLs []string `json:"attachment_urls" binding:"omitempty,dive,url"`
}

type UpdateTodoRequest struct {
	ID             uint     `json:"id" binding:"required"`
	Title          string   `json:"title" binding:"required"`
	Description    string   `json:"description" binding:"required"`
	AttachmentURLs []string `json:"attachment_urls" binding:"omitempty,dive,url"`
}
