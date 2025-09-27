package middleware

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"

	"zpwoot/platform/logger"
)

// LoggerConfig represents configuration for HTTP logger middleware
type LoggerConfig struct {
	Logger *logger.Logger
	Format string
	Output io.Writer
}

// NewLogger creates a new HTTP logger middleware using our custom logger
func NewLogger(customLogger *logger.Logger) fiber.Handler {
	return NewLoggerWithConfig(LoggerConfig{
		Logger: customLogger,
	})
}

// NewLoggerWithConfig creates an HTTP logger middleware with custom configuration
func NewLoggerWithConfig(config LoggerConfig) fiber.Handler {
	// Set default values
	if config.Logger == nil {
		config.Logger = logger.New()
	}

	// Custom format for structured logging
	if config.Format == "" {
		config.Format = "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${error}\n"
	}

	// Use custom writer that integrates with our logger
	if config.Output == nil {
		config.Output = &httpLogWriter{logger: config.Logger}
	}

	return fiberLogger.New(fiberLogger.Config{
		Format:     config.Format,
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Local",
		Output:     config.Output,
		CustomTags: map[string]fiberLogger.LogFunc{
			"custom_log": func(output fiberLogger.Buffer, c *fiber.Ctx, data *fiberLogger.Data, extraParam string) (int, error) {
				// Log using our structured logger
				logHTTPRequest(config.Logger, c, data)
				return 0, nil
			},
		},
	})
}

// httpLogWriter implements io.Writer to integrate HTTP logs with our logger
type httpLogWriter struct {
	logger *logger.Logger
}

// Write implements io.Writer interface
func (w *httpLogWriter) Write(p []byte) (int, error) {
	logLine := strings.TrimSpace(string(p))
	if logLine == "" {
		return len(p), nil
	}

	// Parse the log line and extract information
	parts := strings.Split(logLine, " | ")
	if len(parts) >= 6 {
		timestamp := parts[0]
		status := parts[1]
		latency := parts[2]
		ip := parts[3]
		method := parts[4]
		path := parts[5]
		errorMsg := ""
		if len(parts) > 6 {
			errorMsg = parts[6]
		}

		// Convert status to int for level determination
		statusCode, _ := strconv.Atoi(status)

		// Create structured log entry
		fields := map[string]interface{}{
			"component":   "http",
			"timestamp":   timestamp,
			"status_code": statusCode,
			"latency":     latency,
			"ip":          ip,
			"method":      method,
			"path":        path,
		}

		if errorMsg != "" && errorMsg != "-" {
			fields["error"] = errorMsg
		}

		// Log based on status code
		message := fmt.Sprintf("%s %s", method, path)

		switch {
		case statusCode >= 500:
			w.logger.ErrorWithFields(message, fields)
		case statusCode >= 400:
			w.logger.WarnWithFields(message, fields)
		default:
			w.logger.InfoWithFields(message, fields)
		}
	} else {
		// Fallback for unparseable log lines
		w.logger.Info(logLine)
	}

	return len(p), nil
}

// logHTTPRequest logs HTTP request using structured logging
func logHTTPRequest(logger *logger.Logger, c *fiber.Ctx, data *fiberLogger.Data) {
	// Determine log level based on status code
	statusCode := c.Response().StatusCode()

	// Create structured fields
	fields := map[string]interface{}{
		"component":      "http",
		"method":         c.Method(),
		"path":           c.Path(),
		"route":          c.Route().Path,
		"status_code":    statusCode,
		"latency_ms":     data.Stop.Sub(data.Start).Milliseconds(),
		"ip":             c.IP(),
		"user_agent":     c.Get("User-Agent"),
		"content_length": len(c.Response().Body()),
	}

	// Add query parameters if present
	if c.Request().URI().QueryString() != nil {
		fields["query"] = string(c.Request().URI().QueryString())
	}

	// Add request ID if present
	if requestID := c.Get("X-Request-ID"); requestID != "" {
		fields["request_id"] = requestID
	}

	// Add session ID if present in headers or query
	if sessionID := c.Get("X-Session-ID"); sessionID != "" {
		fields["session_id"] = sessionID
	} else if sessionID := c.Query("session_id"); sessionID != "" {
		fields["session_id"] = sessionID
	}

	// Add error information if present
	if err := c.Locals("error"); err != nil {
		fields["error"] = fmt.Sprintf("%v", err)
	}

	// Create log message
	message := fmt.Sprintf("%s %s", c.Method(), c.Path())

	// Log based on status code
	switch {
	case statusCode >= 500:
		logger.ErrorWithFields(message, fields)
	case statusCode >= 400:
		logger.WarnWithFields(message, fields)
	case statusCode >= 300:
		logger.InfoWithFields(message, fields)
	default:
		logger.InfoWithFields(message, fields)
	}
}

// HTTPLogger creates a middleware that logs HTTP requests with structured logging
func HTTPLogger(logger *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate latency
		latency := time.Since(start)
		statusCode := c.Response().StatusCode()

		// Create structured fields
		fields := map[string]interface{}{
			"component":      "http",
			"method":         c.Method(),
			"path":           c.Path(),
			"route":          c.Route().Path,
			"status_code":    statusCode,
			"latency_ms":     latency.Milliseconds(),
			"latency_human":  latency.String(),
			"ip":             c.IP(),
			"user_agent":     c.Get("User-Agent"),
			"content_length": len(c.Response().Body()),
			"protocol":       c.Protocol(),
		}

		// Add query parameters if present
		if queryString := string(c.Request().URI().QueryString()); queryString != "" {
			fields["query"] = queryString
		}

		// Add request headers if needed (be careful with sensitive data)
		if contentType := c.Get("Content-Type"); contentType != "" {
			fields["content_type"] = contentType
		}

		// Add session information if available
		if sessionID := c.Get("X-Session-ID"); sessionID != "" {
			fields["session_id"] = sessionID
		}

		// Add request ID if available
		if requestID := c.Get("X-Request-ID"); requestID != "" {
			fields["request_id"] = requestID
		}

		// Add error information if present
		if err != nil {
			fields["error"] = err.Error()
		}

		// Create log message
		message := fmt.Sprintf("HTTP %s %s", c.Method(), c.Path())

		// Log based on status code and error
		switch {
		case err != nil:
			logger.ErrorWithFields(message, fields)
		case statusCode >= 500:
			logger.ErrorWithFields(message, fields)
		case statusCode >= 400:
			logger.WarnWithFields(message, fields)
		case statusCode >= 300:
			logger.InfoWithFields(message, fields)
		default:
			logger.DebugWithFields(message, fields)
		}

		return err
	}
}
