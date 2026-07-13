package mailbox

import "github.com/google/uuid"

type FolderType string

const (
	FolderInbox   FolderType = "inbox"
	FolderSent    FolderType = "sent"
	FolderDrafts  FolderType = "drafts"
	FolderTrash   FolderType = "trash"
	FolderSpam    FolderType = "spam"
	FolderArchive FolderType = "archive"
	FolderCustom  FolderType = "custom"
)

type Folder struct {
	ID           uuid.UUID
	MailboxID    uuid.UUID
	Name         string
	Type         FolderType
	MessageCount int64
	UnreadCount  int64
}

func NewFolder(mailboxID uuid.UUID, name string, folderType FolderType) *Folder {
	return &Folder{
		ID:        uuid.New(),
		MailboxID: mailboxID,
		Name:      name,
		Type:      folderType,
	}
}

func (f Folder) IsSystemFolder() bool {
	return f.Type != FolderCustom
}