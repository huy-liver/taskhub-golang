package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserIDKey = "userID"
	ContextRoleKey   = "userRole"
)

// GetUserIDFromContext retrieves the authenticated user ID from Gin context
func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	val, exists := c.Get(ContextUserIDKey)
	if !exists {
		return 0, false
	}
	id, ok := val.(uint)
	return id, ok
}

// GetUserRoleFromContext retrieves the authenticated user role from Gin context
func GetUserRoleFromContext(c *gin.Context) (string, bool) {
	val, exists := c.Get(ContextRoleKey)
	if !exists {
		return "", false
	}
	role, ok := val.(string)
	return role, ok
}

// ParseUintParam parses a uint route parameter
func ParseUintParam(c *gin.Context, paramName string) (uint, error) {
	idStr := c.Param(paramName)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
