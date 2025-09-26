package handlers

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"zpwoot/internal/app/common"
	messageApp "zpwoot/internal/app/message"
	"zpwoot/internal/infra/http/helpers"
	"zpwoot/internal/infra/wameow"
	"zpwoot/platform/logger"
)

// MessageHandler handles message-related HTTP requests
type MessageHandler struct {
	messageUC       messageApp.UseCase
	wameowManager   *wameow.Manager
	sessionResolver *helpers.SessionResolver
	logger          *logger.Logger
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(
	messageUC messageApp.UseCase,
	wameowManager *wameow.Manager,
	sessionRepo helpers.SessionRepository,
	logger *logger.Logger,
) *MessageHandler {
	// Create session resolver using the provided session repository
	sessionResolver := helpers.NewSessionResolver(logger, sessionRepo)

	return &MessageHandler{
		messageUC:       messageUC,
		wameowManager:   wameowManager,
		sessionResolver: sessionResolver,
		logger:          logger,
	}
}



// SendMessage sends a message through WhatsApp
// @Summary Send WhatsApp message
// @Description Send a message through WhatsApp. Supports text, image, audio, video, document, location, and contact messages. Media can be provided via URL or base64.
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth

// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body messageApp.SendMessageRequest true "Message request"
// @Success 200 {object} common.SuccessResponse{data=messageApp.SendMessageResponse} "Message sent successfully"
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 404 {object} common.ErrorResponse "Session not found"
// @Failure 500 {object} common.ErrorResponse "Internal server error"
// @Router /sessions/{sessionId}/messages/send [post]
func (h *MessageHandler) SendMessage(c *fiber.Ctx) error {
	sessionIdentifier := c.Params("sessionId")
	if sessionIdentifier == "" {
		h.logger.Warn("Session identifier is required")
		return c.Status(400).JSON(common.NewErrorResponse("Session identifier is required"))
	}

	// Parse request body
	var req messageApp.SendMessageRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.ErrorWithFields("Failed to parse request body", map[string]interface{}{
			"error": err.Error(),
		})
		return c.Status(400).JSON(common.NewErrorResponse("Invalid request body"))
	}

	// Validate required fields
	if req.To == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Recipient (to) is required"))
	}

	if req.Type == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Message type is required"))
	}

	// Normalize message type
	req.Type = strings.ToLower(req.Type)

	// Validate message type
	validTypes := []string{"text", "image", "audio", "video", "document", "sticker", "location", "contact"}
	isValidType := false
	for _, validType := range validTypes {
		if req.Type == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid message type. Supported types: " + strings.Join(validTypes, ", ")))
	}

	// Resolve session
	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		h.logger.ErrorWithFields("Failed to resolve session", map[string]interface{}{
			"session_identifier": sessionIdentifier,
			"error":              err.Error(),
		})
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

	// Send message
	ctx := c.Context()
	response, err := h.messageUC.SendMessage(ctx, sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to send message", map[string]interface{}{
			"session_id": sess.ID.String(),
			"to":         req.To,
			"type":       req.Type,
			"error":      err.Error(),
		})

		// Check for specific error types
		if strings.Contains(err.Error(), "not connected") {
			return c.Status(400).JSON(common.NewErrorResponse("Session is not connected"))
		}
		if strings.Contains(err.Error(), "not logged in") {
			return c.Status(400).JSON(common.NewErrorResponse("Session is not logged in"))
		}
		if strings.Contains(err.Error(), "invalid request") {
			return c.Status(400).JSON(common.NewErrorResponse(err.Error()))
		}
		if strings.Contains(err.Error(), "failed to process media") {
			return c.Status(400).JSON(common.NewErrorResponse("Failed to process media: " + err.Error()))
		}

		return c.Status(500).JSON(common.NewErrorResponse("Failed to send message"))
	}

	h.logger.InfoWithFields("Message sent successfully", map[string]interface{}{
		"session_id": sess.ID.String(),
		"to":         req.To,
		"type":       req.Type,
		"message_id": response.ID,
	})

	return c.JSON(common.NewSuccessResponse(response, "Message sent successfully"))
}

// SendTextMessage sends a text message (convenience endpoint)
// @Summary Send text message
// @Description Send a simple text message through WhatsApp
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body messageApp.TextMessageRequest true "Text message request"
// @Success 200 {object} common.SuccessResponse{data=messageApp.SendMessageResponse} "Message sent successfully"
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 404 {object} common.ErrorResponse "Session not found"
// @Failure 500 {object} common.ErrorResponse "Internal server error"
// @Router /sessions/{sessionId}/messages/text [post]
func (h *MessageHandler) SendTextMessage(c *fiber.Ctx) error {
	sessionIdentifier := c.Params("sessionId")
	if sessionIdentifier == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Session identifier is required"))
	}

	// Parse simple text request
	var textReq struct {
		To   string `json:"to" validate:"required"`
		Body string `json:"body" validate:"required"`
	}

	if err := c.BodyParser(&textReq); err != nil {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid request body"))
	}

	if textReq.To == "" || textReq.Body == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Both 'to' and 'body' are required"))
	}

	// Convert to full message request
	req := messageApp.SendMessageRequest{
		To:   textReq.To,
		Type: "text",
		Body: textReq.Body,
	}

	// Resolve session
	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

	// Send message
	ctx := c.Context()
	response, err := h.messageUC.SendMessage(ctx, sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to send text message", map[string]interface{}{
			"session_id": sess.ID.String(),
			"to":         req.To,
			"error":      err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			return c.Status(400).JSON(common.NewErrorResponse("Session is not connected"))
		}

		return c.Status(500).JSON(common.NewErrorResponse("Failed to send message"))
	}

	return c.JSON(common.NewSuccessResponse(response, "Text message sent successfully"))
}

// SendText sends a text message (convenience endpoint)
// @Summary Send text message
// @Description Send a simple text message through WhatsApp
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body messageApp.TextMessageRequest true "Text message request"
// @Success 200 {object} common.SuccessResponse{data=messageApp.SendMessageResponse} "Message sent successfully"
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 404 {object} common.ErrorResponse "Session not found"
// @Failure 500 {object} common.ErrorResponse "Internal server error"
// @Router /sessions/{sessionId}/messages/send/text [post]
func (h *MessageHandler) SendText(c *fiber.Ctx) error {
	return h.sendSpecificMessageType(c, "text")
}

// SendMedia sends a media message (generic media endpoint)
// @Summary Send media message
// @Description Send a media message (image, audio, video, document) through WhatsApp
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body messageApp.MediaMessageRequest true "Media message request"
// @Success 200 {object} common.SuccessResponse{data=messageApp.SendMessageResponse} "Message sent successfully"
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 404 {object} common.ErrorResponse "Session not found"
// @Failure 500 {object} common.ErrorResponse "Internal server error"
// @Router /sessions/{sessionId}/messages/send/media [post]
func (h *MessageHandler) SendMedia(c *fiber.Ctx) error {
	return h.SendMessage(c) // Reuse the generic send message logic
}

// SendImage sends an image message
// @Summary Send image message
// @Description Send an image message through WhatsApp
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body messageApp.MediaMessageRequest true "Image message request"
// @Success 200 {object} common.SuccessResponse{data=messageApp.SendMessageResponse} "Message sent successfully"
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 404 {object} common.ErrorResponse "Session not found"
// @Failure 500 {object} common.ErrorResponse "Internal server error"
// @Router /sessions/{sessionId}/messages/send/image [post]
func (h *MessageHandler) SendImage(c *fiber.Ctx) error {
	return h.sendSpecificMessageType(c, "image")
}

// SendAudio sends an audio message
// @Summary Send audio message
// @Description Send an audio message through WhatsApp
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body messageApp.MediaMessageRequest true "Audio message request"
// @Success 200 {object} common.SuccessResponse{data=messageApp.SendMessageResponse} "Message sent successfully"
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 404 {object} common.ErrorResponse "Session not found"
// @Failure 500 {object} common.ErrorResponse "Internal server error"
// @Router /sessions/{sessionId}/messages/send/audio [post]
func (h *MessageHandler) SendAudio(c *fiber.Ctx) error {
	return h.sendSpecificMessageType(c, "audio")
}

// SendVideo sends a video message
// @Summary Send video message
// @Description Send a video message through WhatsApp
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body messageApp.MediaMessageRequest true "Video message request"
// @Success 200 {object} common.SuccessResponse{data=messageApp.SendMessageResponse} "Message sent successfully"
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 404 {object} common.ErrorResponse "Session not found"
// @Failure 500 {object} common.ErrorResponse "Internal server error"
// @Router /sessions/{sessionId}/messages/send/video [post]
func (h *MessageHandler) SendVideo(c *fiber.Ctx) error {
	return h.sendSpecificMessageType(c, "video")
}

// SendDocument sends a document message
// @Summary Send document message
// @Description Send a document message through WhatsApp
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body messageApp.MediaMessageRequest true "Document message request"
// @Success 200 {object} common.SuccessResponse{data=messageApp.SendMessageResponse} "Message sent successfully"
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 404 {object} common.ErrorResponse "Session not found"
// @Failure 500 {object} common.ErrorResponse "Internal server error"
// @Router /sessions/{sessionId}/messages/send/document [post]
func (h *MessageHandler) SendDocument(c *fiber.Ctx) error {
	return h.sendSpecificMessageType(c, "document")
}

// SendSticker sends a sticker message
// @Summary Send sticker message
// @Description Send a sticker message through WhatsApp
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body messageApp.MediaMessageRequest true "Sticker message request"
// @Success 200 {object} common.SuccessResponse{data=messageApp.SendMessageResponse} "Message sent successfully"
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 404 {object} common.ErrorResponse "Session not found"
// @Failure 500 {object} common.ErrorResponse "Internal server error"
// @Router /sessions/{sessionId}/messages/send/sticker [post]
func (h *MessageHandler) SendSticker(c *fiber.Ctx) error {
	return h.sendSpecificMessageType(c, "sticker")
}

// SendLocation sends a location message
// @Summary Send location message
// @Description Send a location message through WhatsApp
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body messageApp.LocationMessageRequest true "Location message request"
// @Success 200 {object} common.SuccessResponse{data=messageApp.SendMessageResponse} "Message sent successfully"
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 404 {object} common.ErrorResponse "Session not found"
// @Failure 500 {object} common.ErrorResponse "Internal server error"
// @Router /sessions/{sessionId}/messages/send/location [post]
func (h *MessageHandler) SendLocation(c *fiber.Ctx) error {
	return h.sendSpecificMessageType(c, "location")
}

// SendContact sends a contact message
// @Summary Send contact message
// @Description Send a contact message through WhatsApp
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body messageApp.ContactMessageRequest true "Contact message request"
// @Success 200 {object} common.SuccessResponse{data=messageApp.SendMessageResponse} "Message sent successfully"
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 404 {object} common.ErrorResponse "Session not found"
// @Failure 500 {object} common.ErrorResponse "Internal server error"
// @Router /sessions/{sessionId}/messages/send/contact [post]
func (h *MessageHandler) SendContact(c *fiber.Ctx) error {
	return h.sendSpecificMessageType(c, "contact")
}

// SendButtonMessage sends a button message
// @Summary Send button message
// @Description Send a message with interactive buttons through WhatsApp
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body messageApp.ButtonMessageRequest true "Button message request"
// @Success 200 {object} common.SuccessResponse{data=messageApp.MessageResponse} "Button message sent successfully"
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 404 {object} common.ErrorResponse "Session not found"
// @Failure 500 {object} common.ErrorResponse "Internal server error"
// @Router /sessions/{sessionId}/messages/send/button [post]
func (h *MessageHandler) SendButtonMessage(c *fiber.Ctx) error {
	sessionIdentifier := c.Params("sessionId")
	if sessionIdentifier == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Session identifier is required"))
	}

	// Parse button message request
	var buttonReq struct {
		To      string `json:"to" validate:"required"`
		Body    string `json:"body" validate:"required"`
		Buttons []struct {
			ID   string `json:"id" validate:"required"`
			Text string `json:"text" validate:"required"`
		} `json:"buttons" validate:"required"`
	}

	if err := c.BodyParser(&buttonReq); err != nil {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid request body"))
	}

	if buttonReq.To == "" || buttonReq.Body == "" || len(buttonReq.Buttons) == 0 {
		return c.Status(400).JSON(common.NewErrorResponse("'to', 'body', and 'buttons' are required"))
	}

	// Convert buttons to the format expected by the manager
	var buttons []map[string]string
	for _, button := range buttonReq.Buttons {
		buttons = append(buttons, map[string]string{
			"id":   button.ID,
			"text": button.Text,
		})
	}

	// Resolve session
	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

	// Send button message using the real implementation
	result, err := h.wameowManager.SendButtonMessage(sess.ID.String(), buttonReq.To, buttonReq.Body, buttons)
	if err != nil {
		h.logger.ErrorWithFields("Failed to send button message", map[string]interface{}{
			"session_id": sess.ID.String(),
			"to":         buttonReq.To,
			"error":      err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			return c.Status(400).JSON(common.NewErrorResponse("Session is not connected"))
		}

		return c.Status(500).JSON(common.NewErrorResponse("Failed to send button message"))
	}

	response := &messageApp.SendMessageResponse{
		ID:        result.MessageID,
		Status:    result.Status,
		Timestamp: result.Timestamp,
	}

	return c.JSON(common.NewSuccessResponse(response, "Button message sent successfully"))
}

// SendListMessage sends a list message
// @Summary Send list message
// @Description Send a message with interactive list through WhatsApp
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body messageApp.ListMessageRequest true "List message request"
// @Success 200 {object} common.SuccessResponse{data=messageApp.MessageResponse} "List message sent successfully"
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 404 {object} common.ErrorResponse "Session not found"
// @Failure 500 {object} common.ErrorResponse "Internal server error"
// @Router /sessions/{sessionId}/messages/send/list [post]
func (h *MessageHandler) SendListMessage(c *fiber.Ctx) error {
	sessionIdentifier := c.Params("sessionId")
	if sessionIdentifier == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Session identifier is required"))
	}

	// Parse list message request
	var listReq struct {
		To         string `json:"to" validate:"required"`
		Body       string `json:"body" validate:"required"`
		ButtonText string `json:"buttonText" validate:"required"`
		Sections   []struct {
			Title string `json:"title" validate:"required"`
			Rows  []struct {
				ID          string `json:"id" validate:"required"`
				Title       string `json:"title" validate:"required"`
				Description string `json:"description,omitempty"`
			} `json:"rows" validate:"required"`
		} `json:"sections" validate:"required"`
	}

	if err := c.BodyParser(&listReq); err != nil {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid request body"))
	}

	if listReq.To == "" || listReq.Body == "" || len(listReq.Sections) == 0 {
		return c.Status(400).JSON(common.NewErrorResponse("'to', 'body', and 'sections' are required"))
	}

	// Convert sections to the format expected by the manager
	var sections []map[string]interface{}
	for _, section := range listReq.Sections {
		var rows []interface{}
		for _, row := range section.Rows {
			rows = append(rows, map[string]interface{}{
				"id":          row.ID,
				"title":       row.Title,
				"description": row.Description,
			})
		}
		sections = append(sections, map[string]interface{}{
			"title": section.Title,
			"rows":  rows,
		})
	}

	// Resolve session
	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

	// Send list message using the real implementation
	result, err := h.wameowManager.SendListMessage(sess.ID.String(), listReq.To, listReq.Body, listReq.ButtonText, sections)
	if err != nil {
		h.logger.ErrorWithFields("Failed to send list message", map[string]interface{}{
			"session_id": sess.ID.String(),
			"to":         listReq.To,
			"error":      err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			return c.Status(400).JSON(common.NewErrorResponse("Session is not connected"))
		}

		return c.Status(500).JSON(common.NewErrorResponse("Failed to send list message"))
	}

	response := &messageApp.SendMessageResponse{
		ID:        result.MessageID,
		Status:    result.Status,
		Timestamp: result.Timestamp,
	}

	return c.JSON(common.NewSuccessResponse(response, "List message sent successfully"))
}

// SendReaction sends a reaction to a message
// @Summary Send reaction
// @Description Send a reaction (emoji) to a specific message
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body messageApp.ReactionMessageRequest true "Reaction request"
// @Success 200 {object} common.SuccessResponse{data=messageApp.ReactionResponse} "Reaction sent successfully"
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 404 {object} common.ErrorResponse "Session not found"
// @Failure 500 {object} common.ErrorResponse "Internal server error"
// @Router /sessions/{sessionId}/messages/send/reaction [post]
func (h *MessageHandler) SendReaction(c *fiber.Ctx) error {
	sessionIdentifier := c.Params("sessionId")
	if sessionIdentifier == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Session identifier is required"))
	}

	// Parse reaction request
	var reactionReq struct {
		To        string `json:"to" validate:"required"`
		MessageID string `json:"messageId" validate:"required"`
		Reaction  string `json:"reaction" validate:"required"`
	}

	if err := c.BodyParser(&reactionReq); err != nil {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid request body"))
	}

	if reactionReq.To == "" || reactionReq.MessageID == "" || reactionReq.Reaction == "" {
		return c.Status(400).JSON(common.NewErrorResponse("'to', 'messageId', and 'reaction' are required"))
	}

	// Resolve session
	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

	// Send reaction using the real implementation
	err = h.wameowManager.SendReaction(sess.ID.String(), reactionReq.To, reactionReq.MessageID, reactionReq.Reaction)
	if err != nil {
		h.logger.ErrorWithFields("Failed to send reaction", map[string]interface{}{
			"session_id": sess.ID.String(),
			"to":         reactionReq.To,
			"message_id": reactionReq.MessageID,
			"error":      err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			return c.Status(400).JSON(common.NewErrorResponse("Session is not connected"))
		}

		return c.Status(500).JSON(common.NewErrorResponse("Failed to send reaction"))
	}

	response := map[string]interface{}{
		"id":        reactionReq.MessageID,
		"reaction":  reactionReq.Reaction,
		"status":    "sent",
		"timestamp": time.Now(),
	}

	return c.JSON(common.NewSuccessResponse(response, "Reaction sent successfully"))
}

// SendPresence sends presence information
// @Summary Send presence
// @Description Send presence information (typing, online, etc.)
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body messageApp.PresenceMessageRequest true "Presence request"
// @Success 200 {object} common.SuccessResponse{data=messageApp.PresenceResponse} "Presence sent successfully"
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 404 {object} common.ErrorResponse "Session not found"
// @Failure 500 {object} common.ErrorResponse "Internal server error"
// @Router /sessions/{sessionId}/messages/send/presence [post]
func (h *MessageHandler) SendPresence(c *fiber.Ctx) error {
	sessionIdentifier := c.Params("sessionId")
	if sessionIdentifier == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Session identifier is required"))
	}

	// Parse presence request
	var presenceReq struct {
		To       string `json:"to" validate:"required"`
		Presence string `json:"presence" validate:"required"` // typing, online, offline
	}

	if err := c.BodyParser(&presenceReq); err != nil {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid request body"))
	}

	if presenceReq.To == "" || presenceReq.Presence == "" {
		return c.Status(400).JSON(common.NewErrorResponse("'to' and 'presence' are required"))
	}

	// Validate presence type
	validPresences := []string{"typing", "online", "offline", "recording", "paused"}
	isValid := false
	for _, valid := range validPresences {
		if presenceReq.Presence == valid {
			isValid = true
			break
		}
	}

	if !isValid {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid presence type. Valid types: " + strings.Join(validPresences, ", ")))
	}

	// Resolve session
	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

	// Send presence using the real implementation
	err = h.wameowManager.SendPresence(sess.ID.String(), presenceReq.To, presenceReq.Presence)
	if err != nil {
		h.logger.ErrorWithFields("Failed to send presence", map[string]interface{}{
			"session_id": sess.ID.String(),
			"to":         presenceReq.To,
			"presence":   presenceReq.Presence,
			"error":      err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			return c.Status(400).JSON(common.NewErrorResponse("Session is not connected"))
		}

		return c.Status(500).JSON(common.NewErrorResponse("Failed to send presence"))
	}

	response := map[string]interface{}{
		"status":    "sent",
		"presence":  presenceReq.Presence,
		"timestamp": time.Now(),
	}

	return c.JSON(common.NewSuccessResponse(response, "Presence sent successfully"))
}

// EditMessage edits an existing message
// @Summary Edit message
// @Description Edit an existing message
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body messageApp.EditMessageRequest true "Edit message request"
// @Success 200 {object} common.SuccessResponse{data=messageApp.EditResponse} "Message edited successfully"
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 404 {object} common.ErrorResponse "Session not found"
// @Failure 500 {object} common.ErrorResponse "Internal server error"
// @Router /sessions/{sessionId}/messages/edit [post]
func (h *MessageHandler) EditMessage(c *fiber.Ctx) error {
	sessionIdentifier := c.Params("sessionId")
	if sessionIdentifier == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Session identifier is required"))
	}

	// Parse edit request
	var editReq struct {
		To        string `json:"to" validate:"required"`
		MessageID string `json:"messageId" validate:"required"`
		NewBody   string `json:"newBody" validate:"required"`
	}

	if err := c.BodyParser(&editReq); err != nil {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid request body"))
	}

	if editReq.To == "" || editReq.MessageID == "" || editReq.NewBody == "" {
		return c.Status(400).JSON(common.NewErrorResponse("'to', 'messageId', and 'newBody' are required"))
	}

	// Resolve session
	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

	// Edit message using the real implementation
	err = h.wameowManager.EditMessage(sess.ID.String(), editReq.To, editReq.MessageID, editReq.NewBody)
	if err != nil {
		h.logger.ErrorWithFields("Failed to edit message", map[string]interface{}{
			"session_id": sess.ID.String(),
			"to":         editReq.To,
			"message_id": editReq.MessageID,
			"error":      err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			return c.Status(400).JSON(common.NewErrorResponse("Session is not connected"))
		}

		return c.Status(500).JSON(common.NewErrorResponse("Failed to edit message"))
	}

	response := map[string]interface{}{
		"id":        editReq.MessageID,
		"status":    "edited",
		"newBody":   editReq.NewBody,
		"timestamp": time.Now(),
	}

	return c.JSON(common.NewSuccessResponse(response, "Message edited successfully"))
}

// DeleteMessage deletes an existing message
// @Summary Delete message
// @Description Delete an existing message
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body messageApp.DeleteMessageRequest true "Delete message request"
// @Success 200 {object} common.SuccessResponse{data=messageApp.DeleteResponse} "Message deleted successfully"
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 404 {object} common.ErrorResponse "Session not found"
// @Failure 500 {object} common.ErrorResponse "Internal server error"
// @Router /sessions/{sessionId}/messages/delete [post]
func (h *MessageHandler) DeleteMessage(c *fiber.Ctx) error {
	sessionIdentifier := c.Params("sessionId")
	if sessionIdentifier == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Session identifier is required"))
	}

	// Parse delete request
	var deleteReq struct {
		To        string `json:"to" validate:"required"`
		MessageID string `json:"messageId" validate:"required"`
		ForAll    bool   `json:"forAll,omitempty"` // Delete for everyone or just for me
	}

	if err := c.BodyParser(&deleteReq); err != nil {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid request body"))
	}

	if deleteReq.To == "" || deleteReq.MessageID == "" {
		return c.Status(400).JSON(common.NewErrorResponse("'to' and 'messageId' are required"))
	}

	// Resolve session
	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

	// Delete message using the real implementation
	err = h.wameowManager.DeleteMessage(sess.ID.String(), deleteReq.To, deleteReq.MessageID, deleteReq.ForAll)
	if err != nil {
		h.logger.ErrorWithFields("Failed to delete message", map[string]interface{}{
			"session_id": sess.ID.String(),
			"to":         deleteReq.To,
			"message_id": deleteReq.MessageID,
			"error":      err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			return c.Status(400).JSON(common.NewErrorResponse("Session is not connected"))
		}

		return c.Status(500).JSON(common.NewErrorResponse("Failed to delete message"))
	}

	response := map[string]interface{}{
		"id":        deleteReq.MessageID,
		"status":    "deleted",
		"forAll":    deleteReq.ForAll,
		"timestamp": time.Now(),
	}

	return c.JSON(common.NewSuccessResponse(response, "Message deleted successfully"))
}

// sendSpecificMessageType is a helper method to send messages of a specific type
func (h *MessageHandler) sendSpecificMessageType(c *fiber.Ctx, messageType string) error {
	sessionIdentifier := c.Params("sessionId")
	if sessionIdentifier == "" {
		h.logger.Warn("Session identifier is required")
		return c.Status(400).JSON(common.NewErrorResponse("Session identifier is required"))
	}

	// Parse request body
	var req messageApp.SendMessageRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.ErrorWithFields("Failed to parse request body", map[string]interface{}{
			"error": err.Error(),
		})
		return c.Status(400).JSON(common.NewErrorResponse("Invalid request body"))
	}

	// Set the message type
	req.Type = messageType

	// Validate required fields
	if req.To == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Recipient (to) is required"))
	}

	// Type-specific validations
	switch messageType {
	case "text":
		if req.Body == "" {
			return c.Status(400).JSON(common.NewErrorResponse("Body is required for text messages"))
		}
	case "image", "audio", "video", "document", "sticker":
		if req.File == "" {
			return c.Status(400).JSON(common.NewErrorResponse("File is required for " + messageType + " messages"))
		}
		if messageType == "document" && req.Filename == "" {
			return c.Status(400).JSON(common.NewErrorResponse("Filename is required for document messages"))
		}
	case "location":
		if req.Latitude == 0 || req.Longitude == 0 {
			return c.Status(400).JSON(common.NewErrorResponse("Latitude and longitude are required for location messages"))
		}
	case "contact":
		if req.ContactName == "" || req.ContactPhone == "" {
			return c.Status(400).JSON(common.NewErrorResponse("ContactName and contactPhone are required for contact messages"))
		}
	}

	// Resolve session
	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		h.logger.ErrorWithFields("Failed to resolve session", map[string]interface{}{
			"session_identifier": sessionIdentifier,
			"error":              err.Error(),
		})
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

	// Send message
	ctx := c.Context()
	response, err := h.messageUC.SendMessage(ctx, sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to send "+messageType+" message", map[string]interface{}{
			"session_id": sess.ID.String(),
			"to":         req.To,
			"type":       messageType,
			"error":      err.Error(),
		})

		// Check for specific error types
		if strings.Contains(err.Error(), "not connected") {
			return c.Status(400).JSON(common.NewErrorResponse("Session is not connected"))
		}
		if strings.Contains(err.Error(), "not logged in") {
			return c.Status(400).JSON(common.NewErrorResponse("Session is not logged in"))
		}
		if strings.Contains(err.Error(), "invalid request") {
			return c.Status(400).JSON(common.NewErrorResponse(err.Error()))
		}
		if strings.Contains(err.Error(), "failed to process media") {
			return c.Status(400).JSON(common.NewErrorResponse("Failed to process media: " + err.Error()))
		}

		return c.Status(500).JSON(common.NewErrorResponse("Failed to send " + messageType + " message"))
	}

	h.logger.InfoWithFields(strings.Title(messageType)+" message sent successfully", map[string]interface{}{
		"session_id": sess.ID.String(),
		"to":         req.To,
		"type":       messageType,
		"message_id": response.ID,
	})

	return c.JSON(common.NewSuccessResponse(response, strings.Title(messageType)+" message sent successfully"))
}
