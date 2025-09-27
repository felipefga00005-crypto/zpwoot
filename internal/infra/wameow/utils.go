// Refactored: centralized utilities; improved validation; standardized error handling
package wameow

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waTypes "go.mau.fi/whatsmeow/types"
	"zpwoot/platform/logger"
)

// JIDValidator handles JID validation and normalization
type JIDValidator struct {
	phoneRegex *regexp.Regexp
}

// NewJIDValidator creates a new JID validator
func NewJIDValidator() *JIDValidator {
	return &JIDValidator{
		phoneRegex: regexp.MustCompile(`^\d+$`),
	}
}

// IsValid checks if a JID is valid
func (v *JIDValidator) IsValid(jid string) bool {
	if jid == "" {
		return false
	}

	// Check for WhatsApp JID format
	if strings.Contains(jid, "@s.whatsapp.net") || strings.Contains(jid, "@g.us") {
		return true
	}

	// Check for phone number format (digits only)
	return v.phoneRegex.MatchString(jid)
}

// Normalize converts a JID to standard WhatsApp format
func (v *JIDValidator) Normalize(jid string) string {
	jid = strings.TrimSpace(jid)

	// If it's already a full JID, return as is
	if strings.Contains(jid, "@") {
		return jid
	}

	// If it's just a phone number, add the WhatsApp suffix
	if v.phoneRegex.MatchString(jid) {
		return jid + "@s.whatsapp.net"
	}

	return jid
}

// Parse converts a string JID to types.JID
func (v *JIDValidator) Parse(jid string) (waTypes.JID, error) {
	normalizedJID := v.Normalize(jid)

	if !v.IsValid(normalizedJID) {
		return waTypes.EmptyJID, fmt.Errorf("invalid JID format: %s", jid)
	}

	parsedJID, err := waTypes.ParseJID(normalizedJID)
	if err != nil {
		return waTypes.EmptyJID, fmt.Errorf("failed to parse JID %s: %w", normalizedJID, err)
	}

	return parsedJID, nil
}

// Global validator instance for backward compatibility
var defaultValidator = NewJIDValidator()

// ConnectionError represents connection-related errors
type ConnectionError struct {
	SessionID string
	Operation string
	Err       error
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("connection error for session %s during %s: %v", e.SessionID, e.Operation, e.Err)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}

func newConnectionError(sessionID, operation string, err error) *ConnectionError {
	return &ConnectionError{
		SessionID: sessionID,
		Operation: operation,
		Err:       err,
	}
}

func ValidateClientAndStore(client *whatsmeow.Client, sessionID string) error {
	if client == nil {
		return fmt.Errorf("client is nil for session %s", sessionID)
	}

	if client.Store == nil {
		return fmt.Errorf("client store is nil for session %s", sessionID)
	}

	if client.Store.ID == nil {
		return fmt.Errorf("client store ID is nil for session %s", sessionID)
	}

	return nil
}

// DeviceStoreManager handles device store operations
type DeviceStoreManager struct {
	container *sqlstore.Container
	logger    *logger.Logger
}

// NewDeviceStoreManager creates a new device store manager
func NewDeviceStoreManager(container *sqlstore.Container, logger *logger.Logger) *DeviceStoreManager {
	return &DeviceStoreManager{
		container: container,
		logger:    logger,
	}
}

// GetOrCreateDeviceStore gets an existing device store or creates a new one
func (dsm *DeviceStoreManager) GetOrCreateDeviceStore(sessionID, expectedDeviceJID string) *store.Device {
	if expectedDeviceJID != "" {
		if deviceStore := dsm.getExistingDeviceStore(sessionID, expectedDeviceJID); deviceStore != nil {
			return deviceStore
		}
	}

	return dsm.createNewDeviceStore(sessionID)
}

func (dsm *DeviceStoreManager) getExistingDeviceStore(sessionID, expectedDeviceJID string) *store.Device {
	dsm.logger.InfoWithFields("Loading existing device store", map[string]interface{}{
		"session_id": sessionID,
		"device_jid": expectedDeviceJID,
	})

	jid, err := waTypes.ParseJID(expectedDeviceJID)
	if err != nil {
		dsm.logger.WarnWithFields("Failed to parse expected JID", map[string]interface{}{
			"session_id": sessionID,
			"device_jid": expectedDeviceJID,
			"error":      err.Error(),
		})
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	deviceStore, err := dsm.container.GetDevice(ctx, jid)
	if err != nil {
		dsm.logger.WarnWithFields("Failed to get device store", map[string]interface{}{
			"session_id": sessionID,
			"device_jid": expectedDeviceJID,
			"error":      err.Error(),
		})
		return nil
	}

	if deviceStore != nil {
		dsm.logger.InfoWithFields("Successfully loaded existing device store", map[string]interface{}{
			"session_id": sessionID,
			"device_jid": expectedDeviceJID,
		})
	}

	return deviceStore
}

func (dsm *DeviceStoreManager) createNewDeviceStore(sessionID string) *store.Device {
	dsm.logger.InfoWithFields("Creating new device store", map[string]interface{}{
		"session_id": sessionID,
	})

	deviceStore := dsm.container.NewDevice()
	if deviceStore == nil {
		dsm.logger.ErrorWithFields("Failed to create device store", map[string]interface{}{
			"session_id": sessionID,
		})
		return nil
	}

	dsm.logger.InfoWithFields("Device store ready", map[string]interface{}{
		"session_id": sessionID,
	})

	return deviceStore
}

// GetDeviceStoreForSession maintains backward compatibility
func GetDeviceStoreForSession(sessionID, expectedDeviceJID string, container *sqlstore.Container) *store.Device {
	// Create a temporary logger for backward compatibility
	tempLogger := &logger.Logger{}
	dsm := NewDeviceStoreManager(container, tempLogger)
	return dsm.GetOrCreateDeviceStore(sessionID, expectedDeviceJID)
}

// IsValidJID checks if a JID is valid (backward compatibility)
func IsValidJID(jidStr string) bool {
	return defaultValidator.IsValid(jidStr)
}

// NormalizeJID normalizes a JID (backward compatibility)
func NormalizeJID(jid string) string {
	return defaultValidator.Normalize(jid)
}

// ParseJID parses a JID (backward compatibility)
func ParseJID(jid string) (waTypes.JID, error) {
	return defaultValidator.Parse(jid)
}

func FormatJID(jid waTypes.JID) string {
	if jid.IsEmpty() {
		return ""
	}
	return jid.String()
}

func GetClientInfo(client *whatsmeow.Client) map[string]interface{} {
	if client == nil {
		return map[string]interface{}{
			"client":    "nil",
			"connected": false,
		}
	}

	info := map[string]interface{}{
		"connected": client.IsConnected(),
		"logged_in": client.IsLoggedIn(),
	}

	if client.Store != nil && client.Store.ID != nil {
		info["device_jid"] = FormatJID(*client.Store.ID)
	}

	return info
}

// ValidateSessionID validates a session ID with improved rules
func ValidateSessionID(sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	if len(sessionID) < 3 {
		return fmt.Errorf("session ID too short (min 3 characters): %s", sessionID)
	}

	if len(sessionID) > 100 {
		return fmt.Errorf("session ID too long (max 100 characters): %s", sessionID)
	}

	// Check for valid characters (alphanumeric, underscore, hyphen)
	validSessionRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validSessionRegex.MatchString(sessionID) {
		return fmt.Errorf("session ID contains invalid characters (only alphanumeric, underscore, and hyphen allowed): %s", sessionID)
	}

	return nil
}

// SafeClientOperation executes an operation safely with validation and panic recovery
func SafeClientOperation(client *whatsmeow.Client, sessionID string, operation func() error, logger *logger.Logger) error {
	if err := ValidateClientAndStore(client, sessionID); err != nil {
		return newConnectionError(sessionID, "validate", err)
	}

	if err := ValidateSessionID(sessionID); err != nil {
		return newConnectionError(sessionID, "validate_session", err)
	}

	defer func() {
		if r := recover(); r != nil {
			if logger != nil {
				logger.ErrorWithFields("Panic in client operation", map[string]interface{}{
					"session_id": sessionID,
					"panic":      r,
				})
			}
		}
	}()

	return operation()
}

func GetStoreInfo(deviceStore *store.Device) map[string]interface{} {
	if deviceStore == nil {
		return map[string]interface{}{
			"store": "nil",
		}
	}

	info := map[string]interface{}{
		"exists": true,
	}

	if deviceStore.ID != nil {
		info["device_jid"] = FormatJID(*deviceStore.ID)
	}

	return info
}

type ConnectionStatus struct {
	SessionID string                 `json:"session_id"`
	Connected bool                   `json:"connected"`
	LoggedIn  bool                   `json:"logged_in"`
	DeviceJID string                 `json:"device_jid,omitempty"`
	LastError string                 `json:"last_error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	UpdatedAt int64                  `json:"updated_at"`
}

func GetConnectionStatus(client *whatsmeow.Client, sessionID string) *ConnectionStatus {
	status := &ConnectionStatus{
		SessionID: sessionID,
		UpdatedAt: getCurrentTimestamp(),
		Metadata:  make(map[string]interface{}),
	}

	if client == nil {
		status.LastError = "client is nil"
		return status
	}

	status.Connected = client.IsConnected()
	status.LoggedIn = client.IsLoggedIn()

	if client.Store != nil && client.Store.ID != nil {
		status.DeviceJID = FormatJID(*client.Store.ID)
	}

	status.Metadata = GetClientInfo(client)

	return status
}

func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

func IsRecoverableError(err error) bool {
	if err == nil {
		return false
	}

	return true
}

// GetErrorCategory categorizes errors for better handling
func GetErrorCategory(err error) string {
	if err == nil {
		return "none"
	}

	errStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errStr, "connection"):
		return "connection"
	case strings.Contains(errStr, "auth") || strings.Contains(errStr, "login"):
		return "authentication"
	case strings.Contains(errStr, "timeout"):
		return "timeout"
	case strings.Contains(errStr, "network"):
		return "network"
	case strings.Contains(errStr, "context"):
		return "context"
	default:
		return "unknown"
	}
}
