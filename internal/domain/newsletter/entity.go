package newsletter

import (
	"errors"
	"strings"
	"time"
)

// Domain errors
var (
	ErrInvalidNewsletterJID  = errors.New("invalid newsletter JID")
	ErrInvalidNewsletterName = errors.New("invalid newsletter name")
	ErrNewsletterNameTooLong = errors.New("newsletter name too long (max 25 characters)")
	ErrDescriptionTooLong    = errors.New("description too long (max 512 characters)")
	ErrInvalidInviteKey      = errors.New("invalid invite key")
	ErrNewsletterNotFound    = errors.New("newsletter not found")
	ErrNotNewsletterAdmin    = errors.New("user is not a newsletter admin")
	ErrEmptyNewsletterName   = errors.New("newsletter name cannot be empty")
	ErrInvalidNewsletterRole = errors.New("invalid newsletter role")
	ErrInvalidNewsletterState = errors.New("invalid newsletter state")
)

// NewsletterRole represents the user's role in a newsletter
type NewsletterRole string

const (
	NewsletterRoleSubscriber NewsletterRole = "subscriber"
	NewsletterRoleGuest      NewsletterRole = "guest"
	NewsletterRoleAdmin      NewsletterRole = "admin"
	NewsletterRoleOwner      NewsletterRole = "owner"
)

// NewsletterState represents the state of a newsletter
type NewsletterState string

const (
	NewsletterStateActive       NewsletterState = "active"
	NewsletterStateSuspended    NewsletterState = "suspended"
	NewsletterStateGeoSuspended NewsletterState = "geosuspended"
)

// NewsletterMuteState represents the mute status
type NewsletterMuteState string

const (
	NewsletterMuteOn  NewsletterMuteState = "on"
	NewsletterMuteOff NewsletterMuteState = "off"
)

// NewsletterVerificationState represents verification status
type NewsletterVerificationState string

const (
	NewsletterVerificationStateVerified   NewsletterVerificationState = "verified"
	NewsletterVerificationStateUnverified NewsletterVerificationState = "unverified"
)

// ProfilePictureInfo represents newsletter profile picture information
type ProfilePictureInfo struct {
	URL    string `json:"url"`
	ID     string `json:"id"`
	Type   string `json:"type"`
	Direct string `json:"direct"`
}

// NewsletterInfo represents a WhatsApp newsletter/channel
type NewsletterInfo struct {
	ID              string                      `json:"id"`
	Name            string                      `json:"name"`
	Description     string                      `json:"description"`
	InviteCode      string                      `json:"inviteCode"`
	SubscriberCount int                         `json:"subscriberCount"`
	State           NewsletterState             `json:"state"`
	Role            NewsletterRole              `json:"role"`
	Muted           bool                        `json:"muted"`
	MuteState       NewsletterMuteState         `json:"muteState"`
	Verified        bool                        `json:"verified"`
	VerificationState NewsletterVerificationState `json:"verificationState"`
	CreationTime    time.Time                   `json:"creationTime"`
	UpdateTime      time.Time                   `json:"updateTime"`
	Picture         *ProfilePictureInfo         `json:"picture,omitempty"`
	Preview         *ProfilePictureInfo         `json:"preview,omitempty"`
}

// CreateNewsletterRequest represents the data needed to create a newsletter
type CreateNewsletterRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Picture     []byte `json:"picture,omitempty"`
}

// GetNewsletterInfoRequest represents a request to get newsletter information
type GetNewsletterInfoRequest struct {
	JID string `json:"jid"`
}

// GetNewsletterInfoWithInviteRequest represents a request to get newsletter info via invite
type GetNewsletterInfoWithInviteRequest struct {
	InviteKey string `json:"inviteKey"`
}

// FollowNewsletterRequest represents a request to follow a newsletter
type FollowNewsletterRequest struct {
	JID string `json:"jid"`
}

// UnfollowNewsletterRequest represents a request to unfollow a newsletter
type UnfollowNewsletterRequest struct {
	JID string `json:"jid"`
}

// Business logic methods

// Validate validates the newsletter information
func (n *NewsletterInfo) Validate() error {
	if n.ID == "" {
		return ErrInvalidNewsletterJID
	}
	if n.Name == "" {
		return ErrEmptyNewsletterName
	}
	if len(n.Name) > 25 {
		return ErrNewsletterNameTooLong
	}
	if len(n.Description) > 512 {
		return ErrDescriptionTooLong
	}
	return nil
}

// IsAdmin checks if the current user is an admin of the newsletter
func (n *NewsletterInfo) IsAdmin() bool {
	return n.Role == NewsletterRoleAdmin || n.Role == NewsletterRoleOwner
}

// IsOwner checks if the current user is the owner of the newsletter
func (n *NewsletterInfo) IsOwner() bool {
	return n.Role == NewsletterRoleOwner
}

// IsActive checks if the newsletter is active
func (n *NewsletterInfo) IsActive() bool {
	return n.State == NewsletterStateActive
}

// IsMuted checks if the newsletter is muted
func (n *NewsletterInfo) IsMuted() bool {
	return n.Muted || n.MuteState == NewsletterMuteOn
}

// IsVerified checks if the newsletter is verified
func (n *NewsletterInfo) IsVerified() bool {
	return n.Verified || n.VerificationState == NewsletterVerificationStateVerified
}

// CanManage checks if the current user can manage the newsletter
func (n *NewsletterInfo) CanManage() bool {
	return n.IsAdmin() && n.IsActive()
}

// GetDisplayName returns the display name for the newsletter
func (n *NewsletterInfo) GetDisplayName() string {
	if n.Name != "" {
		return n.Name
	}
	return n.ID
}

// HasPicture checks if the newsletter has a profile picture
func (n *NewsletterInfo) HasPicture() bool {
	return n.Picture != nil && n.Picture.URL != ""
}

// Validate validates the create newsletter request
func (req *CreateNewsletterRequest) Validate() error {
	if req.Name == "" {
		return ErrEmptyNewsletterName
	}
	if len(req.Name) > 25 {
		return ErrNewsletterNameTooLong
	}
	if len(req.Description) > 512 {
		return ErrDescriptionTooLong
	}
	return nil
}

// Validate validates the get newsletter info request
func (req *GetNewsletterInfoRequest) Validate() error {
	if req.JID == "" {
		return ErrInvalidNewsletterJID
	}
	if !strings.Contains(req.JID, "@newsletter") {
		return ErrInvalidNewsletterJID
	}
	return nil
}

// Validate validates the get newsletter info with invite request
func (req *GetNewsletterInfoWithInviteRequest) Validate() error {
	if req.InviteKey == "" {
		return ErrInvalidInviteKey
	}
	// Remove common prefixes if present
	req.InviteKey = strings.TrimPrefix(req.InviteKey, "https://whatsapp.com/channel/")
	req.InviteKey = strings.TrimPrefix(req.InviteKey, "whatsapp.com/channel/")
	req.InviteKey = strings.TrimPrefix(req.InviteKey, "channel/")
	
	if req.InviteKey == "" {
		return ErrInvalidInviteKey
	}
	return nil
}

// Validate validates the follow newsletter request
func (req *FollowNewsletterRequest) Validate() error {
	if req.JID == "" {
		return ErrInvalidNewsletterJID
	}
	if !strings.Contains(req.JID, "@newsletter") {
		return ErrInvalidNewsletterJID
	}
	return nil
}

// Validate validates the unfollow newsletter request
func (req *UnfollowNewsletterRequest) Validate() error {
	if req.JID == "" {
		return ErrInvalidNewsletterJID
	}
	if !strings.Contains(req.JID, "@newsletter") {
		return ErrInvalidNewsletterJID
	}
	return nil
}

// Helper functions

// IsValidNewsletterJID checks if a JID is a valid newsletter JID
func IsValidNewsletterJID(jid string) bool {
	return strings.Contains(jid, "@newsletter")
}

// IsValidNewsletterRole checks if a role is valid
func IsValidNewsletterRole(role string) bool {
	switch NewsletterRole(role) {
	case NewsletterRoleSubscriber, NewsletterRoleGuest, NewsletterRoleAdmin, NewsletterRoleOwner:
		return true
	default:
		return false
	}
}

// IsValidNewsletterState checks if a state is valid
func IsValidNewsletterState(state string) bool {
	switch NewsletterState(state) {
	case NewsletterStateActive, NewsletterStateSuspended, NewsletterStateGeoSuspended:
		return true
	default:
		return false
	}
}

// ParseNewsletterRole parses a string to NewsletterRole
func ParseNewsletterRole(role string) (NewsletterRole, error) {
	if !IsValidNewsletterRole(role) {
		return "", ErrInvalidNewsletterRole
	}
	return NewsletterRole(role), nil
}

// ParseNewsletterState parses a string to NewsletterState
func ParseNewsletterState(state string) (NewsletterState, error) {
	if !IsValidNewsletterState(state) {
		return "", ErrInvalidNewsletterState
	}
	return NewsletterState(state), nil
}
