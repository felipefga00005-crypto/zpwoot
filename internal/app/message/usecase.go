package message

import (
	"context"
	"fmt"

	"zpwoot/internal/domain/message"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

// UseCase defines the message use case interface
type UseCase interface {
	SendMessage(ctx context.Context, sessionID string, req *SendMessageRequest) (*SendMessageResponse, error)
}

// useCaseImpl implements the message use case
type useCaseImpl struct {
	sessionRepo   ports.SessionRepository
	wameowManager ports.WameowManager
	mediaProcessor *message.MediaProcessor
	logger        *logger.Logger
}

// NewUseCase creates a new message use case
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

// SendMessage sends a message through WhatsApp
func (uc *useCaseImpl) SendMessage(ctx context.Context, sessionID string, req *SendMessageRequest) (*SendMessageResponse, error) {
	uc.logger.InfoWithFields("Sending message", map[string]interface{}{
		"session_id": sessionID,
		"to":         req.To,
		"type":       req.Type,
	})

	// Validate session exists and is connected
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

	// Convert to domain request
	domainReq := req.ToDomainRequest()

	// Validate request
	if err := message.ValidateMessageRequest(domainReq); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Process media if needed
	var filePath string
	var cleanup func() error

	if domainReq.IsMediaMessage() && domainReq.File != "" {
		processedMedia, err := uc.mediaProcessor.ProcessMedia(ctx, domainReq.File)
		if err != nil {
			return nil, fmt.Errorf("failed to process media: %w", err)
		}

		filePath = processedMedia.FilePath
		cleanup = processedMedia.Cleanup

		// Set MIME type if not provided
		if domainReq.MimeType == "" {
			domainReq.MimeType = processedMedia.MimeType
		}

		// Set filename if not provided for documents
		if domainReq.Type == message.MessageTypeDocument && domainReq.Filename == "" {
			domainReq.Filename = "document"
		}

		// Ensure cleanup happens
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

	// Send message through WhatsMeow manager
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
	)

	if err != nil {
		uc.logger.ErrorWithFields("Failed to send message", map[string]interface{}{
			"session_id": sessionID,
			"to":         req.To,
			"type":       req.Type,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	uc.logger.InfoWithFields("Message sent successfully", map[string]interface{}{
		"session_id": sessionID,
		"to":         req.To,
		"type":       req.Type,
		"message_id": result.MessageID,
	})

	// Convert result to response
	response := &SendMessageResponse{
		ID:        result.MessageID,
		Status:    result.Status,
		Timestamp: result.Timestamp,
	}

	return response, nil
}
