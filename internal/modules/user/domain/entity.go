package domain

import shareddomain "project/internal/shared/domain"

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

func (r Role) IsValid() bool {
	return r == RoleUser || r == RoleAdmin
}

type User struct {
	shareddomain.BaseEntity
	Login    string
	Password string
	Role     Role
	IsActive bool
}
