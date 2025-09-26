package chatwoot

import (
	"context"
	"zpwoot/internal/domain/chatwoot"
	"zpwoot/internal/ports"
)

// UseCase defines the chatwoot use case interface
type UseCase interface {
	CreateConfig(ctx context.Context, req *CreateChatwootConfigRequest) (*CreateChatwootConfigResponse, error)
	GetConfig(ctx context.Context) (*ChatwootConfigResponse, error)
	UpdateConfig(ctx context.Context, req *UpdateChatwootConfigRequest) (*ChatwootConfigResponse, error)
	DeleteConfig(ctx context.Context) error
	SyncContact(ctx context.Context, req *SyncContactRequest) (*SyncContactResponse, error)
	SyncConversation(ctx context.Context, req *SyncConversationRequest) (*SyncConversationResponse, error)
	SendMessageToChatwoot(ctx context.Context, req *SendMessageToChatwootRequest) (*SendMessageToChatwootResponse, error)
	ProcessWebhook(ctx context.Context, payload *ChatwootWebhookPayload) error
	TestConnection(ctx context.Context) (*TestChatwootConnectionResponse, error)
	GetStats(ctx context.Context) (*ChatwootStatsResponse, error)
}

// useCaseImpl implements the chatwoot use case
type useCaseImpl struct {
	chatwootRepo        ports.ChatwootRepository
	chatwootIntegration ports.ChatwootIntegration
	chatwootService     *chatwoot.Service
}

// NewUseCase creates a new chatwoot use case
func NewUseCase(
	chatwootRepo ports.ChatwootRepository,
	chatwootIntegration ports.ChatwootIntegration,
	chatwootService *chatwoot.Service,
) UseCase {
	return &useCaseImpl{
		chatwootRepo:        chatwootRepo,
		chatwootIntegration: chatwootIntegration,
		chatwootService:     chatwootService,
	}
}

// CreateConfig creates a new Chatwoot configuration
func (uc *useCaseImpl) CreateConfig(ctx context.Context, req *CreateChatwootConfigRequest) (*CreateChatwootConfigResponse, error) {
	// Convert DTO to domain request
	domainReq := req.ToCreateChatwootConfigRequest()

	// Create config using domain service
	config, err := uc.chatwootService.CreateConfig(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	// Convert domain entity to response DTO
	response := &CreateChatwootConfigResponse{
		ID:        config.ID.String(),
		URL:       config.URL,
		AccountID: config.AccountID,
		InboxID:   config.InboxID,
		Active:    config.Active,
		CreatedAt: config.CreatedAt,
	}

	return response, nil
}

// GetConfig retrieves the current Chatwoot configuration
func (uc *useCaseImpl) GetConfig(ctx context.Context) (*ChatwootConfigResponse, error) {
	// Get config from domain service
	config, err := uc.chatwootService.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	// Convert domain entity to response DTO
	response := FromChatwootConfig(config)
	return response, nil
}

// UpdateConfig updates the Chatwoot configuration
func (uc *useCaseImpl) UpdateConfig(ctx context.Context, req *UpdateChatwootConfigRequest) (*ChatwootConfigResponse, error) {
	// Convert DTO to domain request
	domainReq := req.ToUpdateChatwootConfigRequest()

	// Update config using domain service
	config, err := uc.chatwootService.UpdateConfig(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	// Convert domain entity to response DTO
	response := FromChatwootConfig(config)
	return response, nil
}

// DeleteConfig removes the Chatwoot configuration
func (uc *useCaseImpl) DeleteConfig(ctx context.Context) error {
	return uc.chatwootService.DeleteConfig(ctx)
}

// SyncContact synchronizes a contact with Chatwoot
func (uc *useCaseImpl) SyncContact(ctx context.Context, req *SyncContactRequest) (*SyncContactResponse, error) {
	// Convert DTO to domain request
	domainReq := &chatwoot.SyncContactRequest{
		PhoneNumber: req.PhoneNumber,
		Name:        req.Name,
		Email:       req.Email,
		Attributes:  req.Attributes,
	}

	// Sync contact using domain service
	contact, err := uc.chatwootService.SyncContact(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	// Convert domain entity to response DTO
	response := &SyncContactResponse{
		ID:          contact.ID,
		PhoneNumber: contact.PhoneNumber,
		Name:        contact.Name,
		Email:       contact.Email,
		Attributes:  contact.Attributes,
		CreatedAt:   contact.CreatedAt,
		UpdatedAt:   contact.UpdatedAt,
	}

	return response, nil
}

// SyncConversation synchronizes a conversation with Chatwoot
func (uc *useCaseImpl) SyncConversation(ctx context.Context, req *SyncConversationRequest) (*SyncConversationResponse, error) {
	// Convert DTO to domain request
	domainReq := &chatwoot.SyncConversationRequest{
		ContactID: req.ContactID,
		SessionID: req.SessionID,
	}

	// Sync conversation using domain service
	conversation, err := uc.chatwootService.SyncConversation(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	// Convert domain entity to response DTO
	response := &SyncConversationResponse{
		ID:          conversation.ID,
		ContactID:   conversation.ContactID,
		SessionID:   req.SessionID, // Use from request since domain doesn't have it
		PhoneNumber: "",            // This would need to be retrieved from contact
		Status:      conversation.Status,
		CreatedAt:   conversation.CreatedAt,
		UpdatedAt:   conversation.UpdatedAt,
	}

	return response, nil
}

// SendMessageToChatwoot sends a message to Chatwoot
func (uc *useCaseImpl) SendMessageToChatwoot(ctx context.Context, req *SendMessageToChatwootRequest) (*SendMessageToChatwootResponse, error) {
	// Convert DTO to domain request
	domainReq := &chatwoot.SendMessageToChatwootRequest{
		ConversationID: req.ConversationID,
		Content:        req.Content,
		MessageType:    req.MessageType,
		ContentType:    req.ContentType,
		Attachments:    convertAttachments(req.Attachments),
		Metadata:       req.Metadata,
	}

	// Send message using domain service
	message, err := uc.chatwootService.SendMessage(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	// Convert domain entity to response DTO
	response := &SendMessageToChatwootResponse{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		Content:        message.Content,
		MessageType:    message.MessageType,
		ContentType:    message.ContentType,
		Metadata:       message.Metadata,
		CreatedAt:      message.CreatedAt,
	}

	return response, nil
}

// ProcessWebhook processes a webhook payload from Chatwoot
func (uc *useCaseImpl) ProcessWebhook(ctx context.Context, payload *ChatwootWebhookPayload) error {
	// Convert DTO to domain payload
	domainPayload := &chatwoot.ChatwootWebhookPayload{
		Event: payload.Event,
		Account: chatwoot.ChatwootAccount{
			ID:   payload.Account.ID,
			Name: payload.Account.Name,
		},
	}

	// Process webhook using domain service
	return uc.chatwootService.ProcessWebhook(ctx, domainPayload)
}

// TestConnection tests the connection to Chatwoot
func (uc *useCaseImpl) TestConnection(ctx context.Context) (*TestChatwootConnectionResponse, error) {
	// Test connection using domain service
	result, err := uc.chatwootService.TestConnection(ctx)
	if err != nil {
		return &TestChatwootConnectionResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	response := &TestChatwootConnectionResponse{
		Success:     result.Success,
		AccountName: result.AccountName,
		InboxName:   result.InboxName,
	}

	if result.Error != nil {
		response.Error = result.Error.Error()
	}

	return response, nil
}

// GetStats retrieves Chatwoot integration statistics
func (uc *useCaseImpl) GetStats(ctx context.Context) (*ChatwootStatsResponse, error) {
	// Get stats from domain service
	stats, err := uc.chatwootService.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	response := &ChatwootStatsResponse{
		TotalContacts:       stats.TotalContacts,
		TotalConversations:  stats.TotalConversations,
		ActiveConversations: stats.ActiveConversations,
		MessagesSent:        int(stats.MessagesSent),
		MessagesReceived:    int(stats.MessagesReceived),
	}

	return response, nil
}

// Helper function to convert attachments
func convertAttachments(attachments []ChatwootAttachment) []chatwoot.ChatwootAttachment {
	domainAttachments := make([]chatwoot.ChatwootAttachment, len(attachments))
	for i, att := range attachments {
		domainAttachments[i] = chatwoot.ChatwootAttachment{
			FileType: att.FileType,
			FileName: att.FileName,
		}
	}
	return domainAttachments
}
