package webhook

import (
	"time"

	"zpwoot/internal/domain/webhook"
)

// SetConfigRequest represents the request to create a webhook
type SetConfigRequest struct {
	SessionID *string  `json:"sessionId,omitempty" validate:"omitempty,uuid" example:"session-123"`
	URL       string   `json:"url" validate:"required,url" example:"https://example.com/webhook"`
	Secret    string   `json:"secret,omitempty" example:"webhook-secret-key"`
	Events    []string `json:"events" validate:"required,min=1" example:"message,status"`
} // @name SetConfigRequest

// SetConfigResponse represents the response after creating a webhook
type SetConfigResponse struct {
	ID        string    `json:"id" example:"webhook-123"`
	SessionID *string   `json:"sessionId,omitempty" example:"session-123"`
	URL       string    `json:"url" example:"https://example.com/webhook"`
	Events    []string  `json:"events" example:"message,status"`
	Active    bool      `json:"active" example:"true"`
	CreatedAt time.Time `json:"createdAt" example:"2024-01-01T00:00:00Z"`
} // @name SetConfigResponse

// UpdateWebhookRequest represents the request to update a webhook
type UpdateWebhookRequest struct {
	URL    *string  `json:"url,omitempty" validate:"omitempty,url" example:"https://example.com/new-webhook"`
	Secret *string  `json:"secret,omitempty" example:"new-webhook-secret"`
	Events []string `json:"events,omitempty" validate:"omitempty,min=1" example:"message,status,connection"`
	Active *bool    `json:"active,omitempty" example:"false"`
} // @name UpdateWebhookRequest

// ListWebhooksRequest represents the request to list webhooks
type ListWebhooksRequest struct {
	SessionID *string `json:"sessionId,omitempty" query:"sessionId" example:"session-123"`
	Active    *bool   `json:"active,omitempty" query:"active" example:"true"`
	Limit     int     `json:"limit,omitempty" query:"limit" validate:"omitempty,min=1,max=100" example:"20"`
	Offset    int     `json:"offset,omitempty" query:"offset" validate:"omitempty,min=0" example:"0"`
} // @name ListWebhooksRequest

// ListWebhooksResponse represents the response for listing webhooks
type ListWebhooksResponse struct {
	Webhooks []WebhookResponse `json:"webhooks"`
	Total    int               `json:"total" example:"5"`
	Limit    int               `json:"limit" example:"20"`
	Offset   int               `json:"offset" example:"0"`
} // @name ListWebhooksResponse

// WebhookResponse represents a webhook in responses
type WebhookResponse struct {
	ID        string    `json:"id" example:"webhook-123"`
	SessionID *string   `json:"sessionId,omitempty" example:"session-123"`
	URL       string    `json:"url" example:"https://example.com/webhook"`
	Events    []string  `json:"events" example:"message,status"`
	Active    bool      `json:"active" example:"true"`
	CreatedAt time.Time `json:"createdAt" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updatedAt" example:"2024-01-01T00:00:00Z"`
} // @name WebhookResponse

// WebhookEventResponse represents a webhook event in responses
type WebhookEventResponse struct {
	ID        string                 `json:"id" example:"event-123"`
	SessionID string                 `json:"sessionId" example:"session-123"`
	Type      string                 `json:"type" example:"message"`
	Timestamp time.Time              `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	Data      map[string]interface{} `json:"data"`
} // @name WebhookEventResponse

// TestWebhookRequest represents the request to test a webhook
type TestWebhookRequest struct {
	EventType string                 `json:"event_type" validate:"required" example:"message"`
	TestData  map[string]interface{} `json:"test_data,omitempty"`
} // @name TestWebhookRequest

// TestWebhookResponse represents the response after testing a webhook
type TestWebhookResponse struct {
	Success      bool   `json:"success" example:"true"`
	StatusCode   int    `json:"status_code" example:"200"`
	ResponseTime int64  `json:"response_time_ms" example:"150"`
	Error        string `json:"error,omitempty"`
} // @name TestWebhookResponse

// WebhookEventsResponse represents the list of supported webhook events
type WebhookEventsResponse struct {
	Events []WebhookEventInfo `json:"events"`
} // @name WebhookEventsResponse

// WebhookEventInfo represents information about a webhook event type
type WebhookEventInfo struct {
	Type        string `json:"type" example:"message"`
	Description string `json:"description" example:"Triggered when a message is received or sent"`
	DataSchema  string `json:"data_schema,omitempty" example:"MessageEventData"`
} // @name WebhookEventInfo

// Conversion methods

// ToSetConfigRequest converts to domain request
func (r *SetConfigRequest) ToSetConfigRequest() *webhook.SetConfigRequest {
	return &webhook.SetConfigRequest{
		SessionID: r.SessionID,
		URL:       r.URL,
		Secret:    r.Secret,
		Events:    r.Events,
	}
}

// ToUpdateWebhookRequest converts to domain request
func (r *UpdateWebhookRequest) ToUpdateWebhookRequest() *webhook.UpdateWebhookRequest {
	return &webhook.UpdateWebhookRequest{
		URL:    r.URL,
		Secret: r.Secret,
		Events: r.Events,
		Active: r.Active,
	}
}

// ToListWebhooksRequest converts to domain request
func (r *ListWebhooksRequest) ToListWebhooksRequest() *webhook.ListWebhooksRequest {
	return &webhook.ListWebhooksRequest{
		SessionID: r.SessionID,
		Active:    r.Active,
		Limit:     r.Limit,
		Offset:    r.Offset,
	}
}

// FromWebhook converts from domain webhook to response
func FromWebhook(w *webhook.WebhookConfig) *WebhookResponse {
	return &WebhookResponse{
		ID:        w.ID.String(),
		SessionID: w.SessionID,
		URL:       w.URL,
		Events:    w.Events,
		Active:    w.Active,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
}

// FromWebhookEvent converts from domain webhook event to response
func FromWebhookEvent(we *webhook.WebhookEvent) *WebhookEventResponse {
	return &WebhookEventResponse{
		ID:        we.ID,
		SessionID: we.SessionID,
		Type:      we.Type,
		Timestamp: we.Timestamp,
		Data:      we.Data,
	}
}

// GetSupportedEvents returns information about supported webhook events
func GetSupportedEvents() *WebhookEventsResponse {
	return &WebhookEventsResponse{
		Events: []WebhookEventInfo{
			{
				Type:        "message",
				Description: "Triggered when a message is received or sent",
				DataSchema:  "MessageEventData",
			},
			{
				Type:        "status",
				Description: "Triggered when message status changes (sent, delivered, read)",
				DataSchema:  "StatusEventData",
			},
			{
				Type:        "connection",
				Description: "Triggered when connection status changes",
				DataSchema:  "ConnectionEventData",
			},
			{
				Type:        "qr",
				Description: "Triggered when QR code is generated",
				DataSchema:  "QREventData",
			},
			{
				Type:        "pair",
				Description: "Triggered when phone pairing is successful",
				DataSchema:  "PairEventData",
			},
		},
	}
}
