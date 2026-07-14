package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AttachmentRepo struct {
	pool *pgxpool.Pool
}

func NewAttachmentRepo(pool *pgxpool.Pool) *AttachmentRepo {
	return &AttachmentRepo{pool: pool}
}

func (r *AttachmentRepo) Save(ctx context.Context, messageID, attachmentID uuid.UUID, filename, contentType, s3Key, s3Bucket string, size int64, inline bool, contentID string) error {
	query := `
		INSERT INTO attachments (id, message_id, filename, content_type, size, s3_key, s3_bucket, inline, content_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
	`

	_, err := r.pool.Exec(ctx, query,
		attachmentID, messageID, filename, contentType, size,
		s3Key, s3Bucket, inline, contentID,
	)

	if err != nil {
		return fmt.Errorf("failed to save attachment: %w", err)
	}

	return nil
}

func (r *AttachmentRepo) FindByMessageID(ctx context.Context, messageID uuid.UUID) ([]AttachmentInfo, error) {
	query := `
		SELECT id, filename, content_type, size, s3_key, s3_bucket, inline, content_id
		FROM attachments WHERE message_id = $1
	`

	rows, err := r.pool.Query(ctx, query, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to find attachments: %w", err)
	}
	defer rows.Close()

	var attachments []AttachmentInfo
	for rows.Next() {
		var att AttachmentInfo
		if err := rows.Scan(
			&att.ID, &att.Filename, &att.ContentType, &att.Size,
			&att.S3Key, &att.S3Bucket, &att.Inline, &att.ContentID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan attachment: %w", err)
		}
		attachments = append(attachments, att)
	}

	return attachments, nil
}

func (r *AttachmentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM attachments WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete attachment: %w", err)
	}
	return nil
}

type AttachmentInfo struct {
	ID          uuid.UUID
	Filename    string
	ContentType string
	Size        int64
	S3Key       string
	S3Bucket    string
	Inline      bool
	ContentID   string
}