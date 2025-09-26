package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"zpwoot/internal/domain/webhook"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

// webhookRepository implements the WebhookRepository interface
type webhookRepository struct {
	db     *sqlx.DB
	logger *logger.Logger
}

// NewWebhookRepository creates a new webhook repository
func NewWebhookRepository(db *sqlx.DB, logger *logger.Logger) ports.WebhookRepository {
	return &webhookRepository{
		db:     db,
		logger: logger,
	}
}

// webhookModel represents the database model for webhooks
type webhookModel struct {
	ID        string         `db:"id"`
	SessionID sql.NullString `db:"sessionId"`
	URL       string         `db:"url"`
	Secret    sql.NullString `db:"secret"`
	Events    string         `db:"events"` // JSONB field
	Active    bool           `db:"active"`
	CreatedAt time.Time      `db:"createdAt"`
	UpdatedAt time.Time      `db:"updatedAt"`
}

// Create creates a new webhook configuration
func (r *webhookRepository) Create(ctx context.Context, wh *webhook.WebhookConfig) error {
	r.logger.InfoWithFields("Creating webhook", map[string]interface{}{
		"webhook_id": wh.ID.String(),
		"url":        wh.URL,
		"session_id": wh.SessionID,
	})

	model := r.toModel(wh)

	query := `
		INSERT INTO "zpWebhooks" (id, "sessionId", url, secret, events, active, "createdAt", "updatedAt")
		VALUES (:id, :sessionId, :url, :secret, :events, :active, :createdAt, :updatedAt)
	`

	_, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		r.logger.ErrorWithFields("Failed to create webhook", map[string]interface{}{
			"webhook_id": wh.ID.String(),
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to create webhook: %w", err)
	}

	r.logger.InfoWithFields("Webhook created successfully", map[string]interface{}{
		"webhook_id": wh.ID.String(),
	})

	return nil
}

// GetByID retrieves a webhook by its ID
func (r *webhookRepository) GetByID(ctx context.Context, id string) (*webhook.WebhookConfig, error) {
	r.logger.InfoWithFields("Getting webhook by ID", map[string]interface{}{
		"webhook_id": id,
	})

	var model webhookModel
	query := `SELECT * FROM "zpWebhooks" WHERE id = $1`

	err := r.db.GetContext(ctx, &model, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, webhook.ErrWebhookNotFound
		}
		r.logger.ErrorWithFields("Failed to get webhook by ID", map[string]interface{}{
			"webhook_id": id,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}

	wh, err := r.fromModel(&model)
	if err != nil {
		return nil, fmt.Errorf("failed to convert model to domain: %w", err)
	}

	return wh, nil
}

// GetBySessionID retrieves webhooks for a specific session
func (r *webhookRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*webhook.WebhookConfig, error) {
	r.logger.InfoWithFields("Getting webhooks by session ID", map[string]interface{}{
		"session_id": sessionID,
	})

	var models []webhookModel
	query := `SELECT * FROM "zpWebhooks" WHERE "sessionId" = $1 ORDER BY "createdAt" DESC`

	err := r.db.SelectContext(ctx, &models, query, sessionID)
	if err != nil {
		r.logger.ErrorWithFields("Failed to get webhooks by session ID", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get webhooks: %w", err)
	}

	webhooks := make([]*webhook.WebhookConfig, len(models))
	for i, model := range models {
		wh, err := r.fromModel(&model)
		if err != nil {
			r.logger.ErrorWithFields("Failed to convert webhook model", map[string]interface{}{
				"webhook_id": model.ID,
				"error":      err.Error(),
			})
			continue
		}
		webhooks[i] = wh
	}

	return webhooks, nil
}

// GetGlobalWebhooks retrieves global webhooks (no session ID)
func (r *webhookRepository) GetGlobalWebhooks(ctx context.Context) ([]*webhook.WebhookConfig, error) {
	r.logger.Info("Getting global webhooks")

	var models []webhookModel
	query := `SELECT * FROM "zpWebhooks" WHERE "sessionId" IS NULL ORDER BY "createdAt" DESC`

	err := r.db.SelectContext(ctx, &models, query)
	if err != nil {
		r.logger.ErrorWithFields("Failed to get global webhooks", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to get global webhooks: %w", err)
	}

	webhooks := make([]*webhook.WebhookConfig, len(models))
	for i, model := range models {
		wh, err := r.fromModel(&model)
		if err != nil {
			r.logger.ErrorWithFields("Failed to convert webhook model", map[string]interface{}{
				"webhook_id": model.ID,
				"error":      err.Error(),
			})
			continue
		}
		webhooks[i] = wh
	}

	return webhooks, nil
}

// List retrieves webhooks with optional filters
func (r *webhookRepository) List(ctx context.Context, req *webhook.ListWebhooksRequest) ([]*webhook.WebhookConfig, int, error) {
	r.logger.InfoWithFields("Listing webhooks", map[string]interface{}{
		"session_id": req.SessionID,
		"active":     req.Active,
		"limit":      req.Limit,
		"offset":     req.Offset,
	})

	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if req.SessionID != nil {
		whereClause += fmt.Sprintf(" AND \"sessionId\" = $%d", argIndex)
		args = append(args, *req.SessionID)
		argIndex++
	}

	if req.Active != nil {
		whereClause += fmt.Sprintf(" AND active = $%d", argIndex)
		args = append(args, *req.Active)
		argIndex++
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM \"zpWebhooks\" %s", whereClause)
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		r.logger.ErrorWithFields("Failed to count webhooks", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, 0, fmt.Errorf("failed to count webhooks: %w", err)
	}

	// Get webhooks with pagination
	query := fmt.Sprintf(`
		SELECT * FROM "zpWebhooks" %s
		ORDER BY "createdAt" DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, req.Limit, req.Offset)

	var models []webhookModel
	err = r.db.SelectContext(ctx, &models, query, args...)
	if err != nil {
		r.logger.ErrorWithFields("Failed to list webhooks", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, 0, fmt.Errorf("failed to list webhooks: %w", err)
	}

	// Convert models to domain entities
	webhooks := make([]*webhook.WebhookConfig, len(models))
	for i, model := range models {
		wh, err := r.fromModel(&model)
		if err != nil {
			r.logger.ErrorWithFields("Failed to convert webhook model", map[string]interface{}{
				"webhook_id": model.ID,
				"error":      err.Error(),
			})
			continue
		}
		webhooks[i] = wh
	}

	return webhooks, total, nil
}

// Update updates an existing webhook configuration
func (r *webhookRepository) Update(ctx context.Context, wh *webhook.WebhookConfig) error {
	r.logger.InfoWithFields("Updating webhook", map[string]interface{}{
		"webhook_id": wh.ID.String(),
	})

	model := r.toModel(wh)
	model.UpdatedAt = time.Now()

	query := `
		UPDATE "zpWebhooks"
		SET "sessionId" = :sessionId, url = :url, secret = :secret,
		    events = :events, active = :active, "updatedAt" = :updatedAt
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		r.logger.ErrorWithFields("Failed to update webhook", map[string]interface{}{
			"webhook_id": wh.ID.String(),
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to update webhook: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return webhook.ErrWebhookNotFound
	}

	return nil
}

// Delete removes a webhook by ID
func (r *webhookRepository) Delete(ctx context.Context, id string) error {
	r.logger.InfoWithFields("Deleting webhook", map[string]interface{}{
		"webhook_id": id,
	})

	query := `DELETE FROM "zpWebhooks" WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.ErrorWithFields("Failed to delete webhook", map[string]interface{}{
			"webhook_id": id,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return webhook.ErrWebhookNotFound
	}

	return nil
}

// UpdateStatus updates only the active status of a webhook
func (r *webhookRepository) UpdateStatus(ctx context.Context, id string, active bool) error {
	r.logger.InfoWithFields("Updating webhook status", map[string]interface{}{
		"webhook_id": id,
		"active":     active,
	})

	query := `UPDATE "zpWebhooks" SET active = $1, "updatedAt" = $2 WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, active, time.Now(), id)
	if err != nil {
		r.logger.ErrorWithFields("Failed to update webhook status", map[string]interface{}{
			"webhook_id": id,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to update webhook status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return webhook.ErrWebhookNotFound
	}

	return nil
}

// GetActiveWebhooks retrieves all active webhooks
func (r *webhookRepository) GetActiveWebhooks(ctx context.Context) ([]*webhook.WebhookConfig, error) {
	r.logger.Info("Getting active webhooks")

	query := `SELECT * FROM "zpWebhooks" WHERE active = true ORDER BY "createdAt" DESC`

	var models []webhookModel
	err := r.db.SelectContext(ctx, &models, query)
	if err != nil {
		r.logger.ErrorWithFields("Failed to get active webhooks", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to get active webhooks: %w", err)
	}

	webhooks := make([]*webhook.WebhookConfig, len(models))
	for i, model := range models {
		wh, err := r.fromModel(&model)
		if err != nil {
			r.logger.ErrorWithFields("Failed to convert webhook model", map[string]interface{}{
				"webhook_id": model.ID,
				"error":      err.Error(),
			})
			continue
		}
		webhooks[i] = wh
	}

	return webhooks, nil
}

// GetWebhooksByEvent retrieves webhooks that listen to a specific event
func (r *webhookRepository) GetWebhooksByEvent(ctx context.Context, eventType string) ([]*webhook.WebhookConfig, error) {
	r.logger.InfoWithFields("Getting webhooks by event", map[string]interface{}{
		"event_type": eventType,
	})

	query := `SELECT * FROM "zpWebhooks" WHERE active = true AND $1 = ANY(events) ORDER BY "createdAt" DESC`

	var models []webhookModel
	err := r.db.SelectContext(ctx, &models, query, eventType)
	if err != nil {
		r.logger.ErrorWithFields("Failed to get webhooks by event", map[string]interface{}{
			"event_type": eventType,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get webhooks by event: %w", err)
	}

	webhooks := make([]*webhook.WebhookConfig, len(models))
	for i, model := range models {
		wh, err := r.fromModel(&model)
		if err != nil {
			r.logger.ErrorWithFields("Failed to convert webhook model", map[string]interface{}{
				"webhook_id": model.ID,
				"error":      err.Error(),
			})
			continue
		}
		webhooks[i] = wh
	}

	return webhooks, nil
}

// CountByStatus counts webhooks by active status
func (r *webhookRepository) CountByStatus(ctx context.Context, active bool) (int, error) {
	query := `SELECT COUNT(*) FROM "zpWebhooks" WHERE active = $1`

	var count int
	err := r.db.GetContext(ctx, &count, query, active)
	if err != nil {
		return 0, fmt.Errorf("failed to count webhooks by status: %w", err)
	}

	return count, nil
}

// GetWebhookStats retrieves webhook statistics
func (r *webhookRepository) GetWebhookStats(ctx context.Context, webhookID string) (*ports.WebhookStats, error) {
	// This would typically join with a webhook_stats table
	// For now, return empty stats
	return &ports.WebhookStats{
		WebhookID: webhookID,
	}, nil
}

// UpdateWebhookStats updates webhook statistics
func (r *webhookRepository) UpdateWebhookStats(ctx context.Context, webhookID string, stats *ports.WebhookStats) error {
	// This would typically update a webhook_stats table
	// For now, just log the operation
	r.logger.InfoWithFields("Updating webhook stats", map[string]interface{}{
		"webhook_id": webhookID,
		"stats":      stats,
	})
	return nil
}

// Helper methods

// toModel converts domain entity to database model
func (r *webhookRepository) toModel(wh *webhook.WebhookConfig) *webhookModel {
	model := &webhookModel{
		ID:        wh.ID.String(),
		URL:       wh.URL,
		Active:    wh.Active,
		CreatedAt: wh.CreatedAt,
		UpdatedAt: wh.UpdatedAt,
	}

	if wh.SessionID != nil {
		model.SessionID = sql.NullString{String: *wh.SessionID, Valid: true}
	}

	if wh.Secret != "" {
		model.Secret = sql.NullString{String: wh.Secret, Valid: true}
	}

	// Convert events slice to JSON
	if len(wh.Events) > 0 {
		eventsJSON, err := json.Marshal(wh.Events)
		if err == nil {
			model.Events = string(eventsJSON)
		}
	} else {
		model.Events = "[]"
	}

	return model
}

// fromModel converts database model to domain entity
func (r *webhookRepository) fromModel(model *webhookModel) (*webhook.WebhookConfig, error) {
	id, err := uuid.Parse(model.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid webhook ID: %w", err)
	}

	wh := &webhook.WebhookConfig{
		ID:        id,
		URL:       model.URL,
		Active:    model.Active,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}

	if model.SessionID.Valid {
		wh.SessionID = &model.SessionID.String
	}

	if model.Secret.Valid {
		wh.Secret = model.Secret.String
	}

	// Parse events from JSON
	if model.Events != "" {
		var events []string
		if err := json.Unmarshal([]byte(model.Events), &events); err == nil {
			wh.Events = events
		} else {
			// Fallback to empty array if JSON parsing fails
			wh.Events = []string{}
		}
	} else {
		wh.Events = []string{}
	}

	return wh, nil
}
