package wameow

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	appMessage "zpwoot/internal/app/message"
	"zpwoot/internal/domain/message"
	"zpwoot/internal/domain/session"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

// SessionStats tracks statistics for a session
type SessionStats struct {
	MessagesSent     int64
	MessagesReceived int64
	LastActivity     int64
	StartTime        int64
}

// EventHandlerInfo stores information about registered event handlers
type EventHandlerInfo struct {
	ID      string
	Handler ports.EventHandler
}

// Manager implements the WameowManager interface
type Manager struct {
	clients       map[string]*WameowClient
	clientsMutex  sync.RWMutex
	container     *sqlstore.Container
	connectionMgr *ConnectionManager
	qrGenerator   *QRCodeGenerator
	sessionMgr    *SessionManager
	logger        *logger.Logger

	// Statistics tracking
	sessionStats map[string]*SessionStats
	statsMutex   sync.RWMutex

	// Event handlers
	eventHandlers map[string]map[string]*EventHandlerInfo // sessionID -> handlerID -> handler
	handlersMutex sync.RWMutex
}

// NewManager creates a new Wameow manager
func NewManager(
	container *sqlstore.Container,
	sessionRepo ports.SessionRepository,
	logger *logger.Logger,
) *Manager {
	return &Manager{
		clients:       make(map[string]*WameowClient),
		container:     container,
		connectionMgr: NewConnectionManager(logger),
		qrGenerator:   NewQRCodeGenerator(logger),
		sessionMgr:    NewSessionManager(sessionRepo, logger),
		logger:        logger,
		sessionStats:  make(map[string]*SessionStats),
		eventHandlers: make(map[string]map[string]*EventHandlerInfo),
	}
}

// CreateSession creates a new Wameow session
func (m *Manager) CreateSession(sessionID string, config *session.ProxyConfig) error {
	m.logger.InfoWithFields("Creating Wameow session", map[string]interface{}{
		"session_id": sessionID,
	})

	m.clientsMutex.Lock()
	defer m.clientsMutex.Unlock()

	// Check if session already exists
	if _, exists := m.clients[sessionID]; exists {
		return fmt.Errorf("session %s already exists", sessionID)
	}

	// Create WameowClient
	client, err := NewWameowClient(sessionID, m.container, m.sessionMgr.sessionRepo, m.logger)
	if err != nil {
		return fmt.Errorf("failed to create WameowClient for session %s: %w", sessionID, err)
	}

	// Set up event handlers
	m.setupEventHandlers(client.GetClient(), sessionID)

	// Apply proxy configuration if provided
	if config != nil {
		if err := m.applyProxyConfig(client.GetClient(), config); err != nil {
			m.logger.WarnWithFields("Failed to apply proxy config", map[string]interface{}{
				"session_id": sessionID,
				"error":      err.Error(),
			})
		}
	}

	// Store client
	m.clients[sessionID] = client

	// Initialize session statistics
	m.initSessionStats(sessionID)

	m.logger.InfoWithFields("Wameow session created successfully", map[string]interface{}{
		"session_id": sessionID,
	})

	return nil
}

// ConnectSession connects a Wameow session
func (m *Manager) ConnectSession(sessionID string) error {
	m.logger.InfoWithFields("Connecting Wameow session", map[string]interface{}{
		"session_id": sessionID,
	})

	client := m.getClient(sessionID)
	if client == nil {
		// Session not found in memory, try to load from database and create client
		m.logger.InfoWithFields("Session not found in memory, attempting to load from database", map[string]interface{}{
			"session_id": sessionID,
		})

		// Get session from database
		sess, err := m.sessionMgr.GetSession(sessionID)
		if err != nil {
			m.logger.ErrorWithFields("Failed to get session from database", map[string]interface{}{
				"session_id": sessionID,
				"error":      err.Error(),
			})
			return fmt.Errorf("session %s not found", sessionID)
		}

		// Create Wameow client for the existing session
		// NewWameowClient will automatically detect existing deviceJid
		if err := m.CreateSession(sessionID, sess.ProxyConfig); err != nil {
			m.logger.ErrorWithFields("Failed to create Wameow client for session", map[string]interface{}{
				"session_id": sessionID,
				"device_jid": sess.DeviceJid,
				"error":      err.Error(),
			})
			return fmt.Errorf("failed to initialize Wameow client for session %s: %w", sessionID, err)
		}

		// Get the newly created client
		client = m.getClient(sessionID)
		if client == nil {
			return fmt.Errorf("failed to create Wameow client for session %s", sessionID)
		}

		m.logger.InfoWithFields("Successfully loaded session from database and created Wameow client", map[string]interface{}{
			"session_id": sessionID,
		})
	}

	// Update session status to connecting
	// Connection status will be updated by event handlers

	// Start connection process (will handle QR code if needed)
	err := client.Connect()
	if err != nil {
		m.sessionMgr.UpdateConnectionStatus(sessionID, false)
		return fmt.Errorf("failed to connect session %s: %w", sessionID, err)
	}

	return nil
}

// DisconnectSession disconnects a Wameow session
func (m *Manager) DisconnectSession(sessionID string) error {
	m.logger.InfoWithFields("Disconnecting Wameow session", map[string]interface{}{
		"session_id": sessionID,
	})

	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	err := client.Disconnect()
	if err != nil {
		return fmt.Errorf("failed to disconnect session %s: %w", sessionID, err)
	}

	m.sessionMgr.UpdateConnectionStatus(sessionID, false)

	return nil
}

// LogoutSession logs out a Wameow session
func (m *Manager) LogoutSession(sessionID string) error {
	m.logger.InfoWithFields("Logging out Wameow session", map[string]interface{}{
		"session_id": sessionID,
	})

	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	// Logout from Wameow
	err := client.Logout()
	if err != nil {
		m.logger.WarnWithFields("Error during logout", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}

	// Update session status
	m.sessionMgr.UpdateConnectionStatus(sessionID, false)

	// Remove client from memory
	m.clientsMutex.Lock()
	delete(m.clients, sessionID)
	m.clientsMutex.Unlock()

	return nil
}

// GetQRCode gets QR code for session pairing
func (m *Manager) GetQRCode(sessionID string) (*session.QRCodeResponse, error) {
	m.logger.InfoWithFields("Getting QR code for session", map[string]interface{}{
		"session_id": sessionID,
	})

	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	if client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is already logged in", sessionID)
	}

	qrCode, err := client.GetQRCode()
	if err != nil {
		return nil, fmt.Errorf("failed to get QR code for session %s: %w", sessionID, err)
	}

	return &session.QRCodeResponse{
		QRCode:    qrCode,
		ExpiresAt: time.Now().Add(2 * time.Minute),
		Timeout:   120,
	}, nil
}

// PairPhone pairs a phone number with the session
func (m *Manager) PairPhone(sessionID, phoneNumber string) error {
	m.logger.InfoWithFields("Pairing phone number", map[string]interface{}{
		"session_id":   sessionID,
		"phone_number": phoneNumber,
	})

	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	// This would implement phone pairing logic
	// For now, return not implemented
	return fmt.Errorf("phone pairing not implemented yet")
}

// IsConnected checks if a session is connected
func (m *Manager) IsConnected(sessionID string) bool {
	client := m.getClient(sessionID)
	if client == nil {
		return false
	}
	return client.IsConnected()
}

// GetDeviceInfo gets device information for a session
func (m *Manager) GetDeviceInfo(sessionID string) (*session.DeviceInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	// This would get actual device info from Wameow
	// For now, return placeholder data
	return &session.DeviceInfo{
		Platform:    "web",
		DeviceModel: "Chrome",
		OSVersion:   "Unknown",
		AppVersion:  "2.2412.54",
	}, nil
}

// SetProxy sets proxy configuration for a session
func (m *Manager) SetProxy(sessionID string, config *session.ProxyConfig) error {
	m.logger.InfoWithFields("Setting proxy for session", map[string]interface{}{
		"session_id": sessionID,
		"proxy_type": config.Type,
		"proxy_host": config.Host,
	})

	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	return m.applyProxyConfig(client.GetClient(), config)
}

// initSessionStats initializes statistics for a session
func (m *Manager) initSessionStats(sessionID string) {
	m.statsMutex.Lock()
	defer m.statsMutex.Unlock()

	if _, exists := m.sessionStats[sessionID]; !exists {
		m.sessionStats[sessionID] = &SessionStats{
			StartTime: time.Now().Unix(),
		}
	}
}

// incrementMessagesSent increments the sent messages counter
func (m *Manager) incrementMessagesSent(sessionID string) {
	m.statsMutex.RLock()
	stats, exists := m.sessionStats[sessionID]
	m.statsMutex.RUnlock()

	if exists {
		atomic.AddInt64(&stats.MessagesSent, 1)
		atomic.StoreInt64(&stats.LastActivity, time.Now().Unix())
	}
}

// incrementMessagesReceived increments the received messages counter
func (m *Manager) incrementMessagesReceived(sessionID string) {
	m.statsMutex.RLock()
	stats, exists := m.sessionStats[sessionID]
	m.statsMutex.RUnlock()

	if exists {
		atomic.AddInt64(&stats.MessagesReceived, 1)
		atomic.StoreInt64(&stats.LastActivity, time.Now().Unix())
	}
}

// getSessionStats safely gets session statistics
func (m *Manager) getSessionStats(sessionID string) *SessionStats {
	m.statsMutex.RLock()
	defer m.statsMutex.RUnlock()

	stats, exists := m.sessionStats[sessionID]
	if !exists {
		return &SessionStats{
			StartTime: time.Now().Unix(),
		}
	}

	return stats
}

// GetProxy gets proxy configuration for a session
func (m *Manager) GetProxy(sessionID string) (*session.ProxyConfig, error) {
	// This would get the current proxy configuration
	// For now, return nil (no proxy)
	return nil, nil
}

// GetSessionStats retrieves session statistics
func (m *Manager) GetSessionStats(sessionID string) (*ports.SessionStats, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	// Get session statistics
	stats := m.getSessionStats(sessionID)

	// Calculate uptime
	uptime := int64(0)
	if stats.StartTime > 0 {
		uptime = time.Now().Unix() - stats.StartTime
	}

	return &ports.SessionStats{
		MessagesSent:     atomic.LoadInt64(&stats.MessagesSent),
		MessagesReceived: atomic.LoadInt64(&stats.MessagesReceived),
		LastActivity:     atomic.LoadInt64(&stats.LastActivity),
		Uptime:           uptime,
	}, nil
}

// GetSession retrieves a session by ID
func (m *Manager) GetSession(sessionID string) (*session.Session, error) {
	return m.sessionMgr.GetSession(sessionID)
}

// SendMediaMessage sends a media message
func (m *Manager) SendMediaMessage(sessionID, to string, media []byte, mediaType, caption string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	if !client.IsLoggedIn() {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	// Parse the recipient JID
	recipientJID, err := types.ParseJID(to)
	if err != nil {
		return fmt.Errorf("invalid recipient JID %s: %w", to, err)
	}

	// Upload the media
	uploaded, err := client.GetClient().Upload(context.Background(), media, whatsmeow.MediaType(mediaType))
	if err != nil {
		m.logger.ErrorWithFields("Failed to upload media", map[string]interface{}{
			"session_id": sessionID,
			"to":         to,
			"media_type": mediaType,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to upload media: %w", err)
	}

	// Create the appropriate message based on media type
	var msg *waE2E.Message
	switch mediaType {
	case "image":
		msg = &waE2E.Message{
			ImageMessage: &waE2E.ImageMessage{
				URL:           &uploaded.URL,
				DirectPath:    &uploaded.DirectPath,
				MediaKey:      uploaded.MediaKey,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    &uploaded.FileLength,
				Caption:       &caption,
			},
		}
	case "video":
		msg = &waE2E.Message{
			VideoMessage: &waE2E.VideoMessage{
				URL:           &uploaded.URL,
				DirectPath:    &uploaded.DirectPath,
				MediaKey:      uploaded.MediaKey,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    &uploaded.FileLength,
				Caption:       &caption,
			},
		}
	case "audio":
		msg = &waE2E.Message{
			AudioMessage: &waE2E.AudioMessage{
				URL:           &uploaded.URL,
				DirectPath:    &uploaded.DirectPath,
				MediaKey:      uploaded.MediaKey,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    &uploaded.FileLength,
			},
		}
	case "document":
		msg = &waE2E.Message{
			DocumentMessage: &waE2E.DocumentMessage{
				URL:           &uploaded.URL,
				DirectPath:    &uploaded.DirectPath,
				MediaKey:      uploaded.MediaKey,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    &uploaded.FileLength,
				Caption:       &caption,
			},
		}
	default:
		return fmt.Errorf("unsupported media type: %s", mediaType)
	}

	// Send the message
	_, err = client.GetClient().SendMessage(context.Background(), recipientJID, msg)
	if err != nil {
		m.logger.ErrorWithFields("Failed to send media message", map[string]interface{}{
			"session_id": sessionID,
			"to":         to,
			"media_type": mediaType,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to send media message: %w", err)
	}

	// Increment sent messages counter
	m.incrementMessagesSent(sessionID)

	m.logger.InfoWithFields("Media message sent successfully", map[string]interface{}{
		"session_id": sessionID,
		"to":         to,
		"media_type": mediaType,
	})

	return nil
}

// RegisterEventHandler registers an event handler for Wameow events
func (m *Manager) RegisterEventHandler(sessionID string, handler ports.EventHandler) error {
	m.handlersMutex.Lock()
	defer m.handlersMutex.Unlock()

	// Initialize session handlers map if it doesn't exist
	if m.eventHandlers[sessionID] == nil {
		m.eventHandlers[sessionID] = make(map[string]*EventHandlerInfo)
	}

	// Generate a unique handler ID
	handlerID := fmt.Sprintf("handler_%d", time.Now().UnixNano())

	// Store the handler
	m.eventHandlers[sessionID][handlerID] = &EventHandlerInfo{
		ID:      handlerID,
		Handler: handler,
	}

	// Get the client and register the actual event handler
	client := m.getClient(sessionID)
	if client != nil {
		client.GetClient().AddEventHandler(func(evt interface{}) {
			// Handle different event types and call appropriate handler methods
			switch e := evt.(type) {
			case *events.Message:
				m.incrementMessagesReceived(sessionID)
				// Convert to WameowMessage and call handler
				msg := &ports.WameowMessage{
					ID:   e.Info.ID,
					From: e.Info.Sender.String(),
					To:   e.Info.Chat.String(),
					Body: e.Message.GetConversation(),
				}
				handler.HandleMessage(sessionID, msg)
			case *events.Connected:
				handler.HandleConnection(sessionID, true)
			case *events.Disconnected:
				handler.HandleConnection(sessionID, false)
			case *events.QR:
				handler.HandleQRCode(sessionID, e.Codes[0])
			case *events.PairSuccess:
				handler.HandlePairSuccess(sessionID)
			}
		})
	}

	m.logger.InfoWithFields("Event handler registered", map[string]interface{}{
		"session_id": sessionID,
		"handler_id": handlerID,
	})

	return nil
}

// UnregisterEventHandler removes an event handler
func (m *Manager) UnregisterEventHandler(sessionID string, handlerID string) error {
	m.handlersMutex.Lock()
	defer m.handlersMutex.Unlock()

	// Check if session has handlers
	sessionHandlers, exists := m.eventHandlers[sessionID]
	if !exists {
		return fmt.Errorf("no event handlers found for session %s", sessionID)
	}

	// Check if handler exists
	_, exists = sessionHandlers[handlerID]
	if !exists {
		return fmt.Errorf("event handler %s not found for session %s", handlerID, sessionID)
	}

	// Remove the handler
	delete(sessionHandlers, handlerID)

	// Clean up empty session map
	if len(sessionHandlers) == 0 {
		delete(m.eventHandlers, sessionID)
	}

	m.logger.InfoWithFields("Event handler unregistered", map[string]interface{}{
		"session_id": sessionID,
		"handler_id": handlerID,
	})

	return nil
}

// getClient safely gets a client by session ID
func (m *Manager) getClient(sessionID string) *WameowClient {
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()
	return m.clients[sessionID]
}

// applyProxyConfig applies proxy configuration to a client
func (m *Manager) applyProxyConfig(client *whatsmeow.Client, config *session.ProxyConfig) error {
	// This would implement proxy configuration
	// For now, just log the configuration and validate client
	m.logger.InfoWithFields("Proxy configuration", map[string]interface{}{
		"type":       config.Type,
		"host":       config.Host,
		"port":       config.Port,
		"client_nil": client == nil,
	})

	if client == nil {
		return fmt.Errorf("cannot apply proxy config to nil client")
	}

	// Validate proxy configuration
	if config == nil {
		return fmt.Errorf("proxy configuration is nil")
	}

	// Validate proxy configuration format
	var proxyURL *url.URL
	var err error

	switch config.Type {
	case "http", "https":
		if config.Username != "" && config.Password != "" {
			proxyURL, err = url.Parse(fmt.Sprintf("http://%s:%s@%s:%d",
				config.Username, config.Password, config.Host, config.Port))
		} else {
			proxyURL, err = url.Parse(fmt.Sprintf("http://%s:%d", config.Host, config.Port))
		}
	case "socks5":
		if config.Username != "" && config.Password != "" {
			proxyURL, err = url.Parse(fmt.Sprintf("socks5://%s:%s@%s:%d",
				config.Username, config.Password, config.Host, config.Port))
		} else {
			proxyURL, err = url.Parse(fmt.Sprintf("socks5://%s:%d", config.Host, config.Port))
		}
	default:
		return fmt.Errorf("unsupported proxy type: %s", config.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to parse proxy URL: %w", err)
	}

	// Note: The whatsmeow library doesn't support changing HTTP client after creation.
	// Proxy configuration should be done during client creation in the NewWameowClient function
	// or using environment variables. For now, we validate the proxy URL format.

	m.logger.InfoWithFields("Proxy configuration validated (not yet applied)", map[string]interface{}{
		"type":      config.Type,
		"host":      config.Host,
		"port":      config.Port,
		"proxy_url": proxyURL.String(),
		"note":      "Proxy configuration validation successful, but actual proxy application requires client recreation",
	})

	return nil
}

// SendMessage sends a message through a session
func (m *Manager) SendMessage(sessionID, to, messageType, body, caption, file, filename string, latitude, longitude float64, contactName, contactPhone string) (*message.SendResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	var resp *whatsmeow.SendResponse
	var err error

	switch messageType {
	case "text":
		resp, err = client.SendTextMessage(ctx, to, body)
	case "image":
		resp, err = client.SendImageMessage(ctx, to, file, caption)
	case "audio":
		resp, err = client.SendAudioMessage(ctx, to, file)
	case "video":
		resp, err = client.SendVideoMessage(ctx, to, file, caption)
	case "document":
		resp, err = client.SendDocumentMessage(ctx, to, file, filename)
	case "location":
		resp, err = client.SendLocationMessage(ctx, to, latitude, longitude, body)
	case "contact":
		resp, err = client.SendContactMessage(ctx, to, contactName, contactPhone)
	case "sticker":
		resp, err = client.SendStickerMessage(ctx, to, file)
	default:
		return nil, fmt.Errorf("unsupported message type: %s", messageType)
	}

	if err != nil {
		return &message.SendResult{
			Status:    "failed",
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, err
	}

	return &message.SendResult{
		MessageID: resp.ID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
	}, nil
}

// SendButtonMessage sends a button message through a session
func (m *Manager) SendButtonMessage(sessionID, to, body string, buttons []map[string]string) (*message.SendResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	resp, err := client.SendButtonMessage(ctx, to, body, buttons)
	if err != nil {
		return &message.SendResult{
			Status:    "failed",
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, err
	}

	return &message.SendResult{
		MessageID: resp.ID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
	}, nil
}

// SendListMessage sends a list message through a session
func (m *Manager) SendListMessage(sessionID, to, body, buttonText string, sections []map[string]interface{}) (*message.SendResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	resp, err := client.SendListMessage(ctx, to, body, buttonText, sections)
	if err != nil {
		return &message.SendResult{
			Status:    "failed",
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, err
	}

	return &message.SendResult{
		MessageID: resp.ID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
	}, nil
}

// SendReaction sends a reaction through a session
func (m *Manager) SendReaction(sessionID, to, messageID, reaction string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	return client.SendReaction(ctx, to, messageID, reaction)
}

// SendPresence sends presence information through a session
func (m *Manager) SendPresence(sessionID, to, presence string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	return client.SendPresence(ctx, to, presence)
}

// EditMessage edits a message through a session
func (m *Manager) EditMessage(sessionID, to, messageID, newText string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	return client.EditMessage(ctx, to, messageID, newText)
}

// DeleteMessage deletes a message through a session
func (m *Manager) DeleteMessage(sessionID, to, messageID string, forAll bool) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	return client.DeleteMessage(ctx, to, messageID, forAll)
}

// setupEventHandlers sets up event handlers for a Wameow client
func (m *Manager) setupEventHandlers(client *whatsmeow.Client, sessionID string) {
	m.logger.InfoWithFields("Setting up event handlers", map[string]interface{}{
		"session_id": sessionID,
	})

	// Set up the actual event handlers
	m.SetupEventHandlers(client, sessionID)
}



// ContactListResult represents the result of sending multiple contacts
type ContactListResult struct {
	TotalContacts int
	SuccessCount  int
	FailureCount  int
	Results       []ContactResult
	Timestamp     time.Time
}

// ContactResult represents the result of sending a single contact
type ContactResult struct {
	ContactName string
	MessageID   string
	Status      string
	Error       string
}

// TextMessageResult represents the result of sending a text message
type TextMessageResult struct {
	MessageID string
	Status    string
	Timestamp time.Time
}

// SendTextMessage sends a text message with optional reply/quote
func (m *Manager) SendTextMessage(sessionID, to, text string, contextInfo *appMessage.ContextInfo) (*TextMessageResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("session %s is not connected", sessionID)
	}

	// Parse recipient JID
	recipientJID, err := types.ParseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient JID: %w", err)
	}

	// Generate message ID
	messageID := client.GetClient().GenerateMessageID()

	// Create base message
	msg := &waE2E.Message{
		Conversation: proto.String(text),
	}

	// Add reply context if provided
	if contextInfo != nil {
		// Create context info for reply
		waContextInfo := &waE2E.ContextInfo{
			StanzaID:      proto.String(contextInfo.StanzaID),
			QuotedMessage: &waE2E.Message{Conversation: proto.String("")},
		}

		// Set participant for group messages
		if contextInfo.Participant != "" {
			waContextInfo.Participant = proto.String(contextInfo.Participant)
		}

		// Convert to ExtendedTextMessage to include context
		msg = &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text:        proto.String(text),
				ContextInfo: waContextInfo,
			},
		}
	}

	// Send message
	resp, err := client.GetClient().SendMessage(context.Background(), recipientJID, msg, whatsmeow.SendRequestExtra{ID: messageID})
	if err != nil {
		return nil, fmt.Errorf("failed to send text message: %w", err)
	}

	m.logger.InfoWithFields("Text message sent", map[string]interface{}{
		"session_id": sessionID,
		"to":         to,
		"message_id": messageID,
		"has_reply":  contextInfo != nil,
		"timestamp":  resp.Timestamp,
	})

	return &TextMessageResult{
		MessageID: messageID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
	}, nil
}

// SendContactList sends multiple contacts in a single message
func (m *Manager) SendContactList(sessionID, to string, contacts []ContactInfo) (*ContactListResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	result := &ContactListResult{
		TotalContacts: len(contacts),
		Results:       make([]ContactResult, 0, len(contacts)),
		Timestamp:     time.Now(),
	}

	// Convert to wameow ContactInfo slice
	var wameowContacts []ContactInfo
	for _, contact := range contacts {
		wameowContacts = append(wameowContacts, ContactInfo{
			Name:         contact.Name,
			Phone:        contact.Phone,
			Email:        contact.Email,
			Organization: contact.Organization,
			Title:        contact.Title,
			Website:      contact.Website,
			Address:      contact.Address,
		})
	}

	// Send all contacts in a single message
	resp, err := client.SendContactListMessage(ctx, to, wameowContacts)
	if err != nil {
		// If sending as a list fails, mark all as failed
		for _, contact := range contacts {
			result.Results = append(result.Results, ContactResult{
				ContactName: contact.Name,
				Status:      "failed",
				Error:       err.Error(),
			})
		}
		result.FailureCount = len(contacts)
		return result, err
	}

	// If successful, mark all contacts as sent with the same message ID
	for _, contact := range contacts {
		result.Results = append(result.Results, ContactResult{
			ContactName: contact.Name,
			MessageID:   resp.ID,
			Status:      "sent",
		})
	}
	result.SuccessCount = len(contacts)

	return result, nil
}

// SendContactListBusiness sends multiple contacts using Business format
func (m *Manager) SendContactListBusiness(sessionID, to string, contacts []ContactInfo) (*ContactListResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("session %s is not connected", sessionID)
	}

	// Convert to wameow ContactInfo slice
	var wameowContacts []ContactInfo
	for _, contact := range contacts {
		wameowContacts = append(wameowContacts, ContactInfo{
			Name:         contact.Name,
			Phone:        contact.Phone,
			Email:        contact.Email,
			Organization: contact.Organization,
			Title:        contact.Title,
			Website:      contact.Website,
			Address:      contact.Address,
		})
	}

	// Send using Business format
	resp, err := client.SendContactListMessageBusiness(context.Background(), to, wameowContacts)
	if err != nil {
		return nil, fmt.Errorf("failed to send WhatsApp Business contact list: %w", err)
	}

	// Create result
	result := &ContactListResult{
		TotalContacts: len(contacts),
		SuccessCount:  len(contacts),
		FailureCount:  0,
		Results:       make([]ContactResult, len(contacts)),
		Timestamp:     time.Now(),
	}

	// All contacts share the same message ID in a contact list
	for i, contact := range contacts {
		result.Results[i] = ContactResult{
			ContactName: contact.Name,
			MessageID:   resp.ID,
			Status:      "sent",
		}
	}

	return result, nil
}

// SendSingleContact sends a single contact using ContactMessage (not ContactsArrayMessage)
func (m *Manager) SendSingleContact(sessionID, to string, contact ContactInfo) (*ContactListResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("session %s is not connected", sessionID)
	}

	// Send using single contact format (ContactMessage)
	resp, err := client.SendSingleContactMessage(context.Background(), to, ContactInfo{
		Name:         contact.Name,
		Phone:        contact.Phone,
		Email:        contact.Email,
		Organization: contact.Organization,
		Title:        contact.Title,
		Website:      contact.Website,
		Address:      contact.Address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send single contact: %w", err)
	}

	// Create result
	result := &ContactListResult{
		TotalContacts: 1,
		SuccessCount:  1,
		FailureCount:  0,
		Results:       make([]ContactResult, 1),
		Timestamp:     time.Now(),
	}

	result.Results[0] = ContactResult{
		ContactName: contact.Name,
		MessageID:   resp.ID,
		Status:      "sent",
	}

	return result, nil
}

// SendSingleContactBusiness sends a single contact using Business format
func (m *Manager) SendSingleContactBusiness(sessionID, to string, contact ContactInfo) (*ContactListResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("session %s is not connected", sessionID)
	}

	// Send using single contact Business format
	resp, err := client.SendSingleContactMessageBusiness(context.Background(), to, ContactInfo{
		Name:         contact.Name,
		Phone:        contact.Phone,
		Email:        contact.Email,
		Organization: contact.Organization,
		Title:        contact.Title,
		Website:      contact.Website,
		Address:      contact.Address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send Business single contact: %w", err)
	}

	// Create result
	result := &ContactListResult{
		TotalContacts: 1,
		SuccessCount:  1,
		FailureCount:  0,
		Results:       make([]ContactResult, 1),
		Timestamp:     time.Now(),
	}

	result.Results[0] = ContactResult{
		ContactName: contact.Name,
		MessageID:   resp.ID,
		Status:      "sent",
	}

	return result, nil
}

// SendSingleContactBusinessFormat sends a single contact using Business format
func (m *Manager) SendSingleContactBusinessFormat(sessionID, to string, contact ContactInfo) (*ContactListResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("session %s is not connected", sessionID)
	}

	// Send using single contact Business format
	resp, err := client.SendSingleContactMessageBusiness(context.Background(), to, ContactInfo{
		Name:         contact.Name,
		Phone:        contact.Phone,
		Email:        contact.Email,
		Organization: contact.Organization,
		Title:        contact.Title,
		Website:      contact.Website,
		Address:      contact.Address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send WhatsApp Business single contact: %w", err)
	}

	// Create result
	result := &ContactListResult{
		TotalContacts: 1,
		SuccessCount:  1,
		FailureCount:  0,
		Results:       make([]ContactResult, 1),
		Timestamp:     time.Now(),
	}

	result.Results[0] = ContactResult{
		ContactName: contact.Name,
		MessageID:   resp.ID,
		Status:      "sent",
	}

	return result, nil
}

// SetupEventHandlers sets up all event handlers for a Wameow client
func (m *Manager) SetupEventHandlers(client *whatsmeow.Client, sessionID string) {
	eventHandler := NewEventHandler(m, m.sessionMgr, m.qrGenerator, m.logger)

	client.AddEventHandler(func(evt interface{}) {
		eventHandler.HandleEvent(evt, sessionID)
	})
}
