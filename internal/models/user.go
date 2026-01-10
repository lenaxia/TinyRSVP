package models

import "time"

type UserRole string

const (
	RoleAdmin        UserRole = "admin"
	RoleEventManager UserRole = "event_manager"
	RoleGuest        UserRole = "guest"
)

const (
	SystemUserEmail = "system@tinyrsvp.local"
)

type User struct {
	ID          int64      `db:"id" json:"id"`
	Email       string     `db:"email" json:"email"`
	Name        string     `db:"name" json:"name"`
	Role        UserRole   `db:"role" json:"role"`
	OIDCSubject *string    `db:"oidc_subject" json:"oidc_subject,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
	LastLoginAt *time.Time `db:"last_login_at" json:"last_login_at,omitempty"`
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) IsEventManager() bool {
	return u.Role == RoleEventManager || u.Role == RoleAdmin
}

func (u *User) IsSystem() bool {
	return u.Email == SystemUserEmail
}
