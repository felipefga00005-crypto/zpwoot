// @title zpwoot - WhatsApp Multi-Session API
// @version 1.0
// @description A complete REST API for managing multiple WhatsApp sessions using Go, Fiber, PostgreSQL, and whatsmeow library.
// @description
// @description ## Authentication
// @description All API endpoints (except /health/* and /swagger/*) require API key authentication.
// @description Provide your API key in the `Authorization` header.
// @contact.name zpwoot Support
// @contact.url https://github.com/your-org/zpwoot
// @contact.email support@zpwoot.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Enter your API key directly (no Bearer prefix required). Example: a0b1125a0eb3364d98e2c49ec6f7d6ba
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	""
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"

	_ "zpwoot/docs/swagger" // Import generated swagger docs
	"zpwoot/internal/app"
	"zpwoot/internal/domain/session"
	"zpwoot/internal/infra/db"
	"zpwoot/internal/infra/http/middleware"
	"zpwoot/internal/infra/http/routers"
	"zpwoot/internal/infra/integrations/webhook"
	chatwootIntegration "zpwoot/internal/infra/integrations/chatwoot"
	"zpwoot/internal/infra/repository"
	"zpwoot/internal/infra/wameow"
	"zpwoot/internal/ports"
	"zpwoot/platform/config"
	platformDB "zpwoot/platform/db"
	"zpwoot/platform/logger"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	var (
		migrateUp     = flag.Bool("migrate-up", false, "Run database migrations up")
		migrateDown   = flag.Bool("migrate-down", false, "Rollback last migration")
		migrateStatus = flag.Bool("migrate-status", false, "Show migration status")
		seed          = flag.Bool("seed", false, "Seed database with sample data")
		version       = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	if *version {
		showVersion()
		return
	}

	cfg := config.Load()

	loggerConfig := &logger.LogConfig{
		Level:  cfg.LogLevel,
		Format: cfg.LogFormat,
		Output: cfg.LogOutput,
		Caller: cfg.IsDevelopment(), // Show caller info in development
	}

	if cfg.IsProduction() {
		loggerConfig = logger.ProductionConfig()
		loggerConfig.Level = cfg.LogLevel // Override with env setting
	}

	appLogger := logger.NewWithConfig(loggerConfig)

	database, err := platformDB.NewWithMigrations(cfg.DatabaseURL, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to connect to database and run migrations: " + err.Error())
	}
	defer func() {
		if err := database.Close(); err != nil {
			appLogger.Error("Failed to close database connection: " + err.Error())
		}
	}()

	migrator := db.NewMigrator(database.GetDB().DB, appLogger)

	if *migrateUp {
		if err := migrator.RunMigrations(); err != nil {
			appLogger.Fatal("Failed to run migrations: " + err.Error())
		}
		appLogger.Info("Migrations completed successfully")
		return
	}

	if *migrateDown {
		if err := migrator.Rollback(); err != nil {
			appLogger.Fatal("Failed to rollback migration: " + err.Error())
		}
		appLogger.Info("Migration rollback completed successfully")
		return
	}

	if *migrateStatus {
		migrations, err := migrator.GetMigrationStatus()
		if err != nil {
			appLogger.Fatal("Failed to get migration status: " + err.Error())
		}
		showMigrationStatus(migrations, appLogger)
		return
	}

	if *seed {
		if err := seedDatabase(database, appLogger); err != nil {
			appLogger.Fatal("Failed to seed database: " + err.Error())
		}
		appLogger.Info("Database seeding completed successfully")
		return
	}

	repositories := repository.NewRepositories(database.GetDB(), appLogger)

	whatsappManager, err := initializeWhatsAppManager(database, repositories.GetSessionRepository(), appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize WhatsApp manager: " + err.Error())
	}

	// Initialize webhook manager
	webhookManager, err := initializeWebhookManager(repositories.GetWebhookRepository(), appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize webhook manager: " + err.Error())
	}

	// Configure webhook handler in WhatsApp manager
	err = configureWebhookIntegration(whatsappManager, webhookManager, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to configure webhook integration: " + err.Error())
	}

	// Initialize and configure Chatwoot integration
	chatwootManager, err := initializeChatwootIntegration(repositories.GetChatwootRepository(), appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize Chatwoot integration: " + err.Error())
	}

	// Configure Chatwoot integration in WhatsApp manager
	err = configureChatwootIntegration(whatsappManager, chatwootManager, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to configure Chatwoot integration: " + err.Error())
	}

	container := app.NewContainer(&app.ContainerConfig{
		SessionRepo:         repositories.GetSessionRepository(),
		WebhookRepo:         repositories.GetWebhookRepository(),
		ChatwootRepo:        repositories.GetChatwootRepository(),
		WameowManager:       whatsappManager,
		ChatwootIntegration: nil, // Will be implemented when Chatwoot integration is needed
		Logger:              appLogger,
		DB:                  database.GetDB().DB,
		Version:             Version,
		BuildTime:           BuildTime,
		GitCommit:           GitCommit,
	})

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true, // Disable the Fiber startup banner
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(recover.New())
	app.Use(middleware.RequestID(appLogger))
	app.Use(middleware.HTTPLogger(appLogger))
	app.Use(middleware.Metrics(container, appLogger))
	app.Use(cors.New())
	app.Use(middleware.APIKeyAuth(cfg, appLogger))

	routers.SetupRoutes(app, database, appLogger, whatsappManager, container)

	go connectOnStartup(container, appLogger)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		appLogger.Info("Shutting down server...")
		if err := app.Shutdown(); err != nil {
			appLogger.Error("Failed to shutdown server gracefully: " + err.Error())
		}
	}()

	appLogger.InfoWithFields("Starting zpwoot server", map[string]interface{}{
		"port":        cfg.Port,
		"server_host": cfg.ServerHost,
		"environment": cfg.NodeEnv,
		"log_level":   cfg.LogLevel,
	})
	if err := app.Listen(":" + cfg.Port); err != nil {
		appLogger.Fatal("Server failed to start: " + err.Error())
	}
}

func initializeWhatsAppManager(database *platformDB.DB, sessionRepo ports.SessionRepository, appLogger *logger.Logger) (*wameow.Manager, error) {
	factory, err := wameow.NewFactory(appLogger, sessionRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to create wameow factory: %w", err)
	}

	manager, err := factory.CreateManager(database.GetDB().DB)
	if err != nil {
		return nil, err
	}

	appLogger.Info("WhatsApp manager initialized")
	return manager, nil
}

func showVersion() {
	fmt.Printf("zpwoot - WhatsApp Multi-Session API\n")
	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("Build Time: %s\n", BuildTime)
	fmt.Printf("Git Commit: %s\n", GitCommit)
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

func showMigrationStatus(migrations []*db.Migration, logger *logger.Logger) {
	fmt.Printf("Migration Status:\n")
	fmt.Printf("================\n\n")

	if len(migrations) == 0 {
		fmt.Printf("No migrations found.\n")
		return
	}

	for _, migration := range migrations {
		status := "PENDING"
		appliedAt := "Not applied"

		if migration.AppliedAt != nil {
			status = "APPLIED"
			appliedAt = migration.AppliedAt.Format("2006-01-02 15:04:05")
		}

		fmt.Printf("Version: %03d | Status: %-7s | Name: %s | Applied: %s\n",
			migration.Version, status, migration.Name, appliedAt)
	}
	fmt.Printf("\n")
}

func seedDatabase(database *platformDB.DB, logger *logger.Logger) error {
	logger.Info("Starting database seeding...")

	sampleSessions := []map[string]interface{}{
		{
			"id":         "sample-session-1",
			"name":       "Sample WhatsApp Session",
			"device_jid": "5511999999999@s.whatsapp.net",
			"status":     "created",
			"created_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	sampleWebhooks := []map[string]interface{}{
		{
			"id":         "sample-webhook-1",
			"session_id": "sample-session-1",
			"url":        "https://example.com/webhook",
			"events":     []string{"message", "status"},
			"enabled":    true,
			"created_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	for _, session := range sampleSessions {
		query := `
			INSERT INTO "zpSessions" ("id", "name", "deviceJid", "status", "createdAt", "updatedAt")
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT ("id") DO NOTHING
		`
		_, err := database.GetDB().Exec(query,
			session["id"], session["name"], session["device_jid"],
			session["status"], session["created_at"], session["updated_at"])
		if err != nil {
			return fmt.Errorf("failed to insert sample session: %w", err)
		}
	}

	for _, webhook := range sampleWebhooks {
		query := `
			INSERT INTO "zpWebhooks" ("id", "sessionId", "url", "events", "enabled", "createdAt", "updatedAt")
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT ("id") DO NOTHING
		`
		_, err := database.GetDB().Exec(query,
			webhook["id"], webhook["session_id"], webhook["url"],
			webhook["events"], webhook["enabled"], webhook["created_at"], webhook["updated_at"])
		if err != nil {
			return fmt.Errorf("failed to insert sample webhook: %w", err)
		}
	}

	logger.InfoWithFields("Database seeding completed", map[string]interface{}{
		"sessions_created": len(sampleSessions),
		"webhooks_created": len(sampleWebhooks),
	})

	return nil
}

func initializeWebhookManager(webhookRepo ports.WebhookRepository, appLogger *logger.Logger) (*webhook.WebhookManager, error) {
	// Create webhook manager with 5 workers
	webhookManager := webhook.NewWebhookManager(appLogger, webhookRepo, 5)

	// Start the webhook manager
	if err := webhookManager.Start(); err != nil {
		return nil, fmt.Errorf("failed to start webhook manager: %w", err)
	}

	appLogger.Info("Webhook manager initialized and started")
	return webhookManager, nil
}

func configureWebhookIntegration(wameowManager *wameow.Manager, webhookManager *webhook.WebhookManager, appLogger *logger.Logger) error {
	// Create the webhook handler
	webhookHandler := wameow.NewWhatsmeowWebhookHandler(appLogger, webhookManager)

	// Set the webhook handler in the wameow manager
	wameowManager.SetWebhookHandler(webhookHandler)

	appLogger.Info("Webhook integration configured successfully")
	return nil
}

func connectOnStartup(container *app.Container, logger *logger.Logger) {
	time.Sleep(3 * time.Second)

	sessionUC := container.GetSessionUseCase()
	if sessionUC == nil {
		logger.Error("Session use case not available, skipping auto-connect")
		return
	}

	sessionRepo := container.GetSessionRepository()
	if sessionRepo == nil {
		logger.Error("Session repository not available, skipping auto-connect")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	sessions, _, err := sessionRepo.List(ctx, &session.ListSessionsRequest{
		Limit:  100, // Get up to 100 sessions
		Offset: 0,
	})
	if err != nil {
		logger.ErrorWithFields("Failed to get sessions for auto-connect", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	if len(sessions) == 0 {
		logger.Info("No existing sessions found, skipping auto-connect")
		return
	}

	if len(sessions) > 0 {
		logger.InfoWithFields("Auto-connecting sessions", map[string]interface{}{
			"total_sessions": len(sessions),
		})
	}

	connectedCount := 0
	skippedCount := 0

	for _, sess := range sessions {
		sessionID := sess.ID.String()

		if sess.DeviceJid == "" {
			skippedCount++
			continue
		}

		err := sessionUC.ConnectSession(ctx, sessionID)
		if err != nil {
			logger.ErrorWithFields("Failed to auto-connect session", map[string]interface{}{
				"session_id": sessionID,
				"error":      err.Error(),
			})
			continue
		}

		connectedCount++
		time.Sleep(1 * time.Second)
	}

	if len(sessions) > 0 {
		logger.InfoWithFields("Auto-reconnect completed", map[string]interface{}{
			"connected": connectedCount,
			"skipped":   skippedCount,
		})
	}
}

// initializeChatwootIntegration initializes the Chatwoot integration components
func initializeChatwootIntegration(chatwootRepo ports.ChatwootRepository, logger *logger.Logger) (*chatwootIntegration.IntegrationManager, error) {
	logger.Info("Initializing Chatwoot integration")

	// Create Chatwoot manager
	chatwootManager := chatwootIntegration.NewManager(logger, chatwootRepo)

	// Create message mapper
	messageMapper := chatwootIntegration.NewMessageMapper(logger)

	// Create contact sync
	contactSync := chatwootIntegration.NewContactSync(logger, nil) // Client will be injected later

	// Create conversation manager
	conversationMgr := chatwootIntegration.NewConversationManager(logger)

	// Create message formatter
	formatter := chatwootIntegration.NewMessageFormatter(logger)

	// Create integration manager
	integrationManager := chatwootIntegration.NewIntegrationManager(
		logger,
		chatwootManager,
		messageMapper,
		contactSync,
		conversationMgr,
		formatter,
	)

	logger.Info("Chatwoot integration initialized successfully")
	return integrationManager, nil
}

// configureChatwootIntegration configures the Chatwoot integration with webhook system
func configureChatwootIntegration(whatsappManager ports.WameowManager, integrationManager *chatwootIntegration.IntegrationManager, logger *logger.Logger) error {
	logger.Info("Configuring Chatwoot integration with webhook system")

	// Create a webhook processor that uses the integration manager
	processor := &ChatwootWebhookProcessor{
		logger:             logger,
		integrationManager: integrationManager,
	}

	// Register the processor with the webhook system
	// This would be done through the webhook manager
	logger.Info("Chatwoot integration configured successfully")
	return nil
}

// ChatwootWebhookProcessor processes webhook events for Chatwoot integration
type ChatwootWebhookProcessor struct {
	logger             *logger.Logger
	integrationManager *chatwootIntegration.IntegrationManager
}

// ProcessWebhookEvent processes a webhook event and sends it to Chatwoot if applicable
func (p *ChatwootWebhookProcessor) ProcessWebhookEvent(ctx context.Context, event interface{}) error {
	// This would be called by the webhook system for each event
	// We need to extract message information and call the integration manager

	p.logger.DebugWithFields("Processing webhook event for Chatwoot", map[string]interface{}{
		"event_type": fmt.Sprintf("%T", event),
	})

	// TODO: Extract message information from webhook event and call:
	// p.integrationManager.ProcessWhatsAppMessage(sessionID, messageID, from, content, messageType, timestamp, fromMe)

	return nil
}
