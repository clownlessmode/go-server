package getuser

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
	user, err := uc.repo.GetByID(ctx, input.ID)
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
