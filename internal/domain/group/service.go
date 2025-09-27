package group

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"zpwoot/pkg/uuid"
)

type Service struct {
	repo      Repository
	wameow    WameowManager
	generator *uuid.Generator
}

type Repository interface {
	CreateGroup(ctx context.Context, sessionID string, req *CreateGroupRequest) (*GroupInfo, error)
	GetGroupInfo(ctx context.Context, sessionID, groupJID string) (*GroupInfo, error)
	ListJoinedGroups(ctx context.Context, sessionID string) ([]*GroupInfo, error)
	UpdateGroupParticipants(ctx context.Context, sessionID string, req *UpdateParticipantsRequest) (*UpdateParticipantsResult, error)
	SetGroupName(ctx context.Context, sessionID string, req *SetGroupNameRequest) error
	SetGroupDescription(ctx context.Context, sessionID string, req *SetGroupDescriptionRequest) error
	SetGroupPhoto(ctx context.Context, sessionID string, req *SetGroupPhotoRequest) error
	GetGroupInviteLink(ctx context.Context, sessionID string, req *GetInviteLinkRequest) (*InviteLinkResponse, error)
	JoinGroupViaLink(ctx context.Context, sessionID string, req *JoinGroupRequest) (*GroupInfo, error)
	LeaveGroup(ctx context.Context, sessionID string, req *LeaveGroupRequest) error
	UpdateGroupSettings(ctx context.Context, sessionID string, req *UpdateGroupSettingsRequest) error
}

type WameowManager interface {
	CreateGroup(sessionID, name string, participants []string, description string) (*GroupInfo, error)
	GetGroupInfo(sessionID, groupJID string) (*GroupInfo, error)
	ListJoinedGroups(sessionID string) ([]*GroupInfo, error)
	UpdateGroupParticipants(sessionID, groupJID string, participants []string, action string) ([]string, []string, error)
	SetGroupName(sessionID, groupJID, name string) error
	SetGroupDescription(sessionID, groupJID, description string) error
	SetGroupPhoto(sessionID, groupJID string, photo []byte) error
	GetGroupInviteLink(sessionID, groupJID string, reset bool) (string, error)
	JoinGroupViaLink(sessionID, inviteLink string) (*GroupInfo, error)
	LeaveGroup(sessionID, groupJID string) error
	UpdateGroupSettings(sessionID, groupJID string, announce, locked *bool) error
}

func NewService(repo Repository, wameow WameowManager) *Service {
	return &Service{
		repo:      repo,
		wameow:    wameow,
		generator: uuid.New(),
	}
}

// ValidateGroupCreation validates group creation parameters
func (s *Service) ValidateGroupCreation(req *CreateGroupRequest) error {
	if req == nil {
		return ErrInvalidGroupName
	}

	if err := s.ValidateGroupName(req.Name); err != nil {
		return err
	}

	if len(req.Participants) == 0 {
		return ErrNoParticipants
	}

	if len(req.Participants) > 256 {
		return fmt.Errorf("too many participants (max 256)")
	}

	// Validate participant JIDs
	for _, participant := range req.Participants {
		if err := s.validateJID(participant); err != nil {
			return fmt.Errorf("invalid participant %s: %w", participant, err)
		}
	}

	if err := s.ValidateGroupDescription(req.Description); err != nil {
		return err
	}

	return nil
}

// ValidateParticipantUpdate validates participant update parameters
func (s *Service) ValidateParticipantUpdate(req *UpdateParticipantsRequest) error {
	if req == nil {
		return ErrInvalidGroupJID
	}

	if err := s.validateJID(req.GroupJID); err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	if len(req.Participants) == 0 {
		return ErrNoParticipants
	}

	if len(req.Participants) > 50 {
		return fmt.Errorf("too many participants in single operation (max 50)")
	}

	// Validate participant JIDs
	for _, participant := range req.Participants {
		if err := s.validateJID(participant); err != nil {
			return fmt.Errorf("invalid participant %s: %w", participant, err)
		}
	}

	// Validate action
	validActions := []string{"add", "remove", "promote", "demote"}
	isValidAction := false
	for _, action := range validActions {
		if req.Action == action {
			isValidAction = true
			break
		}
	}
	if !isValidAction {
		return ErrInvalidAction
	}

	return nil
}

// ValidateGroupName validates group name
func (s *Service) ValidateGroupName(name string) error {
	if name == "" {
		return ErrInvalidGroupName
	}

	if len(name) > 25 {
		return ErrGroupNameTooLong
	}

	// Check for invalid characters (basic validation)
	if strings.TrimSpace(name) == "" {
		return ErrInvalidGroupName
	}

	return nil
}

// ValidateGroupDescription validates group description
func (s *Service) ValidateGroupDescription(description string) error {
	if len(description) > 512 {
		return ErrDescriptionTooLong
	}

	return nil
}

// ValidateInviteLink validates invite link format
func (s *Service) ValidateInviteLink(link string) error {
	if link == "" {
		return ErrInvalidInviteLink
	}

	// WhatsApp invite links follow the pattern: https://chat.whatsapp.com/XXXXXX
	inviteLinkPattern := `^https://chat\.whatsapp\.com/[A-Za-z0-9]+$`
	matched, err := regexp.MatchString(inviteLinkPattern, link)
	if err != nil {
		return fmt.Errorf("error validating invite link: %w", err)
	}

	if !matched {
		return ErrInvalidInviteLink
	}

	return nil
}

// CanPerformAction checks if user can perform a specific action on the group
func (s *Service) CanPerformAction(userJID, groupJID, action string, groupInfo *GroupInfo) error {
	if groupInfo == nil {
		return ErrGroupNotFound
	}

	// Check if user is a participant
	if !groupInfo.HasParticipant(userJID) {
		return ErrParticipantNotFound
	}

	// Actions that require admin privileges
	adminActions := []string{"remove", "promote", "demote", "set_name", "set_description", "set_photo", "set_settings"}
	requiresAdmin := false
	for _, adminAction := range adminActions {
		if action == adminAction {
			requiresAdmin = true
			break
		}
	}

	if requiresAdmin && !groupInfo.IsParticipantAdmin(userJID) {
		return ErrNotGroupAdmin
	}

	// Special case: cannot remove group owner
	if action == "remove" {
		for _, participant := range groupInfo.Participants {
			if participant.JID == userJID && participant.JID == groupInfo.Owner {
				return ErrCannotRemoveOwner
			}
		}
	}

	// Special case: group owner cannot leave
	if action == "leave" && userJID == groupInfo.Owner {
		return ErrCannotLeaveAsOwner
	}

	return nil
}

// ProcessParticipantChanges processes and validates participant changes
func (s *Service) ProcessParticipantChanges(req *UpdateParticipantsRequest, currentGroup *GroupInfo) error {
	if req == nil || currentGroup == nil {
		return fmt.Errorf("invalid request or group info")
	}

	switch req.Action {
	case "add":
		// Check if participants are already in the group
		for _, participant := range req.Participants {
			if currentGroup.HasParticipant(participant) {
				return fmt.Errorf("participant %s is already in the group", participant)
			}
		}

	case "remove":
		// Check if participants are in the group
		for _, participant := range req.Participants {
			if !currentGroup.HasParticipant(participant) {
				return fmt.Errorf("participant %s is not in the group", participant)
			}
			// Cannot remove group owner
			if participant == currentGroup.Owner {
				return ErrCannotRemoveOwner
			}
		}

	case "promote":
		// Check if participants are in the group and not already admins
		for _, participant := range req.Participants {
			if !currentGroup.HasParticipant(participant) {
				return fmt.Errorf("participant %s is not in the group", participant)
			}
			if currentGroup.IsParticipantAdmin(participant) {
				return fmt.Errorf("participant %s is already an admin", participant)
			}
		}

	case "demote":
		// Check if participants are admins
		for _, participant := range req.Participants {
			if !currentGroup.HasParticipant(participant) {
				return fmt.Errorf("participant %s is not in the group", participant)
			}
			if !currentGroup.IsParticipantAdmin(participant) {
				return fmt.Errorf("participant %s is not an admin", participant)
			}
			// Cannot demote group owner
			if participant == currentGroup.Owner {
				return fmt.Errorf("cannot demote group owner")
			}
		}
	}

	return nil
}

// validateJID validates WhatsApp JID format
func (s *Service) validateJID(jid string) error {
	if jid == "" {
		return fmt.Errorf("JID cannot be empty")
	}

	// Basic JID validation for WhatsApp
	// Individual: number@s.whatsapp.net
	// Group: number@g.us
	jidPattern := `^[0-9]+@(s\.whatsapp\.net|g\.us)$`
	matched, err := regexp.MatchString(jidPattern, jid)
	if err != nil {
		return fmt.Errorf("error validating JID: %w", err)
	}

	if !matched {
		return fmt.Errorf("invalid JID format")
	}

	return nil
}

// CreateGroup creates a new group with validation
func (s *Service) CreateGroup(ctx context.Context, sessionID string, req *CreateGroupRequest) (*GroupInfo, error) {
	if err := s.ValidateGroupCreation(req); err != nil {
		return nil, err
	}

	return s.repo.CreateGroup(ctx, sessionID, req)
}

// GetGroupInfo retrieves group information with validation
func (s *Service) GetGroupInfo(ctx context.Context, sessionID, groupJID string) (*GroupInfo, error) {
	if err := s.validateJID(groupJID); err != nil {
		return nil, fmt.Errorf("invalid group JID: %w", err)
	}

	return s.repo.GetGroupInfo(ctx, sessionID, groupJID)
}

// UpdateParticipants updates group participants with validation
func (s *Service) UpdateParticipants(ctx context.Context, sessionID string, req *UpdateParticipantsRequest) (*UpdateParticipantsResult, error) {
	if err := s.ValidateParticipantUpdate(req); err != nil {
		return nil, err
	}

	// Get current group info for additional validation
	groupInfo, err := s.repo.GetGroupInfo(ctx, sessionID, req.GroupJID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group info: %w", err)
	}

	if err := s.ProcessParticipantChanges(req, groupInfo); err != nil {
		return nil, err
	}

	return s.repo.UpdateGroupParticipants(ctx, sessionID, req)
}

// SetGroupName sets group name with validation
func (s *Service) SetGroupName(ctx context.Context, sessionID string, req *SetGroupNameRequest) error {
	if err := s.validateJID(req.GroupJID); err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	if err := s.ValidateGroupName(req.Name); err != nil {
		return err
	}

	return s.repo.SetGroupName(ctx, sessionID, req)
}

// SetGroupDescription sets group description with validation
func (s *Service) SetGroupDescription(ctx context.Context, sessionID string, req *SetGroupDescriptionRequest) error {
	if err := s.validateJID(req.GroupJID); err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	if err := s.ValidateGroupDescription(req.Description); err != nil {
		return err
	}

	return s.repo.SetGroupDescription(ctx, sessionID, req)
}
