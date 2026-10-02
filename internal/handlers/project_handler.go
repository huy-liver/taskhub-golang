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

// ProjectHandler handles HTTP requests for project resources
type ProjectHandler struct {
	projectService services.ProjectService
}

// NewProjectHandler creates a new ProjectHandler instance
func NewProjectHandler(projectService services.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

// List handles GET /api/v1/projects
func (h *ProjectHandler) List(c *gin.Context) {
	pagination := utils.ParsePagination(c)

	var ownerID *uint
	if ownerParam := c.Query("owner_id"); ownerParam != "" {
		if id, err := strconv.ParseUint(ownerParam, 10, 32); err == nil {
			uid := uint(id)
			ownerID = &uid
		}
	}

	projects, total, appErr := h.projectService.List(c.Request.Context(), pagination, ownerID)
	if appErr != nil {
		utils.ErrorResponse(c, appErr)
		return
	}

	meta := pagination.BuildMeta(total)
	utils.SuccessResponse(c, http.StatusOK, "Projects retrieved successfully", projects, meta)
}

// GetByID handles GET /api/v1/projects/:id
func (h *ProjectHandler) GetByID(c *gin.Context) {
	id, err := utils.ParseUintParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, errors.NewBadRequestError("Invalid project ID", nil))
		return
	}

	project, appErr := h.projectService.GetByID(c.Request.Context(), id)
	if appErr != nil {
		utils.ErrorResponse(c, appErr)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Project retrieved successfully", project, nil)
}

// Create handles POST /api/v1/projects
func (h *ProjectHandler) Create(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, errors.NewUnauthorizedError("Authentication required"))
		return
	}

	var req dto.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, utils.FormatValidationError(err))
		return
	}

	project, appErr := h.projectService.Create(c.Request.Context(), userID, &req)
	if appErr != nil {
		utils.ErrorResponse(c, appErr)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Project created successfully", project, nil)
}

// Update handles PUT /api/v1/projects/:id
func (h *ProjectHandler) Update(c *gin.Context) {
	id, err := utils.ParseUintParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, errors.NewBadRequestError("Invalid project ID", nil))
		return
	}

	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, errors.NewUnauthorizedError("Authentication required"))
		return
	}
	role, _ := utils.GetUserRoleFromContext(c)

	var req dto.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, utils.FormatValidationError(err))
		return
	}

	project, appErr := h.projectService.Update(c.Request.Context(), id, userID, role, &req)
	if appErr != nil {
		utils.ErrorResponse(c, appErr)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Project updated successfully", project, nil)
}

// Delete handles DELETE /api/v1/projects/:id
func (h *ProjectHandler) Delete(c *gin.Context) {
	id, err := utils.ParseUintParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, errors.NewBadRequestError("Invalid project ID", nil))
		return
	}

	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, errors.NewUnauthorizedError("Authentication required"))
		return
	}
	role, _ := utils.GetUserRoleFromContext(c)

	if appErr := h.projectService.Delete(c.Request.Context(), id, userID, role); appErr != nil {
		utils.ErrorResponse(c, appErr)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Project deleted successfully", nil, nil)
}
