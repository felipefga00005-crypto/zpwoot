package logger

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ContextKey represents keys used in context for logger fields
type ContextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey ContextKey = "request_id"
	// SessionIDKey is the context key for session ID
	SessionIDKey ContextKey = "session_id"
	// LoggerKey is the context key for logger instance
	LoggerKey ContextKey = "logger"
)

// FiberMiddleware creates a Fiber middleware that injects request context into logs
func FiberMiddleware(logger *Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		
		// Generate request ID if not present
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			c.Set("X-Request-ID", requestID)
		}
		
		// Extract session ID from path if present
		sessionID := c.Params("sessionId")
		if sessionID == "" {
			sessionID = c.Params("session_id")
		}
		
		// Create request-scoped logger
		requestLogger := logger.WithRequest(requestID)
		if sessionID != "" {
			requestLogger = requestLogger.WithSession(sessionID)
		}
		
		// Store logger in context
		c.Locals(string(LoggerKey), requestLogger)
		c.Locals(string(RequestIDKey), requestID)
		if sessionID != "" {
			c.Locals(string(SessionIDKey), sessionID)
		}
		
		// Log request start
		requestLogger.EventDebug("http.request.start").
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("user_agent", c.Get("User-Agent")).
			Str("remote_ip", c.IP()).
			Msg("")
		
		// Process request
		err := c.Next()
		
		// Log request completion
		elapsed := time.Since(start)
		event := requestLogger.Event("http.request.complete").
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", c.Response().StatusCode()).
			Int64("elapsed_ms", elapsed.Milliseconds()).
			Int("response_size", len(c.Response().Body()))
		
		if err != nil {
			event = event.Err(err)
		}
		
		event.Msg("")
		
		return err
	}
}

// FromFiberContext extracts the logger from Fiber context
func FromFiberContext(c *fiber.Ctx) *Logger {
	if logger, ok := c.Locals(string(LoggerKey)).(*Logger); ok {
		return logger
	}
	// Fallback to default logger
	return New()
}

// FromContext extracts logger from standard context
func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(LoggerKey).(*Logger); ok {
		return logger
	}
	// Fallback to default logger
	return New()
}

// WithContext adds logger to standard context
func WithContext(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, logger)
}

// GetRequestID extracts request ID from Fiber context
func GetRequestID(c *fiber.Ctx) string {
	if requestID, ok := c.Locals(string(RequestIDKey)).(string); ok {
		return requestID
	}
	return ""
}

// GetSessionID extracts session ID from Fiber context
func GetSessionID(c *fiber.Ctx) string {
	if sessionID, ok := c.Locals(string(SessionIDKey)).(string); ok {
		return sessionID
	}
	return ""
}
