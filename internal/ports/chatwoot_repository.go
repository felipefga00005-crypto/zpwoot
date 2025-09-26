package ports

import (
	"context"

	"zpwoot/internal/domain/chatwoot"
)

// ChatwootRepository defines the interface for Chatwoot data persistence
type ChatwootRepository interface {
	// Config operations
	CreateConfig(ctx context.Context, config *chatwoot.ChatwootConfig) error
	GetConfig(ctx context.Context) (*chatwoot.ChatwootConfig, error)
	UpdateConfig(ctx context.Context, config *chatwoot.ChatwootConfig) error
	DeleteConfig(ctx context.Context) error

	// Contact operations
	CreateContact(ctx context.Context, contact *chatwoot.ChatwootContact) error
	GetContactByID(ctx context.Context, id int) (*chatwoot.ChatwootContact, error)
	GetContactByPhone(ctx context.Context, phoneNumber string) (*chatwoot.ChatwootContact, error)
	UpdateContact(ctx context.Context, contact *chatwoot.ChatwootContact) error
	DeleteContact(ctx context.Context, id int) error
	ListContacts(ctx context.Context, limit, offset int) ([]*chatwoot.ChatwootContact, int, error)

	// Conversation operations
	CreateConversation(ctx context.Context, conversation *chatwoot.ChatwootConversation) error
	GetConversationByID(ctx context.Context, id int) (*chatwoot.ChatwootConversation, error)
	GetConversationByContactID(ctx context.Context, contactID int) (*chatwoot.ChatwootConversation, error)
	GetConversationBySessionID(ctx context.Context, sessionID string) (*chatwoot.ChatwootConversation, error)
	UpdateConversation(ctx context.Context, conversation *chatwoot.ChatwootConversation) error
	DeleteConversation(ctx context.Context, id int) error
	ListConversations(ctx context.Context, limit, offset int) ([]*chatwoot.ChatwootConversation, int, error)
	GetActiveConversations(ctx context.Context) ([]*chatwoot.ChatwootConversation, error)

	// Message operations
	CreateMessage(ctx context.Context, message *chatwoot.ChatwootMessage) error
	GetMessageByID(ctx context.Context, id int) (*chatwoot.ChatwootMessage, error)
	GetMessagesByConversationID(ctx context.Context, conversationID int, limit, offset int) ([]*chatwoot.ChatwootMessage, error)
	UpdateMessage(ctx context.Context, message *chatwoot.ChatwootMessage) error
	DeleteMessage(ctx context.Context, id int) error

	// Sync operations
	CreateSyncRecord(ctx context.Context, record *SyncRecord) error
	GetSyncRecord(ctx context.Context, sessionID, recordType, externalID string) (*SyncRecord, error)
	UpdateSyncRecord(ctx context.Context, record *SyncRecord) error
	DeleteSyncRecord(ctx context.Context, id string) error
	GetSyncRecordsBySession(ctx context.Context, sessionID string) ([]*SyncRecord, error)

	// Statistics operations
	GetContactCount(ctx context.Context) (int, error)
	GetConversationCount(ctx context.Context) (int, error)
	GetActiveConversationCount(ctx context.Context) (int, error)
	GetMessageCount(ctx context.Context) (int, error)
	GetMessageCountByType(ctx context.Context, messageType string) (int, error)
	GetStatsForPeriod(ctx context.Context, from, to int64) (*ChatwootStats, error)
}

// SyncRecord represents a synchronization record between Wameow and Chatwoot
type SyncRecord struct {
	ID           string `json:"id" db:"id"`
	SessionID    string `json:"session_id" db:"session_id"`
	RecordType   string `json:"record_type" db:"record_type"` // contact, conversation, message
	ExternalID   string `json:"external_id" db:"external_id"` // Wameow ID
	ChatwootID   int    `json:"chatwoot_id" db:"chatwoot_id"` // Chatwoot ID
	PhoneNumber  string `json:"phone_number,omitempty" db:"phone_number"`
	LastSyncAt   int64  `json:"last_sync_at" db:"last_sync_at"`
	SyncStatus   string `json:"sync_status" db:"sync_status"` // pending, synced, failed
	ErrorMessage string `json:"error_message,omitempty" db:"error_message"`
	CreatedAt    int64  `json:"created_at" db:"created_at"`
	UpdatedAt    int64  `json:"updated_at" db:"updated_at"`
}

// ChatwootStats represents Chatwoot integration statistics
type ChatwootStats struct {
	TotalContacts       int   `json:"total_contacts"`
	TotalConversations  int   `json:"total_conversations"`
	ActiveConversations int   `json:"active_conversations"`
	MessagesSent        int64 `json:"messages_sent"`
	MessagesReceived    int64 `json:"messages_received"`
	LastSyncAt          int64 `json:"last_sync_at"`
	SyncErrors          int   `json:"sync_errors"`
	From                int64 `json:"from"`
	To                  int64 `json:"to"`
}

// ChatwootIntegrationExtended extends the basic ChatwootIntegration interface
type ChatwootIntegrationExtended interface {
	ChatwootIntegration

	// Advanced operations
	CreateInbox(name, channelType string) (*ChatwootInbox, error)
	GetInbox(inboxID int) (*ChatwootInbox, error)
	UpdateInbox(inboxID int, updates map[string]interface{}) error
	DeleteInbox(inboxID int) error

	// Account operations
	GetAccount() (*ChatwootAccount, error)
	UpdateAccount(updates map[string]interface{}) error

	// Agent operations
	GetAgents() ([]*ChatwootAgent, error)
	GetAgent(agentID int) (*ChatwootAgent, error)
	AssignConversation(conversationID, agentID int) error
	UnassignConversation(conversationID int) error

	// Label operations
	CreateLabel(name, description, color string) (*ChatwootLabel, error)
	GetLabels() ([]*ChatwootLabel, error)
	AddLabelToConversation(conversationID int, labelID int) error
	RemoveLabelFromConversation(conversationID int, labelID int) error

	// Custom attributes
	CreateCustomAttribute(name, attributeType, description string) (*ChatwootCustomAttribute, error)
	GetCustomAttributes() ([]*ChatwootCustomAttribute, error)
	UpdateContactCustomAttribute(contactID int, attributeKey string, value interface{}) error

	// Webhook operations
	SetConfig(url string, events []string) (*ChatwootWebhook, error)
	GetWebhooks() ([]*ChatwootWebhook, error)
	UpdateWebhook(webhookID int, updates map[string]interface{}) error
	DeleteWebhook(webhookID int) error

	// Reporting
	GetConversationMetrics(from, to int64) (*ConversationMetrics, error)
	GetAgentMetrics(agentID int, from, to int64) (*AgentMetrics, error)
	GetAccountMetrics(from, to int64) (*AccountMetrics, error)

	// Bulk operations
	BulkCreateContacts(contacts []*chatwoot.ChatwootContact) ([]*chatwoot.ChatwootContact, error)
	BulkUpdateContacts(updates []ContactUpdate) error
	BulkDeleteContacts(contactIDs []int) error
}

// Extended Chatwoot types
type ChatwootInbox struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	ChannelType string `json:"channel_type"`
	AccountID   int    `json:"account_id"`
	WebsiteURL  string `json:"website_url,omitempty"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

type ChatwootAccount struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Locale       string `json:"locale"`
	Domain       string `json:"domain,omitempty"`
	SupportEmail string `json:"support_email,omitempty"`
}

type ChatwootAgent struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AccountID int    `json:"account_id"`
	Role      string `json:"role"`
	Confirmed bool   `json:"confirmed"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Available bool   `json:"available"`
}

type ChatwootLabel struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Color         string `json:"color"`
	ShowOnSidebar bool   `json:"show_on_sidebar"`
}

type ChatwootCustomAttribute struct {
	ID             int    `json:"id"`
	AttributeKey   string `json:"attribute_key"`
	AttributeType  string `json:"attribute_type"`
	Description    string `json:"description"`
	DefaultValue   string `json:"default_value,omitempty"`
	AttributeModel string `json:"attribute_model"`
}

type ChatwootWebhook struct {
	ID        int      `json:"id"`
	URL       string   `json:"url"`
	Events    []string `json:"events"`
	AccountID int      `json:"account_id"`
}

type ConversationMetrics struct {
	TotalConversations    int     `json:"total_conversations"`
	OpenConversations     int     `json:"open_conversations"`
	ResolvedConversations int     `json:"resolved_conversations"`
	AverageResolutionTime float64 `json:"average_resolution_time"`
	AverageResponseTime   float64 `json:"average_response_time"`
	From                  int64   `json:"from"`
	To                    int64   `json:"to"`
}

type AgentMetrics struct {
	AgentID               int     `json:"agent_id"`
	ConversationsHandled  int     `json:"conversations_handled"`
	ConversationsResolved int     `json:"conversations_resolved"`
	AverageResponseTime   float64 `json:"average_response_time"`
	MessagesSent          int     `json:"messages_sent"`
	From                  int64   `json:"from"`
	To                    int64   `json:"to"`
}

type AccountMetrics struct {
	TotalContacts         int     `json:"total_contacts"`
	TotalConversations    int     `json:"total_conversations"`
	TotalMessages         int     `json:"total_messages"`
	ActiveAgents          int     `json:"active_agents"`
	AverageResolutionTime float64 `json:"average_resolution_time"`
	CustomerSatisfaction  float64 `json:"customer_satisfaction"`
	From                  int64   `json:"from"`
	To                    int64   `json:"to"`
}

type ContactUpdate struct {
	ID      int                    `json:"id"`
	Updates map[string]interface{} `json:"updates"`
}
