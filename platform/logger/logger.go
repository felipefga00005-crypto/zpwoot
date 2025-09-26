package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger wraps zerolog.Logger with additional functionality
type Logger struct {
	logger zerolog.Logger
	level  string
}

// New creates a new logger instance with the specified configuration
func New(level string) *Logger {
	config := &LogConfig{
		Level:      level,
		Format:     "console", // default to console for development
		Output:     "stdout",
		TimeFormat: time.RFC3339,
	}
	return NewWithConfig(config)
}

// NewWithConfig creates a new logger with custom configuration
func NewWithConfig(config *LogConfig) *Logger {
	// Set global log level
	logLevel := parseLogLevel(config.Level)
	zerolog.SetGlobalLevel(logLevel)

	// Configure time format
	if config.TimeFormat != "" {
		zerolog.TimeFieldFormat = config.TimeFormat
	}

	// Configure output writer
	var writer io.Writer
	switch strings.ToLower(config.Output) {
	case "stderr":
		writer = os.Stderr
	case "stdout", "":
		writer = os.Stdout
	case "file":
		writer = createLogFile("logs/app.log")
	default:
		// Treat as file path
		writer = createLogFile(config.Output)
	}

	// Configure format
	if strings.ToLower(config.Format) == "console" {
		writer = zerolog.ConsoleWriter{
			Out:        writer,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}
	}

	// Create logger
	logger := zerolog.New(writer).With().
		Timestamp().
		Caller().
		Logger()

	return &Logger{
		logger: logger,
		level:  config.Level,
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

// createLogFile creates a log file with proper directory structure
func createLogFile(filePath string) io.Writer {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Error().Err(err).Str("path", dir).Msg("Failed to create log directory")
		return os.Stdout
	}

	// Open or create log file
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Error().Err(err).Str("path", filePath).Msg("Failed to create log file")
		return os.Stdout
	}

	return file
}

// Logging methods with various signatures

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

// ErrorWithErr logs an error with an error object
func (l *Logger) ErrorWithErr(err error, msg string) {
	l.logger.Error().Err(err).Msg(msg)
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

// Fatalf logs a fatal level message with formatting and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal().Msgf(format, args...)
}

// Trace logs a trace level message
func (l *Logger) Trace(msg string) {
	l.logger.Trace().Msg(msg)
}

// Tracef logs a trace level message with formatting
func (l *Logger) Tracef(format string, args ...interface{}) {
	l.logger.Trace().Msgf(format, args...)
}

// WithField returns a logger with a single field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		logger: l.logger.With().Interface(key, value).Logger(),
		level:  l.level,
	}
}

// WithFields returns a logger with multiple fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	ctx := l.logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return &Logger{
		logger: ctx.Logger(),
		level:  l.level,
	}
}

// WithError returns a logger with an error field
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		logger: l.logger.With().Err(err).Logger(),
		level:  l.level,
	}
}

// GetLevel returns the current log level
func (l *Logger) GetLevel() string {
	return l.level
}

// SetLevel updates the log level
func (l *Logger) SetLevel(level string) {
	l.level = level
	zerolog.SetGlobalLevel(parseLogLevel(level))
}

// GetZerologLogger returns the underlying zerolog.Logger
func (l *Logger) GetZerologLogger() zerolog.Logger {
	return l.logger
}

// Compatibility methods for existing code that might use variadic args

// InfoArgs logs info with variadic arguments (for compatibility)
func (l *Logger) InfoArgs(args ...interface{}) {
	l.logger.Info().Msg(fmt.Sprint(args...))
}

// ErrorArgs logs error with variadic arguments (for compatibility)
func (l *Logger) ErrorArgs(args ...interface{}) {
	l.logger.Error().Msg(fmt.Sprint(args...))
}

// DebugArgs logs debug with variadic arguments (for compatibility)
func (l *Logger) DebugArgs(args ...interface{}) {
	l.logger.Debug().Msg(fmt.Sprint(args...))
}

// WarnArgs logs warning with variadic arguments (for compatibility)
func (l *Logger) WarnArgs(args ...interface{}) {
	l.logger.Warn().Msg(fmt.Sprint(args...))
}
