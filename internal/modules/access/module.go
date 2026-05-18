package access

import (
	"github.com/gin-gonic/gin"

	"project/internal/modules/access/domain"
	accesshttp "project/internal/modules/access/transport/http"
	"project/internal/modules/access/usecase/grantaccess"
	"project/internal/modules/access/usecase/listmyaccesses"
	"project/internal/modules/access/usecase/revokeaccess"
	bankdomain "project/internal/modules/banks/catalog/domain"
	userdomain "project/internal/modules/user/domain"
)

type Module struct {
	handler *accesshttp.Handler
}

func NewModule(accessRepo domain.Repository, bankRepo bankdomain.Repository, userRepo userdomain.Repository) *Module {
	listMyAccesses := listmyaccesses.New(accessRepo)
	grantAccess := grantaccess.New(accessRepo, bankRepo, userRepo)
	revokeAccess := revokeaccess.New(accessRepo)
	handler := accesshttp.NewHandler(listMyAccesses, grantAccess, revokeAccess)

	return &Module{handler: handler}
}

func (m *Module) RegisterRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc, adminMiddleware gin.HandlerFunc) {
	accesshttp.RegisterRoutes(router, m.handler, authMiddleware, adminMiddleware)
}
