package webhook

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// WebhookConfig represents webhook configuration
type WebhookConfig struct {
	ID        uuid.UUID `json:"id" db:"id"`
	SessionID *string   `json:"session_id,omitempty" db:"session_id"` // null for global webhooks
	URL       string    `json:"url" db:"url"`
	Secret    string    `json:"secret,omitempty" db:"secret"`
	Events    []string  `json:"events" db:"events"`
	Active    bool      `json:"active" db:"active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Domain errors
var (
	ErrWebhookNotFound       = errors.New("webhook not found")
	ErrWebhookAlreadyExists  = errors.New("webhook already exists")
	ErrInvalidWebhookURL     = errors.New("invalid webhook URL")
	ErrWebhookDeliveryFailed = errors.New("webhook delivery failed")
)

// SetConfigRequest represents a request to create a webhook
type SetConfigRequest struct {
	SessionID *string  `json:"session_id,omitempty" validate:"omitempty,uuid"`
	URL       string   `json:"url" validate:"required,url"`
	Secret    string   `json:"secret,omitempty"`
	Events    []string `json:"events" validate:"required,min=1"`
}

// UpdateWebhookRequest represents a request to update a webhook
type UpdateWebhookRequest struct {
	URL    *string  `json:"url,omitempty" validate:"omitempty,url"`
	Secret *string  `json:"secret,omitempty"`
	Events []string `json:"events,omitempty" validate:"omitempty,min=1"`
	Active *bool    `json:"active,omitempty"`
}

// ListWebhooksRequest represents filters for listing webhooks
type ListWebhooksRequest struct {
	SessionID *string `json:"session_id,omitempty" query:"session_id"`
	Active    *bool   `json:"active,omitempty" query:"active"`
	Limit     int     `json:"limit,omitempty" query:"limit" validate:"omitempty,min=1,max=100"`
	Offset    int     `json:"offset,omitempty" query:"offset" validate:"omitempty,min=0"`
}

// WebhookEvent represents an event to be sent to webhooks
type WebhookEvent struct {
	ID        string                 `json:"id"`
	SessionID string                 `json:"session_id"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// List of supported event types
var SupportedEventTypes = []string{
	// Messages and Communication
	"Message",
	"UndecryptableMessage",
	"Receipt",
	"MediaRetry",
	"ReadReceipt",

	// Groups and Contacts
	"GroupInfo",
	"JoinedGroup",
	"Picture",
	"BlocklistChange",
	"Blocklist",

	// Connection and Session
	"Connected",
	"Disconnected",
	"ConnectFailure",
	"KeepAliveRestored",
	"KeepAliveTimeout",
	"LoggedOut",
	"ClientOutdated",
	"TemporaryBan",
	"StreamError",
	"StreamReplaced",
	"PairSuccess",
	"PairError",
	"QR",
	"QRScannedWithoutMultidevice",

	// Privacy and Settings
	"PrivacySettings",
	"PushNameSetting",
	"UserAbout",

	// Synchronization and State
	"AppState",
	"AppStateSyncComplete",
	"HistorySync",
	"OfflineSyncCompleted",
	"OfflineSyncPreview",

	// Calls
	"CallOffer",
	"CallAccept",
	"CallTerminate",
	"CallOfferNotice",
	"CallRelayLatency",

	// Presence and Activity
	"Presence",
	"ChatPresence",

	// Identity
	"IdentityChange",

	// Errors
	"CATRefreshError",

	// Newsletter (Wameow Channels)
	"NewsletterJoin",
	"NewsletterLeave",
	"NewsletterMuteChange",
	"NewsletterLiveUpdate",

	// Facebook/Meta Bridge
	"FBMessage",

	// Special - receives all events
	"All",
}

// Map for quick validation
var eventTypeMap map[string]bool

func init() {
	eventTypeMap = make(map[string]bool)
	for _, eventType := range SupportedEventTypes {
		eventTypeMap[eventType] = true
	}
}

// IsValidEventType validates if an event type is supported
func IsValidEventType(eventType string) bool {
	return eventTypeMap[eventType]
}

// ValidateEvents validates a list of event types
func ValidateEvents(events []string) []string {
	var invalidEvents []string
	for _, event := range events {
		if !IsValidEventType(event) {
			invalidEvents = append(invalidEvents, event)
		}
	}
	return invalidEvents
}

// NewWebhookConfig creates a new webhook configuration
func NewWebhookConfig(sessionID *string, url, secret string, events []string) *WebhookConfig {
	return &WebhookConfig{
		ID:        uuid.New(),
		SessionID: sessionID,
		URL:       url,
		Secret:    secret,
		Events:    events,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// IsGlobal returns true if this is a global webhook (no session ID)
func (w *WebhookConfig) IsGlobal() bool {
	return w.SessionID == nil
}

// HasEvent checks if the webhook is configured to receive a specific event type
func (w *WebhookConfig) HasEvent(eventType string) bool {
	for _, event := range w.Events {
		if event == "All" || event == eventType {
			return true
		}
	}
	return false
}

// Update updates the webhook configuration
func (w *WebhookConfig) Update(req *UpdateWebhookRequest) {
	if req.URL != nil {
		w.URL = *req.URL
	}
	if req.Secret != nil {
		w.Secret = *req.Secret
	}
	if req.Events != nil {
		w.Events = req.Events
	}
	if req.Active != nil {
		w.Active = *req.Active
	}
	w.UpdatedAt = time.Now()
}

// NewWebhookEvent creates a new webhook event
func NewWebhookEvent(sessionID, eventType string, data map[string]interface{}) *WebhookEvent {
	return &WebhookEvent{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
}
