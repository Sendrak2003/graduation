package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PanicRecoveryMiddleware struct {
	logger Logger
}

type Logger interface {
	Error(msg string, fields ...interface{})
	LogPanic(entry PanicLogEntry)
}

type DefaultLogger struct{}

func (l *DefaultLogger) Error(msg string, fields ...interface{}) {
	log.Printf("ERROR: %s %v", msg, fields)
}

func (l *DefaultLogger) LogPanic(entry PanicLogEntry) {
	log.Printf("PANIC RECOVERED: RequestID=%s Method=%s Path=%s Panic=%s", 
		entry.RequestID, entry.Method, entry.Path, entry.PanicValue)
}

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
		Error:     sanitizeErrorMessage("Внутренняя ошибка сервера"),
		RequestID: requestID,
		Path:      path,
	}
}

func sanitizeErrorMessage(message string) string {
	return "Внутренняя ошибка сервера"
}

type PanicLogEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	RequestID   string    `json:"request_id"`
	Method      string    `json:"method"`
	Path        string    `json:"path"`
	UserAgent   string    `json:"user_agent"`
	RemoteAddr  string    `json:"remote_addr"`
	PanicValue  string    `json:"panic_value"`
	StackTrace  string    `json:"stack_trace"`
}

func NewPanicRecoveryMiddleware(logger Logger) gin.HandlerFunc {
	if logger == nil {
		logger = &DefaultLogger{}
	}
	
	middleware := &PanicRecoveryMiddleware{
		logger: logger,
	}
	
	return middleware.handleRecovery()
}

func (m *PanicRecoveryMiddleware) handleRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				m.handlePanic(c, recovered)
			}
		}()
		
		if requestID := c.GetHeader("X-Request-ID"); requestID == "" {
			requestID = uuid.New().String()
			c.Header("X-Request-ID", requestID)
		}
		
		c.Next()
	}
}

func (m *PanicRecoveryMiddleware) handlePanic(c *gin.Context, recovered interface{}) {
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = uuid.New().String()
	}
	
	panicEntry := PanicLogEntry{
		Timestamp:   time.Now(),
		RequestID:   requestID,
		Method:      c.Request.Method,
		Path:        c.Request.URL.Path,
		UserAgent:   c.GetHeader("User-Agent"),
		RemoteAddr:  c.ClientIP(),
		PanicValue:  fmt.Sprintf("%v", recovered),
		StackTrace:  string(debug.Stack()),
	}
	
	m.logger.LogPanic(panicEntry)
	
	errorResponse := NewErrorResponse(requestID, c.Request.URL.Path)
	
	c.JSON(http.StatusInternalServerError, errorResponse)
	c.Abort()
}