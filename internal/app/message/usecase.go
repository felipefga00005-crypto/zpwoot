package message

import (
	"context"
	"fmt"

	"zpwoot/internal/domain/message"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

type UseCase interface {
	SendMessage(ctx context.Context, sessionID string, req *SendMessageRequest) (*SendMessageResponse, error)
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
		"to":         req.To,
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
		processedMedia, err := uc.mediaProcessor.ProcessMedia(ctx, domainReq.File)
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

	response := &SendMessageResponse{
		ID:        result.MessageID,
		Status:    result.Status,
		Timestamp: result.Timestamp,
	}

	return response, nil
}
