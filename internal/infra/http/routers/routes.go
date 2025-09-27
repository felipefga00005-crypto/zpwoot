package routers

import (
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"

	"zpwoot/internal/app"
	"zpwoot/internal/infra/http/handlers"
	"zpwoot/internal/infra/wameow"
	"zpwoot/platform/db"
	"zpwoot/platform/logger"
)

func SetupRoutes(app *fiber.App, database *db.DB, logger *logger.Logger, WameowManager *wameow.Manager, container *app.Container) {
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Health check endpoints
	healthHandler := handlers.NewHealthHandler(logger, WameowManager)
	app.Get("/health", healthHandler.GetHealth)
	app.Get("/health/wameow", healthHandler.GetWameowHealth)

	setupSessionRoutes(app, logger, WameowManager, container)

	setupSessionSpecificRoutes(app, database, logger, WameowManager, container)

	setupGlobalRoutes(app, database, logger, WameowManager, container)
}

func setupSessionRoutes(app *fiber.App, appLogger *logger.Logger, WameowManager *wameow.Manager, container *app.Container) {
	sessionHandler := handlers.NewSessionHandler(appLogger, container.GetSessionUseCase(), container.GetSessionRepository())

	if WameowManager != nil {
		appLogger.Info("Wameow manager is available for session routes")
	} else {
		appLogger.Warn("Wameow manager is nil - session functionality will be limited")
	}

	sessions := app.Group("/sessions")

	sessions.Post("/create", sessionHandler.CreateSession)              // POST /sessions/create
	sessions.Get("/list", sessionHandler.ListSessions)                  // GET /sessions/list
	sessions.Get("/:sessionId/info", sessionHandler.GetSessionInfo)     // GET /sessions/:sessionId/info
	sessions.Delete("/:sessionId/delete", sessionHandler.DeleteSession) // DELETE /sessions/:sessionId/delete
	sessions.Post("/:sessionId/connect", sessionHandler.ConnectSession) // POST /sessions/:sessionId/connect
	sessions.Post("/:sessionId/logout", sessionHandler.LogoutSession)   // POST /sessions/:sessionId/logout
	sessions.Get("/:sessionId/qr", sessionHandler.GetQRCode)            // GET /sessions/:sessionId/qr
	sessions.Post("/:sessionId/pair", sessionHandler.PairPhone)         // POST /sessions/:sessionId/pair
	sessions.Post("/:sessionId/proxy/set", sessionHandler.SetProxy)     // POST /sessions/:sessionId/proxy/set
	sessions.Get("/:sessionId/proxy/find", sessionHandler.GetProxy)     // GET /sessions/:sessionId/proxy/find

	messageHandler := handlers.NewMessageHandler(container.GetMessageUseCase(), WameowManager, container.GetSessionRepository(), appLogger)
	sessions.Post("/:sessionId/messages/send/text", messageHandler.SendText)            // POST /sessions/:sessionId/messages/send/text
	sessions.Post("/:sessionId/messages/send/media", messageHandler.SendMedia)          // POST /sessions/:sessionId/messages/send/media
	sessions.Post("/:sessionId/messages/send/image", messageHandler.SendImage)          // POST /sessions/:sessionId/messages/send/image
	sessions.Post("/:sessionId/messages/send/audio", messageHandler.SendAudio)          // POST /sessions/:sessionId/messages/send/audio
	sessions.Post("/:sessionId/messages/send/video", messageHandler.SendVideo)          // POST /sessions/:sessionId/messages/send/video
	sessions.Post("/:sessionId/messages/send/document", messageHandler.SendDocument)    // POST /sessions/:sessionId/messages/send/document
	sessions.Post("/:sessionId/messages/send/sticker", messageHandler.SendSticker)      // POST /sessions/:sessionId/messages/send/sticker
	sessions.Post("/:sessionId/messages/send/button", messageHandler.SendButtonMessage) // POST /sessions/:sessionId/messages/send/button
	sessions.Post("/:sessionId/messages/send/contact", messageHandler.SendContact)      // POST /sessions/:sessionId/messages/send/contact
	sessions.Post("/:sessionId/messages/send/list", messageHandler.SendListMessage)     // POST /sessions/:sessionId/messages/send/list
	sessions.Post("/:sessionId/messages/send/location", messageHandler.SendLocation)    // POST /sessions/:sessionId/messages/send/location
	sessions.Post("/:sessionId/messages/send/poll", messageHandler.SendPoll)            // POST /sessions/:sessionId/messages/send/poll
	sessions.Post("/:sessionId/messages/send/reaction", messageHandler.SendReaction)    // POST /sessions/:sessionId/messages/send/reaction
	sessions.Post("/:sessionId/messages/send/presence", messageHandler.SendPresence)    // POST /sessions/:sessionId/messages/send/presence
	sessions.Post("/:sessionId/messages/edit", messageHandler.EditMessage)              // POST /sessions/:sessionId/messages/edit
	sessions.Post("/:sessionId/messages/delete", messageHandler.DeleteMessage)          // POST /sessions/:sessionId/messages/delete
	sessions.Post("/:sessionId/messages/mark-read", messageHandler.MarkAsRead)          // POST /sessions/:sessionId/messages/mark-read

	// Advanced message operations
	sessions.Post("/:sessionId/messages/revoke", messageHandler.RevokeMessage)                  // POST /sessions/:sessionId/messages/revoke
	sessions.Get("/:sessionId/messages/poll/:messageId/results", messageHandler.GetPollResults) // GET /sessions/:sessionId/messages/poll/:messageId/results

	// Group management routes
	groupHandler := handlers.NewGroupHandler(appLogger, container.GetGroupUseCase(), container.GetSessionRepository())
	sessions.Post("/:sessionId/groups/create", groupHandler.CreateGroup)                             // POST /sessions/:sessionId/groups/create
	sessions.Get("/:sessionId/groups", groupHandler.ListGroups)                                      // GET /sessions/:sessionId/groups
	sessions.Get("/:sessionId/groups/:groupJid/info", groupHandler.GetGroupInfo)                     // GET /sessions/:sessionId/groups/:groupJid/info
	sessions.Post("/:sessionId/groups/:groupJid/participants", groupHandler.UpdateGroupParticipants) // POST /sessions/:sessionId/groups/:groupJid/participants
	sessions.Put("/:sessionId/groups/:groupJid/name", groupHandler.SetGroupName)                     // PUT /sessions/:sessionId/groups/:groupJid/name
	sessions.Put("/:sessionId/groups/:groupJid/description", groupHandler.SetGroupDescription)       // PUT /sessions/:sessionId/groups/:groupJid/description
	sessions.Put("/:sessionId/groups/:groupJid/photo", groupHandler.SetGroupPhoto)                   // PUT /sessions/:sessionId/groups/:groupJid/photo
	sessions.Get("/:sessionId/groups/:groupJid/invite-link", groupHandler.GetGroupInviteLink)        // GET /sessions/:sessionId/groups/:groupJid/invite-link
	sessions.Post("/:sessionId/groups/join", groupHandler.JoinGroup)                                 // POST /sessions/:sessionId/groups/join
	sessions.Post("/:sessionId/groups/:groupJid/leave", groupHandler.LeaveGroup)                     // POST /sessions/:sessionId/groups/:groupJid/leave
	sessions.Put("/:sessionId/groups/:groupJid/settings", groupHandler.UpdateGroupSettings)          // PUT /sessions/:sessionId/groups/:groupJid/settings

	webhookHandler := handlers.NewWebhookHandler(container.WebhookUseCase, appLogger)

	sessions.Post("/:sessionId/webhook/set", webhookHandler.SetConfig)  // POST /sessions/:sessionId/webhook/set
	sessions.Get("/:sessionId/webhook/find", webhookHandler.FindConfig) // GET /sessions/:sessionId/webhook/find

	chatwootHandler := handlers.NewChatwootHandler(container.GetChatwootUseCase(), appLogger)
	sessions.Post("/:sessionId/chatwoot/set", chatwootHandler.SetConfig)                        // POST /sessions/:sessionId/chatwoot/set (create/update)
	sessions.Get("/:sessionId/chatwoot/find", chatwootHandler.FindConfig)                       // GET /sessions/:sessionId/chatwoot/find
	sessions.Post("/:sessionId/chatwoot/contacts/sync", chatwootHandler.SyncContacts)           // POST /sessions/:sessionId/chatwoot/contacts/sync
	sessions.Post("/:sessionId/chatwoot/conversations/sync", chatwootHandler.SyncConversations) // POST /sessions/:sessionId/chatwoot/conversations/sync

	// TODO: Media download routes - implement media use case
	// mediaHandler := handlers.NewMediaHandler(appLogger, container.GetMediaUseCase(), container.GetSessionRepository())
	// sessions.Get("/:sessionId/media/download/:messageId", mediaHandler.DownloadMedia)                    // GET /sessions/:sessionId/media/download/:messageId

	// TODO: Contact management routes - implement contact domain and use case
	// contactHandler := handlers.NewContactHandler(appLogger, container.GetContactUseCase(), container.GetSessionRepository())
	// sessions.Post("/:sessionId/contacts/check", contactHandler.CheckWhatsApp)                           // POST /sessions/:sessionId/contacts/check
	// sessions.Get("/:sessionId/contacts/:jid/avatar", contactHandler.GetProfilePicture)                  // GET /sessions/:sessionId/contacts/:jid/avatar
	// sessions.Post("/:sessionId/contacts/info", contactHandler.GetUserInfo)                              // POST /sessions/:sessionId/contacts/info
	// sessions.Get("/:sessionId/contacts", contactHandler.ListContacts)                                   // GET /sessions/:sessionId/contacts
	// sessions.Post("/:sessionId/contacts/sync", contactHandler.SyncContacts)                             // POST /sessions/:sessionId/contacts/sync
	// sessions.Get("/:sessionId/contacts/:jid/business", contactHandler.GetBusinessProfile)               // GET /sessions/:sessionId/contacts/:jid/business
}

func setupSessionSpecificRoutes(app *fiber.App, database *db.DB, appLogger *logger.Logger, WameowManager *wameow.Manager, container *app.Container) {
	// Session-specific advanced routes that require additional processing
	// Currently no additional session-specific routes needed
	// All core functionality is handled in setupSessionRoutes
}

func setupGlobalRoutes(app *fiber.App, database *db.DB, appLogger *logger.Logger, WameowManager *wameow.Manager, container *app.Container) {
	// Global routes that don't depend on specific sessions
	// Currently no global routes needed - all functionality is session-specific
}
