package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			c.Header("X-Request-ID", requestID)
		}

		c.Set("request_id", requestID)

		logger.Info("incoming request",
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()))

		c.Next()

		latency := time.Since(start)
		userID, _ := c.Get("user_id")

		logger.Info("request completed",
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("latency_human", latency.String()),
			zap.Any("user_id", userID),
			zap.Int("response_size", c.Writer.Size()))

		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.Error("request error",
					zap.String("request_id", requestID),
					zap.Error(err.Err))
			}
		}
	}
}
