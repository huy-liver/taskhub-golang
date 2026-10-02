package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"taskhub/internal/cache"
	"taskhub/internal/errors"
	"taskhub/internal/utils"

	"github.com/gin-gonic/gin"
)

// RateLimiter returns a Gin middleware that enforces rate limits per IP using Redis
func RateLimiter(redisClient *cache.RedisClient, requestsPerMinute int, prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// If Redis is not available, allow traffic through safely
		if redisClient == nil || redisClient.Client() == nil {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		cacheKey := fmt.Sprintf("ratelimit:%s:%s", prefix, clientIP)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 1*time.Second)
		defer cancel()

		rdb := redisClient.Client()
		count, err := rdb.Incr(ctx, cacheKey).Result()
		if err != nil {
			// On Redis error, fail open to avoid service outage
			c.Next()
			return
		}

		// Set expiry on first request in window
		if count == 1 {
			_ = rdb.Expire(ctx, cacheKey, 60*time.Second).Err()
		}

		if count > int64(requestsPerMinute) {
			ttl, _ := rdb.TTL(ctx, cacheKey).Result()
			retryAfter := int(ttl.Seconds())
			if retryAfter <= 0 {
				retryAfter = 60
			}

			c.Header("Retry-After", strconv.Itoa(retryAfter))
			appErr := &errors.AppError{
				StatusCode: http.StatusTooManyRequests,
				Code:       "RATE_LIMIT_EXCEEDED",
				Message:    "Rate limit exceeded. Please try again later.",
				Details: gin.H{
					"retry_after_seconds": retryAfter,
					"limit_per_minute":    requestsPerMinute,
				},
			}

			c.AbortWithStatusJSON(http.StatusTooManyRequests, utils.StandardResponse{
				Success: false,
				Message: appErr.Message,
				Error:   appErr,
			})
			return
		}

		c.Next()
	}
}
