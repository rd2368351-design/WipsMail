package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wispmail/wispmail/internal/domain/shared"
	"github.com/wispmail/wispmail/internal/domain/template"
)

type TemplateRepo struct {
	pool *pgxpool.Pool
}

func NewTemplateRepo(pool *pgxpool.Pool) *TemplateRepo {
	return &TemplateRepo{pool: pool}
}

func (r *TemplateRepo) Save(ctx context.Context, tmpl *template.Template) error {
	query := `
		INSERT INTO templates (id, tenant_id, name, subject, html_body, plain_body, version, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.pool.Exec(ctx, query,
		tmpl.ID, tmpl.TenantID, tmpl.Name, tmpl.Subject,
		tmpl.HTMLBody, tmpl.PlainBody, tmpl.Version, tmpl.Active,
		tmpl.CreatedAt, tmpl.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save template: %w", err)
	}

	return nil
}

func (r *TemplateRepo) FindByID(ctx context.Context, id uuid.UUID) (*template.Template, error) {
	query := `
		SELECT id, tenant_id, name, subject, html_body, plain_body, version, active, created_at, updated_at
		FROM templates WHERE id = $1
	`

	tmpl := &template.Template{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&tmpl.ID, &tmpl.TenantID, &tmpl.Name, &tmpl.Subject,
		&tmpl.HTMLBody, &tmpl.PlainBody, &tmpl.Version, &tmpl.Active,
		&tmpl.CreatedAt, &tmpl.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, shared.ErrNotFoundError
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find template: %w", err)
	}

	return tmpl, nil
}

func (r *TemplateRepo) FindByName(ctx context.Context, tenantID uuid.UUID, name string) (*template.Template, error) {
	query := `
		SELECT id, tenant_id, name, subject, html_body, plain_body, version, active, created_at, updated_at
		FROM templates WHERE tenant_id = $1 AND name = $2
	`

	tmpl := &template.Template{}
	err := r.pool.QueryRow(ctx, query, tenantID, name).Scan(
		&tmpl.ID, &tmpl.TenantID, &tmpl.Name, &tmpl.Subject,
		&tmpl.HTMLBody, &tmpl.PlainBody, &tmpl.Version, &tmpl.Active,
		&tmpl.CreatedAt, &tmpl.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, shared.ErrNotFoundError
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find template by name: %w", err)
	}

	return tmpl, nil
}

func (r *TemplateRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, pagination shared.Pagination) ([]template.Template, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM templates WHERE tenant_id = $1`, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count templates: %w", err)
	}

	query := `
		SELECT id, tenant_id, name, subject, version, active, created_at, updated_at
		FROM templates WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, tenantID, pagination.Limit(), pagination.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list templates: %w", err)
	}
	defer rows.Close()

	var templates []template.Template
	for rows.Next() {
		var tmpl template.Template
		if err := rows.Scan(
			&tmpl.ID, &tmpl.TenantID, &tmpl.Name, &tmpl.Subject,
			&tmpl.Version, &tmpl.Active, &tmpl.CreatedAt, &tmpl.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan template: %w", err)
		}
		templates = append(templates, tmpl)
	}

	return templates, total, nil
}

func (r *TemplateRepo) Update(ctx context.Context, tmpl *template.Template) error {
	query := `
		UPDATE templates SET
			subject = $1, html_body = $2, plain_body = $3,
			version = $4, active = $5, updated_at = $6
		WHERE id = $7
	`

	_, err := r.pool.Exec(ctx, query,
		tmpl.Subject, tmpl.HTMLBody, tmpl.PlainBody,
		tmpl.Version, tmpl.Active, tmpl.UpdatedAt, tmpl.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	return nil
}

func (r *TemplateRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM templates WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}
	return nil
}

func (r *TemplateRepo) SaveVersion(ctx context.Context, version *template.Version) error {
	query := `
		INSERT INTO template_versions (id, template_id, version, subject, html_body, plain_body, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		version.ID, version.TemplateID, version.Version,
		version.Subject, version.HTMLBody, version.PlainBody,
		version.CreatedBy, version.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save template version: %w", err)
	}

	return nil
}

func (r *TemplateRepo) ListVersions(ctx context.Context, templateID uuid.UUID) ([]template.Version, error) {
	query := `
		SELECT id, template_id, version, subject, html_body, plain_body, created_by, created_at
		FROM template_versions WHERE template_id = $1
		ORDER BY version DESC
	`

	rows, err := r.pool.Query(ctx, query, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to list template versions: %w", err)
	}
	defer rows.Close()

	var versions []template.Version
	for rows.Next() {
		var v template.Version
		if err := rows.Scan(
			&v.ID, &v.TemplateID, &v.Version,
			&v.Subject, &v.HTMLBody, &v.PlainBody,
			&v.CreatedBy, &v.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan template version: %w", err)
		}
		versions = append(versions, v)
	}

	return versions, nil
}