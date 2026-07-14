package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wispmail/wispmail/internal/domain/domain"
	"github.com/wispmail/wispmail/internal/domain/shared"
)

type DomainRepo struct {
	pool *pgxpool.Pool
}

func NewDomainRepo(pool *pgxpool.Pool) *DomainRepo {
	return &DomainRepo{pool: pool}
}

func (r *DomainRepo) Save(ctx context.Context, d *domain.Domain) error {
	query := `
		INSERT INTO domains (id, tenant_id, name, verified, dkim_verified, spf_verified, dmarc_verified,
			verification_token, verification_record_type, verification_record_host,
			verification_record_value, verification_expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := r.pool.Exec(ctx, query,
		d.ID, d.TenantID, d.Name, d.Verified, d.DkimVerified, d.SpfVerified, d.DmarcVerified,
		d.Verification.Token, d.Verification.RecordType, d.Verification.RecordHost,
		d.Verification.RecordValue, d.Verification.ExpiresAt, d.CreatedAt, d.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save domain: %w", err)
	}

	return nil
}

func (r *DomainRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Domain, error) {
	query := `
		SELECT id, tenant_id, name, verified, dkim_verified, spf_verified, dmarc_verified,
			verification_token, verification_record_type, verification_record_host,
			verification_record_value, verification_expires_at, created_at, updated_at
		FROM domains WHERE id = $1
	`

	d := &domain.Domain{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&d.ID, &d.TenantID, &d.Name, &d.Verified, &d.DkimVerified, &d.SpfVerified, &d.DmarcVerified,
		&d.Verification.Token, &d.Verification.RecordType, &d.Verification.RecordHost,
		&d.Verification.RecordValue, &d.Verification.ExpiresAt, &d.CreatedAt, &d.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, shared.ErrNotFoundError
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find domain: %w", err)
	}

	return d, nil
}

func (r *DomainRepo) FindByName(ctx context.Context, name string) (*domain.Domain, error) {
	query := `
		SELECT id, tenant_id, name, verified, dkim_verified, spf_verified, dmarc_verified,
			verification_token, verification_record_type, verification_record_host,
			verification_record_value, verification_expires_at, created_at, updated_at
		FROM domains WHERE name = $1
	`

	d := &domain.Domain{}
	err := r.pool.QueryRow(ctx, query, name).Scan(
		&d.ID, &d.TenantID, &d.Name, &d.Verified, &d.DkimVerified, &d.SpfVerified, &d.DmarcVerified,
		&d.Verification.Token, &d.Verification.RecordType, &d.Verification.RecordHost,
		&d.Verification.RecordValue, &d.Verification.ExpiresAt, &d.CreatedAt, &d.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, shared.ErrNotFoundError
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find domain by name: %w", err)
	}

	return d, nil
}

func (r *DomainRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, pagination shared.Pagination) ([]domain.Domain, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM domains WHERE tenant_id = $1`, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count domains: %w", err)
	}

	query := `
		SELECT id, tenant_id, name, verified, dkim_verified, spf_verified, dmarc_verified, created_at, updated_at
		FROM domains WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, tenantID, pagination.Limit(), pagination.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list domains: %w", err)
	}
	defer rows.Close()

	var domains []domain.Domain
	for rows.Next() {
		var d domain.Domain
		if err := rows.Scan(
			&d.ID, &d.TenantID, &d.Name, &d.Verified, &d.DkimVerified,
			&d.SpfVerified, &d.DmarcVerified, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan domain: %w", err)
		}
		domains = append(domains, d)
	}

	return domains, total, nil
}

func (r *DomainRepo) Update(ctx context.Context, d *domain.Domain) error {
	query := `
		UPDATE domains SET
			verified = $1, dkim_verified = $2, spf_verified = $3, dmarc_verified = $4,
			verification_token = $5, verification_record_type = $6, verification_record_host = $7,
			verification_record_value = $8, verification_expires_at = $9, updated_at = $10
		WHERE id = $11
	`

	_, err := r.pool.Exec(ctx, query,
		d.Verified, d.DkimVerified, d.SpfVerified, d.DmarcVerified,
		d.Verification.Token, d.Verification.RecordType, d.Verification.RecordHost,
		d.Verification.RecordValue, d.Verification.ExpiresAt, time.Now().UTC(), d.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update domain: %w", err)
	}

	return nil
}

func (r *DomainRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM domains WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete domain: %w", err)
	}
	return nil
}