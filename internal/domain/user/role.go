package user

type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleAdmin      Role = "admin"
	RoleManager    Role = "manager"
	RoleUser       Role = "user"
	RoleAPI        Role = "api"
	RoleReadOnly   Role = "read_only"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleSuperAdmin, RoleAdmin, RoleManager, RoleUser, RoleAPI, RoleReadOnly:
		return true
	}
	return false
}

func (r Role) CanManageUsers() bool {
	return r == RoleSuperAdmin || r == RoleAdmin
}

func (r Role) CanManageTenant() bool {
	return r == RoleSuperAdmin || r == RoleAdmin || r == RoleManager
}