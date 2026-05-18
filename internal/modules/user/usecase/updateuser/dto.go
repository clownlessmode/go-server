package updateuser

import "project/internal/modules/user/domain"

type Input struct {
	ID       int64
	Login    *string
	Password *string
	Role     *domain.Role
	IsActive *bool
}
