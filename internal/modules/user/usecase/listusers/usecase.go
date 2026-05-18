package listusers

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
	users, err := uc.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	out := &Output{
		Users: make([]UserOutput, 0, len(users)),
	}
	for _, user := range users {
		out.Users = append(out.Users, UserOutput{
			ID:        user.ID,
			Login:     user.Login,
			Password:  user.Password,
			Role:      string(user.Role),
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	return out, nil
}
