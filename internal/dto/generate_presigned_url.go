// internal/dto/generate_presigned_url.go
package dto

// GeneratePresignedURLRequest represents the expected JSON payload for generating a pre-signed URL.
type GeneratePresignedURLRequest struct {
	// FileName is the name of the file the user intends to upload.
	FileName string `json:"file_name" binding:"required"`
	// Additional fields like ContentType can be added if needed.
	// ContentType string `json:"content_type"`
}

// GeneratePresignedURLResponse represents the response containing the pre-signed URL.
type GeneratePresignedURLResponse struct {
	// URL is the generated pre-signed S3 URL.
	URL string `json:"url"`
}
