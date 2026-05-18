package auth

import (
	"github.com/gin-gonic/gin"

	authdomain "project/internal/modules/auth/domain"
	authhttp "project/internal/modules/auth/transport/http"
	"project/internal/modules/auth/usecase/login"
	"project/internal/modules/auth/usecase/logout"
	"project/internal/modules/auth/usecase/refresh"
	userdomain "project/internal/modules/user/domain"
	sharedauth "project/internal/shared/auth"
)

type Module struct {
	handler *authhttp.Handler
}

func NewModule(authRepo authdomain.Repository, userRepo userdomain.Repository, tokenManager *sharedauth.TokenManager) *Module {
	loginUseCase := login.New(authRepo, userRepo, tokenManager)
	refreshUseCase := refresh.New(authRepo, userRepo, tokenManager)
	logoutUseCase := logout.New(authRepo, tokenManager)
	handler := authhttp.NewHandler(loginUseCase, refreshUseCase, logoutUseCase)

	return &Module{
		handler: handler,
	}
}

func (m *Module) RegisterRoutes(router *gin.Engine) {
	authhttp.RegisterRoutes(router, m.handler)
}
