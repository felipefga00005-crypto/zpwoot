package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"zpwoot/internal/app/chatwoot"
	domainChatwoot "zpwoot/internal/domain/chatwoot"
	"zpwoot/pkg/errors"
	"zpwoot/platform/logger"
)

type ChatwootHandler struct {
	chatwootUC chatwoot.UseCase
	logger     *logger.Logger
}

type ChatwootService interface {
	CreateConfig(ctx context.Context, req *chatwoot.CreateChatwootConfigRequest) (*chatwoot.CreateChatwootConfigResponse, error)
	GetConfig(ctx context.Context) (*chatwoot.ChatwootConfigResponse, error)
	UpdateConfig(ctx context.Context, req *chatwoot.UpdateChatwootConfigRequest) (*chatwoot.ChatwootConfigResponse, error)
	DeleteConfig(ctx context.Context) error
	SyncContact(ctx context.Context, req *chatwoot.SyncContactRequest) (*chatwoot.SyncContactResponse, error)
	SyncConversation(ctx context.Context, req *chatwoot.SyncConversationRequest) (*chatwoot.SyncConversationResponse, error)
	ProcessWebhook(ctx context.Context, payload *chatwoot.ChatwootWebhookPayload) error
}

func NewChatwootHandler(chatwootUC chatwoot.UseCase, logger *logger.Logger) *ChatwootHandler {
	return &ChatwootHandler{
		chatwootUC: chatwootUC,
		logger:     logger,
	}
}

// @Summary Set Chatwoot configuration
// @Description Set or update Chatwoot integration configuration for a WhatsApp session
// @Tags Chatwoot
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body chatwoot.CreateChatwootConfigRequest true "Chatwoot configuration request"
// @Success 200 {object} chatwoot.CreateChatwootConfigResponse "Chatwoot configuration set successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/chatwoot/set [post]
func (h *ChatwootHandler) CreateConfig(c *fiber.Ctx) error {
	var req chatwoot.CreateChatwootConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	config, err := h.chatwootUC.CreateConfig(c.Context(), &req)
	if err != nil {
		if appErr := errors.GetAppError(err); appErr != nil {
			return c.Status(appErr.Code).JSON(fiber.Map{
				"error":   appErr.Message,
				"details": appErr.Details,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"data":    config,
	})
}

func (h *ChatwootHandler) GetConfig(c *fiber.Ctx) error {
	config, err := h.chatwootUC.GetConfig(c.Context())
	if err != nil {
		if appErr := errors.GetAppError(err); appErr != nil {
			return c.Status(appErr.Code).JSON(fiber.Map{
				"error": appErr.Message,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    config,
	})
}

func (h *ChatwootHandler) UpdateConfig(c *fiber.Ctx) error {
	var req chatwoot.UpdateChatwootConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	config, err := h.chatwootUC.UpdateConfig(c.Context(), &req)
	if err != nil {
		if appErr := errors.GetAppError(err); appErr != nil {
			return c.Status(appErr.Code).JSON(fiber.Map{
				"error":   appErr.Message,
				"details": appErr.Details,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    config,
	})
}

func (h *ChatwootHandler) DeleteConfig(c *fiber.Ctx) error {
	err := h.chatwootUC.DeleteConfig(c.Context())
	if err != nil {
		if appErr := errors.GetAppError(err); appErr != nil {
			return c.Status(appErr.Code).JSON(fiber.Map{
				"error": appErr.Message,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Chatwoot configuration deleted successfully",
	})
}

func (h *ChatwootHandler) SyncContacts(c *fiber.Ctx) error {
	var req chatwoot.SyncContactRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	contact, err := h.chatwootUC.SyncContact(c.Context(), &req)
	if err != nil {
		if appErr := errors.GetAppError(err); appErr != nil {
			return c.Status(appErr.Code).JSON(fiber.Map{
				"error":   appErr.Message,
				"details": appErr.Details,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    contact,
	})
}

func (h *ChatwootHandler) SyncConversations(c *fiber.Ctx) error {
	var req chatwoot.SyncConversationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	conversation, err := h.chatwootUC.SyncConversation(c.Context(), &req)
	if err != nil {
		if appErr := errors.GetAppError(err); appErr != nil {
			return c.Status(appErr.Code).JSON(fiber.Map{
				"error":   appErr.Message,
				"details": appErr.Details,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    conversation,
	})
}

func (h *ChatwootHandler) ReceiveWebhook(c *fiber.Ctx) error {
	var payload chatwoot.ChatwootWebhookPayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid webhook payload",
		})
	}

	if !domainChatwoot.IsValidChatwootEvent(payload.Event) {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid event type",
			"event": payload.Event,
		})
	}

	err := h.chatwootUC.ProcessWebhook(c.Context(), &payload)
	if err != nil {
		if appErr := errors.GetAppError(err); appErr != nil {
			return c.Status(appErr.Code).JSON(fiber.Map{
				"error":   appErr.Message,
				"details": appErr.Details,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Webhook processed successfully",
		"event":   payload.Event,
	})
}

func (h *ChatwootHandler) TestConnection(c *fiber.Ctx) error {
	ctx := c.Context()

	result, err := h.chatwootUC.TestConnection(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Chatwoot connection test failed",
			"error":   err.Error(),
			"status":  "failed",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Chatwoot connection test completed",
		"data":    result,
		"status":  "connected",
	})
}

func (h *ChatwootHandler) GetStats(c *fiber.Ctx) error {
	ctx := c.Context()

	stats, err := h.chatwootUC.GetStats(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get Chatwoot statistics",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    stats,
	})
}

// @Summary Set Chatwoot configuration
// @Description Set or update Chatwoot integration configuration for a WhatsApp session
// @Tags Chatwoot
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body chatwoot.CreateChatwootConfigRequest true "Chatwoot configuration request"
// @Success 200 {object} chatwoot.CreateChatwootConfigResponse "Chatwoot configuration set successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/chatwoot/set [post]
func (h *ChatwootHandler) SetConfig(c *fiber.Ctx) error {

	var req chatwoot.CreateChatwootConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	ctx := c.Context()

	_, err := h.chatwootUC.GetConfig(ctx)

	if err != nil {
		result, createErr := h.chatwootUC.CreateConfig(ctx, &req)
		if createErr != nil {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Failed to create Chatwoot configuration",
				"error":   createErr.Error(),
			})
		}

		return c.Status(201).JSON(fiber.Map{
			"success": true,
			"message": "Chatwoot configuration created successfully",
			"data":    result,
		})
	}

	updateReq := chatwoot.UpdateChatwootConfigRequest{
		URL:       &req.URL,
		APIKey:    &req.APIKey,
		AccountID: &req.AccountID,
		InboxID:   req.InboxID,
	}

	result, updateErr := h.chatwootUC.UpdateConfig(ctx, &updateReq)
	if updateErr != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update Chatwoot configuration",
			"error":   updateErr.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Chatwoot configuration updated successfully",
		"data":    result,
	})
}

// @Summary Get Chatwoot configuration
// @Description Get current Chatwoot integration configuration for a WhatsApp session
// @Tags Chatwoot
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} chatwoot.ChatwootConfigResponse "Chatwoot configuration retrieved successfully"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/chatwoot/find [get]
func (h *ChatwootHandler) FindConfig(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	ctx := c.Context()
	config, err := h.chatwootUC.GetConfig(ctx)

	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"success":    false,
			"message":    "Chatwoot configuration not found for this session",
			"session_id": sessionID,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Chatwoot configuration found",
		"data":    config,
	})
}
