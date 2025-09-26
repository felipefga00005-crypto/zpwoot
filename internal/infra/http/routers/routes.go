package routers

import (
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"

	"zpwoot/internal/app"
	"zpwoot/internal/app/common"
	"zpwoot/internal/infra/http/handlers"
	"zpwoot/internal/infra/wameow"
	"zpwoot/platform/db"
	"zpwoot/platform/logger"
)

// SetupRoutes configures all application routes
func SetupRoutes(app *fiber.App, database *db.DB, logger *logger.Logger, WameowManager *wameow.Manager, container *app.Container) {
	// Swagger documentation
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Health check
	// @Summary Health check
	// @Description Check if the API is running and healthy
	// @Tags Health
	// @Produce json
	// @Success 200 {object} object "API is healthy"
	// @Router /health [get]
	app.Get("/health", func(c *fiber.Ctx) error {
		response := &common.HealthResponse{
			Status:  "ok",
			Service: "zpwoot",
		}
		return c.JSON(response)
	})

	// Wameow health check
	// @Summary Wameow health check
	// @Description Check if Wameow manager and whatsmeow tables are available
	// @Tags Health
	// @Produce json
	// @Success 200 {object} object "Wameow manager is healthy"
	// @Router /health/Wameow [get]
	app.Get("/health/Wameow", func(c *fiber.Ctx) error {
		if WameowManager == nil {
			return c.Status(503).JSON(fiber.Map{
				"status":  "error",
				"service": "Wameow",
				"message": "Wameow manager not initialized",
			})
		}

		// Get health check from manager
		healthData := WameowManager.HealthCheck()
		healthData["service"] = "Wameow"
		healthData["message"] = "Wameow manager is healthy and whatsmeow tables are available"

		return c.JSON(healthData)
	})

	// Session management routes
	setupSessionRoutes(app, logger, WameowManager, container)

	// Session-specific routes (grouped by session ID)
	setupSessionSpecificRoutes(app, database, logger, WameowManager, container)

	// Global webhook and chatwoot configuration routes
	setupGlobalRoutes(app, database, logger, WameowManager, container)
}

// setupSessionRoutes configures session management routes
func setupSessionRoutes(app *fiber.App, appLogger *logger.Logger, WameowManager *wameow.Manager, container *app.Container) {
	// Initialize session handler with use case and repository from container
	sessionHandler := handlers.NewSessionHandler(appLogger, container.GetSessionUseCase(), container.GetSessionRepository())

	// Log Wameow manager availability
	if WameowManager != nil {
		appLogger.Info("Wameow manager is available for session routes")
	} else {
		appLogger.Warn("Wameow manager is nil - session functionality will be limited")
	}

	sessions := app.Group("/sessions")

	// Session management routes (supports both UUID and session names)
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

	// Initialize webhook handler for session-specific routes
	webhookHandler := handlers.NewWebhookHandler(container.WebhookUseCase, appLogger)

	// Session-specific webhook configuration (supports both UUID and session names)
	sessions.Post("/:sessionId/webhook/set", webhookHandler.SetConfig)  // POST /sessions/:sessionId/webhook/set
	sessions.Get("/:sessionId/webhook/find", webhookHandler.FindConfig) // GET /sessions/:sessionId/webhook/find

	// Session-specific Chatwoot configuration (simplified to 2 endpoints)
	chatwootHandler := handlers.NewChatwootHandler(container.GetChatwootUseCase(), appLogger)
	sessions.Post("/:sessionId/chatwoot/set", chatwootHandler.SetConfig)  // POST /sessions/:sessionId/chatwoot/set (create/update)
	sessions.Get("/:sessionId/chatwoot/find", chatwootHandler.FindConfig) // GET /sessions/:sessionId/chatwoot/find

	// Message sending routes
	messageHandler := handlers.NewMessageHandler(container.GetMessageUseCase(), WameowManager, container.GetSessionRepository(), appLogger)
	sessions.Post("/:sessionId/messages/send", messageHandler.SendMessage)             // POST /sessions/:sessionId/messages/send (generic)
	sessions.Post("/:sessionId/messages/send/text", messageHandler.SendText)           // POST /sessions/:sessionId/messages/send/text
	sessions.Post("/:sessionId/messages/send/media", messageHandler.SendMedia)          // POST /sessions/:sessionId/messages/send/media
	sessions.Post("/:sessionId/messages/send/image", messageHandler.SendImage)          // POST /sessions/:sessionId/messages/send/image
	sessions.Post("/:sessionId/messages/send/audio", messageHandler.SendAudio)          // POST /sessions/:sessionId/messages/send/audio
	sessions.Post("/:sessionId/messages/send/video", messageHandler.SendVideo)          // POST /sessions/:sessionId/messages/send/video
	sessions.Post("/:sessionId/messages/send/document", messageHandler.SendDocument)    // POST /sessions/:sessionId/messages/send/document
	sessions.Post("/:sessionId/messages/send/sticker", messageHandler.SendSticker)      // POST /sessions/:sessionId/messages/send/sticker
	sessions.Post("/:sessionId/messages/send/button", messageHandler.SendButtonMessage) // POST /sessions/:sessionId/messages/send/button
	sessions.Post("/:sessionId/messages/send/list", messageHandler.SendListMessage)     // POST /sessions/:sessionId/messages/send/list
	sessions.Post("/:sessionId/messages/send/location", messageHandler.SendLocation)    // POST /sessions/:sessionId/messages/send/location
	sessions.Post("/:sessionId/messages/send/contact", messageHandler.SendContact)      // POST /sessions/:sessionId/messages/send/contact
	sessions.Post("/:sessionId/messages/send/reaction", messageHandler.SendReaction)    // POST /sessions/:sessionId/messages/send/reaction
	sessions.Post("/:sessionId/messages/send/presence", messageHandler.SendPresence)    // POST /sessions/:sessionId/messages/send/presence
	sessions.Post("/:sessionId/messages/edit", messageHandler.EditMessage)              // POST /sessions/:sessionId/messages/edit
	sessions.Post("/:sessionId/messages/delete", messageHandler.DeleteMessage)          // POST /sessions/:sessionId/messages/delete

}

// setupSessionSpecificRoutes configures routes grouped by session ID
func setupSessionSpecificRoutes(app *fiber.App, database *db.DB, appLogger *logger.Logger, WameowManager *wameow.Manager, container *app.Container) {
	// Placeholder for future session-specific routes if needed
	// Currently all required routes are in setupSessionRoutes
}

// setupGlobalRoutes configures global routes (currently none needed)
func setupGlobalRoutes(app *fiber.App, database *db.DB, appLogger *logger.Logger, WameowManager *wameow.Manager, container *app.Container) {
	// All configuration routes are now session-specific
	// This function is kept for future global routes if needed
}
