package validators

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateTitle(t *testing.T) {
	tests := []struct {
		name          string
		title         string
		expectedError error
	}{
		{
			name:          "Valid title",
			title:         "A Valid Title",
			expectedError: nil,
		},
		{
			name:          "Title with leading and trailing spaces",
			title:         "   Valid Title   ",
			expectedError: nil,
		},
		{
			name:          "Title exactly 3 characters",
			title:         "abc",
			expectedError: nil,
		},
		{
			name:          "Title too short",
			title:         "ab", // trimmed length 2 (< 3)
			expectedError: errors.New("title must be at least 3 characters long"),
		},
		{
			name:          "Title empty after trim",
			title:         "    ",
			expectedError: errors.New("title must be at least 3 characters long"),
		},
		{
			name:          "Title exactly 100 characters",
			title:         strings.Repeat("a", 100),
			expectedError: nil,
		},
		{
			name:          "Title too long (101 characters)",
			title:         strings.Repeat("a", 101),
			expectedError: errors.New("title must be less than 100 characters long"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTitle(tt.title)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDescription(t *testing.T) {
	tests := []struct {
		name          string
		description   string
		expectedError error
	}{
		{
			name:          "Valid description",
			description:   "This is a valid description.",
			expectedError: nil,
		},
		{
			name:          "Empty description is allowed",
			description:   "",
			expectedError: nil,
		},
		{
			name:          "Description exactly 10000 characters",
			description:   strings.Repeat("a", 10000),
			expectedError: nil,
		},
		{
			name:          "Description too long (10001 characters)",
			description:   strings.Repeat("a", 10001),
			expectedError: errors.New("description must be less than 10000 characters long"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDescription(tt.description)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAttachmentURL(t *testing.T) {
	// Use dummy bucket values for testing.
	bucketName := "dummy-bucket"
	bucketRegion := "dummy-region"
	validURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/1234/file.jpg", bucketName, bucketRegion)
	validURLAlt := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/5678/my_file.png", bucketName, bucketRegion)

	tests := []struct {
		name          string
		url           string
		expectedError error
	}{
		{
			name:          "Valid attachment URL",
			url:           validURL,
			expectedError: nil,
		},
		{
			name:          "Valid attachment URL with alternate filename",
			url:           validURLAlt,
			expectedError: nil,
		},
		{
			name: "URL too long",
			url: fmt.Sprintf("https://%s.s3.%s.amazonaws.com/1234/%s",
				bucketName, bucketRegion, strings.Repeat("a", 90)), // will exceed 100 characters
			expectedError: fmt.Errorf("attachment URL must be less than 100 characters: %s",
				strings.TrimSpace(fmt.Sprintf("https://%s.s3.%s.amazonaws.com/1234/%s",
					bucketName, bucketRegion, strings.Repeat("a", 90)))),
		},
		{
			name: "Wrong protocol (http instead of https)",
			url:  strings.Replace(validURL, "https://", "http://", 1),
			expectedError: fmt.Errorf("attachment URL does not match expected pattern: %s. Hint: it should be in the format https://{bucketName}.s3.{bucketRegion}.amazonaws.com/{number}/{filename-with-extension}",
				strings.Replace(validURL, "https://", "http://", 1)),
		},
		{
			name: "Missing digit segment",
			url:  fmt.Sprintf("https://%s.s3.%s.amazonaws.com//file.jpg", bucketName, bucketRegion),
			expectedError: fmt.Errorf("attachment URL does not match expected pattern: %s. Hint: it should be in the format https://{bucketName}.s3.{bucketRegion}.amazonaws.com/{number}/{filename-with-extension}",
				fmt.Sprintf("https://%s.s3.%s.amazonaws.com//file.jpg", bucketName, bucketRegion)),
		},
		{
			name: "Invalid filename (no extension)",
			url:  fmt.Sprintf("https://%s.s3.%s.amazonaws.com/1234/file", bucketName, bucketRegion),
			expectedError: fmt.Errorf("attachment URL does not match expected pattern: %s. Hint: it should be in the format https://{bucketName}.s3.{bucketRegion}.amazonaws.com/{number}/{filename-with-extension}",
				fmt.Sprintf("https://%s.s3.%s.amazonaws.com/1234/file", bucketName, bucketRegion)),
		},
		{
			name: "Extra path segments in URL",
			url:  fmt.Sprintf("https://%s.s3.%s.amazonaws.com/1234/path/file.jpg", bucketName, bucketRegion),
			expectedError: fmt.Errorf("attachment URL does not match expected pattern: %s. Hint: it should be in the format https://{bucketName}.s3.{bucketRegion}.amazonaws.com/{number}/{filename-with-extension}",
				fmt.Sprintf("https://%s.s3.%s.amazonaws.com/1234/path/file.jpg", bucketName, bucketRegion)),
		},
		{
			name:          "URL with leading/trailing spaces",
			url:           "   " + validURL + "   ",
			expectedError: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAttachmentURL(tt.url, bucketName, bucketRegion)
			if tt.expectedError != nil {
				assert.Error(t, err)
				// Since the error message is dynamically built, we verify that it contains key substrings.
				if strings.HasPrefix(tt.expectedError.Error(), "attachment URL must be less than 100 characters") {
					assert.Contains(t, err.Error(), "attachment URL must be less than 100 characters")
				} else if strings.HasPrefix(tt.expectedError.Error(), "attachment URL does not match expected pattern") {
					assert.Contains(t, err.Error(), "attachment URL does not match expected pattern")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidFileWithExtension(t *testing.T) {
	tests := []struct {
		name          string
		fileName      string
		expectedError error
	}{
		{
			name:          "Valid PDF file",
			fileName:      "document.pdf",
			expectedError: nil,
		},
		{
			name:          "Valid JPEG file (case insensitive)",
			fileName:      "image.JpEg",
			expectedError: nil,
		},
		{
			name:          "Valid file with multiple dots",
			fileName:      "my.document.pdf",
			expectedError: nil,
		},
		{
			name:          "Empty file name",
			fileName:      "   ",
			expectedError: errors.New("filename cannot be empty"),
		},
		{
			name:          "Missing extension",
			fileName:      "file",
			expectedError: fmt.Errorf("filename %q is invalid; it must be valid and have one of the extensions: pdf, jpeg, jpg, png", "file"),
		},
		{
			name:          "Disallowed extension",
			fileName:      "file.txt",
			expectedError: fmt.Errorf("filename %q is invalid; it must be valid and have one of the extensions: pdf, jpeg, jpg, png", "file.txt"),
		},
		{
			name:          "Filename with a forward slash",
			fileName:      "folder/file.pdf",
			expectedError: fmt.Errorf("filename %q is invalid; it must be valid and have one of the extensions: pdf, jpeg, jpg, png", "folder/file.pdf"),
		},
		{
			name:          "Filename with a backslash",
			fileName:      "folder\\file.pdf",
			expectedError: fmt.Errorf("filename %q is invalid; it must be valid and have one of the extensions: pdf, jpeg, jpg, png", "folder\\file.pdf"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidFileWithExtension(tt.fileName)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
