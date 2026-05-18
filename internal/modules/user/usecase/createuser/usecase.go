package createuser

import (
	"context"
	"crypto/rand"
	"math/big"

	"project/internal/modules/user/domain"
)

const passwordLength = 16

type UseCase struct {
	repo domain.Repository
}

func New(repo domain.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Execute(ctx context.Context, input Input) (*Output, error) {
	if !input.Role.IsValid() {
		return nil, domain.ErrInvalidRole
	}

	password, err := generatePassword(passwordLength)
	if err != nil {
		return nil, err
	}

	user, err := uc.repo.Create(ctx, &domain.User{
		Login:    input.Login,
		Password: password,
		Role:     input.Role,
		IsActive: input.IsActive,
	})
	if err != nil {
		return nil, err
	}

	return toOutput(user), nil
}

func generatePassword(length int) (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	password := make([]byte, length)
	max := big.NewInt(int64(len(alphabet)))

	for i := range password {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}

		password[i] = alphabet[n.Int64()]
	}

	return string(password), nil
}

func toOutput(user *domain.User) *Output {
	return &Output{
		ID:        user.ID,
		Login:     user.Login,
		Password:  user.Password,
		Role:      string(user.Role),
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
