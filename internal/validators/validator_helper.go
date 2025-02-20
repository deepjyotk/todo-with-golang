// internal/validators/validator_helper.go
package validators

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ValidateTitle checks that the title has a length >2 and <100.
func ValidateTitle(title string) error {
	title = strings.TrimSpace(title)
	if len(title) < 3 {
		return errors.New("title must be at least 3 characters long")
	}
	if len(title) > 100 {
		return errors.New("title must be less than 100 characters long")
	}
	return nil
}

// ValidateDescription checks that the description length is less than 10000.
func ValidateDescription(description string) error {
	if len(description) > 10000 {
		return errors.New("description must be less than 10000 characters long")
	}
	return nil
}

// ValidateAttachmentURL checks that an attachment URL is less than 100 characters
// and follows the pattern:
// https://{bucketName}.s3.{bucketRegion}.amazonaws.com/{number}/{filename-with-extension}
func ValidateAttachmentURL(url, bucketName, bucketRegion string) error {
	url = strings.TrimSpace(url)
	if len(url) > 100 {
		return fmt.Errorf("attachment URL must be less than 100 characters: %s", url)
	}
	// Build the regex pattern dynamically using the bucket name and region.
	// regexp.QuoteMeta is used to escape any special characters in the bucketName/region.
	pattern := fmt.Sprintf(
		`^https:\/\/%s\.s3\.%s\.amazonaws\.com\/(\d+)\/([^\/]+\.[a-zA-Z0-9]+)$`,
		regexp.QuoteMeta(bucketName),
		regexp.QuoteMeta(bucketRegion),
	)

	re := regexp.MustCompile(pattern)
	if !re.MatchString(url) {
		return fmt.Errorf("attachment URL does not match expected pattern: %s. Hint: it should be in the format https://{bucketName}.s3.{bucketRegion}.amazonaws.com/{number}/{filename-with-extension}", url)
	}
	return nil
}

// ValidFileWithExtension checks if the given fileName is valid and has one of the allowed extensions: pdf, jpeg, jpg, png.
func ValidFileWithExtension(fileName string) error {
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		return errors.New("filename cannot be empty")
	}

	// This regex does the following:
	// 1. (?i) enables case-insensitive matching.
	// 2. ^[^\\/]+ ensures that the filename does not contain any forward or back slashes.
	// 3. \. matches the literal dot.
	// 4. (pdf|jpeg|jpg|png)$ ensures that the filename ends with one of the allowed extensions.
	pattern := `(?i)^[^\\/]+\.(pdf|jpeg|jpg|png)$`
	re := regexp.MustCompile(pattern)
	if !re.MatchString(fileName) {
		return fmt.Errorf("filename %q is invalid; it must be valid and have one of the extensions: pdf, jpeg, jpg, png", fileName)
	}
	return nil
}
