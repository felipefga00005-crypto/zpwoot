package app

import (
	"database/sql"
	"fmt"

	"zpwoot/internal/app/chatwoot"
	"zpwoot/internal/app/common"
	"zpwoot/internal/app/community"
	"zpwoot/internal/app/contact"
	"zpwoot/internal/app/group"
	"zpwoot/internal/app/media"
	"zpwoot/internal/app/message"
	"zpwoot/internal/app/newsletter"
	"zpwoot/internal/app/session"
	"zpwoot/internal/app/webhook"
	domainChatwoot "zpwoot/internal/domain/chatwoot"
	domainCommunity "zpwoot/internal/domain/community"
	domainContact "zpwoot/internal/domain/contact"
	domainGroup "zpwoot/internal/domain/group"
	domainMedia "zpwoot/internal/domain/media"
	domainNewsletter "zpwoot/internal/domain/newsletter"
	domainSession "zpwoot/internal/domain/session"
	domainWebhook "zpwoot/internal/domain/webhook"
	chatwootIntegration "zpwoot/internal/infra/integrations/chatwoot"
	"zpwoot/internal/infra/wameow"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

type Container struct {
	CommonUseCase     common.UseCase
	SessionUseCase    session.UseCase
	WebhookUseCase    webhook.UseCase
	ChatwootUseCase   chatwoot.UseCase
	MessageUseCase    message.UseCase
	MediaUseCase      media.UseCase
	GroupUseCase      group.UseCase
	ContactUseCase    contact.UseCase
	NewsletterUseCase newsletter.UseCase
	CommunityUseCase  community.UseCase

	logger      *logger.Logger
	sessionRepo ports.SessionRepository
}

type ContainerConfig struct {
	SessionRepo         ports.SessionRepository
	WebhookRepo         ports.WebhookRepository
	ChatwootRepo        ports.ChatwootRepository
	ChatwootMessageRepo ports.ChatwootMessageRepository
	MediaRepo           ports.MediaRepository

	WameowManager       ports.WameowManager
	ChatwootIntegration ports.ChatwootIntegration

	Logger *logger.Logger
	DB     *sql.DB

	Version   string
	BuildTime string
	GitCommit string
}

func NewContainer(config *ContainerConfig) *Container {
	services := createDomainServices(config)
	useCases := createUseCases(config, services)

	return &Container{
		CommonUseCase:     useCases.common,
		SessionUseCase:    useCases.session,
		WebhookUseCase:    useCases.webhook,
		ChatwootUseCase:   useCases.chatwoot,
		MessageUseCase:    useCases.message,
		MediaUseCase:      useCases.media,
		GroupUseCase:      useCases.group,
		ContactUseCase:    useCases.contact,
		NewsletterUseCase: useCases.newsletter,
		CommunityUseCase:  useCases.community,
		logger:            config.Logger,
		sessionRepo:       config.SessionRepo,
	}
}

// domainServices holds all domain services
type domainServices struct {
	session    *domainSession.Service
	webhook    *domainWebhook.Service
	chatwoot   *domainChatwoot.Service
	group      *domainGroup.Service
	contact    domainContact.Service
	media      domainMedia.Service
	newsletter *domainNewsletter.Service
	community  domainCommunity.Service
}

// useCases holds all use cases
type useCases struct {
	common     common.UseCase
	session    session.UseCase
	webhook    webhook.UseCase
	chatwoot   chatwoot.UseCase
	message    message.UseCase
	media      media.UseCase
	group      group.UseCase
	contact    contact.UseCase
	newsletter newsletter.UseCase
	community  community.UseCase
}

// createDomainServices creates all domain services
func createDomainServices(config *ContainerConfig) *domainServices {
	sessionService := domainSession.NewService(
		config.SessionRepo,
		config.WameowManager,
	)

	webhookService := domainWebhook.NewService(
		config.Logger,
		config.WebhookRepo,
	)

	chatwootService := domainChatwoot.NewService(
		config.Logger,
		config.ChatwootRepo,
		config.WameowManager,
	)

	// Create and inject MessageMapper if available
	if config.ChatwootMessageRepo != nil {
		messageMapper := createMessageMapper(config.Logger, config.ChatwootMessageRepo)
		chatwootService.SetMessageMapper(messageMapper)
	}

	// Create JID validator for group service
	jidValidator := wameow.NewJIDValidatorAdapter()

	groupService := domainGroup.NewService(
		nil, // No repository needed for groups
		config.WameowManager,
		jidValidator,
	)

	contactService := domainContact.NewService(
		config.WameowManager, // WhatsAppClient interface
		config.Logger,
	)

	// Create media service
	// Note: MediaService requires WhatsAppClient and CacheManager which are not available in this context
	// For now, we'll pass nil values and handle this in the actual implementation
	mediaService := domainMedia.NewService(nil, nil, config.Logger, "/tmp/media_cache")

	// Create newsletter service
	newsletterService := domainNewsletter.NewService(nil) // JIDValidator is optional for now

	// Create community service
	communityService := domainCommunity.NewService()

	return &domainServices{
		session:    sessionService,
		webhook:    webhookService,
		chatwoot:   chatwootService,
		group:      groupService,
		contact:    contactService,
		media:      mediaService,
		newsletter: newsletterService,
		community:  communityService,
	}
}

// createUseCases creates all use cases
func createUseCases(config *ContainerConfig, services *domainServices) *useCases {
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
		services.session,
	)

	webhookUseCase := webhook.NewUseCase(
		config.WebhookRepo,
		services.webhook,
	)

	chatwootUseCase := chatwoot.NewUseCase(
		config.ChatwootRepo,
		config.ChatwootIntegration,
		services.chatwoot,
		config.Logger,
	)

	messageUseCase := message.NewUseCase(
		config.SessionRepo,
		config.WameowManager,
		config.Logger,
	)

	mediaUseCase := media.NewUseCase(
		services.media,
		config.MediaRepo,
		config.Logger,
	)

	groupUseCase := group.NewUseCase(
		nil, // No repository needed for groups
		config.WameowManager,
		services.group,
	)

	contactUseCase := contact.NewUseCase(
		services.contact,
		config.Logger,
	)

	// Create newsletter adapter and use case
	newsletterManager := wameow.NewNewsletterAdapter(config.WameowManager, *config.Logger)
	newsletterUseCase := newsletter.NewUseCase(
		newsletterManager,
		services.newsletter,
		config.SessionRepo,
		*config.Logger,
	)

	// Create community adapter and use case
	communityManager := wameow.NewCommunityAdapter(config.WameowManager, *config.Logger)
	communityUseCase := community.NewUseCase(
		communityManager,
		services.community,
		config.SessionRepo,
		*config.Logger,
	)

	return &useCases{
		common:     commonUseCase,
		session:    sessionUseCase,
		webhook:    webhookUseCase,
		chatwoot:   chatwootUseCase,
		message:    messageUseCase,
		media:      mediaUseCase,
		group:      groupUseCase,
		contact:    contactUseCase,
		newsletter: newsletterUseCase,
		community:  communityUseCase,
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

func (c *Container) GetMediaUseCase() media.UseCase {
	return c.MediaUseCase
}

func (c *Container) GetContactUseCase() contact.UseCase {
	return c.ContactUseCase
}

func (c *Container) GetNewsletterUseCase() newsletter.UseCase {
	return c.NewsletterUseCase
}

func (c *Container) GetCommunityUseCase() community.UseCase {
	return c.CommunityUseCase
}

func (c *Container) GetSessionResolver() func(sessionID string) (ports.WameowManager, error) {
	return func(sessionID string) (ports.WameowManager, error) {
		return nil, fmt.Errorf("session resolver not properly implemented")
	}
}

// createMessageMapper creates a new MessageMapper instance
func createMessageMapper(logger *logger.Logger, repository ports.ChatwootMessageRepository) ports.ChatwootMessageMapper {
	return chatwootIntegration.NewMessageMapper(logger, repository)
}
