package ports

import (
	"context"

	"zpwoot/internal/domain/message"
	"zpwoot/internal/domain/session"
)

// SessionRepository defines the interface for session data persistence
type SessionRepository interface {
	// Create creates a new session
	Create(ctx context.Context, session *session.Session) error

	// GetByID retrieves a session by its ID
	GetByID(ctx context.Context, id string) (*session.Session, error)

	// GetByName retrieves a session by its name
	GetByName(ctx context.Context, name string) (*session.Session, error)

	// GetByDeviceJid retrieves a session by device JID
	GetByDeviceJid(ctx context.Context, deviceJid string) (*session.Session, error)

	// List retrieves sessions with optional filters
	List(ctx context.Context, req *session.ListSessionsRequest) ([]*session.Session, int, error)

	// Update updates an existing session
	Update(ctx context.Context, session *session.Session) error

	// Delete removes a session by ID
	Delete(ctx context.Context, id string) error

	// UpdateConnectionStatus updates the connection status of a session
	UpdateConnectionStatus(ctx context.Context, id string, isConnected bool) error

	// UpdateLastSeen updates the last seen timestamp
	UpdateLastSeen(ctx context.Context, id string) error

	// GetActiveSessions retrieves all connected sessions
	GetActiveSessions(ctx context.Context) ([]*session.Session, error)

	// CountByConnectionStatus counts sessions by connection status
	CountByConnectionStatus(ctx context.Context, isConnected bool) (int, error)
}

// WameowManager defines the interface for Wameow operations
type WameowManager interface {
	// CreateSession initializes a new Wameow session
	CreateSession(sessionID string, config *session.ProxyConfig) error

	// ConnectSession establishes connection to Wameow
	ConnectSession(sessionID string) error

	// DisconnectSession disconnects from Wameow
	DisconnectSession(sessionID string) error

	// LogoutSession logs out from Wameow
	LogoutSession(sessionID string) error

	// GetQRCode retrieves the current QR code for pairing
	GetQRCode(sessionID string) (*session.QRCodeResponse, error)

	// PairPhone pairs a phone number with the session
	PairPhone(sessionID, phoneNumber string) error

	// IsConnected checks if the session is connected
	IsConnected(sessionID string) bool

	// GetDeviceInfo retrieves device information
	GetDeviceInfo(sessionID string) (*session.DeviceInfo, error)

	// SetProxy configures proxy for the session
	SetProxy(sessionID string, config *session.ProxyConfig) error

	// GetProxy retrieves proxy configuration
	GetProxy(sessionID string) (*session.ProxyConfig, error)

	// SendMessage sends a message through Wameow with full support for all message types
	SendMessage(sessionID, to, messageType, body, caption, file, filename string, latitude, longitude float64, contactName, contactPhone string) (*message.SendResult, error)

	// SendMediaMessage sends a media message
	SendMediaMessage(sessionID, to string, media []byte, mediaType, caption string) error

	// SendButtonMessage sends a button message
	SendButtonMessage(sessionID, to, body string, buttons []map[string]string) (*message.SendResult, error)

	// SendListMessage sends a list message
	SendListMessage(sessionID, to, body, buttonText string, sections []map[string]interface{}) (*message.SendResult, error)

	// SendReaction sends a reaction to a message
	SendReaction(sessionID, to, messageID, reaction string) error

	// SendPresence sends presence information
	SendPresence(sessionID, to, presence string) error

	// EditMessage edits an existing message
	EditMessage(sessionID, to, messageID, newText string) error

	// DeleteMessage deletes an existing message
	DeleteMessage(sessionID, to, messageID string, forAll bool) error

	// GetSessionStats retrieves session statistics
	GetSessionStats(sessionID string) (*SessionStats, error)

	// RegisterEventHandler registers an event handler for Wameow events
	RegisterEventHandler(sessionID string, handler EventHandler) error

	// UnregisterEventHandler removes an event handler
	UnregisterEventHandler(sessionID string, handlerID string) error
}

// SessionStats represents session statistics
type SessionStats struct {
	MessagesSent     int64 `json:"messages_sent"`
	MessagesReceived int64 `json:"messages_received"`
	LastActivity     int64 `json:"last_activity"`
	Uptime           int64 `json:"uptime"`
}

// EventHandler defines the interface for handling Wameow events
type EventHandler interface {
	HandleMessage(sessionID string, message *WameowMessage) error
	HandleConnection(sessionID string, connected bool) error
	HandleQRCode(sessionID string, qrCode string) error
	HandlePairSuccess(sessionID string) error
	HandleError(sessionID string, err error) error
}

// WameowMessage represents a Wameow message
type WameowMessage struct {
	ID        string `json:"id"`
	From      string `json:"from"`
	To        string `json:"to"`
	Body      string `json:"body"`
	Timestamp int64  `json:"timestamp"`
	Type      string `json:"type"`
	MediaURL  string `json:"media_url,omitempty"`
	Caption   string `json:"caption,omitempty"`
}

// ChatwootIntegration defines the interface for Chatwoot integration
type ChatwootIntegration interface {
	// CreateContact creates a contact in Chatwoot
	CreateContact(phoneNumber, name string) (*ChatwootContact, error)

	// CreateConversation creates a conversation in Chatwoot
	CreateConversation(contactID string, sessionID string) (*ChatwootConversation, error)

	// SendMessage sends a message to Chatwoot
	SendMessage(conversationID, content, messageType string) error

	// GetContact retrieves a contact by phone number
	GetContact(phoneNumber string) (*ChatwootContact, error)

	// GetConversation retrieves a conversation by ID
	GetConversation(conversationID string) (*ChatwootConversation, error)

	// UpdateContactAttributes updates contact attributes
	UpdateContactAttributes(contactID string, attributes map[string]interface{}) error
}

// ChatwootContact represents a Chatwoot contact
type ChatwootContact struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phoneNumber"`
	Email       string `json:"email"`
}

// ChatwootConversation represents a Chatwoot conversation
type ChatwootConversation struct {
	ID        int    `json:"id"`
	ContactID int    `json:"contact_id"`
	InboxID   int    `json:"inbox_id"`
	Status    string `json:"status"`
}

// WebhookService defines the interface for webhook operations
type WebhookService interface {
	// SendWebhook sends a webhook notification
	SendWebhook(url string, payload interface{}) error

	// RegisterWebhook registers a webhook URL for events
	RegisterWebhook(sessionID, url, secret string, events []string) error

	// UnregisterWebhook removes a webhook registration
	UnregisterWebhook(sessionID, url string) error

	// GetWebhooks retrieves registered webhooks for a session
	GetWebhooks(sessionID string) ([]*WebhookRegistration, error)
}

// WebhookRegistration represents a webhook registration
type WebhookRegistration struct {
	ID        string   `json:"id"`
	SessionID string   `json:"session_id"`
	URL       string   `json:"url"`
	Secret    string   `json:"secret"`
	Events    []string `json:"events"`
	Active    bool     `json:"active"`
}
