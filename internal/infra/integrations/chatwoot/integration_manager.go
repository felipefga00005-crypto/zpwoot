package chatwoot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

// IntegrationManager handles the integration between WhatsApp and Chatwoot
type IntegrationManager struct {
	logger          *logger.Logger
	chatwootManager ports.ChatwootManager
	messageMapper   *MessageMapper
	contactSync     *ContactSync
	conversationMgr *ConversationManager
	formatter       *MessageFormatter
}

// NewIntegrationManager creates a new integration manager
func NewIntegrationManager(
	logger *logger.Logger,
	chatwootManager ports.ChatwootManager,
	messageMapper *MessageMapper,
	contactSync *ContactSync,
	conversationMgr *ConversationManager,
	formatter *MessageFormatter,
) *IntegrationManager {
	return &IntegrationManager{
		logger:          logger,
		chatwootManager: chatwootManager,
		messageMapper:   messageMapper,
		contactSync:     contactSync,
		conversationMgr: conversationMgr,
		formatter:       formatter,
	}
}

// IsEnabled checks if Chatwoot integration is enabled for a session
func (im *IntegrationManager) IsEnabled(sessionID string) bool {
	return im.chatwootManager.IsEnabled(sessionID)
}

// ProcessWhatsAppMessage processes a WhatsApp message for Chatwoot integration
func (im *IntegrationManager) ProcessWhatsAppMessage(sessionID, messageID, from, content, messageType string, timestamp time.Time, fromMe bool) error {
	ctx := context.Background()

	im.logger.InfoWithFields("Processing WhatsApp message for Chatwoot", map[string]interface{}{
		"session_id":   sessionID,
		"message_id":   messageID,
		"from":         from,
		"message_type": messageType,
		"from_me":      fromMe,
	})

	// Skip if message is already mapped
	if im.messageMapper.IsMessageMapped(ctx, sessionID, messageID) {
		im.logger.DebugWithFields("Message already mapped, skipping", map[string]interface{}{
			"session_id": sessionID,
			"message_id": messageID,
		})
		return nil
	}

	// Create initial mapping
	_, err := im.messageMapper.CreateMapping(ctx, sessionID, messageID)
	if err != nil {
		return fmt.Errorf("failed to create message mapping: %w", err)
	}

	// Get Chatwoot client
	client, err := im.chatwootManager.GetClient(sessionID)
	if err != nil {
		im.messageMapper.MarkAsFailed(ctx, messageID)
		return fmt.Errorf("failed to get Chatwoot client: %w", err)
	}

	// Extract phone number from JID
	phoneNumber := im.extractPhoneFromJID(from)
	if phoneNumber == "" {
		im.messageMapper.MarkAsFailed(ctx, messageID)
		return fmt.Errorf("failed to extract phone number from JID: %s", from)
	}

	// Get or create contact
	contact, err := im.getOrCreateContact(client, phoneNumber, sessionID)
	if err != nil {
		im.messageMapper.MarkAsFailed(ctx, messageID)
		return fmt.Errorf("failed to get or create contact: %w", err)
	}

	// Get or create conversation
	conversation, err := im.getOrCreateConversation(client, contact.ID, sessionID)
	if err != nil {
		im.messageMapper.MarkAsFailed(ctx, messageID)
		return fmt.Errorf("failed to get or create conversation: %w", err)
	}

	// Format content for Chatwoot
	formattedContent := im.formatContentForChatwoot(content, messageType)

	// Send message to Chatwoot
	chatwootMessage, err := client.SendMessage(conversation.ID, formattedContent)
	if err != nil {
		im.messageMapper.MarkAsFailed(ctx, messageID)
		return fmt.Errorf("failed to send message to Chatwoot: %w", err)
	}

	// Update mapping with Chatwoot IDs
	err = im.messageMapper.UpdateMapping(ctx, messageID, chatwootMessage.ID, conversation.ID)
	if err != nil {
		im.logger.WarnWithFields("Failed to update mapping", map[string]interface{}{
			"message_id":         messageID,
			"cw_message_id":      chatwootMessage.ID,
			"cw_conversation_id": conversation.ID,
			"error":              err.Error(),
		})
		// Don't return error here as the message was sent successfully
	}

	im.logger.InfoWithFields("WhatsApp message processed successfully", map[string]interface{}{
		"session_id":         sessionID,
		"message_id":         messageID,
		"cw_message_id":      chatwootMessage.ID,
		"cw_conversation_id": conversation.ID,
	})

	return nil
}

// extractPhoneFromJID extracts phone number from WhatsApp JID
func (im *IntegrationManager) extractPhoneFromJID(jid string) string {
	// Remove @s.whatsapp.net or @g.us suffix
	phone := strings.Split(jid, "@")[0]

	// For group JIDs, extract the creator's phone
	if strings.Contains(phone, "-") {
		parts := strings.Split(phone, "-")
		if len(parts) > 0 {
			phone = parts[0]
		}
	}

	return phone
}

// getOrCreateContact gets or creates a contact in Chatwoot
func (im *IntegrationManager) getOrCreateContact(client ports.ChatwootClient, phoneNumber, sessionID string) (*ports.ChatwootContact, error) {
	// Try to find existing contact
	contact, err := client.FindContact(phoneNumber, 1) // Assuming inbox ID 1 for now
	if err == nil {
		return contact, nil
	}

	// Create new contact
	contactName := phoneNumber // Use phone as name initially
	contact, err = client.CreateContact(phoneNumber, contactName, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	im.logger.InfoWithFields("Created new Chatwoot contact", map[string]interface{}{
		"contact_id":   contact.ID,
		"phone_number": phoneNumber,
		"session_id":   sessionID,
	})

	return contact, nil
}

// getOrCreateConversation gets or creates a conversation in Chatwoot
func (im *IntegrationManager) getOrCreateConversation(client ports.ChatwootClient, contactID int, sessionID string) (*ports.ChatwootConversation, error) {
	// Try to find existing conversation
	conversation, err := client.GetConversation(contactID, 1) // Assuming inbox ID 1 for now
	if err == nil {
		return conversation, nil
	}

	// Create new conversation
	conversation, err = client.CreateConversation(contactID, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	im.logger.InfoWithFields("Created new Chatwoot conversation", map[string]interface{}{
		"conversation_id": conversation.ID,
		"contact_id":      contactID,
		"session_id":      sessionID,
	})

	return conversation, nil
}

// formatContentForChatwoot formats message content for Chatwoot
func (im *IntegrationManager) formatContentForChatwoot(content, messageType string) string {
	switch messageType {
	case "text":
		// Format markdown for Chatwoot
		return im.formatter.FormatMarkdownForChatwoot(content)
	case "image":
		return "🖼️ **Image**\n" + content
	case "video":
		return "🎥 **Video**\n" + content
	case "audio":
		return "🎵 **Audio**\n" + content
	case "document":
		return "📄 **Document**\n" + content
	case "contact":
		return "👤 **Contact**\n" + content
	case "contacts":
		return "👥 **Contacts**\n" + content
	case "location":
		return "📍 **Location**\n" + content
	case "sticker":
		return "😊 **Sticker**"
	default:
		return content
	}
}

// GetMappingStats returns statistics about message mappings
func (im *IntegrationManager) GetMappingStats(sessionID string) (*MappingStats, error) {
	ctx := context.Background()
	return im.messageMapper.GetMappingStats(ctx, sessionID)
}

// ProcessPendingMessages processes pending message mappings
func (im *IntegrationManager) ProcessPendingMessages(sessionID string, limit int) error {
	ctx := context.Background()

	im.logger.InfoWithFields("Processing pending messages", map[string]interface{}{
		"session_id": sessionID,
		"limit":      limit,
	})

	pendingMappings, err := im.messageMapper.GetPendingMappings(ctx, sessionID, limit)
	if err != nil {
		return fmt.Errorf("failed to get pending mappings: %w", err)
	}

	if len(pendingMappings) == 0 {
		im.logger.DebugWithFields("No pending messages to process", map[string]interface{}{
			"session_id": sessionID,
		})
		return nil
	}

	processed := 0
	failed := 0

	for _, mapping := range pendingMappings {
		// For pending mappings, we would need to get the original message data
		// This is a simplified version that just marks them as failed
		// In a real implementation, you'd store more message data or retrieve it from WhatsApp

		err := im.messageMapper.MarkAsFailed(ctx, mapping.ZpMessageID)
		if err != nil {
			im.logger.WarnWithFields("Failed to mark mapping as failed", map[string]interface{}{
				"message_id": mapping.ZpMessageID,
				"error":      err.Error(),
			})
			failed++
		} else {
			processed++
		}
	}

	im.logger.InfoWithFields("Processed pending messages", map[string]interface{}{
		"session_id": sessionID,
		"processed":  processed,
		"failed":     failed,
		"total":      len(pendingMappings),
	})

	return nil
}

// CleanupOldMappings removes old message mappings
func (im *IntegrationManager) CleanupOldMappings(sessionID string, olderThanDays int) (int, error) {
	ctx := context.Background()
	return im.messageMapper.CleanupOldMappings(ctx, sessionID, olderThanDays)
}
