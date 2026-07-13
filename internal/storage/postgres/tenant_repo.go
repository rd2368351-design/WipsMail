package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wispmail/wispmail/internal/domain/shared"
	"github.com/wispmail/wispmail/internal/domain/tenant"
)

type TenantRepo struct {
	pool *pgxpool.Pool
}

func NewTenantRepo(pool *pgxpool.Pool) *TenantRepo {
	return &TenantRepo{pool: pool}
}

func (r *TenantRepo) Save(ctx context.Context, t *tenant.Tenant) error {
	settingsJSON, _ := json.Marshal(t.Settings)
	limitsJSON, _ := json.Marshal(t.Limits)

	query := `
		INSERT INTO tenants (id, name, slug, active, plan, settings, limits, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.pool.Exec(ctx, query,
		t.ID, t.Name, t.Slug, t.Active, t.Plan,
		settingsJSON, limitsJSON, t.CreatedAt, t.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save tenant: %w", err)
	}

	return nil
}

func (r *TenantRepo) FindByID(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
	query := `
		SELECT id, name, slug, active, plan, settings, limits, created_at, updated_at
		FROM tenants WHERE id = $1
	`

	t := &tenant.Tenant{}
	var settingsJSON, limitsJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.Name, &t.Slug, &t.Active, &t.Plan,
		&settingsJSON, &limitsJSON, &t.CreatedAt, &t.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, shared.ErrNotFoundError
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find tenant: %w", err)
	}

	json.Unmarshal(settingsJSON, &t.Settings)
	json.Unmarshal(limitsJSON, &t.Limits)

	return t, nil
}

func (r *TenantRepo) FindBySlug(ctx context.Context, slug string) (*tenant.Tenant, error) {
	query := `
		SELECT id, name, slug, active, plan, settings, limits, created_at, updated_at
		FROM tenants WHERE slug = $1
	`

	t := &tenant.Tenant{}
	var settingsJSON, limitsJSON []byte

	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&t.ID, &t.Name, &t.Slug, &t.Active, &t.Plan,
		&settingsJSON, &limitsJSON, &t.CreatedAt, &t.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, shared.ErrNotFoundError
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find tenant by slug: %w", err)
	}

	json.Unmarshal(settingsJSON, &t.Settings)
	json.Unmarshal(limitsJSON, &t.Limits)

	return t, nil
}

func (r *TenantRepo) List(ctx context.Context, pagination shared.Pagination) ([]tenant.Tenant, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM tenants`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count tenants: %w", err)
	}

	query := `
		SELECT id, name, slug, active, plan, settings, limits, created_at, updated_at
		FROM tenants
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, pagination.Limit(), pagination.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tenants: %w", err)
	}
	defer rows.Close()

	var tenants []tenant.Tenant
	for rows.Next() {
		var t tenant.Tenant
		var settingsJSON, limitsJSON []byte
		if err := rows.Scan(
			&t.ID, &t.Name, &t.Slug, &t.Active, &t.Plan,
			&settingsJSON, &limitsJSON, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan tenant: %w", err)
		}
		json.Unmarshal(settingsJSON, &t.Settings)
		json.Unmarshal(limitsJSON, &t.Limits)
		tenants = append(tenants, t)
	}

	return tenants, total, nil
}

func (r *TenantRepo) Update(ctx context.Context, t *tenant.Tenant) error {
	settingsJSON, _ := json.Marshal(t.Settings)
	limitsJSON, _ := json.Marshal(t.Limits)

	query := `
		UPDATE tenants SET
			name = $1, active = $2, plan = $3,
			settings = $4, limits = $5, updated_at = $6
		WHERE id = $7
	`

	_, err := r.pool.Exec(ctx, query,
		t.Name, t.Active, t.Plan,
		settingsJSON, limitsJSON, t.UpdatedAt, t.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	return nil
}

func (r *TenantRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tenants WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}
	return nil
}