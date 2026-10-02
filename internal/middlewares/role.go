package middlewares

import (
	"taskhub/internal/errors"
	"taskhub/internal/utils"

	"github.com/gin-gonic/gin"
)

// RequireRole enforces that the authenticated user possesses one of the required roles
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := utils.GetUserRoleFromContext(c)
		if !exists {
			utils.ErrorResponse(c, errors.NewUnauthorizedError("Authentication required"))
			c.Abort()
			return
		}

		for _, r := range allowedRoles {
			if r == role {
				c.Next()
				return
			}
		}

		utils.ErrorResponse(c, errors.NewForbiddenError("You do not have the required role to access this resource"))
		c.Abort()
	}
}
