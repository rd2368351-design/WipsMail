package user

type Permission string

const (
	PermSendEmail       Permission = "email:send"
	PermReadEmail       Permission = "email:read"
	PermDeleteEmail     Permission = "email:delete"
	PermManageUsers     Permission = "users:manage"
	PermManageTenants   Permission = "tenants:manage"
	PermManageDomains   Permission = "domains:manage"
	PermManageBilling   Permission = "billing:manage"
	PermManageWebhooks  Permission = "webhooks:manage"
	PermManageAPIKeys   Permission = "apikeys:manage"
	PermViewAnalytics   Permission = "analytics:view"
	PermManageTemplates Permission = "templates:manage"
)

var RolePermissions = map[Role][]Permission{
	RoleSuperAdmin: {
		PermSendEmail, PermReadEmail, PermDeleteEmail,
		PermManageUsers, PermManageTenants, PermManageDomains,
		PermManageBilling, PermManageWebhooks, PermManageAPIKeys,
		PermViewAnalytics, PermManageTemplates,
	},
	RoleAdmin: {
		PermSendEmail, PermReadEmail, PermDeleteEmail,
		PermManageUsers, PermManageDomains, PermManageWebhooks,
		PermManageAPIKeys, PermViewAnalytics, PermManageTemplates,
	},
	RoleManager: {
		PermSendEmail, PermReadEmail, PermDeleteEmail,
		PermViewAnalytics, PermManageTemplates,
	},
	RoleUser: {
		PermSendEmail, PermReadEmail,
	},
	RoleAPI: {
		PermSendEmail, PermReadEmail,
	},
	RoleReadOnly: {
		PermReadEmail, PermViewAnalytics,
	},
}