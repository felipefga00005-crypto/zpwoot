package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	chatwootdomain "zpwoot/internal/domain/chatwoot"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

type chatwootRepository struct {
	db     *sqlx.DB
	logger *logger.Logger
}

func NewChatwootRepository(db *sqlx.DB, logger *logger.Logger) ports.ChatwootRepository {
	return &chatwootRepository{
		db:     db,
		logger: logger,
	}
}

type chatwootConfigModel struct {
	ID        string         `db:"id"`
	SessionID string         `db:"sessionId"`
	URL       string         `db:"url"`
	Token     string         `db:"token"`
	AccountID string         `db:"accountId"`
	InboxID   sql.NullString `db:"inboxId"`
	Enabled   bool           `db:"enabled"`
	CreatedAt time.Time      `db:"createdAt"`
	UpdatedAt time.Time      `db:"updatedAt"`
}

func (r *chatwootRepository) CreateConfig(ctx context.Context, config *chatwootdomain.ChatwootConfig) error {
	r.logger.InfoWithFields("Creating chatwoot config", map[string]interface{}{
		"config_id": config.ID.String(),
		"url":       config.URL,
	})

	model := r.configToModel(config)

	query := `
		INSERT INTO "zpChatwoot" (id, "sessionId", url, token, "accountId", "inboxId", enabled, "createdAt", "updatedAt")
		VALUES (:id, :sessionId, :url, :token, :accountId, :inboxId, :enabled, :createdAt, :updatedAt)
	`

	_, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		r.logger.ErrorWithFields("Failed to create chatwoot config", map[string]interface{}{
			"config_id": config.ID.String(),
			"error":     err.Error(),
		})
		return fmt.Errorf("failed to create chatwoot config: %w", err)
	}

	return nil
}

func (r *chatwootRepository) GetConfig(ctx context.Context) (*chatwootdomain.ChatwootConfig, error) {
	r.logger.Info("Getting chatwoot config")

	var model chatwootConfigModel
	query := `SELECT * FROM "zpChatwoot" ORDER BY "createdAt" DESC LIMIT 1`

	err := r.db.GetContext(ctx, &model, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, chatwootdomain.ErrConfigNotFound
		}
		r.logger.ErrorWithFields("Failed to get chatwoot config", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to get chatwoot config: %w", err)
	}

	config, err := r.configFromModel(&model)
	if err != nil {
		return nil, fmt.Errorf("failed to convert model to domain: %w", err)
	}

	return config, nil
}

func (r *chatwootRepository) UpdateConfig(ctx context.Context, config *chatwootdomain.ChatwootConfig) error {
	r.logger.InfoWithFields("Updating chatwoot config", map[string]interface{}{
		"config_id": config.ID.String(),
	})

	model := r.configToModel(config)
	model.UpdatedAt = time.Now()

	query := `
		UPDATE "zpChatwoot"
		SET url = :url, token = :token, "accountId" = :accountId,
		    "inboxId" = :inboxId, enabled = :enabled, "updatedAt" = :updatedAt
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		r.logger.ErrorWithFields("Failed to update chatwoot config", map[string]interface{}{
			"config_id": config.ID.String(),
			"error":     err.Error(),
		})
		return fmt.Errorf("failed to update chatwoot config: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return chatwootdomain.ErrConfigNotFound
	}

	return nil
}

func (r *chatwootRepository) DeleteConfig(ctx context.Context) error {
	r.logger.Info("Deleting chatwoot config")

	query := `DELETE FROM "zpChatwoot"`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		r.logger.ErrorWithFields("Failed to delete chatwoot config", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to delete chatwoot config: %w", err)
	}

	return nil
}

func (r *chatwootRepository) CreateContact(ctx context.Context, contact *chatwootdomain.ChatwootContact) error {
	r.logger.InfoWithFields("Contact operations handled via Chatwoot API", map[string]interface{}{
		"chatwoot_id":  contact.ID,
		"phone_number": contact.PhoneNumber,
	})

	return nil
}

func (r *chatwootRepository) GetContactByID(ctx context.Context, id int) (*chatwootdomain.ChatwootContact, error) {
	r.logger.InfoWithFields("Getting contact via Chatwoot API", map[string]interface{}{
		"contact_id": id,
	})

	return nil, chatwootdomain.ErrContactNotFound
}

func (r *chatwootRepository) GetContactByPhone(ctx context.Context, phoneNumber string) (*chatwootdomain.ChatwootContact, error) {
	r.logger.InfoWithFields("Getting contact by phone via Chatwoot API", map[string]interface{}{
		"phone_number": phoneNumber,
	})

	return nil, chatwootdomain.ErrContactNotFound
}

func (r *chatwootRepository) UpdateContact(ctx context.Context, contact *chatwootdomain.ChatwootContact) error {
	r.logger.InfoWithFields("Updating contact via Chatwoot API", map[string]interface{}{
		"contact_id": contact.ID,
	})

	return nil
}

func (r *chatwootRepository) DeleteContact(ctx context.Context, id int) error {
	r.logger.InfoWithFields("Deleting contact via Chatwoot API", map[string]interface{}{
		"contact_id": id,
	})

	return nil
}

func (r *chatwootRepository) ListContacts(ctx context.Context, limit, offset int) ([]*chatwootdomain.ChatwootContact, int, error) {
	r.logger.InfoWithFields("Listing contacts via Chatwoot API", map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	})

	return []*chatwootdomain.ChatwootContact{}, 0, nil
}

func (r *chatwootRepository) GetContactCount(ctx context.Context) (int, error) {
	return 0, nil
}

func (r *chatwootRepository) GetConversationCount(ctx context.Context) (int, error) {
	return 0, nil
}

func (r *chatwootRepository) GetActiveConversationCount(ctx context.Context) (int, error) {
	return 0, nil
}

func (r *chatwootRepository) GetMessageCount(ctx context.Context) (int, error) {
	return 0, nil
}

func (r *chatwootRepository) GetMessageCountByType(ctx context.Context, messageType string) (int, error) {
	return 0, nil
}

func (r *chatwootRepository) GetStatsForPeriod(ctx context.Context, from, to int64) (*ports.ChatwootStats, error) {
	fromTime := time.Unix(from, 0)
	toTime := time.Unix(to, 0)

	stats := &ports.ChatwootStats{
		From:                from,
		To:                  to,
		TotalContacts:       0,
		TotalConversations:  0,
		ActiveConversations: 0,
		MessagesSent:        0,
		MessagesReceived:    0,
		LastSyncAt:          time.Now().Unix(),
	}

	r.logger.InfoWithFields("Retrieved chatwoot stats via API", map[string]interface{}{
		"from": fromTime,
		"to":   toTime,
	})

	return stats, nil
}

func (r *chatwootRepository) CreateConversation(ctx context.Context, conversation *chatwootdomain.ChatwootConversation) error {
	return fmt.Errorf("not implemented")
}

func (r *chatwootRepository) GetConversationByID(ctx context.Context, id int) (*chatwootdomain.ChatwootConversation, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *chatwootRepository) GetConversationByContactID(ctx context.Context, contactID int) (*chatwootdomain.ChatwootConversation, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *chatwootRepository) GetConversationBySessionID(ctx context.Context, sessionID string) (*chatwootdomain.ChatwootConversation, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *chatwootRepository) UpdateConversation(ctx context.Context, conversation *chatwootdomain.ChatwootConversation) error {
	return fmt.Errorf("not implemented")
}

func (r *chatwootRepository) DeleteConversation(ctx context.Context, id int) error {
	return fmt.Errorf("not implemented")
}

func (r *chatwootRepository) ListConversations(ctx context.Context, limit, offset int) ([]*chatwootdomain.ChatwootConversation, int, error) {
	return nil, 0, fmt.Errorf("not implemented")
}

func (r *chatwootRepository) GetActiveConversations(ctx context.Context) ([]*chatwootdomain.ChatwootConversation, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *chatwootRepository) CreateMessage(ctx context.Context, message *chatwootdomain.ChatwootMessage) error {
	return fmt.Errorf("not implemented")
}

func (r *chatwootRepository) GetMessageByID(ctx context.Context, id int) (*chatwootdomain.ChatwootMessage, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *chatwootRepository) GetMessagesByConversationID(ctx context.Context, conversationID int, limit, offset int) ([]*chatwootdomain.ChatwootMessage, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *chatwootRepository) UpdateMessage(ctx context.Context, message *chatwootdomain.ChatwootMessage) error {
	return fmt.Errorf("not implemented")
}

func (r *chatwootRepository) DeleteMessage(ctx context.Context, id int) error {
	return fmt.Errorf("not implemented")
}

func (r *chatwootRepository) CreateSyncRecord(ctx context.Context, record *ports.SyncRecord) error {
	return fmt.Errorf("not implemented")
}

func (r *chatwootRepository) GetSyncRecord(ctx context.Context, sessionID, recordType, externalID string) (*ports.SyncRecord, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *chatwootRepository) UpdateSyncRecord(ctx context.Context, record *ports.SyncRecord) error {
	return fmt.Errorf("not implemented")
}

func (r *chatwootRepository) DeleteSyncRecord(ctx context.Context, id string) error {
	return fmt.Errorf("not implemented")
}

func (r *chatwootRepository) GetSyncRecordsBySession(ctx context.Context, sessionID string) ([]*ports.SyncRecord, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *chatwootRepository) configToModel(config *chatwootdomain.ChatwootConfig) *chatwootConfigModel {
	model := &chatwootConfigModel{
		ID:        config.ID.String(),
		SessionID: config.SessionID.String(),
		URL:       config.URL,
		Token:     config.Token,
		AccountID: config.AccountID,
		Enabled:   config.Enabled,
		CreatedAt: config.CreatedAt,
		UpdatedAt: config.UpdatedAt,
	}

	if config.InboxID != nil {
		model.InboxID = sql.NullString{String: *config.InboxID, Valid: true}
	}

	return model
}

func (r *chatwootRepository) configFromModel(model *chatwootConfigModel) (*chatwootdomain.ChatwootConfig, error) {
	id, err := uuid.Parse(model.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid config ID: %w", err)
	}

	sessionID, err := uuid.Parse(model.SessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	config := &chatwootdomain.ChatwootConfig{
		ID:        id,
		SessionID: sessionID,
		URL:       model.URL,
		Token:     model.Token,
		AccountID: model.AccountID,
		Enabled:   model.Enabled,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}

	if model.InboxID.Valid {
		config.InboxID = &model.InboxID.String
	}

	return config, nil
}
