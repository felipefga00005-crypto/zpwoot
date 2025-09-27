package app

import (
	"database/sql"
	"fmt"

	"zpwoot/internal/app/chatwoot"
	"zpwoot/internal/app/common"
	"zpwoot/internal/app/group"
	"zpwoot/internal/app/message"
	"zpwoot/internal/app/session"
	"zpwoot/internal/app/webhook"
	domainChatwoot "zpwoot/internal/domain/chatwoot"
	domainGroup "zpwoot/internal/domain/group"
	domainSession "zpwoot/internal/domain/session"
	domainWebhook "zpwoot/internal/domain/webhook"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

type Container struct {
	CommonUseCase   common.UseCase
	SessionUseCase  session.UseCase
	WebhookUseCase  webhook.UseCase
	ChatwootUseCase chatwoot.UseCase
	MessageUseCase  message.UseCase
	GroupUseCase    group.UseCase

	logger      *logger.Logger
	sessionRepo ports.SessionRepository
}

type ContainerConfig struct {
	SessionRepo  ports.SessionRepository
	WebhookRepo  ports.WebhookRepository
	ChatwootRepo ports.ChatwootRepository
	GroupRepo    ports.GroupRepository

	WameowManager       ports.WameowManager
	ChatwootIntegration ports.ChatwootIntegration

	Logger *logger.Logger
	DB     *sql.DB

	Version   string
	BuildTime string
	GitCommit string
}

func NewContainer(config *ContainerConfig) *Container {
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

	groupService := domainGroup.NewService(
		config.GroupRepo,
		config.WameowManager,
	)

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

	groupUseCase := group.NewUseCase(
		config.GroupRepo,
		config.WameowManager,
		groupService,
	)

	return &Container{
		CommonUseCase:   commonUseCase,
		SessionUseCase:  sessionUseCase,
		WebhookUseCase:  webhookUseCase,
		ChatwootUseCase: chatwootUseCase,
		MessageUseCase:  messageUseCase,
		GroupUseCase:    groupUseCase,
		logger:          config.Logger,
		sessionRepo:     config.SessionRepo,
	}
}

func (c *Container) GetCommonUseCase() common.UseCase {
	return c.CommonUseCase
}

func (c *Container) GetSessionUseCase() session.UseCase {
	return c.SessionUseCase
}

func (c *Container) GetWebhookUseCase() webhook.UseCase {
	return c.WebhookUseCase
}

func (c *Container) GetChatwootUseCase() chatwoot.UseCase {
	return c.ChatwootUseCase
}

func (c *Container) GetLogger() *logger.Logger {
	return c.logger
}

func (c *Container) GetSessionRepository() ports.SessionRepository {
	return c.sessionRepo
}

func (c *Container) GetMessageUseCase() message.UseCase {
	return c.MessageUseCase
}

func (c *Container) GetGroupUseCase() group.UseCase {
	return c.GroupUseCase
}

func (c *Container) GetSessionResolver() func(sessionID string) (ports.WameowManager, error) {
	return func(sessionID string) (ports.WameowManager, error) {
		return nil, fmt.Errorf("session resolver not properly implemented")
	}
}
