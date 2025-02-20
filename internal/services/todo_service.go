// internal/services/todo_service.go
package services

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/deepjyotk/todo-with-golang/configs"
	"github.com/deepjyotk/todo-with-golang/internal/models"
	"github.com/deepjyotk/todo-with-golang/internal/repository/postgres"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// TodoService defines the business logic interface.
type TodoService interface {
	CreateTodo(todo *models.Todo) error
	GetTodoByIDAndUserID(id uint, userID uint) (*models.Todo, error)
	GetTodosByUser(userID uint) ([]models.Todo, error)
	UpdateTodo(todo *models.Todo) error
	DeleteItemWithItsAttachments(todo uint) error
	AddAttachment(attachment *models.Attachment) error
	GeneratePresignedS3UrlPutRequest(userID uint, fileName string) (string, error)
}

// todoService is the concrete implementation of TodoService.
type todoService struct {
	repo     postgres.TodoRepository
	s3Client *s3.S3
	bucket   string
}

// NewTodoService creates a new TodoService.
// It initializes the AWS SDK v1 S3 client using the provided configuration.
func NewTodoService(repo postgres.TodoRepository, s3Config configs.S3Config) TodoService {
	// Create a new AWS session with static credentials and region.
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(s3Config.Region),
		Credentials: credentials.NewStaticCredentials(s3Config.AccessKey, s3Config.SecretKey, ""),
	})
	if err != nil {
		panic("unable to create AWS session, " + err.Error())
	}

	return &todoService{
		repo:     repo,
		s3Client: s3.New(sess),
		bucket:   s3Config.Bucket,
	}
}

// CreateTodo creates a new Todo item.
func (s *todoService) CreateTodo(todo *models.Todo) error {
	//TODO: For safety we could have an extra check: is the attachment with passed url is really uploaded!! (in this is just demo skipping this step!)
	// Business validations can be added here.
	return s.repo.Create(todo)
}

// GetTodoByID retrieves a Todo item by its ID.
func (s *todoService) GetTodoByIDAndUserID(id uint, userID uint) (*models.Todo, error) {
	todo, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	// Check that the todo belongs to the provided userID.
	if todo.UserID != userID {
		return nil, errors.New("todo item not found for this user")
	}
	return todo, nil
}

// GetTodosByUser retrieves all Todo items for a specific user.
func (s *todoService) GetTodosByUser(userID uint) ([]models.Todo, error) {
	todos, err := s.repo.GetByUser(userID)
	if err != nil {
		return nil, err
	}
	return todos, nil
}

// UpdateTodo updates an existing Todo item.
func (s *todoService) UpdateTodo(todo *models.Todo) error {
	// 1. Check if the todo exists
	existing, err := s.repo.GetByID(todo.ID)
	if err != nil || existing == nil {
		return errors.New("todo item not found")
	}

	// --- Additional safety check about attachments can go here if needed ---
	// For instance: Validate that 'todo.AttachmentURL' has indeed been uploaded

	// 2. Delete all attachments for this Todo (refactored to its own function).
	if err := s.deleteAllAttachmentsForTodoItem(todo.ID); err != nil {
		return err
	}

	// 3. Update the Todo itself
	return s.repo.Update(todo)
}

// Completed: create a export function named deleteItemWithItsAttachments(todoID uint) --> this first calls deleteAllAttachmentsForTodoItem, and once all the attachments are deleted, it also delete from todos table, can you implement this

// DeleteItemWithItsAttachments deletes all attachments associated with the given todoID
// and then deletes the Todo item from the database.
func (s *todoService) DeleteItemWithItsAttachments(todoID uint) error {
	// Step 1: Delete all attachments associated with the Todo item
	if err := s.deleteAllAttachmentsForTodoItem(todoID); err != nil {
		return fmt.Errorf("failed to delete attachments for todo ID %d: %w", todoID, err)
	}

	// Step 2: Delete the Todo item from the database
	if err := s.repo.DeleteByID(todoID); err != nil {
		return fmt.Errorf("failed to delete todo item with ID %d: %w", todoID, err)
	}

	return nil
}

// deleteAllAttachmentsForTodoItem encapsulates fetching and removing attachments
// from both S3 and the database for a given todoID.
func (s *todoService) deleteAllAttachmentsForTodoItem(todoID uint) error {
	// 1. Fetch all the attachments for the todoID
	attachments, err := s.repo.GetAttachmentsByTodo(todoID)
	if err != nil {
		return fmt.Errorf("failed to fetch attachments: %w", err)
	}

	// 2. Delete each attachment from S3
	for _, attachment := range attachments {
		if err := s.deleteS3Attachment(attachment.AttachmentURL); err != nil {
			return fmt.Errorf("failed to delete attachment from S3: %w", err)
		}
	}

	// 3. Delete attachments from the database
	if err := s.repo.DeleteAttachmentsByItemID(todoID); err != nil {
		return fmt.Errorf("failed to delete attachments: %w", err)
	}

	return nil
}

// deleteS3Attachment deletes an object from your S3 bucket given the full URL.
func (s *todoService) deleteS3Attachment(attachmentURL string) error {
	// Parse the URL to extract the path, which is used as the key in S3
	parsedURL, err := url.Parse(attachmentURL)
	if err != nil {
		return fmt.Errorf("failed to parse attachment URL: %w", err)
	}

	// The path typically starts with a leading slash, e.g. "/<userID>/<fileName>"
	objectKey := strings.TrimPrefix(parsedURL.Path, "/")

	_, err = s.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object from S3: %w", err)
	}

	// (Optionally) wait until the object is deleted.
	// This ensures eventual consistency but adds overhead; use only if really required.
	err = s.s3Client.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("error while waiting for S3 object to be deleted: %w", err)
	}

	return nil
}

// AddAttachment adds an attachment to a Todo item.
func (s *todoService) AddAttachment(attachment *models.Attachment) error {
	// You may add validations (e.g., checking file type) here.
	return s.repo.CreateAttachment(attachment)
}

// GeneratePresignedS3UrlPutRequest generates an S3 presigned URL for the specified user and filename.
// It constructs the S3 object key using the user ID as the folder name.
func (s *todoService) GeneratePresignedS3UrlPutRequest(userID uint, fileName string) (string, error) {
	// Construct the object key using the user ID as the folder name.
	objectKey := strconv.Itoa(int(userID)) + "/" + fileName

	// Prepare the PutObjectInput.
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	}

	// Create a request for the PutObject operation.
	req, _ := s.s3Client.PutObjectRequest(input)
	// Generate a presigned URL with a 15-minute expiry.
	urlStr, err := req.Presign(15 * time.Minute)
	if err != nil {
		return "", err
	}

	return urlStr, nil
}
