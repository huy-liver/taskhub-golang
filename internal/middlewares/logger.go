package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ZapLogger logs HTTP requests using Uber's Zap logger
func ZapLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()
		requestID, _ := c.Get(HeaderXRequestID)

		if rawQuery != "" {
			path = path + "?" + rawQuery
		}

		fields := []zap.Field{
			zap.Int("status", statusCode),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("ip", clientIP),
			zap.Duration("latency", latency),
			zap.Int("bytes", bodySize),
			zap.Any("request_id", requestID),
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
			logger.Error("HTTP Request Error", fields...)
		} else if statusCode >= 500 {
			logger.Error("Server Error", fields...)
		} else if statusCode >= 400 {
			logger.Warn("Client Error", fields...)
		} else {
			logger.Info("HTTP Request", fields...)
		}
	}
}
