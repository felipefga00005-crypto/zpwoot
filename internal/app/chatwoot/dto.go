package chatwoot

import (
	"time"
	"zpwoot/internal/domain/chatwoot"
)

type CreateChatwootConfigRequest struct {
	URL       string  `json:"url" validate:"required,url" example:"https://chatwoot.example.com"`
	APIKey    string  `json:"apiKey" validate:"required" example:"chatwoot-api-key-123"`
	AccountID string  `json:"accountId" validate:"required" example:"1"`
	InboxID   *string `json:"inboxId,omitempty" example:"1"`
} //@name CreateChatwootConfigRequest

type CreateChatwootConfigResponse struct {
	ID        string    `json:"id" example:"chatwoot-config-123"`
	URL       string    `json:"url" example:"https://chatwoot.example.com"`
	AccountID string    `json:"accountId" example:"1"`
	InboxID   *string   `json:"inboxId,omitempty" example:"1"`
	Active    bool      `json:"active" example:"true"`
	CreatedAt time.Time `json:"createdAt" example:"2024-01-01T00:00:00Z"`
} //@name CreateChatwootConfigResponse

type UpdateChatwootConfigRequest struct {
	URL       *string `json:"url,omitempty" validate:"omitempty,url" example:"https://new-chatwoot.example.com"`
	APIKey    *string `json:"api_key,omitempty" example:"new-api-key-123"`
	AccountID *string `json:"account_id,omitempty" example:"2"`
	InboxID   *string `json:"inbox_id,omitempty" example:"2"`
	Active    *bool   `json:"active,omitempty" example:"false"`
}

type ChatwootConfigResponse struct {
	ID        string    `json:"id" example:"chatwoot-config-123"`
	URL       string    `json:"url" example:"https://chatwoot.example.com"`
	AccountID string    `json:"accountId" example:"1"`
	InboxID   *string   `json:"inboxId,omitempty" example:"1"`
	Active    bool      `json:"active" example:"true"`
	CreatedAt time.Time `json:"createdAt" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updatedAt" example:"2024-01-01T00:00:00Z"`
} //@name ChatwootConfigResponse

type SyncContactRequest struct {
	PhoneNumber string                 `json:"phoneNumber" validate:"required" example:"+5511999999999"`
	Name        string                 `json:"name" validate:"required" example:"John Doe"`
	Email       string                 `json:"email,omitempty" validate:"omitempty,email" example:"john@example.com"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

type SyncContactResponse struct {
	ID          int                    `json:"id" example:"123"`
	PhoneNumber string                 `json:"phoneNumber" example:"+5511999999999"`
	Name        string                 `json:"name" example:"John Doe"`
	Email       string                 `json:"email,omitempty" example:"john@example.com"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
	CreatedAt   time.Time              `json:"createdAt" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time              `json:"updatedAt" example:"2024-01-01T00:00:00Z"`
}

type SyncConversationRequest struct {
	ContactID   int    `json:"contactId" validate:"required" example:"123"`
	SessionID   string `json:"sessionId" validate:"required" example:"session-123"`
	PhoneNumber string `json:"phoneNumber" validate:"required" example:"+5511999999999"`
}

type SyncConversationResponse struct {
	ID          int       `json:"id" example:"456"`
	ContactID   int       `json:"contactId" example:"123"`
	SessionID   string    `json:"sessionId" example:"session-123"`
	PhoneNumber string    `json:"phoneNumber" example:"+5511999999999"`
	Status      string    `json:"status" example:"open"`
	CreatedAt   time.Time `json:"createdAt" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time `json:"updatedAt" example:"2024-01-01T00:00:00Z"`
}

type SendMessageToChatwootRequest struct {
	ConversationID int                    `json:"conversationId" validate:"required" example:"456"`
	Content        string                 `json:"content" validate:"required" example:"Hello from Wameow!"`
	MessageType    string                 `json:"messageType" validate:"required,oneof=incoming outgoing" example:"incoming"`
	ContentType    string                 `json:"contentType,omitempty" example:"text"`
	Attachments    []ChatwootAttachment   `json:"attachments,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

type ChatwootAttachment struct {
	URL      string `json:"url" example:"https://example.com/image.jpg"`
	FileType string `json:"fileType" example:"image"`
	FileName string `json:"fileName" example:"image.jpg"`
}

type SendMessageToChatwootResponse struct {
	ID             int                    `json:"id" example:"789"`
	ConversationID int                    `json:"conversationId" example:"456"`
	Content        string                 `json:"content" example:"Hello from Wameow!"`
	MessageType    string                 `json:"messageType" example:"incoming"`
	ContentType    string                 `json:"contentType" example:"text"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"createdAt" example:"2024-01-01T00:00:00Z"`
}

type ChatwootWebhookPayload struct {
	Event        string                 `json:"event" example:"message_created"`
	Data         map[string]interface{} `json:"data"`
	Account      ChatwootAccount        `json:"account"`
	Conversation ChatwootConversation   `json:"conversation,omitempty"`
	Message      ChatwootMessage        `json:"message,omitempty"`
}

type ChatwootAccount struct {
	ID   int    `json:"id" example:"1"`
	Name string `json:"name" example:"My Company"`
}

type ChatwootConversation struct {
	ID     int    `json:"id" example:"456"`
	Status string `json:"status" example:"open"`
}

type ChatwootMessage struct {
	ID          int    `json:"id" example:"789"`
	Content     string `json:"content" example:"Hello!"`
	MessageType string `json:"messageType" example:"incoming"`
	ContentType string `json:"contentType" example:"text"`
}

type TestChatwootConnectionResponse struct {
	Success     bool   `json:"success" example:"true"`
	AccountName string `json:"accountName,omitempty" example:"My Company"`
	InboxName   string `json:"inboxName,omitempty" example:"Wameow Inbox"`
	Error       string `json:"error,omitempty"`
} // @name TestChatwootConnectionResponse

type ChatwootStatsResponse struct {
	TotalContacts       int `json:"totalContacts" example:"150"`
	TotalConversations  int `json:"totalConversations" example:"89"`
	ActiveConversations int `json:"activeConversations" example:"12"`
	MessagesSent        int `json:"messagesSent" example:"1250"`
	MessagesReceived    int `json:"messagesReceived" example:"890"`
} // @name ChatwootStatsResponse

func (r *CreateChatwootConfigRequest) ToCreateChatwootConfigRequest() *chatwoot.CreateChatwootConfigRequest {
	return &chatwoot.CreateChatwootConfigRequest{
		URL:       r.URL,
		APIKey:    r.APIKey,
		AccountID: r.AccountID,
		InboxID:   r.InboxID,
	}
}

func (r *UpdateChatwootConfigRequest) ToUpdateChatwootConfigRequest() *chatwoot.UpdateChatwootConfigRequest {
	return &chatwoot.UpdateChatwootConfigRequest{
		URL:       r.URL,
		APIKey:    r.APIKey,
		AccountID: r.AccountID,
		InboxID:   r.InboxID,
		Active:    r.Active,
	}
}

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
