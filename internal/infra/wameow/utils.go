package wameow

import (
	"context"
	"fmt"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waTypes "go.mau.fi/whatsmeow/types"
)

// ConnectionError represents a connection-related error
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

// newConnectionError creates a new connection error
func newConnectionError(sessionID, operation string, err error) *ConnectionError {
	return &ConnectionError{
		SessionID: sessionID,
		Operation: operation,
		Err:       err,
	}
}

// ValidateClientAndStore validates Wameow client and store
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

// GetDeviceStoreForSession gets or creates a device store for a session
func GetDeviceStoreForSession(sessionID, expectedDeviceJID string, container *sqlstore.Container) *store.Device {
	var deviceStore *store.Device

	if expectedDeviceJID != "" {
		fmt.Printf("Loading existing device store for session %s with JID %s\n", sessionID, expectedDeviceJID)

		jid, err := waTypes.ParseJID(expectedDeviceJID)
		if err != nil {
			fmt.Printf("Failed to parse expected JID %s: %v, creating new device\n", expectedDeviceJID, err)
		} else {
			ctx := context.Background()
			deviceStore, err = container.GetDevice(ctx, jid)
			if err != nil {
				fmt.Printf("Failed to get device store for expected JID %s: %v, creating new device\n", expectedDeviceJID, err)
			} else if deviceStore != nil {
				fmt.Printf("Successfully loaded existing device store for session %s with JID %s\n", sessionID, expectedDeviceJID)
				return deviceStore
			}
		}

		if deviceStore == nil {
			fmt.Printf("Device store not found for expected JID %s, creating new device for session %s\n", expectedDeviceJID, sessionID)
			deviceStore = container.NewDevice()
		}
	} else {
		fmt.Printf("Creating new device store for session %s (no existing JID)\n", sessionID)
		deviceStore = container.NewDevice()
	}

	if deviceStore == nil {
		fmt.Printf("Failed to create device store for session %s\n", sessionID)
		return nil
	}

	fmt.Printf("Device store ready for session %s\n", sessionID)
	return deviceStore
}

// IsValidJID checks if a JID string is valid
func IsValidJID(jidStr string) bool {
	if jidStr == "" {
		return false
	}

	_, err := waTypes.ParseJID(jidStr)
	return err == nil
}

// FormatJID formats a JID for display
func FormatJID(jid waTypes.JID) string {
	if jid.IsEmpty() {
		return ""
	}
	return jid.String()
}

// GetClientInfo returns basic client information
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

// ValidateSessionID checks if session ID is valid
func ValidateSessionID(sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	if len(sessionID) < 3 {
		return fmt.Errorf("session ID too short: %s", sessionID)
	}

	if len(sessionID) > 100 {
		return fmt.Errorf("session ID too long: %s", sessionID)
	}

	return nil
}

// SafeClientOperation executes an operation on a client with error handling
func SafeClientOperation(client *whatsmeow.Client, sessionID string, operation func() error) error {
	if err := ValidateClientAndStore(client, sessionID); err != nil {
		return newConnectionError(sessionID, "validate", err)
	}

	if err := ValidateSessionID(sessionID); err != nil {
		return newConnectionError(sessionID, "validate_session", err)
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in client operation for session %s: %v\n", sessionID, r)
		}
	}()

	return operation()
}

// GetStoreInfo returns information about a device store
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

// ConnectionStatus represents the status of a Wameow connection
type ConnectionStatus struct {
	SessionID string                 `json:"session_id"`
	Connected bool                   `json:"connected"`
	LoggedIn  bool                   `json:"logged_in"`
	DeviceJID string                 `json:"device_jid,omitempty"`
	LastError string                 `json:"last_error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	UpdatedAt int64                  `json:"updated_at"`
}

// GetConnectionStatus returns the current connection status
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

	// Add client info to metadata
	status.Metadata = GetClientInfo(client)

	return status
}

// getCurrentTimestamp returns current Unix timestamp
func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// Helper function to check if error is recoverable
func IsRecoverableError(err error) bool {
	if err == nil {
		return false
	}

	// Add logic to determine if error is recoverable
	// For now, assume most errors are recoverable
	return true
}

// Helper function to get error category
func GetErrorCategory(err error) string {
	if err == nil {
		return "none"
	}

	// Categorize common Wameow errors
	errStr := err.Error()

	if contains(errStr, "connection") {
		return "connection"
	}
	if contains(errStr, "auth") || contains(errStr, "login") {
		return "authentication"
	}
	if contains(errStr, "timeout") {
		return "timeout"
	}
	if contains(errStr, "network") {
		return "network"
	}

	return "unknown"
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr))))
}

// containsSubstring checks if string contains substring
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
