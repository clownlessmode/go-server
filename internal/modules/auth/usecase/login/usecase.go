package login

import (
	"context"

	authdomain "project/internal/modules/auth/domain"
	userdomain "project/internal/modules/user/domain"
	sharedauth "project/internal/shared/auth"
)

type UseCase struct {
	authRepo     authdomain.Repository
	userRepo     userdomain.Repository
	tokenManager *sharedauth.TokenManager
}

func New(authRepo authdomain.Repository, userRepo userdomain.Repository, tokenManager *sharedauth.TokenManager) *UseCase {
	return &UseCase{
		authRepo:     authRepo,
		userRepo:     userRepo,
		tokenManager: tokenManager,
	}
}

func (uc *UseCase) Execute(ctx context.Context, input Input) (*Output, error) {
	user, err := uc.userRepo.GetByLogin(ctx, input.Login)
	if err != nil {
		if err == userdomain.ErrUserNotFound {
			return nil, authdomain.ErrInvalidCredentials
		}

		return nil, err
	}
	if user.Password != input.Password {
		return nil, authdomain.ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, authdomain.ErrInactiveUser
	}

	accessToken, _, err := uc.tokenManager.GenerateAccessToken(user.ID, user.Login, string(user.Role))
	if err != nil {
		return nil, err
	}

	refreshToken, refreshExpiresAt, err := uc.tokenManager.GenerateRefreshToken(user.ID, user.Login, string(user.Role))
	if err != nil {
		return nil, err
	}

	if _, err := uc.authRepo.CreateRefreshSession(ctx, &authdomain.RefreshSession{
		UserID:    user.ID,
		TokenHash: sharedauth.HashToken(refreshToken),
		ExpiresAt: refreshExpiresAt,
	}); err != nil {
		return nil, err
	}

	return &Output{
		User: UserOutput{
			ID:        user.ID,
			Login:     user.Login,
			Password:  user.Password,
			Role:      string(user.Role),
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
