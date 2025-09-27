package middleware

import (
	"zpwoot/internal/app"
	"zpwoot/platform/logger"

	"github.com/gofiber/fiber/v2"
)

// Metrics middleware tracks request counts and errors
func Metrics(container *app.Container, logger *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Increment request count
		container.GetCommonUseCase().IncrementRequestCount()

		// Continue with the request
		err := c.Next()

		// If there was an error, increment error count
		if err != nil {
			container.GetCommonUseCase().IncrementErrorCount()

			// Log the error for debugging
			logger.ErrorWithFields("Request error", map[string]interface{}{
				"method":     c.Method(),
				"path":       c.Path(),
				"status":     c.Response().StatusCode(),
				"error":      err.Error(),
				"request_id": c.Locals("request_id"),
			})
		}

		return err
	}
}
