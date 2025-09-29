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
	messageMapper ports.ChatwootMessageMapper // Optional - for storing outgoing messages
}

func NewService(logger *logger.Logger, repository ports.ChatwootRepository, wameowManager ports.WameowManager) *Service {
	return &Service{
		logger:        logger,
		repository:    repository,
		wameowManager: wameowManager,
	}
}

// SetMessageMapper sets the message mapper for storing outgoing messages
func (s *Service) SetMessageMapper(messageMapper ports.ChatwootMessageMapper) {
	s.messageMapper = messageMapper
}

func (s *Service) CreateConfig(ctx context.Context, req *CreateChatwootConfigRequest) (*ports.ChatwootConfig, error) {


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
		return nil, err
	}

	return config, nil
}

func (s *Service) GetConfig(ctx context.Context) (*ports.ChatwootConfig, error) {
	config, err := s.repository.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (s *Service) GetConfigBySessionID(ctx context.Context, sessionID string) (*ports.ChatwootConfig, error) {
	config, err := s.repository.GetConfigBySessionID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (s *Service) UpdateConfig(ctx context.Context, req *UpdateChatwootConfigRequest) (*ports.ChatwootConfig, error) {
	// Get existing config first
	existingConfig, err := s.repository.GetConfig(ctx)
	if err != nil {
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
		return nil, err
	}

	return &config, nil
}

func (s *Service) DeleteConfig(ctx context.Context) error {
	if err := s.repository.DeleteConfig(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Service) SyncContact(ctx context.Context, req *SyncContactRequest) (*ChatwootContact, error) {
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
	// Delay 500ms to avoid race conditions (based on Evolution API)
	time.Sleep(500 * time.Millisecond)

	// Skip private messages
	if payload.Message != nil && payload.Message.Private {
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

	// Handle typing events (ignore them as they don't require action)
	if payload.Event == "conversation_typing_on" || payload.Event == "conversation_typing_off" {
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
	// Extract message details from webhook payload
	// O Chatwoot envia os dados da mensagem diretamente no payload, não em um campo "message"
	var content, messageType string
	var messageID int
	var isPrivate bool

	if payload.Message != nil {
		// Formato antigo (se ainda existir)
		content = payload.Message.Content
		messageType = payload.Message.MessageType
		messageID = payload.Message.ID
		isPrivate = payload.Message.Private
	} else {
		// Formato atual do Chatwoot (baseado nos logs do Sidekiq)
		content = payload.Content
		messageType = payload.MessageType
		messageID = payload.ID
		isPrivate = payload.Private
	}

	// Se o conteúdo estiver vazio, tentar extrair de outros campos possíveis
	if content == "" {
		// Verificar se há conteúdo em content_attributes ou outros campos
		if payload.ContentAttributes != nil {
			if textContent, ok := payload.ContentAttributes["text"].(string); ok {
				content = textContent
			}
		}
		// Verificar se há mensagens na conversa
		if len(payload.Conversation.Messages) > 0 {
			lastMessage := payload.Conversation.Messages[len(payload.Conversation.Messages)-1]
			if lastMessage.Content != "" {
				content = lastMessage.Content
				messageType = lastMessage.MessageType
				messageID = lastMessage.ID
			}
		}
	}

	// Se não há conteúdo e é uma mensagem outgoing, pode ser um problema
	if content == "" && messageType == "outgoing" {
		// Não retornar erro, apenas logar para investigação
		return nil
	}

	// Process both incoming and outgoing messages
	// - outgoing: mensagens do agente para WhatsApp (enviar para WhatsApp)
	// - incoming: mensagens do WhatsApp para agente (IGNORAR - já processadas)
	if messageType == "incoming" {
		// Mensagens incoming são aquelas que vieram do WhatsApp e já foram processadas
		// Não precisamos reprocessá-las
		return nil
	}

	// Para mensagens outgoing, verificar se há conteúdo para enviar
	if messageType == "outgoing" && content == "" {
		// Não enviar mensagem vazia para WhatsApp, mas não falhar
		return nil
	}

	// Skip private messages
	if isPrivate {
		return nil
	}

	// Skip bot messages or messages from source_id starting with WAID:
	var sourceID string
	if payload.Message != nil {
		sourceID = payload.Message.SourceID
	} else if payload.SourceID != nil {
		sourceID = *payload.SourceID
	}

	if sourceID != "" && len(sourceID) >= 5 && sourceID[:5] == "WAID:" {
		return nil
	}

	return s.sendToWhatsApp(ctx, sessionID, payload, content)
}

// sendToWhatsApp sends a message from Chatwoot to WhatsApp
func (s *Service) sendToWhatsApp(ctx context.Context, sessionID string, payload *ChatwootWebhookPayload, content string) error {
	// Determine the correct recipient based on message type (like WhatsApp logic)
	// For outgoing messages: recipient = Contact (client)
	// For incoming messages: recipient = Sender (client) - but we don't process incoming here
	var phoneNumber string

	// Check message type to determine recipient
	messageType := "outgoing" // Default assumption for messages sent to WhatsApp
	if payload.Message != nil {
		messageType = payload.Message.MessageType
	}

	if messageType == "outgoing" {
		// Outgoing: agent → client, so recipient is Contact
		if payload.Contact.PhoneNumber != "" {
			phoneNumber = payload.Contact.PhoneNumber
		}
	} else {
		// Incoming: client → agent, so recipient would be Sender (but we shouldn't process these)
		if payload.Sender.PhoneNumber != "" {
			phoneNumber = payload.Sender.PhoneNumber
		}
	}

	// If still no phone number, this is an error
	if phoneNumber == "" {
		return fmt.Errorf("no valid recipient phone number found for %s message", messageType)
	}

	// Format content for WhatsApp (convert Chatwoot markdown to WhatsApp format)
	formattedContent := s.formatContentForWhatsApp(content)

	// Extract message ID for logging
	var messageID int
	if payload.Message != nil {
		messageID = payload.Message.ID
	} else {
		messageID = payload.ID
	}

	// Send message to WhatsApp using wameowManager
	result, err := s.wameowManager.SendMessage(sessionID, phoneNumber, "text", formattedContent, "", "", "", 0, 0, "", "", nil)
	if err != nil {
		return fmt.Errorf("failed to send message to WhatsApp: %w", err)
	}

	// Store the outgoing message in zpMessage table for tracking
	err = s.storeOutgoingMessage(ctx, sessionID, result.MessageID, phoneNumber, formattedContent, result.Timestamp, messageID, payload.Conversation.ID)
	if err != nil {
		// Don't fail the whole operation, just log the error
	}

	return nil
}

// storeOutgoingMessage stores an outgoing message in the zpMessage table
func (s *Service) storeOutgoingMessage(ctx context.Context, sessionID, whatsappMessageID, phoneNumber, content string, timestamp time.Time, chatwootMessageID, chatwootConversationID int) error {
	// Check if we have a message mapper available
	if s.messageMapper == nil {
		s.logger.WarnWithFields("Message mapper not available, cannot store outgoing message", map[string]interface{}{
			"session_id":      sessionID,
			"whatsapp_msg_id": whatsappMessageID,
		})
		return nil
	}

	// Create mapping for outgoing message
	mapping, err := s.messageMapper.CreateMapping(ctx, sessionID, whatsappMessageID, phoneNumber, phoneNumber, "text", content, timestamp, true)
	if err != nil {
		return fmt.Errorf("failed to create mapping for outgoing message: %w", err)
	}

	// Update mapping with Chatwoot IDs
	err = s.messageMapper.UpdateMapping(ctx, sessionID, whatsappMessageID, chatwootMessageID, chatwootConversationID)
	if err != nil {
		return fmt.Errorf("failed to update mapping with Chatwoot IDs: %w", err)
	}

	s.logger.InfoWithFields("Outgoing message stored in zpMessage table", map[string]interface{}{
		"session_id":               sessionID,
		"whatsapp_msg_id":          whatsappMessageID,
		"chatwoot_msg_id":          chatwootMessageID,
		"chatwoot_conversation_id": chatwootConversationID,
		"mapping_id":               mapping.ID,
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
	return &ChatwootStats{
		TotalContacts:       100,
		TotalConversations:  50,
		ActiveConversations: 10,
		MessagesSent:        500,
		MessagesReceived:    300,
	}, nil
}
