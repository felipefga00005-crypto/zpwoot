package ports

import "zpwoot/internal/domain/chatwoot"

// ChatwootIntegration defines the basic interface for Chatwoot integration operations
type ChatwootIntegration interface {
	CreateContact(phoneNumber, name string) (*ChatwootContact, error)
	CreateConversation(contactID string, sessionID string) (*ChatwootConversation, error)
	SendMessage(conversationID, content, messageType string) error
	GetContact(phoneNumber string) (*ChatwootContact, error)
	GetConversation(conversationID string) (*ChatwootConversation, error)
	UpdateContactAttributes(contactID string, attributes map[string]interface{}) error
}

// ChatwootIntegrationExtended extends ChatwootIntegration with advanced operations
type ChatwootIntegrationExtended interface {
	ChatwootIntegration

	CreateInbox(name, channelType string) (*ChatwootInbox, error)
	GetInbox(inboxID int) (*ChatwootInbox, error)
	UpdateInbox(inboxID int, updates map[string]interface{}) error
	DeleteInbox(inboxID int) error

	GetAccount() (*ChatwootAccount, error)
	UpdateAccount(updates map[string]interface{}) error

	GetAgents() ([]*ChatwootAgent, error)
	GetAgent(agentID int) (*ChatwootAgent, error)
	AssignConversation(conversationID, agentID int) error
	UnassignConversation(conversationID int) error

	CreateLabel(name, description, color string) (*ChatwootLabel, error)
	GetLabels() ([]*ChatwootLabel, error)
	AddLabelToConversation(conversationID int, labelID int) error
	RemoveLabelFromConversation(conversationID int, labelID int) error

	CreateCustomAttribute(name, attributeType, description string) (*ChatwootCustomAttribute, error)
	GetCustomAttributes() ([]*ChatwootCustomAttribute, error)
	UpdateContactCustomAttribute(contactID int, attributeKey string, value interface{}) error

	SetConfig(url string, events []string) (*ChatwootWebhook, error)
	GetWebhooks() ([]*ChatwootWebhook, error)
	UpdateWebhook(webhookID int, updates map[string]interface{}) error
	DeleteWebhook(webhookID int) error

	GetConversationMetrics(from, to int64) (*ConversationMetrics, error)
	GetAgentMetrics(agentID int, from, to int64) (*AgentMetrics, error)
	GetAccountMetrics(from, to int64) (*AccountMetrics, error)

	BulkCreateContacts(contacts []*chatwoot.ChatwootContact) ([]*chatwoot.ChatwootContact, error)
	BulkUpdateContacts(updates []ContactUpdate) error
	BulkDeleteContacts(contactIDs []int) error
}

// ChatwootContact represents a contact in Chatwoot
type ChatwootContact struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phoneNumber"`
	Email       string `json:"email"`
}

// ChatwootConversation represents a conversation in Chatwoot
type ChatwootConversation struct {
	ID        int    `json:"id"`
	ContactID int    `json:"contact_id"`
	InboxID   int    `json:"inbox_id"`
	Status    string `json:"status"`
}

// ChatwootInbox represents an inbox in Chatwoot
type ChatwootInbox struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	ChannelType string `json:"channel_type"`
	AccountID   int    `json:"account_id"`
	WebsiteURL  string `json:"website_url,omitempty"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

// ChatwootAccount represents an account in Chatwoot
type ChatwootAccount struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Locale       string `json:"locale"`
	Domain       string `json:"domain,omitempty"`
	SupportEmail string `json:"support_email,omitempty"`
}

// ChatwootAgent represents an agent in Chatwoot
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

// ChatwootLabel represents a label in Chatwoot
type ChatwootLabel struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Color         string `json:"color"`
	ShowOnSidebar bool   `json:"show_on_sidebar"`
}

// ChatwootCustomAttribute represents a custom attribute in Chatwoot
type ChatwootCustomAttribute struct {
	ID             int    `json:"id"`
	AttributeKey   string `json:"attribute_key"`
	AttributeType  string `json:"attribute_type"`
	Description    string `json:"description"`
	DefaultValue   string `json:"default_value,omitempty"`
	AttributeModel string `json:"attribute_model"`
}

// ChatwootWebhook represents a webhook configuration in Chatwoot
type ChatwootWebhook struct {
	ID        int      `json:"id"`
	URL       string   `json:"url"`
	Events    []string `json:"events"`
	AccountID int      `json:"account_id"`
}

// ConversationMetrics represents conversation metrics from Chatwoot
type ConversationMetrics struct {
	TotalConversations    int     `json:"total_conversations"`
	OpenConversations     int     `json:"open_conversations"`
	ResolvedConversations int     `json:"resolved_conversations"`
	AverageResolutionTime float64 `json:"average_resolution_time"`
	AverageResponseTime   float64 `json:"average_response_time"`
	From                  int64   `json:"from"`
	To                    int64   `json:"to"`
}

// AgentMetrics represents agent performance metrics from Chatwoot
type AgentMetrics struct {
	AgentID               int     `json:"agent_id"`
	ConversationsHandled  int     `json:"conversations_handled"`
	ConversationsResolved int     `json:"conversations_resolved"`
	AverageResponseTime   float64 `json:"average_response_time"`
	MessagesSent          int     `json:"messages_sent"`
	From                  int64   `json:"from"`
	To                    int64   `json:"to"`
}

// AccountMetrics represents account-level metrics from Chatwoot
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

// ContactUpdate represents an update operation for a contact
type ContactUpdate struct {
	ID      int                    `json:"id"`
	Updates map[string]interface{} `json:"updates"`
}
