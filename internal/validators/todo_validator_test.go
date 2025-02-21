package validators

import (
	"strings"
	"testing"

	"github.com/deepjyotk/todo-with-golang/configs"
	"github.com/deepjyotk/todo-with-golang/internal/dto"

	"github.com/stretchr/testify/assert"
)

func dummyConfig() *configs.Config {
	return &configs.Config{
		S3: configs.S3Config{
			Bucket: "todo-go-app-s3-bucket",
			Region: "us-east-1",
		},
	}
}

func TestValidateCreateTodoRequest(t *testing.T) {
	validator := NewTodoValidator(dummyConfig())

	// Valid case.

	validReq := &dto.CreateTodoRequest{
		Title:          "A Valid Title",
		Description:    "A valid description",
		AttachmentURLs: []string{"https://todo-go-app-s3-bucket.s3.us-east-1.amazonaws.com/1/fi2l123.jpeg"},
	}
	err := validator.ValidateCreateTodoRequest(validReq)
	assert.NoError(t, err)

	// Empty title.
	emptyTitleReq := &dto.CreateTodoRequest{
		Title:       "",
		Description: "Some description",
	}
	err = validator.ValidateCreateTodoRequest(emptyTitleReq)
	assert.EqualError(t, err, "title is required")

	// Empty description.
	emptyDescReq := &dto.CreateTodoRequest{
		Title:       "A Valid Title",
		Description: "",
	}
	err = validator.ValidateCreateTodoRequest(emptyDescReq)
	assert.EqualError(t, err, "description is required")

	// Title too short (assuming ValidateTitle enforces a minimum length >2).
	shortTitleReq := &dto.CreateTodoRequest{
		Title:       "A", // too short
		Description: "Some description",
	}
	err = validator.ValidateCreateTodoRequest(shortTitleReq)
	assert.Error(t, err) // Expect error from ValidateTitle.

	// Title too long (assuming max length < 101).
	longTitle := strings.Repeat("a", 101)
	longTitleReq := &dto.CreateTodoRequest{
		Title:       longTitle,
		Description: "Some description",
	}
	err = validator.ValidateCreateTodoRequest(longTitleReq)
	assert.Error(t, err)

	// Description too long (assuming max length is < 10000).
	longDescription := strings.Repeat("a", 10001)
	longDescReq := &dto.CreateTodoRequest{
		Title:       "A Valid Title",
		Description: longDescription,
	}
	err = validator.ValidateCreateTodoRequest(longDescReq)
	assert.Error(t, err)

	// Invalid attachment URL.
	// (Assuming ValidateAttachmentURL returns an error if the URL is too long; e.g. length >= 100.)
	invalidURL := strings.Repeat("a", 101)
	invalidAttachmentReq := &dto.CreateTodoRequest{
		Title:          "A Valid Title",
		Description:    "A valid description",
		AttachmentURLs: []string{invalidURL},
	}
	err = validator.ValidateCreateTodoRequest(invalidAttachmentReq)
	assert.Error(t, err)
}

func TestValidateUpdateTodoRequest(t *testing.T) {
	validator := NewTodoValidator(dummyConfig())

	// Valid update request.
	validUpdateReq := &dto.UpdateTodoRequest{
		ID:             1,
		Title:          "A Valid Title",
		Description:    "A valid description",
		AttachmentURLs: []string{"https://todo-go-app-s3-bucket.s3.us-east-1.amazonaws.com/1/fi2l123.jpeg"},
	}
	err := validator.ValidateUpdateTodoRequest(validUpdateReq)
	assert.NoError(t, err)

	// ID is zero.
	zeroIDReq := &dto.UpdateTodoRequest{
		ID:          0,
		Title:       "A Valid Title",
		Description: "A valid description",
	}
	err = validator.ValidateUpdateTodoRequest(zeroIDReq)
	assert.EqualError(t, err, "id is required and must be a non-zero unsigned integer")

	// Empty title.
	emptyTitleReq := &dto.UpdateTodoRequest{
		ID:          1,
		Title:       "",
		Description: "A valid description",
	}
	err = validator.ValidateUpdateTodoRequest(emptyTitleReq)
	assert.EqualError(t, err, "title is required")

	// Empty description.
	emptyDescReq := &dto.UpdateTodoRequest{
		ID:          1,
		Title:       "A Valid Title",
		Description: "",
	}
	err = validator.ValidateUpdateTodoRequest(emptyDescReq)
	assert.EqualError(t, err, "description is required")

	// Title too short.
	shortTitleReq := &dto.UpdateTodoRequest{
		ID:          1,
		Title:       "A", // too short
		Description: "A valid description",
	}
	err = validator.ValidateUpdateTodoRequest(shortTitleReq)
	assert.Error(t, err)

	// Title too long.
	longTitle := strings.Repeat("a", 101)
	longTitleReq := &dto.UpdateTodoRequest{
		ID:          1,
		Title:       longTitle,
		Description: "A valid description",
	}
	err = validator.ValidateUpdateTodoRequest(longTitleReq)
	assert.Error(t, err)

	// Description too long.
	longDesc := strings.Repeat("a", 10001)
	longDescReq := &dto.UpdateTodoRequest{
		ID:          1,
		Title:       "A Valid Title",
		Description: longDesc,
	}
	err = validator.ValidateUpdateTodoRequest(longDescReq)
	assert.Error(t, err)

	// Invalid attachment URL.
	invalidURL := strings.Repeat("a", 101)
	invalidAttachmentReq := &dto.UpdateTodoRequest{
		ID:             1,
		Title:          "A Valid Title",
		Description:    "A valid description",
		AttachmentURLs: []string{invalidURL},
	}
	err = validator.ValidateUpdateTodoRequest(invalidAttachmentReq)
	assert.Error(t, err)
}

func TestValidFileName(t *testing.T) {
	validator := NewTodoValidator(dummyConfig())

	// Valid file name.
	err := validator.ValidFileName("file.jpg")
	assert.NoError(t, err)

	// Invalid file name: missing extension.
	err = validator.ValidFileName("file")
	assert.Error(t, err)

	// You may also check for disallowed extensions.
	// For example, if only image files are allowed:
	err = validator.ValidFileName("file.txt")
	assert.Error(t, err)
}
