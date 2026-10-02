package handlers

import (
	"net/http"
	"taskhub/internal/dto"
	"taskhub/internal/errors"
	"taskhub/internal/services"
	"taskhub/internal/utils"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService services.AuthService
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, utils.FormatValidationError(err))
		return
	}

	res, appErr := h.authService.Register(c.Request.Context(), &req)
	if appErr != nil {
		utils.ErrorResponse(c, appErr)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "User registered successfully", res, nil)
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, utils.FormatValidationError(err))
		return
	}

	res, appErr := h.authService.Login(c.Request.Context(), &req)
	if appErr != nil {
		utils.ErrorResponse(c, appErr)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", res, nil)
}

// RefreshToken handles POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, utils.FormatValidationError(err))
		return
	}

	res, appErr := h.authService.RefreshToken(c.Request.Context(), &req)
	if appErr != nil {
		utils.ErrorResponse(c, appErr)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Token refreshed successfully", res, nil)
}

// GetMe handles GET /api/v1/auth/me
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, errors.NewUnauthorizedError("Authentication required"))
		return
	}

	user, appErr := h.authService.GetMe(c.Request.Context(), userID)
	if appErr != nil {
		utils.ErrorResponse(c, appErr)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Current user retrieved successfully", user, nil)
}
