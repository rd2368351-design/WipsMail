package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wispmail/wispmail/internal/domain/mailbox"
	"github.com/wispmail/wispmail/internal/domain/shared"
)

type MailboxRepo struct {
	pool *pgxpool.Pool
}

func NewMailboxRepo(pool *pgxpool.Pool) *MailboxRepo {
	return &MailboxRepo{pool: pool}
}

func (r *MailboxRepo) Save(ctx context.Context, mb *mailbox.Mailbox) error {
	query := `
		INSERT INTO mailboxes (
			id, tenant_id, user_id, email, name,
			quota_max_messages, quota_max_storage, quota_used_storage,
			message_count, unread_count, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.pool.Exec(ctx, query,
		mb.ID, mb.TenantID, mb.UserID, mb.Email.String(), mb.Name,
		mb.Quota.MaxMessages, mb.Quota.MaxStorage, mb.Quota.UsedStorage,
		mb.MessageCount, mb.UnreadCount, mb.CreatedAt, mb.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save mailbox: %w", err)
	}

	return nil
}

func (r *MailboxRepo) FindByID(ctx context.Context, id uuid.UUID) (*mailbox.Mailbox, error) {
	query := `
		SELECT id, tenant_id, user_id, email, name,
			quota_max_messages, quota_max_storage, quota_used_storage,
			message_count, unread_count, created_at, updated_at
		FROM mailboxes WHERE id = $1
	`

	mb := &mailbox.Mailbox{}
	var emailStr string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&mb.ID, &mb.TenantID, &mb.UserID, &emailStr, &mb.Name,
		&mb.Quota.MaxMessages, &mb.Quota.MaxStorage, &mb.Quota.UsedStorage,
		&mb.MessageCount, &mb.UnreadCount, &mb.CreatedAt, &mb.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, shared.ErrNotFoundError
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find mailbox: %w", err)
	}

	mb.Email, _ = shared.NewEmailAddress(emailStr)
	return mb, nil
}

func (r *MailboxRepo) FindByEmail(ctx context.Context, email shared.EmailAddress) (*mailbox.Mailbox, error) {
	query := `
		SELECT id, tenant_id, user_id, email, name,
			quota_max_messages, quota_max_storage, quota_used_storage,
			message_count, unread_count, created_at, updated_at
		FROM mailboxes WHERE email = $1
	`

	mb := &mailbox.Mailbox{}
	var emailStr string

	err := r.pool.QueryRow(ctx, query, email.String()).Scan(
		&mb.ID, &mb.TenantID, &mb.UserID, &emailStr, &mb.Name,
		&mb.Quota.MaxMessages, &mb.Quota.MaxStorage, &mb.Quota.UsedStorage,
		&mb.MessageCount, &mb.UnreadCount, &mb.CreatedAt, &mb.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, shared.ErrNotFoundError
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find mailbox by email: %w", err)
	}

	mb.Email, _ = shared.NewEmailAddress(emailStr)
	return mb, nil
}

func (r *MailboxRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]mailbox.Mailbox, error) {
	query := `
		SELECT id, tenant_id, user_id, email, name,
			quota_max_messages, quota_max_storage, quota_used_storage,
			message_count, unread_count, created_at, updated_at
		FROM mailboxes WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list mailboxes: %w", err)
	}
	defer rows.Close()

	var mailboxes []mailbox.Mailbox
	for rows.Next() {
		var mb mailbox.Mailbox
		var emailStr string
		if err := rows.Scan(
			&mb.ID, &mb.TenantID, &mb.UserID, &emailStr, &mb.Name,
			&mb.Quota.MaxMessages, &mb.Quota.MaxStorage, &mb.Quota.UsedStorage,
			&mb.MessageCount, &mb.UnreadCount, &mb.CreatedAt, &mb.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan mailbox: %w", err)
		}
		mb.Email, _ = shared.NewEmailAddress(emailStr)
		mailboxes = append(mailboxes, mb)
	}

	return mailboxes, nil
}

func (r *MailboxRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, pagination shared.Pagination) ([]mailbox.Mailbox, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM mailboxes WHERE tenant_id = $1`
	if err := r.pool.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count mailboxes: %w", err)
	}

	query := `
		SELECT id, tenant_id, user_id, email, name,
			quota_max_messages, quota_max_storage, quota_used_storage,
			message_count, unread_count, created_at, updated_at
		FROM mailboxes WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, tenantID, pagination.Limit(), pagination.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list mailboxes: %w", err)
	}
	defer rows.Close()

	var mailboxes []mailbox.Mailbox
	for rows.Next() {
		var mb mailbox.Mailbox
		var emailStr string
		if err := rows.Scan(
			&mb.ID, &mb.TenantID, &mb.UserID, &emailStr, &mb.Name,
			&mb.Quota.MaxMessages, &mb.Quota.MaxStorage, &mb.Quota.UsedStorage,
			&mb.MessageCount, &mb.UnreadCount, &mb.CreatedAt, &mb.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan mailbox: %w", err)
		}
		mb.Email, _ = shared.NewEmailAddress(emailStr)
		mailboxes = append(mailboxes, mb)
	}

	return mailboxes, total, nil
}

func (r *MailboxRepo) Update(ctx context.Context, mb *mailbox.Mailbox) error {
	query := `
		UPDATE mailboxes SET
			name = $1, quota_max_messages = $2, quota_max_storage = $3,
			quota_used_storage = $4, message_count = $5, unread_count = $6,
			updated_at = $7
		WHERE id = $8
	`

	_, err := r.pool.Exec(ctx, query,
		mb.Name, mb.Quota.MaxMessages, mb.Quota.MaxStorage,
		mb.Quota.UsedStorage, mb.MessageCount, mb.UnreadCount,
		mb.UpdatedAt, mb.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update mailbox: %w", err)
	}

	return nil
}

func (r *MailboxRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM mailboxes WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete mailbox: %w", err)
	}
	return nil
}