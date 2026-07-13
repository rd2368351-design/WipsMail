package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wispmail/wispmail/internal/domain/message"
	"github.com/wispmail/wispmail/internal/domain/shared"
)

type MessageRepo struct {
	pool *pgxpool.Pool
}

func NewMessageRepo(pool *pgxpool.Pool) *MessageRepo {
	return &MessageRepo{pool: pool}
}

func (r *MessageRepo) Save(ctx context.Context, msg *message.Message) error {
	query := `
		INSERT INTO messages (
			id, tenant_id, message_id, in_reply_to, references,
			subject, from_address, sender, reply_to,
			to_addresses, cc_addresses, bcc_addresses,
			body_plain, body_html, body_raw,
			status, priority, size, received_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12,
			$13, $14, $15,
			$16, $17, $18, $19, $20, $21
		)
	`

	_, err := r.pool.Exec(ctx, query,
		msg.ID, msg.TenantID, msg.MessageID, msg.InReplyTo, msg.References,
		msg.Subject, msg.From.String(), msg.Sender.String(), msg.ReplyTo.String(),
		formatAddressList(msg.To), formatAddressList(msg.Cc), formatAddressList(msg.Bcc),
		msg.Body.PlainText, msg.Body.HTML, msg.Body.Raw,
		msg.Status, msg.Priority, msg.Size, msg.ReceivedAt, msg.CreatedAt, msg.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	return nil
}

func (r *MessageRepo) FindByID(ctx context.Context, id uuid.UUID) (*message.Message, error) {
	query := `
		SELECT id, tenant_id, message_id, in_reply_to, references,
			subject, from_address, sender, reply_to,
			to_addresses, cc_addresses, bcc_addresses,
			body_plain, body_html, body_raw,
			status, priority, size, received_at, created_at, updated_at
		FROM messages WHERE id = $1
	`

	msg := &message.Message{}
	var fromStr, senderStr, replyToStr, toStr, ccStr, bccStr string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&msg.ID, &msg.TenantID, &msg.MessageID, &msg.InReplyTo, &msg.References,
		&msg.Subject, &fromStr, &senderStr, &replyToStr,
		&toStr, &ccStr, &bccStr,
		&msg.Body.PlainText, &msg.Body.HTML, &msg.Body.Raw,
		&msg.Status, &msg.Priority, &msg.Size, &msg.ReceivedAt, &msg.CreatedAt, &msg.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, shared.ErrNotFoundError
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find message: %w", err)
	}

	msg.From, _ = shared.NewEmailAddress(fromStr)
	msg.Sender, _ = shared.NewEmailAddress(senderStr)
	msg.ReplyTo, _ = shared.NewEmailAddress(replyToStr)
	msg.To = parseAddressList(toStr)
	msg.Cc = parseAddressList(ccStr)
	msg.Bcc = parseAddressList(bccStr)

	return msg, nil
}

func (r *MessageRepo) FindByMessageID(ctx context.Context, messageID string) (*message.Message, error) {
	query := `SELECT id FROM messages WHERE message_id = $1`
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, query, messageID).Scan(&id)
	if err == pgx.ErrNoRows {
		return nil, shared.ErrNotFoundError
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find message by message_id: %w", err)
	}
	return r.FindByID(ctx, id)
}

func (r *MessageRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, pagination shared.Pagination, filters *shared.FilterSet, sort shared.Sort) ([]message.Message, int64, error) {
	whereClause := "WHERE tenant_id = $1"
	args := []any{tenantID}
	argIndex := 2

	if filters != nil && !filters.IsEmpty() {
		for _, f := range filters.Filters {
			switch f.Operator {
			case shared.OpEquals:
				whereClause += fmt.Sprintf(" AND %s = $%d", f.Field, argIndex)
				args = append(args, f.Value)
				argIndex++
			case shared.OpContains:
				whereClause += fmt.Sprintf(" AND %s ILIKE $%d", f.Field, argIndex)
				args = append(args, "%"+fmt.Sprintf("%v", f.Value)+"%")
				argIndex++
			case shared.OpIn:
				placeholders := make([]string, len(f.Values))
				for i, v := range f.Values {
					placeholders[i] = fmt.Sprintf("$%d", argIndex)
					args = append(args, v)
					argIndex++
				}
				whereClause += fmt.Sprintf(" AND %s IN (%s)", f.Field, strings.Join(placeholders, ", "))
			}
		}
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM messages %s", whereClause)
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count messages: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, tenant_id, message_id, subject, from_address,
			to_addresses, status, priority, size, created_at
		FROM messages %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sort.String(), argIndex, argIndex+1)

	args = append(args, pagination.Limit(), pagination.Offset())

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list messages: %w", err)
	}
	defer rows.Close()

	var messages []message.Message
	for rows.Next() {
		var msg message.Message
		var fromStr, toStr string
		if err := rows.Scan(
			&msg.ID, &msg.TenantID, &msg.MessageID, &msg.Subject,
			&fromStr, &toStr, &msg.Status, &msg.Priority, &msg.Size, &msg.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan message: %w", err)
		}
		msg.From, _ = shared.NewEmailAddress(fromStr)
		msg.To = parseAddressList(toStr)
		messages = append(messages, msg)
	}

	return messages, total, nil
}

func (r *MessageRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status message.Status) error {
	query := `UPDATE messages SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, status, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to update message status: %w", err)
	}
	return nil
}

func (r *MessageRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM messages WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}
	return nil
}

func formatAddressList(addresses []shared.EmailAddress) string {
	var result []string
	for _, addr := range addresses {
		result = append(result, addr.String())
	}
	return strings.Join(result, ", ")
}

func parseAddressList(s string) []shared.EmailAddress {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ", ")
	var addresses []shared.EmailAddress
	for _, part := range parts {
		addr, err := shared.NewEmailAddress(strings.TrimSpace(part))
		if err == nil {
			addresses = append(addresses, addr)
		}
	}
	return addresses
}