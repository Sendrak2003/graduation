package middleware

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ErrorResponse struct {
	Timestamp string `json:"timestamp"`
	Status    int    `json:"status"`
	Error     string `json:"error"`
	RequestID string `json:"request_id"`
	Path      string `json:"path"`
}

func NewErrorResponse(requestID, path string) *ErrorResponse {
	return &ErrorResponse{
		Timestamp: time.Now().Format(time.RFC3339),
		Status:    http.StatusInternalServerError,
		Error:     "Внутренняя ошибка сервера",
		RequestID: requestID,
		Path:      path,
	}
}

func NewPanicRecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				requestID := c.GetHeader("X-Request-ID")
				if requestID == "" {
					requestID = uuid.New().String()
				}

				if logger != nil {
					logger.Error("panic recovered",
						zap.String("request_id", requestID),
						zap.String("method", c.Request.Method),
						zap.String("path", c.Request.URL.Path),
						zap.String("user_agent", c.GetHeader("User-Agent")),
						zap.String("remote_addr", c.ClientIP()),
						zap.Any("panic_value", recovered),
						zap.String("stack_trace", string(debug.Stack())),
					)
				}

				errorResponse := NewErrorResponse(requestID, c.Request.URL.Path)
				c.JSON(http.StatusInternalServerError, errorResponse)
				c.Abort()
			}
		}()

		if requestID := c.GetHeader("X-Request-ID"); requestID == "" {
			requestID = uuid.New().String()
			c.Header("X-Request-ID", requestID)
		}

		c.Next()
	}
}
