package handlers

import (
	"zpwoot/internal/app/group"
	domainSession "zpwoot/internal/domain/session"
	"zpwoot/internal/infra/http/helpers"
	"zpwoot/platform/logger"

	"github.com/gofiber/fiber/v2"
)

type GroupHandler struct {
	logger          *logger.Logger
	groupUC         group.UseCase
	sessionResolver *helpers.SessionResolver
}

func NewGroupHandler(appLogger *logger.Logger, groupUC group.UseCase, sessionRepo helpers.SessionRepository) *GroupHandler {
	return &GroupHandler{
		logger:          appLogger,
		groupUC:         groupUC,
		sessionResolver: helpers.NewSessionResolver(appLogger, sessionRepo),
	}
}

func (h *GroupHandler) resolveSession(c *fiber.Ctx) (*domainSession.Session, *fiber.Error) {
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

// CreateGroup creates a new WhatsApp group
func (h *GroupHandler) CreateGroup(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req group.CreateGroupRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Invalid request body", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	h.logger.InfoWithFields("Creating group", map[string]interface{}{
		"session_id": sess.ID.String(),
		"name":       req.Name,
		"participants": len(req.Participants),
	})

	response, err := h.groupUC.CreateGroup(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to create group", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(response)
}

// GetGroupInfo retrieves information about a specific group
func (h *GroupHandler) GetGroupInfo(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	groupJID := c.Params("groupJid")
	if groupJID == "" {
		return fiber.NewError(400, "Group JID is required")
	}

	req := &group.GetGroupInfoRequest{
		GroupJID: groupJID,
	}

	h.logger.InfoWithFields("Getting group info", map[string]interface{}{
		"session_id": sess.ID.String(),
		"group_jid":  groupJID,
	})

	response, err := h.groupUC.GetGroupInfo(c.Context(), sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get group info", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(response)
}

// ListGroups lists all groups the user is a member of
func (h *GroupHandler) ListGroups(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	h.logger.InfoWithFields("Listing groups", map[string]interface{}{
		"session_id": sess.ID.String(),
	})

	response, err := h.groupUC.ListGroups(c.Context(), sess.ID.String())
	if err != nil {
		h.logger.ErrorWithFields("Failed to list groups", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(response)
}

// UpdateGroupParticipants adds, removes, promotes, or demotes group participants
func (h *GroupHandler) UpdateGroupParticipants(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	groupJID := c.Params("groupJid")
	if groupJID == "" {
		return fiber.NewError(400, "Group JID is required")
	}

	var req group.UpdateGroupParticipantsRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Invalid request body", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	req.GroupJID = groupJID

	h.logger.InfoWithFields("Updating group participants", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"group_jid":    groupJID,
		"action":       req.Action,
		"participants": len(req.Participants),
	})

	response, err := h.groupUC.UpdateGroupParticipants(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to update group participants", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(response)
}

// SetGroupName updates the group name
func (h *GroupHandler) SetGroupName(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	groupJID := c.Params("groupJid")
	if groupJID == "" {
		return fiber.NewError(400, "Group JID is required")
	}

	var req group.SetGroupNameRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Invalid request body", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	req.GroupJID = groupJID

	h.logger.InfoWithFields("Setting group name", map[string]interface{}{
		"session_id": sess.ID.String(),
		"group_jid":  groupJID,
		"name":       req.Name,
	})

	response, err := h.groupUC.SetGroupName(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to set group name", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(response)
}

// SetGroupDescription updates the group description
func (h *GroupHandler) SetGroupDescription(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	groupJID := c.Params("groupJid")
	if groupJID == "" {
		return fiber.NewError(400, "Group JID is required")
	}

	var req group.SetGroupDescriptionRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Invalid request body", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	req.GroupJID = groupJID

	h.logger.InfoWithFields("Setting group description", map[string]interface{}{
		"session_id": sess.ID.String(),
		"group_jid":  groupJID,
		"description": req.Description,
	})

	response, err := h.groupUC.SetGroupDescription(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to set group description", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(response)
}

// SetGroupPhoto updates the group photo
func (h *GroupHandler) SetGroupPhoto(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	groupJID := c.Params("groupJid")
	if groupJID == "" {
		return fiber.NewError(400, "Group JID is required")
	}

	var req group.SetGroupPhotoRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Invalid request body", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	req.GroupJID = groupJID

	h.logger.InfoWithFields("Setting group photo", map[string]interface{}{
		"session_id": sess.ID.String(),
		"group_jid":  groupJID,
	})

	response, err := h.groupUC.SetGroupPhoto(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to set group photo", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(response)
}

// GetGroupInviteLink retrieves or generates a group invite link
func (h *GroupHandler) GetGroupInviteLink(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	groupJID := c.Params("groupJid")
	if groupJID == "" {
		return fiber.NewError(400, "Group JID is required")
	}

	reset := c.QueryBool("reset", false)

	req := &group.GetGroupInviteLinkRequest{
		GroupJID: groupJID,
		Reset:    reset,
	}

	h.logger.InfoWithFields("Getting group invite link", map[string]interface{}{
		"session_id": sess.ID.String(),
		"group_jid":  groupJID,
		"reset":      reset,
	})

	response, err := h.groupUC.GetGroupInviteLink(c.Context(), sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get group invite link", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(response)
}

// JoinGroup joins a group using an invite link
func (h *GroupHandler) JoinGroup(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req group.JoinGroupRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Invalid request body", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	h.logger.InfoWithFields("Joining group", map[string]interface{}{
		"session_id": sess.ID.String(),
	})

	response, err := h.groupUC.JoinGroup(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to join group", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(response)
}

// LeaveGroup leaves a group
func (h *GroupHandler) LeaveGroup(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	groupJID := c.Params("groupJid")
	if groupJID == "" {
		return fiber.NewError(400, "Group JID is required")
	}

	req := &group.LeaveGroupRequest{
		GroupJID: groupJID,
	}

	h.logger.InfoWithFields("Leaving group", map[string]interface{}{
		"session_id": sess.ID.String(),
		"group_jid":  groupJID,
	})

	response, err := h.groupUC.LeaveGroup(c.Context(), sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to leave group", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(response)
}

// UpdateGroupSettings updates group settings (announce, locked)
func (h *GroupHandler) UpdateGroupSettings(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	groupJID := c.Params("groupJid")
	if groupJID == "" {
		return fiber.NewError(400, "Group JID is required")
	}

	var req group.UpdateGroupSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Invalid request body", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	req.GroupJID = groupJID

	h.logger.InfoWithFields("Updating group settings", map[string]interface{}{
		"session_id": sess.ID.String(),
		"group_jid":  groupJID,
		"announce":   req.Announce,
		"locked":     req.Locked,
	})

	response, err := h.groupUC.UpdateGroupSettings(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to update group settings", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(response)
}
