package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wispmail/wispmail/internal/domain/shared"
	"github.com/wispmail/wispmail/internal/domain/user"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Save(ctx context.Context, u *user.User) error {
	query := `
		INSERT INTO users (
			id, tenant_id, email, username, password_hash,
			first_name, last_name, role, email_verified, mfa_enabled,
			active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (tenant_id, email) DO UPDATE SET
			username = EXCLUDED.username,
			password_hash = EXCLUDED.password_hash,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.pool.Exec(ctx, query,
		u.ID, u.TenantID, u.Email.String(), u.Username, u.PasswordHash,
		u.FirstName, u.LastName, u.Role, u.EmailVerified, u.MFAEnabled,
		u.Active, u.CreatedAt, u.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}

func (r *UserRepo) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	query := `
		SELECT id, tenant_id, email, username, password_hash,
			first_name, last_name, role, email_verified, mfa_enabled,
			active, last_login_at, created_at, updated_at
		FROM users WHERE id = $1
	`

	u := &user.User{}
	var emailStr string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.TenantID, &emailStr, &u.Username, &u.PasswordHash,
		&u.FirstName, &u.LastName, &u.Role, &u.EmailVerified, &u.MFAEnabled,
		&u.Active, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, shared.ErrNotFoundError
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	u.Email, _ = shared.NewEmailAddress(emailStr)
	return u, nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email shared.EmailAddress) (*user.User, error) {
	query := `
		SELECT id, tenant_id, email, username, password_hash,
			first_name, last_name, role, email_verified, mfa_enabled,
			active, last_login_at, created_at, updated_at
		FROM users WHERE email = $1
	`

	u := &user.User{}
	var emailStr string

	err := r.pool.QueryRow(ctx, query, email.String()).Scan(
		&u.ID, &u.TenantID, &emailStr, &u.Username, &u.PasswordHash,
		&u.FirstName, &u.LastName, &u.Role, &u.EmailVerified, &u.MFAEnabled,
		&u.Active, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, shared.ErrNotFoundError
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	u.Email, _ = shared.NewEmailAddress(emailStr)
	return u, nil
}

func (r *UserRepo) FindByUsername(ctx context.Context, username string) (*user.User, error) {
	query := `
		SELECT id, tenant_id, email, username, password_hash,
			first_name, last_name, role, email_verified, mfa_enabled,
			active, last_login_at, created_at, updated_at
		FROM users WHERE username = $1
	`

	u := &user.User{}
	var emailStr string

	err := r.pool.QueryRow(ctx, query, username).Scan(
		&u.ID, &u.TenantID, &emailStr, &u.Username, &u.PasswordHash,
		&u.FirstName, &u.LastName, &u.Role, &u.EmailVerified, &u.MFAEnabled,
		&u.Active, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, shared.ErrNotFoundError
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user by username: %w", err)
	}

	u.Email, _ = shared.NewEmailAddress(emailStr)
	return u, nil
}

func (r *UserRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, pagination shared.Pagination) ([]user.User, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM users WHERE tenant_id = $1`
	if err := r.pool.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	query := `
		SELECT id, tenant_id, email, username, first_name, last_name,
			role, email_verified, active, created_at
		FROM users WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, tenantID, pagination.Limit(), pagination.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []user.User
	for rows.Next() {
		var u user.User
		var emailStr string
		if err := rows.Scan(
			&u.ID, &u.TenantID, &emailStr, &u.Username,
			&u.FirstName, &u.LastName, &u.Role, &u.EmailVerified,
			&u.Active, &u.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		u.Email, _ = shared.NewEmailAddress(emailStr)
		users = append(users, u)
	}

	return users, total, nil
}

func (r *UserRepo) Update(ctx context.Context, u *user.User) error {
	query := `
		UPDATE users SET
			username = $1, first_name = $2, last_name = $3,
			role = $4, email_verified = $5, mfa_enabled = $6,
			active = $7, last_login_at = $8, updated_at = $9
		WHERE id = $10
	`

	_, err := r.pool.Exec(ctx, query,
		u.Username, u.FirstName, u.LastName,
		u.Role, u.EmailVerified, u.MFAEnabled,
		u.Active, u.LastLoginAt, u.UpdatedAt, u.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}