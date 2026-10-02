package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const HeaderXRequestID = "X-Request-ID"

// RequestID attaches a unique request identifier to each incoming request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(HeaderXRequestID)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set(HeaderXRequestID, requestID)
		c.Writer.Header().Set(HeaderXRequestID, requestID)
		c.Next()
	}
}
