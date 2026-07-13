package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wispmail/wispmail/internal/domain/billing"
	"github.com/wispmail/wispmail/internal/domain/shared"
)

type BillingRepo struct {
	pool *pgxpool.Pool
}

func NewBillingRepo(pool *pgxpool.Pool) *BillingRepo {
	return &BillingRepo{pool: pool}
}

func (r *BillingRepo) SaveInvoice(ctx context.Context, invoice *billing.Invoice) error {
	query := `
		INSERT INTO invoices (id, tenant_id, invoice_number, status, amount, currency,
			description, period_start, period_end, due_date, paid_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.pool.Exec(ctx, query,
		invoice.ID, invoice.TenantID, invoice.InvoiceNumber, invoice.Status,
		invoice.Amount, invoice.Currency, invoice.Description,
		invoice.PeriodStart, invoice.PeriodEnd, invoice.DueDate,
		invoice.PaidAt, invoice.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save invoice: %w", err)
	}

	return nil
}

func (r *BillingRepo) FindInvoice(ctx context.Context, id uuid.UUID) (*billing.Invoice, error) {
	query := `
		SELECT id, tenant_id, invoice_number, status, amount, currency,
			description, period_start, period_end, due_date, paid_at, created_at
		FROM invoices WHERE id = $1
	`

	inv := &billing.Invoice{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&inv.ID, &inv.TenantID, &inv.InvoiceNumber, &inv.Status,
		&inv.Amount, &inv.Currency, &inv.Description,
		&inv.PeriodStart, &inv.PeriodEnd, &inv.DueDate,
		&inv.PaidAt, &inv.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, shared.ErrNotFoundError
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find invoice: %w", err)
	}

	return inv, nil
}

func (r *BillingRepo) ListInvoices(ctx context.Context, tenantID uuid.UUID, pagination shared.Pagination) ([]billing.Invoice, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM invoices WHERE tenant_id = $1`, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices: %w", err)
	}

	query := `
		SELECT id, tenant_id, invoice_number, status, amount, currency,
			period_start, period_end, due_date, created_at
		FROM invoices WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, tenantID, pagination.Limit(), pagination.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list invoices: %w", err)
	}
	defer rows.Close()

	var invoices []billing.Invoice
	for rows.Next() {
		var inv billing.Invoice
		if err := rows.Scan(
			&inv.ID, &inv.TenantID, &inv.InvoiceNumber, &inv.Status,
			&inv.Amount, &inv.Currency, &inv.PeriodStart, &inv.PeriodEnd,
			&inv.DueDate, &inv.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, inv)
	}

	return invoices, total, nil
}

func (r *BillingRepo) UpdateInvoiceStatus(ctx context.Context, id uuid.UUID, status billing.InvoiceStatus) error {
	query := `UPDATE invoices SET status = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update invoice status: %w", err)
	}
	return nil
}

func (r *BillingRepo) SaveUsage(ctx context.Context, usage *billing.UsageRecord) error {
	query := `
		INSERT INTO usage_records (id, tenant_id, date, emails_sent, emails_failed,
			emails_bounced, storage_bytes, api_calls)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (tenant_id, date) DO UPDATE SET
			emails_sent = usage_records.emails_sent + EXCLUDED.emails_sent,
			emails_failed = usage_records.emails_failed + EXCLUDED.emails_failed,
			emails_bounced = usage_records.emails_bounced + EXCLUDED.emails_bounced,
			storage_bytes = EXCLUDED.storage_bytes,
			api_calls = usage_records.api_calls + EXCLUDED.api_calls
	`

	_, err := r.pool.Exec(ctx, query,
		usage.ID, usage.TenantID, usage.Date,
		usage.EmailsSent, usage.EmailsFailed, usage.EmailsBounced,
		usage.StorageBytes, usage.APICalls,
	)

	if err != nil {
		return fmt.Errorf("failed to save usage: %w", err)
	}

	return nil
}

func (r *BillingRepo) GetUsage(ctx context.Context, tenantID uuid.UUID, periodStart, periodEnd time.Time) ([]billing.UsageRecord, error) {
	query := `
		SELECT id, tenant_id, date, emails_sent, emails_failed,
			emails_bounced, storage_bytes, api_calls
		FROM usage_records
		WHERE tenant_id = $1 AND date BETWEEN $2 AND $3
		ORDER BY date ASC
	`

	rows, err := r.pool.Query(ctx, query, tenantID, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage: %w", err)
	}
	defer rows.Close()

	var records []billing.UsageRecord
	for rows.Next() {
		var record billing.UsageRecord
		if err := rows.Scan(
			&record.ID, &record.TenantID, &record.Date,
			&record.EmailsSent, &record.EmailsFailed, &record.EmailsBounced,
			&record.StorageBytes, &record.APICalls,
		); err != nil {
			return nil, fmt.Errorf("failed to scan usage record: %w", err)
		}
		records = append(records, record)
	}

	return records, nil
}

func (r *BillingRepo) GetCurrentUsage(ctx context.Context, tenantID uuid.UUID) (*billing.UsageRecord, error) {
	query := `
		SELECT id, tenant_id, date, emails_sent, emails_failed,
			emails_bounced, storage_bytes, api_calls
		FROM usage_records
		WHERE tenant_id = $1 AND date = CURRENT_DATE
	`

	record := &billing.UsageRecord{}
	err := r.pool.QueryRow(ctx, query, tenantID).Scan(
		&record.ID, &record.TenantID, &record.Date,
		&record.EmailsSent, &record.EmailsFailed, &record.EmailsBounced,
		&record.StorageBytes, &record.APICalls,
	)

	if err == pgx.ErrNoRows {
		return billing.NewUsageRecord(tenantID), nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get current usage: %w", err)
	}

	return record, nil
}