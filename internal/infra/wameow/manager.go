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

type SessionStats struct {
	MessagesSent     int64
	MessagesReceived int64
	LastActivity     int64
	StartTime        int64
}

type EventHandlerInfo struct {
	ID      string
	Handler ports.EventHandler
}

type Manager struct {
	clients       map[string]*WameowClient
	clientsMutex  sync.RWMutex
	container     *sqlstore.Container
	connectionMgr *ConnectionManager
	qrGenerator   *QRCodeGenerator
	sessionMgr    *SessionManager
	logger        *logger.Logger

	sessionStats map[string]*SessionStats
	statsMutex   sync.RWMutex

	eventHandlers map[string]map[string]*EventHandlerInfo // sessionID -> handlerID -> handler
	handlersMutex sync.RWMutex
}

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

func (m *Manager) CreateSession(sessionID string, config *session.ProxyConfig) error {
	m.logger.InfoWithFields("Creating Wameow session", map[string]interface{}{
		"session_id": sessionID,
	})

	m.clientsMutex.Lock()
	defer m.clientsMutex.Unlock()

	if _, exists := m.clients[sessionID]; exists {
		return fmt.Errorf("session %s already exists", sessionID)
	}

	client, err := NewWameowClient(sessionID, m.container, m.sessionMgr.sessionRepo, m.logger)
	if err != nil {
		return fmt.Errorf("failed to create WameowClient for session %s: %w", sessionID, err)
	}

	m.setupEventHandlers(client.GetClient(), sessionID)

	if config != nil {
		if err := m.applyProxyConfig(client.GetClient(), config); err != nil {
			m.logger.WarnWithFields("Failed to apply proxy config", map[string]interface{}{
				"session_id": sessionID,
				"error":      err.Error(),
			})
		}
	}

	m.clients[sessionID] = client

	m.initSessionStats(sessionID)

	m.logger.InfoWithFields("Wameow session created successfully", map[string]interface{}{
		"session_id": sessionID,
	})

	return nil
}

func (m *Manager) ConnectSession(sessionID string) error {
	m.logger.InfoWithFields("Connecting Wameow session", map[string]interface{}{
		"session_id": sessionID,
	})

	client := m.getClient(sessionID)
	if client == nil {
		m.logger.InfoWithFields("Session not found in memory, attempting to load from database", map[string]interface{}{
			"session_id": sessionID,
		})

		sess, err := m.sessionMgr.GetSession(sessionID)
		if err != nil {
			m.logger.ErrorWithFields("Failed to get session from database", map[string]interface{}{
				"session_id": sessionID,
				"error":      err.Error(),
			})
			return fmt.Errorf("session %s not found", sessionID)
		}

		if err := m.CreateSession(sessionID, sess.ProxyConfig); err != nil {
			m.logger.ErrorWithFields("Failed to create Wameow client for session", map[string]interface{}{
				"session_id": sessionID,
				"device_jid": sess.DeviceJid,
				"error":      err.Error(),
			})
			return fmt.Errorf("failed to initialize Wameow client for session %s: %w", sessionID, err)
		}

		client = m.getClient(sessionID)
		if client == nil {
			return fmt.Errorf("failed to create Wameow client for session %s", sessionID)
		}

		m.logger.InfoWithFields("Successfully loaded session from database and created Wameow client", map[string]interface{}{
			"session_id": sessionID,
		})
	}

	err := client.Connect()
	if err != nil {
		m.sessionMgr.UpdateConnectionStatus(sessionID, false)
		return fmt.Errorf("failed to connect session %s: %w", sessionID, err)
	}

	return nil
}

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

func (m *Manager) LogoutSession(sessionID string) error {
	m.logger.InfoWithFields("Logging out Wameow session", map[string]interface{}{
		"session_id": sessionID,
	})

	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	err := client.Logout()
	if err != nil {
		m.logger.WarnWithFields("Error during logout", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}

	m.sessionMgr.UpdateConnectionStatus(sessionID, false)

	m.clientsMutex.Lock()
	delete(m.clients, sessionID)
	m.clientsMutex.Unlock()

	return nil
}

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

func (m *Manager) PairPhone(sessionID, phoneNumber string) error {
	m.logger.InfoWithFields("Pairing phone number", map[string]interface{}{
		"session_id":   sessionID,
		"phone_number": phoneNumber,
	})

	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	return fmt.Errorf("phone pairing not implemented yet")
}

func (m *Manager) IsConnected(sessionID string) bool {
	client := m.getClient(sessionID)
	if client == nil {
		return false
	}
	return client.IsConnected()
}

func (m *Manager) GetDeviceInfo(sessionID string) (*session.DeviceInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	return &session.DeviceInfo{
		Platform:    "web",
		DeviceModel: "Chrome",
		OSVersion:   "Unknown",
		AppVersion:  "2.2412.54",
	}, nil
}

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

func (m *Manager) initSessionStats(sessionID string) {
	m.statsMutex.Lock()
	defer m.statsMutex.Unlock()

	if _, exists := m.sessionStats[sessionID]; !exists {
		m.sessionStats[sessionID] = &SessionStats{
			StartTime: time.Now().Unix(),
		}
	}
}

func (m *Manager) incrementMessagesSent(sessionID string) {
	m.statsMutex.RLock()
	stats, exists := m.sessionStats[sessionID]
	m.statsMutex.RUnlock()

	if exists {
		atomic.AddInt64(&stats.MessagesSent, 1)
		atomic.StoreInt64(&stats.LastActivity, time.Now().Unix())
	}
}

func (m *Manager) incrementMessagesReceived(sessionID string) {
	m.statsMutex.RLock()
	stats, exists := m.sessionStats[sessionID]
	m.statsMutex.RUnlock()

	if exists {
		atomic.AddInt64(&stats.MessagesReceived, 1)
		atomic.StoreInt64(&stats.LastActivity, time.Now().Unix())
	}
}

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

func (m *Manager) GetProxy(sessionID string) (*session.ProxyConfig, error) {
	return nil, nil
}

func (m *Manager) GetSessionStats(sessionID string) (*ports.SessionStats, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	stats := m.getSessionStats(sessionID)

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

func (m *Manager) GetSession(sessionID string) (*session.Session, error) {
	return m.sessionMgr.GetSession(sessionID)
}

func (m *Manager) SendMediaMessage(sessionID, to string, media []byte, mediaType, caption string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	if !client.IsLoggedIn() {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	recipientJID, err := types.ParseJID(to)
	if err != nil {
		return fmt.Errorf("invalid recipient JID %s: %w", to, err)
	}

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

	m.incrementMessagesSent(sessionID)

	m.logger.InfoWithFields("Media message sent successfully", map[string]interface{}{
		"session_id": sessionID,
		"to":         to,
		"media_type": mediaType,
	})

	return nil
}

func (m *Manager) RegisterEventHandler(sessionID string, handler ports.EventHandler) error {
	m.handlersMutex.Lock()
	defer m.handlersMutex.Unlock()

	if m.eventHandlers[sessionID] == nil {
		m.eventHandlers[sessionID] = make(map[string]*EventHandlerInfo)
	}

	handlerID := fmt.Sprintf("handler_%d", time.Now().UnixNano())

	m.eventHandlers[sessionID][handlerID] = &EventHandlerInfo{
		ID:      handlerID,
		Handler: handler,
	}

	client := m.getClient(sessionID)
	if client != nil {
		client.GetClient().AddEventHandler(func(evt interface{}) {
			switch e := evt.(type) {
			case *events.Message:
				m.incrementMessagesReceived(sessionID)
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

func (m *Manager) UnregisterEventHandler(sessionID string, handlerID string) error {
	m.handlersMutex.Lock()
	defer m.handlersMutex.Unlock()

	sessionHandlers, exists := m.eventHandlers[sessionID]
	if !exists {
		return fmt.Errorf("no event handlers found for session %s", sessionID)
	}

	_, exists = sessionHandlers[handlerID]
	if !exists {
		return fmt.Errorf("event handler %s not found for session %s", handlerID, sessionID)
	}

	delete(sessionHandlers, handlerID)

	if len(sessionHandlers) == 0 {
		delete(m.eventHandlers, sessionID)
	}

	m.logger.InfoWithFields("Event handler unregistered", map[string]interface{}{
		"session_id": sessionID,
		"handler_id": handlerID,
	})

	return nil
}

func (m *Manager) getClient(sessionID string) *WameowClient {
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()
	return m.clients[sessionID]
}

func (m *Manager) applyProxyConfig(client *whatsmeow.Client, config *session.ProxyConfig) error {
	m.logger.InfoWithFields("Proxy configuration", map[string]interface{}{
		"type":       config.Type,
		"host":       config.Host,
		"port":       config.Port,
		"client_nil": client == nil,
	})

	if client == nil {
		return fmt.Errorf("cannot apply proxy config to nil client")
	}

	if config == nil {
		return fmt.Errorf("proxy configuration is nil")
	}

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

	m.logger.InfoWithFields("Proxy configuration validated (not yet applied)", map[string]interface{}{
		"type":      config.Type,
		"host":      config.Host,
		"port":      config.Port,
		"proxy_url": proxyURL.String(),
		"note":      "Proxy configuration validation successful, but actual proxy application requires client recreation",
	})

	return nil
}

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

func (m *Manager) setupEventHandlers(client *whatsmeow.Client, sessionID string) {
	m.logger.InfoWithFields("Setting up event handlers", map[string]interface{}{
		"session_id": sessionID,
	})

	m.SetupEventHandlers(client, sessionID)
}

type ContactListResult struct {
	TotalContacts int
	SuccessCount  int
	FailureCount  int
	Results       []ContactResult
	Timestamp     time.Time
}

type ContactResult struct {
	ContactName string
	MessageID   string
	Status      string
	Error       string
}

type TextMessageResult struct {
	MessageID string
	Status    string
	Timestamp time.Time
}

func (m *Manager) SendTextMessage(sessionID, to, text string, contextInfo *appMessage.ContextInfo) (*TextMessageResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("session %s is not connected", sessionID)
	}

	recipientJID, err := types.ParseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient JID: %w", err)
	}

	messageID := client.GetClient().GenerateMessageID()

	msg := &waE2E.Message{
		Conversation: proto.String(text),
	}

	if contextInfo != nil {
		waContextInfo := &waE2E.ContextInfo{
			StanzaID:      proto.String(contextInfo.StanzaID),
			QuotedMessage: &waE2E.Message{Conversation: proto.String("")},
		}

		if contextInfo.Participant != "" {
			waContextInfo.Participant = proto.String(contextInfo.Participant)
		}

		msg = &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text:        proto.String(text),
				ContextInfo: waContextInfo,
			},
		}
	}

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

	resp, err := client.SendContactListMessage(ctx, to, wameowContacts)
	if err != nil {
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

func (m *Manager) SendContactListBusiness(sessionID, to string, contacts []ContactInfo) (*ContactListResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("session %s is not connected", sessionID)
	}

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

	resp, err := client.SendContactListMessageBusiness(context.Background(), to, wameowContacts)
	if err != nil {
		return nil, fmt.Errorf("failed to send WhatsApp Business contact list: %w", err)
	}

	result := &ContactListResult{
		TotalContacts: len(contacts),
		SuccessCount:  len(contacts),
		FailureCount:  0,
		Results:       make([]ContactResult, len(contacts)),
		Timestamp:     time.Now(),
	}

	for i, contact := range contacts {
		result.Results[i] = ContactResult{
			ContactName: contact.Name,
			MessageID:   resp.ID,
			Status:      "sent",
		}
	}

	return result, nil
}

func (m *Manager) SendSingleContact(sessionID, to string, contact ContactInfo) (*ContactListResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("session %s is not connected", sessionID)
	}

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

func (m *Manager) SendSingleContactBusiness(sessionID, to string, contact ContactInfo) (*ContactListResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("session %s is not connected", sessionID)
	}

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

func (m *Manager) SendSingleContactBusinessFormat(sessionID, to string, contact ContactInfo) (*ContactListResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("session %s is not connected", sessionID)
	}

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

func (m *Manager) SetupEventHandlers(client *whatsmeow.Client, sessionID string) {
	eventHandler := NewEventHandler(m, m.sessionMgr, m.qrGenerator, m.logger)

	client.AddEventHandler(func(evt interface{}) {
		eventHandler.HandleEvent(evt, sessionID)
	})
}
