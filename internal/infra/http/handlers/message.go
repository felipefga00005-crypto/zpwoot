package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"zpwoot/internal/app/common"
	"zpwoot/internal/app/message"
	messageApp "zpwoot/internal/app/message"
	"zpwoot/internal/infra/http/helpers"
	"zpwoot/internal/infra/wameow"
	"zpwoot/platform/logger"
)

type MessageHandler struct {
	messageUC       messageApp.UseCase
	wameowManager   *wameow.Manager
	sessionResolver *helpers.SessionResolver
	logger          *logger.Logger
}

func NewMessageHandler(
	messageUC messageApp.UseCase,
	wameowManager *wameow.Manager,
	sessionRepo helpers.SessionRepository,
	logger *logger.Logger,
) *MessageHandler {
	sessionResolver := helpers.NewSessionResolver(logger, sessionRepo)

	return &MessageHandler{
		messageUC:       messageUC,
		wameowManager:   wameowManager,
		sessionResolver: sessionResolver,
		logger:          logger,
	}
}

// @Summary Send media message
// @Description Send a media file (image, audio, video, document) with optional caption
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body message.MediaMessageRequest true "Media message request"
// @Success 200 {object} message.MessageResponse "Media message sent successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/messages/send/media [post]
func (h *MessageHandler) SendMedia(c *fiber.Ctx) error {
	sessionIdentifier := c.Params("sessionId")
	if sessionIdentifier == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Session identifier is required"))
	}

	var mediaReq message.MediaMessageRequest
	if err := c.BodyParser(&mediaReq); err != nil {
		h.logger.ErrorWithFields("Failed to parse media request", map[string]interface{}{
			"error": err.Error(),
		})
		return c.Status(400).JSON(common.NewErrorResponse("Invalid request body"))
	}

	if mediaReq.To == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Recipient (to) is required"))
	}

	if mediaReq.File == "" {
		return c.Status(400).JSON(common.NewErrorResponse("File is required"))
	}

	// Convert MediaMessageRequest to SendMessageRequest
	req := &message.SendMessageRequest{
		To:       mediaReq.To,
		Type:     "media", // Generic media type
		File:     mediaReq.File,
		Caption:  mediaReq.Caption,
		MimeType: mediaReq.MimeType,
		Filename: mediaReq.Filename,
	}

	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		h.logger.ErrorWithFields("Failed to resolve session", map[string]interface{}{
			"session_identifier": sessionIdentifier,
			"error":              err.Error(),
		})
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

	ctx := c.Context()
	response, err := h.messageUC.SendMessage(ctx, sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to send media message", map[string]interface{}{
			"session_id": sess.ID.String(),
			"to":         req.To,
			"error":      err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			return c.Status(400).JSON(common.NewErrorResponse("Session is not connected"))
		}
		if strings.Contains(err.Error(), "not logged in") {
			return c.Status(400).JSON(common.NewErrorResponse("Session is not logged in"))
		}
		if strings.Contains(err.Error(), "failed to process media") {
			return c.Status(400).JSON(common.NewErrorResponse("Failed to process media: " + err.Error()))
		}

		return c.Status(500).JSON(common.NewErrorResponse("Failed to send media message"))
	}

	h.logger.InfoWithFields("Media message sent successfully", map[string]interface{}{
		"session_id": sess.ID.String(),
		"to":         req.To,
		"message_id": response.ID,
	})

	return c.JSON(common.NewSuccessResponse(response, "Media message sent successfully"))
}

// @Summary Send image message
// @Description Send an image message through WhatsApp with optional reply context
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body message.ImageMessageRequest true "Image message request"
// @Success 200 {object} common.SuccessResponse{data=message.SendMessageResponse} "Message sent successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/send/image [post]
func (h *MessageHandler) SendImage(c *fiber.Ctx) error {
	sessionIdentifier := c.Params("sessionId")
	if sessionIdentifier == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Session identifier is required"))
	}

	var imageReq message.ImageMessageRequest
	if err := c.BodyParser(&imageReq); err != nil {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid image message format"))
	}

	if imageReq.To == "" {
		return c.Status(400).JSON(common.NewErrorResponse("'to' field is required"))
	}

	if imageReq.File == "" {
		return c.Status(400).JSON(common.NewErrorResponse("'file' field is required"))
	}

	if imageReq.ContextInfo != nil {
		if imageReq.ContextInfo.StanzaID == "" {
			return c.Status(400).JSON(common.NewErrorResponse("'contextInfo.stanzaId' is required when replying"))
		}
	}

	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

	// Convert to SendMessageRequest for compatibility
	req := message.SendMessageRequest{
		To:          imageReq.To,
		Type:        "image",
		File:        imageReq.File,
		Caption:     imageReq.Caption,
		MimeType:    imageReq.MimeType,
		Filename:    imageReq.Filename,
		ContextInfo: imageReq.ContextInfo,
	}

	ctx := c.Context()
	response, err := h.messageUC.SendMessage(ctx, sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to send image message", map[string]interface{}{
			"session_id": sess.ID.String(),
			"to":         imageReq.To,
			"has_reply":  imageReq.ContextInfo != nil,
			"error":      err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			return c.Status(400).JSON(common.NewErrorResponse("Session is not connected"))
		}

		return c.Status(500).JSON(common.NewErrorResponse("Failed to send image message"))
	}

	return c.JSON(common.NewSuccessResponse(response, "Image message sent successfully"))
}

// @Summary Send audio message
// @Description Send an audio message through WhatsApp with optional reply context
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body message.AudioMessageRequest true "Audio message request"
// @Success 200 {object} common.SuccessResponse{data=message.SendMessageResponse} "Message sent successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/send/audio [post]
func (h *MessageHandler) SendAudio(c *fiber.Ctx) error {
	sessionIdentifier := c.Params("sessionId")
	if sessionIdentifier == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Session identifier is required"))
	}

	var audioReq message.AudioMessageRequest
	if err := c.BodyParser(&audioReq); err != nil {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid audio message format"))
	}

	if audioReq.To == "" {
		return c.Status(400).JSON(common.NewErrorResponse("'to' field is required"))
	}

	if audioReq.File == "" {
		return c.Status(400).JSON(common.NewErrorResponse("'file' field is required"))
	}

	if audioReq.ContextInfo != nil {
		if audioReq.ContextInfo.StanzaID == "" {
			return c.Status(400).JSON(common.NewErrorResponse("'contextInfo.stanzaId' is required when replying"))
		}
	}

	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

	// Convert to SendMessageRequest for compatibility
	req := message.SendMessageRequest{
		To:          audioReq.To,
		Type:        "audio",
		File:        audioReq.File,
		Caption:     audioReq.Caption,
		MimeType:    audioReq.MimeType,
		ContextInfo: audioReq.ContextInfo,
	}

	ctx := c.Context()
	response, err := h.messageUC.SendMessage(ctx, sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to send audio message", map[string]interface{}{
			"session_id": sess.ID.String(),
			"to":         audioReq.To,
			"has_reply":  audioReq.ContextInfo != nil,
			"error":      err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			return c.Status(400).JSON(common.NewErrorResponse("Session is not connected"))
		}

		return c.Status(500).JSON(common.NewErrorResponse("Failed to send audio message"))
	}

	return c.JSON(common.NewSuccessResponse(response, "Audio message sent successfully"))
}

// @Summary Send video message
// @Description Send a video message through WhatsApp with optional reply context
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body message.VideoMessageRequest true "Video message request"
// @Success 200 {object} common.SuccessResponse{data=message.SendMessageResponse} "Message sent successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/send/video [post]
func (h *MessageHandler) SendVideo(c *fiber.Ctx) error {
	sessionIdentifier := c.Params("sessionId")
	if sessionIdentifier == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Session identifier is required"))
	}

	var videoReq message.VideoMessageRequest
	if err := c.BodyParser(&videoReq); err != nil {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid video message format"))
	}

	if videoReq.To == "" {
		return c.Status(400).JSON(common.NewErrorResponse("'to' field is required"))
	}

	if videoReq.File == "" {
		return c.Status(400).JSON(common.NewErrorResponse("'file' field is required"))
	}

	if videoReq.ContextInfo != nil {
		if videoReq.ContextInfo.StanzaID == "" {
			return c.Status(400).JSON(common.NewErrorResponse("'contextInfo.stanzaId' is required when replying"))
		}
	}

	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

	// Convert to SendMessageRequest for compatibility
	req := message.SendMessageRequest{
		To:          videoReq.To,
		Type:        "video",
		File:        videoReq.File,
		Caption:     videoReq.Caption,
		MimeType:    videoReq.MimeType,
		Filename:    videoReq.Filename,
		ContextInfo: videoReq.ContextInfo,
	}

	ctx := c.Context()
	response, err := h.messageUC.SendMessage(ctx, sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to send video message", map[string]interface{}{
			"session_id": sess.ID.String(),
			"to":         videoReq.To,
			"has_reply":  videoReq.ContextInfo != nil,
			"error":      err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			return c.Status(400).JSON(common.NewErrorResponse("Session is not connected"))
		}

		return c.Status(500).JSON(common.NewErrorResponse("Failed to send video message"))
	}

	return c.JSON(common.NewSuccessResponse(response, "Video message sent successfully"))
}

// @Summary Send document message
// @Description Send a document message through WhatsApp with optional reply context
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body message.DocumentMessageRequest true "Document message request"
// @Success 200 {object} common.SuccessResponse{data=message.SendMessageResponse} "Message sent successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/send/document [post]
func (h *MessageHandler) SendDocument(c *fiber.Ctx) error {
	sessionIdentifier := c.Params("sessionId")
	if sessionIdentifier == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Session identifier is required"))
	}

	var docReq message.DocumentMessageRequest
	if err := c.BodyParser(&docReq); err != nil {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid document message format"))
	}

	if docReq.To == "" {
		return c.Status(400).JSON(common.NewErrorResponse("'to' field is required"))
	}

	if docReq.File == "" {
		return c.Status(400).JSON(common.NewErrorResponse("'file' field is required"))
	}

	if docReq.Filename == "" {
		return c.Status(400).JSON(common.NewErrorResponse("'filename' field is required"))
	}

	if docReq.ContextInfo != nil {
		if docReq.ContextInfo.StanzaID == "" {
			return c.Status(400).JSON(common.NewErrorResponse("'contextInfo.stanzaId' is required when replying"))
		}
	}

	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

	// Convert to SendMessageRequest for compatibility
	req := message.SendMessageRequest{
		To:          docReq.To,
		Type:        "document",
		File:        docReq.File,
		Caption:     docReq.Caption,
		MimeType:    docReq.MimeType,
		Filename:    docReq.Filename,
		ContextInfo: docReq.ContextInfo,
	}

	ctx := c.Context()
	response, err := h.messageUC.SendMessage(ctx, sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to send document message", map[string]interface{}{
			"session_id": sess.ID.String(),
			"to":         docReq.To,
			"filename":   docReq.Filename,
			"has_reply":  docReq.ContextInfo != nil,
			"error":      err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			return c.Status(400).JSON(common.NewErrorResponse("Session is not connected"))
		}

		return c.Status(500).JSON(common.NewErrorResponse("Failed to send document message"))
	}

	return c.JSON(common.NewSuccessResponse(response, "Document message sent successfully"))
}

// @Summary Send sticker message
// @Description Send a sticker message through WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body message.MediaMessageRequest true "Sticker message request"
// @Success 200 {object} common.SuccessResponse{data=message.SendMessageResponse} "Message sent successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/send/sticker [post]
func (h *MessageHandler) SendSticker(c *fiber.Ctx) error {
	return h.sendSpecificMessageType(c, "sticker")
}

// @Summary Send location message
// @Description Send a location message through WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body message.LocationMessageRequest true "Location message request"
// @Success 200 {object} common.SuccessResponse{data=message.SendMessageResponse} "Message sent successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/send/location [post]
func (h *MessageHandler) SendLocation(c *fiber.Ctx) error {
	return h.sendSpecificMessageType(c, "location")
}

// @Summary Send contact message(s)
// @Description Send a single contact or multiple contacts through WhatsApp. Automatically detects if it's a single contact (ContactMessage) or multiple contacts (ContactsArrayMessage) based on the array length.
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body message.ContactMessageRequest true "Contact message request (single contact) or ContactListMessageRequest (multiple contacts)"
// @Success 200 {object} common.SuccessResponse{data=message.SendMessageResponse} "Single contact sent successfully"
// @Success 200 {object} common.SuccessResponse{data=message.ContactListMessageResponse} "Contact list sent successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/send/contact [post]
func (h *MessageHandler) SendContact(c *fiber.Ctx) error {
	sessionIdentifier := c.Params("sessionId")
	if sessionIdentifier == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Session identifier is required"))
	}

	var rawBody map[string]interface{}
	if err := c.BodyParser(&rawBody); err != nil {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid request body"))
	}

	if _, hasContacts := rawBody["contacts"]; hasContacts {
		return h.handleContactList(c, sessionIdentifier)
	} else if _, hasContactName := rawBody["contactName"]; hasContactName {
		return h.handleSingleContact(c, sessionIdentifier)
	} else {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid contact format. Use either single contact format (contactName, contactPhone) or contact list format (contacts array)"))
	}
}

func (h *MessageHandler) handleSingleContact(c *fiber.Ctx, sessionIdentifier string) error {
	var contactReq message.ContactMessageRequest
	if err := c.BodyParser(&contactReq); err != nil {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid single contact format"))
	}

	if contactReq.To == "" {
		return c.Status(400).JSON(common.NewErrorResponse("'to' field is required"))
	}
	if contactReq.ContactName == "" {
		return c.Status(400).JSON(common.NewErrorResponse("'contactName' field is required"))
	}
	if contactReq.ContactPhone == "" {
		return c.Status(400).JSON(common.NewErrorResponse("'contactPhone' field is required"))
	}

	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

	result, err := h.wameowManager.SendMessage(
		sess.ID.String(),
		contactReq.To,
		"contact",
		"",
		"",
		"",
		"",
		0,
		0,
		contactReq.ContactName,
		contactReq.ContactPhone,
	)

	if err != nil {
		h.logger.ErrorWithFields("Failed to send contact message", map[string]interface{}{
			"session_id":   sess.ID.String(),
			"to":           contactReq.To,
			"contact_name": contactReq.ContactName,
			"error":        err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			return c.Status(400).JSON(common.NewErrorResponse("Session is not connected"))
		}

		return c.Status(500).JSON(common.NewErrorResponse("Failed to send contact message"))
	}

	response := &message.SendMessageResponse{
		ID:        result.MessageID,
		Status:    result.Status,
		Timestamp: result.Timestamp,
	}

	h.logger.InfoWithFields("Contact message sent successfully", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"to":           contactReq.To,
		"contact_name": contactReq.ContactName,
		"message_id":   result.MessageID,
	})

	return c.JSON(common.NewSuccessResponse(response, "Contact message sent successfully"))
}

func (h *MessageHandler) handleContactList(c *fiber.Ctx, sessionIdentifier string) error {
	var contactListReq message.ContactListMessageRequest
	if err := c.BodyParser(&contactListReq); err != nil {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid contact list format"))
	}

	if contactListReq.To == "" {
		return c.Status(400).JSON(common.NewErrorResponse("'to' field is required"))
	}

	if len(contactListReq.Contacts) == 0 {
		return c.Status(400).JSON(common.NewErrorResponse("At least one contact is required"))
	}

	if len(contactListReq.Contacts) > 10 {
		return c.Status(400).JSON(common.NewErrorResponse("Maximum 10 contacts allowed per request"))
	}

	for i, contact := range contactListReq.Contacts {
		if contact.Name == "" {
			return c.Status(400).JSON(common.NewErrorResponse(fmt.Sprintf("Contact %d: name is required", i+1)))
		}
		if contact.Phone == "" {
			return c.Status(400).JSON(common.NewErrorResponse(fmt.Sprintf("Contact %d: phone is required", i+1)))
		}
	}

	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

	var contacts []wameow.ContactInfo
	for _, contact := range contactListReq.Contacts {
		contacts = append(contacts, wameow.ContactInfo{
			Name:         contact.Name,
			Phone:        contact.Phone,
			Email:        contact.Email,
			Organization: contact.Organization,
			Title:        contact.Title,
			Website:      contact.Website,
			Address:      contact.Address,
		})
	}

	var result *wameow.ContactListResult

	if len(contactListReq.Contacts) == 1 {
		contact := contacts[0]

		result, err = h.wameowManager.SendSingleContact(sess.ID.String(), contactListReq.To, contact)
		if err != nil {
			h.logger.ErrorWithFields("Failed to send single contact", map[string]interface{}{
				"session_id":   sess.ID.String(),
				"to":           contactListReq.To,
				"contact_name": contact.Name,
				"error":        err.Error(),
			})
		}
	} else {
		result, err = h.wameowManager.SendContactList(sess.ID.String(), contactListReq.To, contacts)
		if err != nil {
			h.logger.ErrorWithFields("Failed to send contact list", map[string]interface{}{
				"session_id":    sess.ID.String(),
				"to":            contactListReq.To,
				"contact_count": len(contactListReq.Contacts),
				"error":         err.Error(),
			})
		}
	}

	if err != nil {
		if strings.Contains(err.Error(), "not connected") {
			return c.Status(400).JSON(common.NewErrorResponse("Session is not connected"))
		}

		return c.Status(500).JSON(common.NewErrorResponse("Failed to send contact list"))
	}

	var contactResults []message.ContactSendResult
	for _, r := range result.Results {
		contactResults = append(contactResults, message.ContactSendResult{
			ContactName: r.ContactName,
			MessageID:   r.MessageID,
			Status:      r.Status,
			Error:       r.Error,
		})
	}

	response := &message.ContactListMessageResponse{
		TotalContacts: result.TotalContacts,
		SuccessCount:  result.SuccessCount,
		FailureCount:  result.FailureCount,
		Results:       contactResults,
		Timestamp:     result.Timestamp.Format(time.RFC3339),
	}

	contactType := "single contact"
	if len(contactListReq.Contacts) > 1 {
		contactType = "contact list"
	}

	h.logger.InfoWithFields("Contact sent successfully", map[string]interface{}{
		"session_id":     sess.ID.String(),
		"to":             contactListReq.To,
		"total_contacts": result.TotalContacts,
		"success_count":  result.SuccessCount,
		"failure_count":  result.FailureCount,
		"contact_type":   contactType,
		"format_type":    "standard",
	})

	successMessage := "Contact sent successfully"
	if len(contactListReq.Contacts) > 1 {
		successMessage = "Contact list sent successfully"
	}

	return c.JSON(common.NewSuccessResponse(response, successMessage))
}

// @Summary Send business profile contact
// @Description Send a business profile contact using Business format (with waid, X-ABLabel, X-WA-BIZ-NAME)
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body message.BusinessProfileRequest true "Business profile request"
// @Success 200 {object} common.SuccessResponse{data=message.SendMessageResponse} "Business profile sent successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/send/profile/business [post]
func (h *MessageHandler) SendBusinessProfile(c *fiber.Ctx) error {
	sessionIdentifier := c.Params("sessionId")
	if sessionIdentifier == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Session identifier is required"))
	}

	var businessReq message.BusinessProfileRequest
	if err := c.BodyParser(&businessReq); err != nil {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid business profile format"))
	}

	if businessReq.To == "" {
		return c.Status(400).JSON(common.NewErrorResponse("'to' field is required"))
	}

	if businessReq.Name == "" {
		return c.Status(400).JSON(common.NewErrorResponse("'name' field is required"))
	}

	if businessReq.Phone == "" {
		return c.Status(400).JSON(common.NewErrorResponse("'phone' field is required"))
	}

	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

	contact := wameow.ContactInfo{
		Name:         businessReq.Name,
		Phone:        businessReq.Phone,
		Email:        businessReq.Email,
		Organization: businessReq.Organization,
		Title:        businessReq.Title,
		Website:      businessReq.Website,
		Address:      businessReq.Address,
	}

	result, err := h.wameowManager.SendSingleContactBusinessFormat(sess.ID.String(), businessReq.To, contact)
	if err != nil {
		h.logger.ErrorWithFields("Failed to send business profile", map[string]interface{}{
			"session_id":    sess.ID.String(),
			"to":            businessReq.To,
			"business_name": businessReq.Name,
			"error":         err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			return c.Status(400).JSON(common.NewErrorResponse("Session is not connected"))
		}

		return c.Status(500).JSON(common.NewErrorResponse("Failed to send business profile"))
	}

	h.logger.InfoWithFields("Business profile sent successfully", map[string]interface{}{
		"session_id":    sess.ID.String(),
		"to":            businessReq.To,
		"business_name": businessReq.Name,
		"format_type":   "Business",
	})

	response := message.SendMessageResponse{
		ID:        result.Results[0].MessageID,
		Status:    result.Results[0].Status,
		Timestamp: result.Timestamp,
	}

	return c.Status(200).JSON(common.NewSuccessResponse(response, "Business profile sent successfully"))
}

// @Summary Send text message
// @Description Send a text message with optional context info for replies
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body message.TextMessageRequest true "Text message request"
// @Success 200 {object} message.MessageResponse "Text message sent successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/messages/send/text [post]
func (h *MessageHandler) SendText(c *fiber.Ctx) error {
	sessionIdentifier := c.Params("sessionId")
	if sessionIdentifier == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Session identifier is required"))
	}

	var textReq message.TextMessageRequest
	if err := c.BodyParser(&textReq); err != nil {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid text message format"))
	}

	if textReq.To == "" {
		return c.Status(400).JSON(common.NewErrorResponse("'to' field is required"))
	}

	if textReq.Body == "" {
		return c.Status(400).JSON(common.NewErrorResponse("'body' field is required"))
	}

	if textReq.ContextInfo != nil {
		if textReq.ContextInfo.StanzaID == "" {
			return c.Status(400).JSON(common.NewErrorResponse("'contextInfo.stanzaId' is required when replying"))
		}
	}

	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

	result, err := h.wameowManager.SendTextMessage(sess.ID.String(), textReq.To, textReq.Body, textReq.ContextInfo)
	if err != nil {
		h.logger.ErrorWithFields("Failed to send text message", map[string]interface{}{
			"session_id": sess.ID.String(),
			"to":         textReq.To,
			"has_reply":  textReq.ContextInfo != nil,
			"error":      err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			return c.Status(400).JSON(common.NewErrorResponse("Session is not connected"))
		}

		return c.Status(500).JSON(common.NewErrorResponse("Failed to send text message"))
	}

	h.logger.InfoWithFields("Text message sent successfully", map[string]interface{}{
		"session_id": sess.ID.String(),
		"to":         textReq.To,
		"message_id": result.MessageID,
		"has_reply":  textReq.ContextInfo != nil,
	})

	response := message.SendMessageResponse{
		ID:        result.MessageID,
		Status:    result.Status,
		Timestamp: result.Timestamp,
	}

	return c.Status(200).JSON(common.NewSuccessResponse(response, "Text message sent successfully"))
}

// @Summary Send button message
// @Description Send a message with interactive buttons through WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body message.ButtonMessageRequest true "Button message request"
// @Success 200 {object} common.SuccessResponse{data=message.MessageResponse} "Button message sent successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/send/button [post]
func (h *MessageHandler) SendButtonMessage(c *fiber.Ctx) error {
	sessionIdentifier := c.Params("sessionId")
	if sessionIdentifier == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Session identifier is required"))
	}

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

	var buttons []map[string]string
	for _, button := range buttonReq.Buttons {
		buttons = append(buttons, map[string]string{
			"id":   button.ID,
			"text": button.Text,
		})
	}

	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

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

	response := &message.SendMessageResponse{
		ID:        result.MessageID,
		Status:    result.Status,
		Timestamp: result.Timestamp,
	}

	return c.JSON(common.NewSuccessResponse(response, "Button message sent successfully"))
}

// @Summary Send list message
// @Description Send a message with interactive list through WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body message.ListMessageRequest true "List message request"
// @Success 200 {object} common.SuccessResponse{data=message.MessageResponse} "List message sent successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/send/list [post]
func (h *MessageHandler) SendListMessage(c *fiber.Ctx) error {
	sessionIdentifier := c.Params("sessionId")
	if sessionIdentifier == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Session identifier is required"))
	}

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

	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

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

	response := &message.SendMessageResponse{
		ID:        result.MessageID,
		Status:    result.Status,
		Timestamp: result.Timestamp,
	}

	return c.JSON(common.NewSuccessResponse(response, "List message sent successfully"))
}

// @Summary Send reaction
// @Description Send a reaction (emoji) to a specific message
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body message.ReactionMessageRequest true "Reaction request"
// @Success 200 {object} common.SuccessResponse{data=message.ReactionResponse} "Reaction sent successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/send/reaction [post]
func (h *MessageHandler) SendReaction(c *fiber.Ctx) error {
	sessionIdentifier := c.Params("sessionId")
	if sessionIdentifier == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Session identifier is required"))
	}

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

	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

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

// @Summary Send presence
// @Description Send presence information (typing, online, etc.)
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body message.PresenceMessageRequest true "Presence request"
// @Success 200 {object} common.SuccessResponse{data=message.PresenceResponse} "Presence sent successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/send/presence [post]
func (h *MessageHandler) SendPresence(c *fiber.Ctx) error {
	sessionIdentifier := c.Params("sessionId")
	if sessionIdentifier == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Session identifier is required"))
	}

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

	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

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

// @Summary Edit message
// @Description Edit an existing message
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body message.EditMessageRequest true "Edit message request"
// @Success 200 {object} common.SuccessResponse{data=message.EditResponse} "Message edited successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/edit [post]
func (h *MessageHandler) EditMessage(c *fiber.Ctx) error {
	sessionIdentifier := c.Params("sessionId")
	if sessionIdentifier == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Session identifier is required"))
	}

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

	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

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

// @Summary Delete message
// @Description Delete an existing message
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body message.DeleteMessageRequest true "Delete message request"
// @Success 200 {object} common.SuccessResponse{data=message.DeleteResponse} "Message deleted successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/delete [post]
func (h *MessageHandler) DeleteMessage(c *fiber.Ctx) error {
	sessionIdentifier := c.Params("sessionId")
	if sessionIdentifier == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Session identifier is required"))
	}

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

	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

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

// @Summary Mark message as read
// @Description Mark a specific message as read
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body message.MarkReadRequest true "Mark as read request"
// @Success 200 {object} common.SuccessResponse{data=message.MarkReadResponse} "Message marked as read successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/mark-read [post]
func (h *MessageHandler) MarkAsRead(c *fiber.Ctx) error {
	sessionIdentifier := c.Params("sessionId")
	if sessionIdentifier == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Session identifier is required"))
	}

	var markReadReq struct {
		To        string `json:"to" validate:"required"`
		MessageID string `json:"messageId" validate:"required"`
	}

	if err := c.BodyParser(&markReadReq); err != nil {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid request body"))
	}

	if markReadReq.To == "" || markReadReq.MessageID == "" {
		return c.Status(400).JSON(common.NewErrorResponse("'to' and 'messageId' are required"))
	}

	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

	err = h.wameowManager.MarkRead(sess.ID.String(), markReadReq.To, markReadReq.MessageID)
	if err != nil {
		h.logger.ErrorWithFields("Failed to mark message as read", map[string]interface{}{
			"session_id": sess.ID.String(),
			"to":         markReadReq.To,
			"message_id": markReadReq.MessageID,
			"error":      err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			return c.Status(400).JSON(common.NewErrorResponse("Session is not connected"))
		}

		return c.Status(500).JSON(common.NewErrorResponse("Failed to mark message as read"))
	}

	response := map[string]interface{}{
		"messageId": markReadReq.MessageID,
		"status":    "read",
		"timestamp": time.Now(),
	}

	return c.JSON(common.NewSuccessResponse(response, "Message marked as read successfully"))
}

func (h *MessageHandler) sendSpecificMessageType(c *fiber.Ctx, messageType string) error {
	sessionIdentifier := c.Params("sessionId")
	if sessionIdentifier == "" {
		h.logger.Warn("Session identifier is required")
		return c.Status(400).JSON(common.NewErrorResponse("Session identifier is required"))
	}

	var req message.SendMessageRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.ErrorWithFields("Failed to parse request body", map[string]interface{}{
			"error": err.Error(),
		})
		return c.Status(400).JSON(common.NewErrorResponse("Invalid request body"))
	}

	req.Type = messageType

	if req.To == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Recipient (to) is required"))
	}

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

	sess, err := h.sessionResolver.ResolveSession(c.Context(), sessionIdentifier)
	if err != nil {
		h.logger.ErrorWithFields("Failed to resolve session", map[string]interface{}{
			"session_identifier": sessionIdentifier,
			"error":              err.Error(),
		})
		return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
	}

	ctx := c.Context()
	response, err := h.messageUC.SendMessage(ctx, sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to send "+messageType+" message", map[string]interface{}{
			"session_id": sess.ID.String(),
			"to":         req.To,
			"type":       messageType,
			"error":      err.Error(),
		})

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
