package beeline

import (
	"github.com/gin-gonic/gin"

	"project/internal/modules/banks/beeline/domain"
	beelinehttp "project/internal/modules/banks/beeline/transport/http"
	"project/internal/modules/banks/beeline/usecase/createpayment"
	"project/internal/modules/banks/beeline/usecase/createsim"
	"project/internal/modules/banks/beeline/usecase/deletepayment"
	"project/internal/modules/banks/beeline/usecase/deletesim"
	"project/internal/modules/banks/beeline/usecase/getconfig"
	"project/internal/modules/banks/beeline/usecase/getdetalization"
	"project/internal/modules/banks/beeline/usecase/getpayment"
	"project/internal/modules/banks/beeline/usecase/getsim"
	"project/internal/modules/banks/beeline/usecase/hidedetalizationtransaction"
	"project/internal/modules/banks/beeline/usecase/listpayments"
	"project/internal/modules/banks/beeline/usecase/listsims"
	"project/internal/modules/banks/beeline/usecase/updatepayment"
)

type Module struct {
	handler *beelinehttp.Handler
}

func NewModule(repo domain.Repository) *Module {
	handler := beelinehttp.NewHandler(
		listsims.New(repo),
		getsim.New(repo),
		createsim.New(repo),
		deletesim.New(repo),
		getconfig.New(repo),
		getdetalization.New(repo),
		hidedetalizationtransaction.New(repo),
		listpayments.New(repo),
		getpayment.New(repo),
		createpayment.New(repo),
		updatepayment.New(repo),
		deletepayment.New(repo),
	)

	return &Module{handler: handler}
}

func (m *Module) RegisterRoutes(router *gin.Engine, middlewares ...gin.HandlerFunc) {
	beelinehttp.RegisterRoutes(router, m.handler, middlewares...)
}
