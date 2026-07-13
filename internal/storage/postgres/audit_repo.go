package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wispmail/wispmail/internal/domain/audit"
	"github.com/wispmail/wispmail/internal/domain/shared"
)

type AuditRepo struct {
	pool *pgxpool.Pool
}

func NewAuditRepo(pool *pgxpool.Pool) *AuditRepo {
	return &AuditRepo{pool: pool}
}

func (r *AuditRepo) Save(ctx context.Context, event *audit.Event) error {
	query := `
		INSERT INTO audit_logs (id, tenant_id, actor_id, actor_email, action,
			resource, resource_id, details, ip_address, user_agent,
			success, error_message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.pool.Exec(ctx, query,
		event.ID, event.TenantID, event.ActorID, event.ActorEmail, event.Action,
		event.Resource, event.ResourceID, event.Details, event.IPAddress, event.UserAgent,
		event.Success, event.ErrorMessage, event.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save audit event: %w", err)
	}

	return nil
}

func (r *AuditRepo) FindByID(ctx context.Context, id uuid.UUID) (*audit.Event, error) {
	query := `
		SELECT id, tenant_id, actor_id, actor_email, action,
			resource, resource_id, details, ip_address, user_agent,
			success, error_message, created_at
		FROM audit_logs WHERE id = $1
	`

	event := &audit.Event{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&event.ID, &event.TenantID, &event.ActorID, &event.ActorEmail, &event.Action,
		&event.Resource, &event.ResourceID, &event.Details, &event.IPAddress, &event.UserAgent,
		&event.Success, &event.ErrorMessage, &event.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, shared.ErrNotFoundError
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find audit event: %w", err)
	}

	return event, nil
}

func (r *AuditRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, pagination shared.Pagination) ([]audit.Event, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs WHERE tenant_id = $1`, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count audit events: %w", err)
	}

	query := `
		SELECT id, tenant_id, actor_id, actor_email, action,
			resource, resource_id, success, created_at
		FROM audit_logs WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, tenantID, pagination.Limit(), pagination.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list audit events: %w", err)
	}
	defer rows.Close()

	var events []audit.Event
	for rows.Next() {
		var event audit.Event
		if err := rows.Scan(
			&event.ID, &event.TenantID, &event.ActorID, &event.ActorEmail, &event.Action,
			&event.Resource, &event.ResourceID, &event.Success, &event.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit event: %w", err)
		}
		events = append(events, event)
	}

	return events, total, nil
}

func (r *AuditRepo) ListByActor(ctx context.Context, actorID uuid.UUID, pagination shared.Pagination) ([]audit.Event, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs WHERE actor_id = $1`, actorID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count audit events: %w", err)
	}

	query := `
		SELECT id, tenant_id, actor_id, actor_email, action,
			resource, resource_id, success, created_at
		FROM audit_logs WHERE actor_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, actorID, pagination.Limit(), pagination.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list audit events by actor: %w", err)
	}
	defer rows.Close()

	var events []audit.Event
	for rows.Next() {
		var event audit.Event
		if err := rows.Scan(
			&event.ID, &event.TenantID, &event.ActorID, &event.ActorEmail, &event.Action,
			&event.Resource, &event.ResourceID, &event.Success, &event.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit event: %w", err)
		}
		events = append(events, event)
	}

	return events, total, nil
}

func (r *AuditRepo) ListByAction(ctx context.Context, tenantID uuid.UUID, action audit.Action, pagination shared.Pagination) ([]audit.Event, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs WHERE tenant_id = $1 AND action = $2`, tenantID, action).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count audit events: %w", err)
	}

	query := `
		SELECT id, tenant_id, actor_id, actor_email, action,
			resource, resource_id, success, created_at
		FROM audit_logs WHERE tenant_id = $1 AND action = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(ctx, query, tenantID, action, pagination.Limit(), pagination.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list audit events by action: %w", err)
	}
	defer rows.Close()

	var events []audit.Event
	for rows.Next() {
		var event audit.Event
		if err := rows.Scan(
			&event.ID, &event.TenantID, &event.ActorID, &event.ActorEmail, &event.Action,
			&event.Resource, &event.ResourceID, &event.Success, &event.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit event: %w", err)
		}
		events = append(events, event)
	}

	return events, total, nil
}