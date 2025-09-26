package wameow

import (
	"context"
	"time"

	"zpwoot/platform/logger"

	"go.mau.fi/whatsmeow/types/events"
)

// EventHandler handles Wameow events
type EventHandler struct {
	manager    *Manager
	sessionMgr *SessionManager
	qrGen      *QRCodeGenerator
	logger     *logger.Logger
}

// NewEventHandler creates a new event handler
func NewEventHandler(manager *Manager, sessionMgr *SessionManager, qrGen *QRCodeGenerator, logger *logger.Logger) *EventHandler {
	return &EventHandler{
		manager:    manager,
		sessionMgr: sessionMgr,
		qrGen:      qrGen,
		logger:     logger,
	}
}

// SetupEventHandlers is now defined in manager.go to avoid circular imports

// HandleEvent handles all Wameow events
func (h *EventHandler) HandleEvent(evt interface{}, sessionID string) {
	switch v := evt.(type) {
	case *events.Connected:
		h.handleConnected(v, sessionID)
	case *events.Disconnected:
		h.handleDisconnected(v, sessionID)
	case *events.LoggedOut:
		h.handleLoggedOut(v, sessionID)
	case *events.QR:
		h.handleQR(v, sessionID)
	case *events.PairSuccess:
		h.handlePairSuccess(v, sessionID)
	case *events.PairError:
		h.handlePairError(v, sessionID)
	case *events.Message:
		h.handleMessage(v, sessionID)
	case *events.Receipt:
		h.handleReceipt(v, sessionID)
	case *events.Presence:
		h.handlePresence(v, sessionID)
	case *events.ChatPresence:
		h.handleChatPresence(v, sessionID)
	case *events.HistorySync:
		h.handleHistorySync(v, sessionID)
	// Add more common event types to reduce noise
	case *events.AppState:
		h.handleAppState(v, sessionID)
	case *events.AppStateSyncComplete:
		h.handleAppStateSyncComplete(v, sessionID)
	case *events.KeepAliveTimeout:
		h.handleKeepAliveTimeout(v, sessionID)
	case *events.KeepAliveRestored:
		h.handleKeepAliveRestored(v, sessionID)
	case *events.Contact:
		h.handleContact(v, sessionID)
	case *events.GroupInfo:
		h.handleGroupInfo(v, sessionID)
	case *events.Picture:
		h.handlePicture(v, sessionID)
	case *events.BusinessName:
		h.handleBusinessName(v, sessionID)
	case *events.PushName:
		h.handlePushName(v, sessionID)
	case *events.Archive:
		h.handleArchive(v, sessionID)
	case *events.Pin:
		h.handlePin(v, sessionID)
	case *events.Mute:
		h.handleMute(v, sessionID)
	case *events.Star:
		h.handleStar(v, sessionID)
	case *events.DeleteForMe:
		h.handleDeleteForMe(v, sessionID)
	case *events.MarkChatAsRead:
		h.handleMarkChatAsRead(v, sessionID)
	case *events.UndecryptableMessage:
		h.handleUndecryptableMessage(v, sessionID)
	case *events.OfflineSyncPreview:
		h.handleOfflineSyncPreview(v, sessionID)
	case *events.OfflineSyncCompleted:
		h.handleOfflineSyncCompleted(v, sessionID)
	default:
		// Use DEBUG level instead of INFO to reduce noise for truly unknown events
		h.logger.DebugWithFields("Unhandled event", map[string]interface{}{
			"session_id": sessionID,
			"event_type": getEventType(evt),
		})
	}
}

// handleConnected handles connection events
func (h *EventHandler) handleConnected(evt *events.Connected, sessionID string) {
	h.logger.InfoWithFields("Wameow connected", map[string]interface{}{
		"session_id":   sessionID,
		"event_type":   "Connected",
		"connected_at": time.Now().Unix(),
	})

	// Use evt to avoid unused parameter warning
	_ = evt

	h.sessionMgr.UpdateConnectionStatus(sessionID, true)
}

// handleDisconnected handles disconnection events
func (h *EventHandler) handleDisconnected(evt *events.Disconnected, sessionID string) {
	h.logger.InfoWithFields("Wameow disconnected", map[string]interface{}{
		"session_id":      sessionID,
		"event_type":      "Disconnected",
		"disconnected_at": time.Now().Unix(),
	})

	// Use evt to avoid unused parameter warning
	_ = evt

	h.sessionMgr.UpdateConnectionStatus(sessionID, false)
}

// handleLoggedOut handles logout events
func (h *EventHandler) handleLoggedOut(evt *events.LoggedOut, sessionID string) {
	h.logger.InfoWithFields("Wameow logged out", map[string]interface{}{
		"session_id": sessionID,
		"reason":     evt.Reason,
	})

	h.sessionMgr.UpdateConnectionStatus(sessionID, false)
}

// handleQR handles QR code events
func (h *EventHandler) handleQR(evt *events.QR, sessionID string) {
	h.logger.InfoWithFields("QR code received", map[string]interface{}{
		"session_id":  sessionID,
		"codes_count": len(evt.Codes),
	})

	// Generate QR code image
	qrImage := h.qrGen.GenerateQRCodeImage(evt.Codes[0])

	// Update session with QR code
	h.updateSessionQRCode(sessionID, qrImage)

	// Note: QR code display is handled in client.go to avoid duplication
}

// handlePairSuccess handles successful pairing
func (h *EventHandler) handlePairSuccess(evt *events.PairSuccess, sessionID string) {
	h.logger.InfoWithFields("Pairing successful", map[string]interface{}{
		"session_id": sessionID,
		"device_jid": evt.ID.String(),
	})

	h.sessionMgr.UpdateConnectionStatus(sessionID, true)

	// Update session with device JID
	h.updateSessionDeviceJID(sessionID, evt.ID.String())

	// Clear QR code after successful pairing
	h.clearSessionQRCode(sessionID)
}

// handlePairError handles pairing errors
func (h *EventHandler) handlePairError(evt *events.PairError, sessionID string) {
	h.logger.ErrorWithFields("Pairing failed", map[string]interface{}{
		"session_id": sessionID,
		"error":      evt.Error.Error(),
	})

	h.sessionMgr.UpdateConnectionStatus(sessionID, false)
}

// handleMessage handles incoming messages
func (h *EventHandler) handleMessage(evt *events.Message, sessionID string) {
	h.logger.InfoWithFields("Message received", map[string]interface{}{
		"session_id": sessionID,
		"from":       evt.Info.Sender.String(),
		"message_id": evt.Info.ID,
		"timestamp":  evt.Info.Timestamp,
	})

	// Update last seen
	h.updateSessionLastSeen(sessionID)

	// Here you would typically:
	// 1. Process the message
	// 2. Send to webhooks
	// 3. Forward to Chatwoot if configured
	// 4. Store in database if needed
}

// handleReceipt handles message receipts
func (h *EventHandler) handleReceipt(evt *events.Receipt, sessionID string) {
	h.logger.InfoWithFields("Receipt received", map[string]interface{}{
		"session_id": sessionID,
		"type":       evt.Type,
		"sender":     evt.Sender.String(),
		"timestamp":  evt.Timestamp,
	})
}

// handlePresence handles presence updates
func (h *EventHandler) handlePresence(evt *events.Presence, sessionID string) {
	h.logger.InfoWithFields("Presence update", map[string]interface{}{
		"session_id":  sessionID,
		"from":        evt.From.String(),
		"unavailable": evt.Unavailable,
		"last_seen":   evt.LastSeen,
	})
}

// handleChatPresence handles chat presence updates
func (h *EventHandler) handleChatPresence(evt *events.ChatPresence, sessionID string) {
	h.logger.InfoWithFields("Chat presence update", map[string]interface{}{
		"session_id": sessionID,
		"chat":       evt.Chat.String(),
		"state":      evt.State,
	})
}

// handleHistorySync handles history sync events
func (h *EventHandler) handleHistorySync(evt *events.HistorySync, sessionID string) {
	h.logger.InfoWithFields("History sync", map[string]interface{}{
		"session_id": sessionID,
		"data_size":  len(evt.Data.String()), // Just log the data size for now
	})
}

// handleAppState handles app state events
func (h *EventHandler) handleAppState(evt *events.AppState, sessionID string) {
	h.logger.DebugWithFields("App state update", map[string]interface{}{
		"session_id": sessionID,
	})
	_ = evt // Avoid unused parameter warning
}

// handleAppStateSyncComplete handles app state sync completion
func (h *EventHandler) handleAppStateSyncComplete(evt *events.AppStateSyncComplete, sessionID string) {
	h.logger.DebugWithFields("App state sync complete", map[string]interface{}{
		"session_id": sessionID,
		"name":       evt.Name,
	})
}

// handleKeepAliveTimeout handles keep alive timeout events
func (h *EventHandler) handleKeepAliveTimeout(evt *events.KeepAliveTimeout, sessionID string) {
	h.logger.DebugWithFields("Keep alive timeout", map[string]interface{}{
		"session_id": sessionID,
	})
	_ = evt // Avoid unused parameter warning
}

// handleKeepAliveRestored handles keep alive restored events
func (h *EventHandler) handleKeepAliveRestored(evt *events.KeepAliveRestored, sessionID string) {
	h.logger.DebugWithFields("Keep alive restored", map[string]interface{}{
		"session_id": sessionID,
	})
	_ = evt // Avoid unused parameter warning
}

// handleContact handles contact events
func (h *EventHandler) handleContact(evt *events.Contact, sessionID string) {
	h.logger.DebugWithFields("Contact update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

// handleGroupInfo handles group info events
func (h *EventHandler) handleGroupInfo(evt *events.GroupInfo, sessionID string) {
	h.logger.DebugWithFields("Group info update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

// handlePicture handles picture events
func (h *EventHandler) handlePicture(evt *events.Picture, sessionID string) {
	h.logger.DebugWithFields("Picture update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

// handleBusinessName handles business name events
func (h *EventHandler) handleBusinessName(evt *events.BusinessName, sessionID string) {
	h.logger.DebugWithFields("Business name update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

// handlePushName handles push name events
func (h *EventHandler) handlePushName(evt *events.PushName, sessionID string) {
	h.logger.DebugWithFields("Push name update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

// handleArchive handles archive events
func (h *EventHandler) handleArchive(evt *events.Archive, sessionID string) {
	h.logger.DebugWithFields("Archive update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

// handlePin handles pin events
func (h *EventHandler) handlePin(evt *events.Pin, sessionID string) {
	h.logger.DebugWithFields("Pin update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

// handleMute handles mute events
func (h *EventHandler) handleMute(evt *events.Mute, sessionID string) {
	h.logger.DebugWithFields("Mute update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

// handleStar handles star events
func (h *EventHandler) handleStar(evt *events.Star, sessionID string) {
	h.logger.DebugWithFields("Star update", map[string]interface{}{
		"session_id": sessionID,
	})
	_ = evt // Avoid unused parameter warning
}

// handleDeleteForMe handles delete for me events
func (h *EventHandler) handleDeleteForMe(evt *events.DeleteForMe, sessionID string) {
	h.logger.DebugWithFields("Delete for me", map[string]interface{}{
		"session_id": sessionID,
		"chat":       evt.ChatJID.String(),
	})
}

// handleMarkChatAsRead handles mark chat as read events
func (h *EventHandler) handleMarkChatAsRead(evt *events.MarkChatAsRead, sessionID string) {
	h.logger.DebugWithFields("Mark chat as read", map[string]interface{}{
		"session_id": sessionID,
		"chat":       evt.JID.String(),
	})
}

// handleUndecryptableMessage handles undecryptable message events
func (h *EventHandler) handleUndecryptableMessage(evt *events.UndecryptableMessage, sessionID string) {
	h.logger.DebugWithFields("Undecryptable message", map[string]interface{}{
		"session_id": sessionID,
		"from":       evt.Info.Sender.String(),
	})
}

// handleOfflineSyncPreview handles offline sync preview events
func (h *EventHandler) handleOfflineSyncPreview(evt *events.OfflineSyncPreview, sessionID string) {
	h.logger.DebugWithFields("Offline sync preview", map[string]interface{}{
		"session_id": sessionID,
		"messages":   evt.Messages,
	})
}

// handleOfflineSyncCompleted handles offline sync completed events
func (h *EventHandler) handleOfflineSyncCompleted(evt *events.OfflineSyncCompleted, sessionID string) {
	h.logger.DebugWithFields("Offline sync completed", map[string]interface{}{
		"session_id": sessionID,
		"count":      evt.Count,
	})
}

// updateSessionQRCode updates the QR code for a session
func (h *EventHandler) updateSessionQRCode(sessionID, qrCode string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sess, err := h.sessionMgr.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get session for QR update", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return
	}

	// Update QR code
	sess.QRCode = qrCode
	sess.UpdatedAt = time.Now()

	if err := h.sessionMgr.sessionRepo.Update(ctx, sess); err != nil {
		h.logger.ErrorWithFields("Failed to update session QR code", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}
}

// updateSessionDeviceJID updates the device JID for a session
func (h *EventHandler) updateSessionDeviceJID(sessionID, deviceJID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sess, err := h.sessionMgr.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get session for device JID update", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return
	}

	sess.DeviceJid = deviceJID
	sess.UpdatedAt = time.Now()

	if err := h.sessionMgr.sessionRepo.Update(ctx, sess); err != nil {
		h.logger.ErrorWithFields("Failed to update session device JID", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}
}

// updateSessionLastSeen updates the last seen timestamp for a session
func (h *EventHandler) updateSessionLastSeen(sessionID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sess, err := h.sessionMgr.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get session for last seen update", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return
	}

	now := time.Now()
	sess.LastSeen = &now
	sess.UpdatedAt = now

	if err := h.sessionMgr.sessionRepo.Update(ctx, sess); err != nil {
		h.logger.ErrorWithFields("Failed to update session last seen", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}
}

// clearSessionQRCode clears the QR code for a session after successful connection
func (h *EventHandler) clearSessionQRCode(sessionID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sess, err := h.sessionMgr.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get session for QR code clearing", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return
	}

	// Clear QR code and expiration
	sess.QRCode = ""
	sess.QRCodeExpiresAt = nil
	sess.UpdatedAt = time.Now()

	if err := h.sessionMgr.sessionRepo.Update(ctx, sess); err != nil {
		h.logger.ErrorWithFields("Failed to clear session QR code", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	} else {
		h.logger.InfoWithFields("QR code cleared after successful connection", map[string]interface{}{
			"session_id": sessionID,
		})
	}
}

// getEventType returns the type name of an event
func getEventType(evt interface{}) string {
	switch evt.(type) {
	case *events.Connected:
		return "Connected"
	case *events.Disconnected:
		return "Disconnected"
	case *events.LoggedOut:
		return "LoggedOut"
	case *events.QR:
		return "QR"
	case *events.PairSuccess:
		return "PairSuccess"
	case *events.PairError:
		return "PairError"
	case *events.Message:
		return "Message"
	case *events.Receipt:
		return "Receipt"
	case *events.Presence:
		return "Presence"
	case *events.ChatPresence:
		return "ChatPresence"
	case *events.HistorySync:
		return "HistorySync"
	case *events.AppState:
		return "AppState"
	case *events.AppStateSyncComplete:
		return "AppStateSyncComplete"
	case *events.KeepAliveTimeout:
		return "KeepAliveTimeout"
	case *events.KeepAliveRestored:
		return "KeepAliveRestored"
	case *events.Contact:
		return "Contact"
	case *events.GroupInfo:
		return "GroupInfo"
	case *events.Picture:
		return "Picture"
	case *events.BusinessName:
		return "BusinessName"
	case *events.PushName:
		return "PushName"
	case *events.Archive:
		return "Archive"
	case *events.Pin:
		return "Pin"
	case *events.Mute:
		return "Mute"
	case *events.Star:
		return "Star"
	case *events.DeleteForMe:
		return "DeleteForMe"
	case *events.MarkChatAsRead:
		return "MarkChatAsRead"
	case *events.UndecryptableMessage:
		return "UndecryptableMessage"
	case *events.OfflineSyncPreview:
		return "OfflineSyncPreview"
	case *events.OfflineSyncCompleted:
		return "OfflineSyncCompleted"
	default:
		return "Unknown"
	}
}
