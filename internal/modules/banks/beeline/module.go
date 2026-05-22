package beeline

import (
	"github.com/gin-gonic/gin"

	"project/internal/modules/banks/beeline/domain"
	beelinehttp "project/internal/modules/banks/beeline/transport/http"
	"project/internal/modules/banks/beeline/usecase/getconfig"
)

type Module struct {
	handler *beelinehttp.Handler
}

func NewModule(repo domain.Repository) *Module {
	getConfig := getconfig.New(repo)
	handler := beelinehttp.NewHandler(getConfig)

	return &Module{handler: handler}
}

func (m *Module) RegisterRoutes(router *gin.Engine, middlewares ...gin.HandlerFunc) {
	beelinehttp.RegisterRoutes(router, m.handler, middlewares...)
}
