// @title zpwoot - WhatsApp Multi-Session API
// @version 1.0
// @description A complete REST API for managing multiple WhatsApp sessions using Go, Fiber, PostgreSQL, and whatsmeow library.
// @description
// @description ## Authentication
// @description All API endpoints (except /health and /swagger/*) require API key authentication.
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
// @description Enter your API key directly (no Bearer prefix required). Example: dev-api-key-12345
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
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

	appLogger.Info("Initializing WhatsApp manager and creating whatsmeow tables...")
	whatsappManager, err := initializeWhatsAppManager(database, repositories.GetSessionRepository(), appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize WhatsApp manager: " + err.Error())
	}
	appLogger.Info("WhatsApp manager initialized successfully with whatsmeow tables created")

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
	appLogger.Info("Creating WhatsApp manager factory...")

	factory := wameow.NewFactory(appLogger, sessionRepo)

	appLogger.Info("Creating WhatsApp manager with database connection...")
	manager, err := factory.CreateManager(database.GetDB().DB)
	if err != nil {
		return nil, err
	}

	appLogger.Info("WhatsApp manager created successfully - whatsmeow tables are now available")
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

func connectOnStartup(container *app.Container, logger *logger.Logger) {
	logger.Info("Starting connection process for all sessions on startup")

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

	logger.InfoWithFields("Found sessions for auto-connect", map[string]interface{}{
		"total_sessions": len(sessions),
	})

	connectedCount := 0
	skippedCount := 0

	for _, sess := range sessions {
		sessionID := sess.ID.String()

		if sess.DeviceJid == "" {
			logger.InfoWithFields("Skipping session without device JID (never paired)", map[string]interface{}{
				"session_id":   sessionID,
				"session_name": sess.Name,
			})
			skippedCount++
			continue
		}

		logger.InfoWithFields("Attempting to reconnect session with saved credentials", map[string]interface{}{
			"session_id":    sessionID,
			"session_name":  sess.Name,
			"device_jid":    sess.DeviceJid,
			"was_connected": sess.IsConnected,
		})

		err := sessionUC.ConnectSession(ctx, sessionID)
		if err != nil {
			logger.ErrorWithFields("Failed to auto-connect session", map[string]interface{}{
				"session_id":   sessionID,
				"session_name": sess.Name,
				"error":        err.Error(),
			})
			continue
		}

		connectedCount++
		logger.InfoWithFields("Successfully initiated reconnection for session", map[string]interface{}{
			"session_id":   sessionID,
			"session_name": sess.Name,
		})

		time.Sleep(1 * time.Second)
	}

	logger.InfoWithFields("Auto-reconnect process completed", map[string]interface{}{
		"total_sessions":        len(sessions),
		"reconnection_attempts": connectedCount,
		"skipped_sessions":      skippedCount,
	})
}
