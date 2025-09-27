package repository

import (
	"context"
	"fmt"

	"zpwoot/internal/domain/group"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

type GroupRepository struct {
	logger       *logger.Logger
	wameowMgr    ports.WameowManager
}

func NewGroupRepository(logger *logger.Logger, wameowMgr ports.WameowManager) *GroupRepository {
	return &GroupRepository{
		logger:       logger,
		wameowMgr:    wameowMgr,
	}
}

func (r *GroupRepository) CreateGroup(ctx context.Context, sessionID string, req *group.CreateGroupRequest) (*group.GroupInfo, error) {
	r.logger.InfoWithFields("Creating group via repository", map[string]interface{}{
		"session_id": sessionID,
		"name":       req.Name,
		"participants": len(req.Participants),
	})

	// Use wameow manager to create group
	groupInfo, err := r.wameowMgr.CreateGroup(sessionID, req.Name, req.Participants, req.Description)
	if err != nil {
		r.logger.ErrorWithFields("Failed to create group", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, err
	}

	// Convert ports.GroupInfo to domain.GroupInfo
	return r.convertToGroupInfo(groupInfo), nil
}

func (r *GroupRepository) GetGroupInfo(ctx context.Context, sessionID, groupJID string) (*group.GroupInfo, error) {
	r.logger.InfoWithFields("Getting group info via repository", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  groupJID,
	})

	groupInfo, err := r.wameowMgr.GetGroupInfo(sessionID, groupJID)
	if err != nil {
		r.logger.ErrorWithFields("Failed to get group info", map[string]interface{}{
			"session_id": sessionID,
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return nil, err
	}

	return r.convertToGroupInfo(groupInfo), nil
}

func (r *GroupRepository) ListJoinedGroups(ctx context.Context, sessionID string) ([]*group.GroupInfo, error) {
	r.logger.InfoWithFields("Listing joined groups via repository", map[string]interface{}{
		"session_id": sessionID,
	})

	groups, err := r.wameowMgr.ListJoinedGroups(sessionID)
	if err != nil {
		r.logger.ErrorWithFields("Failed to list joined groups", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, err
	}

	var result []*group.GroupInfo
	for _, g := range groups {
		result = append(result, r.convertToGroupInfo(g))
	}

	return result, nil
}

func (r *GroupRepository) UpdateGroupParticipants(ctx context.Context, sessionID string, req *group.UpdateParticipantsRequest) (*group.UpdateParticipantsResult, error) {
	r.logger.InfoWithFields("Updating group participants via repository", map[string]interface{}{
		"session_id":   sessionID,
		"group_jid":    req.GroupJID,
		"action":       req.Action,
		"participants": len(req.Participants),
	})

	success, failed, err := r.wameowMgr.UpdateGroupParticipants(sessionID, req.GroupJID, req.Participants, req.Action)
	if err != nil {
		r.logger.ErrorWithFields("Failed to update group participants", map[string]interface{}{
			"session_id": sessionID,
			"group_jid":  req.GroupJID,
			"error":      err.Error(),
		})
		return nil, err
	}

	return &group.UpdateParticipantsResult{
		GroupJID:     req.GroupJID,
		Participants: req.Participants,
		Action:       req.Action,
		Success:      success,
		Failed:       failed,
	}, nil
}

func (r *GroupRepository) SetGroupName(ctx context.Context, sessionID string, req *group.SetGroupNameRequest) error {
	r.logger.InfoWithFields("Setting group name via repository", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  req.GroupJID,
		"name":       req.Name,
	})

	err := r.wameowMgr.SetGroupName(sessionID, req.GroupJID, req.Name)
	if err != nil {
		r.logger.ErrorWithFields("Failed to set group name", map[string]interface{}{
			"session_id": sessionID,
			"group_jid":  req.GroupJID,
			"error":      err.Error(),
		})
		return err
	}

	return nil
}

func (r *GroupRepository) SetGroupDescription(ctx context.Context, sessionID string, req *group.SetGroupDescriptionRequest) error {
	r.logger.InfoWithFields("Setting group description via repository", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  req.GroupJID,
		"description": req.Description,
	})

	err := r.wameowMgr.SetGroupDescription(sessionID, req.GroupJID, req.Description)
	if err != nil {
		r.logger.ErrorWithFields("Failed to set group description", map[string]interface{}{
			"session_id": sessionID,
			"group_jid":  req.GroupJID,
			"error":      err.Error(),
		})
		return err
	}

	return nil
}

func (r *GroupRepository) SetGroupPhoto(ctx context.Context, sessionID string, req *group.SetGroupPhotoRequest) error {
	r.logger.InfoWithFields("Setting group photo via repository", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  req.GroupJID,
	})

	// Convert photo string to bytes (simplified)
	photoBytes := []byte(req.Photo)

	err := r.wameowMgr.SetGroupPhoto(sessionID, req.GroupJID, photoBytes)
	if err != nil {
		r.logger.ErrorWithFields("Failed to set group photo", map[string]interface{}{
			"session_id": sessionID,
			"group_jid":  req.GroupJID,
			"error":      err.Error(),
		})
		return err
	}

	return nil
}

func (r *GroupRepository) GetGroupInviteLink(ctx context.Context, sessionID string, req *group.GetInviteLinkRequest) (*group.InviteLinkResponse, error) {
	r.logger.InfoWithFields("Getting group invite link via repository", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  req.GroupJID,
		"reset":      req.Reset,
	})

	inviteLink, err := r.wameowMgr.GetGroupInviteLink(sessionID, req.GroupJID, req.Reset)
	if err != nil {
		r.logger.ErrorWithFields("Failed to get group invite link", map[string]interface{}{
			"session_id": sessionID,
			"group_jid":  req.GroupJID,
			"error":      err.Error(),
		})
		return nil, err
	}

	return &group.InviteLinkResponse{
		GroupJID:   req.GroupJID,
		InviteLink: inviteLink,
	}, nil
}

func (r *GroupRepository) JoinGroupViaLink(ctx context.Context, sessionID string, req *group.JoinGroupRequest) (*group.GroupInfo, error) {
	r.logger.InfoWithFields("Joining group via link via repository", map[string]interface{}{
		"session_id": sessionID,
	})

	groupInfo, err := r.wameowMgr.JoinGroupViaLink(sessionID, req.InviteLink)
	if err != nil {
		r.logger.ErrorWithFields("Failed to join group via link", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, err
	}

	return r.convertToGroupInfo(groupInfo), nil
}

func (r *GroupRepository) LeaveGroup(ctx context.Context, sessionID string, req *group.LeaveGroupRequest) error {
	r.logger.InfoWithFields("Leaving group via repository", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  req.GroupJID,
	})

	err := r.wameowMgr.LeaveGroup(sessionID, req.GroupJID)
	if err != nil {
		r.logger.ErrorWithFields("Failed to leave group", map[string]interface{}{
			"session_id": sessionID,
			"group_jid":  req.GroupJID,
			"error":      err.Error(),
		})
		return err
	}

	return nil
}

func (r *GroupRepository) UpdateGroupSettings(ctx context.Context, sessionID string, req *group.UpdateGroupSettingsRequest) error {
	r.logger.InfoWithFields("Updating group settings via repository", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  req.GroupJID,
		"announce":   req.Announce,
		"locked":     req.Locked,
	})

	err := r.wameowMgr.UpdateGroupSettings(sessionID, req.GroupJID, req.Announce, req.Locked)
	if err != nil {
		r.logger.ErrorWithFields("Failed to update group settings", map[string]interface{}{
			"session_id": sessionID,
			"group_jid":  req.GroupJID,
			"error":      err.Error(),
		})
		return err
	}

	return nil
}

// Helper function to convert ports.GroupInfo to domain.GroupInfo
func (r *GroupRepository) convertToGroupInfo(portsGroup *ports.GroupInfo) *group.GroupInfo {
	if portsGroup == nil {
		return nil
	}

	var participants []group.GroupParticipant
	for _, p := range portsGroup.Participants {
		participants = append(participants, group.GroupParticipant{
			JID:          p.JID,
			IsAdmin:      p.IsAdmin,
			IsSuperAdmin: p.IsSuperAdmin,
		})
	}

	return &group.GroupInfo{
		GroupJID:     portsGroup.GroupJID,
		Name:         portsGroup.Name,
		Description:  portsGroup.Description,
		Owner:        portsGroup.Owner,
		Participants: participants,
		Settings: group.GroupSettings{
			Announce: portsGroup.Settings.Announce,
			Locked:   portsGroup.Settings.Locked,
		},
		CreatedAt: portsGroup.CreatedAt,
		UpdatedAt: portsGroup.UpdatedAt,
	}
}
