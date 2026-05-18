package refresh

import (
	"context"
	"time"

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
	claims, err := uc.tokenManager.ParseRefreshToken(input.RefreshToken)
	if err != nil {
		return nil, authdomain.ErrInvalidToken
	}

	tokenHash := sharedauth.HashToken(input.RefreshToken)
	session, err := uc.authRepo.GetRefreshSessionByHash(ctx, tokenHash)
	if err != nil {
		return nil, authdomain.ErrInvalidToken
	}
	if session.UserID != claims.UserID {
		return nil, authdomain.ErrInvalidToken
	}
	if session.RevokedAt != nil {
		return nil, authdomain.ErrRefreshRevoked
	}
	if time.Now().UTC().After(session.ExpiresAt) {
		return nil, authdomain.ErrInvalidToken
	}

	user, err := uc.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		if err == userdomain.ErrUserNotFound {
			return nil, authdomain.ErrInvalidToken
		}

		return nil, err
	}
	if !user.IsActive {
		return nil, authdomain.ErrInactiveUser
	}

	if err := uc.authRepo.RevokeRefreshSession(ctx, tokenHash); err != nil {
		return nil, err
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
