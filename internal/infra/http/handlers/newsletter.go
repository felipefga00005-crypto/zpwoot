package handlers

import (
	"zpwoot/internal/app/newsletter"
	domainSession "zpwoot/internal/domain/session"
	"zpwoot/internal/infra/http/helpers"
	"zpwoot/platform/logger"

	"github.com/gofiber/fiber/v2"
)

// NewsletterHandler handles newsletter-related HTTP requests
type NewsletterHandler struct {
	logger          *logger.Logger
	newsletterUC    newsletter.UseCase
	sessionResolver *helpers.SessionResolver
}

// NewNewsletterHandler creates a new newsletter handler
func NewNewsletterHandler(appLogger *logger.Logger, newsletterUC newsletter.UseCase, sessionRepo helpers.SessionRepository) *NewsletterHandler {
	return &NewsletterHandler{
		logger:          appLogger,
		newsletterUC:    newsletterUC,
		sessionResolver: helpers.NewSessionResolver(appLogger, sessionRepo),
	}
}

// resolveSession resolves session from URL parameter
func (h *NewsletterHandler) resolveSession(c *fiber.Ctx) (*domainSession.Session, *fiber.Error) {
	idOrName := c.Params("sessionId")

	sess, err := h.sessionResolver.ResolveSession(c.Context(), idOrName)
	if err != nil {
		h.logger.WarnWithFields("Failed to resolve session", map[string]interface{}{
			"identifier": idOrName,
			"error":      err.Error(),
			"path":       c.Path(),
		})

		if err.Error() == "session not found" || err == domainSession.ErrSessionNotFound {
			return nil, fiber.NewError(404, "Session not found")
		}

		return nil, fiber.NewError(500, "Internal server error")
	}

	return sess, nil
}

// CreateNewsletter creates a new WhatsApp newsletter/channel
// POST /sessions/:sessionId/newsletters/create
func (h *NewsletterHandler) CreateNewsletter(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req newsletter.CreateNewsletterRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Failed to parse create newsletter request", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	h.logger.InfoWithFields("Creating newsletter", map[string]interface{}{
		"session_id": sess.ID.String(),
		"name":       req.Name,
	})

	response, err := h.newsletterUC.CreateNewsletter(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to create newsletter", map[string]interface{}{
			"session_id": sess.ID.String(),
			"name":       req.Name,
			"error":      err.Error(),
		})

		if err.Error() == "session is not connected" {
			return fiber.NewError(400, "Session is not connected")
		}

		return fiber.NewError(500, "Failed to create newsletter")
	}

	h.logger.InfoWithFields("Newsletter created successfully", map[string]interface{}{
		"session_id":    sess.ID,
		"newsletter_id": response.ID,
		"name":          response.Name,
	})

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// GetNewsletterInfo gets information about a newsletter by JID
// GET /sessions/:sessionId/newsletters/info?jid=...
func (h *NewsletterHandler) GetNewsletterInfo(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	jid := c.Query("jid")
	if jid == "" {
		return fiber.NewError(400, "JID parameter is required")
	}

	req := &newsletter.GetNewsletterInfoRequest{
		JID: jid,
	}

	h.logger.InfoWithFields("Getting newsletter info", map[string]interface{}{
		"session_id": sess.ID.String(),
		"jid":        jid,
	})

	response, err := h.newsletterUC.GetNewsletterInfo(c.Context(), sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get newsletter info", map[string]interface{}{
			"session_id": sess.ID.String(),
			"jid":        jid,
			"error":      err.Error(),
		})

		if err.Error() == "session is not connected" {
			return fiber.NewError(400, "Session is not connected")
		}

		if err.Error() == "newsletter not found" {
			return fiber.NewError(404, "Newsletter not found")
		}

		return fiber.NewError(500, "Failed to get newsletter info")
	}

	h.logger.InfoWithFields("Newsletter info retrieved successfully", map[string]interface{}{
		"session_id":    sess.ID,
		"newsletter_id": response.ID,
		"name":          response.Name,
	})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// GetNewsletterInfoWithInvite gets newsletter information using an invite key
// POST /sessions/:sessionId/newsletters/info-from-invite
func (h *NewsletterHandler) GetNewsletterInfoWithInvite(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req newsletter.GetNewsletterInfoWithInviteRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Failed to parse get newsletter info with invite request", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	h.logger.InfoWithFields("Getting newsletter info with invite", map[string]interface{}{
		"session_id": sess.ID.String(),
		"invite_key": req.InviteKey,
	})

	response, err := h.newsletterUC.GetNewsletterInfoWithInvite(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get newsletter info with invite", map[string]interface{}{
			"session_id": sess.ID.String(),
			"invite_key": req.InviteKey,
			"error":      err.Error(),
		})

		if err.Error() == "session is not connected" {
			return fiber.NewError(400, "Session is not connected")
		}

		if err.Error() == "newsletter not found" {
			return fiber.NewError(404, "Newsletter not found")
		}

		return fiber.NewError(500, "Failed to get newsletter info with invite")
	}

	h.logger.InfoWithFields("Newsletter info retrieved with invite successfully", map[string]interface{}{
		"session_id":    sess.ID,
		"newsletter_id": response.ID,
		"name":          response.Name,
	})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// FollowNewsletter follows a newsletter
// POST /sessions/:sessionId/newsletters/follow
func (h *NewsletterHandler) FollowNewsletter(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req newsletter.FollowNewsletterRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Failed to parse follow newsletter request", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	h.logger.InfoWithFields("Following newsletter", map[string]interface{}{
		"session_id": sess.ID.String(),
		"jid":        req.JID,
	})

	response, err := h.newsletterUC.FollowNewsletter(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to follow newsletter", map[string]interface{}{
			"session_id": sess.ID.String(),
			"jid":        req.JID,
			"error":      err.Error(),
		})

		if err.Error() == "session is not connected" {
			return fiber.NewError(400, "Session is not connected")
		}

		if err.Error() == "newsletter not found" {
			return fiber.NewError(404, "Newsletter not found")
		}

		return fiber.NewError(500, "Failed to follow newsletter")
	}

	h.logger.InfoWithFields("Newsletter followed successfully", map[string]interface{}{
		"session_id": sess.ID.String(),
		"jid":        req.JID,
	})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// UnfollowNewsletter unfollows a newsletter
// POST /sessions/:sessionId/newsletters/unfollow
func (h *NewsletterHandler) UnfollowNewsletter(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req newsletter.UnfollowNewsletterRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Failed to parse unfollow newsletter request", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	h.logger.InfoWithFields("Unfollowing newsletter", map[string]interface{}{
		"session_id": sess.ID.String(),
		"jid":        req.JID,
	})

	response, err := h.newsletterUC.UnfollowNewsletter(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to unfollow newsletter", map[string]interface{}{
			"session_id": sess.ID.String(),
			"jid":        req.JID,
			"error":      err.Error(),
		})

		if err.Error() == "session is not connected" {
			return fiber.NewError(400, "Session is not connected")
		}

		if err.Error() == "newsletter not found" {
			return fiber.NewError(404, "Newsletter not found")
		}

		return fiber.NewError(500, "Failed to unfollow newsletter")
	}

	h.logger.InfoWithFields("Newsletter unfollowed successfully", map[string]interface{}{
		"session_id": sess.ID.String(),
		"jid":        req.JID,
	})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// GetSubscribedNewsletters gets all newsletters the user is subscribed to
// GET /sessions/:sessionId/newsletters
func (h *NewsletterHandler) GetSubscribedNewsletters(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	h.logger.InfoWithFields("Getting subscribed newsletters", map[string]interface{}{
		"session_id": sess.ID.String(),
	})

	response, err := h.newsletterUC.GetSubscribedNewsletters(c.Context(), sess.ID.String())
	if err != nil {
		h.logger.ErrorWithFields("Failed to get subscribed newsletters", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})

		if err.Error() == "session is not connected" {
			return fiber.NewError(400, "Session is not connected")
		}

		return fiber.NewError(500, "Failed to get subscribed newsletters")
	}

	h.logger.InfoWithFields("Subscribed newsletters retrieved successfully", map[string]interface{}{
		"session_id": sess.ID.String(),
		"count":      response.Total,
	})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}
