package tenant

type Settings struct {
	Timezone          string
	DefaultLanguage   string
	MaxUsers          int
	AllowAPIAccess    bool
	AllowSMTPSending  bool
	RequireDKIM       bool
	RequireSPF        bool
	WebhookURL        string
	CustomDomain      string
	NotificationEmail string
}

func DefaultSettings() Settings {
	return Settings{
		Timezone:         "UTC",
		DefaultLanguage:  "en",
		MaxUsers:         10,
		AllowAPIAccess:   true,
		AllowSMTPSending: true,
		RequireDKIM:      true,
		RequireSPF:       false,
	}
}