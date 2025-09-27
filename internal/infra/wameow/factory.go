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

type WameowLogger struct {
	logger *logger.Logger
}

func NewWameowLogger(logger *logger.Logger) *WameowLogger {
	return &WameowLogger{
		logger: logger,
	}
}

func (w *WameowLogger) Errorf(msg string, args ...interface{}) {
	w.logger.Errorf(msg, args...)
}

func (w *WameowLogger) Warnf(msg string, args ...interface{}) {
	w.logger.Warnf(msg, args...)
}

func (w *WameowLogger) Infof(msg string, args ...interface{}) {
	w.logger.Infof(msg, args...)
}

func (w *WameowLogger) Debugf(msg string, args ...interface{}) {
	w.logger.Debugf(msg, args...)
}

func (w *WameowLogger) Sub(module string) waLog.Logger {
	return &WameowLogger{
		logger: w.logger, // Use the same underlying logger
	}
}

type Factory struct {
	logger      *logger.Logger
	sessionRepo ports.SessionRepository
}

func NewFactory(logger *logger.Logger, sessionRepo ports.SessionRepository) *Factory {
	return &Factory{
		logger:      logger,
		sessionRepo: sessionRepo,
	}
}

func (f *Factory) CreateManager(db *sql.DB) (*Manager, error) {
	container, err := f.createSQLStoreContainer(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL store container: %w", err)
	}

	manager := NewManager(container, f.sessionRepo, f.logger)

	f.logger.Info("Wameow manager created successfully")
	return manager, nil
}

func (f *Factory) createSQLStoreContainer(db *sql.DB) (*sqlstore.Container, error) {
	waLogger := NewWameowLogger(f.logger)

	container := sqlstore.NewWithDB(db, "postgres", waLogger)
	if container == nil {
		return nil, fmt.Errorf("failed to create SQL store container")
	}

	ctx := context.Background()
	err := container.Upgrade(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade database schema: %w", err)
	}

	f.logger.Info("SQL store container created and upgraded successfully")
	return container, nil
}

type ManagerBuilder struct {
	logger      *logger.Logger
	sessionRepo ports.SessionRepository
	db          *sql.DB
}

func NewManagerBuilder() *ManagerBuilder {
	return &ManagerBuilder{}
}

func (b *ManagerBuilder) WithLogger(logger *logger.Logger) *ManagerBuilder {
	b.logger = logger
	return b
}

func (b *ManagerBuilder) WithSessionRepository(repo ports.SessionRepository) *ManagerBuilder {
	b.sessionRepo = repo
	return b
}

func (b *ManagerBuilder) WithDatabase(db *sql.DB) *ManagerBuilder {
	b.db = db
	return b
}

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

func (m *Manager) GetStats() map[string]interface{} {
	return m.HealthCheck()
}

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
