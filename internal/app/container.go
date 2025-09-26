package app

import (
	"database/sql"
	"fmt"

	"zpwoot/internal/domain/chatwoot"
	"zpwoot/internal/domain/session"
	"zpwoot/internal/domain/webhook"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

// Container holds all use cases and their dependencies
type Container struct {
	// Use Cases
	CommonUseCase   CommonUseCase
	SessionUseCase  SessionUseCase
	WebhookUseCase  WebhookUseCase
	ChatwootUseCase ChatwootUseCase
	MessageUseCase  MessageUseCase

	// Dependencies
	logger      *logger.Logger
	sessionRepo ports.SessionRepository
}

// ContainerConfig holds configuration for creating the container
type ContainerConfig struct {
	// Repositories
	SessionRepo  ports.SessionRepository
	WebhookRepo  ports.WebhookRepository
	ChatwootRepo ports.ChatwootRepository

	// External integrations
	WameowManager       ports.WameowManager
	ChatwootIntegration ports.ChatwootIntegration

	// Infrastructure
	Logger *logger.Logger
	DB     *sql.DB

	// Application metadata
	Version   string
	BuildTime string
	GitCommit string
}

// NewContainer creates a new application container with all use cases
func NewContainer(config *ContainerConfig) *Container {
	// Create domain services
	sessionService := session.NewService(
		config.SessionRepo,
		config.WameowManager,
	)

	webhookService := webhook.NewService(
		config.Logger,
	)

	chatwootService := chatwoot.NewService(
		config.Logger,
	)

	// Create use cases
	commonUseCase := NewCommonUseCase(
		config.Version,
		config.BuildTime,
		config.GitCommit,
		config.DB,
		config.SessionRepo,
		config.WebhookRepo,
	)

	sessionUseCase := NewSessionUseCase(
		config.SessionRepo,
		config.WameowManager,
		sessionService,
	)

	webhookUseCase := NewWebhookUseCase(
		config.WebhookRepo,
		webhookService,
	)

	chatwootUseCase := NewChatwootUseCase(
		config.ChatwootRepo,
		config.ChatwootIntegration,
		chatwootService,
	)

	messageUseCase := NewMessageUseCase(
		config.SessionRepo,
		config.WameowManager,
		config.Logger,
	)

	return &Container{
		CommonUseCase:   commonUseCase,
		SessionUseCase:  sessionUseCase,
		WebhookUseCase:  webhookUseCase,
		ChatwootUseCase: chatwootUseCase,
		MessageUseCase:  messageUseCase,
		logger:          config.Logger,
		sessionRepo:     config.SessionRepo,
	}
}

// GetCommonUseCase returns the common use case
func (c *Container) GetCommonUseCase() CommonUseCase {
	return c.CommonUseCase
}

// GetSessionUseCase returns the session use case
func (c *Container) GetSessionUseCase() SessionUseCase {
	return c.SessionUseCase
}

// GetWebhookUseCase returns the webhook use case
func (c *Container) GetWebhookUseCase() WebhookUseCase {
	return c.WebhookUseCase
}

// GetChatwootUseCase returns the chatwoot use case
func (c *Container) GetChatwootUseCase() ChatwootUseCase {
	return c.ChatwootUseCase
}

// GetLogger returns the logger instance
func (c *Container) GetLogger() *logger.Logger {
	return c.logger
}

// GetSessionRepository returns the session repository instance
func (c *Container) GetSessionRepository() ports.SessionRepository {
	return c.sessionRepo
}

// GetMessageUseCase returns the message use case
func (c *Container) GetMessageUseCase() MessageUseCase {
	return c.MessageUseCase
}

// GetSessionResolver returns a session resolver function
func (c *Container) GetSessionResolver() func(sessionID string) (ports.WameowManager, error) {
	return func(sessionID string) (ports.WameowManager, error) {
		// This should return the WameowManager, not the session repository
		// We need to get the WameowManager from the container config
		// For now, this is a placeholder that will need proper implementation
		return nil, fmt.Errorf("session resolver not properly implemented")
	}
}
