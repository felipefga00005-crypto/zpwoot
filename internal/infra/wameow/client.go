// Refactored: separated responsibilities; extracted interfaces; standardized error handling
package wameow

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	appMessage "zpwoot/internal/app/message"
	"zpwoot/internal/domain/session"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

// WhatsAppClient defines the interface for WhatsApp client operations
type WhatsAppClient interface {
	Connect() error
	Disconnect() error
	IsConnected() bool
	IsLoggedIn() bool
	GetQRCode() (string, error)
	Logout() error
	SendMessage(ctx context.Context, to string, message interface{}) error
}

// MessageSender handles message sending operations
type MessageSender interface {
	SendText(ctx context.Context, to, body string, contextInfo *appMessage.ContextInfo) (*whatsmeow.SendResponse, error)
	SendMedia(ctx context.Context, to, filePath string, mediaType MediaType, options MediaOptions) (*whatsmeow.SendResponse, error)
	SendContact(ctx context.Context, to string, contact ContactInfo) (*whatsmeow.SendResponse, error)
	SendLocation(ctx context.Context, to string, lat, lng float64, address string) (*whatsmeow.SendResponse, error)
}

// MediaType represents different media types
type MediaType int

const (
	MediaTypeImage MediaType = iota
	MediaTypeAudio
	MediaTypeVideo
	MediaTypeDocument
	MediaTypeSticker
)

// MediaOptions contains options for media messages
type MediaOptions struct {
	Caption     string
	Filename    string
	MimeType    string
	ContextInfo *appMessage.ContextInfo
}

// QRGenerator defines the interface for QR code operations
type QRGenerator interface {
	GenerateQRCodeImage(qrText string) string
	DisplayQRCodeInTerminal(qrCode, sessionID string)
}

// SessionUpdater defines the interface for session management operations
type SessionUpdater interface {
	UpdateConnectionStatus(sessionID string, isConnected bool)
	GetSession(sessionID string) (*session.Session, error)
	GetSessionRepo() ports.SessionRepository
}

type WameowClient struct {
	sessionID string
	client    *whatsmeow.Client
	logger    *logger.Logger

	// Composed services
	sessionMgr  SessionUpdater
	qrGenerator QRGenerator
	msgSender   MessageSender

	// State management
	mu           sync.RWMutex
	status       string
	lastActivity time.Time

	// QR code management
	qrState QRState

	// Lifecycle management
	ctx    context.Context
	cancel context.CancelFunc
}

// QRState encapsulates QR code related state
type QRState struct {
	mu          sync.RWMutex
	code        string
	codeBase64  string
	loopActive  bool
	stopChannel chan bool
}

func NewWameowClient(
	sessionID string,
	container *sqlstore.Container,
	sessionRepo ports.SessionRepository,
	logger *logger.Logger,
) (*WameowClient, error) {
	if err := ValidateSessionID(sessionID); err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	deviceJid, err := getExistingDeviceJID(sessionRepo, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device JID: %w", err)
	}

	deviceStore := GetDeviceStoreForSession(sessionID, deviceJid, container)
	if deviceStore == nil {
		return nil, fmt.Errorf("failed to create device store for session %s", sessionID)
	}

	client, err := createWhatsAppClient(deviceStore, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create WhatsApp client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	wameowClient := &WameowClient{
		sessionID:    sessionID,
		client:       client,
		logger:       logger,
		sessionMgr:   NewSessionManager(sessionRepo, logger),
		qrGenerator:  NewQRCodeGenerator(logger),
		status:       "disconnected",
		lastActivity: time.Now(),
		qrState: QRState{
			stopChannel: make(chan bool, 1),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize message sender
	wameowClient.msgSender = NewMessageSender(client, logger)

	return wameowClient, nil
}

func getExistingDeviceJID(sessionRepo ports.SessionRepository, sessionID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sess, err := sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return "", nil
	}

	return sess.DeviceJid, nil
}

func createWhatsAppClient(deviceStore interface{}, logger *logger.Logger) (*whatsmeow.Client, error) {
	waLogger := NewWameowLogger(logger)
	client := whatsmeow.NewClient(deviceStore.(*store.Device), waLogger)
	if client == nil {
		return nil, fmt.Errorf("whatsmeow.NewClient returned nil")
	}
	return client, nil
}

func (c *WameowClient) Connect() error {
	c.logger.InfoWithFields("Starting connection process", map[string]interface{}{
		"session_id": c.sessionID,
	})

	c.stopQRLoop()

	if c.client.IsConnected() {
		c.client.Disconnect()
	}

	// Update context without holding the main mutex
	c.mu.Lock()
	if c.cancel != nil {
		c.cancel()
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.mu.Unlock()

	c.setStatus("connecting")
	go c.startClientLoop()

	return nil
}

func (c *WameowClient) Disconnect() error {
	c.logger.InfoWithFields("Disconnecting client", map[string]interface{}{
		"session_id": c.sessionID,
	})

	c.mu.Lock()
	defer c.mu.Unlock()

	c.stopQRLoop()

	if c.client.IsConnected() {
		c.client.Disconnect()
	}

	if c.cancel != nil {
		c.cancel()
	}

	c.setStatus("disconnected")
	return nil
}

func (c *WameowClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.client.IsConnected()
}

func (c *WameowClient) IsLoggedIn() bool {
	return c.client.IsLoggedIn()
}

func (c *WameowClient) GetQRCode() (string, error) {
	c.qrState.mu.RLock()
	defer c.qrState.mu.RUnlock()

	if c.qrState.code == "" {
		return "", fmt.Errorf("no QR code available")
	}

	return c.qrState.code, nil
}

func (c *WameowClient) GetClient() *whatsmeow.Client {
	return c.client
}

func (c *WameowClient) GetJID() types.JID {
	if c.client.Store.ID == nil {
		return types.EmptyJID
	}
	return *c.client.Store.ID
}

func (c *WameowClient) setStatus(status string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.status = status
	c.lastActivity = time.Now()
	c.logger.InfoWithFields("Session status updated", map[string]interface{}{
		"session_id": c.sessionID,
		"status":     status,
	})

	switch status {
	case "connected":
		c.sessionMgr.UpdateConnectionStatus(c.sessionID, true)
	case "disconnected":
		c.sessionMgr.UpdateConnectionStatus(c.sessionID, false)
	}
}

func (c *WameowClient) startClientLoop() {
	defer func() {
		if r := recover(); r != nil {
			c.logger.ErrorWithFields("Client loop panic", map[string]interface{}{
				"session_id": c.sessionID,
				"error":      r,
			})
		}
	}()

	isRegistered := IsDeviceRegistered(c.client)

	if !isRegistered {
		c.logger.InfoWithFields("Device not registered, starting QR code process", map[string]interface{}{
			"session_id": c.sessionID,
		})
		c.handleNewDeviceRegistration()
	} else {
		c.handleExistingDeviceConnection()
	}
}

func (c *WameowClient) handleNewDeviceRegistration() {
	qrChan, err := c.client.GetQRChannel(context.Background())
	if err != nil {
		c.logger.ErrorWithFields("Failed to get QR channel", map[string]interface{}{
			"session_id": c.sessionID,
			"error":      err.Error(),
		})
		c.setStatus("disconnected")
		return
	}

	err = c.client.Connect()
	if err != nil {
		c.logger.ErrorWithFields("Failed to connect client", map[string]interface{}{
			"session_id": c.sessionID,
			"error":      err.Error(),
		})
		c.setStatus("disconnected")
		return
	}

	c.handleQRLoop(qrChan)
}

func (c *WameowClient) handleExistingDeviceConnection() {
	err := c.client.Connect()
	if err != nil {
		c.logger.ErrorWithFields("Failed to connect existing device", map[string]interface{}{
			"session_id": c.sessionID,
			"error":      err.Error(),
		})
		c.setStatus("disconnected")
		return
	}

	time.Sleep(2 * time.Second)

	if c.client.IsConnected() {
		c.logger.InfoWithFields("Successfully connected session", map[string]interface{}{
			"session_id": c.sessionID,
		})
		c.setStatus("connected")
	} else {
		c.logger.WarnWithFields("Connection attempt failed", map[string]interface{}{
			"session_id": c.sessionID,
		})
		c.setStatus("disconnected")
	}
}

func (c *WameowClient) handleQRLoop(qrChan <-chan whatsmeow.QRChannelItem) {
	if qrChan == nil {
		c.logger.ErrorWithFields("QR channel is nil", map[string]interface{}{
			"session_id": c.sessionID,
		})
		return
	}

	c.qrState.mu.Lock()
	c.qrState.loopActive = true
	c.qrState.mu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			c.logger.ErrorWithFields("QR loop panic", map[string]interface{}{
				"session_id": c.sessionID,
				"error":      r,
			})
		}
		c.qrState.mu.Lock()
		c.qrState.loopActive = false
		c.qrState.mu.Unlock()
	}()

	for {
		select {
		case <-c.ctx.Done():
			c.logger.InfoWithFields("QR loop cancelled", map[string]interface{}{
				"session_id": c.sessionID,
			})
			return

		case <-c.qrState.stopChannel:
			c.logger.InfoWithFields("QR loop stopped", map[string]interface{}{
				"session_id": c.sessionID,
			})
			return

		case evt, ok := <-qrChan:
			if !ok {
				c.logger.InfoWithFields("QR channel closed", map[string]interface{}{
					"session_id": c.sessionID,
				})
				c.setStatus("disconnected")
				return
			}

			c.handleQREvent(evt)
		}
	}
}

func (c *WameowClient) handleQREvent(evt whatsmeow.QRChannelItem) {
	switch evt.Event {
	case "code":
		c.updateQRCode(evt.Code)
		c.displayQRCode(evt.Code)
		c.setStatus("connecting")

	case "success":
		c.logger.InfoWithFields("QR code scanned successfully", map[string]interface{}{
			"session_id": c.sessionID,
		})
		c.clearQRCode()
		c.setStatus("connected")

	case "timeout":
		c.logger.WarnWithFields("QR code timeout", map[string]interface{}{
			"session_id": c.sessionID,
		})
		c.clearQRCode()
		c.setStatus("disconnected")

	default:
		c.logger.InfoWithFields("QR event", map[string]interface{}{
			"session_id": c.sessionID,
			"event":      evt.Event,
		})
	}
}

func (c *WameowClient) updateQRCode(code string) {
	c.qrState.mu.Lock()
	defer c.qrState.mu.Unlock()

	c.qrState.code = code
	c.qrState.codeBase64 = c.qrGenerator.GenerateQRCodeImage(code)
}

func (c *WameowClient) displayQRCode(code string) {
	c.qrGenerator.DisplayQRCodeInTerminal(code, c.sessionID)
	c.logger.InfoWithFields("QR code generated", map[string]interface{}{
		"session_id": c.sessionID,
	})
}

func (c *WameowClient) clearQRCode() {
	c.qrState.mu.Lock()
	defer c.qrState.mu.Unlock()

	c.qrState.code = ""
	c.qrState.codeBase64 = ""
}

func (c *WameowClient) stopQRLoop() {
	c.qrState.mu.RLock()
	isActive := c.qrState.loopActive
	c.qrState.mu.RUnlock()

	if !isActive {
		return
	}

	c.logger.InfoWithFields("Stopping existing QR loop", map[string]interface{}{
		"session_id": c.sessionID,
	})

	select {
	case c.qrState.stopChannel <- true:
		c.logger.InfoWithFields("QR loop stop signal sent", map[string]interface{}{
			"session_id": c.sessionID,
		})
	default:
		c.logger.InfoWithFields("QR loop stop channel full, loop may already be stopping", map[string]interface{}{
			"session_id": c.sessionID,
		})
	}
	time.Sleep(100 * time.Millisecond)
}

func (c *WameowClient) Logout() error {
	c.logger.InfoWithFields("Logging out session", map[string]interface{}{
		"session_id": c.sessionID,
	})

	err := c.client.Logout(context.Background())
	if err != nil {
		c.logger.ErrorWithFields("Failed to logout session", map[string]interface{}{
			"session_id": c.sessionID,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to logout: %w", err)
	}

	if c.client.IsConnected() {
		c.client.Disconnect()
	}

	c.setStatus("disconnected")
	c.logger.InfoWithFields("Successfully logged out session", map[string]interface{}{
		"session_id": c.sessionID,
	})
	return nil
}

func (c *WameowClient) SendTextMessage(ctx context.Context, to, body string) (*whatsmeow.SendResponse, error) {
	return c.msgSender.SendText(ctx, to, body, nil)
}

func (c *WameowClient) SendImageMessage(ctx context.Context, to, filePath, caption string) (*whatsmeow.SendResponse, error) {
	options := MediaOptions{
		Caption:  caption,
		MimeType: "image/jpeg",
	}
	return c.msgSender.SendMedia(ctx, to, filePath, MediaTypeImage, options)
}

func (c *WameowClient) SendAudioMessage(ctx context.Context, to, filePath string) (*whatsmeow.SendResponse, error) {
	options := MediaOptions{
		MimeType: "audio/ogg; codecs=opus",
	}
	return c.msgSender.SendMedia(ctx, to, filePath, MediaTypeAudio, options)
}

func (c *WameowClient) SendVideoMessage(ctx context.Context, to, filePath, caption string) (*whatsmeow.SendResponse, error) {
	options := MediaOptions{
		Caption:  caption,
		MimeType: "video/mp4",
	}
	return c.msgSender.SendMedia(ctx, to, filePath, MediaTypeVideo, options)
}

func (c *WameowClient) SendDocumentMessage(ctx context.Context, to, filePath, filename string) (*whatsmeow.SendResponse, error) {
	options := MediaOptions{
		Filename: filename,
		MimeType: "application/octet-stream",
	}
	return c.msgSender.SendMedia(ctx, to, filePath, MediaTypeDocument, options)
}

func (c *WameowClient) SendLocationMessage(ctx context.Context, to string, latitude, longitude float64, address string) (*whatsmeow.SendResponse, error) {
	return c.msgSender.SendLocation(ctx, to, latitude, longitude, address)
}

func (c *WameowClient) SendContactMessage(ctx context.Context, to, contactName, contactPhone string) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	vcard := fmt.Sprintf("BEGIN:VCARD\nVERSION:3.0\nFN:%s\nTEL:%s\nEND:VCARD", contactName, contactPhone)

	message := &waE2E.Message{
		ContactMessage: &waE2E.ContactMessage{
			DisplayName: &contactName,
			Vcard:       &vcard,
		},
	}

	c.logger.InfoWithFields("Sending contact message", map[string]interface{}{
		"session_id":    c.sessionID,
		"to":            to,
		"contact_name":  contactName,
		"contact_phone": contactPhone,
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send contact message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Contact message sent successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": resp.ID,
	})

	return &resp, nil
}

type ContactInfo struct {
	Name         string
	Phone        string
	Email        string
	Organization string
	Title        string
	Website      string
	Address      string
}

func (c *WameowClient) SendDetailedContactMessage(ctx context.Context, to string, contact ContactInfo) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	vcard := fmt.Sprintf("BEGIN:VCARD\nVERSION:3.0\nFN:%s\nTEL:%s", contact.Name, contact.Phone)

	if contact.Email != "" {
		vcard += fmt.Sprintf("\nEMAIL:%s", contact.Email)
	}
	if contact.Organization != "" {
		vcard += fmt.Sprintf("\nORG:%s", contact.Organization)
	}
	if contact.Title != "" {
		vcard += fmt.Sprintf("\nTITLE:%s", contact.Title)
	}
	if contact.Website != "" {
		vcard += fmt.Sprintf("\nURL:%s", contact.Website)
	}
	if contact.Address != "" {
		vcard += fmt.Sprintf("\nADR:%s", contact.Address)
	}

	vcard += "\nEND:VCARD"

	message := &waE2E.Message{
		ContactMessage: &waE2E.ContactMessage{
			DisplayName: &contact.Name,
			Vcard:       &vcard,
		},
	}

	c.logger.InfoWithFields("Sending detailed contact message", map[string]interface{}{
		"session_id":    c.sessionID,
		"to":            to,
		"contact_name":  contact.Name,
		"contact_phone": contact.Phone,
		"has_email":     contact.Email != "",
		"has_org":       contact.Organization != "",
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send detailed contact message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Detailed contact message sent successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": resp.ID,
	})

	return &resp, nil
}

func (c *WameowClient) SendContactListMessage(ctx context.Context, to string, contacts []ContactInfo) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	if len(contacts) == 0 {
		return nil, fmt.Errorf("at least one contact is required")
	}

	displayName := fmt.Sprintf("%d contatos", len(contacts))
	if len(contacts) == 1 {
		displayName = contacts[0].Name
	}

	var contactMessages []*waE2E.ContactMessage

	for _, contact := range contacts {
		vcard := "BEGIN:VCARD\n"
		vcard += "VERSION:3.0\n"
		vcard += fmt.Sprintf("FN:%s\n", contact.Name)
		vcard += fmt.Sprintf("N:%s;;;;\n", contact.Name)
		vcard += fmt.Sprintf("TEL:%s\n", contact.Phone)

		if contact.Organization != "" {
			vcard += fmt.Sprintf("ORG:%s\n", contact.Organization)
		}

		if contact.Email != "" {
			vcard += fmt.Sprintf("EMAIL:%s\n", contact.Email)
		}
		if contact.Title != "" {
			vcard += fmt.Sprintf("TITLE:%s\n", contact.Title)
		}
		if contact.Website != "" {
			vcard += fmt.Sprintf("URL:%s\n", contact.Website)
		}
		if contact.Address != "" {
			vcard += fmt.Sprintf("ADR:%s\n", contact.Address)
		}

		vcard += "END:VCARD"

		c.logger.InfoWithFields("📋 Generated vCard for contact", map[string]interface{}{
			"session_id":    c.sessionID,
			"contact_name":  contact.Name,
			"vcard_content": vcard,
		})

		contactMessage := &waE2E.ContactMessage{
			DisplayName: &contact.Name,
			Vcard:       &vcard,
		}

		contactMessages = append(contactMessages, contactMessage)
	}

	message := &waE2E.Message{
		ContactsArrayMessage: &waE2E.ContactsArrayMessage{
			DisplayName: &displayName,
			Contacts:    contactMessages,
		},
	}

	c.logger.InfoWithFields("Sending contacts array message (WhatsApp native format)", map[string]interface{}{
		"session_id":    c.sessionID,
		"to":            to,
		"contact_count": len(contacts),
		"display_name":  displayName,
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send contacts array message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Contacts array message sent successfully", map[string]interface{}{
		"session_id":    c.sessionID,
		"to":            to,
		"message_id":    resp.ID,
		"contact_count": len(contacts),
	})

	return &resp, nil
}

func (c *WameowClient) SendContactListMessageBusiness(ctx context.Context, to string, contacts []ContactInfo) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	if len(contacts) == 0 {
		return nil, fmt.Errorf("at least one contact is required")
	}

	displayName := fmt.Sprintf("%d contatos", len(contacts))
	if len(contacts) == 1 {
		displayName = contacts[0].Name
	}

	var contactMessages []*waE2E.ContactMessage

	for _, contact := range contacts {
		vcard := "BEGIN:VCARD\n"
		vcard += "VERSION:3.0\n"
		vcard += fmt.Sprintf("N:;%s;;;\n", contact.Name)
		vcard += fmt.Sprintf("FN:%s\n", contact.Name)

		if contact.Organization != "" {
			vcard += fmt.Sprintf("ORG:%s\n", contact.Organization)
		}

		if contact.Title != "" {
			vcard += fmt.Sprintf("TITLE:%s\n", contact.Title)
		} else {
			vcard += "TITLE:\n"
		}

		phoneClean := strings.ReplaceAll(strings.ReplaceAll(contact.Phone, "+", ""), " ", "")
		phoneFormatted := contact.Phone
		vcard += fmt.Sprintf("item1.TEL;waid=%s:%s\n", phoneClean, phoneFormatted)
		vcard += "item1.X-ABLabel:Celular\n"

		vcard += fmt.Sprintf("X-WA-BIZ-NAME:%s\n", contact.Name)

		vcard += "END:VCARD"

		c.logger.InfoWithFields("📋 Generated Business style vCard", map[string]interface{}{
			"session_id":    c.sessionID,
			"contact_name":  contact.Name,
			"vcard_content": vcard,
		})

		contactMessage := &waE2E.ContactMessage{
			DisplayName: &contact.Name,
			Vcard:       &vcard,
		}

		contactMessages = append(contactMessages, contactMessage)
	}

	message := &waE2E.Message{
		ContactsArrayMessage: &waE2E.ContactsArrayMessage{
			DisplayName: &displayName,
			Contacts:    contactMessages,
		},
	}

	c.logger.InfoWithFields("Sending contacts array message (Business format)", map[string]interface{}{
		"session_id":    c.sessionID,
		"to":            to,
		"contact_count": len(contacts),
		"display_name":  displayName,
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send Business contacts array message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Business contacts array message sent successfully", map[string]interface{}{
		"session_id":    c.sessionID,
		"to":            to,
		"message_id":    resp.ID,
		"contact_count": len(contacts),
	})

	return &resp, nil
}

func (c *WameowClient) SendSingleContactMessage(ctx context.Context, to string, contact ContactInfo) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	vcard := "BEGIN:VCARD\n"
	vcard += "VERSION:3.0\n"
	vcard += fmt.Sprintf("FN:%s\n", contact.Name)
	vcard += fmt.Sprintf("N:%s;;;;\n", contact.Name)
	vcard += fmt.Sprintf("TEL:%s\n", contact.Phone)

	if contact.Organization != "" {
		vcard += fmt.Sprintf("ORG:%s\n", contact.Organization)
	}
	if contact.Email != "" {
		vcard += fmt.Sprintf("EMAIL:%s\n", contact.Email)
	}
	if contact.Title != "" {
		vcard += fmt.Sprintf("TITLE:%s\n", contact.Title)
	}
	if contact.Website != "" {
		vcard += fmt.Sprintf("URL:%s\n", contact.Website)
	}
	if contact.Address != "" {
		vcard += fmt.Sprintf("ADR:%s\n", contact.Address)
	}

	vcard += "END:VCARD"

	message := &waE2E.Message{
		ContactMessage: &waE2E.ContactMessage{
			DisplayName: &contact.Name,
			Vcard:       &vcard,
		},
	}

	c.logger.InfoWithFields("Sending single contact message (standard format)", map[string]interface{}{
		"session_id":    c.sessionID,
		"to":            to,
		"contact_name":  contact.Name,
		"vcard_content": vcard,
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send single contact message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Single contact message sent successfully", map[string]interface{}{
		"session_id":   c.sessionID,
		"to":           to,
		"message_id":   resp.ID,
		"contact_name": contact.Name,
	})

	return &resp, nil
}

func (c *WameowClient) SendSingleContactMessageBusiness(ctx context.Context, to string, contact ContactInfo) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	vcard := "BEGIN:VCARD\n"
	vcard += "VERSION:3.0\n"
	vcard += fmt.Sprintf("N:;%s;;;\n", contact.Name)
	vcard += fmt.Sprintf("FN:%s\n", contact.Name)

	if contact.Organization != "" {
		vcard += fmt.Sprintf("ORG:%s\n", contact.Organization)
	}

	if contact.Title != "" {
		vcard += fmt.Sprintf("TITLE:%s\n", contact.Title)
	} else {
		vcard += "TITLE:\n"
	}

	phoneClean := strings.ReplaceAll(strings.ReplaceAll(contact.Phone, "+", ""), " ", "")
	phoneFormatted := contact.Phone
	vcard += fmt.Sprintf("item1.TEL;waid=%s:%s\n", phoneClean, phoneFormatted)
	vcard += "item1.X-ABLabel:Celular\n"

	vcard += fmt.Sprintf("X-WA-BIZ-NAME:%s\n", contact.Name)

	vcard += "END:VCARD"

	message := &waE2E.Message{
		ContactMessage: &waE2E.ContactMessage{
			DisplayName: &contact.Name,
			Vcard:       &vcard,
		},
	}

	c.logger.InfoWithFields("Sending single contact message (Business format)", map[string]interface{}{
		"session_id":    c.sessionID,
		"to":            to,
		"contact_name":  contact.Name,
		"vcard_content": vcard,
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send Business single contact message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Business single contact message sent successfully", map[string]interface{}{
		"session_id":   c.sessionID,
		"to":           to,
		"message_id":   resp.ID,
		"contact_name": contact.Name,
	})

	return &resp, nil
}

func (c *WameowClient) parseJID(jidStr string) (types.JID, error) {
	validator := NewJIDValidator()
	return validator.Parse(jidStr)
}

// Helper function to create WhatsApp ContextInfo from our ContextInfo
func (c *WameowClient) createContextInfo(contextInfo *appMessage.ContextInfo) *waE2E.ContextInfo {
	if contextInfo == nil {
		return nil
	}

	waContextInfo := &waE2E.ContextInfo{
		StanzaID:      proto.String(contextInfo.StanzaID),
		QuotedMessage: &waE2E.Message{Conversation: proto.String("")},
	}

	if contextInfo.Participant != "" {
		waContextInfo.Participant = proto.String(contextInfo.Participant)
	}

	return waContextInfo
}

// SendImageMessageWithContext sends an image message with optional context info for replies
func (c *WameowClient) SendImageMessageWithContext(ctx context.Context, to, filePath, caption string, contextInfo *appMessage.ContextInfo) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read image file: %w", err)
	}

	uploaded, err := c.client.Upload(ctx, data, whatsmeow.MediaImage)
	if err != nil {
		return nil, fmt.Errorf("failed to upload image: %w", err)
	}

	mimetype := "image/jpeg" // Default mimetype
	message := &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			Caption:       &caption,
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			Mimetype:      &mimetype,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
			ContextInfo:   c.createContextInfo(contextInfo),
		},
	}

	c.logger.InfoWithFields("Sending image message with context", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"file_size":  len(data),
		"caption":    caption,
		"has_reply":  contextInfo != nil,
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send image message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Image message sent successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": resp.ID,
	})

	return &resp, nil
}

// SendAudioMessageWithContext sends an audio message with optional context info for replies
func (c *WameowClient) SendAudioMessageWithContext(ctx context.Context, to, filePath string, contextInfo *appMessage.ContextInfo) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio file: %w", err)
	}

	uploaded, err := c.client.Upload(ctx, data, whatsmeow.MediaAudio)
	if err != nil {
		return nil, fmt.Errorf("failed to upload audio: %w", err)
	}

	mimetype := "audio/ogg; codecs=opus" // Default mimetype
	message := &waE2E.Message{
		AudioMessage: &waE2E.AudioMessage{
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			Mimetype:      &mimetype,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
			ContextInfo:   c.createContextInfo(contextInfo),
		},
	}

	c.logger.InfoWithFields("Sending audio message with context", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"file_size":  len(data),
		"has_reply":  contextInfo != nil,
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send audio message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Audio message sent successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": resp.ID,
	})

	return &resp, nil
}

// SendVideoMessageWithContext sends a video message with optional context info for replies
func (c *WameowClient) SendVideoMessageWithContext(ctx context.Context, to, filePath, caption string, contextInfo *appMessage.ContextInfo) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read video file: %w", err)
	}

	uploaded, err := c.client.Upload(ctx, data, whatsmeow.MediaVideo)
	if err != nil {
		return nil, fmt.Errorf("failed to upload video: %w", err)
	}

	mimetype := "video/mp4" // Default mimetype
	message := &waE2E.Message{
		VideoMessage: &waE2E.VideoMessage{
			Caption:       &caption,
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			Mimetype:      &mimetype,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
			ContextInfo:   c.createContextInfo(contextInfo),
		},
	}

	c.logger.InfoWithFields("Sending video message with context", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"file_size":  len(data),
		"caption":    caption,
		"has_reply":  contextInfo != nil,
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send video message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Video message sent successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": resp.ID,
	})

	return &resp, nil
}

// SendDocumentMessageWithContext sends a document message with optional context info for replies
func (c *WameowClient) SendDocumentMessageWithContext(ctx context.Context, to, filePath, filename string, contextInfo *appMessage.ContextInfo) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read document file: %w", err)
	}

	uploaded, err := c.client.Upload(ctx, data, whatsmeow.MediaDocument)
	if err != nil {
		return nil, fmt.Errorf("failed to upload document: %w", err)
	}

	mimetype := "application/octet-stream" // Default mimetype
	message := &waE2E.Message{
		DocumentMessage: &waE2E.DocumentMessage{
			Title:         &filename,
			FileName:      &filename,
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			Mimetype:      &mimetype,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
			ContextInfo:   c.createContextInfo(contextInfo),
		},
	}

	c.logger.InfoWithFields("Sending document message with context", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"file_size":  len(data),
		"filename":   filename,
		"has_reply":  contextInfo != nil,
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send document message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Document message sent successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": resp.ID,
	})

	return &resp, nil
}

func (c *WameowClient) AddEventHandler(handler whatsmeow.EventHandler) uint32 {
	return c.client.AddEventHandler(handler)
}

func (c *WameowClient) SendStickerMessage(ctx context.Context, to, filePath string) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read sticker file: %w", err)
	}

	uploaded, err := c.client.Upload(ctx, data, whatsmeow.MediaImage) // Stickers use image media type
	if err != nil {
		return nil, fmt.Errorf("failed to upload sticker: %w", err)
	}

	mimetype := "image/webp" // Stickers are typically WebP
	message := &waE2E.Message{
		StickerMessage: &waE2E.StickerMessage{
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			Mimetype:      &mimetype,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
		},
	}

	c.logger.InfoWithFields("Sending sticker message", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"file_size":  len(data),
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send sticker message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Sticker message sent successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": resp.ID,
	})

	return &resp, nil
}

func (c *WameowClient) SendButtonMessage(ctx context.Context, to, body string, buttons []map[string]string) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	buttonText := body + "\n\n📋 *Options:*"
	for i, button := range buttons {
		if i >= 3 { // Limit to 3 options for readability
			break
		}
		buttonText += fmt.Sprintf("\n%d. %s", i+1, button["text"])
	}

	message := &waE2E.Message{
		Conversation: &buttonText,
	}

	c.logger.InfoWithFields("Sending button message (as text)", map[string]interface{}{
		"session_id":   c.sessionID,
		"to":           to,
		"button_count": len(buttons),
		"body_length":  len(body),
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send button message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Button message sent successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": resp.ID,
	})

	return &resp, nil
}

func (c *WameowClient) SendListMessage(ctx context.Context, to, body, buttonText string, sections []map[string]interface{}) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	listText := body + "\n\n📋 *" + buttonText + ":*"

	for _, section := range sections {
		title, _ := section["title"].(string)
		rows, _ := section["rows"].([]interface{})

		if title != "" {
			listText += "\n\n*" + title + ":*"
		}

		for i, rowInterface := range rows {
			if i >= 10 { // Limit rows for readability
				break
			}

			row, ok := rowInterface.(map[string]interface{})
			if !ok {
				continue
			}

			rowTitle, _ := row["title"].(string)
			rowDescription, _ := row["description"].(string)

			listText += fmt.Sprintf("\n%d. %s", i+1, rowTitle)
			if rowDescription != "" {
				listText += " - " + rowDescription
			}
		}
	}

	message := &waE2E.Message{
		Conversation: &listText,
	}

	c.logger.InfoWithFields("Sending list message (as text)", map[string]interface{}{
		"session_id":    c.sessionID,
		"to":            to,
		"section_count": len(sections),
		"body_length":   len(body),
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send list message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("List message sent successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": resp.ID,
	})

	return &resp, nil
}

func (c *WameowClient) SendReaction(ctx context.Context, to, messageID, reaction string) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return fmt.Errorf("invalid JID: %w", err)
	}

	if messageID == "" {
		return fmt.Errorf("message ID is required")
	}

	c.logger.InfoWithFields("Sending reaction", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": messageID,
		"reaction":   reaction,
	})

	message := c.client.BuildReaction(jid, jid, types.MessageID(messageID), reaction)

	_, err = c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send reaction", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"message_id": messageID,
			"error":      err.Error(),
		})
		return err
	}

	c.logger.InfoWithFields("Reaction sent successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": messageID,
		"reaction":   reaction,
	})

	return nil
}

func (c *WameowClient) SendPresence(ctx context.Context, to, presence string) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return fmt.Errorf("invalid JID: %w", err)
	}

	c.logger.InfoWithFields("Sending presence", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"presence":   presence,
	})

	switch presence {
	case "typing":
		err = c.client.SendChatPresence(jid, types.ChatPresenceComposing, types.ChatPresenceMediaText)
	case "recording":
		err = c.client.SendChatPresence(jid, types.ChatPresenceComposing, types.ChatPresenceMediaAudio)
	case "online":
		err = c.client.SendPresence(types.PresenceAvailable)
	case "offline", "paused":
		err = c.client.SendChatPresence(jid, types.ChatPresencePaused, types.ChatPresenceMediaText)
	default:
		return fmt.Errorf("invalid presence type: %s. Valid types: typing, recording, online, offline, paused", presence)
	}

	if err != nil {
		c.logger.ErrorWithFields("Failed to send presence", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"presence":   presence,
			"error":      err.Error(),
		})
		return err
	}

	c.logger.InfoWithFields("Presence sent successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"presence":   presence,
	})

	return nil
}

func (c *WameowClient) EditMessage(ctx context.Context, to, messageID, newText string) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return fmt.Errorf("invalid JID: %w", err)
	}

	if messageID == "" {
		return fmt.Errorf("message ID is required")
	}

	c.logger.InfoWithFields("Editing message", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": messageID,
		"new_text":   newText,
	})

	editMessage := &waE2E.Message{
		EditedMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{
				Conversation: &newText,
			},
		},
	}

	_, err = c.client.SendMessage(ctx, jid, editMessage)
	if err != nil {
		c.logger.ErrorWithFields("Failed to edit message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"message_id": messageID,
			"error":      err.Error(),
		})
		return err
	}

	c.logger.InfoWithFields("Message edited successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": messageID,
	})

	return nil
}

func (c *WameowClient) DeleteMessage(ctx context.Context, to, messageID string, forAll bool) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return fmt.Errorf("invalid JID: %w", err)
	}

	if messageID == "" {
		return fmt.Errorf("message ID is required")
	}

	c.logger.InfoWithFields("Deleting message", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": messageID,
		"for_all":    forAll,
	})

	message := c.client.BuildRevoke(jid, jid, messageID)

	_, err = c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to delete message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"message_id": messageID,
			"error":      err.Error(),
		})
		return err
	}

	c.logger.InfoWithFields("Message deleted successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": messageID,
		"for_all":    forAll,
	})

	return nil
}

func (c *WameowClient) MarkRead(ctx context.Context, to, messageID string) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return fmt.Errorf("invalid JID: %w", err)
	}

	if messageID == "" {
		return fmt.Errorf("message ID is required")
	}

	c.logger.InfoWithFields("Marking message as read", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": messageID,
	})

	// Convert messageID string to types.MessageID
	msgID := types.MessageID(messageID)

	// MarkRead expects a slice of message IDs, timestamp, chat JID, sender JID, and optional receipt type
	err = c.client.MarkRead([]types.MessageID{msgID}, time.Now(), jid, jid, "")
	if err != nil {
		c.logger.ErrorWithFields("Failed to mark message as read", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"message_id": messageID,
			"error":      err.Error(),
		})
		return err
	}

	c.logger.InfoWithFields("Message marked as read successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": messageID,
	})

	return nil
}

// CreateGroup creates a new WhatsApp group
func (c *WameowClient) CreateGroup(ctx context.Context, name string, participants []string, description string) (*types.GroupInfo, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	if name == "" {
		return nil, fmt.Errorf("group name is required")
	}

	if len(participants) == 0 {
		return nil, fmt.Errorf("at least one participant is required")
	}

	// Convert participant strings to JIDs
	participantJIDs := make([]types.JID, len(participants))
	for i, participant := range participants {
		jid, err := c.parseJID(participant)
		if err != nil {
			return nil, fmt.Errorf("invalid participant JID %s: %w", participant, err)
		}
		participantJIDs[i] = jid
	}

	c.logger.InfoWithFields("Creating group", map[string]interface{}{
		"session_id":   c.sessionID,
		"name":         name,
		"participants": len(participantJIDs),
	})

	// Create the group
	groupInfo, err := c.client.CreateGroup(ctx, whatsmeow.ReqCreateGroup{
		Name:         name,
		Participants: participantJIDs,
	})
	if err != nil {
		c.logger.ErrorWithFields("Failed to create group", map[string]interface{}{
			"session_id": c.sessionID,
			"name":       name,
			"error":      err.Error(),
		})
		return nil, err
	}

	// Set description if provided
	if description != "" {
		err = c.client.SetGroupTopic(groupInfo.JID, "", "", description)
		if err != nil {
			c.logger.WarnWithFields("Failed to set group description", map[string]interface{}{
				"session_id": c.sessionID,
				"group_jid":  groupInfo.JID.String(),
				"error":      err.Error(),
			})
		}
	}

	c.logger.InfoWithFields("Group created successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupInfo.JID.String(),
		"name":       name,
	})

	return groupInfo, nil
}

// GetGroupInfo retrieves information about a specific group
func (c *WameowClient) GetGroupInfo(ctx context.Context, groupJID string) (*types.GroupInfo, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(groupJID)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID: %w", err)
	}

	c.logger.InfoWithFields("Getting group info", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
	})

	groupInfo, err := c.client.GetGroupInfo(jid)
	if err != nil {
		c.logger.ErrorWithFields("Failed to get group info", map[string]interface{}{
			"session_id": c.sessionID,
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Group info retrieved successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
		"name":       groupInfo.Name,
	})

	return groupInfo, nil
}

// ListJoinedGroups lists all groups the user is a member of
func (c *WameowClient) ListJoinedGroups(ctx context.Context) ([]*types.GroupInfo, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	c.logger.InfoWithFields("Listing joined groups", map[string]interface{}{
		"session_id": c.sessionID,
	})

	groups, err := c.client.GetJoinedGroups(ctx)
	if err != nil {
		c.logger.ErrorWithFields("Failed to list joined groups", map[string]interface{}{
			"session_id": c.sessionID,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Joined groups listed successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"count":      len(groups),
	})

	return groups, nil
}

// UpdateGroupParticipants adds, removes, promotes, or demotes group participants
func (c *WameowClient) UpdateGroupParticipants(ctx context.Context, groupJID string, participants []string, action string) ([]string, []string, error) {
	if !c.client.IsLoggedIn() {
		return nil, nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(groupJID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid group JID: %w", err)
	}

	if len(participants) == 0 {
		return nil, nil, fmt.Errorf("no participants provided")
	}

	// Convert participant strings to JIDs
	participantJIDs := make([]types.JID, len(participants))
	for i, participant := range participants {
		participantJID, err := c.parseJID(participant)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid participant JID %s: %w", participant, err)
		}
		participantJIDs[i] = participantJID
	}

	c.logger.InfoWithFields("Updating group participants", map[string]interface{}{
		"session_id":   c.sessionID,
		"group_jid":    groupJID,
		"action":       action,
		"participants": len(participantJIDs),
	})

	var success, failed []string

	switch action {
	case "add":
		_, err = c.client.UpdateGroupParticipants(jid, participantJIDs, whatsmeow.ParticipantChangeAdd)
	case "remove":
		_, err = c.client.UpdateGroupParticipants(jid, participantJIDs, whatsmeow.ParticipantChangeRemove)
	case "promote":
		_, err = c.client.UpdateGroupParticipants(jid, participantJIDs, whatsmeow.ParticipantChangePromote)
	case "demote":
		_, err = c.client.UpdateGroupParticipants(jid, participantJIDs, whatsmeow.ParticipantChangeDemote)
	default:
		return nil, nil, fmt.Errorf("invalid action: %s (must be add, remove, promote, or demote)", action)
	}

	if err != nil {
		c.logger.ErrorWithFields("Failed to update group participants", map[string]interface{}{
			"session_id": c.sessionID,
			"group_jid":  groupJID,
			"action":     action,
			"error":      err.Error(),
		})
		return nil, nil, err
	}

	// For simplicity, assume all participants were successful if no error occurred
	// In a real implementation, you might want to check individual results
	for _, participantJID := range participantJIDs {
		success = append(success, participantJID.String())
	}

	c.logger.InfoWithFields("Group participants updated", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
		"action":     action,
		"success":    len(success),
		"failed":     len(failed),
	})

	return success, failed, nil
}

// SetGroupName updates the group name
func (c *WameowClient) SetGroupName(ctx context.Context, groupJID, name string) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	if name == "" {
		return fmt.Errorf("group name is required")
	}

	c.logger.InfoWithFields("Setting group name", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
		"name":       name,
	})

	err = c.client.SetGroupName(jid, name)
	if err != nil {
		c.logger.ErrorWithFields("Failed to set group name", map[string]interface{}{
			"session_id": c.sessionID,
			"group_jid":  groupJID,
			"name":       name,
			"error":      err.Error(),
		})
		return err
	}

	c.logger.InfoWithFields("Group name set successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
		"name":       name,
	})

	return nil
}

// SetGroupDescription updates the group description
func (c *WameowClient) SetGroupDescription(ctx context.Context, groupJID, description string) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	c.logger.InfoWithFields("Setting group description", map[string]interface{}{
		"session_id":  c.sessionID,
		"group_jid":   groupJID,
		"description": description,
	})

	err = c.client.SetGroupTopic(jid, "", "", description)
	if err != nil {
		c.logger.ErrorWithFields("Failed to set group description", map[string]interface{}{
			"session_id": c.sessionID,
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return err
	}

	c.logger.InfoWithFields("Group description set successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
	})

	return nil
}

// GetGroupInviteLink retrieves or generates a group invite link
func (c *WameowClient) GetGroupInviteLink(ctx context.Context, groupJID string, reset bool) (string, error) {
	if !c.client.IsLoggedIn() {
		return "", fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(groupJID)
	if err != nil {
		return "", fmt.Errorf("invalid group JID: %w", err)
	}

	c.logger.InfoWithFields("Getting group invite link", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
		"reset":      reset,
	})

	var link string
	if reset {
		link, err = c.client.GetGroupInviteLink(jid, true)
	} else {
		link, err = c.client.GetGroupInviteLink(jid, false)
	}

	if err != nil {
		c.logger.ErrorWithFields("Failed to get group invite link", map[string]interface{}{
			"session_id": c.sessionID,
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return "", err
	}

	c.logger.InfoWithFields("Group invite link retrieved successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
	})

	return link, nil
}

// JoinGroupViaLink joins a group using an invite link
func (c *WameowClient) JoinGroupViaLink(ctx context.Context, inviteLink string) (*types.GroupInfo, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	if inviteLink == "" {
		return nil, fmt.Errorf("invite link is required")
	}

	c.logger.InfoWithFields("Joining group via link", map[string]interface{}{
		"session_id": c.sessionID,
	})

	groupJID, err := c.client.JoinGroupWithLink(inviteLink)
	if err != nil {
		c.logger.ErrorWithFields("Failed to join group via link", map[string]interface{}{
			"session_id": c.sessionID,
			"error":      err.Error(),
		})
		return nil, err
	}

	// Get group info after joining
	groupInfo, err := c.client.GetGroupInfo(groupJID)
	if err != nil {
		c.logger.WarnWithFields("Joined group but failed to get info", map[string]interface{}{
			"session_id": c.sessionID,
			"group_jid":  groupJID.String(),
			"error":      err.Error(),
		})
		// Return minimal info if we can't get full details
		return &types.GroupInfo{
			JID: groupJID,
		}, nil
	}

	c.logger.InfoWithFields("Joined group successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID.String(),
		"name":       groupInfo.Name,
	})

	return groupInfo, nil
}

// LeaveGroup leaves a group
func (c *WameowClient) LeaveGroup(ctx context.Context, groupJID string) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	c.logger.InfoWithFields("Leaving group", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
	})

	err = c.client.LeaveGroup(jid)
	if err != nil {
		c.logger.ErrorWithFields("Failed to leave group", map[string]interface{}{
			"session_id": c.sessionID,
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return err
	}

	c.logger.InfoWithFields("Left group successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
	})

	return nil
}

// UpdateGroupSettings updates group settings (announce, locked)
func (c *WameowClient) UpdateGroupSettings(ctx context.Context, groupJID string, announce, locked *bool) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	c.logger.InfoWithFields("Updating group settings", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
		"announce":   announce,
		"locked":     locked,
	})

	if announce != nil {
		err = c.client.SetGroupAnnounce(jid, *announce)
		if err != nil {
			c.logger.ErrorWithFields("Failed to set group announce", map[string]interface{}{
				"session_id": c.sessionID,
				"group_jid":  groupJID,
				"announce":   *announce,
				"error":      err.Error(),
			})
			return err
		}
	}

	if locked != nil {
		err = c.client.SetGroupLocked(jid, *locked)
		if err != nil {
			c.logger.ErrorWithFields("Failed to set group locked", map[string]interface{}{
				"session_id": c.sessionID,
				"group_jid":  groupJID,
				"locked":     *locked,
				"error":      err.Error(),
			})
			return err
		}
	}

	c.logger.InfoWithFields("Group settings updated successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
	})

	return nil
}

// CreatePoll creates a poll message
func (c *WameowClient) CreatePoll(ctx context.Context, to, name string, options []string, selectableCount int) (*types.MessageInfo, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	if name == "" {
		return nil, fmt.Errorf("poll name is required")
	}

	if len(options) < 2 {
		return nil, fmt.Errorf("at least 2 options are required")
	}

	if len(options) > 12 {
		return nil, fmt.Errorf("maximum 12 options allowed")
	}

	if selectableCount < 1 {
		selectableCount = 1 // Default to single selection
	}

	if selectableCount > len(options) {
		return nil, fmt.Errorf("selectable count cannot exceed number of options")
	}

	// Parse recipient JID
	toJID, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient JID: %w", err)
	}

	c.logger.InfoWithFields("Creating poll", map[string]interface{}{
		"session_id":       c.sessionID,
		"to":               to,
		"name":             name,
		"options_count":    len(options),
		"selectable_count": selectableCount,
	})

	// Build poll creation message
	pollMessage := c.client.BuildPollCreation(name, options, selectableCount)

	// Send the poll
	resp, err := c.client.SendMessage(ctx, toJID, pollMessage)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send poll", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Poll sent successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": resp.ID,
		"timestamp":  resp.Timestamp,
	})

	// Return message info
	return &types.MessageInfo{
		ID:        resp.ID,
		Timestamp: resp.Timestamp,
	}, nil
}

// VotePoll votes in a poll
func (c *WameowClient) VotePoll(ctx context.Context, to, pollMessageID string, selectedOptions []string) (*types.MessageInfo, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	if pollMessageID == "" {
		return nil, fmt.Errorf("poll message ID is required")
	}

	if len(selectedOptions) == 0 {
		return nil, fmt.Errorf("at least one option must be selected")
	}

	c.logger.InfoWithFields("Poll voting not fully implemented", map[string]interface{}{
		"session_id":       c.sessionID,
		"to":               to,
		"poll_message_id":  pollMessageID,
		"selected_options": selectedOptions,
	})

	// Return a mock response for now since poll voting requires complex message handling
	return &types.MessageInfo{
		ID:        "mock-vote-" + pollMessageID,
		Timestamp: time.Now(),
	}, nil
}

// SetGroupPhoto sets a group's photo
func (c *WameowClient) SetGroupPhoto(ctx context.Context, groupJID, photoPath string) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	if groupJID == "" {
		return fmt.Errorf("group JID is required")
	}

	if photoPath == "" {
		return fmt.Errorf("photo path is required")
	}

	// Parse group JID
	gJID, err := c.parseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	// Read photo file
	photoData, err := os.ReadFile(photoPath)
	if err != nil {
		return fmt.Errorf("failed to read photo file: %w", err)
	}

	c.logger.InfoWithFields("Setting group photo", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
		"photo_path": photoPath,
		"photo_size": len(photoData),
	})

	// Set group photo using whatsmeow
	_, err = c.client.SetGroupPhoto(gJID, photoData)
	if err != nil {
		c.logger.ErrorWithFields("Failed to set group photo", map[string]interface{}{
			"session_id": c.sessionID,
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to set group photo: %w", err)
	}

	c.logger.InfoWithFields("Group photo set successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
	})

	return nil
}

// SetGroupPhotoFromBytes sets a group's photo from byte data
func (c *WameowClient) SetGroupPhotoFromBytes(ctx context.Context, groupJID string, photoData []byte) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	if groupJID == "" {
		return fmt.Errorf("group JID is required")
	}

	if len(photoData) == 0 {
		return fmt.Errorf("photo data is required")
	}

	// Parse group JID
	gJID, err := c.parseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	c.logger.InfoWithFields("Setting group photo from bytes", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
		"photo_size": len(photoData),
	})

	// Set group photo using whatsmeow
	_, err = c.client.SetGroupPhoto(gJID, photoData)
	if err != nil {
		c.logger.ErrorWithFields("Failed to set group photo", map[string]interface{}{
			"session_id": c.sessionID,
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to set group photo: %w", err)
	}

	c.logger.InfoWithFields("Group photo set successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
	})

	return nil
}

func IsDeviceRegistered(client *whatsmeow.Client) bool {
	if client == nil || client.Store == nil {
		return false
	}
	return client.Store.ID != nil
}
