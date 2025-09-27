package chatwoot

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

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

var (
	ErrConfigNotFound       = errors.New("chatwoot config not found")
	ErrContactNotFound      = errors.New("chatwoot contact not found")
	ErrConversationNotFound = errors.New("chatwoot conversation not found")
	ErrMessageNotFound      = errors.New("chatwoot message not found")
	ErrInvalidAPIKey        = errors.New("invalid chatwoot API key")
	ErrInvalidAccountID     = errors.New("invalid chatwoot account ID")
	ErrChatwootAPIError     = errors.New("chatwoot API error")
)

type CreateChatwootConfigRequest struct {
	URL       string  `json:"url" validate:"required,url"`
	APIKey    string  `json:"api_key" validate:"required"`
	AccountID string  `json:"account_id" validate:"required"`
	InboxID   *string `json:"inbox_id,omitempty"`
}

type UpdateChatwootConfigRequest struct {
	URL       *string `json:"url,omitempty" validate:"omitempty,url"`
	APIKey    *string `json:"api_key,omitempty"`
	AccountID *string `json:"account_id,omitempty"`
	InboxID   *string `json:"inbox_id,omitempty"`
	Active    *bool   `json:"active,omitempty"`
}

type ChatwootContact struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	PhoneNumber string                 `json:"phone_number"`
	Email       string                 `json:"email"`
	Attributes  map[string]interface{} `json:"custom_attributes"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type ChatwootConversation struct {
	ID        int       `json:"id"`
	ContactID int       `json:"contact_id"`
	InboxID   int       `json:"inbox_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

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

type ChatwootAttachment struct {
	ID       int    `json:"id"`
	FileType string `json:"file_type"`
	FileURL  string `json:"data_url"`
	FileName string `json:"file_name"`
}

type ChatwootWebhookPayload struct {
	Event        string                 `json:"event"`
	Account      ChatwootAccount        `json:"account"`
	Conversation ChatwootConversation   `json:"conversation"`
	Contact      ChatwootContact        `json:"contact"`
	Message      *ChatwootMessage       `json:"message,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type ChatwootAccount struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type SyncContactRequest struct {
	PhoneNumber string                 `json:"phone_number" validate:"required"`
	Name        string                 `json:"name" validate:"required"`
	Email       string                 `json:"email,omitempty" validate:"omitempty,email"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

type SyncConversationRequest struct {
	ContactID int    `json:"contact_id" validate:"required"`
	SessionID string `json:"session_id" validate:"required"`
}

type SendMessageToChatwootRequest struct {
	ConversationID int                    `json:"conversation_id" validate:"required"`
	Content        string                 `json:"content" validate:"required"`
	MessageType    string                 `json:"message_type" validate:"required,oneof=incoming outgoing"`
	ContentType    string                 `json:"content_type,omitempty"`
	Attachments    []ChatwootAttachment   `json:"attachments,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

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

func (c *ChatwootConfig) IsConfigured() bool {
	return c.URL != "" && c.APIKey != "" && c.AccountID != ""
}

func (c *ChatwootConfig) GetBaseURL() string {
	return c.URL + "/accounts/" + c.AccountID
}

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
