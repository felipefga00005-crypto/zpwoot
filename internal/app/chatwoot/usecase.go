package chatwoot

import (
	"context"
	"fmt"

	"zpwoot/internal/domain/chatwoot"
	chatwootIntegration "zpwoot/internal/infra/integrations/chatwoot"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

type UseCase interface {
	CreateConfig(ctx context.Context, sessionID string, req *CreateChatwootConfigRequest) (*CreateChatwootConfigResponse, error)
	GetConfig(ctx context.Context) (*ChatwootConfigResponse, error)
	UpdateConfig(ctx context.Context, req *UpdateChatwootConfigRequest) (*ChatwootConfigResponse, error)
	DeleteConfig(ctx context.Context) error
	SyncContact(ctx context.Context, req *SyncContactRequest) (*SyncContactResponse, error)
	SyncConversation(ctx context.Context, req *SyncConversationRequest) (*SyncConversationResponse, error)
	SendMessageToChatwoot(ctx context.Context, req *SendMessageToChatwootRequest) (*SendMessageToChatwootResponse, error)
	ProcessWebhook(ctx context.Context, sessionID string, payload *ChatwootWebhookPayload) error
	TestConnection(ctx context.Context) (*TestChatwootConnectionResponse, error)
	GetStats(ctx context.Context) (*ChatwootStatsResponse, error)
	AutoCreateInbox(ctx context.Context, sessionID, inboxName, webhookURL string) error
}

type useCaseImpl struct {
	chatwootRepo        ports.ChatwootRepository
	chatwootIntegration ports.ChatwootIntegration
	chatwootService     *chatwoot.Service
	logger              *logger.Logger
}

func NewUseCase(
	chatwootRepo ports.ChatwootRepository,
	chatwootIntegration ports.ChatwootIntegration,
	chatwootService *chatwoot.Service,
	logger *logger.Logger,
) UseCase {
	return &useCaseImpl{
		chatwootRepo:        chatwootRepo,
		chatwootIntegration: chatwootIntegration,
		chatwootService:     chatwootService,
		logger:              logger,
	}
}

func (uc *useCaseImpl) CreateConfig(ctx context.Context, sessionID string, req *CreateChatwootConfigRequest) (*CreateChatwootConfigResponse, error) {
	domainReq, err := req.ToCreateChatwootConfigRequest(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	config, err := uc.chatwootService.CreateConfig(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	response := &CreateChatwootConfigResponse{
		ID:        config.ID.String(),
		URL:       config.URL,
		AccountID: config.AccountID,
		InboxID:   config.InboxID,
		Active:    config.Enabled,
		CreatedAt: config.CreatedAt,
	}

	return response, nil
}

func (uc *useCaseImpl) GetConfig(ctx context.Context) (*ChatwootConfigResponse, error) {
	config, err := uc.chatwootService.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	response := FromChatwootConfig(config)
	return response, nil
}

func (uc *useCaseImpl) UpdateConfig(ctx context.Context, req *UpdateChatwootConfigRequest) (*ChatwootConfigResponse, error) {
	domainReq := req.ToUpdateChatwootConfigRequest()

	config, err := uc.chatwootService.UpdateConfig(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	response := FromChatwootConfig(config)
	return response, nil
}

func (uc *useCaseImpl) DeleteConfig(ctx context.Context) error {
	return uc.chatwootService.DeleteConfig(ctx)
}

func (uc *useCaseImpl) SyncContact(ctx context.Context, req *SyncContactRequest) (*SyncContactResponse, error) {
	domainReq := &chatwoot.SyncContactRequest{
		PhoneNumber: req.PhoneNumber,
		Name:        req.Name,
		Email:       req.Email,
		Attributes:  req.Attributes,
	}

	contact, err := uc.chatwootService.SyncContact(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	response := &SyncContactResponse{
		ID:          contact.ID,
		PhoneNumber: contact.PhoneNumber,
		Name:        contact.Name,
		Email:       contact.Email,
		Attributes:  contact.CustomAttributes,
		CreatedAt:   contact.CreatedAt,
		UpdatedAt:   contact.UpdatedAt,
	}

	return response, nil
}

func (uc *useCaseImpl) SyncConversation(ctx context.Context, req *SyncConversationRequest) (*SyncConversationResponse, error) {
	domainReq := &chatwoot.SyncConversationRequest{
		ContactID: req.ContactID,
		SessionID: req.SessionID,
	}

	conversation, err := uc.chatwootService.SyncConversation(ctx, domainReq)
	if err != nil {
		return nil, err
	}

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

func (uc *useCaseImpl) SendMessageToChatwoot(ctx context.Context, req *SendMessageToChatwootRequest) (*SendMessageToChatwootResponse, error) {
	domainReq := &chatwoot.SendMessageToChatwootRequest{
		ConversationID: req.ConversationID,
		Content:        req.Content,
		MessageType:    req.MessageType,
		ContentType:    req.ContentType,
		Attachments:    convertAttachments(req.Attachments),
		Metadata:       req.Metadata,
	}

	message, err := uc.chatwootService.SendMessage(ctx, domainReq)
	if err != nil {
		return nil, err
	}

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

func (uc *useCaseImpl) ProcessWebhook(ctx context.Context, sessionID string, payload *ChatwootWebhookPayload) error {
	domainPayload := &chatwoot.ChatwootWebhookPayload{
		Event: payload.Event,
		Account: chatwoot.ChatwootAccount{
			ID:   payload.Account.ID,
			Name: payload.Account.Name,
		},
	}

	return uc.chatwootService.ProcessWebhook(ctx, sessionID, domainPayload)
}

func (uc *useCaseImpl) TestConnection(ctx context.Context) (*TestChatwootConnectionResponse, error) {
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

func (uc *useCaseImpl) GetStats(ctx context.Context) (*ChatwootStatsResponse, error) {
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

func (uc *useCaseImpl) AutoCreateInbox(ctx context.Context, sessionID, inboxName, webhookURL string) error {
	// Get existing Chatwoot configuration for this session
	config, err := uc.chatwootService.GetConfigBySessionID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("chatwoot not configured for session %s: %w", sessionID, err)
	}

	// Create Chatwoot client
	client := chatwootIntegration.NewClient(config.URL, config.Token, config.AccountID, uc.logger)

	// Create inbox in Chatwoot
	inbox, err := client.CreateInbox(inboxName, webhookURL)
	if err != nil {
		return fmt.Errorf("failed to create inbox in Chatwoot: %w", err)
	}

	// Update configuration with the new inbox ID
	inboxIDStr := fmt.Sprintf("%d", inbox.ID)
	updateReq := &chatwoot.UpdateChatwootConfigRequest{
		InboxID: &inboxIDStr,
	}

	_, err = uc.chatwootService.UpdateConfig(ctx, updateReq)
	if err != nil {
		return fmt.Errorf("failed to update config with inbox ID: %w", err)
	}

	return nil
}
