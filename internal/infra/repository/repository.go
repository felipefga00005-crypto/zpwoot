package repository

import (
	"github.com/jmoiron/sqlx"

	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

// Repositories holds all repository implementations
type Repositories struct {
	Session  ports.SessionRepository
	Webhook  ports.WebhookRepository
	Chatwoot ports.ChatwootRepository
}

// NewRepositories creates all repository implementations
func NewRepositories(db *sqlx.DB, logger *logger.Logger) *Repositories {
	return &Repositories{
		Session:  NewSessionRepository(db, logger),
		Webhook:  NewWebhookRepository(db, logger),
		Chatwoot: NewChatwootRepository(db, logger),
	}
}

// GetSessionRepository returns the session repository
func (r *Repositories) GetSessionRepository() ports.SessionRepository {
	return r.Session
}

// GetWebhookRepository returns the webhook repository
func (r *Repositories) GetWebhookRepository() ports.WebhookRepository {
	return r.Webhook
}

// GetChatwootRepository returns the chatwoot repository
func (r *Repositories) GetChatwootRepository() ports.ChatwootRepository {
	return r.Chatwoot
}
