package webhook

import (
	"context"

	"zpwoot/internal/domain/webhook"
	"zpwoot/internal/ports"
)

// UseCase defines the webhook use case interface
type UseCase interface {
	SetConfig(ctx context.Context, req *SetConfigRequest) (*SetConfigResponse, error)
	FindConfig(ctx context.Context, sessionID string) (*WebhookResponse, error)
	UpdateWebhook(ctx context.Context, webhookID string, req *UpdateWebhookRequest) (*WebhookResponse, error)
	DeleteWebhook(ctx context.Context, webhookID string) error
	ListWebhooks(ctx context.Context, req *ListWebhooksRequest) (*ListWebhooksResponse, error)
	TestWebhook(ctx context.Context, webhookID string, req *TestWebhookRequest) (*TestWebhookResponse, error)
	GetSupportedWebhookEvents(ctx context.Context) (*WebhookEventsResponse, error)
	ProcessWebhookEvent(ctx context.Context, event *webhook.WebhookEvent) error
}

// useCaseImpl implements the webhook use case
type useCaseImpl struct {
	webhookRepo    ports.WebhookRepository
	webhookService *webhook.Service
}

// NewUseCase creates a new webhook use case
func NewUseCase(
	webhookRepo ports.WebhookRepository,
	webhookService *webhook.Service,
) UseCase {
	return &useCaseImpl{
		webhookRepo:    webhookRepo,
		webhookService: webhookService,
	}
}

// SetConfig creates a new webhook configuration
func (uc *useCaseImpl) SetConfig(ctx context.Context, req *SetConfigRequest) (*SetConfigResponse, error) {
	// Convert DTO to domain request
	domainReq := req.ToSetConfigRequest()

	// Create webhook using domain service
	webhookConfig, err := uc.webhookService.SetConfig(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	// Convert domain entity to response DTO
	response := &SetConfigResponse{
		ID:        webhookConfig.ID.String(),
		SessionID: webhookConfig.SessionID,
		URL:       webhookConfig.URL,
		Events:    webhookConfig.Events,
		Active:    webhookConfig.Active,
		CreatedAt: webhookConfig.CreatedAt,
	}

	return response, nil
}

// FindConfig retrieves webhook configuration for a session
func (uc *useCaseImpl) FindConfig(ctx context.Context, sessionID string) (*WebhookResponse, error) {
	// Get webhook config from domain service
	webhookConfig, err := uc.webhookService.GetWebhookBySession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Convert domain entity to response DTO
	response := FromWebhook(webhookConfig)
	return response, nil
}

// UpdateWebhook updates an existing webhook configuration
func (uc *useCaseImpl) UpdateWebhook(ctx context.Context, webhookID string, req *UpdateWebhookRequest) (*WebhookResponse, error) {
	// Convert DTO to domain request
	domainReq := req.ToUpdateWebhookRequest()

	// Update webhook using domain service
	webhookConfig, err := uc.webhookService.UpdateWebhook(ctx, webhookID, domainReq)
	if err != nil {
		return nil, err
	}

	// Convert domain entity to response DTO
	response := FromWebhook(webhookConfig)
	return response, nil
}

// DeleteWebhook removes a webhook configuration
func (uc *useCaseImpl) DeleteWebhook(ctx context.Context, webhookID string) error {
	return uc.webhookService.DeleteWebhook(ctx, webhookID)
}

// ListWebhooks retrieves a list of webhook configurations
func (uc *useCaseImpl) ListWebhooks(ctx context.Context, req *ListWebhooksRequest) (*ListWebhooksResponse, error) {
	// Convert DTO to domain request
	domainReq := req.ToListWebhooksRequest()

	// Set defaults
	if domainReq.Limit == 0 {
		domainReq.Limit = 20
	}

	// Get webhooks from domain service
	webhooks, total, err := uc.webhookService.ListWebhooks(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	// Convert domain entities to response DTOs
	webhookResponses := make([]WebhookResponse, len(webhooks))
	for i, wh := range webhooks {
		webhookResponses[i] = *FromWebhook(wh)
	}

	response := &ListWebhooksResponse{
		Webhooks: webhookResponses,
		Total:    total,
		Limit:    domainReq.Limit,
		Offset:   domainReq.Offset,
	}

	return response, nil
}

// TestWebhook tests a webhook configuration by sending a test event
func (uc *useCaseImpl) TestWebhook(ctx context.Context, webhookID string, req *TestWebhookRequest) (*TestWebhookResponse, error) {
	// Create test event
	testEvent := &webhook.WebhookEvent{
		ID:        "test-" + webhookID,
		SessionID: "test-session",
		Type:      req.EventType,
		Data:      req.TestData,
	}

	// Test webhook using domain service
	result, err := uc.webhookService.TestWebhook(ctx, webhookID, testEvent)
	if err != nil {
		return &TestWebhookResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	response := &TestWebhookResponse{
		Success:      result.Success,
		StatusCode:   result.StatusCode,
		ResponseTime: result.ResponseTime,
	}

	if result.Error != nil {
		response.Error = result.Error.Error()
	}

	return response, nil
}

// GetSupportedWebhookEvents returns information about supported webhook events
func (uc *useCaseImpl) GetSupportedWebhookEvents(ctx context.Context) (*WebhookEventsResponse, error) {
	return GetSupportedEvents(), nil
}

// ProcessWebhookEvent processes a webhook event and sends it to configured webhooks
func (uc *useCaseImpl) ProcessWebhookEvent(ctx context.Context, event *webhook.WebhookEvent) error {
	return uc.webhookService.ProcessEvent(ctx, event)
}
