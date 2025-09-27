package wameow

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"zpwoot/internal/ports"
	"zpwoot/platform/logger"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waTypes "go.mau.fi/whatsmeow/types"
)

// WameowClient wraps whatsmeow.Client with additional functionality
type WameowClient struct {
	sessionID   string
	client      *whatsmeow.Client
	logger      *logger.Logger
	sessionMgr  *SessionManager
	qrGenerator *QRCodeGenerator

	mu           sync.RWMutex
	status       string
	lastActivity time.Time

	qrCode       string
	qrCodeBase64 string
	qrLoopActive bool

	ctx           context.Context
	cancel        context.CancelFunc
	qrStopChannel chan bool
}

// NewWameowClient creates a new WameowClient
func NewWameowClient(
	sessionID string,
	container *sqlstore.Container,
	sessionRepo ports.SessionRepository,
	logger *logger.Logger,
) (*WameowClient, error) {
	// Get session from repository to check for existing deviceJid
	ctx := context.Background()
	sess, err := sessionRepo.GetByID(ctx, sessionID)
	var deviceJid string
	if err == nil && sess != nil {
		deviceJid = sess.DeviceJid
		logger.InfoWithFields("Found existing session", map[string]interface{}{
			"session_id": sessionID,
			"device_jid": deviceJid,
		})
	} else {
		logger.InfoWithFields("Creating new session", map[string]interface{}{
			"session_id": sessionID,
		})
	}

	// Get device store for session with the correct deviceJid
	deviceStore := GetDeviceStoreForSession(sessionID, deviceJid, container)
	if deviceStore == nil {
		return nil, fmt.Errorf("failed to create device store for session %s", sessionID)
	}

	// Create whatsmeow logger
	waLogger := NewWameowLogger(logger)

	// Create whatsmeow client
	client := whatsmeow.NewClient(deviceStore, waLogger)
	if client == nil {
		return nil, fmt.Errorf("failed to create WhatsApp client for session %s", sessionID)
	}

	ctx, cancel := context.WithCancel(context.Background())

	wameowClient := &WameowClient{
		sessionID:     sessionID,
		client:        client,
		logger:        logger,
		sessionMgr:    NewSessionManager(sessionRepo, logger),
		qrGenerator:   NewQRCodeGenerator(logger),
		status:        "disconnected",
		lastActivity:  time.Now(),
		ctx:           ctx,
		cancel:        cancel,
		qrStopChannel: make(chan bool, 1),
	}

	return wameowClient, nil
}

// Connect starts the connection process
func (c *WameowClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.InfoWithFields("Starting connection process (will restart if already running)", map[string]interface{}{
		"session_id": c.sessionID,
	})

	// Always stop any existing QR loop first
	c.stopQRLoop()

	// If client is connected, disconnect first to restart the process
	if c.client.IsConnected() {
		c.logger.InfoWithFields("Client already connected, disconnecting to restart", map[string]interface{}{
			"session_id": c.sessionID,
		})
		c.client.Disconnect()
	}

	// Cancel any existing context and create a new one
	if c.cancel != nil {
		c.cancel()
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())

	c.setStatus("connecting")

	// Start connection process in background
	go c.startClientLoop()

	return nil
}

// Disconnect stops the connection
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

// IsConnected returns connection status
func (c *WameowClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.client.IsConnected()
}

// IsLoggedIn returns login status
func (c *WameowClient) IsLoggedIn() bool {
	return c.client.IsLoggedIn()
}

// GetQRCode returns the current QR code
func (c *WameowClient) GetQRCode() (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.qrCode == "" {
		return "", fmt.Errorf("no QR code available")
	}

	return c.qrCode, nil
}

// GetClient returns the underlying whatsmeow client
func (c *WameowClient) GetClient() *whatsmeow.Client {
	return c.client
}

// GetJID returns the device JID
func (c *WameowClient) GetJID() waTypes.JID {
	if c.client.Store.ID == nil {
		return waTypes.EmptyJID
	}
	return *c.client.Store.ID
}

// setStatus sets the internal status
func (c *WameowClient) setStatus(status string) {
	c.status = status
	c.lastActivity = time.Now()
	c.logger.InfoWithFields("Session status updated", map[string]interface{}{
		"session_id": c.sessionID,
		"status":     status,
	})

	// Update database when status changes to connected or disconnected
	switch status {
	case "connected":
		c.sessionMgr.UpdateConnectionStatus(c.sessionID, true)
	case "disconnected":
		c.sessionMgr.UpdateConnectionStatus(c.sessionID, false)
	}
}

// startClientLoop handles the connection logic
func (c *WameowClient) startClientLoop() {
	defer func() {
		if r := recover(); r != nil {
			c.logger.ErrorWithFields("Client loop panic", map[string]interface{}{
				"session_id": c.sessionID,
				"error":      r,
			})
		}
	}()

	if !IsDeviceRegistered(c.client) {
		c.logger.InfoWithFields("Device not registered, starting QR code process", map[string]interface{}{
			"session_id": c.sessionID,
		})
		c.handleNewDeviceRegistration()
	} else {
		c.logger.InfoWithFields("Device already registered, connecting directly", map[string]interface{}{
			"session_id": c.sessionID,
		})
		c.handleExistingDeviceConnection()
	}
}

// handleNewDeviceRegistration handles QR code generation for new devices
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

// handleExistingDeviceConnection handles connection for registered devices
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
		c.logger.WarnWithFields("Connection attempt completed but client not connected", map[string]interface{}{
			"session_id": c.sessionID,
		})
		c.setStatus("disconnected")
	}
}

// handleQRLoop handles QR code events
func (c *WameowClient) handleQRLoop(qrChan <-chan whatsmeow.QRChannelItem) {
	if qrChan == nil {
		c.logger.ErrorWithFields("QR channel is nil", map[string]interface{}{
			"session_id": c.sessionID,
		})
		return
	}

	c.mu.Lock()
	c.qrLoopActive = true
	c.mu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			c.logger.ErrorWithFields("QR loop panic", map[string]interface{}{
				"session_id": c.sessionID,
				"error":      r,
			})
		}
		c.mu.Lock()
		c.qrLoopActive = false
		c.mu.Unlock()
	}()

	for {
		select {
		case <-c.ctx.Done():
			c.logger.InfoWithFields("QR loop cancelled", map[string]interface{}{
				"session_id": c.sessionID,
			})
			return

		case <-c.qrStopChannel:
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

			switch evt.Event {
			case "code":
				c.mu.Lock()
				c.qrCode = evt.Code
				if c.qrGenerator != nil {
					c.qrCodeBase64 = c.qrGenerator.GenerateQRCodeImage(evt.Code)
				}
				c.mu.Unlock()

				// Display compact QR code in terminal (only once)
				if c.qrGenerator != nil {
					c.qrGenerator.DisplayQRCodeInTerminal(evt.Code, c.sessionID)
				}

				c.logger.InfoWithFields("QR code generated", map[string]interface{}{
					"session_id": c.sessionID,
				})
				c.setStatus("connecting")

			case "success":
				c.logger.InfoWithFields("QR code scanned successfully", map[string]interface{}{
					"session_id": c.sessionID,
				})
				// Clear QR code from client memory
				c.mu.Lock()
				c.qrCode = ""
				c.qrCodeBase64 = ""
				c.mu.Unlock()
				c.setStatus("connected")
				return

			case "timeout":
				c.logger.WarnWithFields("QR code timeout", map[string]interface{}{
					"session_id": c.sessionID,
				})
				c.mu.Lock()
				c.qrCode = ""
				c.qrCodeBase64 = ""
				c.mu.Unlock()
				c.setStatus("disconnected")
				return

			default:
				c.logger.InfoWithFields("QR event", map[string]interface{}{
					"session_id": c.sessionID,
					"event":      evt.Event,
				})
			}
		}
	}
}

// stopQRLoop stops the QR code loop
func (c *WameowClient) stopQRLoop() {
	if c.qrLoopActive {
		c.logger.InfoWithFields("Stopping existing QR loop", map[string]interface{}{
			"session_id": c.sessionID,
		})
		select {
		case c.qrStopChannel <- true:
			c.logger.InfoWithFields("QR loop stop signal sent", map[string]interface{}{
				"session_id": c.sessionID,
			})
		default:
			c.logger.InfoWithFields("QR loop stop channel full, loop may already be stopping", map[string]interface{}{
				"session_id": c.sessionID,
			})
		}
		// Wait a bit for the loop to stop
		time.Sleep(100 * time.Millisecond)
	}
}

// Logout logs out the session
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

// SendTextMessage sends a text message
func (c *WameowClient) SendTextMessage(ctx context.Context, to, body string) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	message := &waE2E.Message{
		Conversation: &body,
	}

	c.logger.InfoWithFields("Sending text message", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"body_len":   len(body),
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send text message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Text message sent successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": resp.ID,
	})

	return &resp, nil
}

// SendImageMessage sends an image message
func (c *WameowClient) SendImageMessage(ctx context.Context, to, filePath, caption string) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read image file: %w", err)
	}

	// Upload media
	uploaded, err := c.client.Upload(ctx, data, whatsmeow.MediaImage)
	if err != nil {
		return nil, fmt.Errorf("failed to upload image: %w", err)
	}

	// Create image message
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
		},
	}

	c.logger.InfoWithFields("Sending image message", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"file_size":  len(data),
		"caption":    caption,
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

// SendAudioMessage sends an audio message
func (c *WameowClient) SendAudioMessage(ctx context.Context, to, filePath string) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio file: %w", err)
	}

	// Upload media
	uploaded, err := c.client.Upload(ctx, data, whatsmeow.MediaAudio)
	if err != nil {
		return nil, fmt.Errorf("failed to upload audio: %w", err)
	}

	// Create audio message
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
		},
	}

	c.logger.InfoWithFields("Sending audio message", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"file_size":  len(data),
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

// SendVideoMessage sends a video message
func (c *WameowClient) SendVideoMessage(ctx context.Context, to, filePath, caption string) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read video file: %w", err)
	}

	// Upload media
	uploaded, err := c.client.Upload(ctx, data, whatsmeow.MediaVideo)
	if err != nil {
		return nil, fmt.Errorf("failed to upload video: %w", err)
	}

	// Create video message
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
		},
	}

	c.logger.InfoWithFields("Sending video message", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"file_size":  len(data),
		"caption":    caption,
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

// SendDocumentMessage sends a document message
func (c *WameowClient) SendDocumentMessage(ctx context.Context, to, filePath, filename string) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read document file: %w", err)
	}

	// Upload media
	uploaded, err := c.client.Upload(ctx, data, whatsmeow.MediaDocument)
	if err != nil {
		return nil, fmt.Errorf("failed to upload document: %w", err)
	}

	// Create document message
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
		},
	}

	c.logger.InfoWithFields("Sending document message", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"file_size":  len(data),
		"filename":   filename,
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

// SendLocationMessage sends a location message
func (c *WameowClient) SendLocationMessage(ctx context.Context, to string, latitude, longitude float64, address string) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	// Create location message
	message := &waE2E.Message{
		LocationMessage: &waE2E.LocationMessage{
			DegreesLatitude:  &latitude,
			DegreesLongitude: &longitude,
			Name:             &address,
		},
	}

	c.logger.InfoWithFields("Sending location message", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"latitude":   latitude,
		"longitude":  longitude,
		"address":    address,
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send location message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Location message sent successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": resp.ID,
	})

	return &resp, nil
}

// SendContactMessage sends a contact message
func (c *WameowClient) SendContactMessage(ctx context.Context, to, contactName, contactPhone string) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	// Create vCard
	vcard := fmt.Sprintf("BEGIN:VCARD\nVERSION:3.0\nFN:%s\nTEL:%s\nEND:VCARD", contactName, contactPhone)

	// Create contact message
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

// ContactInfo represents detailed contact information for vCard
type ContactInfo struct {
	Name         string
	Phone        string
	Email        string
	Organization string
	Title        string
	Website      string
	Address      string
}

// SendDetailedContactMessage sends a contact message with detailed information
func (c *WameowClient) SendDetailedContactMessage(ctx context.Context, to string, contact ContactInfo) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	// Create detailed vCard
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

	// Create contact message
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

// SendContactListMessage sends multiple contacts using ContactsArrayMessage (like WhatsApp native)
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

	// Create display name for the contact list
	displayName := fmt.Sprintf("%d contatos", len(contacts))
	if len(contacts) == 1 {
		displayName = contacts[0].Name
	}

	// Create individual contact messages for ContactsArrayMessage
	var contactMessages []*waE2E.ContactMessage

	for _, contact := range contacts {
		// Create optimized vCard for WhatsApp (only supported fields)
		vcard := "BEGIN:VCARD\n"
		vcard += "VERSION:3.0\n"
		vcard += fmt.Sprintf("FN:%s\n", contact.Name)
		vcard += fmt.Sprintf("N:%s;;;;\n", contact.Name)
		vcard += fmt.Sprintf("TEL:%s\n", contact.Phone)

		// Only include fields that WhatsApp actually supports and displays
		if contact.Organization != "" {
			vcard += fmt.Sprintf("ORG:%s\n", contact.Organization)
		}

		// Note: EMAIL, TITLE, URL, ADR are not displayed by WhatsApp
		// but we'll include them as optional for compatibility with other apps
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

		// Log the generated vCard for analysis
		c.logger.InfoWithFields("📋 Generated vCard for contact", map[string]interface{}{
			"session_id":    c.sessionID,
			"contact_name":  contact.Name,
			"vcard_content": vcard,
		})

		// Create individual contact message
		contactMessage := &waE2E.ContactMessage{
			DisplayName: &contact.Name,
			Vcard:       &vcard,
		}

		contactMessages = append(contactMessages, contactMessage)
	}

	// Create ContactsArrayMessage with all contacts
	message := &waE2E.Message{
		ContactsArrayMessage: &waE2E.ContactsArrayMessage{
			DisplayName: &displayName,
			Contacts:    contactMessages,
		},
	}

	c.logger.InfoWithFields("Sending contacts array message (WhatsApp native format)", map[string]interface{}{
		"session_id":     c.sessionID,
		"to":             to,
		"contact_count":  len(contacts),
		"display_name":   displayName,
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
		"session_id": c.sessionID,
		"to":         to,
		"message_id": resp.ID,
		"contact_count": len(contacts),
	})

	return &resp, nil
}

// SendContactListMessageBusiness sends multiple contacts using Business format
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

	// Create display name for the contact list
	displayName := fmt.Sprintf("%d contatos", len(contacts))
	if len(contacts) == 1 {
		displayName = contacts[0].Name
	}

	// Create individual contact messages for ContactsArrayMessage using WhatsApp Business format
	var contactMessages []*waE2E.ContactMessage

	for _, contact := range contacts {
		// Create WhatsApp Business style vCard
		vcard := "BEGIN:VCARD\n"
		vcard += "VERSION:3.0\n"
		vcard += fmt.Sprintf("N:;%s;;;\n", contact.Name)
		vcard += fmt.Sprintf("FN:%s\n", contact.Name)

		if contact.Organization != "" {
			vcard += fmt.Sprintf("ORG:%s\n", contact.Organization)
		}

		// WhatsApp Business style TITLE (empty if not provided)
		if contact.Title != "" {
			vcard += fmt.Sprintf("TITLE:%s\n", contact.Title)
		} else {
			vcard += "TITLE:\n"
		}

		// WhatsApp Business style phone with waid
		phoneClean := strings.ReplaceAll(strings.ReplaceAll(contact.Phone, "+", ""), " ", "")
		phoneFormatted := contact.Phone
		vcard += fmt.Sprintf("item1.TEL;waid=%s:%s\n", phoneClean, phoneFormatted)
		vcard += "item1.X-ABLabel:Celular\n"

		// WhatsApp Business name field
		vcard += fmt.Sprintf("X-WA-BIZ-NAME:%s\n", contact.Name)

		vcard += "END:VCARD"

		// Log the generated Business style vCard
		c.logger.InfoWithFields("📋 Generated Business style vCard", map[string]interface{}{
			"session_id":    c.sessionID,
			"contact_name":  contact.Name,
			"vcard_content": vcard,
		})

		// Create individual contact message
		contactMessage := &waE2E.ContactMessage{
			DisplayName: &contact.Name,
			Vcard:       &vcard,
		}

		contactMessages = append(contactMessages, contactMessage)
	}

	// Create ContactsArrayMessage with all contacts
	message := &waE2E.Message{
		ContactsArrayMessage: &waE2E.ContactsArrayMessage{
			DisplayName: &displayName,
			Contacts:    contactMessages,
		},
	}

	c.logger.InfoWithFields("Sending contacts array message (Business format)", map[string]interface{}{
		"session_id":     c.sessionID,
		"to":             to,
		"contact_count":  len(contacts),
		"display_name":   displayName,
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
		"session_id": c.sessionID,
		"to":         to,
		"message_id": resp.ID,
		"contact_count": len(contacts),
	})

	return &resp, nil
}

// SendSingleContactMessage sends a single contact using ContactMessage (standard format)
func (c *WameowClient) SendSingleContactMessage(ctx context.Context, to string, contact ContactInfo) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	// Create standard vCard
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

	// Create single ContactMessage
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

// SendSingleContactMessageBusiness sends a single contact using Business format
func (c *WameowClient) SendSingleContactMessageBusiness(ctx context.Context, to string, contact ContactInfo) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	// Create Business style vCard
	vcard := "BEGIN:VCARD\n"
	vcard += "VERSION:3.0\n"
	vcard += fmt.Sprintf("N:;%s;;;\n", contact.Name)
	vcard += fmt.Sprintf("FN:%s\n", contact.Name)

	if contact.Organization != "" {
		vcard += fmt.Sprintf("ORG:%s\n", contact.Organization)
	}

	// Business style TITLE (empty if not provided)
	if contact.Title != "" {
		vcard += fmt.Sprintf("TITLE:%s\n", contact.Title)
	} else {
		vcard += "TITLE:\n"
	}

	// Business style phone with waid
	phoneClean := strings.ReplaceAll(strings.ReplaceAll(contact.Phone, "+", ""), " ", "")
	phoneFormatted := contact.Phone
	vcard += fmt.Sprintf("item1.TEL;waid=%s:%s\n", phoneClean, phoneFormatted)
	vcard += "item1.X-ABLabel:Celular\n"

	// Business name field
	vcard += fmt.Sprintf("X-WA-BIZ-NAME:%s\n", contact.Name)

	vcard += "END:VCARD"

	// Create single ContactMessage with Business format
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

// parseJID parses a JID string into a types.JID
func (c *WameowClient) parseJID(jidStr string) (waTypes.JID, error) {
	if jidStr == "" {
		return waTypes.EmptyJID, fmt.Errorf("JID cannot be empty")
	}

	// If it doesn't contain @, assume it's a phone number and add @s.whatsapp.net
	if !strings.Contains(jidStr, "@") {
		jidStr = jidStr + "@s.whatsapp.net"
	}

	jid, err := waTypes.ParseJID(jidStr)
	if err != nil {
		return waTypes.EmptyJID, fmt.Errorf("failed to parse JID: %w", err)
	}

	return jid, nil
}

// AddEventHandler adds an event handler
func (c *WameowClient) AddEventHandler(handler whatsmeow.EventHandler) uint32 {
	return c.client.AddEventHandler(handler)
}

// SendStickerMessage sends a sticker message
func (c *WameowClient) SendStickerMessage(ctx context.Context, to, filePath string) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read sticker file: %w", err)
	}

	// Upload media
	uploaded, err := c.client.Upload(ctx, data, whatsmeow.MediaImage) // Stickers use image media type
	if err != nil {
		return nil, fmt.Errorf("failed to upload sticker: %w", err)
	}

	// Create sticker message
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

// SendButtonMessage sends a message with interactive buttons (fallback to text with options)
func (c *WameowClient) SendButtonMessage(ctx context.Context, to, body string, buttons []map[string]string) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	// Since interactive buttons may not be fully supported, create a text message with options
	buttonText := body + "\n\n📋 *Options:*"
	for i, button := range buttons {
		if i >= 3 { // Limit to 3 options for readability
			break
		}
		buttonText += fmt.Sprintf("\n%d. %s", i+1, button["text"])
	}

	// Create text message
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

// SendListMessage sends a message with interactive list (fallback to text with sections)
func (c *WameowClient) SendListMessage(ctx context.Context, to, body, buttonText string, sections []map[string]interface{}) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	// Since interactive lists may not be fully supported, create a text message with sections
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

	// Create text message
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

// SendReaction sends a reaction to a message
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

	// Build reaction message using WhatsMeow's BuildReaction
	// Parameters: chat JID, sender JID, message ID, reaction emoji
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

// SendPresence sends presence information (typing, online, etc.)
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

	// Use the available presence methods in WhatsMeow
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

// EditMessage edits an existing message
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

	// Create edit message
	editMessage := &waE2E.Message{
		EditedMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{
				Conversation: &newText,
			},
		},
	}

	// Note: Message editing in WhatsApp is complex and may require the original message key
	// This is a simplified implementation
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

// DeleteMessage deletes an existing message
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

	// Build revoke message using WhatsMeow's BuildRevoke
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

// IsDeviceRegistered checks if the device is registered (has a store ID)
func IsDeviceRegistered(client *whatsmeow.Client) bool {
	if client == nil || client.Store == nil {
		return false
	}
	return client.Store.ID != nil
}
