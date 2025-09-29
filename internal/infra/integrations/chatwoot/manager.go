package chatwoot

import (
	"context"
	"fmt"
	"sync"

	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

// Manager implements the ChatwootManager interface
type Manager struct {
	logger     *logger.Logger
	repository ports.ChatwootRepository
	clients    map[string]*Client
	configs    map[string]*ports.ChatwootConfig
	mu         sync.RWMutex
}

// NewManager creates a new Chatwoot manager
func NewManager(logger *logger.Logger, repository ports.ChatwootRepository) *Manager {
	return &Manager{
		logger:     logger,
		repository: repository,
		clients:    make(map[string]*Client),
		configs:    make(map[string]*ports.ChatwootConfig),
	}
}

// GetClient returns a Chatwoot client for the given session
func (m *Manager) GetClient(sessionID string) (ports.ChatwootClient, error) {
	m.mu.RLock()
	client, exists := m.clients[sessionID]
	m.mu.RUnlock()

	if exists {
		return client, nil
	}

	// Load config and create client
	config, err := m.GetConfig(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get config for session %s: %w", sessionID, err)
	}

	if !config.Enabled {
		return nil, fmt.Errorf("chatwoot integration is disabled for session %s", sessionID)
	}

	client = NewClient(config.URL, config.Token, config.AccountID, m.logger)

	m.mu.Lock()
	m.clients[sessionID] = client
	m.mu.Unlock()

	return client, nil
}

// IsEnabled checks if Chatwoot integration is enabled for a session
func (m *Manager) IsEnabled(sessionID string) bool {
	config, err := m.GetConfig(sessionID)
	if err != nil {
		m.logger.ErrorWithFields("Failed to check if Chatwoot is enabled", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return false
	}

	return config.Enabled
}

// InitInstanceChatwoot initializes Chatwoot integration for a session
func (m *Manager) InitInstanceChatwoot(sessionID, inboxName, webhookURL string, autoCreate bool) error {
	m.logger.InfoWithFields("Initializing Chatwoot instance", map[string]interface{}{
		"session_id":  sessionID,
		"inbox_name":  inboxName,
		"webhook_url": webhookURL,
		"auto_create": autoCreate,
	})

	client, err := m.GetClient(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	if autoCreate {
		// Check if inbox already exists
		inboxes, err := client.ListInboxes()
		if err != nil {
			return fmt.Errorf("failed to list inboxes: %w", err)
		}

		var targetInbox *ports.ChatwootInbox
		for _, inbox := range inboxes {
			if inbox.Name == inboxName {
				targetInbox = &inbox
				break
			}
		}

		// Create inbox if it doesn't exist
		if targetInbox == nil {
			m.logger.InfoWithFields("Creating new inbox", map[string]interface{}{
				"inbox_name": inboxName,
			})

			createdInbox, err := client.CreateInbox(inboxName, webhookURL)
			if err != nil {
				return fmt.Errorf("failed to create inbox: %w", err)
			}
			targetInbox = createdInbox
		}

		// Update config with inbox ID
		config, err := m.GetConfig(sessionID)
		if err != nil {
			return fmt.Errorf("failed to get config: %w", err)
		}

		inboxIDStr := fmt.Sprintf("%d", targetInbox.ID)
		config.InboxID = &inboxIDStr

		err = m.SetConfig(sessionID, config)
		if err != nil {
			return fmt.Errorf("failed to update config with inbox ID: %w", err)
		}

		// Create bot contact if enabled (123456)
		err = m.createBotContact(client, targetInbox.ID)
		if err != nil {
			m.logger.WarnWithFields("Failed to create bot contact", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	return nil
}

// SetConfig sets the Chatwoot configuration for a session
func (m *Manager) SetConfig(sessionID string, config *ports.ChatwootConfig) error {
	m.logger.InfoWithFields("Setting Chatwoot config", map[string]interface{}{
		"session_id": sessionID,
		"config_id":  config.ID.String(),
	})

	// Store in repository
	ctx := context.Background()
	err := m.repository.UpdateConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to update config in repository: %w", err)
	}

	// Update cache
	m.mu.Lock()
	m.configs[sessionID] = config
	// Clear client cache to force recreation with new config
	delete(m.clients, sessionID)
	m.mu.Unlock()

	return nil
}

// GetConfig gets the Chatwoot configuration for a session
func (m *Manager) GetConfig(sessionID string) (*ports.ChatwootConfig, error) {
	m.mu.RLock()
	config, exists := m.configs[sessionID]
	m.mu.RUnlock()

	if exists {
		return config, nil
	}

	// Load from repository
	ctx := context.Background()
	config, err := m.repository.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get config from repository: %w", err)
	}

	// Cache it
	m.mu.Lock()
	m.configs[sessionID] = config
	m.mu.Unlock()

	return config, nil
}

// Cleanup cleans up resources for a session
func (m *Manager) Cleanup(sessionID string) error {
	m.logger.InfoWithFields("Cleaning up Chatwoot resources", map[string]interface{}{
		"session_id": sessionID,
	})

	m.mu.Lock()
	delete(m.clients, sessionID)
	delete(m.configs, sessionID)
	m.mu.Unlock()

	return nil
}

// createBotContact creates a bot contact (123456) if it doesn't exist
func (m *Manager) createBotContact(client ports.ChatwootClient, inboxID int) error {
	botPhone := "123456"
	botName := "Bot"

	// Try to find existing bot contact
	_, err := client.FindContact(botPhone, inboxID)
	if err == nil {
		// Bot contact already exists
		return nil
	}

	// Create bot contact
	m.logger.InfoWithFields("Creating bot contact", map[string]interface{}{
		"phone":    botPhone,
		"name":     botName,
		"inbox_id": inboxID,
	})

	_, err = client.CreateContact(botPhone, botName, inboxID)
	if err != nil {
		return fmt.Errorf("failed to create bot contact: %w", err)
	}

	return nil
}

// GetStats returns statistics for Chatwoot integration
func (m *Manager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"active_clients": len(m.clients),
		"cached_configs": len(m.configs),
	}
}
