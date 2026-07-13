package tenant

type Limits struct {
	DailyEmailLimit    int64
	MonthlyEmailLimit  int64
	MaxStorageGB       int64
	MaxAttachmentsSize int64
	MaxRecipients      int
	RateLimitPerMinute int
	RetainDays         int
}

func DefaultLimits() Limits {
	return Limits{
		DailyEmailLimit:    1000,
		MonthlyEmailLimit:  10000,
		MaxStorageGB:       10,
		MaxAttachmentsSize: 26214400,
		MaxRecipients:      50,
		RateLimitPerMinute: 100,
		RetainDays:         90,
	}
}