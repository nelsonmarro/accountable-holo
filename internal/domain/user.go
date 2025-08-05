package domain

// UserRole defines the roles a user can have.
type UserRole string

const (
	AdminRole    UserRole = "Admin"
	CustomerRole UserRole = "Customer"
)

// User represents a user of the application.
type User struct {
	BaseEntity
	Username     string   `db:"username"`
	PasswordHash string   `db:"password_hash"`
	Role         UserRole `db:"role"`
}
