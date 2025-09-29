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
	sessions.Post("/:sessionId/messages/mark-read", messageHandler.MarkAsRead)          // POST /sessions/:sessionId/messages/mark-read

	// Advanced message operations
	sessions.Post("/:sessionId/messages/revoke", messageHandler.RevokeMessage)                  // POST /sessions/:sessionId/messages/revoke
	sessions.Get("/:sessionId/messages/poll/:messageId/results", messageHandler.GetPollResults) // GET /sessions/:sessionId/messages/poll/:messageId/results

	// Group management routes
	groupHandler := handlers.NewGroupHandler(appLogger, container.GetGroupUseCase(), container.GetSessionRepository())
	sessions.Post("/:sessionId/groups/create", groupHandler.CreateGroup)                   // POST /sessions/:sessionId/groups/create
	sessions.Get("/:sessionId/groups", groupHandler.ListGroups)                            // GET /sessions/:sessionId/groups
	sessions.Get("/:sessionId/groups/info", groupHandler.GetGroupInfo)                     // GET /sessions/:sessionId/groups/info?jid=...
	sessions.Post("/:sessionId/groups/participants", groupHandler.UpdateGroupParticipants) // POST /sessions/:sessionId/groups/participants
	sessions.Put("/:sessionId/groups/name", groupHandler.SetGroupName)                     // PUT /sessions/:sessionId/groups/name
	sessions.Put("/:sessionId/groups/description", groupHandler.SetGroupDescription)       // PUT /sessions/:sessionId/groups/description
	sessions.Put("/:sessionId/groups/photo", groupHandler.SetGroupPhoto)                   // PUT /sessions/:sessionId/groups/photo
	sessions.Get("/:sessionId/groups/invite-link", groupHandler.GetGroupInviteLink)        // GET /sessions/:sessionId/groups/invite-link?jid=...
	sessions.Post("/:sessionId/groups/join", groupHandler.JoinGroup)                       // POST /sessions/:sessionId/groups/join
	sessions.Post("/:sessionId/groups/leave", groupHandler.LeaveGroup)                     // POST /sessions/:sessionId/groups/leave
	sessions.Put("/:sessionId/groups/settings", groupHandler.UpdateGroupSettings)          // PUT /sessions/:sessionId/groups/settings

	// Group request management routes
	sessions.Get("/:sessionId/groups/requests", groupHandler.GetGroupRequestParticipants)     // GET /sessions/:sessionId/groups/requests?jid=...
	sessions.Post("/:sessionId/groups/requests", groupHandler.UpdateGroupRequestParticipants) // POST /sessions/:sessionId/groups/requests
	sessions.Put("/:sessionId/groups/join-approval", groupHandler.SetGroupJoinApprovalMode)   // PUT /sessions/:sessionId/groups/join-approval
	sessions.Put("/:sessionId/groups/member-add-mode", groupHandler.SetGroupMemberAddMode)    // PUT /sessions/:sessionId/groups/member-add-mode

	// Advanced group routes
	sessions.Get("/:sessionId/groups/info-from-link", groupHandler.GetGroupInfoFromLink)      // GET /sessions/:sessionId/groups/info-from-link?inviteLink=...
	sessions.Post("/:sessionId/groups/info-from-invite", groupHandler.GetGroupInfoFromInvite) // POST /sessions/:sessionId/groups/info-from-invite
	sessions.Post("/:sessionId/groups/join-with-invite", groupHandler.JoinGroupWithInvite)    // POST /sessions/:sessionId/groups/join-with-invite

	// Newsletter management routes
	newsletterHandler := handlers.NewNewsletterHandler(appLogger, container.GetNewsletterUseCase(), container.GetSessionRepository())
	sessions.Post("/:sessionId/newsletters/create", newsletterHandler.CreateNewsletter)                       // POST /sessions/:sessionId/newsletters/create
	sessions.Get("/:sessionId/newsletters/info", newsletterHandler.GetNewsletterInfo)                         // GET /sessions/:sessionId/newsletters/info?jid=...
	sessions.Post("/:sessionId/newsletters/info-from-invite", newsletterHandler.GetNewsletterInfoWithInvite)  // POST /sessions/:sessionId/newsletters/info-from-invite
	sessions.Post("/:sessionId/newsletters/follow", newsletterHandler.FollowNewsletter)                       // POST /sessions/:sessionId/newsletters/follow
	sessions.Post("/:sessionId/newsletters/unfollow", newsletterHandler.UnfollowNewsletter)                   // POST /sessions/:sessionId/newsletters/unfollow
	sessions.Get("/:sessionId/newsletters/messages", newsletterHandler.GetNewsletterMessages)                 // GET /sessions/:sessionId/newsletters/messages?jid=...&count=20&before=...
	sessions.Get("/:sessionId/newsletters/updates", newsletterHandler.GetNewsletterMessageUpdates)            // GET /sessions/:sessionId/newsletters/updates?jid=...&count=20&since=...&after=...
	sessions.Post("/:sessionId/newsletters/mark-viewed", newsletterHandler.NewsletterMarkViewed)              // POST /sessions/:sessionId/newsletters/mark-viewed
	sessions.Post("/:sessionId/newsletters/send-reaction", newsletterHandler.NewsletterSendReaction)          // POST /sessions/:sessionId/newsletters/send-reaction
	sessions.Post("/:sessionId/newsletters/subscribe-live", newsletterHandler.NewsletterSubscribeLiveUpdates) // POST /sessions/:sessionId/newsletters/subscribe-live
	sessions.Post("/:sessionId/newsletters/toggle-mute", newsletterHandler.NewsletterToggleMute)              // POST /sessions/:sessionId/newsletters/toggle-mute
	sessions.Post("/:sessionId/newsletters/accept-tos", newsletterHandler.AcceptTOSNotice)                    // POST /sessions/:sessionId/newsletters/accept-tos
	sessions.Post("/:sessionId/newsletters/upload", newsletterHandler.UploadNewsletter)                       // POST /sessions/:sessionId/newsletters/upload
	sessions.Post("/:sessionId/newsletters/upload-reader", newsletterHandler.UploadNewsletterReader)          // POST /sessions/:sessionId/newsletters/upload-reader
	sessions.Get("/:sessionId/newsletters", newsletterHandler.GetSubscribedNewsletters)                       // GET /sessions/:sessionId/newsletters

	// Community management routes
	communityHandler := handlers.NewCommunityHandler(appLogger, container.GetCommunityUseCase(), container.GetSessionRepository())
	sessions.Post("/:sessionId/communities/link-group", communityHandler.LinkGroup)     // POST /sessions/:sessionId/communities/link-group
	sessions.Post("/:sessionId/communities/unlink-group", communityHandler.UnlinkGroup) // POST /sessions/:sessionId/communities/unlink-group
	sessions.Get("/:sessionId/communities/info", communityHandler.GetCommunityInfo)     // GET /sessions/:sessionId/communities/info?communityJid=...
	sessions.Get("/:sessionId/communities/subgroups", communityHandler.GetSubGroups)    // GET /sessions/:sessionId/communities/subgroups?communityJid=...

	// Contact management routes
	contactHandler := handlers.NewContactHandler(appLogger, container.GetContactUseCase(), container.GetSessionRepository())
	sessions.Post("/:sessionId/contacts/check", contactHandler.CheckWhatsApp)        // POST /sessions/:sessionId/contacts/check
	sessions.Get("/:sessionId/contacts/avatar", contactHandler.GetProfilePicture)    // GET /sessions/:sessionId/contacts/avatar?jid=...
	sessions.Post("/:sessionId/contacts/info", contactHandler.GetUserInfo)           // POST /sessions/:sessionId/contacts/info
	sessions.Get("/:sessionId/contacts", contactHandler.ListContacts)                // GET /sessions/:sessionId/contacts
	sessions.Post("/:sessionId/contacts/sync", contactHandler.SyncContacts)          // POST /sessions/:sessionId/contacts/sync
	sessions.Get("/:sessionId/contacts/business", contactHandler.GetBusinessProfile) // GET /sessions/:sessionId/contacts/business?jid=...

	webhookHandler := handlers.NewWebhookHandler(container.WebhookUseCase, appLogger)

	// Webhook management routes (padrão find/set - configuração única por sessão)
	sessions.Post("/:sessionId/webhook/set", webhookHandler.SetConfig)    // POST /sessions/:sessionId/webhook/set (create/update/disable)
	sessions.Get("/:sessionId/webhook/find", webhookHandler.FindConfig)   // GET /sessions/:sessionId/webhook/find
	sessions.Post("/:sessionId/webhook/test", webhookHandler.TestWebhook) // POST /sessions/:sessionId/webhook/test

	chatwootHandler := handlers.NewChatwootHandler(container.GetChatwootUseCase(), appLogger)
	sessions.Post("/:sessionId/chatwoot/set", chatwootHandler.SetConfig)                        // POST /sessions/:sessionId/chatwoot/set (create/update)
	sessions.Get("/:sessionId/chatwoot/find", chatwootHandler.FindConfig)                       // GET /sessions/:sessionId/chatwoot/find
	sessions.Post("/:sessionId/chatwoot/contacts/sync", chatwootHandler.SyncContacts)           // POST /sessions/:sessionId/chatwoot/contacts/sync
	sessions.Post("/:sessionId/chatwoot/conversations/sync", chatwootHandler.SyncConversations) // POST /sessions/:sessionId/chatwoot/conversations/sync

	// TODO: Media download routes - implement media use case
	// mediaHandler := handlers.NewMediaHandler(appLogger, container.GetMediaUseCase(), container.GetSessionRepository())
	// sessions.Get("/:sessionId/media/download/:messageId", mediaHandler.DownloadMedia)                    // GET /sessions/:sessionId/media/download/:messageId
}

func setupSessionSpecificRoutes(app *fiber.App, database *db.DB, appLogger *logger.Logger, WameowManager *wameow.Manager, container *app.Container) {
	// Session-specific advanced routes that require additional processing
	// Currently no additional session-specific routes needed
	// All core functionality is handled in setupSessionRoutes
}

func setupGlobalRoutes(app *fiber.App, database *db.DB, appLogger *logger.Logger, WameowManager *wameow.Manager, container *app.Container) {
	// Global webhook info routes
	webhookHandler := handlers.NewWebhookHandler(container.WebhookUseCase, appLogger)
	app.Get("/webhook/events", webhookHandler.GetSupportedEvents) // GET /webhook/events

	// Chatwoot webhook (without authentication - like Evolution API)
	chatwootHandler := handlers.NewChatwootHandler(container.GetChatwootUseCase(), appLogger)
	app.Post("/sessions/:sessionId/chatwoot/webhook", chatwootHandler.ReceiveWebhook) // POST /sessions/:sessionId/chatwoot/webhook
}
