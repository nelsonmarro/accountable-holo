package domain

// UserRole defines the roles a user can have.
type UserRole string

const (
	RoleAdmin      UserRole = "Admin"
	RoleSupervisor UserRole = "Supervisor" // Contador / Auditor
	RoleCashier    UserRole = "Cajero"     // Vendedor (Legacy: Customer)
)

// User represents a user of the application.
type User struct {
	BaseEntity
	Username     string   `db:"username"`
	PasswordHash string   `db:"password_hash"`
	FirstName    string   `db:"first_name"`
	LastName     string   `db:"last_name"`
	Role         UserRole `db:"role"`
}

// --- Permission Helpers (Encapsulación) ---

func (u *User) CanViewReports() bool {
	return u.Role == RoleAdmin || u.Role == RoleSupervisor
}

func (u *User) CanConfigureSystem() bool {
	return u.Role == RoleAdmin
}

func (u *User) CanManageUsers() bool {
	return u.Role == RoleAdmin
}

func (u *User) CanVoidTransactions() bool {
	return u.Role == RoleAdmin || u.Role == RoleSupervisor
}

func (u *User) CanReconcile() bool {
	return u.Role == RoleAdmin || u.Role == RoleSupervisor
}
