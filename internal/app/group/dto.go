package group

import (
	"time"

	"zpwoot/internal/domain/group"
)

// CreateGroupRequest represents the request to create a new group
type CreateGroupRequest struct {
	Name         string   `json:"name" validate:"required,min=1,max=25" example:"My Group"`
	Participants []string `json:"participants" validate:"required,min=1" example:"5511999999999@s.whatsapp.net,5511888888888@s.whatsapp.net"`
	Description  string   `json:"description,omitempty" validate:"max=512" example:"Group description"`
} //@name CreateGroupRequest

// CreateGroupResponse represents the response after creating a group
type CreateGroupResponse struct {
	GroupJID     string    `json:"groupJid" example:"120363123456789012@g.us"`
	Name         string    `json:"name" example:"My Group"`
	Description  string    `json:"description,omitempty" example:"Group description"`
	Participants []string  `json:"participants" example:"5511999999999@s.whatsapp.net,5511888888888@s.whatsapp.net"`
	CreatedAt    time.Time `json:"createdAt" example:"2024-01-01T00:00:00Z"`
} //@name CreateGroupResponse

// GetGroupInfoRequest represents the request to get group information
type GetGroupInfoRequest struct {
	GroupJID string `json:"groupJid" validate:"required" example:"120363123456789012@g.us"`
} //@name GetGroupInfoRequest

// GetGroupInfoResponse represents the group information response
type GetGroupInfoResponse struct {
	GroupJID     string              `json:"groupJid" example:"120363123456789012@g.us"`
	Name         string              `json:"name" example:"My Group"`
	Description  string              `json:"description,omitempty" example:"Group description"`
	Owner        string              `json:"owner" example:"5511999999999@s.whatsapp.net"`
	Participants []GroupParticipant  `json:"participants"`
	Settings     GroupSettings       `json:"settings"`
	CreatedAt    time.Time           `json:"createdAt" example:"2024-01-01T00:00:00Z"`
	UpdatedAt    time.Time           `json:"updatedAt" example:"2024-01-01T00:00:00Z"`
} //@name GetGroupInfoResponse

// GroupParticipant represents a group participant
type GroupParticipant struct {
	JID      string `json:"jid" example:"5511999999999@s.whatsapp.net"`
	IsAdmin  bool   `json:"isAdmin" example:"false"`
	IsSuperAdmin bool `json:"isSuperAdmin" example:"false"`
} //@name GroupParticipant

// GroupSettings represents group settings
type GroupSettings struct {
	Announce bool `json:"announce" example:"false"` // Only admins can send messages
	Locked   bool `json:"locked" example:"false"`   // Only admins can edit group info
} //@name GroupSettings

// ListGroupsResponse represents the response for listing joined groups
type ListGroupsResponse struct {
	Groups []GroupInfo `json:"groups"`
	Total  int         `json:"total" example:"5"`
} //@name ListGroupsResponse

// GroupInfo represents basic group information
type GroupInfo struct {
	GroupJID         string    `json:"groupJid" example:"120363123456789012@g.us"`
	Name             string    `json:"name" example:"My Group"`
	Description      string    `json:"description,omitempty" example:"Group description"`
	ParticipantCount int       `json:"participantCount" example:"10"`
	IsAdmin          bool      `json:"isAdmin" example:"true"`
	CreatedAt        time.Time `json:"createdAt" example:"2024-01-01T00:00:00Z"`
} //@name GroupInfo

// UpdateGroupParticipantsRequest represents the request to add/remove participants
type UpdateGroupParticipantsRequest struct {
	GroupJID     string   `json:"groupJid" validate:"required" example:"120363123456789012@g.us"`
	Participants []string `json:"participants" validate:"required,min=1" example:"5511999999999@s.whatsapp.net,5511888888888@s.whatsapp.net"`
	Action       string   `json:"action" validate:"required,oneof=add remove promote demote" example:"add"`
} //@name UpdateGroupParticipantsRequest

// UpdateGroupParticipantsResponse represents the response after updating participants
type UpdateGroupParticipantsResponse struct {
	GroupJID     string   `json:"groupJid" example:"120363123456789012@g.us"`
	Participants []string `json:"participants" example:"5511999999999@s.whatsapp.net,5511888888888@s.whatsapp.net"`
	Action       string   `json:"action" example:"add"`
	Success      []string `json:"success" example:"5511999999999@s.whatsapp.net"`
	Failed       []string `json:"failed,omitempty" example:"5511888888888@s.whatsapp.net"`
} //@name UpdateGroupParticipantsResponse

// SetGroupNameRequest represents the request to set group name
type SetGroupNameRequest struct {
	GroupJID string `json:"groupJid" validate:"required" example:"120363123456789012@g.us"`
	Name     string `json:"name" validate:"required,min=1,max=25" example:"New Group Name"`
} //@name SetGroupNameRequest

// SetGroupDescriptionRequest represents the request to set group description
type SetGroupDescriptionRequest struct {
	GroupJID    string `json:"groupJid" validate:"required" example:"120363123456789012@g.us"`
	Description string `json:"description" validate:"max=512" example:"New group description"`
} //@name SetGroupDescriptionRequest

// GetGroupInviteLinkRequest represents the request to get group invite link
type GetGroupInviteLinkRequest struct {
	GroupJID string `json:"groupJid" validate:"required" example:"120363123456789012@g.us"`
	Reset    bool   `json:"reset,omitempty" example:"false"`
} //@name GetGroupInviteLinkRequest

// GetGroupInviteLinkResponse represents the group invite link response
type GetGroupInviteLinkResponse struct {
	GroupJID   string `json:"groupJid" example:"120363123456789012@g.us"`
	InviteLink string `json:"inviteLink" example:"https://chat.whatsapp.com/ABC123DEF456"`
} //@name GetGroupInviteLinkResponse

// JoinGroupRequest represents the request to join a group via invite link
type JoinGroupRequest struct {
	InviteLink string `json:"inviteLink" validate:"required" example:"https://chat.whatsapp.com/ABC123DEF456"`
} //@name JoinGroupRequest

// JoinGroupResponse represents the response after joining a group
type JoinGroupResponse struct {
	GroupJID string `json:"groupJid" example:"120363123456789012@g.us"`
	Name     string `json:"name" example:"My Group"`
	Status   string `json:"status" example:"joined"`
} //@name JoinGroupResponse

// LeaveGroupRequest represents the request to leave a group
type LeaveGroupRequest struct {
	GroupJID string `json:"groupJid" validate:"required" example:"120363123456789012@g.us"`
} //@name LeaveGroupRequest

// LeaveGroupResponse represents the response after leaving a group
type LeaveGroupResponse struct {
	GroupJID string `json:"groupJid" example:"120363123456789012@g.us"`
	Status   string `json:"status" example:"left"`
} //@name LeaveGroupResponse

// SetGroupPhotoRequest represents the request to set group photo
type SetGroupPhotoRequest struct {
	GroupJID string `json:"groupJid" validate:"required" example:"120363123456789012@g.us"`
	Photo    string `json:"photo" validate:"required" example:"data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQ..."`
} //@name SetGroupPhotoRequest

// UpdateGroupSettingsRequest represents the request to update group settings
type UpdateGroupSettingsRequest struct {
	GroupJID string `json:"groupJid" validate:"required" example:"120363123456789012@g.us"`
	Announce *bool  `json:"announce,omitempty" example:"true"`
	Locked   *bool  `json:"locked,omitempty" example:"false"`
} //@name UpdateGroupSettingsRequest

// GroupActionResponse represents a generic response for group actions
type GroupActionResponse struct {
	GroupJID  string    `json:"groupJid" example:"120363123456789012@g.us"`
	Status    string    `json:"status" example:"success"`
	Message   string    `json:"message" example:"Action completed successfully"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T00:00:00Z"`
} //@name GroupActionResponse

// Conversion functions to/from domain models
func (r *CreateGroupRequest) ToDomain() *group.CreateGroupRequest {
	return &group.CreateGroupRequest{
		Name:         r.Name,
		Participants: r.Participants,
		Description:  r.Description,
	}
}

func FromDomainGroupInfo(g *group.GroupInfo) *GetGroupInfoResponse {
	participants := make([]GroupParticipant, len(g.Participants))
	for i, p := range g.Participants {
		participants[i] = GroupParticipant{
			JID:      p.JID,
			IsAdmin:  p.IsAdmin,
			IsSuperAdmin: p.IsSuperAdmin,
		}
	}

	return &GetGroupInfoResponse{
		GroupJID:     g.GroupJID,
		Name:         g.Name,
		Description:  g.Description,
		Owner:        g.Owner,
		Participants: participants,
		Settings: GroupSettings{
			Announce: g.Settings.Announce,
			Locked:   g.Settings.Locked,
		},
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
	}
}

func FromDomainGroupList(groups []*group.GroupInfo) []GroupInfo {
	result := make([]GroupInfo, len(groups))
	for i, g := range groups {
		result[i] = GroupInfo{
			GroupJID:         g.GroupJID,
			Name:             g.Name,
			Description:      g.Description,
			ParticipantCount: len(g.Participants),
			IsAdmin:          g.IsCurrentUserAdmin(),
			CreatedAt:        g.CreatedAt,
		}
	}
	return result
}
