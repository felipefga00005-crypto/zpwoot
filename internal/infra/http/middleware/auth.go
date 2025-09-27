package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"zpwoot/platform/config"
	"zpwoot/platform/logger"
)

// APIKeyAuth creates a middleware that validates API key authentication
func APIKeyAuth(cfg *config.Config, logger *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip authentication for health check and swagger docs
		path := c.Path()
		if path == "/health" || strings.HasPrefix(path, "/swagger") {
			return c.Next()
		}

		// Get API key from Authorization header (direct value, no Bearer prefix)
		apiKey := c.Get("Authorization")
		if apiKey == "" {
			// Fallback to X-API-Key header for compatibility
			apiKey = c.Get("X-API-Key")
		}

		// Validate API key
		if apiKey == "" {
			logger.WarnWithFields("Missing API key", map[string]interface{}{
				"path":   path,
				"method": c.Method(),
				"ip":     c.IP(),
			})
			return c.Status(401).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "API key is required. Provide it via Authorization header or X-API-Key header",
				"code":    "MISSING_API_KEY",
			})
		}

		// Check if API key is valid
		if apiKey != cfg.GlobalAPIKey {
			logger.WarnWithFields("Invalid API key", map[string]interface{}{
				"path":    path,
				"method":  c.Method(),
				"ip":      c.IP(),
				"api_key": maskAPIKey(apiKey),
			})
			return c.Status(401).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid API key",
				"code":    "INVALID_API_KEY",
			})
		}

		// Log successful authentication
		logger.DebugWithFields("API key authenticated", map[string]interface{}{
			"path":    path,
			"method":  c.Method(),
			"ip":      c.IP(),
			"api_key": maskAPIKey(apiKey),
		})

		// Store API key info in context for handlers
		c.Locals("api_key", apiKey)
		c.Locals("authenticated", true)

		return c.Next()
	}
}

// maskAPIKey masks the API key for logging (shows only first 8 and last 4 characters)
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 12 {
		return strings.Repeat("*", len(apiKey))
	}
	return apiKey[:8] + strings.Repeat("*", len(apiKey)-12) + apiKey[len(apiKey)-4:]
}

// GetAPIKeyFromContext retrieves the API key from Fiber context
func GetAPIKeyFromContext(c *fiber.Ctx) string {
	if apiKey, ok := c.Locals("api_key").(string); ok {
		return apiKey
	}
	return ""
}

// IsAuthenticated checks if the request is authenticated
func IsAuthenticated(c *fiber.Ctx) bool {
	if authenticated, ok := c.Locals("authenticated").(bool); ok {
		return authenticated
	}
	return false
}
