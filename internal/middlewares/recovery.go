package middlewares

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"taskhub/internal/errors"
	"taskhub/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RecoveryMiddleware recovers from panics and returns standard 500 JSON response
func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())
				logger.Error("Panic recovered",
					zap.Any("panic", r),
					zap.String("stack", stack),
				)

				appErr := errors.NewInternalServerError(
					"Internal server error occurred",
					fmt.Errorf("%v", r),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, utils.StandardResponse{
					Success: false,
					Message: appErr.Message,
					Error:   appErr,
				})
			}
		}()
		c.Next()
	}
}
