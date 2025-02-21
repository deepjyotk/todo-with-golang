package todo_service

import (
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client/metadata"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/deepjyotk/todo-with-golang/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ----- Custom No-Op Retryer -----

// NoOpRetryer implements request.Retryer with no retries.
type NoOpRetryer struct{}

func (r NoOpRetryer) MaxRetries() int {
	return 0
}

func (r NoOpRetryer) RetryRules(rq *request.Request) time.Duration {
	return 0
}

func (r NoOpRetryer) ShouldRetry(rq *request.Request) bool {
	return false
}

func (r NoOpRetryer) RetryDelay(rq *request.Request) time.Duration {
	return 0
}

// ----- Mocks for the repository -----

type MockTodoRepository struct {
	mock.Mock
}

func (m *MockTodoRepository) Create(todo *models.Todo) error {
	args := m.Called(todo)
	return args.Error(0)
}

func (m *MockTodoRepository) GetByID(id uint) (*models.Todo, error) {
	args := m.Called(id)
	if t, ok := args.Get(0).(*models.Todo); ok {
		return t, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTodoRepository) GetByUser(userID uint) ([]models.Todo, error) {
	args := m.Called(userID)
	if todos, ok := args.Get(0).([]models.Todo); ok {
		return todos, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTodoRepository) Update(todo *models.Todo) error {
	args := m.Called(todo)
	return args.Error(0)
}

func (m *MockTodoRepository) DeleteByID(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTodoRepository) GetAttachmentsByTodo(todoID uint) ([]models.Attachment, error) {
	args := m.Called(todoID)
	if attachments, ok := args.Get(0).([]models.Attachment); ok {
		return attachments, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTodoRepository) DeleteAttachmentsByItemID(todoID uint) error {
	args := m.Called(todoID)
	return args.Error(0)
}

func (m *MockTodoRepository) CreateAttachment(attachment *models.Attachment) error {
	args := m.Called(attachment)
	return args.Error(0)
}

func (m *MockTodoRepository) Delete(todo *models.Todo) error {
	args := m.Called(todo)
	return args.Error(0)
}

// ----- Mocks for the S3 Client -----

type MockS3Client struct {
	mock.Mock
	s3iface.S3API
}

func (m *MockS3Client) DeleteObject(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	args := m.Called(input)
	return &s3.DeleteObjectOutput{}, args.Error(1)
}

func (m *MockS3Client) WaitUntilObjectNotExists(input *s3.HeadObjectInput) error {
	args := m.Called(input)
	return args.Error(0)
}

func (m *MockS3Client) PutObjectRequest(input *s3.PutObjectInput) (*request.Request, *s3.PutObjectOutput) {
	m.Called(input)
	// Create a dummy request using request.New.
	fakeReq := request.New(
		aws.Config{},
		metadata.ClientInfo{ServiceName: "s3", SigningName: "s3"},
		request.Handlers{},
		NoOpRetryer{},
		&request.Operation{
			Name:       "PutObject",
			HTTPMethod: "PUT",
			HTTPPath:   "/",
		},
		nil, // params
		nil, // data
	)
	fakeReq.Handlers.Sign.Clear()
	fakeReq.Handlers.Build.PushBack(func(r *request.Request) {
		r.HTTPRequest.URL, _ = url.Parse("https://dummy-s3-url/put/" + *input.Key)
	})
	return fakeReq, &s3.PutObjectOutput{}
}

func (m *MockS3Client) GetObjectRequest(input *s3.GetObjectInput) (*request.Request, *s3.GetObjectOutput) {
	m.Called(input)
	// Create a dummy request using request.New.
	fakeReq := request.New(
		aws.Config{},
		metadata.ClientInfo{ServiceName: "s3", SigningName: "s3"},
		request.Handlers{},
		NoOpRetryer{},
		&request.Operation{
			Name:       "GetObject",
			HTTPMethod: "GET",
			HTTPPath:   "/",
		},
		nil, // params
		nil, // data
	)
	fakeReq.Handlers.Sign.Clear()
	fakeReq.Handlers.Build.PushBack(func(r *request.Request) {
		r.HTTPRequest.URL, _ = url.Parse("https://dummy-s3-url/get/" + *input.Key)
	})
	return fakeReq, &s3.GetObjectOutput{}
}

// ----- Helper to initialize the service with mocks -----

func newTestTodoService(repo *MockTodoRepository, s3Client *MockS3Client, bucket string) TodoService {
	return &todoService{
		repo:     repo,
		s3Client: s3Client,
		bucket:   bucket,
	}
}

// ----- Unit Tests -----

func TestCreateTodo(t *testing.T) {
	mockRepo := new(MockTodoRepository)
	mockS3 := new(MockS3Client)
	service := newTestTodoService(mockRepo, mockS3, "test-bucket")

	todo := &models.Todo{ID: 1, UserID: 100}
	mockRepo.On("Create", todo).Return(nil)

	err := service.CreateTodo(todo)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGetTodoByIDAndUserID(t *testing.T) {
	mockRepo := new(MockTodoRepository)
	mockS3 := new(MockS3Client)
	service := newTestTodoService(mockRepo, mockS3, "test-bucket")

	// Successful case.
	expectedTodo := &models.Todo{ID: 1, UserID: 100}
	mockRepo.On("GetByID", uint(1)).Return(expectedTodo, nil)

	todo, err := service.GetTodoByIDAndUserID(1, 100)
	assert.NoError(t, err)
	assert.Equal(t, expectedTodo, todo)

	// Unauthorized access.
	mockRepo.On("GetByID", uint(2)).Return(&models.Todo{ID: 2, UserID: 101}, nil)
	todo, err = service.GetTodoByIDAndUserID(2, 100)
	assert.Error(t, err)
	assert.Nil(t, todo)
	mockRepo.AssertExpectations(t)
}

func TestGetTodosByUser(t *testing.T) {
	mockRepo := new(MockTodoRepository)
	mockS3 := new(MockS3Client)
	service := newTestTodoService(mockRepo, mockS3, "test-bucket")

	attachmentURL := "https://s3.amazonaws.com/test-bucket/100/file.txt"
	todo := models.Todo{
		ID:     uint(1),
		UserID: uint(100),
		Attachments: []models.Attachment{
			{ID: uint(10), AttachmentURL: attachmentURL},
		},
	}
	mockRepo.On("GetByUser", uint(100)).Return([]models.Todo{todo}, nil)

	// For GET calls we use an argument matcher.
	mockS3.On("GetObjectRequest", mock.AnythingOfType("*s3.GetObjectInput")).Return(&request.Request{}, &s3.GetObjectOutput{})

	todos, err := service.GetTodosByUser(100)
	assert.NoError(t, err)
	assert.Len(t, todos, 1)
	assert.Contains(t, todos[0].Attachments[0].AttachmentURL, "https://dummy-s3-url/get/")
	mockRepo.AssertExpectations(t)
	mockS3.AssertExpectations(t)
}

func TestUpdateTodo(t *testing.T) {
	mockRepo := new(MockTodoRepository)
	mockS3 := new(MockS3Client)
	service := newTestTodoService(mockRepo, mockS3, "test-bucket")

	existingTodo := &models.Todo{ID: 1, UserID: 100}
	updatedTodo := &models.Todo{ID: 1, UserID: 100, Title: "Updated"}

	mockRepo.On("GetByID", uint(1)).Return(existingTodo, nil)

	attachmentURL := "https://s3.amazonaws.com/test-bucket/100/file.txt"
	attachments := []models.Attachment{
		{ID: 10, AttachmentURL: attachmentURL},
	}
	mockRepo.On("GetAttachmentsByTodo", uint(1)).Return(attachments, nil)

	parsedURL, _ := url.Parse(attachmentURL)
	key := strings.TrimPrefix(parsedURL.Path, "/")
	mockS3.On("DeleteObject", &s3.DeleteObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String(key),
	}).Return(&s3.DeleteObjectOutput{}, nil)
	mockS3.On("WaitUntilObjectNotExists", &s3.HeadObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String(key),
	}).Return(nil)

	mockRepo.On("DeleteAttachmentsByItemID", uint(1)).Return(nil)
	mockRepo.On("Update", updatedTodo).Return(nil)

	err := service.UpdateTodo(updatedTodo)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockS3.AssertExpectations(t)
}

func TestDeleteItemWithItsAttachments(t *testing.T) {
	mockRepo := new(MockTodoRepository)
	mockS3 := new(MockS3Client)
	service := newTestTodoService(mockRepo, mockS3, "test-bucket")

	attachmentURL := "https://s3.amazonaws.com/test-bucket/100/file.txt"
	attachments := []models.Attachment{
		{ID: 10, AttachmentURL: attachmentURL},
	}
	mockRepo.On("GetAttachmentsByTodo", uint(1)).Return(attachments, nil)

	parsedURL, _ := url.Parse(attachmentURL)
	key := strings.TrimPrefix(parsedURL.Path, "/")
	mockS3.On("DeleteObject", &s3.DeleteObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String(key),
	}).Return(&s3.DeleteObjectOutput{}, nil)
	mockS3.On("WaitUntilObjectNotExists", &s3.HeadObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String(key),
	}).Return(nil)

	mockRepo.On("DeleteAttachmentsByItemID", uint(1)).Return(nil)
	mockRepo.On("DeleteByID", uint(1)).Return(nil)

	err := service.DeleteItemWithItsAttachments(1)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockS3.AssertExpectations(t)
}

func TestAddAttachment(t *testing.T) {
	mockRepo := new(MockTodoRepository)
	mockS3 := new(MockS3Client)
	service := newTestTodoService(mockRepo, mockS3, "test-bucket")

	attachment := &models.Attachment{ID: 20, AttachmentURL: "https://dummy-url"}
	mockRepo.On("CreateAttachment", attachment).Return(nil)

	err := service.AddAttachment(attachment)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGeneratePresignedS3Url(t *testing.T) {
	mockRepo := new(MockTodoRepository)
	mockS3 := new(MockS3Client)
	bucket := "test-bucket"
	service := newTestTodoService(mockRepo, mockS3, bucket)

	// For PUT request, fileNameOrURL is a new file name.
	keyPut := "100/newfile.txt"
	fakePutReq := request.New(
		aws.Config{},
		metadata.ClientInfo{ServiceName: "s3", SigningName: "s3"},
		request.Handlers{},
		NoOpRetryer{},
		&request.Operation{
			Name:       "PutObject",
			HTTPMethod: "PUT",
			HTTPPath:   "/",
		},
		nil,
		nil,
	)
	fakePutReq.Handlers.Sign.Clear()
	fakePutReq.Handlers.Build.PushBack(func(r *request.Request) {
		r.HTTPRequest.URL, _ = url.Parse("https://dummy-s3-url/put/" + keyPut)
	})
	mockS3.On("PutObjectRequest", mock.MatchedBy(func(input *s3.PutObjectInput) bool {
		return input.Bucket != nil && *input.Bucket == "test-bucket" &&
			input.Key != nil && *input.Key == keyPut
	})).Return(fakePutReq, &s3.PutObjectOutput{})

	putURL, err := service.GeneratePresignedS3Url(100, "newfile.txt", "PUT")
	assert.NoError(t, err)
	assert.Contains(t, putURL, "https://dummy-s3-url/put/")
	assert.Contains(t, putURL, strconv.Itoa(100))

	// For GET request, fileNameOrURL is an existing S3 URL.
	existingURL := "https://s3.amazonaws.com/test-bucket/100/existing.txt"
	// When parsed, the key will be "test-bucket/100/existing.txt"
	keyGet := "test-bucket/100/existing.txt"
	fakeGetReq := request.New(
		aws.Config{},
		metadata.ClientInfo{ServiceName: "s3", SigningName: "s3"},
		request.Handlers{},
		NoOpRetryer{},
		&request.Operation{
			Name:       "GetObject",
			HTTPMethod: "GET",
			HTTPPath:   "/",
		},
		nil,
		nil,
	)
	fakeGetReq.Handlers.Sign.Clear()
	fakeGetReq.Handlers.Build.PushBack(func(r *request.Request) {
		r.HTTPRequest.URL, _ = url.Parse("https://dummy-s3-url/get/" + keyGet)
	})
	mockS3.On("GetObjectRequest", mock.MatchedBy(func(input *s3.GetObjectInput) bool {
		return input.Bucket != nil && *input.Bucket == "test-bucket" &&
			input.Key != nil && *input.Key == keyGet
	})).Return(fakeGetReq, &s3.GetObjectOutput{})

	getURL, err := service.GeneratePresignedS3Url(100, existingURL, "GET")
	assert.NoError(t, err)
	assert.Contains(t, getURL, "https://dummy-s3-url/get/")
	parsed, _ := url.Parse(existingURL)
	expectedKey := strings.TrimPrefix(parsed.Path, "/")
	assert.Contains(t, getURL, expectedKey)

	// Test invalid request type.
	_, err = service.GeneratePresignedS3Url(100, "file.txt", "DELETE")
	assert.Error(t, err)
}
