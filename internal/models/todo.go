// internal/models/todo.go
package models

import (
	"time"
)

// _Todo represents a TODO item belonging to a user.
// swagger:model Todo
type Todo struct {
	// The unique identifier of the todo item.
	// example: 1
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// Title of the todo item.
	// required: true
	// example: Buy groceries
	Title string `gorm:"type:varchar(255);not null" json:"title"`

	// Description of the todo item.
	// example: Milk, eggs, bread, and coffee.
	Description string `gorm:"type:text" json:"description"`

	// UserID is the foreign key referencing the owner user.
	// required: true
	// example: 42
	UserID uint `gorm:"index;not null" json:"user_id"`

	// Attachments holds the related attachments (images/PDFs).
	Attachments []Attachment `gorm:"foreignKey:ItemID" json:"attachments,omitempty"`

	// CreatedAt timestamp
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt timestamp
	UpdatedAt time.Time `json:"updated_at"`
}

// Attachment represents an attachment for a todo item.
// swagger:model Attachment
type Attachment struct {
	// The unique identifier of the attachment.
	// example: 1
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// AttachmentURL is the S3 URL where the attachment is stored.
	// required: true
	// example: https://s3.amazonaws.com/yourbucket/attachment.jpg
	AttachmentURL string `gorm:"type:text;not null" json:"attachment_url"`

	// ItemID is the foreign key referencing the parent todo item.
	// required: true
	// example: 1
	ItemID uint `gorm:"index;not null" json:"item_id"`

	// CreatedAt timestamp
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt timestamp
	UpdatedAt time.Time `json:"updated_at"`
}
