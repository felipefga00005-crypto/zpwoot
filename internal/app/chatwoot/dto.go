package chatwoot

import (
	"time"
	"zpwoot/internal/domain/chatwoot"
)

// CreateChatwootConfigRequest represents the request to create Chatwoot configuration
type CreateChatwootConfigRequest struct {
	URL       string  `json:"url" validate:"required,url" example:"https://chatwoot.example.com"`
	APIKey    string  `json:"apiKey" validate:"required" example:"chatwoot-api-key-123"`
	AccountID string  `json:"accountId" validate:"required" example:"1"`
	InboxID   *string `json:"inboxId,omitempty" example:"1"`
} //@name CreateChatwootConfigRequest

// CreateChatwootConfigResponse represents the response after creating Chatwoot configuration
type CreateChatwootConfigResponse struct {
	ID        string    `json:"id" example:"chatwoot-config-123"`
	URL       string    `json:"url" example:"https://chatwoot.example.com"`
	AccountID string    `json:"accountId" example:"1"`
	InboxID   *string   `json:"inboxId,omitempty" example:"1"`
	Active    bool      `json:"active" example:"true"`
	CreatedAt time.Time `json:"createdAt" example:"2024-01-01T00:00:00Z"`
} //@name CreateChatwootConfigResponse

// UpdateChatwootConfigRequest represents the request to update Chatwoot configuration
type UpdateChatwootConfigRequest struct {
	URL       *string `json:"url,omitempty" validate:"omitempty,url" example:"https://new-chatwoot.example.com"`
	APIKey    *string `json:"api_key,omitempty" example:"new-api-key-123"`
	AccountID *string `json:"account_id,omitempty" example:"2"`
	InboxID   *string `json:"inbox_id,omitempty" example:"2"`
	Active    *bool   `json:"active,omitempty" example:"false"`
}

// ChatwootConfigResponse represents Chatwoot configuration in responses
type ChatwootConfigResponse struct {
	ID        string    `json:"id" example:"chatwoot-config-123"`
	URL       string    `json:"url" example:"https://chatwoot.example.com"`
	AccountID string    `json:"account_id" example:"1"`
	InboxID   *string   `json:"inbox_id,omitempty" example:"1"`
	Active    bool      `json:"active" example:"true"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// SyncContactRequest represents the request to sync a contact with Chatwoot
type SyncContactRequest struct {
	PhoneNumber string                 `json:"phone_number" validate:"required" example:"+5511999999999"`
	Name        string                 `json:"name" validate:"required" example:"John Doe"`
	Email       string                 `json:"email,omitempty" validate:"omitempty,email" example:"john@example.com"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

// SyncContactResponse represents the response after syncing a contact
type SyncContactResponse struct {
	ID          int                    `json:"id" example:"123"`
	PhoneNumber string                 `json:"phone_number" example:"+5511999999999"`
	Name        string                 `json:"name" example:"John Doe"`
	Email       string                 `json:"email,omitempty" example:"john@example.com"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
	CreatedAt   time.Time              `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time              `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// SyncConversationRequest represents the request to sync a conversation with Chatwoot
type SyncConversationRequest struct {
	ContactID   int    `json:"contact_id" validate:"required" example:"123"`
	SessionID   string `json:"session_id" validate:"required" example:"session-123"`
	PhoneNumber string `json:"phone_number" validate:"required" example:"+5511999999999"`
}

// SyncConversationResponse represents the response after syncing a conversation
type SyncConversationResponse struct {
	ID          int       `json:"id" example:"456"`
	ContactID   int       `json:"contact_id" example:"123"`
	SessionID   string    `json:"session_id" example:"session-123"`
	PhoneNumber string    `json:"phone_number" example:"+5511999999999"`
	Status      string    `json:"status" example:"open"`
	CreatedAt   time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// SendMessageToChatwootRequest represents the request to send a message to Chatwoot
type SendMessageToChatwootRequest struct {
	ConversationID int                    `json:"conversation_id" validate:"required" example:"456"`
	Content        string                 `json:"content" validate:"required" example:"Hello from Wameow!"`
	MessageType    string                 `json:"message_type" validate:"required,oneof=incoming outgoing" example:"incoming"`
	ContentType    string                 `json:"content_type,omitempty" example:"text"`
	Attachments    []ChatwootAttachment   `json:"attachments,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ChatwootAttachment represents an attachment in Chatwoot messages
type ChatwootAttachment struct {
	URL      string `json:"url" example:"https://example.com/image.jpg"`
	FileType string `json:"file_type" example:"image"`
	FileName string `json:"file_name" example:"image.jpg"`
}

// SendMessageToChatwootResponse represents the response after sending a message to Chatwoot
type SendMessageToChatwootResponse struct {
	ID             int                    `json:"id" example:"789"`
	ConversationID int                    `json:"conversation_id" example:"456"`
	Content        string                 `json:"content" example:"Hello from Wameow!"`
	MessageType    string                 `json:"message_type" example:"incoming"`
	ContentType    string                 `json:"content_type" example:"text"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

// ChatwootWebhookPayload represents the payload received from Chatwoot webhooks
type ChatwootWebhookPayload struct {
	Event        string                 `json:"event" example:"message_created"`
	Data         map[string]interface{} `json:"data"`
	Account      ChatwootAccount        `json:"account"`
	Conversation ChatwootConversation   `json:"conversation,omitempty"`
	Message      ChatwootMessage        `json:"message,omitempty"`
}

// ChatwootAccount represents account information in webhook
type ChatwootAccount struct {
	ID   int    `json:"id" example:"1"`
	Name string `json:"name" example:"My Company"`
}

// ChatwootConversation represents conversation information in webhook
type ChatwootConversation struct {
	ID     int    `json:"id" example:"456"`
	Status string `json:"status" example:"open"`
}

// ChatwootMessage represents message information in webhook
type ChatwootMessage struct {
	ID          int    `json:"id" example:"789"`
	Content     string `json:"content" example:"Hello!"`
	MessageType string `json:"message_type" example:"incoming"`
	ContentType string `json:"content_type" example:"text"`
}

// TestChatwootConnectionResponse represents the response after testing Chatwoot connection
type TestChatwootConnectionResponse struct {
	Success     bool   `json:"success" example:"true"`
	AccountName string `json:"account_name,omitempty" example:"My Company"`
	InboxName   string `json:"inbox_name,omitempty" example:"Wameow Inbox"`
	Error       string `json:"error,omitempty"`
} // @name TestChatwootConnectionResponse

// ChatwootStatsResponse represents Chatwoot integration statistics
type ChatwootStatsResponse struct {
	TotalContacts       int `json:"total_contacts" example:"150"`
	TotalConversations  int `json:"total_conversations" example:"89"`
	ActiveConversations int `json:"active_conversations" example:"12"`
	MessagesSent        int `json:"messages_sent" example:"1250"`
	MessagesReceived    int `json:"messages_received" example:"890"`
} // @name ChatwootStatsResponse

// Conversion methods

// ToCreateChatwootConfigRequest converts to domain request
func (r *CreateChatwootConfigRequest) ToCreateChatwootConfigRequest() *chatwoot.CreateChatwootConfigRequest {
	return &chatwoot.CreateChatwootConfigRequest{
		URL:       r.URL,
		APIKey:    r.APIKey,
		AccountID: r.AccountID,
		InboxID:   r.InboxID,
	}
}

// ToUpdateChatwootConfigRequest converts to domain request
func (r *UpdateChatwootConfigRequest) ToUpdateChatwootConfigRequest() *chatwoot.UpdateChatwootConfigRequest {
	return &chatwoot.UpdateChatwootConfigRequest{
		URL:       r.URL,
		APIKey:    r.APIKey,
		AccountID: r.AccountID,
		InboxID:   r.InboxID,
		Active:    r.Active,
	}
}

// FromChatwootConfig converts from domain config to response
func FromChatwootConfig(c *chatwoot.ChatwootConfig) *ChatwootConfigResponse {
	return &ChatwootConfigResponse{
		ID:        c.ID.String(),
		URL:       c.URL,
		AccountID: c.AccountID,
		InboxID:   c.InboxID,
		Active:    c.Active,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}
