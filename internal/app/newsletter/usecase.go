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
			"jid":        req.JID,
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
	jid := uc.newsletterService.FormatNewsletterJID(req.JID)

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
			"jid":        req.JID,
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
	jid := uc.newsletterService.FormatNewsletterJID(req.JID)

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
			"jid":        req.JID,
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
	jid := uc.newsletterService.FormatNewsletterJID(req.JID)

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


