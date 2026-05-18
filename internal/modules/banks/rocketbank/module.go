package rocketbank

import (
	"github.com/gin-gonic/gin"

	"project/internal/modules/banks/rocketbank/domain"
	rocketbankhttp "project/internal/modules/banks/rocketbank/transport/http"
	"project/internal/modules/banks/rocketbank/usecase/clearhistory"
	"project/internal/modules/banks/rocketbank/usecase/createcardtransfer"
	"project/internal/modules/banks/rocketbank/usecase/createcashtransfer"
	"project/internal/modules/banks/rocketbank/usecase/createsbptransfer"
	"project/internal/modules/banks/rocketbank/usecase/deletehistoryitem"
	"project/internal/modules/banks/rocketbank/usecase/getconfig"
	"project/internal/modules/banks/rocketbank/usecase/gethistoryitem"
	"project/internal/modules/banks/rocketbank/usecase/listhistory"
	"project/internal/modules/banks/rocketbank/usecase/updatebalance"
	"project/internal/modules/banks/rocketbank/usecase/updatecardtransfer"
	"project/internal/modules/banks/rocketbank/usecase/updatecashtransfer"
	"project/internal/modules/banks/rocketbank/usecase/updateclientinfo"
	"project/internal/modules/banks/rocketbank/usecase/updatesbptransfer"
)

type Module struct {
	handler *rocketbankhttp.Handler
}

func NewModule(repo domain.Repository, chequeGenerator domain.ChequeGenerator) *Module {
	getConfig := getconfig.New(repo)
	updateBalance := updatebalance.New(repo)
	updateClientInfo := updateclientinfo.New(repo)
	listHistory := listhistory.New(repo)
	getHistoryItem := gethistoryitem.New(repo)
	createCardTransfer := createcardtransfer.New(repo)
	updateCardTransfer := updatecardtransfer.New(repo)
	createCashTransfer := createcashtransfer.New(repo)
	updateCashTransfer := updatecashtransfer.New(repo)
	createSBPTransfer := createsbptransfer.New(repo)
	updateSBPTransfer := updatesbptransfer.New(repo)
	deleteHistoryItem := deletehistoryitem.New(repo)
	clearHistory := clearhistory.New(repo)
	handler := rocketbankhttp.NewHandler(
		getConfig,
		updateBalance,
		updateClientInfo,
		listHistory,
		getHistoryItem,
		createCardTransfer,
		updateCardTransfer,
		createCashTransfer,
		updateCashTransfer,
		createSBPTransfer,
		updateSBPTransfer,
		deleteHistoryItem,
		clearHistory,
		chequeGenerator,
	)

	return &Module{handler: handler}
}

func (m *Module) RegisterRoutes(router *gin.Engine, middlewares ...gin.HandlerFunc) {
	rocketbankhttp.RegisterRoutes(router, m.handler, middlewares...)
}
