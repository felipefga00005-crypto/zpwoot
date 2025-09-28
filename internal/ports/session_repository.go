package ports

import (
	"context"
	"time"

	"go.mau.fi/whatsmeow/types"
	"zpwoot/internal/domain/message"
	"zpwoot/internal/domain/session"
)

type SessionRepository interface {
	Create(ctx context.Context, session *session.Session) error

	GetByID(ctx context.Context, id string) (*session.Session, error)

	GetByName(ctx context.Context, name string) (*session.Session, error)

	GetByDeviceJid(ctx context.Context, deviceJid string) (*session.Session, error)

	List(ctx context.Context, req *session.ListSessionsRequest) ([]*session.Session, int, error)

	Update(ctx context.Context, session *session.Session) error

	Delete(ctx context.Context, id string) error

	UpdateConnectionStatus(ctx context.Context, id string, isConnected bool) error

	UpdateLastSeen(ctx context.Context, id string) error

	GetActiveSessions(ctx context.Context) ([]*session.Session, error)

	CountByConnectionStatus(ctx context.Context, isConnected bool) (int, error)
}

type WameowManager interface {
	CreateSession(sessionID string, config *session.ProxyConfig) error

	ConnectSession(sessionID string) error

	DisconnectSession(sessionID string) error

	LogoutSession(sessionID string) error

	GetQRCode(sessionID string) (*session.QRCodeResponse, error)

	PairPhone(sessionID, phoneNumber string) error

	IsConnected(sessionID string) bool

	GetDeviceInfo(sessionID string) (*session.DeviceInfo, error)

	SetProxy(sessionID string, config *session.ProxyConfig) error

	GetProxy(sessionID string) (*session.ProxyConfig, error)

	GetUserJID(sessionID string) (string, error)

	SendMessage(sessionID, to, messageType, body, caption, file, filename string, latitude, longitude float64, contactName, contactPhone string, contextInfo *message.ContextInfo) (*message.SendResult, error)

	SendMediaMessage(sessionID, to string, media []byte, mediaType, caption string) error

	SendButtonMessage(sessionID, to, body string, buttons []map[string]string) (*message.SendResult, error)

	SendListMessage(sessionID, to, body, buttonText string, sections []map[string]interface{}) (*message.SendResult, error)

	SendReaction(sessionID, to, messageID, reaction string) error

	SendPresence(sessionID, to, presence string) error

	EditMessage(sessionID, to, messageID, newText string) error

	MarkRead(sessionID, to, messageID string) error

	// Advanced message operations
	RevokeMessage(sessionID, to, messageID string) (*message.SendResult, error)

	// Group management methods
	CreateGroup(sessionID, name string, participants []string, description string) (*GroupInfo, error)

	GetGroupInfo(sessionID, groupJID string) (*GroupInfo, error)

	ListJoinedGroups(sessionID string) ([]*GroupInfo, error)

	UpdateGroupParticipants(sessionID, groupJID string, participants []string, action string) ([]string, []string, error)

	SetGroupName(sessionID, groupJID, name string) error

	SetGroupDescription(sessionID, groupJID, description string) error

	SetGroupPhoto(sessionID, groupJID string, photo []byte) error

	GetGroupInviteLink(sessionID, groupJID string, reset bool) (string, error)

	JoinGroupViaLink(sessionID, inviteLink string) (*GroupInfo, error)

	LeaveGroup(sessionID, groupJID string) error

	UpdateGroupSettings(sessionID, groupJID string, announce, locked *bool) error

	GetGroupRequestParticipants(sessionID, groupJID string) ([]types.GroupParticipantRequest, error)

	UpdateGroupRequestParticipants(sessionID, groupJID string, participants []string, action string) ([]string, []string, error)

	SetGroupJoinApprovalMode(sessionID, groupJID string, requireApproval bool) error

	SetGroupMemberAddMode(sessionID, groupJID string, mode string) error

	GetSessionStats(sessionID string) (*SessionStats, error)

	RegisterEventHandler(sessionID string, handler EventHandler) error

	UnregisterEventHandler(sessionID string, handlerID string) error

	// Contact-related methods for WhatsApp client operations
	IsOnWhatsApp(ctx context.Context, sessionID string, phoneNumbers []string) (map[string]interface{}, error)
	GetProfilePictureInfo(ctx context.Context, sessionID, jid string, preview bool) (map[string]interface{}, error)
	GetUserInfo(ctx context.Context, sessionID string, jids []string) ([]map[string]interface{}, error)
	GetBusinessProfile(ctx context.Context, sessionID, jid string) (map[string]interface{}, error)
	GetAllContacts(ctx context.Context, sessionID string) (map[string]interface{}, error)
}

type GroupInfo struct {
	GroupJID     string             `json:"groupJid"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Owner        string             `json:"owner"`
	Participants []GroupParticipant `json:"participants"`
	Settings     GroupSettings      `json:"settings"`
	CreatedAt    time.Time          `json:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt"`
}

type GroupParticipant struct {
	JID          string `json:"jid"`
	IsAdmin      bool   `json:"isAdmin"`
	IsSuperAdmin bool   `json:"isSuperAdmin"`
}

type GroupSettings struct {
	Announce bool `json:"announce"`
	Locked   bool `json:"locked"`
}

type SessionStats struct {
	MessagesSent     int64 `json:"messages_sent"`
	MessagesReceived int64 `json:"messages_received"`
	LastActivity     int64 `json:"last_activity"`
	Uptime           int64 `json:"uptime"`
}

type EventHandler interface {
	HandleMessage(sessionID string, message *WameowMessage) error
	HandleConnection(sessionID string, connected bool) error
	HandleQRCode(sessionID string, qrCode string) error
	HandlePairSuccess(sessionID string) error
	HandleError(sessionID string, err error) error
}

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

type MessageInfo struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Chat      string    `json:"chat"`
}

type ChatwootIntegration interface {
	CreateContact(phoneNumber, name string) (*ChatwootContact, error)

	CreateConversation(contactID string, sessionID string) (*ChatwootConversation, error)

	SendMessage(conversationID, content, messageType string) error

	GetContact(phoneNumber string) (*ChatwootContact, error)

	GetConversation(conversationID string) (*ChatwootConversation, error)

	UpdateContactAttributes(contactID string, attributes map[string]interface{}) error
}

type ChatwootContact struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phoneNumber"`
	Email       string `json:"email"`
}

type ChatwootConversation struct {
	ID        int    `json:"id"`
	ContactID int    `json:"contact_id"`
	InboxID   int    `json:"inbox_id"`
	Status    string `json:"status"`
}

type WebhookService interface {
	SendWebhook(url string, payload interface{}) error

	RegisterWebhook(sessionID, url, secret string, events []string) error

	UnregisterWebhook(sessionID, url string) error

	GetWebhooks(sessionID string) ([]*WebhookRegistration, error)
}

type WebhookRegistration struct {
	ID        string   `json:"id"`
	SessionID string   `json:"session_id"`
	URL       string   `json:"url"`
	Secret    string   `json:"secret"`
	Events    []string `json:"events"`
	Active    bool     `json:"active"`
}
