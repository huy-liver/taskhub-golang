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

// CommentHandler handles HTTP requests for comment resources
type CommentHandler struct {
	commentService services.CommentService
}

// NewCommentHandler creates a new CommentHandler instance
func NewCommentHandler(commentService services.CommentService) *CommentHandler {
	return &CommentHandler{commentService: commentService}
}

// ListByTask handles GET /api/v1/comments?task_id=:task_id or GET /api/v1/tasks/:id/comments
func (h *CommentHandler) ListByTask(c *gin.Context) {
	pagination := utils.ParsePagination(c)

	taskIDStr := c.Param("id")
	if taskIDStr == "" {
		taskIDStr = c.Query("task_id")
	}

	if taskIDStr == "" {
		utils.ErrorResponse(c, errors.NewBadRequestError("task_id is required", nil))
		return
	}

	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, errors.NewBadRequestError("Invalid task_id", nil))
		return
	}

	comments, total, appErr := h.commentService.ListByTask(c.Request.Context(), uint(taskID), pagination)
	if appErr != nil {
		utils.ErrorResponse(c, appErr)
		return
	}

	meta := pagination.BuildMeta(total)
	utils.SuccessResponse(c, http.StatusOK, "Comments retrieved successfully", comments, meta)
}

// GetByID handles GET /api/v1/comments/:id
func (h *CommentHandler) GetByID(c *gin.Context) {
	id, err := utils.ParseUintParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, errors.NewBadRequestError("Invalid comment ID", nil))
		return
	}

	comment, appErr := h.commentService.GetByID(c.Request.Context(), id)
	if appErr != nil {
		utils.ErrorResponse(c, appErr)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Comment retrieved successfully", comment, nil)
}

// Create handles POST /api/v1/comments
func (h *CommentHandler) Create(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, errors.NewUnauthorizedError("Authentication required"))
		return
	}

	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, utils.FormatValidationError(err))
		return
	}

	// Also support task_id from URL param if available
	if taskIDStr := c.Param("id"); taskIDStr != "" {
		if tid, err := strconv.ParseUint(taskIDStr, 10, 32); err == nil {
			req.TaskID = uint(tid)
		}
	}

	comment, appErr := h.commentService.Create(c.Request.Context(), userID, &req)
	if appErr != nil {
		utils.ErrorResponse(c, appErr)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Comment created successfully", comment, nil)
}

// Update handles PUT /api/v1/comments/:id
func (h *CommentHandler) Update(c *gin.Context) {
	id, err := utils.ParseUintParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, errors.NewBadRequestError("Invalid comment ID", nil))
		return
	}

	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, errors.NewUnauthorizedError("Authentication required"))
		return
	}
	role, _ := utils.GetUserRoleFromContext(c)

	var req dto.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, utils.FormatValidationError(err))
		return
	}

	comment, appErr := h.commentService.Update(c.Request.Context(), id, userID, role, &req)
	if appErr != nil {
		utils.ErrorResponse(c, appErr)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Comment updated successfully", comment, nil)
}

// Delete handles DELETE /api/v1/comments/:id
func (h *CommentHandler) Delete(c *gin.Context) {
	id, err := utils.ParseUintParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, errors.NewBadRequestError("Invalid comment ID", nil))
		return
	}

	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, errors.NewUnauthorizedError("Authentication required"))
		return
	}
	role, _ := utils.GetUserRoleFromContext(c)

	if appErr := h.commentService.Delete(c.Request.Context(), id, userID, role); appErr != nil {
		utils.ErrorResponse(c, appErr)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Comment deleted successfully", nil, nil)
}
