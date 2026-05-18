package logout

import (
	"context"

	authdomain "project/internal/modules/auth/domain"
	sharedauth "project/internal/shared/auth"
)

type UseCase struct {
	authRepo     authdomain.Repository
	tokenManager *sharedauth.TokenManager
}

func New(authRepo authdomain.Repository, tokenManager *sharedauth.TokenManager) *UseCase {
	return &UseCase{
		authRepo:     authRepo,
		tokenManager: tokenManager,
	}
}

func (uc *UseCase) Execute(ctx context.Context, input Input) (*Output, error) {
	if _, err := uc.tokenManager.ParseRefreshToken(input.RefreshToken); err != nil {
		return nil, authdomain.ErrInvalidToken
	}

	if err := uc.authRepo.RevokeRefreshSession(ctx, sharedauth.HashToken(input.RefreshToken)); err != nil {
		return nil, authdomain.ErrInvalidToken
	}

	return &Output{}, nil
}
