package catalog

import (
	"github.com/gin-gonic/gin"

	"project/internal/modules/banks/catalog/domain"
	cataloghttp "project/internal/modules/banks/catalog/transport/http"
	"project/internal/modules/banks/catalog/usecase/listbanks"
)

type Module struct {
	handler *cataloghttp.Handler
}

func NewModule(repo domain.Repository) *Module {
	listBanks := listbanks.New(repo)
	handler := cataloghttp.NewHandler(listBanks)

	return &Module{handler: handler}
}

func (m *Module) RegisterRoutes(router *gin.Engine, middlewares ...gin.HandlerFunc) {
	cataloghttp.RegisterRoutes(router, m.handler, middlewares...)
}
