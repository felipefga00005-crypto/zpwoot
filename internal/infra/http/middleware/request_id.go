package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"

	"zpwoot/platform/logger"
)

// RequestID adds a request ID to each request for tracing
func RequestID(logger *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get or generate request ID
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			// Generate new request ID using our UUID generator
			requestID = generateRequestID()
			c.Set("X-Request-ID", requestID)
		}

		// Store in locals for access in handlers
		c.Locals("request_id", requestID)

		// Add to logger context
		requestLogger := logger.WithField("request_id", requestID)
		c.Locals("logger", requestLogger)

		return c.Next()
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	// Simple request ID generation - you can use UUID package here
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// GetLoggerFromContext retrieves the logger from Fiber context
func GetLoggerFromContext(c *fiber.Ctx) *logger.Logger {
	if logger, ok := c.Locals("logger").(*logger.Logger); ok {
		return logger
	}
	// Fallback to default logger
	return logger.New("info")
}

// LogError logs an error with request context
func LogError(c *fiber.Ctx, err error, message string) {
	requestLogger := GetLoggerFromContext(c)

	fields := map[string]interface{}{
		"component": "http",
		"method":    c.Method(),
		"path":      c.Path(),
		"ip":        c.IP(),
	}

	if requestID := c.Locals("request_id"); requestID != nil {
		fields["request_id"] = requestID
	}

	requestLogger.ErrorWithFields(message, fields)
}

// LogInfo logs an info message with request context
func LogInfo(c *fiber.Ctx, message string, additionalFields ...map[string]interface{}) {
	requestLogger := GetLoggerFromContext(c)

	fields := map[string]interface{}{
		"component": "http",
		"method":    c.Method(),
		"path":      c.Path(),
	}

	if requestID := c.Locals("request_id"); requestID != nil {
		fields["request_id"] = requestID
	}

	// Merge additional fields
	for _, additional := range additionalFields {
		for k, v := range additional {
			fields[k] = v
		}
	}

	requestLogger.InfoWithFields(message, fields)
}
