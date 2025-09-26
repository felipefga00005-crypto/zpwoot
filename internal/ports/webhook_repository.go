package ports

import (
	"context"

	"zpwoot/internal/domain/webhook"
)

// WebhookRepository defines the interface for webhook data persistence
type WebhookRepository interface {
	// Create creates a new webhook configuration
	Create(ctx context.Context, webhook *webhook.WebhookConfig) error

	// GetByID retrieves a webhook by its ID
	GetByID(ctx context.Context, id string) (*webhook.WebhookConfig, error)

	// GetBySessionID retrieves webhooks for a specific session
	GetBySessionID(ctx context.Context, sessionID string) ([]*webhook.WebhookConfig, error)

	// GetGlobalWebhooks retrieves global webhooks (no session ID)
	GetGlobalWebhooks(ctx context.Context) ([]*webhook.WebhookConfig, error)

	// List retrieves webhooks with optional filters
	List(ctx context.Context, req *webhook.ListWebhooksRequest) ([]*webhook.WebhookConfig, int, error)

	// Update updates an existing webhook configuration
	Update(ctx context.Context, webhook *webhook.WebhookConfig) error

	// Delete removes a webhook by ID
	Delete(ctx context.Context, id string) error

	// UpdateStatus updates only the active status of a webhook
	UpdateStatus(ctx context.Context, id string, active bool) error

	// GetActiveWebhooks retrieves all active webhooks
	GetActiveWebhooks(ctx context.Context) ([]*webhook.WebhookConfig, error)

	// GetWebhooksByEvent retrieves webhooks that listen to a specific event
	GetWebhooksByEvent(ctx context.Context, eventType string) ([]*webhook.WebhookConfig, error)

	// CountByStatus counts webhooks by active status
	CountByStatus(ctx context.Context, active bool) (int, error)

	// GetWebhookStats retrieves webhook statistics
	GetWebhookStats(ctx context.Context, webhookID string) (*WebhookStats, error)

	// UpdateWebhookStats updates webhook statistics
	UpdateWebhookStats(ctx context.Context, webhookID string, stats *WebhookStats) error
}

// WebhookStats represents webhook delivery statistics
type WebhookStats struct {
	WebhookID       string `json:"webhook_id" db:"webhook_id"`
	TotalDeliveries int64  `json:"total_deliveries" db:"total_deliveries"`
	SuccessCount    int64  `json:"success_count" db:"success_count"`
	FailureCount    int64  `json:"failure_count" db:"failure_count"`
	LastDelivery    int64  `json:"last_delivery" db:"last_delivery"`
	LastSuccess     int64  `json:"last_success" db:"last_success"`
	LastFailure     int64  `json:"last_failure" db:"last_failure"`
	AverageLatency  int64  `json:"average_latency" db:"average_latency"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID           string `json:"id" db:"id"`
	WebhookID    string `json:"webhook_id" db:"webhook_id"`
	EventID      string `json:"event_id" db:"event_id"`
	URL          string `json:"url" db:"url"`
	Payload      string `json:"payload" db:"payload"`
	StatusCode   int    `json:"status_code" db:"status_code"`
	ResponseBody string `json:"response_body" db:"response_body"`
	Latency      int64  `json:"latency" db:"latency"`
	Success      bool   `json:"success" db:"success"`
	Error        string `json:"error,omitempty" db:"error"`
	CreatedAt    int64  `json:"created_at" db:"created_at"`
}

// WebhookDeliveryRepository defines the interface for webhook delivery persistence
type WebhookDeliveryRepository interface {
	// Create creates a new webhook delivery record
	Create(ctx context.Context, delivery *WebhookDelivery) error

	// GetByWebhookID retrieves deliveries for a specific webhook
	GetByWebhookID(ctx context.Context, webhookID string, limit, offset int) ([]*WebhookDelivery, error)

	// GetByEventID retrieves deliveries for a specific event
	GetByEventID(ctx context.Context, eventID string) ([]*WebhookDelivery, error)

	// GetFailedDeliveries retrieves failed deliveries for retry
	GetFailedDeliveries(ctx context.Context, limit int) ([]*WebhookDelivery, error)

	// UpdateDeliveryStatus updates the status of a delivery
	UpdateDeliveryStatus(ctx context.Context, deliveryID string, success bool, statusCode int, responseBody, error string) error

	// DeleteOldDeliveries removes old delivery records
	DeleteOldDeliveries(ctx context.Context, olderThan int64) error

	// GetDeliveryStats retrieves delivery statistics
	GetDeliveryStats(ctx context.Context, webhookID string, from, to int64) (*DeliveryStats, error)
}

// DeliveryStats represents delivery statistics for a time period
type DeliveryStats struct {
	WebhookID       string  `json:"webhook_id"`
	TotalDeliveries int64   `json:"total_deliveries"`
	SuccessCount    int64   `json:"success_count"`
	FailureCount    int64   `json:"failure_count"`
	SuccessRate     float64 `json:"success_rate"`
	AverageLatency  float64 `json:"average_latency"`
	From            int64   `json:"from"`
	To              int64   `json:"to"`
}
