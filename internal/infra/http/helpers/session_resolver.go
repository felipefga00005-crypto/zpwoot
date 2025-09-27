package helpers

import (
	"context"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"zpwoot/internal/domain/session"
	"zpwoot/platform/logger"
)

// SessionRepository interface for session operations
type SessionRepository interface {
	GetByID(ctx context.Context, id string) (*session.Session, error)
	GetByName(ctx context.Context, name string) (*session.Session, error)
}

// SessionResolver provides utilities for resolving session identifiers
type SessionResolver struct {
	logger      *logger.Logger
	sessionRepo SessionRepository
}

// NewSessionResolver creates a new session resolver
func NewSessionResolver(logger *logger.Logger, sessionRepo SessionRepository) *SessionResolver {
	return &SessionResolver{
		logger:      logger,
		sessionRepo: sessionRepo,
	}
}

// ResolveSessionIdentifier determines if the provided identifier is a UUID or a name
// and returns the appropriate identifier type and value
func (sr *SessionResolver) ResolveSessionIdentifier(idOrName string) (identifierType string, value string, isValid bool) {
	// Clean the input
	idOrName = strings.TrimSpace(idOrName)

	if idOrName == "" {
		sr.logger.Warn("Empty session identifier provided")
		return "", "", false
	}

	// Check if it's a valid UUID
	if sr.isValidUUID(idOrName) {
		sr.logger.DebugWithFields("Resolved as UUID", map[string]interface{}{
			"identifier": idOrName,
			"type":       "uuid",
		})
		return "uuid", idOrName, true
	}

	// Check if it's a valid session name (URL-safe)
	if sr.isValidSessionName(idOrName) {
		sr.logger.DebugWithFields("Resolved as session name", map[string]interface{}{
			"identifier": idOrName,
			"type":       "name",
		})
		return "name", idOrName, true
	}

	sr.logger.WarnWithFields("Invalid session identifier", map[string]interface{}{
		"identifier": idOrName,
		"reason":     "not a valid UUID or session name",
	})
	return "", "", false
}

// isValidUUID checks if the string is a valid UUID
func (sr *SessionResolver) isValidUUID(str string) bool {
	_, err := uuid.Parse(str)
	return err == nil
}

// isValidSessionName checks if the string is a valid session name for URL usage
// Session names should be URL-safe: alphanumeric, hyphens, underscores, dots
// Length: 1-100 characters
func (sr *SessionResolver) isValidSessionName(name string) bool {
	// Check length
	if len(name) < 1 || len(name) > 100 {
		return false
	}

	// Check if it contains only URL-safe characters
	// Allow: letters, numbers, hyphens, underscores, dots
	urlSafePattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !urlSafePattern.MatchString(name) {
		return false
	}

	// Don't allow names that look like UUIDs to avoid confusion
	if sr.looksLikeUUID(name) {
		return false
	}

	// Don't allow reserved names
	reservedNames := []string{
		"create", "list", "info", "delete", "connect", "logout",
		"qr", "pair", "proxy", "webhook", "chatwoot", "health",
		"swagger", "api", "admin", "config", "status", "test",
	}

	lowerName := strings.ToLower(name)
	for _, reserved := range reservedNames {
		if lowerName == reserved {
			return false
		}
	}

	return true
}

// looksLikeUUID checks if a string looks like a UUID (to avoid confusion)
func (sr *SessionResolver) looksLikeUUID(str string) bool {
	// UUID pattern: 8-4-4-4-12 hexadecimal characters
	uuidPattern := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	return uuidPattern.MatchString(str)
}

// ValidateSessionName validates a session name for creation
// This is more strict than URL validation and should be used when creating sessions
func (sr *SessionResolver) ValidateSessionName(name string) (bool, string) {
	name = strings.TrimSpace(name)

	if name == "" {
		return false, "Session name cannot be empty"
	}

	if len(name) < 3 {
		return false, "Session name must be at least 3 characters long"
	}

	if len(name) > 50 {
		return false, "Session name must be at most 50 characters long"
	}

	// More restrictive pattern for creation: start with letter, then letters/numbers/hyphens/underscores
	creationPattern := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
	if !creationPattern.MatchString(name) {
		return false, "Session name must start with a letter and contain only letters, numbers, hyphens, and underscores"
	}

	// Check reserved names
	reservedNames := []string{
		"create", "list", "info", "delete", "connect", "logout",
		"qr", "pair", "proxy", "webhook", "chatwoot", "health",
		"swagger", "api", "admin", "config", "status", "test",
		"new", "add", "remove", "update", "edit", "view", "show",
	}

	lowerName := strings.ToLower(name)
	for _, reserved := range reservedNames {
		if lowerName == reserved {
			return false, "Session name '" + name + "' is reserved and cannot be used"
		}
	}

	return true, ""
}

// SuggestValidName suggests a valid session name based on input
func (sr *SessionResolver) SuggestValidName(input string) string {
	// Clean the input
	input = strings.TrimSpace(input)
	input = strings.ToLower(input)

	// Replace invalid characters with hyphens
	validPattern := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	suggested := validPattern.ReplaceAllString(input, "-")

	// Remove multiple consecutive hyphens
	multiHyphen := regexp.MustCompile(`-+`)
	suggested = multiHyphen.ReplaceAllString(suggested, "-")

	// Ensure it starts with a letter
	if len(suggested) > 0 && !regexp.MustCompile(`^[a-zA-Z]`).MatchString(suggested) {
		suggested = "session-" + suggested
	}

	// Trim hyphens from start and end
	suggested = strings.Trim(suggested, "-")

	// Ensure minimum length
	if len(suggested) < 3 {
		suggested = "session-" + suggested
	}

	// Ensure maximum length
	if len(suggested) > 50 {
		suggested = suggested[:50]
		suggested = strings.TrimRight(suggested, "-")
	}

	return suggested
}

// ResolveSession resolves a session by identifier (UUID or name) and returns the actual session
func (sr *SessionResolver) ResolveSession(ctx context.Context, idOrName string) (*session.Session, error) {
	// First, resolve the identifier type
	identifierType, value, isValid := sr.ResolveSessionIdentifier(idOrName)
	if !isValid {
		return nil, session.ErrSessionNotFound
	}

	sr.logger.InfoWithFields("Resolving session", map[string]interface{}{
		"identifier":      value,
		"identifier_type": identifierType,
	})

	// Get session based on identifier type
	var sess *session.Session
	var err error

	switch identifierType {
	case "uuid":
		sess, err = sr.sessionRepo.GetByID(ctx, value)
	case "name":
		sess, err = sr.sessionRepo.GetByName(ctx, value)
	default:
		return nil, session.ErrSessionNotFound
	}

	if err != nil {
		sr.logger.ErrorWithFields("Failed to resolve session", map[string]interface{}{
			"identifier":      value,
			"identifier_type": identifierType,
			"error":           err.Error(),
		})
		return nil, err
	}

	sr.logger.InfoWithFields("Session resolved successfully", map[string]interface{}{
		"identifier":      value,
		"identifier_type": identifierType,
		"session_id":      sess.ID.String(),
		"session_name":    sess.Name,
	})

	return sess, nil
}
