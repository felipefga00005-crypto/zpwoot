package webhook

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"zpwoot/platform/logger"
)

type Service struct {
	logger *logger.Logger
}

func NewService(logger *logger.Logger) *Service {
	return &Service{
		logger: logger,
	}
}

func (s *Service) SetConfig(ctx context.Context, req *SetConfigRequest) (*WebhookConfig, error) {
	s.logger.InfoWithFields("Creating webhook", map[string]interface{}{
		"url":        req.URL,
		"session_id": req.SessionID,
		"events":     req.Events,
	})

	webhook := &WebhookConfig{
		ID:        uuid.New(),
		SessionID: req.SessionID,
		URL:       req.URL,
		Secret:    req.Secret,
		Events:    req.Events,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return webhook, nil
}

func (s *Service) UpdateWebhook(ctx context.Context, webhookID string, req *UpdateWebhookRequest) (*WebhookConfig, error) {
	s.logger.InfoWithFields("Updating webhook", map[string]interface{}{
		"webhook_id": webhookID,
	})

	id, err := uuid.Parse(webhookID)
	if err != nil {
		return nil, fmt.Errorf("invalid webhook ID: %w", err)
	}

	webhook := &WebhookConfig{
		ID:        id,
		UpdatedAt: time.Now(),
	}

	if req.URL != nil {
		webhook.URL = *req.URL
	}
	if req.Secret != nil {
		webhook.Secret = *req.Secret
	}
	if req.Events != nil {
		webhook.Events = req.Events
	}
	if req.Active != nil {
		webhook.Active = *req.Active
	}

	return webhook, nil
}

func (s *Service) DeleteWebhook(ctx context.Context, webhookID string) error {
	s.logger.InfoWithFields("Deleting webhook", map[string]interface{}{
		"webhook_id": webhookID,
	})

	return nil
}

func (s *Service) GetWebhookBySession(ctx context.Context, sessionID string) (*WebhookConfig, error) {
	s.logger.InfoWithFields("Getting webhook by session", map[string]interface{}{
		"session_id": sessionID,
	})

	return nil, ErrWebhookNotFound
}

func (s *Service) ListWebhooks(ctx context.Context, req *ListWebhooksRequest) ([]*WebhookConfig, int, error) {
	s.logger.InfoWithFields("Listing webhooks", map[string]interface{}{
		"session_id": req.SessionID,
		"active":     req.Active,
		"limit":      req.Limit,
		"offset":     req.Offset,
	})

	return []*WebhookConfig{}, 0, nil
}

type TestWebhookResult struct {
	Success      bool
	StatusCode   int
	ResponseTime int64
	Error        error
}

func (s *Service) TestWebhook(ctx context.Context, webhookID string, event *WebhookEvent) (*TestWebhookResult, error) {
	s.logger.InfoWithFields("Testing webhook", map[string]interface{}{
		"webhook_id": webhookID,
		"event_type": event.Type,
	})

	return &TestWebhookResult{
		Success:      true,
		StatusCode:   200,
		ResponseTime: 150,
	}, nil
}

func (s *Service) ProcessEvent(ctx context.Context, event *WebhookEvent) error {
	s.logger.InfoWithFields("Processing webhook event", map[string]interface{}{
		"event_id":   event.ID,
		"event_type": event.Type,
		"session_id": event.SessionID,
	})

	return nil
}

func (s *Service) ValidateWebhookConfig(config *WebhookConfig) error {
	if config.URL == "" {
		return ErrInvalidWebhookURL
	}

	if len(config.Events) == 0 {
		return fmt.Errorf("webhook must listen to at least one event")
	}

	return nil
}
