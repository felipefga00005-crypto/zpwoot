package message

import (
	"context"
	"fmt"
	"time"

	"zpwoot/internal/domain/message"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

type UseCase interface {
	SendMessage(ctx context.Context, sessionID string, req *SendMessageRequest) (*SendMessageResponse, error)
	GetPollResults(ctx context.Context, req *GetPollResultsRequest) (*GetPollResultsResponse, error)
	RevokeMessage(ctx context.Context, req *RevokeMessageRequest) (*RevokeMessageResponse, error)
	EditMessage(ctx context.Context, req *EditMessageRequest) (*EditMessageResponse, error)
	MarkAsRead(ctx context.Context, req *MarkAsReadRequest) (*MarkAsReadResponse, error)
}

type useCaseImpl struct {
	sessionRepo    ports.SessionRepository
	wameowManager  ports.WameowManager
	mediaProcessor *message.MediaProcessor
	logger         *logger.Logger
}

func NewUseCase(
	sessionRepo ports.SessionRepository,
	wameowManager ports.WameowManager,
	logger *logger.Logger,
) UseCase {
	return &useCaseImpl{
		sessionRepo:    sessionRepo,
		wameowManager:  wameowManager,
		mediaProcessor: message.NewMediaProcessor(logger),
		logger:         logger,
	}
}

func (uc *useCaseImpl) SendMessage(ctx context.Context, sessionID string, req *SendMessageRequest) (*SendMessageResponse, error) {
	uc.logger.InfoWithFields("Sending message", map[string]interface{}{
		"session_id": sessionID,
		"to":         req.JID,
		"type":       req.Type,
	})

	sess, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if sess == nil {
		return nil, fmt.Errorf("session not found")
	}

	if !sess.IsConnected {
		return nil, fmt.Errorf("session is not connected")
	}

	domainReq := req.ToDomainRequest()

	if err := message.ValidateMessageRequest(domainReq); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var filePath string
	var cleanup func() error

	if domainReq.IsMediaMessage() && domainReq.File != "" {
		processedMedia, err := uc.mediaProcessor.ProcessMediaForType(ctx, domainReq.File, domainReq.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to process media: %w", err)
		}

		filePath = processedMedia.FilePath
		cleanup = processedMedia.Cleanup

		if domainReq.MimeType == "" {
			domainReq.MimeType = processedMedia.MimeType
		}

		if domainReq.Type == message.MessageTypeDocument && domainReq.Filename == "" {
			domainReq.Filename = "document"
		}

		defer func() {
			if cleanup != nil {
				if cleanupErr := cleanup(); cleanupErr != nil {
					uc.logger.WarnWithFields("Failed to cleanup temporary file", map[string]interface{}{
						"file_path": filePath,
						"error":     cleanupErr.Error(),
					})
				}
			}
		}()
	}

	// Convert domain ContextInfo to message ContextInfo
	var msgContextInfo *message.ContextInfo
	if domainReq.ContextInfo != nil {
		msgContextInfo = &message.ContextInfo{
			StanzaID:    domainReq.ContextInfo.StanzaID,
			Participant: domainReq.ContextInfo.Participant,
		}
	}

	result, err := uc.wameowManager.SendMessage(
		sessionID,
		domainReq.To,
		string(domainReq.Type),
		domainReq.Body,
		domainReq.Caption,
		filePath,
		domainReq.Filename,
		domainReq.Latitude,
		domainReq.Longitude,
		domainReq.ContactName,
		domainReq.ContactPhone,
		msgContextInfo,
	)

	if err != nil {
		uc.logger.ErrorWithFields("Failed to send message", map[string]interface{}{
			"session_id": sessionID,
			"to":         req.JID,
			"type":       req.Type,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	uc.logger.InfoWithFields("Message sent successfully", map[string]interface{}{
		"session_id": sessionID,
		"to":         req.JID,
		"type":       req.Type,
		"message_id": result.MessageID,
	})

	response := &SendMessageResponse{
		ID:        result.MessageID,
		Status:    result.Status,
		Timestamp: result.Timestamp,
	}

	return response, nil
}

// GetPollResults retrieves poll results for a specific poll message
func (uc *useCaseImpl) GetPollResults(ctx context.Context, req *GetPollResultsRequest) (*GetPollResultsResponse, error) {
	uc.logger.InfoWithFields("Getting poll results", map[string]interface{}{
		"to":              req.JID,
		"poll_message_id": req.PollMessageID,
	})

	// Note: whatsmeow doesn't have a direct GetPollResults method
	// Poll results are typically collected via events (DecryptPollVote)
	// This is a placeholder implementation that would need to be enhanced
	// with actual poll vote collection from events

	return &GetPollResultsResponse{
		PollMessageID:         req.PollMessageID,
		PollName:              "Poll results not yet implemented",
		Options:               []PollOption{},
		TotalVotes:            0,
		SelectableOptionCount: 1,
		AllowMultipleAnswers:  false,
		JID:                   req.JID,
	}, fmt.Errorf("poll results collection not yet implemented - requires event handling")
}

// RevokeMessage revokes a message using whatsmeow's RevokeMessage method
func (uc *useCaseImpl) RevokeMessage(ctx context.Context, req *RevokeMessageRequest) (*RevokeMessageResponse, error) {
	uc.logger.InfoWithFields("Revoking message", map[string]interface{}{
		"to":         req.JID,
		"message_id": req.MessageID,
	})

	// Use whatsmeow's RevokeMessage method
	result, err := uc.wameowManager.RevokeMessage(req.SessionID, req.JID, req.MessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to revoke message: %w", err)
	}

	return &RevokeMessageResponse{
		ID:        result.MessageID,
		Status:    "revoked",
		Timestamp: result.Timestamp,
	}, nil
}

// EditMessage edits a message using whatsmeow's BuildEdit method
func (uc *useCaseImpl) EditMessage(ctx context.Context, req *EditMessageRequest) (*EditMessageResponse, error) {
	uc.logger.InfoWithFields("Editing message", map[string]interface{}{
		"to":         req.JID,
		"message_id": req.MessageID,
		"new_body":   req.NewBody,
	})

	// Use whatsmeow's BuildEdit method
	err := uc.wameowManager.EditMessage(req.SessionID, req.JID, req.MessageID, req.NewBody)
	if err != nil {
		return nil, fmt.Errorf("failed to edit message: %w", err)
	}

	return &EditMessageResponse{
		ID:        req.MessageID,
		Status:    "edited",
		NewBody:   req.NewBody,
		Timestamp: time.Now(),
	}, nil
}

// MarkAsRead marks messages as read using whatsmeow's MarkRead method
func (uc *useCaseImpl) MarkAsRead(ctx context.Context, req *MarkAsReadRequest) (*MarkAsReadResponse, error) {
	uc.logger.InfoWithFields("Marking messages as read", map[string]interface{}{
		"to":          req.JID,
		"message_ids": req.MessageIDs,
	})

	// Use whatsmeow's MarkRead method (currently supports single message)
	// For multiple messages, we'll mark each one individually
	for _, messageID := range req.MessageIDs {
		err := uc.wameowManager.MarkRead(req.SessionID, req.JID, messageID)
		if err != nil {
			uc.logger.WarnWithFields("Failed to mark message as read", map[string]interface{}{
				"session_id": req.SessionID,
				"to":         req.JID,
				"message_id": messageID,
				"error":      err.Error(),
			})
			// Continue with other messages even if one fails
		}
	}

	return &MarkAsReadResponse{
		MessageIDs: req.MessageIDs,
		Status:     "read",
		Timestamp:  time.Now(),
	}, nil
}
