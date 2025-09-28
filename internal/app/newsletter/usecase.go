package newsletter

import (
	"context"
	"fmt"

	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

// UseCase defines the interface for newsletter use cases
type UseCase interface {
	// CreateNewsletter creates a new newsletter
	CreateNewsletter(ctx context.Context, sessionID string, req *CreateNewsletterRequest) (*CreateNewsletterResponse, error)
	
	// GetNewsletterInfo gets newsletter information by JID
	GetNewsletterInfo(ctx context.Context, sessionID string, req *GetNewsletterInfoRequest) (*NewsletterInfoResponse, error)
	
	// GetNewsletterInfoWithInvite gets newsletter information using invite key
	GetNewsletterInfoWithInvite(ctx context.Context, sessionID string, req *GetNewsletterInfoWithInviteRequest) (*NewsletterInfoResponse, error)
	
	// FollowNewsletter follows a newsletter
	FollowNewsletter(ctx context.Context, sessionID string, req *FollowNewsletterRequest) (*NewsletterActionResponse, error)
	
	// UnfollowNewsletter unfollows a newsletter
	UnfollowNewsletter(ctx context.Context, sessionID string, req *UnfollowNewsletterRequest) (*NewsletterActionResponse, error)
	
	// GetSubscribedNewsletters gets all subscribed newsletters
	GetSubscribedNewsletters(ctx context.Context, sessionID string) (*SubscribedNewslettersResponse, error)

	// GetNewsletterMessages gets messages from a newsletter
	GetNewsletterMessages(ctx context.Context, sessionID string, req *GetNewsletterMessagesRequest) (*GetNewsletterMessagesResponse, error)

	// GetNewsletterMessageUpdates gets message updates from a newsletter
	GetNewsletterMessageUpdates(ctx context.Context, sessionID string, req *GetNewsletterMessageUpdatesRequest) (*GetNewsletterMessageUpdatesResponse, error)

	// NewsletterMarkViewed marks newsletter messages as viewed
	NewsletterMarkViewed(ctx context.Context, sessionID string, req *NewsletterMarkViewedRequest) (*NewsletterActionResponse, error)

	// NewsletterSendReaction sends a reaction to a newsletter message
	NewsletterSendReaction(ctx context.Context, sessionID string, req *NewsletterSendReactionRequest) (*NewsletterActionResponse, error)

	// NewsletterSubscribeLiveUpdates subscribes to live updates from a newsletter
	NewsletterSubscribeLiveUpdates(ctx context.Context, sessionID string, req *NewsletterSubscribeLiveUpdatesRequest) (*NewsletterSubscribeLiveUpdatesResponse, error)

	// NewsletterToggleMute toggles mute status of a newsletter
	NewsletterToggleMute(ctx context.Context, sessionID string, req *NewsletterToggleMuteRequest) (*NewsletterActionResponse, error)

	// AcceptTOSNotice accepts a terms of service notice
	AcceptTOSNotice(ctx context.Context, sessionID string, req *AcceptTOSNoticeRequest) (*NewsletterActionResponse, error)

	// UploadNewsletter uploads media for newsletters
	UploadNewsletter(ctx context.Context, sessionID string, req *UploadNewsletterRequest) (*UploadNewsletterResponse, error)

	// UploadNewsletterReader uploads media for newsletters from a reader
	UploadNewsletterReader(ctx context.Context, sessionID string, req *UploadNewsletterRequest) (*UploadNewsletterResponse, error)
}

// useCaseImpl implements the UseCase interface
type useCaseImpl struct {
	newsletterManager ports.NewsletterManager
	newsletterService ports.NewsletterService
	sessionRepo       ports.SessionRepository
	logger            logger.Logger
}

// NewUseCase creates a new newsletter use case
func NewUseCase(
	newsletterManager ports.NewsletterManager,
	newsletterService ports.NewsletterService,
	sessionRepo ports.SessionRepository,
	logger logger.Logger,
) UseCase {
	return &useCaseImpl{
		newsletterManager: newsletterManager,
		newsletterService: newsletterService,
		sessionRepo:       sessionRepo,
		logger:            logger,
	}
}

// CreateNewsletter creates a new newsletter
func (uc *useCaseImpl) CreateNewsletter(ctx context.Context, sessionID string, req *CreateNewsletterRequest) (*CreateNewsletterResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid create newsletter request", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Validate session
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if !session.IsConnected {
		return nil, fmt.Errorf("session is not connected")
	}

	// Sanitize input
	name := uc.newsletterService.SanitizeNewsletterName(req.Name)
	description := uc.newsletterService.SanitizeNewsletterDescription(req.Description)

	uc.logger.InfoWithFields("Creating newsletter", map[string]interface{}{
		"session_id":  sessionID,
		"name":        name,
		"description": description,
	})

	// Create newsletter via WhatsApp
	newsletterInfo, err := uc.newsletterManager.CreateNewsletter(ctx, sessionID, name, description)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to create newsletter", map[string]interface{}{
			"session_id": sessionID,
			"name":       name,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to create newsletter: %w", err)
	}

	// Process newsletter info
	if err := uc.newsletterService.ProcessNewsletterInfo(newsletterInfo); err != nil {
		uc.logger.ErrorWithFields("Failed to process newsletter info", map[string]interface{}{
			"session_id":    sessionID,
			"newsletter_id": newsletterInfo.ID,
			"error":         err.Error(),
		})
		return nil, fmt.Errorf("failed to process newsletter info: %w", err)
	}

	uc.logger.InfoWithFields("Newsletter created successfully", map[string]interface{}{
		"session_id":    sessionID,
		"newsletter_id": newsletterInfo.ID,
		"name":          newsletterInfo.Name,
	})

	return NewCreateNewsletterResponse(newsletterInfo), nil
}

// GetNewsletterInfo gets newsletter information by JID
func (uc *useCaseImpl) GetNewsletterInfo(ctx context.Context, sessionID string, req *GetNewsletterInfoRequest) (*NewsletterInfoResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid get newsletter info request", map[string]interface{}{
			"session_id": sessionID,
			"jid":        req.NewsletterJID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Validate session
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if !session.IsConnected {
		return nil, fmt.Errorf("session is not connected")
	}

	// Format JID
	jid := uc.newsletterService.FormatNewsletterJID(req.NewsletterJID)

	uc.logger.InfoWithFields("Getting newsletter info", map[string]interface{}{
		"session_id": sessionID,
		"jid":        jid,
	})

	// Get newsletter info via WhatsApp
	newsletterInfo, err := uc.newsletterManager.GetNewsletterInfo(ctx, sessionID, jid)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to get newsletter info", map[string]interface{}{
			"session_id": sessionID,
			"jid":        jid,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get newsletter info: %w", err)
	}

	// Process newsletter info
	if err := uc.newsletterService.ProcessNewsletterInfo(newsletterInfo); err != nil {
		uc.logger.ErrorWithFields("Failed to process newsletter info", map[string]interface{}{
			"session_id":    sessionID,
			"newsletter_id": newsletterInfo.ID,
			"error":         err.Error(),
		})
		return nil, fmt.Errorf("failed to process newsletter info: %w", err)
	}

	uc.logger.InfoWithFields("Newsletter info retrieved successfully", map[string]interface{}{
		"session_id":    sessionID,
		"newsletter_id": newsletterInfo.ID,
		"name":          newsletterInfo.Name,
	})

	return NewNewsletterInfoResponse(newsletterInfo), nil
}

// GetNewsletterInfoWithInvite gets newsletter information using invite key
func (uc *useCaseImpl) GetNewsletterInfoWithInvite(ctx context.Context, sessionID string, req *GetNewsletterInfoWithInviteRequest) (*NewsletterInfoResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid get newsletter info with invite request", map[string]interface{}{
			"session_id": sessionID,
			"invite_key": req.InviteKey,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Validate session
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if !session.IsConnected {
		return nil, fmt.Errorf("session is not connected")
	}

	// Clean invite key
	inviteKey := uc.newsletterService.CleanInviteKey(req.InviteKey)

	uc.logger.InfoWithFields("Getting newsletter info with invite", map[string]interface{}{
		"session_id": sessionID,
		"invite_key": inviteKey,
	})

	// Get newsletter info via WhatsApp
	newsletterInfo, err := uc.newsletterManager.GetNewsletterInfoWithInvite(ctx, sessionID, inviteKey)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to get newsletter info with invite", map[string]interface{}{
			"session_id": sessionID,
			"invite_key": inviteKey,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get newsletter info with invite: %w", err)
	}

	// Process newsletter info
	if err := uc.newsletterService.ProcessNewsletterInfo(newsletterInfo); err != nil {
		uc.logger.ErrorWithFields("Failed to process newsletter info", map[string]interface{}{
			"session_id":    sessionID,
			"newsletter_id": newsletterInfo.ID,
			"error":         err.Error(),
		})
		return nil, fmt.Errorf("failed to process newsletter info: %w", err)
	}

	uc.logger.InfoWithFields("Newsletter info retrieved with invite successfully", map[string]interface{}{
		"session_id":    sessionID,
		"newsletter_id": newsletterInfo.ID,
		"name":          newsletterInfo.Name,
	})

	return NewNewsletterInfoResponse(newsletterInfo), nil
}

// FollowNewsletter follows a newsletter
func (uc *useCaseImpl) FollowNewsletter(ctx context.Context, sessionID string, req *FollowNewsletterRequest) (*NewsletterActionResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid follow newsletter request", map[string]interface{}{
			"session_id": sessionID,
			"jid":        req.NewsletterJID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Validate session
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if !session.IsConnected {
		return nil, fmt.Errorf("session is not connected")
	}

	// Format JID
	jid := uc.newsletterService.FormatNewsletterJID(req.NewsletterJID)

	uc.logger.InfoWithFields("Following newsletter", map[string]interface{}{
		"session_id": sessionID,
		"jid":        jid,
	})

	// Follow newsletter via WhatsApp
	err = uc.newsletterManager.FollowNewsletter(ctx, sessionID, jid)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to follow newsletter", map[string]interface{}{
			"session_id": sessionID,
			"jid":        jid,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to follow newsletter: %w", err)
	}

	uc.logger.InfoWithFields("Newsletter followed successfully", map[string]interface{}{
		"session_id": sessionID,
		"jid":        jid,
	})

	return NewSuccessFollowResponse(jid), nil
}

// UnfollowNewsletter unfollows a newsletter
func (uc *useCaseImpl) UnfollowNewsletter(ctx context.Context, sessionID string, req *UnfollowNewsletterRequest) (*NewsletterActionResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid unfollow newsletter request", map[string]interface{}{
			"session_id": sessionID,
			"jid":        req.NewsletterJID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Validate session
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if !session.IsConnected {
		return nil, fmt.Errorf("session is not connected")
	}

	// Format JID
	jid := uc.newsletterService.FormatNewsletterJID(req.NewsletterJID)

	uc.logger.InfoWithFields("Unfollowing newsletter", map[string]interface{}{
		"session_id": sessionID,
		"jid":        jid,
	})

	// Unfollow newsletter via WhatsApp
	err = uc.newsletterManager.UnfollowNewsletter(ctx, sessionID, jid)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to unfollow newsletter", map[string]interface{}{
			"session_id": sessionID,
			"jid":        jid,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to unfollow newsletter: %w", err)
	}

	uc.logger.InfoWithFields("Newsletter unfollowed successfully", map[string]interface{}{
		"session_id": sessionID,
		"jid":        jid,
	})

	return NewSuccessUnfollowResponse(jid), nil
}

// GetSubscribedNewsletters gets all subscribed newsletters
func (uc *useCaseImpl) GetSubscribedNewsletters(ctx context.Context, sessionID string) (*SubscribedNewslettersResponse, error) {
	// Validate session
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if !session.IsConnected {
		return nil, fmt.Errorf("session is not connected")
	}

	uc.logger.InfoWithFields("Getting subscribed newsletters", map[string]interface{}{
		"session_id": sessionID,
	})

	// Get subscribed newsletters via WhatsApp
	newsletters, err := uc.newsletterManager.GetSubscribedNewsletters(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to get subscribed newsletters", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get subscribed newsletters: %w", err)
	}

	// Process each newsletter info
	for _, newsletterInfo := range newsletters {
		if err := uc.newsletterService.ProcessNewsletterInfo(newsletterInfo); err != nil {
			uc.logger.WarnWithFields("Failed to process newsletter info", map[string]interface{}{
				"session_id":    sessionID,
				"newsletter_id": newsletterInfo.ID,
				"error":         err.Error(),
			})
			// Continue processing other newsletters
		}
	}

	uc.logger.InfoWithFields("Subscribed newsletters retrieved successfully", map[string]interface{}{
		"session_id": sessionID,
		"count":      len(newsletters),
	})

	return NewSubscribedNewslettersResponse(newsletters), nil
}

// GetNewsletterMessages gets messages from a newsletter
func (uc *useCaseImpl) GetNewsletterMessages(ctx context.Context, sessionID string, req *GetNewsletterMessagesRequest) (*GetNewsletterMessagesResponse, error) {
	uc.logger.InfoWithFields("Getting newsletter messages", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.NewsletterJID,
		"count":      req.Count,
		"before":     req.Before,
	})

	// Validate session
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session == nil {
		uc.logger.ErrorWithFields("Session is nil", map[string]interface{}{
			"session_id": sessionID,
		})
		return nil, fmt.Errorf("session not found")
	}

	// Validate request
	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid request", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Get messages from newsletter manager
	messages, err := uc.newsletterManager.GetNewsletterMessages(ctx, sessionID, req.NewsletterJID, req.Count, req.Before)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to get newsletter messages", map[string]interface{}{
			"session_id": sessionID,
			"jid":        req.NewsletterJID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get newsletter messages: %w", err)
	}

	uc.logger.InfoWithFields("Newsletter messages retrieved successfully", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.NewsletterJID,
		"count":      len(messages),
	})

	return NewGetNewsletterMessagesResponse(messages), nil
}

// GetNewsletterMessageUpdates gets message updates from a newsletter
func (uc *useCaseImpl) GetNewsletterMessageUpdates(ctx context.Context, sessionID string, req *GetNewsletterMessageUpdatesRequest) (*GetNewsletterMessageUpdatesResponse, error) {
	uc.logger.InfoWithFields("Getting newsletter message updates", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.NewsletterJID,
		"count":      req.Count,
		"since":      req.Since,
		"after":      req.After,
	})

	// Validate session
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session == nil {
		uc.logger.ErrorWithFields("Session is nil", map[string]interface{}{
			"session_id": sessionID,
		})
		return nil, fmt.Errorf("session not found")
	}

	// Validate request
	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid request", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Get updates from newsletter manager
	updates, err := uc.newsletterManager.GetNewsletterMessageUpdates(ctx, sessionID, req.NewsletterJID, req.Count, req.Since, req.After)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to get newsletter message updates", map[string]interface{}{
			"session_id": sessionID,
			"jid":        req.NewsletterJID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get newsletter message updates: %w", err)
	}

	uc.logger.InfoWithFields("Newsletter message updates retrieved successfully", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.NewsletterJID,
		"count":      len(updates),
	})

	return NewGetNewsletterMessageUpdatesResponse(updates), nil
}

// NewsletterMarkViewed marks newsletter messages as viewed
func (uc *useCaseImpl) NewsletterMarkViewed(ctx context.Context, sessionID string, req *NewsletterMarkViewedRequest) (*NewsletterActionResponse, error) {
	uc.logger.InfoWithFields("Marking newsletter messages as viewed", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.NewsletterJID,
		"count":      len(req.ServerIDs),
	})

	// Validate session
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session == nil {
		return nil, fmt.Errorf("session not found")
	}

	// Validate request
	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid request", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Mark messages as viewed
	err = uc.newsletterManager.NewsletterMarkViewed(ctx, sessionID, req.NewsletterJID, req.ServerIDs)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to mark newsletter messages as viewed", map[string]interface{}{
			"session_id": sessionID,
			"jid":        req.NewsletterJID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to mark newsletter messages as viewed: %w", err)
	}

	uc.logger.InfoWithFields("Newsletter messages marked as viewed successfully", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.NewsletterJID,
		"count":      len(req.ServerIDs),
	})

	return NewNewsletterActionResponse(req.NewsletterJID, "success", "Messages marked as viewed successfully"), nil
}

// NewsletterSendReaction sends a reaction to a newsletter message
func (uc *useCaseImpl) NewsletterSendReaction(ctx context.Context, sessionID string, req *NewsletterSendReactionRequest) (*NewsletterActionResponse, error) {
	uc.logger.InfoWithFields("Sending newsletter reaction", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.NewsletterJID,
		"server_id":  req.ServerID,
		"reaction":   req.Reaction,
	})

	// Validate session
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session == nil {
		return nil, fmt.Errorf("session not found")
	}

	// Validate request
	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid request", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// If ServerID is not provided, try to find it from the MessageID
	serverID := req.ServerID
	if serverID == "" {
		uc.logger.InfoWithFields("ServerID not provided, looking up from MessageID", map[string]interface{}{
			"session_id": sessionID,
			"jid":        req.NewsletterJID,
			"message_id": req.MessageID,
		})

		// Get recent messages to find the ServerID for this MessageID
		messages, err := uc.newsletterManager.GetNewsletterMessages(ctx, sessionID, req.NewsletterJID, 50, "")
		if err != nil {
			uc.logger.ErrorWithFields("Failed to get newsletter messages for ServerID lookup", map[string]interface{}{
				"session_id": sessionID,
				"jid":        req.NewsletterJID,
				"error":      err.Error(),
			})
			return nil, fmt.Errorf("failed to lookup ServerID: %w", err)
		}

		// Find the message with matching MessageID
		for _, msg := range messages {
			if msg.ID == req.MessageID {
				serverID = msg.ServerID
				uc.logger.InfoWithFields("Found ServerID for MessageID", map[string]interface{}{
					"session_id": sessionID,
					"message_id": req.MessageID,
					"server_id":  serverID,
				})
				break
			}
		}

		if serverID == "" {
			return nil, fmt.Errorf("could not find ServerID for MessageID %s in newsletter %s", req.MessageID, req.NewsletterJID)
		}
	}

	// Send reaction
	err = uc.newsletterManager.NewsletterSendReaction(ctx, sessionID, req.NewsletterJID, serverID, req.Reaction, req.MessageID)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to send newsletter reaction", map[string]interface{}{
			"session_id": sessionID,
			"jid":        req.NewsletterJID,
			"server_id":  serverID,
			"message_id": req.MessageID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to send newsletter reaction: %w", err)
	}

	uc.logger.InfoWithFields("Newsletter reaction sent successfully", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.NewsletterJID,
		"server_id":  serverID,
		"message_id": req.MessageID,
		"reaction":   req.Reaction,
	})

	message := "Reaction sent successfully"
	if req.Reaction == "" {
		message = "Reaction removed successfully"
	}

	return NewNewsletterActionResponse(req.NewsletterJID, "success", message), nil
}

// NewsletterSubscribeLiveUpdates subscribes to live updates from a newsletter
func (uc *useCaseImpl) NewsletterSubscribeLiveUpdates(ctx context.Context, sessionID string, req *NewsletterSubscribeLiveUpdatesRequest) (*NewsletterSubscribeLiveUpdatesResponse, error) {
	uc.logger.InfoWithFields("Subscribing to newsletter live updates", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.NewsletterJID,
	})

	// Validate session
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session == nil {
		return nil, fmt.Errorf("session not found")
	}

	// Validate request
	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid request", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Subscribe to live updates
	duration, err := uc.newsletterManager.NewsletterSubscribeLiveUpdates(ctx, sessionID, req.NewsletterJID)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to subscribe to newsletter live updates", map[string]interface{}{
			"session_id": sessionID,
			"jid":        req.NewsletterJID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to subscribe to newsletter live updates: %w", err)
	}

	uc.logger.InfoWithFields("Subscribed to newsletter live updates successfully", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.NewsletterJID,
		"duration":   duration,
	})

	return &NewsletterSubscribeLiveUpdatesResponse{Duration: duration}, nil
}

// NewsletterToggleMute toggles mute status of a newsletter
func (uc *useCaseImpl) NewsletterToggleMute(ctx context.Context, sessionID string, req *NewsletterToggleMuteRequest) (*NewsletterActionResponse, error) {
	uc.logger.InfoWithFields("Toggling newsletter mute status", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.NewsletterJID,
		"mute":       req.Mute,
	})

	// Validate session
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session == nil {
		return nil, fmt.Errorf("session not found")
	}

	// Validate request
	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid request", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Toggle mute status
	err = uc.newsletterManager.NewsletterToggleMute(ctx, sessionID, req.NewsletterJID, req.Mute)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to toggle newsletter mute status", map[string]interface{}{
			"session_id": sessionID,
			"jid":        req.NewsletterJID,
			"mute":       req.Mute,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to toggle newsletter mute status: %w", err)
	}

	uc.logger.InfoWithFields("Newsletter mute status toggled successfully", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.NewsletterJID,
		"mute":       req.Mute,
	})

	message := "Newsletter muted successfully"
	if !req.Mute {
		message = "Newsletter unmuted successfully"
	}

	return NewNewsletterActionResponse(req.NewsletterJID, "success", message), nil
}

// AcceptTOSNotice accepts a terms of service notice
func (uc *useCaseImpl) AcceptTOSNotice(ctx context.Context, sessionID string, req *AcceptTOSNoticeRequest) (*NewsletterActionResponse, error) {
	uc.logger.InfoWithFields("Accepting TOS notice", map[string]interface{}{
		"session_id": sessionID,
		"notice_id":  req.NoticeID,
		"stage":      req.Stage,
	})

	// Validate session
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session == nil {
		return nil, fmt.Errorf("session not found")
	}

	// Validate request
	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid request", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Accept TOS notice
	err = uc.newsletterManager.AcceptTOSNotice(ctx, sessionID, req.NoticeID, req.Stage)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to accept TOS notice", map[string]interface{}{
			"session_id": sessionID,
			"notice_id":  req.NoticeID,
			"stage":      req.Stage,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to accept TOS notice: %w", err)
	}

	uc.logger.InfoWithFields("TOS notice accepted successfully", map[string]interface{}{
		"session_id": sessionID,
		"notice_id":  req.NoticeID,
		"stage":      req.Stage,
	})

	return NewNewsletterActionResponse("", "success", "TOS notice accepted successfully"), nil
}

// UploadNewsletter uploads media for newsletters
func (uc *useCaseImpl) UploadNewsletter(ctx context.Context, sessionID string, req *UploadNewsletterRequest) (*UploadNewsletterResponse, error) {
	uc.logger.InfoWithFields("Uploading newsletter media", map[string]interface{}{
		"session_id": sessionID,
		"mime_type":  req.MimeType,
		"media_type": req.MediaType,
		"data_size":  len(req.Data),
	})

	// Validate session
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session == nil {
		return nil, fmt.Errorf("session not found")
	}

	// Validate request
	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid request", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Upload media
	response, err := uc.newsletterManager.UploadNewsletter(ctx, sessionID, req.Data, req.MimeType, req.MediaType)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to upload newsletter media", map[string]interface{}{
			"session_id": sessionID,
			"mime_type":  req.MimeType,
			"media_type": req.MediaType,
			"data_size":  len(req.Data),
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to upload newsletter media: %w", err)
	}

	uc.logger.InfoWithFields("Newsletter media uploaded successfully", map[string]interface{}{
		"session_id":  sessionID,
		"mime_type":   req.MimeType,
		"media_type":  req.MediaType,
		"data_size":   len(req.Data),
		"url":         response.URL,
		"handle":      response.Handle,
		"file_length": response.FileLength,
	})

	// Convert domain response to DTO response
	return &UploadNewsletterResponse{
		URL:        response.URL,
		DirectPath: response.DirectPath,
		Handle:     response.Handle,
		ObjectID:   response.ObjectID,
		FileSHA256: response.FileSHA256,
		FileLength: response.FileLength,
	}, nil
}

// UploadNewsletterReader uploads media for newsletters from a reader
func (uc *useCaseImpl) UploadNewsletterReader(ctx context.Context, sessionID string, req *UploadNewsletterRequest) (*UploadNewsletterResponse, error) {
	uc.logger.InfoWithFields("Uploading newsletter media with reader", map[string]interface{}{
		"session_id": sessionID,
		"mime_type":  req.MimeType,
		"media_type": req.MediaType,
		"data_size":  len(req.Data),
	})

	// Validate session
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session == nil {
		return nil, fmt.Errorf("session not found")
	}

	// Validate request
	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid request", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Upload media with reader
	response, err := uc.newsletterManager.UploadNewsletterReader(ctx, sessionID, req.Data, req.MimeType, req.MediaType)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to upload newsletter media with reader", map[string]interface{}{
			"session_id": sessionID,
			"mime_type":  req.MimeType,
			"media_type": req.MediaType,
			"data_size":  len(req.Data),
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to upload newsletter media with reader: %w", err)
	}

	uc.logger.InfoWithFields("Newsletter media uploaded successfully with reader", map[string]interface{}{
		"session_id":  sessionID,
		"mime_type":   req.MimeType,
		"media_type":  req.MediaType,
		"data_size":   len(req.Data),
		"url":         response.URL,
		"handle":      response.Handle,
		"file_length": response.FileLength,
	})

	// Convert domain response to DTO response
	return &UploadNewsletterResponse{
		URL:        response.URL,
		DirectPath: response.DirectPath,
		Handle:     response.Handle,
		ObjectID:   response.ObjectID,
		FileSHA256: response.FileSHA256,
		FileLength: response.FileLength,
	}, nil
}


