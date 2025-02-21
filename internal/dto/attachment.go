package dto

// Attachment represents an attachment for a Todo item.
type Attachment struct {
	// Define the fields needed for your attachment.
	URL  string `json:"url"`
	Name string `json:"name"`
}
