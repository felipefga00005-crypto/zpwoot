package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port       string
	ServerHost string
	LogLevel   string
	LogFormat  string // "json" or "console"
	LogOutput  string // "stdout", "stderr", "file", or file path

	// Database
	DatabaseURL string

	// Wameow
	WameowLogLevel string

	// Global Webhooks
	GlobalWebhookURL string
	WebhookSecret    string

	// Security
	GlobalAPIKey string

	// Environment
	NodeEnv string
}

func Load() *Config {
	// Load .env file if exists (ignore error if file doesn't exist)
	_ = godotenv.Load()

	return &Config{
		Port:       getEnv("PORT", "8080"),
		ServerHost: getEnv("SERVER_HOST", "http://localhost:8080"),
		LogLevel:   getEnv("LOG_LEVEL", "info"),
		LogFormat:  getEnv("LOG_FORMAT", "console"),
		LogOutput:  getEnv("LOG_OUTPUT", "stdout"),

		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:password@localhost/zpwoot?sslmode=disable"),

		WameowLogLevel: getEnv("WA_LOG_LEVEL", "INFO"),

		GlobalWebhookURL: getEnv("GLOBAL_WEBHOOK_URL", ""),
		WebhookSecret:    getEnv("WEBHOOK_SECRET", ""),

		GlobalAPIKey: getEnv("ZP_API_KEY", "a0b1125a0eb3364d98e2c49ec6f7d6ba"),

		NodeEnv: getEnv("NODE_ENV", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Removed unused helper functions getEnvAsInt and getEnvAsBool
// They can be added back when needed

// Helper methods for configuration

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.NodeEnv == "production"
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.NodeEnv == "development"
}

// IsTest returns true if running in test environment
func (c *Config) IsTest() bool {
	return c.NodeEnv == "test"
}

// GetServerURL returns the full server URL
func (c *Config) GetServerURL() string {
	return c.ServerHost
}

// HasWebhookSecret returns true if webhook secret is configured
func (c *Config) HasWebhookSecret() bool {
	return c.WebhookSecret != ""
}
