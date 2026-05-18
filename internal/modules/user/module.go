package user

import (
	"github.com/gin-gonic/gin"

	"project/internal/modules/user/domain"
	userhttp "project/internal/modules/user/transport/http"
	"project/internal/modules/user/usecase/createuser"
	"project/internal/modules/user/usecase/deleteuser"
	"project/internal/modules/user/usecase/getuser"
	"project/internal/modules/user/usecase/listusers"
	"project/internal/modules/user/usecase/updateuser"
)

type Module struct {
	handler *userhttp.Handler
}

func NewModule(repo domain.Repository) *Module {
	createUser := createuser.New(repo)
	listUsers := listusers.New(repo)
	getUser := getuser.New(repo)
	updateUser := updateuser.New(repo)
	deleteUser := deleteuser.New(repo)
	handler := userhttp.NewHandler(createUser, listUsers, getUser, updateUser, deleteUser)

	return &Module{
		handler: handler,
	}
}

func (m *Module) RegisterRoutes(router *gin.Engine) {
	userhttp.RegisterRoutes(router, m.handler)
}

func (m *Module) RegisterRoutesWithMiddleware(router *gin.Engine, middlewares ...gin.HandlerFunc) {
	userhttp.RegisterRoutes(router, m.handler, middlewares...)
}
