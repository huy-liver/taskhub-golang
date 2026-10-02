package handlers

import (
	"net/http"
	"taskhub/internal/dto"
	"taskhub/internal/errors"
	"taskhub/internal/services"
	"taskhub/internal/utils"

	"github.com/gin-gonic/gin"
)

// UserHandler handles HTTP requests for user resources
type UserHandler struct {
	userService services.UserService
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// List handles GET /api/v1/users
func (h *UserHandler) List(c *gin.Context) {
	pagination := utils.ParsePagination(c)

	users, total, appErr := h.userService.List(c.Request.Context(), pagination)
	if appErr != nil {
		utils.ErrorResponse(c, appErr)
		return
	}

	meta := pagination.BuildMeta(total)
	utils.SuccessResponse(c, http.StatusOK, "Users retrieved successfully", users, meta)
}

// GetByID handles GET /api/v1/users/:id
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := utils.ParseUintParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, errors.NewBadRequestError("Invalid user ID", nil))
		return
	}

	user, appErr := h.userService.GetByID(c.Request.Context(), id)
	if appErr != nil {
		utils.ErrorResponse(c, appErr)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User retrieved successfully", user, nil)
}

// Create handles POST /api/v1/users
func (h *UserHandler) Create(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, utils.FormatValidationError(err))
		return
	}

	user, appErr := h.userService.Create(c.Request.Context(), &req)
	if appErr != nil {
		utils.ErrorResponse(c, appErr)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "User created successfully", user, nil)
}

// Update handles PUT /api/v1/users/:id
func (h *UserHandler) Update(c *gin.Context) {
	id, err := utils.ParseUintParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, errors.NewBadRequestError("Invalid user ID", nil))
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, utils.FormatValidationError(err))
		return
	}

	user, appErr := h.userService.Update(c.Request.Context(), id, &req)
	if appErr != nil {
		utils.ErrorResponse(c, appErr)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User updated successfully", user, nil)
}

// Delete handles DELETE /api/v1/users/:id
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := utils.ParseUintParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, errors.NewBadRequestError("Invalid user ID", nil))
		return
	}

	if appErr := h.userService.Delete(c.Request.Context(), id); appErr != nil {
		utils.ErrorResponse(c, appErr)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User deleted successfully", nil, nil)
}
