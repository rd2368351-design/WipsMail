package billing

type Plan struct {
	ID             string
	Name           string
	Price          float64
	DailyLimit     int64
	MonthlyLimit   int64
	MaxStorage     int64
	MaxUsers       int
	MaxDomains     int
	PremiumSupport bool
	DedicatedIP    bool
	CustomDomain   bool
}

var Plans = map[string]Plan{
	"free": {
		ID: "free", Name: "Free", Price: 0, DailyLimit: 100, MonthlyLimit: 1000,
		MaxStorage: 1, MaxUsers: 1, MaxDomains: 1,
	},
	"starter": {
		ID: "starter", Name: "Starter", Price: 29, DailyLimit: 1000, MonthlyLimit: 10000,
		MaxStorage: 10, MaxUsers: 5, MaxDomains: 3,
	},
	"pro": {
		ID: "pro", Name: "Professional", Price: 99, DailyLimit: 10000, MonthlyLimit: 100000,
		MaxStorage: 50, MaxUsers: 20, MaxDomains: 10, PremiumSupport: true, CustomDomain: true,
	},
	"enterprise": {
		ID: "enterprise", Name: "Enterprise", Price: 499, DailyLimit: 100000, MonthlyLimit: 1000000,
		MaxStorage: 500, MaxUsers: 100, MaxDomains: 100, PremiumSupport: true, DedicatedIP: true, CustomDomain: true,
	},
}