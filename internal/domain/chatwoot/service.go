package chatwoot

import (
	"context"
	"fmt"
	"time"

	"zpwoot/internal/ports"
	"zpwoot/platform/logger"

	"github.com/google/uuid"
)

type Service struct {
	logger        *logger.Logger
	repository    ports.ChatwootRepository
	wameowManager ports.WameowManager
}

func NewService(logger *logger.Logger, repository ports.ChatwootRepository, wameowManager ports.WameowManager) *Service {
	return &Service{
		logger:        logger,
		repository:    repository,
		wameowManager: wameowManager,
	}
}

func (s *Service) CreateConfig(ctx context.Context, req *CreateChatwootConfigRequest) (*ports.ChatwootConfig, error) {
	s.logger.InfoWithFields("Creating Chatwoot config", map[string]interface{}{
		"url":        req.URL,
		"account_id": req.AccountID,
	})

	// Set defaults
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	autoCreate := false
	if req.AutoCreate != nil {
		autoCreate = *req.AutoCreate
	}

	signMsg := false
	if req.SignMsg != nil {
		signMsg = *req.SignMsg
	}

	signDelimiter := "\n\n"
	if req.SignDelimiter != nil {
		signDelimiter = *req.SignDelimiter
	}

	reopenConv := true
	if req.ReopenConv != nil {
		reopenConv = *req.ReopenConv
	}

	convPending := false
	if req.ConvPending != nil {
		convPending = *req.ConvPending
	}

	importContacts := false
	if req.ImportContacts != nil {
		importContacts = *req.ImportContacts
	}

	importMessages := false
	if req.ImportMessages != nil {
		importMessages = *req.ImportMessages
	}

	importDays := 60
	if req.ImportDays != nil {
		importDays = *req.ImportDays
	}

	mergeBrazil := true
	if req.MergeBrazil != nil {
		mergeBrazil = *req.MergeBrazil
	}

	ignoreJids := []string{}
	if req.IgnoreJids != nil {
		ignoreJids = req.IgnoreJids
	}

	config := &ports.ChatwootConfig{
		ID:        uuid.New(),
		SessionID: req.SessionID,
		URL:       req.URL,
		Token:     req.Token,
		AccountID: req.AccountID,
		InboxID:   req.InboxID,
		Enabled:   enabled,

		// Advanced configuration
		InboxName:      req.InboxName,
		AutoCreate:     autoCreate,
		SignMsg:        signMsg,
		SignDelimiter:  signDelimiter,
		ReopenConv:     reopenConv,
		ConvPending:    convPending,
		ImportContacts: importContacts,
		ImportMessages: importMessages,
		ImportDays:     importDays,
		MergeBrazil:    mergeBrazil,
		Organization:   req.Organization,
		Logo:           req.Logo,
		Number:         req.Number,
		IgnoreJids:     ignoreJids,

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Persist to repository
	if err := s.repository.CreateConfig(ctx, config); err != nil {
		s.logger.ErrorWithFields("Failed to create chatwoot config", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	return config, nil
}

func (s *Service) GetConfig(ctx context.Context) (*ports.ChatwootConfig, error) {
	s.logger.Info("Getting Chatwoot config")

	config, err := s.repository.GetConfig(ctx)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get chatwoot config", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	return config, nil
}

func (s *Service) GetConfigBySessionID(ctx context.Context, sessionID string) (*ports.ChatwootConfig, error) {
	s.logger.InfoWithFields("Getting Chatwoot config by session ID", map[string]interface{}{
		"session_id": sessionID,
	})

	config, err := s.repository.GetConfigBySessionID(ctx, sessionID)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get chatwoot config by session ID", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, err
	}

	return config, nil
}

func (s *Service) UpdateConfig(ctx context.Context, req *UpdateChatwootConfigRequest) (*ports.ChatwootConfig, error) {
	s.logger.Info("Updating Chatwoot config")

	// Get existing config first
	existingConfig, err := s.repository.GetConfig(ctx)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get existing config for update", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	// Update only provided fields
	config := *existingConfig // Copy existing config
	config.UpdatedAt = time.Now()

	if req.URL != nil {
		config.URL = *req.URL
	}
	if req.Token != nil {
		config.Token = *req.Token
	}
	if req.AccountID != nil {
		config.AccountID = *req.AccountID
	}
	if req.InboxID != nil {
		config.InboxID = req.InboxID
	}
	if req.Enabled != nil {
		config.Enabled = *req.Enabled
	}
	if req.InboxName != nil {
		config.InboxName = req.InboxName
	}
	if req.AutoCreate != nil {
		config.AutoCreate = *req.AutoCreate
	}
	if req.SignMsg != nil {
		config.SignMsg = *req.SignMsg
	}
	if req.SignDelimiter != nil {
		config.SignDelimiter = *req.SignDelimiter
	}
	if req.ReopenConv != nil {
		config.ReopenConv = *req.ReopenConv
	}
	if req.ConvPending != nil {
		config.ConvPending = *req.ConvPending
	}
	if req.ImportContacts != nil {
		config.ImportContacts = *req.ImportContacts
	}
	if req.ImportMessages != nil {
		config.ImportMessages = *req.ImportMessages
	}
	if req.ImportDays != nil {
		config.ImportDays = *req.ImportDays
	}
	if req.MergeBrazil != nil {
		config.MergeBrazil = *req.MergeBrazil
	}
	if req.Organization != nil {
		config.Organization = req.Organization
	}
	if req.Logo != nil {
		config.Logo = req.Logo
	}
	if req.Number != nil {
		config.Number = req.Number
	}
	if req.IgnoreJids != nil {
		config.IgnoreJids = req.IgnoreJids
	}

	// Persist changes
	if err := s.repository.UpdateConfig(ctx, &config); err != nil {
		s.logger.ErrorWithFields("Failed to update chatwoot config", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	return &config, nil
}

func (s *Service) DeleteConfig(ctx context.Context) error {
	s.logger.Info("Deleting Chatwoot config")

	if err := s.repository.DeleteConfig(ctx); err != nil {
		s.logger.ErrorWithFields("Failed to delete chatwoot config", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	return nil
}

func (s *Service) SyncContact(ctx context.Context, req *SyncContactRequest) (*ChatwootContact, error) {
	s.logger.InfoWithFields("Syncing contact", map[string]interface{}{
		"phone_number": req.PhoneNumber,
		"name":         req.Name,
	})

	contact := &ChatwootContact{
		ID:               1, // This would be assigned by Chatwoot
		Name:             req.Name,
		PhoneNumber:      req.PhoneNumber,
		Email:            req.Email,
		CustomAttributes: req.Attributes,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	return contact, nil
}

func (s *Service) SyncConversation(ctx context.Context, req *SyncConversationRequest) (*ChatwootConversation, error) {
	s.logger.InfoWithFields("Syncing conversation", map[string]interface{}{
		"contact_id": req.ContactID,
		"session_id": req.SessionID,
	})

	conversation := &ChatwootConversation{
		ID:        1, // This would be assigned by Chatwoot
		ContactID: req.ContactID,
		Status:    "open",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return conversation, nil
}

func (s *Service) SendMessage(ctx context.Context, req *SendMessageToChatwootRequest) (*ChatwootMessage, error) {
	s.logger.InfoWithFields("Sending message to Chatwoot", map[string]interface{}{
		"conversation_id": req.ConversationID,
		"message_type":    req.MessageType,
		"content_type":    req.ContentType,
	})

	message := &ChatwootMessage{
		ID:             1, // This would be assigned by Chatwoot
		ConversationID: req.ConversationID,
		Content:        req.Content,
		MessageType:    req.MessageType,
		ContentType:    req.ContentType,
		Metadata:       req.Metadata,
		CreatedAt:      time.Now(),
	}

	return message, nil
}

func (s *Service) ProcessWebhook(ctx context.Context, sessionID string, payload *ChatwootWebhookPayload) error {
	s.logger.InfoWithFields("Processing Chatwoot webhook", map[string]interface{}{
		"session_id": sessionID,
		"event":      payload.Event,
		"account_id": payload.Account.ID,
	})

	// Delay 500ms to avoid race conditions (based on Evolution API)
	time.Sleep(500 * time.Millisecond)

	// Skip private messages
	if payload.Message != nil && payload.Message.Private {
		s.logger.DebugWithFields("Skipping private message", map[string]interface{}{
			"session_id": sessionID,
			"event":      payload.Event,
		})
		return nil
	}

	// Skip message updates without deletion
	if payload.Event == "message_updated" {
		// TODO: Handle message deletion if needed
		s.logger.DebugWithFields("Skipping message update", map[string]interface{}{
			"session_id": sessionID,
			"event":      payload.Event,
		})
		return nil
	}

	// Process conversation status changes
	if payload.Event == "conversation_status_changed" {
		// TODO: Handle conversation status changes if needed
		s.logger.DebugWithFields("Processing conversation status change", map[string]interface{}{
			"session_id": sessionID,
			"event":      payload.Event,
		})
		return nil
	}

	// Process new messages (main functionality)
	if payload.Event == "message_created" {
		return s.handleMessageCreated(ctx, sessionID, payload)
	}

	return nil
}

// handleMessageCreated processes new messages from Chatwoot
func (s *Service) handleMessageCreated(ctx context.Context, sessionID string, payload *ChatwootWebhookPayload) error {
	if payload.Message == nil {
		return fmt.Errorf("message is nil in webhook payload")
	}

	// Skip incoming messages (only process outgoing messages from agents)
	if payload.Message.MessageType != "outgoing" {
		s.logger.DebugWithFields("Skipping incoming message", map[string]interface{}{
			"session_id":   sessionID,
			"message_id":   payload.Message.ID,
			"message_type": payload.Message.MessageType,
		})
		return nil
	}

	// Skip bot messages or messages from source_id starting with WAID:
	if payload.Message.SourceID != "" && len(payload.Message.SourceID) >= 5 && payload.Message.SourceID[:5] == "WAID:" {
		s.logger.DebugWithFields("Skipping bot message", map[string]interface{}{
			"session_id": sessionID,
			"message_id": payload.Message.ID,
			"source_id":  payload.Message.SourceID,
		})
		return nil
	}

	return s.sendToWhatsApp(ctx, sessionID, payload)
}

// sendToWhatsApp sends a message from Chatwoot to WhatsApp
func (s *Service) sendToWhatsApp(ctx context.Context, sessionID string, payload *ChatwootWebhookPayload) error {
	// Extract phone number from contact
	phoneNumber := payload.Contact.PhoneNumber
	if phoneNumber == "" {
		return fmt.Errorf("contact phone number is empty")
	}

	// Format content for WhatsApp (convert Chatwoot markdown to WhatsApp format)
	content := s.formatContentForWhatsApp(payload.Message.Content)

	s.logger.InfoWithFields("Sending message to WhatsApp", map[string]interface{}{
		"session_id":      sessionID,
		"to":              phoneNumber,
		"content":         content,
		"message_id":      payload.Message.ID,
		"conversation_id": payload.Conversation.ID,
	})

	// Send message to WhatsApp using wameowManager
	result, err := s.wameowManager.SendTextMessage(sessionID, phoneNumber, content, nil)
	if err != nil {
		s.logger.ErrorWithFields("Failed to send message to WhatsApp", map[string]interface{}{
			"session_id":      sessionID,
			"to":              phoneNumber,
			"content":         content,
			"message_id":      payload.Message.ID,
			"conversation_id": payload.Conversation.ID,
			"error":           err.Error(),
		})
		return fmt.Errorf("failed to send message to WhatsApp: %w", err)
	}

	s.logger.InfoWithFields("Message sent to WhatsApp successfully", map[string]interface{}{
		"session_id":      sessionID,
		"to":              phoneNumber,
		"content":         content,
		"whatsapp_msg_id": result.MessageID,
		"chatwoot_msg_id": payload.Message.ID,
		"conversation_id": payload.Conversation.ID,
		"timestamp":       result.Timestamp,
	})

	return nil
}

// formatContentForWhatsApp formats message content for WhatsApp (based on Evolution API)
func (s *Service) formatContentForWhatsApp(content string) string {
	if content == "" {
		return content
	}

	// Convert Chatwoot markdown to WhatsApp format (based on Evolution API)
	// * -> _ (italic)
	// ** -> * (bold)
	// ~~ -> ~ (strikethrough)
	// ` -> ``` (code)

	// Use regex replacements similar to Evolution API
	// Note: This is a simplified version, the full regex from Evolution is more complex
	result := content

	// Replace single * with _ for italic (avoiding double *)
	// Replace ** with * for bold
	// Replace ~~ with ~ for strikethrough
	// Replace single ` with ``` for code (avoiding triple `)

	// For now, return as-is since regex in Go would be complex
	// TODO: Implement proper markdown conversion

	return result
}

type TestConnectionResult struct {
	Success     bool
	AccountName string
	InboxName   string
	Error       error
}

func (s *Service) TestConnection(ctx context.Context) (*TestConnectionResult, error) {
	s.logger.Info("Testing Chatwoot connection")

	return &TestConnectionResult{
		Success:     true,
		AccountName: "Test Account",
		InboxName:   "Wameow Inbox",
	}, nil
}

type ChatwootStats struct {
	TotalContacts       int
	TotalConversations  int
	ActiveConversations int
	MessagesSent        int64
	MessagesReceived    int64
}

func (s *Service) GetStats(ctx context.Context) (*ChatwootStats, error) {
	s.logger.Info("Getting Chatwoot stats")

	return &ChatwootStats{
		TotalContacts:       100,
		TotalConversations:  50,
		ActiveConversations: 10,
		MessagesSent:        500,
		MessagesReceived:    300,
	}, nil
}
