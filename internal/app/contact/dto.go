package contact

import "time"

// CheckWhatsAppRequest represents a request to check if phone numbers are on WhatsApp
type CheckWhatsAppRequest struct {
	SessionID    string   `json:"session_id,omitempty"`
	PhoneNumbers []string `json:"phone_numbers" validate:"required,min=1,max=50" example:"[\"+5511999999999\", \"+5511888888888\"]"`
}

// WhatsAppStatus represents the status of a phone number on WhatsApp
type WhatsAppStatus struct {
	PhoneNumber  string `json:"phone_number" example:"+5511999999999"`
	IsOnWhatsApp bool   `json:"is_on_whatsapp" example:"true"`
	JID          string `json:"jid,omitempty" example:"5511999999999@s.whatsapp.net"`
	IsBusiness   bool   `json:"is_business" example:"false"`
	VerifiedName string `json:"verified_name,omitempty" example:"Company Name"`
}

// CheckWhatsAppResponse represents the response for checking WhatsApp numbers
type CheckWhatsAppResponse struct {
	Results []WhatsAppStatus `json:"results"`
	Total   int              `json:"total" example:"2"`
	Checked int              `json:"checked" example:"2"`
}

// GetProfilePictureRequest represents a request to get profile picture
type GetProfilePictureRequest struct {
	SessionID string `json:"session_id,omitempty"`
	JID       string `json:"jid" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Preview   bool   `json:"preview" example:"false"`
}

// ProfilePictureResponse represents profile picture information
type ProfilePictureResponse struct {
	JID        string    `json:"jid" example:"5511999999999@s.whatsapp.net"`
	URL        string    `json:"url,omitempty" example:"https://pps.whatsapp.net/v/..."`
	ID         string    `json:"id,omitempty" example:"1234567890"`
	Type       string    `json:"type,omitempty" example:"image"`
	DirectPath string    `json:"direct_path,omitempty"`
	UpdatedAt  time.Time `json:"updated_at,omitempty" example:"2024-01-01T12:00:00Z"`
	HasPicture bool      `json:"has_picture" example:"true"`
}

// GetUserInfoRequest represents a request to get user information
type GetUserInfoRequest struct {
	SessionID string   `json:"session_id,omitempty"`
	JIDs      []string `json:"jids" validate:"required,min=1,max=20" example:"[\"5511999999999@s.whatsapp.net\", \"5511888888888@s.whatsapp.net\"]"`
}

// UserInfo represents information about a WhatsApp user
type UserInfo struct {
	JID          string     `json:"jid" example:"5511999999999@s.whatsapp.net"`
	PhoneNumber  string     `json:"phone_number" example:"+5511999999999"`
	Name         string     `json:"name,omitempty" example:"John Doe"`
	Status       string     `json:"status,omitempty" example:"Hey there! I am using WhatsApp."`
	PictureID    string     `json:"picture_id,omitempty" example:"1234567890"`
	IsBusiness   bool       `json:"is_business" example:"false"`
	VerifiedName string     `json:"verified_name,omitempty" example:"Company Name"`
	IsContact    bool       `json:"is_contact" example:"true"`
	LastSeen     *time.Time `json:"last_seen,omitempty" example:"2024-01-01T12:00:00Z"`
	IsOnline     bool       `json:"is_online" example:"false"`
}

// GetUserInfoResponse represents the response for getting user information
type GetUserInfoResponse struct {
	Users []UserInfo `json:"users"`
	Total int        `json:"total" example:"2"`
	Found int        `json:"found" example:"2"`
}

// ListContactsRequest represents a request to list contacts
type ListContactsRequest struct {
	SessionID string `json:"session_id,omitempty"`
	Limit     int    `json:"limit" validate:"min=1,max=100" example:"50"`
	Offset    int    `json:"offset" validate:"min=0" example:"0"`
	Search    string `json:"search,omitempty" example:"John"`
}

// Contact represents a contact in the contact list
type Contact struct {
	JID         string    `json:"jid" example:"5511999999999@s.whatsapp.net"`
	PhoneNumber string    `json:"phone_number" example:"+5511999999999"`
	Name        string    `json:"name,omitempty" example:"John Doe"`
	ShortName   string    `json:"short_name,omitempty" example:"John"`
	PushName    string    `json:"push_name,omitempty" example:"John"`
	IsBusiness  bool      `json:"is_business" example:"false"`
	IsContact   bool      `json:"is_contact" example:"true"`
	IsBlocked   bool      `json:"is_blocked" example:"false"`
	AddedAt     time.Time `json:"added_at,omitempty" example:"2024-01-01T12:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at,omitempty" example:"2024-01-01T12:00:00Z"`
}

// ListContactsResponse represents the response for listing contacts
type ListContactsResponse struct {
	Contacts []Contact `json:"contacts"`
	Total    int       `json:"total" example:"150"`
	Limit    int       `json:"limit" example:"50"`
	Offset   int       `json:"offset" example:"0"`
	HasMore  bool      `json:"has_more" example:"true"`
}

// SyncContactsRequest represents a request to sync contacts
type SyncContactsRequest struct {
	SessionID string `json:"session_id,omitempty"`
	Force     bool   `json:"force" example:"false"` // Force full sync even if recently synced
}

// SyncContactsResponse represents the response for syncing contacts
type SyncContactsResponse struct {
	Synced   int       `json:"synced" example:"25"`
	Added    int       `json:"added" example:"5"`
	Updated  int       `json:"updated" example:"3"`
	Removed  int       `json:"removed" example:"1"`
	Total    int       `json:"total" example:"150"`
	SyncedAt time.Time `json:"synced_at" example:"2024-01-01T12:00:00Z"`
	Message  string    `json:"message" example:"Contacts synchronized successfully"`
}

// GetBusinessProfileRequest represents a request to get business profile
type GetBusinessProfileRequest struct {
	SessionID string `json:"session_id,omitempty"`
	JID       string `json:"jid" validate:"required" example:"5511999999999@s.whatsapp.net"`
}

// BusinessProfile represents a WhatsApp Business profile
type BusinessProfile struct {
	JID         string `json:"jid" example:"5511999999999@s.whatsapp.net"`
	Name        string `json:"name,omitempty" example:"My Business"`
	Category    string `json:"category,omitempty" example:"Retail"`
	Description string `json:"description,omitempty" example:"We sell amazing products"`
	Website     string `json:"website,omitempty" example:"https://mybusiness.com"`
	Email       string `json:"email,omitempty" example:"contact@mybusiness.com"`
	Address     string `json:"address,omitempty" example:"123 Main St, City"`
	Verified    bool   `json:"verified" example:"true"`
}

// BusinessProfileResponse represents the response for getting business profile
type BusinessProfileResponse struct {
	Profile   BusinessProfile `json:"profile"`
	Found     bool            `json:"found" example:"true"`
	UpdatedAt time.Time       `json:"updated_at" example:"2024-01-01T12:00:00Z"`
}

// ContactStats represents statistics about contacts
type ContactStats struct {
	TotalContacts    int        `json:"total_contacts" example:"150"`
	WhatsAppContacts int        `json:"whatsapp_contacts" example:"120"`
	BusinessContacts int        `json:"business_contacts" example:"10"`
	BlockedContacts  int        `json:"blocked_contacts" example:"2"`
	SyncRate         float64    `json:"sync_rate" example:"0.8"`
	LastSyncAt       *time.Time `json:"last_sync_at,omitempty" example:"2024-01-01T12:00:00Z"`
}

// GetContactStatsRequest represents a request to get contact statistics
type GetContactStatsRequest struct {
	SessionID string `json:"session_id,omitempty"`
}

// GetContactStatsResponse represents the response for contact statistics
type GetContactStatsResponse struct {
	SessionID string       `json:"session_id" example:"session-123"`
	Stats     ContactStats `json:"stats"`
	UpdatedAt time.Time    `json:"updated_at" example:"2024-01-01T12:00:00Z"`
}
