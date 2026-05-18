package createuser

import "project/internal/modules/user/domain"

type Input struct {
	Login    string
	Role     domain.Role
	IsActive bool
}
