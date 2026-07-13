package mailbox

type Quota struct {
	MaxMessages int64
	MaxStorage  int64
	UsedStorage int64
}

func DefaultQuota() Quota {
	return Quota{
		MaxMessages: 100000,
		MaxStorage:  10737418240,
	}
}

func (q Quota) StorageUsedPercent() float64 {
	if q.MaxStorage == 0 {
		return 0
	}
	return float64(q.UsedStorage) / float64(q.MaxStorage) * 100
}

func (q Quota) IsStorageFull() bool {
	return q.UsedStorage >= q.MaxStorage
}