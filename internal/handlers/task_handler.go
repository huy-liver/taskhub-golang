package handlers

import (
	"net/http"
	"strconv"
	"taskhub/internal/dto"
	"taskhub/internal/errors"
	"taskhub/internal/services"
	"taskhub/internal/utils"

	"github.com/gin-gonic/gin"
)

// TaskHandler handles HTTP requests for task resources
type TaskHandler struct {
	taskService services.TaskService
}

// NewTaskHandler creates a new TaskHandler instance
func NewTaskHandler(taskService services.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

// List handles GET /api/v1/tasks
func (h *TaskHandler) List(c *gin.Context) {
	pagination := utils.ParsePagination(c)

	var filter dto.TaskFilterQuery
	if projectIDStr := c.Query("project_id"); projectIDStr != "" {
		if id, err := strconv.ParseUint(projectIDStr, 10, 32); err == nil {
			uid := uint(id)
			filter.ProjectID = &uid
		}
	}
	if assigneeIDStr := c.Query("assignee_id"); assigneeIDStr != "" {
		if id, err := strconv.ParseUint(assigneeIDStr, 10, 32); err == nil {
			uid := uint(id)
			filter.AssigneeID = &uid
		}
	}
	filter.Status = c.Query("status")
	filter.Priority = c.Query("priority")

	tasks, total, appErr := h.taskService.List(c.Request.Context(), pagination, filter)
	if appErr != nil {
		utils.ErrorResponse(c, appErr)
		return
	}

	meta := pagination.BuildMeta(total)
	utils.SuccessResponse(c, http.StatusOK, "Tasks retrieved successfully", tasks, meta)
}

// GetByID handles GET /api/v1/tasks/:id
func (h *TaskHandler) GetByID(c *gin.Context) {
	id, err := utils.ParseUintParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, errors.NewBadRequestError("Invalid task ID", nil))
		return
	}

	task, appErr := h.taskService.GetByID(c.Request.Context(), id)
	if appErr != nil {
		utils.ErrorResponse(c, appErr)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Task retrieved successfully", task, nil)
}

// Create handles POST /api/v1/tasks
func (h *TaskHandler) Create(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, errors.NewUnauthorizedError("Authentication required"))
		return
	}
	role, _ := utils.GetUserRoleFromContext(c)

	var req dto.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, utils.FormatValidationError(err))
		return
	}

	task, appErr := h.taskService.Create(c.Request.Context(), userID, role, &req)
	if appErr != nil {
		utils.ErrorResponse(c, appErr)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Task created successfully", task, nil)
}

// Update handles PUT /api/v1/tasks/:id
func (h *TaskHandler) Update(c *gin.Context) {
	id, err := utils.ParseUintParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, errors.NewBadRequestError("Invalid task ID", nil))
		return
	}

	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, errors.NewUnauthorizedError("Authentication required"))
		return
	}
	role, _ := utils.GetUserRoleFromContext(c)

	var req dto.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, utils.FormatValidationError(err))
		return
	}

	task, appErr := h.taskService.Update(c.Request.Context(), id, userID, role, &req)
	if appErr != nil {
		utils.ErrorResponse(c, appErr)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Task updated successfully", task, nil)
}

// PatchStatus handles PATCH /api/v1/tasks/:id/status
func (h *TaskHandler) PatchStatus(c *gin.Context) {
	id, err := utils.ParseUintParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, errors.NewBadRequestError("Invalid task ID", nil))
		return
	}

	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, errors.NewUnauthorizedError("Authentication required"))
		return
	}
	role, _ := utils.GetUserRoleFromContext(c)

	var req dto.PatchTaskStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, utils.FormatValidationError(err))
		return
	}

	task, appErr := h.taskService.PatchStatus(c.Request.Context(), id, userID, role, req.Status)
	if appErr != nil {
		utils.ErrorResponse(c, appErr)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Task status updated successfully", task, nil)
}

// Delete handles DELETE /api/v1/tasks/:id
func (h *TaskHandler) Delete(c *gin.Context) {
	id, err := utils.ParseUintParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, errors.NewBadRequestError("Invalid task ID", nil))
		return
	}

	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, errors.NewUnauthorizedError("Authentication required"))
		return
	}
	role, _ := utils.GetUserRoleFromContext(c)

	if appErr := h.taskService.Delete(c.Request.Context(), id, userID, role); appErr != nil {
		utils.ErrorResponse(c, appErr)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Task deleted successfully", nil, nil)
}
