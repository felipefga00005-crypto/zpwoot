package contact

import (
	"context"
	"fmt"
	"strings"
	"time"

	"zpwoot/platform/logger"
)

// Service defines the interface for contact domain service
type Service interface {
	CheckWhatsApp(ctx context.Context, req *CheckWhatsAppRequest) (*CheckWhatsAppResponse, error)
	GetProfilePicture(ctx context.Context, req *GetProfilePictureRequest) (*ProfilePictureInfo, error)
	GetUserInfo(ctx context.Context, req *GetUserInfoRequest) (*GetUserInfoResponse, error)
	ListContacts(ctx context.Context, req *ListContactsRequest) (*ListContactsResponse, error)
	SyncContacts(ctx context.Context, req *SyncContactsRequest) (*SyncContactsResponse, error)
	GetBusinessProfile(ctx context.Context, req *GetBusinessProfileRequest) (*GetBusinessProfileResponse, error)
	GetContactStats(ctx context.Context, req *GetContactStatsRequest) (*GetContactStatsResponse, error)
}

// WameowManager defines the interface for multi-session WhatsApp operations
type WameowManager interface {
	IsOnWhatsApp(ctx context.Context, sessionID string, phoneNumbers []string) (map[string]interface{}, error)
	GetProfilePictureInfo(ctx context.Context, sessionID, jid string, preview bool) (map[string]interface{}, error)
	GetUserInfo(ctx context.Context, sessionID string, jids []string) ([]map[string]interface{}, error)
	GetBusinessProfile(ctx context.Context, sessionID, jid string) (map[string]interface{}, error)
	GetAllContacts(ctx context.Context, sessionID string) (map[string]interface{}, error)
}

type service struct {
	wameowManager WameowManager
	logger        *logger.Logger
}

// NewService creates a new contact service
func NewService(wameowManager WameowManager, logger *logger.Logger) Service {
	return &service{
		wameowManager: wameowManager,
		logger:        logger,
	}
}

// CheckWhatsApp checks if phone numbers are registered on WhatsApp
func (s *service) CheckWhatsApp(ctx context.Context, req *CheckWhatsAppRequest) (*CheckWhatsAppResponse, error) {
	if err := s.validateCheckWhatsAppRequest(req); err != nil {
		return nil, err
	}

	s.logger.InfoWithFields("Checking WhatsApp numbers", map[string]interface{}{
		"session_id":  req.SessionID,
		"phone_count": len(req.PhoneNumbers),
	})

	// Check with WhatsApp using real whatsmeow method
	statusMap, err := s.wameowManager.IsOnWhatsApp(ctx, req.SessionID, req.PhoneNumbers)
	if err != nil {
		s.logger.ErrorWithFields("Failed to check WhatsApp numbers", map[string]interface{}{
			"session_id": req.SessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to check WhatsApp numbers: %w", err)
	}

	// Convert map to slice
	results := make([]WhatsAppStatus, 0, len(req.PhoneNumbers))
	checked := 0
	for _, phoneNumber := range req.PhoneNumbers {
		if statusData, exists := statusMap[phoneNumber]; exists {
			if statusMap, ok := statusData.(map[string]interface{}); ok {
				status := WhatsAppStatus{
					PhoneNumber:  getStringFromMap(statusMap, "phone_number"),
					IsOnWhatsApp: getBoolFromMap(statusMap, "is_on_whatsapp"),
					JID:          getStringFromMap(statusMap, "jid"),
					IsBusiness:   getBoolFromMap(statusMap, "is_business"),
					VerifiedName: getStringFromMap(statusMap, "verified_name"),
				}
				results = append(results, status)
				checked++
			}
		} else {
			// Add default status for numbers that couldn't be checked
			results = append(results, WhatsAppStatus{
				PhoneNumber:  phoneNumber,
				IsOnWhatsApp: false,
			})
		}
	}

	return &CheckWhatsAppResponse{
		Results: results,
		Total:   len(req.PhoneNumbers),
		Checked: checked,
	}, nil
}

// GetProfilePicture gets profile picture information for a contact
func (s *service) GetProfilePicture(ctx context.Context, req *GetProfilePictureRequest) (*ProfilePictureInfo, error) {
	if err := s.validateGetProfilePictureRequest(req); err != nil {
		return nil, err
	}

	s.logger.InfoWithFields("Getting profile picture", map[string]interface{}{
		"session_id": req.SessionID,
		"jid":        req.JID,
		"preview":    req.Preview,
	})

	profileData, err := s.wameowManager.GetProfilePictureInfo(ctx, req.SessionID, req.JID, req.Preview)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get profile picture", map[string]interface{}{
			"session_id": req.SessionID,
			"jid":        req.JID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get profile picture: %w", err)
	}

	// Convert map to ProfilePictureInfo
	profilePicture := &ProfilePictureInfo{
		JID:        getStringFromMap(profileData, "jid"),
		URL:        getStringFromMap(profileData, "url"),
		ID:         getStringFromMap(profileData, "id"),
		Type:       getStringFromMap(profileData, "type"),
		DirectPath: getStringFromMap(profileData, "direct_path"),
		UpdatedAt:  time.Now(),
		HasPicture: getBoolFromMap(profileData, "has_picture"),
	}

	return profilePicture, nil
}

// GetUserInfo gets detailed information about WhatsApp users
func (s *service) GetUserInfo(ctx context.Context, req *GetUserInfoRequest) (*GetUserInfoResponse, error) {
	if err := s.validateGetUserInfoRequest(req); err != nil {
		return nil, err
	}

	s.logger.InfoWithFields("Getting user info", map[string]interface{}{
		"session_id": req.SessionID,
		"jid_count":  len(req.JIDs),
	})

	usersData, err := s.wameowManager.GetUserInfo(ctx, req.SessionID, req.JIDs)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get user info", map[string]interface{}{
			"session_id": req.SessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Convert slice of maps to slice of UserInfo
	users := make([]UserInfo, 0, len(usersData))
	for _, userData := range usersData {
		user := UserInfo{
			JID:          getStringFromMap(userData, "jid"),
			PhoneNumber:  getStringFromMap(userData, "phone_number"),
			Name:         getStringFromMap(userData, "name"),
			Status:       getStringFromMap(userData, "status"),
			PictureID:    getStringFromMap(userData, "picture_id"),
			IsBusiness:   getBoolFromMap(userData, "is_business"),
			VerifiedName: getStringFromMap(userData, "verified_name"),
			IsContact:    getBoolFromMap(userData, "is_contact"),
			LastSeen:     getTimeFromMap(userData, "last_seen"),
			IsOnline:     getBoolFromMap(userData, "is_online"),
		}
		users = append(users, user)
	}

	return &GetUserInfoResponse{
		Users: users,
		Total: len(req.JIDs),
		Found: len(users),
	}, nil
}

// ListContacts lists contacts from the WhatsApp account
func (s *service) ListContacts(ctx context.Context, req *ListContactsRequest) (*ListContactsResponse, error) {
	if err := s.validateListContactsRequest(req); err != nil {
		return nil, err
	}

	s.logger.InfoWithFields("Listing contacts", map[string]interface{}{
		"session_id": req.SessionID,
		"limit":      req.Limit,
		"offset":     req.Offset,
		"search":     req.Search,
	})

	// Get all contacts from whatsmeow store
	contactsData, err := s.wameowManager.GetAllContacts(ctx, req.SessionID)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get contacts", map[string]interface{}{
			"session_id": req.SessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get contacts: %w", err)
	}

	// Extract contacts from response
	contactsInterface, exists := contactsData["contacts"]
	if !exists {
		return &ListContactsResponse{
			Contacts: []Contact{},
			Total:    0,
			Limit:    req.Limit,
			Offset:   req.Offset,
			HasMore:  false,
		}, nil
	}

	contactsList, ok := contactsInterface.([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid contacts data format")
	}

	// Convert to domain objects and apply filtering
	allContacts := make([]Contact, 0, len(contactsList))
	for _, contactData := range contactsList {
		// Handle time fields that might be nil
		var addedAt, updatedAt time.Time
		if addedAtPtr := getTimeFromMap(contactData, "addedAt"); addedAtPtr != nil {
			addedAt = *addedAtPtr
		}
		if updatedAtPtr := getTimeFromMap(contactData, "updatedAt"); updatedAtPtr != nil {
			updatedAt = *updatedAtPtr
		}

		contact := Contact{
			JID:         getStringFromMap(contactData, "jid"),
			PhoneNumber: getStringFromMap(contactData, "phoneNumber"),
			Name:        getStringFromMap(contactData, "name"),
			ShortName:   getStringFromMap(contactData, "shortName"),
			PushName:    getStringFromMap(contactData, "pushName"),
			IsBusiness:  getBoolFromMap(contactData, "isBusiness"),
			IsContact:   getBoolFromMap(contactData, "isContact"),
			IsBlocked:   getBoolFromMap(contactData, "isBlocked"),
			AddedAt:     addedAt,
			UpdatedAt:   updatedAt,
		}

		// Apply search filter if provided
		if req.Search != "" {
			searchLower := strings.ToLower(req.Search)
			if !strings.Contains(strings.ToLower(contact.Name), searchLower) &&
				!strings.Contains(strings.ToLower(contact.ShortName), searchLower) &&
				!strings.Contains(strings.ToLower(contact.PushName), searchLower) &&
				!strings.Contains(contact.PhoneNumber, req.Search) {
				continue
			}
		}

		allContacts = append(allContacts, contact)
	}

	// Apply pagination
	total := len(allContacts)
	start := req.Offset
	end := start + req.Limit

	if start >= total {
		return &ListContactsResponse{
			Contacts: []Contact{},
			Total:    total,
			Limit:    req.Limit,
			Offset:   req.Offset,
			HasMore:  false,
		}, nil
	}

	if end > total {
		end = total
	}

	paginatedContacts := allContacts[start:end]
	hasMore := end < total

	return &ListContactsResponse{
		Contacts: paginatedContacts,
		Total:    total,
		Limit:    req.Limit,
		Offset:   req.Offset,
		HasMore:  hasMore,
	}, nil
}

// SyncContacts synchronizes contacts from the device with WhatsApp
// Note: whatsmeow doesn't have a native SyncContacts method
// Contact sync happens automatically via app state
func (s *service) SyncContacts(ctx context.Context, req *SyncContactsRequest) (*SyncContactsResponse, error) {
	if err := s.validateSyncContactsRequest(req); err != nil {
		return nil, err
	}

	s.logger.WarnWithFields("SyncContacts not supported by whatsmeow", map[string]interface{}{
		"session_id": req.SessionID,
		"method":     "SyncContacts",
	})

	// Return placeholder response since whatsmeow doesn't support this natively
	return &SyncContactsResponse{
		Synced:   0,
		Added:    0,
		Updated:  0,
		Removed:  0,
		Total:    0,
		SyncedAt: time.Now(),
	}, fmt.Errorf("SyncContacts not supported by whatsmeow - contacts sync automatically via app state")
}

// GetBusinessProfile gets business profile information
func (s *service) GetBusinessProfile(ctx context.Context, req *GetBusinessProfileRequest) (*GetBusinessProfileResponse, error) {
	if err := s.validateGetBusinessProfileRequest(req); err != nil {
		return nil, err
	}

	s.logger.InfoWithFields("Getting business profile", map[string]interface{}{
		"session_id": req.SessionID,
		"jid":        req.JID,
	})

	profileData, err := s.wameowManager.GetBusinessProfile(ctx, req.SessionID, req.JID)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get business profile", map[string]interface{}{
			"session_id": req.SessionID,
			"jid":        req.JID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get business profile: %w", err)
	}

	// Convert map to BusinessProfile
	profile := BusinessProfile{
		JID:         getStringFromMap(profileData, "jid"),
		Name:        getStringFromMap(profileData, "name"),
		Category:    getStringFromMap(profileData, "category"),
		Description: getStringFromMap(profileData, "description"),
		Website:     getStringFromMap(profileData, "website"),
		Email:       getStringFromMap(profileData, "email"),
		Address:     getStringFromMap(profileData, "address"),
		Verified:    getBoolFromMap(profileData, "verified"),
	}

	return &GetBusinessProfileResponse{
		Profile:   profile,
		Found:     true,
		UpdatedAt: time.Now(),
	}, nil
}

// GetContactStats gets statistics about contacts
// Note: whatsmeow doesn't provide contact statistics natively
func (s *service) GetContactStats(ctx context.Context, req *GetContactStatsRequest) (*GetContactStatsResponse, error) {
	if err := s.validateGetContactStatsRequest(req); err != nil {
		return nil, err
	}

	s.logger.WarnWithFields("GetContactStats not supported by whatsmeow", map[string]interface{}{
		"session_id": req.SessionID,
		"method":     "GetContactStats",
	})

	// Return placeholder stats since whatsmeow doesn't support this natively
	return &GetContactStatsResponse{
		SessionID: req.SessionID,
		Stats: ContactStats{
			TotalContacts:    0,
			WhatsAppContacts: 0,
			BusinessContacts: 0,
			BlockedContacts:  0,
			SyncRate:         0.0,
			LastSyncAt:       nil,
		},
		UpdatedAt: time.Now(),
	}, fmt.Errorf("GetContactStats not supported by whatsmeow - contact stats not available")
}

// Validation methods
func (s *service) validateCheckWhatsAppRequest(req *CheckWhatsAppRequest) error {
	if req.SessionID == "" {
		return ErrInvalidSessionID
	}
	if len(req.PhoneNumbers) == 0 {
		return fmt.Errorf("at least one phone number is required")
	}
	if len(req.PhoneNumbers) > 50 {
		return fmt.Errorf("maximum 50 phone numbers allowed")
	}
	return nil
}

func (s *service) validateGetProfilePictureRequest(req *GetProfilePictureRequest) error {
	if req.SessionID == "" {
		return ErrInvalidSessionID
	}
	if req.JID == "" {
		return ErrInvalidJID
	}
	return nil
}

func (s *service) validateGetUserInfoRequest(req *GetUserInfoRequest) error {
	if req.SessionID == "" {
		return ErrInvalidSessionID
	}
	if len(req.JIDs) == 0 {
		return fmt.Errorf("at least one JID is required")
	}
	if len(req.JIDs) > 20 {
		return fmt.Errorf("maximum 20 JIDs allowed")
	}
	return nil
}

func (s *service) validateListContactsRequest(req *ListContactsRequest) error {
	if req.SessionID == "" {
		return ErrInvalidSessionID
	}
	if req.Limit < 0 || req.Limit > 100 {
		return ErrInvalidLimit
	}
	if req.Offset < 0 {
		return ErrInvalidOffset
	}
	return nil
}

func (s *service) validateSyncContactsRequest(req *SyncContactsRequest) error {
	if req.SessionID == "" {
		return ErrInvalidSessionID
	}
	return nil
}

func (s *service) validateGetBusinessProfileRequest(req *GetBusinessProfileRequest) error {
	if req.SessionID == "" {
		return ErrInvalidSessionID
	}
	if req.JID == "" {
		return ErrInvalidJID
	}
	return nil
}

func (s *service) validateGetContactStatsRequest(req *GetContactStatsRequest) error {
	if req.SessionID == "" {
		return ErrInvalidSessionID
	}
	return nil
}

// Helper functions to extract values from map[string]interface{}
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getBoolFromMap(m map[string]interface{}, key string) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func getTimeFromMap(m map[string]interface{}, key string) *time.Time {
	if val, ok := m[key]; ok {
		if t, ok := val.(time.Time); ok {
			return &t
		}
		if t, ok := val.(*time.Time); ok {
			return t
		}
	}
	return nil
}
