package utils

import (
	"taskhub/internal/errors"

	"github.com/gin-gonic/gin"
)

// StandardResponse defines the unified API response structure
type StandardResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Data    any              `json:"data,omitempty"`
	Error   *errors.AppError `json:"error,omitempty"`
	Meta    any              `json:"meta,omitempty"`
}

// PaginationMeta provides pagination information
type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// SuccessResponse sends a standard JSON success response
func SuccessResponse(c *gin.Context, statusCode int, message string, data any, meta any) {
	c.JSON(statusCode, StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// ErrorResponse sends a standard JSON error response mapped from AppError
func ErrorResponse(c *gin.Context, appErr *errors.AppError) {
	if appErr == nil {
		appErr = errors.NewInternalServerError("An unknown error occurred", nil)
	}

	c.JSON(appErr.StatusCode, StandardResponse{
		Success: false,
		Message: appErr.Message,
		Error:   appErr,
	})
}
