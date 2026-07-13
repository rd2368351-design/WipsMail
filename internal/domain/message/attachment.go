package message

import "github.com/google/uuid"

type Attachment struct {
	ID          uuid.UUID
	MessageID   uuid.UUID
	Filename    string
	ContentType string
	Size        int64
	S3Key       string
	S3Bucket    string
	Inline      bool
	ContentID   string
}

func NewAttachment(filename, contentType string, size int64) *Attachment {
	return &Attachment{
		ID:          uuid.New(),
		Filename:    filename,
		ContentType: contentType,
		Size:        size,
	}
}