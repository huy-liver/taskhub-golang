package middlewares

import (
	"strings"
	"taskhub/internal/errors"
	"taskhub/internal/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the JWT Bearer token in Authorization header
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(c, errors.NewUnauthorizedError("Authorization header is required"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && strings.EqualFold(parts[0], "Bearer")) {
			utils.ErrorResponse(c, errors.NewUnauthorizedError("Authorization header format must be Bearer {token}"))
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(parts[1], jwtSecret)
		if err != nil {
			utils.ErrorResponse(c, errors.NewUnauthorizedError("Invalid or expired token"))
			c.Abort()
			return
		}

		// Store user id and role in Gin context
		c.Set(utils.ContextUserIDKey, claims.UserID)
		c.Set(utils.ContextRoleKey, claims.Role)

		c.Next()
	}
}
