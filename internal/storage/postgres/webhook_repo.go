package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wispmail/wispmail/internal/domain/shared"
	"github.com/wispmail/wispmail/internal/domain/webhook"
)

type WebhookRepo struct {
	pool *pgxpool.Pool
}

func NewWebhookRepo(pool *pgxpool.Pool) *WebhookRepo {
	return &WebhookRepo{pool: pool}
}

func (r *WebhookRepo) Save(ctx context.Context, w *webhook.Webhook) error {
	query := `
		INSERT INTO webhooks (id, tenant_id, name, url, secret, events, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.pool.Exec(ctx, query,
		w.ID, w.TenantID, w.Name, w.URL, w.Secret,
		w.Events, w.Active, w.CreatedAt, w.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save webhook: %w", err)
	}

	return nil
}

func (r *WebhookRepo) FindByID(ctx context.Context, id uuid.UUID) (*webhook.Webhook, error) {
	query := `
		SELECT id, tenant_id, name, url, secret, events, active, created_at, updated_at
		FROM webhooks WHERE id = $1
	`

	w := &webhook.Webhook{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&w.ID, &w.TenantID, &w.Name, &w.URL, &w.Secret,
		&w.Events, &w.Active, &w.CreatedAt, &w.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, shared.ErrNotFoundError
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find webhook: %w", err)
	}

	return w, nil
}

func (r *WebhookRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, pagination shared.Pagination) ([]webhook.Webhook, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM webhooks WHERE tenant_id = $1`, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count webhooks: %w", err)
	}

	query := `
		SELECT id, tenant_id, name, url, events, active, created_at, updated_at
		FROM webhooks WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, tenantID, pagination.Limit(), pagination.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list webhooks: %w", err)
	}
	defer rows.Close()

	var webhooks []webhook.Webhook
	for rows.Next() {
		var w webhook.Webhook
		if err := rows.Scan(
			&w.ID, &w.TenantID, &w.Name, &w.URL,
			&w.Events, &w.Active, &w.CreatedAt, &w.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan webhook: %w", err)
		}
		webhooks = append(webhooks, w)
	}

	return webhooks, total, nil
}

func (r *WebhookRepo) ListByEvent(ctx context.Context, tenantID uuid.UUID, event webhook.EventType) ([]webhook.Webhook, error) {
	query := `
		SELECT id, tenant_id, name, url, secret, events, active, created_at, updated_at
		FROM webhooks WHERE tenant_id = $1 AND $2 = ANY(events) AND active = true
	`

	rows, err := r.pool.Query(ctx, query, tenantID, event)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks by event: %w", err)
	}
	defer rows.Close()

	var webhooks []webhook.Webhook
	for rows.Next() {
		var w webhook.Webhook
		if err := rows.Scan(
			&w.ID, &w.TenantID, &w.Name, &w.URL, &w.Secret,
			&w.Events, &w.Active, &w.CreatedAt, &w.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan webhook: %w", err)
		}
		webhooks = append(webhooks, w)
	}

	return webhooks, nil
}

func (r *WebhookRepo) Update(ctx context.Context, w *webhook.Webhook) error {
	query := `
		UPDATE webhooks SET
			name = $1, url = $2, secret = $3,
			events = $4, active = $5, updated_at = $6
		WHERE id = $7
	`

	_, err := r.pool.Exec(ctx, query,
		w.Name, w.URL, w.Secret,
		w.Events, w.Active, w.UpdatedAt, w.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update webhook: %w", err)
	}

	return nil
}

func (r *WebhookRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM webhooks WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}
	return nil
}