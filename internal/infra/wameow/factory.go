package wameow

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"zpwoot/internal/ports"
	"zpwoot/platform/logger"

	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// WameowLogger wraps our logger to implement whatsmeow's log interface
type WameowLogger struct {
	logger *logger.Logger
}

// NewWameowLogger creates a new Wameow logger wrapper
func NewWameowLogger(logger *logger.Logger) *WameowLogger {
	return &WameowLogger{
		logger: logger,
	}
}

// Errorf implements waLog.Logger
func (w *WameowLogger) Errorf(msg string, args ...interface{}) {
	w.logger.Errorf(msg, args...)
}

// Warnf implements waLog.Logger
func (w *WameowLogger) Warnf(msg string, args ...interface{}) {
	w.logger.Warnf(msg, args...)
}

// Infof implements waLog.Logger
func (w *WameowLogger) Infof(msg string, args ...interface{}) {
	w.logger.Infof(msg, args...)
}

// Debugf implements waLog.Logger
func (w *WameowLogger) Debugf(msg string, args ...interface{}) {
	w.logger.Debugf(msg, args...)
}

// Sub implements waLog.Logger
func (w *WameowLogger) Sub(module string) waLog.Logger {
	// Create a new logger instance for the sub-module
	return &WameowLogger{
		logger: w.logger, // Use the same underlying logger
	}
}

// Factory creates and configures Wameow manager components
type Factory struct {
	logger      *logger.Logger
	sessionRepo ports.SessionRepository
}

// NewFactory creates a new factory
func NewFactory(logger *logger.Logger, sessionRepo ports.SessionRepository) *Factory {
	return &Factory{
		logger:      logger,
		sessionRepo: sessionRepo,
	}
}

// CreateManager creates a new Wameow manager with all dependencies
func (f *Factory) CreateManager(db *sql.DB) (*Manager, error) {
	// Create SQL store container
	container, err := f.createSQLStoreContainer(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL store container: %w", err)
	}

	// Create manager
	manager := NewManager(container, f.sessionRepo, f.logger)

	f.logger.Info("Wameow manager created successfully")
	return manager, nil
}

// createSQLStoreContainer creates and configures the SQL store container
func (f *Factory) createSQLStoreContainer(db *sql.DB) (*sqlstore.Container, error) {
	// Create Wameow logger
	waLogger := NewWameowLogger(f.logger)

	// Create SQL store container
	container := sqlstore.NewWithDB(db, "postgres", waLogger)
	if container == nil {
		return nil, fmt.Errorf("failed to create SQL store container")
	}

	// Upgrade database schema if needed
	ctx := context.Background()
	err := container.Upgrade(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade database schema: %w", err)
	}

	f.logger.Info("SQL store container created and upgraded successfully")
	return container, nil
}

// ManagerBuilder provides a fluent interface for building managers
type ManagerBuilder struct {
	logger      *logger.Logger
	sessionRepo ports.SessionRepository
	db          *sql.DB
}

// NewManagerBuilder creates a new manager builder
func NewManagerBuilder() *ManagerBuilder {
	return &ManagerBuilder{}
}

// WithLogger sets the logger
func (b *ManagerBuilder) WithLogger(logger *logger.Logger) *ManagerBuilder {
	b.logger = logger
	return b
}

// WithSessionRepository sets the session repository
func (b *ManagerBuilder) WithSessionRepository(repo ports.SessionRepository) *ManagerBuilder {
	b.sessionRepo = repo
	return b
}

// WithDatabase sets the database connection
func (b *ManagerBuilder) WithDatabase(db *sql.DB) *ManagerBuilder {
	b.db = db
	return b
}

// Build creates the manager with the configured options
func (b *ManagerBuilder) Build() (*Manager, error) {
	if b.logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	if b.sessionRepo == nil {
		return nil, fmt.Errorf("session repository is required")
	}

	if b.db == nil {
		return nil, fmt.Errorf("database connection is required")
	}

	factory := NewFactory(b.logger, b.sessionRepo)
	return factory.CreateManager(b.db)
}

// HealthCheck performs a health check on the Wameow manager
func (m *Manager) HealthCheck() map[string]interface{} {
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()

	totalSessions := len(m.clients)
	connectedSessions := 0
	loggedInSessions := 0

	for sessionID, client := range m.clients {
		if client.IsConnected() {
			connectedSessions++
		}
		if client.IsLoggedIn() {
			loggedInSessions++
		}

		// Log session status for debugging
		m.logger.InfoWithFields("Session status", map[string]interface{}{
			"session_id": sessionID,
			"connected":  client.IsConnected(),
			"logged_in":  client.IsLoggedIn(),
		})
	}

	return map[string]interface{}{
		"total_sessions":     totalSessions,
		"connected_sessions": connectedSessions,
		"logged_in_sessions": loggedInSessions,
		"healthy":            true,
		"timestamp":          time.Now().Unix(),
	}
}

// GetStats returns statistics about the Wameow manager
func (m *Manager) GetStats() map[string]interface{} {
	return m.HealthCheck()
}

// LogLevelToWALevel converts string log level to whatsmeow log level
// Note: whatsmeow doesn't use log levels in the same way, this is kept for compatibility
func LogLevelToWALevel(level string) string {
	switch level {
	case "ERROR":
		return "error"
	case "WARN":
		return "warn"
	case "INFO":
		return "info"
	case "DEBUG":
		return "debug"
	default:
		return "info"
	}
}
