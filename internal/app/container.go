package app

import (
	"database/sql"
	"fmt"

	"zpwoot/internal/app/chatwoot"
	"zpwoot/internal/app/common"
	"zpwoot/internal/app/message"
	"zpwoot/internal/app/session"
	"zpwoot/internal/app/webhook"
	domainChatwoot "zpwoot/internal/domain/chatwoot"
	domainSession "zpwoot/internal/domain/session"
	domainWebhook "zpwoot/internal/domain/webhook"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

// Container holds all use cases and their dependencies
type Container struct {
	// Use Cases
	CommonUseCase   common.UseCase
	SessionUseCase  session.UseCase
	WebhookUseCase  webhook.UseCase
	ChatwootUseCase chatwoot.UseCase
	MessageUseCase  message.UseCase

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
	sessionService := domainSession.NewService(
		config.SessionRepo,
		config.WameowManager,
	)

	webhookService := domainWebhook.NewService(
		config.Logger,
	)

	chatwootService := domainChatwoot.NewService(
		config.Logger,
	)

	// Create use cases
	commonUseCase := common.NewUseCase(
		config.Version,
		config.BuildTime,
		config.GitCommit,
		config.DB,
		config.SessionRepo,
		config.WebhookRepo,
	)

	sessionUseCase := session.NewUseCase(
		config.SessionRepo,
		config.WameowManager,
		sessionService,
	)

	webhookUseCase := webhook.NewUseCase(
		config.WebhookRepo,
		webhookService,
	)

	chatwootUseCase := chatwoot.NewUseCase(
		config.ChatwootRepo,
		config.ChatwootIntegration,
		chatwootService,
	)

	messageUseCase := message.NewUseCase(
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
func (c *Container) GetCommonUseCase() common.UseCase {
	return c.CommonUseCase
}

// GetSessionUseCase returns the session use case
func (c *Container) GetSessionUseCase() session.UseCase {
	return c.SessionUseCase
}

// GetWebhookUseCase returns the webhook use case
func (c *Container) GetWebhookUseCase() webhook.UseCase {
	return c.WebhookUseCase
}

// GetChatwootUseCase returns the chatwoot use case
func (c *Container) GetChatwootUseCase() chatwoot.UseCase {
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
func (c *Container) GetMessageUseCase() message.UseCase {
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
