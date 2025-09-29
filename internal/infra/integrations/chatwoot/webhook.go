package chatwoot

import (
	"context"
	"fmt"
	"time"

	chatwootdomain "zpwoot/internal/domain/chatwoot"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

// WebhookHandler implements the WebhookHandler interface
type WebhookHandler struct {
	logger        *logger.Logger
	manager       ports.ChatwootManager
	wameowManager ports.WameowManager // For sending messages to WhatsApp
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(logger *logger.Logger, manager ports.ChatwootManager, wameowManager ports.WameowManager) *WebhookHandler {
	return &WebhookHandler{
		logger:        logger,
		manager:       manager,
		wameowManager: wameowManager,
	}
}

// ProcessWebhook processes incoming webhooks from Chatwoot
func (h *WebhookHandler) ProcessWebhook(ctx context.Context, webhook *chatwootdomain.ChatwootWebhookPayload, sessionID string) error {
	// Delay 500ms to avoid race conditions (based on Evolution API)
	time.Sleep(500 * time.Millisecond)

	// Filter private messages
	if h.isPrivateMessage(webhook) {
		return nil
	}

	// Filter message updates without deletion
	if webhook.Event == "message_updated" && !h.isMessageDeleted(webhook) {
		return nil
	}

	// Process conversation status changes
	if webhook.Event == "conversation_status_changed" {
		return h.handleConversationStatusChanged(ctx, webhook, sessionID)
	}

	// Process message deletions
	if webhook.Event == "message_updated" && h.isMessageDeleted(webhook) {
		return h.handleMessageDeleted(ctx, webhook, sessionID)
	}

	// Process new messages
	if webhook.Event == "message_created" {
		return h.handleMessageCreated(ctx, webhook, sessionID)
	}

	h.logger.DebugWithFields("Unhandled webhook event", map[string]interface{}{
		"event":      webhook.Event,
		"session_id": sessionID,
	})

	return nil
}

// handleConversationStatusChanged handles conversation status changes
func (h *WebhookHandler) handleConversationStatusChanged(ctx context.Context, webhook *chatwootdomain.ChatwootWebhookPayload, sessionID string) error {
	if webhook.Conversation.Status == "resolved" {
		h.logger.InfoWithFields("Conversation resolved", map[string]interface{}{
			"conversation_id": webhook.Conversation.ID,
			"session_id":      sessionID,
		})

		// TODO: Implement conversation resolution logic
		// This could involve updating local state or sending notifications
	}

	return nil
}

// handleMessageDeleted handles deleted messages
func (h *WebhookHandler) handleMessageDeleted(ctx context.Context, webhook *chatwootdomain.ChatwootWebhookPayload, sessionID string) error {
	h.logger.InfoWithFields("Message deleted", map[string]interface{}{
		"message_id": webhook.Message.ID,
		"session_id": sessionID,
	})

	// TODO: Implement message deletion logic
	// This could involve deleting the message from WhatsApp if supported

	return nil
}

// handleMessageCreated handles new messages from Chatwoot
func (h *WebhookHandler) handleMessageCreated(ctx context.Context, webhook *chatwootdomain.ChatwootWebhookPayload, sessionID string) error {
	if webhook.Message == nil {
		return fmt.Errorf("message is nil in webhook payload")
	}

	// Skip bot messages
	if h.isBotMessage(webhook) {
		h.logger.DebugWithFields("Skipping bot message", map[string]interface{}{
			"message_id": webhook.Message.ID,
			"session_id": sessionID,
		})
		return nil
	}

	// Skip incoming messages (only process outgoing messages from agents)
	if webhook.Message.MessageType != "outgoing" {
		return nil
	}

	return h.sendToWhatsApp(ctx, webhook, sessionID)
}

// sendToWhatsApp sends a message from Chatwoot to WhatsApp
func (h *WebhookHandler) sendToWhatsApp(ctx context.Context, webhook *chatwootdomain.ChatwootWebhookPayload, sessionID string) error {
	h.logger.InfoWithFields("Sending message to WhatsApp", map[string]interface{}{
		"message_id":      webhook.Message.ID,
		"conversation_id": webhook.Conversation.ID,
		"session_id":      sessionID,
	})

	// Extract phone number from contact
	phoneNumber := webhook.Contact.PhoneNumber
	if phoneNumber == "" {
		return fmt.Errorf("contact phone number is empty")
	}

	// Format content for WhatsApp (convert Chatwoot markdown to WhatsApp format)
	content := h.formatContentForWhatsApp(webhook.Message.Content)

	// Send message to WhatsApp using wameowManager
	_, err := h.wameowManager.SendMessage(sessionID, phoneNumber, "text", content, "", "", "", 0, 0, "", "", nil)
	if err != nil {
		return fmt.Errorf("failed to send message to WhatsApp: %w", err)
	}

	return nil
}

// formatContentForWhatsApp formats message content for WhatsApp
func (h *WebhookHandler) formatContentForWhatsApp(content string) string {
	// TODO: Implement markdown conversion from Chatwoot to WhatsApp format
	// Chatwoot uses standard markdown, WhatsApp uses different formatting
	// ** -> *
	// * -> _
	// ~~ -> ~
	return content
}

// isPrivateMessage checks if the message is private
func (h *WebhookHandler) isPrivateMessage(webhook *chatwootdomain.ChatwootWebhookPayload) bool {
	if webhook.Message == nil {
		return false
	}
	return webhook.Message.Private
}

// isMessageDeleted checks if the message was deleted
func (h *WebhookHandler) isMessageDeleted(webhook *chatwootdomain.ChatwootWebhookPayload) bool {
	if webhook.Message == nil {
		return false
	}

	// Check content_attributes for deleted flag
	if webhook.Message.ContentAttributes != nil {
		if deleted, exists := webhook.Message.ContentAttributes["deleted"]; exists && deleted != nil {
			return true
		}
	}

	return false
}

// isBotMessage checks if the message is from a bot
func (h *WebhookHandler) isBotMessage(webhook *chatwootdomain.ChatwootWebhookPayload) bool {
	if webhook.Message == nil || webhook.Message.Sender == nil {
		return false
	}

	// Check if sender type is agent_bot
	if webhook.Message.Sender.Type == "agent_bot" {
		return true
	}

	// Check if sender identifier is the bot contact (123456)
	if webhook.Message.Sender.Identifier == "123456" {
		return true
	}

	return false
}

// extractPhoneFromContact extracts phone number from contact
func (h *WebhookHandler) extractPhoneFromContact(contact chatwootdomain.ChatwootContact) string {
	// Try phone_number field first
	if contact.PhoneNumber != "" {
		return contact.PhoneNumber
	}

	// Try identifier field
	if contact.Identifier != "" {
		return contact.Identifier
	}

	// Try additional_attributes
	if contact.AdditionalAttributes != nil {
		if phone, exists := contact.AdditionalAttributes["phone_number"]; exists {
			if phoneStr, ok := phone.(string); ok {
				return phoneStr
			}
		}
	}

	return ""
}

// normalizePhoneNumber normalizes phone number for WhatsApp
func (h *WebhookHandler) normalizePhoneNumber(phone string) string {
	// TODO: Implement phone number normalization
	// Remove special characters, add country code if needed, etc.
	return phone
}
