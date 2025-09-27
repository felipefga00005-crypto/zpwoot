package wameow

import (
	"context"
	"fmt"
	"time"

	"zpwoot/platform/logger"

	"go.mau.fi/whatsmeow/types/events"
)

type EventHandler struct {
	manager    *Manager
	sessionMgr *SessionManager
	qrGen      *QRCodeGenerator
	logger     *logger.Logger
}

func NewEventHandler(manager *Manager, sessionMgr *SessionManager, qrGen *QRCodeGenerator, logger *logger.Logger) *EventHandler {
	return &EventHandler{
		manager:    manager,
		sessionMgr: sessionMgr,
		qrGen:      qrGen,
		logger:     logger,
	}
}


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
		h.logger.DebugWithFields("Unhandled event", map[string]interface{}{
			"session_id": sessionID,
			"event_type": getEventType(evt),
		})
	}
}

func (h *EventHandler) handleConnected(evt *events.Connected, sessionID string) {
	h.logger.InfoWithFields("Wameow connected", map[string]interface{}{
		"session_id":   sessionID,
		"event_type":   "Connected",
		"connected_at": time.Now().Unix(),
	})

	_ = evt

	h.sessionMgr.UpdateConnectionStatus(sessionID, true)
}

func (h *EventHandler) handleDisconnected(evt *events.Disconnected, sessionID string) {
	h.logger.InfoWithFields("Wameow disconnected", map[string]interface{}{
		"session_id":      sessionID,
		"event_type":      "Disconnected",
		"disconnected_at": time.Now().Unix(),
	})

	_ = evt

	h.sessionMgr.UpdateConnectionStatus(sessionID, false)
}

func (h *EventHandler) handleLoggedOut(evt *events.LoggedOut, sessionID string) {
	h.logger.InfoWithFields("Wameow logged out", map[string]interface{}{
		"session_id": sessionID,
		"reason":     evt.Reason,
	})

	h.sessionMgr.UpdateConnectionStatus(sessionID, false)
}

func (h *EventHandler) handleQR(evt *events.QR, sessionID string) {
	h.logger.InfoWithFields("QR code received", map[string]interface{}{
		"session_id":  sessionID,
		"codes_count": len(evt.Codes),
	})

	qrImage := h.qrGen.GenerateQRCodeImage(evt.Codes[0])

	h.updateSessionQRCode(sessionID, qrImage)

}

func (h *EventHandler) handlePairSuccess(evt *events.PairSuccess, sessionID string) {
	h.logger.InfoWithFields("Pairing successful", map[string]interface{}{
		"session_id": sessionID,
		"device_jid": evt.ID.String(),
	})

	h.sessionMgr.UpdateConnectionStatus(sessionID, true)

	h.updateSessionDeviceJID(sessionID, evt.ID.String())

	h.clearSessionQRCode(sessionID)
}

func (h *EventHandler) handlePairError(evt *events.PairError, sessionID string) {
	h.logger.ErrorWithFields("Pairing failed", map[string]interface{}{
		"session_id": sessionID,
		"error":      evt.Error.Error(),
	})

	h.sessionMgr.UpdateConnectionStatus(sessionID, false)
}

func (h *EventHandler) handleMessage(evt *events.Message, sessionID string) {
	messageInfo := map[string]interface{}{
		"session_id": sessionID,
		"from":       evt.Info.Sender.String(),
		"message_id": evt.Info.ID,
		"timestamp":  evt.Info.Timestamp,
	}

	h.logger.InfoWithFields("🔍 DETAILED MESSAGE ANALYSIS", map[string]interface{}{
		"session_id":                sessionID,
		"from":                      evt.Info.Sender.String(),
		"message_id":                evt.Info.ID,
		"has_contact_message":       evt.Message.ContactMessage != nil,
		"has_contacts_array_message": evt.Message.ContactsArrayMessage != nil,
		"has_conversation":          evt.Message.GetConversation() != "",
		"has_image_message":         evt.Message.ImageMessage != nil,
		"has_audio_message":         evt.Message.AudioMessage != nil,
		"has_video_message":         evt.Message.VideoMessage != nil,
		"has_document_message":      evt.Message.DocumentMessage != nil,
		"has_sticker_message":       evt.Message.StickerMessage != nil,
		"has_location_message":      evt.Message.LocationMessage != nil,
		"has_extended_text_message": evt.Message.ExtendedTextMessage != nil,
		"has_template_message":      evt.Message.TemplateMessage != nil,
		"has_list_message":          evt.Message.ListMessage != nil,
		"has_buttons_message":       evt.Message.ButtonsMessage != nil,
	})

	if evt.Message.ContactMessage != nil {
		contactMsg := evt.Message.ContactMessage
		messageInfo["message_type"] = "contact"

		if contactMsg.DisplayName != nil {
			messageInfo["contact_display_name"] = *contactMsg.DisplayName
		}

		if contactMsg.Vcard != nil {
			messageInfo["contact_vcard"] = *contactMsg.Vcard
			messageInfo["vcard_length"] = len(*contactMsg.Vcard)
		}

		h.logger.InfoWithFields("📞 CONTACT MESSAGE RECEIVED", messageInfo)

		if contactMsg.Vcard != nil {
			h.logger.InfoWithFields("📋 FULL VCARD CONTENT", map[string]interface{}{
				"session_id": sessionID,
				"from":       evt.Info.Sender.String(),
				"vcard":      *contactMsg.Vcard,
			})
		}
	} else if evt.Message.ContactsArrayMessage != nil {
		contactsMsg := evt.Message.ContactsArrayMessage
		messageInfo["message_type"] = "contacts_array"

		if contactsMsg.DisplayName != nil {
			messageInfo["contacts_display_name"] = *contactsMsg.DisplayName
		}

		if contactsMsg.Contacts != nil {
			messageInfo["contacts_count"] = len(contactsMsg.Contacts)
		}

		h.logger.InfoWithFields("📞📞📞 CONTACTS ARRAY MESSAGE RECEIVED", messageInfo)

		if contactsMsg.Contacts != nil {
			for i, contact := range contactsMsg.Contacts {
				contactInfo := map[string]interface{}{
					"session_id":    sessionID,
					"from":          evt.Info.Sender.String(),
					"contact_index": i,
				}

				if contact.DisplayName != nil {
					contactInfo["contact_display_name"] = *contact.DisplayName
				}

				if contact.Vcard != nil {
					contactInfo["contact_vcard"] = *contact.Vcard
					contactInfo["vcard_length"] = len(*contact.Vcard)
				}

				h.logger.InfoWithFields(fmt.Sprintf("📋 CONTACT %d VCARD CONTENT", i+1), contactInfo)
			}
		}
	} else {
		messageType := "text"
		if evt.Message.ImageMessage != nil {
			messageType = "image"
		} else if evt.Message.AudioMessage != nil {
			messageType = "audio"
		} else if evt.Message.VideoMessage != nil {
			messageType = "video"
		} else if evt.Message.DocumentMessage != nil {
			messageType = "document"
		} else if evt.Message.StickerMessage != nil {
			messageType = "sticker"
		} else if evt.Message.LocationMessage != nil {
			messageType = "location"
		}

		messageInfo["message_type"] = messageType

		if evt.Message.GetConversation() != "" {
			messageInfo["text_content"] = evt.Message.GetConversation()
		}

		h.logger.InfoWithFields("Message received", messageInfo)
	}

	h.updateSessionLastSeen(sessionID)

}

func (h *EventHandler) handleReceipt(evt *events.Receipt, sessionID string) {
	h.logger.InfoWithFields("Receipt received", map[string]interface{}{
		"session_id": sessionID,
		"type":       evt.Type,
		"sender":     evt.Sender.String(),
		"timestamp":  evt.Timestamp,
	})
}

func (h *EventHandler) handlePresence(evt *events.Presence, sessionID string) {
	h.logger.InfoWithFields("Presence update", map[string]interface{}{
		"session_id":  sessionID,
		"from":        evt.From.String(),
		"unavailable": evt.Unavailable,
		"last_seen":   evt.LastSeen,
	})
}

func (h *EventHandler) handleChatPresence(evt *events.ChatPresence, sessionID string) {
	h.logger.InfoWithFields("Chat presence update", map[string]interface{}{
		"session_id": sessionID,
		"chat":       evt.Chat.String(),
		"state":      evt.State,
	})
}

func (h *EventHandler) handleHistorySync(evt *events.HistorySync, sessionID string) {
	h.logger.InfoWithFields("History sync", map[string]interface{}{
		"session_id": sessionID,
		"data_size":  len(evt.Data.String()), // Just log the data size for now
	})
}

func (h *EventHandler) handleAppState(evt *events.AppState, sessionID string) {
	h.logger.DebugWithFields("App state update", map[string]interface{}{
		"session_id": sessionID,
	})
	_ = evt // Avoid unused parameter warning
}

func (h *EventHandler) handleAppStateSyncComplete(evt *events.AppStateSyncComplete, sessionID string) {
	h.logger.DebugWithFields("App state sync complete", map[string]interface{}{
		"session_id": sessionID,
		"name":       evt.Name,
	})
}

func (h *EventHandler) handleKeepAliveTimeout(evt *events.KeepAliveTimeout, sessionID string) {
	h.logger.DebugWithFields("Keep alive timeout", map[string]interface{}{
		"session_id": sessionID,
	})
	_ = evt // Avoid unused parameter warning
}

func (h *EventHandler) handleKeepAliveRestored(evt *events.KeepAliveRestored, sessionID string) {
	h.logger.DebugWithFields("Keep alive restored", map[string]interface{}{
		"session_id": sessionID,
	})
	_ = evt // Avoid unused parameter warning
}

func (h *EventHandler) handleContact(evt *events.Contact, sessionID string) {
	h.logger.DebugWithFields("Contact update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

func (h *EventHandler) handleGroupInfo(evt *events.GroupInfo, sessionID string) {
	h.logger.DebugWithFields("Group info update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

func (h *EventHandler) handlePicture(evt *events.Picture, sessionID string) {
	h.logger.DebugWithFields("Picture update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

func (h *EventHandler) handleBusinessName(evt *events.BusinessName, sessionID string) {
	h.logger.DebugWithFields("Business name update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

func (h *EventHandler) handlePushName(evt *events.PushName, sessionID string) {
	h.logger.DebugWithFields("Push name update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

func (h *EventHandler) handleArchive(evt *events.Archive, sessionID string) {
	h.logger.DebugWithFields("Archive update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

func (h *EventHandler) handlePin(evt *events.Pin, sessionID string) {
	h.logger.DebugWithFields("Pin update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

func (h *EventHandler) handleMute(evt *events.Mute, sessionID string) {
	h.logger.DebugWithFields("Mute update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

func (h *EventHandler) handleStar(evt *events.Star, sessionID string) {
	h.logger.DebugWithFields("Star update", map[string]interface{}{
		"session_id": sessionID,
	})
	_ = evt // Avoid unused parameter warning
}

func (h *EventHandler) handleDeleteForMe(evt *events.DeleteForMe, sessionID string) {
	h.logger.DebugWithFields("Delete for me", map[string]interface{}{
		"session_id": sessionID,
		"chat":       evt.ChatJID.String(),
	})
}

func (h *EventHandler) handleMarkChatAsRead(evt *events.MarkChatAsRead, sessionID string) {
	h.logger.DebugWithFields("Mark chat as read", map[string]interface{}{
		"session_id": sessionID,
		"chat":       evt.JID.String(),
	})
}

func (h *EventHandler) handleUndecryptableMessage(evt *events.UndecryptableMessage, sessionID string) {
	h.logger.DebugWithFields("Undecryptable message", map[string]interface{}{
		"session_id": sessionID,
		"from":       evt.Info.Sender.String(),
	})
}

func (h *EventHandler) handleOfflineSyncPreview(evt *events.OfflineSyncPreview, sessionID string) {
	h.logger.DebugWithFields("Offline sync preview", map[string]interface{}{
		"session_id": sessionID,
		"messages":   evt.Messages,
	})
}

func (h *EventHandler) handleOfflineSyncCompleted(evt *events.OfflineSyncCompleted, sessionID string) {
	h.logger.DebugWithFields("Offline sync completed", map[string]interface{}{
		"session_id": sessionID,
		"count":      evt.Count,
	})
}

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

	sess.QRCode = qrCode
	sess.UpdatedAt = time.Now()

	if err := h.sessionMgr.sessionRepo.Update(ctx, sess); err != nil {
		h.logger.ErrorWithFields("Failed to update session QR code", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}
}

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
