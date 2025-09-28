package handlers

import (
	"zpwoot/internal/app/community"
	domainSession "zpwoot/internal/domain/session"
	"zpwoot/internal/infra/http/helpers"
	"zpwoot/platform/logger"

	"github.com/gofiber/fiber/v2"
)

// CommunityHandler handles community-related HTTP requests
type CommunityHandler struct {
	logger          *logger.Logger
	communityUC     community.UseCase
	sessionResolver *helpers.SessionResolver
}

// NewCommunityHandler creates a new community handler
func NewCommunityHandler(appLogger *logger.Logger, communityUC community.UseCase, sessionRepo helpers.SessionRepository) *CommunityHandler {
	return &CommunityHandler{
		logger:          appLogger,
		communityUC:     communityUC,
		sessionResolver: helpers.NewSessionResolver(appLogger, sessionRepo),
	}
}

// resolveSession resolves session from URL parameter
func (h *CommunityHandler) resolveSession(c *fiber.Ctx) (*domainSession.Session, *fiber.Error) {
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

// LinkGroup links a group to a community
// POST /sessions/:sessionId/communities/link-group
func (h *CommunityHandler) LinkGroup(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req community.LinkGroupRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Failed to parse link group request", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	h.logger.InfoWithFields("Linking group to community", map[string]interface{}{
		"session_id":    sess.ID.String(),
		"community_jid": req.CommunityJID,
		"group_jid":     req.GroupJID,
	})

	response, err := h.communityUC.LinkGroup(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to link group to community", map[string]interface{}{
			"session_id":    sess.ID.String(),
			"community_jid": req.CommunityJID,
			"group_jid":     req.GroupJID,
			"error":         err.Error(),
		})

		if err.Error() == "session is not connected" {
			return fiber.NewError(400, "Session is not connected")
		}

		if err.Error() == "validation failed" {
			return fiber.NewError(400, "Invalid request data")
		}

		return fiber.NewError(500, "Failed to link group to community")
	}

	h.logger.InfoWithFields("Group linked to community successfully", map[string]interface{}{
		"session_id":    sess.ID,
		"community_jid": response.CommunityJID,
		"group_jid":     response.GroupJID,
	})

	return c.Status(200).JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// UnlinkGroup unlinks a group from a community
// POST /sessions/:sessionId/communities/unlink-group
func (h *CommunityHandler) UnlinkGroup(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req community.UnlinkGroupRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Failed to parse unlink group request", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	h.logger.InfoWithFields("Unlinking group from community", map[string]interface{}{
		"session_id":    sess.ID.String(),
		"community_jid": req.CommunityJID,
		"group_jid":     req.GroupJID,
	})

	response, err := h.communityUC.UnlinkGroup(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to unlink group from community", map[string]interface{}{
			"session_id":    sess.ID.String(),
			"community_jid": req.CommunityJID,
			"group_jid":     req.GroupJID,
			"error":         err.Error(),
		})

		if err.Error() == "session is not connected" {
			return fiber.NewError(400, "Session is not connected")
		}

		if err.Error() == "validation failed" {
			return fiber.NewError(400, "Invalid request data")
		}

		return fiber.NewError(500, "Failed to unlink group from community")
	}

	h.logger.InfoWithFields("Group unlinked from community successfully", map[string]interface{}{
		"session_id":    sess.ID,
		"community_jid": response.CommunityJID,
		"group_jid":     response.GroupJID,
	})

	return c.Status(200).JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// GetCommunityInfo gets information about a community
// GET /sessions/:sessionId/communities/info?communityJid=...
func (h *CommunityHandler) GetCommunityInfo(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	communityJid := c.Query("communityJid")
	if communityJid == "" {
		return fiber.NewError(400, "Community JID parameter is required")
	}

	req := &community.GetCommunityInfoRequest{
		CommunityJID: communityJid,
	}

	h.logger.InfoWithFields("Getting community info", map[string]interface{}{
		"session_id":    sess.ID.String(),
		"community_jid": communityJid,
	})

	response, err := h.communityUC.GetCommunityInfo(c.Context(), sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get community info", map[string]interface{}{
			"session_id":    sess.ID.String(),
			"community_jid": communityJid,
			"error":         err.Error(),
		})

		if err.Error() == "session is not connected" {
			return fiber.NewError(400, "Session is not connected")
		}

		if err.Error() == "community not found" {
			return fiber.NewError(404, "Community not found")
		}

		return fiber.NewError(500, "Failed to get community info")
	}

	h.logger.InfoWithFields("Community info retrieved successfully", map[string]interface{}{
		"session_id":    sess.ID,
		"community_jid": response.JID,
		"name":          response.Name,
	})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// GetSubGroups gets all sub-groups of a community
// GET /sessions/:sessionId/communities/subgroups?communityJid=...
func (h *CommunityHandler) GetSubGroups(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	communityJid := c.Query("communityJid")
	if communityJid == "" {
		return fiber.NewError(400, "Community JID parameter is required")
	}

	req := &community.GetSubGroupsRequest{
		CommunityJID: communityJid,
	}

	h.logger.InfoWithFields("Getting community sub-groups", map[string]interface{}{
		"session_id":    sess.ID.String(),
		"community_jid": communityJid,
	})

	response, err := h.communityUC.GetSubGroups(c.Context(), sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get community sub-groups", map[string]interface{}{
			"session_id":    sess.ID.String(),
			"community_jid": communityJid,
			"error":         err.Error(),
		})

		if err.Error() == "session is not connected" {
			return fiber.NewError(400, "Session is not connected")
		}

		if err.Error() == "community not found" {
			return fiber.NewError(404, "Community not found")
		}

		return fiber.NewError(500, "Failed to get community sub-groups")
	}

	h.logger.InfoWithFields("Community sub-groups retrieved successfully", map[string]interface{}{
		"session_id":    sess.ID,
		"community_jid": response.CommunityJID,
		"count":         response.TotalCount,
	})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}
