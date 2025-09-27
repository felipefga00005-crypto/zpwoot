package handlers

import (
	"zpwoot/internal/app/common"
	"zpwoot/internal/app/webhook"
	"zpwoot/platform/logger"

	"github.com/gofiber/fiber/v2"
)

type WebhookHandler struct {
	webhookUC webhook.UseCase
	logger    *logger.Logger
}

func NewWebhookHandler(webhookUC webhook.UseCase, appLogger *logger.Logger) *WebhookHandler {
	return &WebhookHandler{
		webhookUC: webhookUC,
		logger:    appLogger,
	}
}

// @Summary Set webhook configuration
// @Description Set or update webhook configuration for a WhatsApp session
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body webhook.SetConfigRequest true "Webhook configuration request"
// @Success 200 {object} webhook.SetConfigResponse "Webhook configuration set successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/webhook/set [post]
func (h *WebhookHandler) SetConfig(c *fiber.Ctx) error {
	sessionID := c.Params("id")
	h.logger.InfoWithFields("Creating webhook config", map[string]interface{}{
		"session_id": sessionID,
	})

	var req webhook.SetConfigRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse webhook request: " + err.Error())
		return c.Status(400).JSON(common.NewErrorResponse("Invalid request body"))
	}

	req.SessionID = &sessionID

	ctx := c.Context()
	result, err := h.webhookUC.SetConfig(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to create webhook: " + err.Error())
		return c.Status(500).JSON(common.NewErrorResponse("Failed to create webhook"))
	}

	response := common.NewSuccessResponse(result, "Webhook configuration created successfully")
	return c.Status(201).JSON(response)
}

// @Summary Get webhook configuration
// @Description Get current webhook configuration for a WhatsApp session
// @Tags Webhooks
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} webhook.WebhookResponse "Webhook configuration retrieved successfully"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/webhook/find [get]
func (h *WebhookHandler) FindConfig(c *fiber.Ctx) error {
	sessionID := c.Params("id")
	h.logger.InfoWithFields("Getting webhook config", map[string]interface{}{
		"session_id": sessionID,
	})
	ctx := c.Context()
	webhook, err := h.webhookUC.FindConfig(ctx, sessionID)
	if err != nil {
		h.logger.Error("Failed to get webhook config: " + err.Error())
		return c.Status(500).JSON(common.NewErrorResponse("Failed to get webhook configuration"))
	}

	response := common.NewSuccessResponse(webhook, "Webhook configuration retrieved successfully")
	return c.JSON(response)
}
