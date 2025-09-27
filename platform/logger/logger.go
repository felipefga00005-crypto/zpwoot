package logger

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// Logger wraps zerolog.Logger with zpwoot-specific functionality
type Logger struct {
	logger zerolog.Logger
	config *LogConfig
}

// New creates a new logger instance based on environment
func New() *Logger {
	env := strings.ToLower(os.Getenv("ZPWOOT_ENV"))
	logLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))

	var config *LogConfig
	switch env {
	case "development", "dev":
		config = DevelopmentConfig()
	case "production", "prod":
		config = ProductionConfig()
	default:
		// Default to development for safety
		config = DevelopmentConfig()
	}

	// Override log level if specified
	if logLevel != "" {
		config.Level = logLevel
	}

	return NewWithConfig(config)
}

// NewWithConfig creates a new logger with custom configuration
func NewWithConfig(config *LogConfig) *Logger {
	// Validate and set defaults
	config.Validate()

	// Set global log level
	logLevel := parseLogLevel(config.Level)
	zerolog.SetGlobalLevel(logLevel)

	// Configure time format
	zerolog.TimeFieldFormat = time.RFC3339

	// Configure output writer
	var writer io.Writer = os.Stdout

	if config.Output == "file" {
		// For now, just use stdout - file rotation can be added later
		writer = os.Stdout
	}

	// Configure format based on environment
	if config.Format == "console" {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        writer,
			TimeFormat: time.RFC3339,
		}

		// Custom caller formatter to show relative path for clickable links
		if config.Caller {
			consoleWriter.FormatCaller = func(i interface{}) string {
				if caller, ok := i.(string); ok {
					// Convert absolute path to relative path from project root
					if strings.Contains(caller, "/workspaces/zpwoot/") {
						relativePath := strings.TrimPrefix(caller, "/workspaces/zpwoot/")
						return relativePath
					}
					// If not in workspace, try to get relative path from current directory
					if strings.Contains(caller, "zpwoot/") {
						parts := strings.Split(caller, "zpwoot/")
						if len(parts) > 1 {
							return parts[len(parts)-1]
						}
					}
					// Fallback to just filename
					return filepath.Base(caller)
				}
				return ""
			}
		}

		writer = consoleWriter
	}

	// Create base logger with global fields
	ctx := zerolog.New(writer).With().
		Timestamp().
		Str("service", "zpwoot")

	// Add environment info
	if env := os.Getenv("ZPWOOT_ENV"); env != "" {
		ctx = ctx.Str("env", env)
	}

	// Add caller info with proper skip level
	if config.Caller {
		ctx = ctx.CallerWithSkipFrameCount(3) // Skip wrapper functions
	}

	logger := ctx.Logger()

	return &Logger{
		logger: logger,
		config: config,
	}
}

// Event-based logging methods following zpwoot patterns

// Event logs a structured event with consistent fields
func (l *Logger) Event(event string) *zerolog.Event {
	return l.logger.Info().Str("event", event)
}

// EventDebug logs a debug-level structured event
func (l *Logger) EventDebug(event string) *zerolog.Event {
	return l.logger.Debug().Str("event", event)
}

// EventError logs an error-level structured event
func (l *Logger) EventError(event string) *zerolog.Event {
	return l.logger.Error().Str("event", event)
}

// EventWarn logs a warning-level structured event
func (l *Logger) EventWarn(event string) *zerolog.Event {
	return l.logger.Warn().Str("event", event)
}

// WithSession returns a logger with session context
func (l *Logger) WithSession(sessionID string) *Logger {
	return &Logger{
		logger: l.logger.With().Str("session_id", sessionID).Logger(),
		config: l.config,
	}
}

// WithRequest returns a logger with request context
func (l *Logger) WithRequest(requestID string) *Logger {
	return &Logger{
		logger: l.logger.With().Str("request_id", requestID).Logger(),
		config: l.config,
	}
}

// WithMessage returns a logger with message context
func (l *Logger) WithMessage(messageID string) *Logger {
	return &Logger{
		logger: l.logger.With().Str("message_id", messageID).Logger(),
		config: l.config,
	}
}

// WithElapsed adds elapsed time in milliseconds
func (l *Logger) WithElapsed(start time.Time) *Logger {
	elapsed := time.Since(start).Milliseconds()
	return &Logger{
		logger: l.logger.With().Int64("elapsed_ms", elapsed).Logger(),
		config: l.config,
	}
}

// parseLogLevel converts string level to zerolog.Level
func parseLogLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info", "":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	case "disabled":
		return zerolog.Disabled
	default:
		return zerolog.InfoLevel
	}
}

// Simple logging methods for backward compatibility

// Info logs an info level message
func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

// Infof logs an info level message with formatting
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

// InfoWithFields logs an info level message with structured fields
func (l *Logger) InfoWithFields(msg string, fields map[string]interface{}) {
	event := l.logger.Info()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Error logs an error level message
func (l *Logger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

// Errorf logs an error level message with formatting
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

// ErrorWithFields logs an error level message with structured fields
func (l *Logger) ErrorWithFields(msg string, fields map[string]interface{}) {
	event := l.logger.Error()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Debug logs a debug level message
func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

// Debugf logs a debug level message with formatting
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

// DebugWithFields logs a debug level message with structured fields
func (l *Logger) DebugWithFields(msg string, fields map[string]interface{}) {
	event := l.logger.Debug()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Warn logs a warning level message
func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

// Warnf logs a warning level message with formatting
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

// WarnWithFields logs a warning level message with structured fields
func (l *Logger) WarnWithFields(msg string, fields map[string]interface{}) {
	event := l.logger.Warn()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Fatal logs a fatal level message and exits
func (l *Logger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}

// WithError returns a logger with an error field
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		logger: l.logger.With().Err(err).Logger(),
		config: l.config,
	}
}

// WithField returns a logger with a single field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		logger: l.logger.With().Interface(key, value).Logger(),
		config: l.config,
	}
}

// GetZerologLogger returns the underlying zerolog.Logger for advanced usage
func (l *Logger) GetZerologLogger() zerolog.Logger {
	return l.logger
}
