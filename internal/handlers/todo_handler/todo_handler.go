package todo_handler

import (
	"net/http"
	"strconv"

	"github.com/deepjyotk/todo-with-golang/internal/dto"
	"github.com/deepjyotk/todo-with-golang/internal/mapper"
	"github.com/deepjyotk/todo-with-golang/internal/services"
	"github.com/deepjyotk/todo-with-golang/internal/validators"
	"github.com/gin-gonic/gin"
)

// _TodoHandler handles HTTP requests related to todo items.
type TodoHandler struct {
	todoService services.TodoService
	Validator   validators.TodoValidatorInterface
	authService services.AuthService
}

// NewTodoHandler creates a new TodoHandler.
func NewTodoHandler(todoService services.TodoService, validator validators.TodoValidatorInterface, authService services.AuthService) *TodoHandler {
	return &TodoHandler{todoService: todoService, Validator: validator, authService: authService}
}

// CreateTodo godoc
// @Summary Create a new Todo item
// @Description Create a new Todo item with optional attachments.
// @Tags todo
// @Accept json
// @Produce json
// @Param todo body dto.CreateTodoRequest true "Todo item"
// @Success 201 {object} models.Todo
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/todos [post]
func (h *TodoHandler) CreateTodo(c *gin.Context) {
	var req dto.CreateTodoRequest

	// Bind incoming JSON to the CreateTodoRequest DTO.
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Retrieve the user ID from the context.
	userID, err := h.authService.GetUserIDFromContext(c)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Validate the incoming request.
	if err := h.Validator.ValidateCreateTodoRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Map the DTO to the Todo model.
	todo := mapper.MapCreateTodoRequestToTodo(req, userID)

	// Call the service layer to create the Todo item.
	if err := h.todoService.CreateTodo(&todo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create todo"})
		return
	}
	c.JSON(http.StatusCreated, todo)
}

// GeneratePresignedS3UrlPutRequest godoc
// @Summary Generate S3 presigned URL for file uploads
// @Description Generate a presigned S3 URL unique to the authenticated user. The URL will allow the front-end to upload a file (attachment) directly to S3. The file will be stored in a folder named after the user ID.
// @Tags todo
// @Accept json
// @Produce json
// @Param filename query string true "Filename for the attachment"
// @Success 200 {object} map[string]string "presigned URL"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/todos/presigned-url [get]
func (h *TodoHandler) GeneratePresignedS3UrlPutRequest(c *gin.Context) {
	// Retrieve the authenticated user ID.
	userID, err := h.authService.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	//Valid INPUT
	fileName := c.Query("filename")
	if fileName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename is required"})
		return
	}

	if err := h.Validator.ValidFileName(fileName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate the presigned URL using the service.
	url, err := h.todoService.GeneratePresignedS3Url(userID, fileName, "PUT")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate presigned URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

// GetTodo godoc
// @Summary Retrieve a Todo item by ID
// @Description Retrieve a Todo item using its ID. But, the current userID should be accessible to view this TodoID(intersection-AND)
// @Tags todo
// @Accept json
// @Produce json
// @Param id path int true "Todo ID"
// @Success 200 {object} models.Todo
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/todos/{id} [get]
func (h *TodoHandler) GetSpecificTodo(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid todo id"})
		return
	}

	// Retrieve the authenticated user ID.
	userID, err := h.authService.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	todo, err := h.todoService.GetTodoByIDAndUserID(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "todo not found"})
		return
	}
	c.JSON(http.StatusOK, todo)
}

// UpdateTodo godoc
// @Summary Update a Todo item
// @Description Update an existing Todo item by ID.
// @Tags todo
// @Accept json
// @Produce json
// @Param todo body dto.UpdateTodoRequest true "Updated Todo item"
// @Success 200 {object} models.Todo
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/todos/{id} [put]
func (h *TodoHandler) UpdateTodo(c *gin.Context) {
	var req dto.UpdateTodoRequest

	// Bind incoming JSON to the CreateTodoRequest DTO.
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Retrieve the user ID from the context.
	userID, err := h.authService.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	// Validate the incoming request.
	if err := h.Validator.ValidateUpdateTodoRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = h.todoService.GetTodoByIDAndUserID(uint(int(req.ID)), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "todo not found"})
		return
	}

	// Map the DTO to the Todo model.
	todo := mapper.MapUpdateTodoRequestToTodo(req, userID)

	if err := h.todoService.UpdateTodo(&todo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update todo"})
		return
	}
	c.JSON(http.StatusOK, todo)
}

// DeleteTodo godoc
// @Summary Delete a Todo item
// @Description Delete a Todo item by ID.
// @Tags todo
// @Accept json
// @Produce json
// @Param id path int true "Todo ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/todos/{id} [delete]
func (h *TodoHandler) DeleteTodo(c *gin.Context) {
	userID, err := h.authService.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid todo id"})
		return
	}

	todo, err := h.todoService.GetTodoByIDAndUserID(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "todo not found"})
		return
	}
	if err := h.todoService.DeleteItemWithItsAttachments(todo.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete todo"})
		return
	}
	c.Status(http.StatusNoContent)
}

// GetTodos godoc
// @Summary Retrieve all Todos for a user
// @Description Retrieve all Todo items associated with a user ID.
// @Tags todo
// @Accept json
// @Produce json
// @Success 200 {array} models.Todo
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/todos/get-all [get]
func (h *TodoHandler) GetAllTodosForUser(c *gin.Context) {
	userID, err := h.authService.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	todos, err := h.todoService.GetTodosByUser(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve todos"})
		return
	}
	c.JSON(http.StatusOK, todos)
}

/*
TODO: The below 2 functions are the future scope of the project, not gonna implement it now.

// AddAttachment godoc
// @Summary Add an Attachment to a Todo item
// @Description Add an attachment (e.g. S3 URL) to a Todo item.
// @Tags todo
// @Accept json
// @Produce json
// @Param id path int true "Todo ID"
// @Param attachment body models.Attachment true "Attachment data"
// @Success 201 {object} models.Attachment
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/todos/{id}/attachments [post]
func (h *TodoHandler) AddAttachment(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid todo id"})
		return
	}

	var attachment models.Attachment
	if err := c.ShouldBindJSON(&attachment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	attachment.ItemID = uint(id)
	if err := h.todoService.AddAttachment(&attachment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add attachment"})
		return
	}
	c.JSON(http.StatusCreated, attachment)
}
*/
