package updateuser

import (
	"context"

	"project/internal/modules/user/domain"
)

type UseCase struct {
	repo domain.Repository
}

func New(repo domain.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Execute(ctx context.Context, input Input) (*Output, error) {
	if input.Role != nil && !input.Role.IsValid() {
		return nil, domain.ErrInvalidRole
	}

	user, err := uc.repo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if input.Login != nil {
		user.Login = *input.Login
	}
	if input.Password != nil {
		user.Password = *input.Password
	}
	if input.Role != nil {
		user.Role = *input.Role
	}
	if input.IsActive != nil {
		user.IsActive = *input.IsActive
	}

	user, err = uc.repo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	return &Output{
		ID:        user.ID,
		Login:     user.Login,
		Password:  user.Password,
		Role:      string(user.Role),
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}
