package chatwoot

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ChatwootConfig represents Chatwoot integration configuration
type ChatwootConfig struct {
	ID        uuid.UUID `json:"id" db:"id"`
	URL       string    `json:"url" db:"url"`
	APIKey    string    `json:"api_key" db:"api_key"`
	AccountID string    `json:"account_id" db:"account_id"`
	InboxID   *string   `json:"inbox_id,omitempty" db:"inbox_id"`
	Active    bool      `json:"active" db:"active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Domain errors
var (
	ErrConfigNotFound       = errors.New("chatwoot config not found")
	ErrContactNotFound      = errors.New("chatwoot contact not found")
	ErrConversationNotFound = errors.New("chatwoot conversation not found")
	ErrMessageNotFound      = errors.New("chatwoot message not found")
	ErrInvalidAPIKey        = errors.New("invalid chatwoot API key")
	ErrInvalidAccountID     = errors.New("invalid chatwoot account ID")
	ErrChatwootAPIError     = errors.New("chatwoot API error")
)

// CreateChatwootConfigRequest represents a request to create Chatwoot configuration
type CreateChatwootConfigRequest struct {
	URL       string  `json:"url" validate:"required,url"`
	APIKey    string  `json:"api_key" validate:"required"`
	AccountID string  `json:"account_id" validate:"required"`
	InboxID   *string `json:"inbox_id,omitempty"`
}

// UpdateChatwootConfigRequest represents a request to update Chatwoot configuration
type UpdateChatwootConfigRequest struct {
	URL       *string `json:"url,omitempty" validate:"omitempty,url"`
	APIKey    *string `json:"api_key,omitempty"`
	AccountID *string `json:"account_id,omitempty"`
	InboxID   *string `json:"inbox_id,omitempty"`
	Active    *bool   `json:"active,omitempty"`
}

// ChatwootContact represents a contact in Chatwoot
type ChatwootContact struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	PhoneNumber string                 `json:"phone_number"`
	Email       string                 `json:"email"`
	Attributes  map[string]interface{} `json:"custom_attributes"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ChatwootConversation represents a conversation in Chatwoot
type ChatwootConversation struct {
	ID        int       `json:"id"`
	ContactID int       `json:"contact_id"`
	InboxID   int       `json:"inbox_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ChatwootMessage represents a message in Chatwoot
type ChatwootMessage struct {
	ID             int                    `json:"id"`
	ConversationID int                    `json:"conversation_id"`
	Content        string                 `json:"content"`
	MessageType    string                 `json:"message_type"`
	ContentType    string                 `json:"content_type"`
	Attachments    []ChatwootAttachment   `json:"attachments"`
	Metadata       map[string]interface{} `json:"metadata"`
	CreatedAt      time.Time              `json:"created_at"`
}

// ChatwootAttachment represents an attachment in Chatwoot
type ChatwootAttachment struct {
	ID       int    `json:"id"`
	FileType string `json:"file_type"`
	FileURL  string `json:"data_url"`
	FileName string `json:"file_name"`
}

// ChatwootWebhookPayload represents incoming webhook payload from Chatwoot
type ChatwootWebhookPayload struct {
	Event        string                 `json:"event"`
	Account      ChatwootAccount        `json:"account"`
	Conversation ChatwootConversation   `json:"conversation"`
	Contact      ChatwootContact        `json:"contact"`
	Message      *ChatwootMessage       `json:"message,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ChatwootAccount represents account information in webhook
type ChatwootAccount struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// SyncContactRequest represents a request to sync contact with Chatwoot
type SyncContactRequest struct {
	PhoneNumber string                 `json:"phone_number" validate:"required"`
	Name        string                 `json:"name" validate:"required"`
	Email       string                 `json:"email,omitempty" validate:"omitempty,email"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

// SyncConversationRequest represents a request to sync conversation with Chatwoot
type SyncConversationRequest struct {
	ContactID int    `json:"contact_id" validate:"required"`
	SessionID string `json:"session_id" validate:"required"`
}

// SendMessageToChatwootRequest represents a request to send message to Chatwoot
type SendMessageToChatwootRequest struct {
	ConversationID int                    `json:"conversation_id" validate:"required"`
	Content        string                 `json:"content" validate:"required"`
	MessageType    string                 `json:"message_type" validate:"required,oneof=incoming outgoing"`
	ContentType    string                 `json:"content_type,omitempty"`
	Attachments    []ChatwootAttachment   `json:"attachments,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// NewChatwootConfig creates a new Chatwoot configuration
func NewChatwootConfig(url, apiKey, accountID string, inboxID *string) *ChatwootConfig {
	return &ChatwootConfig{
		ID:        uuid.New(),
		URL:       url,
		APIKey:    apiKey,
		AccountID: accountID,
		InboxID:   inboxID,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Update updates the Chatwoot configuration
func (c *ChatwootConfig) Update(req *UpdateChatwootConfigRequest) {
	if req.URL != nil {
		c.URL = *req.URL
	}
	if req.APIKey != nil {
		c.APIKey = *req.APIKey
	}
	if req.AccountID != nil {
		c.AccountID = *req.AccountID
	}
	if req.InboxID != nil {
		c.InboxID = req.InboxID
	}
	if req.Active != nil {
		c.Active = *req.Active
	}
	c.UpdatedAt = time.Now()
}

// IsConfigured returns true if the configuration has all required fields
func (c *ChatwootConfig) IsConfigured() bool {
	return c.URL != "" && c.APIKey != "" && c.AccountID != ""
}

// GetBaseURL returns the base URL for API calls
func (c *ChatwootConfig) GetBaseURL() string {
	return c.URL + "/accounts/" + c.AccountID
}

// ChatwootEventType represents the type of Chatwoot webhook event
type ChatwootEventType string

const (
	ChatwootEventConversationCreated       ChatwootEventType = "conversation_created"
	ChatwootEventConversationUpdated       ChatwootEventType = "conversation_updated"
	ChatwootEventConversationResolved      ChatwootEventType = "conversation_resolved"
	ChatwootEventMessageCreated            ChatwootEventType = "message_created"
	ChatwootEventMessageUpdated            ChatwootEventType = "message_updated"
	ChatwootEventContactCreated            ChatwootEventType = "contact_created"
	ChatwootEventContactUpdated            ChatwootEventType = "contact_updated"
	ChatwootEventConversationStatusChanged ChatwootEventType = "conversation_status_changed"
)

// IsValidChatwootEvent validates if the event type is supported
func IsValidChatwootEvent(eventType string) bool {
	switch ChatwootEventType(eventType) {
	case ChatwootEventConversationCreated,
		ChatwootEventConversationUpdated,
		ChatwootEventConversationResolved,
		ChatwootEventMessageCreated,
		ChatwootEventMessageUpdated,
		ChatwootEventContactCreated,
		ChatwootEventContactUpdated,
		ChatwootEventConversationStatusChanged:
		return true
	default:
		return false
	}
}
