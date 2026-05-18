package http

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"project/internal/app/logger"
	"project/internal/modules/banks/rocketbank/domain"
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

var handlerLog = logger.New("rocketbank-http")

type Handler struct {
	getConfig          *getconfig.UseCase
	updateBalance      *updatebalance.UseCase
	updateClientInfo   *updateclientinfo.UseCase
	listHistory        *listhistory.UseCase
	getHistoryItem     *gethistoryitem.UseCase
	createCardTransfer *createcardtransfer.UseCase
	updateCardTransfer *updatecardtransfer.UseCase
	createCashTransfer *createcashtransfer.UseCase
	updateCashTransfer *updatecashtransfer.UseCase
	createSBPTransfer  *createsbptransfer.UseCase
	updateSBPTransfer  *updatesbptransfer.UseCase
	deleteHistoryItem  *deletehistoryitem.UseCase
	clearHistory       *clearhistory.UseCase
	chequeGenerator    domain.ChequeGenerator
}

func NewHandler(
	getConfig *getconfig.UseCase,
	updateBalance *updatebalance.UseCase,
	updateClientInfo *updateclientinfo.UseCase,
	listHistory *listhistory.UseCase,
	getHistoryItem *gethistoryitem.UseCase,
	createCardTransfer *createcardtransfer.UseCase,
	updateCardTransfer *updatecardtransfer.UseCase,
	createCashTransfer *createcashtransfer.UseCase,
	updateCashTransfer *updatecashtransfer.UseCase,
	createSBPTransfer *createsbptransfer.UseCase,
	updateSBPTransfer *updatesbptransfer.UseCase,
	deleteHistoryItem *deletehistoryitem.UseCase,
	clearHistory *clearhistory.UseCase,
	chequeGenerator domain.ChequeGenerator,
) *Handler {
	return &Handler{
		getConfig:          getConfig,
		updateBalance:      updateBalance,
		updateClientInfo:   updateClientInfo,
		listHistory:        listHistory,
		getHistoryItem:     getHistoryItem,
		createCardTransfer: createCardTransfer,
		updateCardTransfer: updateCardTransfer,
		createCashTransfer: createCashTransfer,
		updateCashTransfer: updateCashTransfer,
		createSBPTransfer:  createSBPTransfer,
		updateSBPTransfer:  updateSBPTransfer,
		deleteHistoryItem:  deleteHistoryItem,
		clearHistory:       clearHistory,
		chequeGenerator:    chequeGenerator,
	}
}

type UpdateBalanceRequest struct {
	Balance *float64 `json:"balance" binding:"required" example:"99999"`
}

type UpdateClientInfoRequest struct {
	FirstName   *string `json:"firstName" binding:"required" example:"Иван"`
	MiddleName  *string `json:"middleName" binding:"required" example:"Иванович"`
	LastName    *string `json:"lastName" binding:"required" example:"Иванов"`
	PhoneNumber *string `json:"phoneNumber" binding:"required" example:"+79001234567"`
	CardNumber  *string `json:"cardNumber" binding:"required" example:"40817810000000000000"`
}

type CashTransferRequest struct {
	Amount        *float64 `json:"amount" binding:"required" example:"8000"`
	BalanceBefore *float64 `json:"balanceBefore" binding:"required" example:"22096.74"`
	Direction     string   `json:"direction" binding:"required" enums:"OUTGOING,INCOMING" example:"OUTGOING"`
	Time          string   `json:"time" binding:"required" example:"2026-05-02T11:08:52+0700"`
}

type UpdateCashTransferRequest struct {
	Amount        *float64 `json:"amount,omitempty" example:"8000"`
	BalanceBefore *float64 `json:"balanceBefore,omitempty" example:"22096.74"`
	Direction     *string  `json:"direction,omitempty" enums:"OUTGOING,INCOMING" example:"OUTGOING"`
	Time          *string  `json:"time,omitempty" example:"2026-05-02T11:08:52+0700"`
}

type CardTransferRequest struct {
	Amount              *float64 `json:"amount" binding:"required" example:"5000"`
	BalanceBefore       *float64 `json:"balanceBefore" binding:"required" example:"22096.74"`
	Direction           string   `json:"direction" binding:"required" enums:"OUTGOING,INCOMING" example:"OUTGOING"`
	Time                string   `json:"time" binding:"required" example:"2026-05-16T22:14:12+0700"`
	RecipientCardNumber string   `json:"recipientCardNumber" binding:"required" example:"1234 5678 9100 0000"`
	BankID              string   `json:"bankId" binding:"required" example:"tbank"`
}

type UpdateCardTransferRequest struct {
	Amount              *float64 `json:"amount,omitempty" example:"5000"`
	BalanceBefore       *float64 `json:"balanceBefore,omitempty" example:"22096.74"`
	Direction           *string  `json:"direction,omitempty" enums:"OUTGOING,INCOMING" example:"OUTGOING"`
	Time                *string  `json:"time,omitempty" example:"2026-05-16T22:14:12+0700"`
	RecipientCardNumber *string  `json:"recipientCardNumber,omitempty" example:"1234 5678 9100 0000"`
	BankID              *string  `json:"bankId,omitempty" example:"tbank"`
}

type SBPTransferRequest struct {
	Amount              *float64 `json:"amount" binding:"required" example:"50"`
	BalanceBefore       *float64 `json:"balanceBefore" binding:"required" example:"190"`
	Direction           string   `json:"direction" binding:"required" enums:"OUTGOING,INCOMING" example:"OUTGOING"`
	Time                string   `json:"time" binding:"required" example:"2026-05-16T22:14:12+0700"`
	OperationFirstName  string   `json:"operationFirstName" binding:"required" example:"Азат"`
	OperationMiddleName string   `json:"operationMiddleName" binding:"required" example:"Аликович"`
	OperationLastName   string   `json:"operationLastName" binding:"required" example:"Гайнутдинов"`
	BankID              string   `json:"bankId" binding:"required" example:"tbank"`
	PhoneNumber         string   `json:"phoneNumber,omitempty" example:"+79099334005"`
}

type UpdateSBPTransferRequest struct {
	Amount              *float64 `json:"amount,omitempty" example:"50"`
	BalanceBefore       *float64 `json:"balanceBefore,omitempty" example:"190"`
	Direction           *string  `json:"direction,omitempty" enums:"OUTGOING,INCOMING" example:"OUTGOING"`
	Time                *string  `json:"time,omitempty" example:"2026-05-16T22:14:12+0700"`
	OperationFirstName  *string  `json:"operationFirstName,omitempty" example:"Азат"`
	OperationMiddleName *string  `json:"operationMiddleName,omitempty" example:"Аликович"`
	OperationLastName   *string  `json:"operationLastName,omitempty" example:"Гайнутдинов"`
	BankID              *string  `json:"bankId,omitempty" example:"tbank"`
	PhoneNumber         *string  `json:"phoneNumber,omitempty" example:"+79099334005"`
}

// GetConfig godoc
// @Summary Get Rocketbank config
// @Description Returns Rocketbank config.
// @Tags rocketbank config
// @Produce json
// @Success 200 {object} ConfigResponse
// @Failure 500 {object} RocketbankErrorResponse
// @Router /banks/rocketbank/config [get]
func (h *Handler) GetConfig(c *gin.Context) {
	out, err := h.getConfig.Execute(c.Request.Context(), getconfig.Input{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Balance:    out.Balance,
		ClientInfo: clientInfoResponse(out.ClientInfo),
		History:    historyResponse(out.History),
		CreatedAt:  out.CreatedAt,
		UpdatedAt:  out.UpdatedAt,
	})
}

// UpdateBalance godoc
// @Summary Update Rocketbank balance
// @Description Updates only Rocketbank balance. Full config updates are not allowed.
// @Tags rocketbank config
// @Accept json
// @Produce json
// @Param input body UpdateBalanceRequest true "Balance update payload"
// @Success 200 {object} ConfigResponse
// @Failure 400 {object} RocketbankErrorResponse
// @Failure 500 {object} RocketbankErrorResponse
// @Router /banks/rocketbank/config/balance [patch]
func (h *Handler) UpdateBalance(c *gin.Context) {
	var req UpdateBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	out, err := h.updateBalance.Execute(c.Request.Context(), updatebalance.Input{
		Balance: req.Balance,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Balance:    out.Balance,
		ClientInfo: clientInfoResponse(out.ClientInfo),
		History:    historyResponse(out.History),
		CreatedAt:  out.CreatedAt,
		UpdatedAt:  out.UpdatedAt,
	})
}

// UpdateClientInfo godoc
// @Summary Update Rocketbank client info
// @Description Updates only Rocketbank client info. Full config updates are not allowed.
// @Tags rocketbank client info
// @Accept json
// @Produce json
// @Param input body UpdateClientInfoRequest true "Client info update payload"
// @Success 200 {object} ConfigResponse
// @Failure 400 {object} RocketbankErrorResponse
// @Failure 500 {object} RocketbankErrorResponse
// @Router /banks/rocketbank/config/client-info [patch]
func (h *Handler) UpdateClientInfo(c *gin.Context) {
	var req UpdateClientInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	out, err := h.updateClientInfo.Execute(c.Request.Context(), updateclientinfo.Input{
		ClientInfo: domain.ClientInfo{
			FirstName:   req.FirstName,
			MiddleName:  req.MiddleName,
			LastName:    req.LastName,
			PhoneNumber: req.PhoneNumber,
			CardNumber:  req.CardNumber,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Balance:    out.Balance,
		ClientInfo: clientInfoResponse(out.ClientInfo),
		History:    historyResponse(out.History),
		CreatedAt:  out.CreatedAt,
		UpdatedAt:  out.UpdatedAt,
	})
}

// ListHistory godoc
// @Summary List Rocketbank history
// @Description Returns Rocketbank configured history items.
// @Tags rocketbank history
// @Produce json
// @Success 200 {array} HistoryResponse
// @Failure 500 {object} RocketbankErrorResponse
// @Router /banks/rocketbank/config/history [get]
func (h *Handler) ListHistory(c *gin.Context) {
	out, err := h.listHistory.Execute(c.Request.Context(), listhistory.Input{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, historyResponse(out.History))
}

// GetHistoryItem godoc
// @Summary Get Rocketbank history item
// @Description Returns Rocketbank configured history item by transaction id.
// @Tags rocketbank history
// @Produce json
// @Param id path string true "History transaction id"
// @Success 200 {object} HistoryResponse
// @Failure 404 {object} RocketbankErrorResponse
// @Failure 500 {object} RocketbankErrorResponse
// @Router /banks/rocketbank/config/history/items/{id} [get]
func (h *Handler) GetHistoryItem(c *gin.Context) {
	out, err := h.getHistoryItem.Execute(c.Request.Context(), gethistoryitem.Input{
		ID: c.Param("id"),
	})
	if errors.Is(err, domain.ErrHistoryItemNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "history item not found"})
		return
	}
	if errors.Is(err, domain.ErrHistoryItemExists) {
		c.JSON(http.StatusConflict, gin.H{"error": "history item already exists"})
		return
	}
	if errors.Is(err, domain.ErrInsufficientBalance) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "balanceBefore cannot be less than amount for OUTGOING direction"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, historyItemResponse(out.Item))
}

// CreateCashTransfer godoc
// @Summary Create Rocketbank cash transfer history item
// @Description Creates a Rocketbank cash transfer history item.
// @Tags rocketbank cash transfer
// @Accept json
// @Produce json
// @Param input body CashTransferRequest true "Cash transfer payload"
// @Success 201 {object} HistoryResponse
// @Failure 400 {object} RocketbankErrorResponse
// @Failure 409 {object} RocketbankErrorResponse
// @Failure 500 {object} RocketbankErrorResponse
// @Router /banks/rocketbank/config/history/cash-transfer [post]
func (h *Handler) CreateCashTransfer(c *gin.Context) {
	req, ok := bindCashTransferRequest(c)
	if !ok {
		return
	}

	out, err := h.createCashTransfer.Execute(c.Request.Context(), createcashtransfer.Input{
		Amount:        *req.Amount,
		BalanceBefore: *req.BalanceBefore,
		Direction:     req.Direction,
		Time:          req.Time,
	})
	if errors.Is(err, domain.ErrHistoryItemExists) {
		c.JSON(http.StatusConflict, gin.H{"error": "history item already exists"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	h.generateSBPTransferCheque(c, out.Item)

	c.JSON(http.StatusCreated, historyItemResponse(out.Item))
}

// CreateSBPTransfer godoc
// @Summary Create Rocketbank SBP transfer history item
// @Description Creates a Rocketbank SBP transfer history item.
// @Tags rocketbank sbp transfer
// @Accept json
// @Produce json
// @Param input body SBPTransferRequest true "SBP transfer payload"
// @Success 201 {object} HistoryResponse
// @Failure 400 {object} RocketbankErrorResponse
// @Failure 409 {object} RocketbankErrorResponse
// @Failure 500 {object} RocketbankErrorResponse
// @Router /banks/rocketbank/config/history/sbp-transfer [post]
func (h *Handler) CreateSBPTransfer(c *gin.Context) {
	req, ok := bindSBPTransferRequest(c)
	if !ok {
		return
	}

	out, err := h.createSBPTransfer.Execute(c.Request.Context(), createsbptransfer.Input{
		Amount:              *req.Amount,
		BalanceBefore:       *req.BalanceBefore,
		Direction:           req.Direction,
		Time:                req.Time,
		OperationFirstName:  req.OperationFirstName,
		OperationMiddleName: req.OperationMiddleName,
		OperationLastName:   req.OperationLastName,
		BankID:              req.BankID,
		PhoneNumber:         req.PhoneNumber,
	})
	if errors.Is(err, domain.ErrHistoryItemExists) {
		c.JSON(http.StatusConflict, gin.H{"error": "history item already exists"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	h.generateSBPTransferCheque(c, out.Item)

	c.JSON(http.StatusCreated, historyItemResponse(out.Item))
}

// CreateCardTransfer godoc
// @Summary Create Rocketbank card transfer history item
// @Description Creates a Rocketbank card transfer history item.
// @Tags rocketbank card transfer
// @Accept json
// @Produce json
// @Param input body CardTransferRequest true "Card transfer payload"
// @Success 201 {object} HistoryResponse
// @Failure 400 {object} RocketbankErrorResponse
// @Failure 409 {object} RocketbankErrorResponse
// @Failure 500 {object} RocketbankErrorResponse
// @Router /banks/rocketbank/config/history/card-transfer [post]
func (h *Handler) CreateCardTransfer(c *gin.Context) {
	req, ok := bindCardTransferRequest(c)
	if !ok {
		return
	}

	out, err := h.createCardTransfer.Execute(c.Request.Context(), createcardtransfer.Input{
		Amount:              *req.Amount,
		BalanceBefore:       *req.BalanceBefore,
		Direction:           req.Direction,
		Time:                req.Time,
		BankID:              req.BankID,
		RecipientCardNumber: req.RecipientCardNumber,
	})
	if errors.Is(err, domain.ErrHistoryItemExists) {
		c.JSON(http.StatusConflict, gin.H{"error": "history item already exists"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	h.generateCardTransferCheque(c, out.Item)

	c.JSON(http.StatusCreated, historyItemResponse(out.Item))
}

// UpdateCashTransfer godoc
// @Summary Update Rocketbank cash transfer history item
// @Description Updates a Rocketbank cash transfer history item.
// @Tags rocketbank cash transfer
// @Accept json
// @Produce json
// @Param id path string true "History transaction id"
// @Param input body UpdateCashTransferRequest true "Cash transfer patch payload"
// @Success 200 {object} HistoryResponse
// @Failure 400 {object} RocketbankErrorResponse
// @Failure 404 {object} RocketbankErrorResponse
// @Failure 409 {object} RocketbankErrorResponse
// @Failure 500 {object} RocketbankErrorResponse
// @Router /banks/rocketbank/config/history/items/{id}/cash-transfer [patch]
func (h *Handler) UpdateCashTransfer(c *gin.Context) {
	req, ok := bindUpdateCashTransferRequest(c)
	if !ok {
		return
	}

	out, err := h.updateCashTransfer.Execute(c.Request.Context(), updatecashtransfer.Input{
		ID:            c.Param("id"),
		Amount:        req.Amount,
		BalanceBefore: req.BalanceBefore,
		Direction:     req.Direction,
		Time:          req.Time,
	})
	if errors.Is(err, domain.ErrHistoryItemNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "history item not found"})
		return
	}
	if errors.Is(err, domain.ErrHistoryItemExists) {
		c.JSON(http.StatusConflict, gin.H{"error": "history item already exists"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	h.generateSBPTransferCheque(c, out.Item)

	c.JSON(http.StatusOK, historyItemResponse(out.Item))
}

// UpdateCardTransfer godoc
// @Summary Update Rocketbank card transfer history item
// @Description Updates a Rocketbank card transfer history item.
// @Tags rocketbank card transfer
// @Accept json
// @Produce json
// @Param id path string true "History transaction id"
// @Param input body UpdateCardTransferRequest true "Card transfer patch payload"
// @Success 200 {object} HistoryResponse
// @Failure 400 {object} RocketbankErrorResponse
// @Failure 404 {object} RocketbankErrorResponse
// @Failure 409 {object} RocketbankErrorResponse
// @Failure 500 {object} RocketbankErrorResponse
// @Router /banks/rocketbank/config/history/items/{id}/card-transfer [patch]
func (h *Handler) UpdateCardTransfer(c *gin.Context) {
	req, ok := bindUpdateCardTransferRequest(c)
	if !ok {
		return
	}

	out, err := h.updateCardTransfer.Execute(c.Request.Context(), updatecardtransfer.Input{
		ID:                  c.Param("id"),
		Amount:              req.Amount,
		BalanceBefore:       req.BalanceBefore,
		Direction:           req.Direction,
		Time:                req.Time,
		BankID:              req.BankID,
		RecipientCardNumber: req.RecipientCardNumber,
	})
	if errors.Is(err, domain.ErrHistoryItemNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "history item not found"})
		return
	}
	if errors.Is(err, domain.ErrHistoryItemExists) {
		c.JSON(http.StatusConflict, gin.H{"error": "history item already exists"})
		return
	}
	if errors.Is(err, domain.ErrInsufficientBalance) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "balanceBefore cannot be less than amount"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	h.generateCardTransferCheque(c, out.Item)

	c.JSON(http.StatusOK, historyItemResponse(out.Item))
}

// UpdateSBPTransfer godoc
// @Summary Update Rocketbank SBP transfer history item
// @Description Updates a Rocketbank SBP transfer history item.
// @Tags rocketbank sbp transfer
// @Accept json
// @Produce json
// @Param id path string true "History transaction id"
// @Param input body UpdateSBPTransferRequest true "SBP transfer patch payload"
// @Success 200 {object} HistoryResponse
// @Failure 400 {object} RocketbankErrorResponse
// @Failure 404 {object} RocketbankErrorResponse
// @Failure 409 {object} RocketbankErrorResponse
// @Failure 500 {object} RocketbankErrorResponse
// @Router /banks/rocketbank/config/history/items/{id}/sbp-transfer [patch]
func (h *Handler) UpdateSBPTransfer(c *gin.Context) {
	req, ok := bindUpdateSBPTransferRequest(c)
	if !ok {
		return
	}

	out, err := h.updateSBPTransfer.Execute(c.Request.Context(), updatesbptransfer.Input{
		ID:                  c.Param("id"),
		Amount:              req.Amount,
		BalanceBefore:       req.BalanceBefore,
		Direction:           req.Direction,
		Time:                req.Time,
		OperationFirstName:  req.OperationFirstName,
		OperationMiddleName: req.OperationMiddleName,
		OperationLastName:   req.OperationLastName,
		BankID:              req.BankID,
		PhoneNumber:         req.PhoneNumber,
	})
	if errors.Is(err, domain.ErrHistoryItemNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "history item not found"})
		return
	}
	if errors.Is(err, domain.ErrHistoryItemExists) {
		c.JSON(http.StatusConflict, gin.H{"error": "history item already exists"})
		return
	}
	if errors.Is(err, domain.ErrInsufficientBalance) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "balanceBefore cannot be less than amount for OUTGOING direction"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, historyItemResponse(out.Item))
}

// DeleteHistoryItem godoc
// @Summary Delete Rocketbank history item
// @Description Deletes Rocketbank configured history item by transaction id.
// @Tags rocketbank history
// @Produce json
// @Param id path string true "History transaction id"
// @Success 204
// @Failure 404 {object} RocketbankErrorResponse
// @Failure 500 {object} RocketbankErrorResponse
// @Router /banks/rocketbank/config/history/items/{id} [delete]
func (h *Handler) DeleteHistoryItem(c *gin.Context) {
	_, err := h.deleteHistoryItem.Execute(c.Request.Context(), deletehistoryitem.Input{
		ID: c.Param("id"),
	})
	if errors.Is(err, domain.ErrHistoryItemNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "history item not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

// ClearHistory godoc
// @Summary Clear Rocketbank history
// @Description Deletes all configured Rocketbank history items.
// @Tags rocketbank history
// @Produce json
// @Success 204
// @Failure 500 {object} RocketbankErrorResponse
// @Router /banks/rocketbank/config/history [delete]
func (h *Handler) ClearHistory(c *gin.Context) {
	if _, err := h.clearHistory.Execute(c.Request.Context(), clearhistory.Input{}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

func clientInfoResponse(clientInfo domain.ClientInfo) ClientInfoResponse {
	return ClientInfoResponse{
		FirstName:   clientInfo.FirstName,
		MiddleName:  clientInfo.MiddleName,
		LastName:    clientInfo.LastName,
		PhoneNumber: clientInfo.PhoneNumber,
		CardNumber:  clientInfo.CardNumber,
	}
}

func (h *Handler) generateSBPTransferCheque(c *gin.Context, item domain.HistoryItem) {
	if h.chequeGenerator == nil || item.Type != domain.HistoryTypeSBPTransfer {
		return
	}

	config, err := h.getConfig.Execute(c.Request.Context(), getconfig.Input{})
	if err != nil {
		handlerLog.Warnf("generate sbp cheque skipped: read config failed: transactionId=%s err=%v", domain.HistoryItemID(item), err)
		return
	}

	if err := h.chequeGenerator.GenerateSBPTransferCheque(item, config.ClientInfo); err != nil {
		handlerLog.Warnf("generate sbp cheque failed: transactionId=%s err=%v", domain.HistoryItemID(item), err)
	}
}

func (h *Handler) generateCardTransferCheque(c *gin.Context, item domain.HistoryItem) {
	if h.chequeGenerator == nil || item.Type != domain.HistoryTypeCardTransfer {
		return
	}

	config, err := h.getConfig.Execute(c.Request.Context(), getconfig.Input{})
	if err != nil {
		handlerLog.Warnf("generate card cheque skipped: read config failed: transactionId=%s err=%v", domain.HistoryItemID(item), err)
		return
	}

	if err := h.chequeGenerator.GenerateCardTransferCheque(item, config.ClientInfo); err != nil {
		handlerLog.Warnf("generate card cheque failed: transactionId=%s err=%v", domain.HistoryItemID(item), err)
	}
}

func historyResponse(history []domain.HistoryItem) []HistoryResponse {
	if history == nil {
		return []HistoryResponse{}
	}

	response := make([]HistoryResponse, 0, len(history))
	for _, item := range history {
		response = append(response, historyItemResponse(item))
	}

	return response
}

func historyItemResponse(item domain.HistoryItem) HistoryResponse {
	return HistoryResponse{
		ID:                  domain.HistoryItemID(item),
		Type:                item.Type,
		Amount:              item.Amount,
		BalanceBefore:       item.BalanceBefore,
		Direction:           item.Direction,
		Time:                item.Time,
		OperationFirstName:  item.OperationFirstName,
		OperationMiddleName: item.OperationMiddleName,
		OperationLastName:   item.OperationLastName,
		BankID:              item.BankID,
		PhoneNumber:         item.PhoneNumber,
		RecipientCardNumber: item.RecipientCardNumber,
	}
}

func bindCashTransferRequest(c *gin.Context) (CashTransferRequest, bool) {
	var req CashTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return req, false
	}
	if !domain.IsValidHistoryDirection(req.Direction) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "direction must be OUTGOING or INCOMING"})
		return req, false
	}
	if req.Amount == nil || *req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be greater than 0"})
		return req, false
	}
	if req.BalanceBefore == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "balanceBefore is required"})
		return req, false
	}
	if !domain.IsValidCashTransferBalance(*req.Amount, *req.BalanceBefore, req.Direction) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "balanceBefore cannot be less than amount for OUTGOING direction"})
		return req, false
	}
	if _, err := time.Parse(domain.HistoryTimeLayout, req.Time); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "time must match format 2026-05-02T11:08:52+0700"})
		return req, false
	}

	return req, true
}

func bindCardTransferRequest(c *gin.Context) (CardTransferRequest, bool) {
	var req CardTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return req, false
	}
	if !domain.IsValidHistoryDirection(req.Direction) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "direction must be OUTGOING or INCOMING"})
		return req, false
	}
	if req.Amount == nil || *req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be greater than 0"})
		return req, false
	}
	if req.BalanceBefore == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "balanceBefore is required"})
		return req, false
	}
	if !domain.IsValidCashTransferBalance(*req.Amount, *req.BalanceBefore, req.Direction) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "balanceBefore cannot be less than amount for OUTGOING direction"})
		return req, false
	}
	if _, err := time.Parse(domain.HistoryTimeLayout, req.Time); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "time must match format 2026-05-02T11:08:52+0700"})
		return req, false
	}
	if !isValidRecipientCardNumber(req.RecipientCardNumber) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "recipientCardNumber must match format 1234 5678 9100 0000"})
		return req, false
	}
	if !domain.IsValidCardTransferBankID(req.BankID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bankId is invalid"})
		return req, false
	}

	return req, true
}

func bindUpdateCardTransferRequest(c *gin.Context) (UpdateCardTransferRequest, bool) {
	var req UpdateCardTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return req, false
	}
	if req.Amount == nil &&
		req.BalanceBefore == nil &&
		req.Direction == nil &&
		req.Time == nil &&
		req.RecipientCardNumber == nil &&
		req.BankID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one field must be provided"})
		return req, false
	}
	if req.Amount != nil && *req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be greater than 0"})
		return req, false
	}
	if req.Direction != nil && !domain.IsValidHistoryDirection(*req.Direction) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "direction must be OUTGOING or INCOMING"})
		return req, false
	}
	if req.Time != nil {
		if _, err := time.Parse(domain.HistoryTimeLayout, *req.Time); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "time must match format 2026-05-02T11:08:52+0700"})
			return req, false
		}
	}
	if req.RecipientCardNumber != nil && !isValidRecipientCardNumber(*req.RecipientCardNumber) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "recipientCardNumber must match format 1234 5678 9100 0000"})
		return req, false
	}
	if req.BankID != nil && !domain.IsValidCardTransferBankID(*req.BankID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bankId is invalid"})
		return req, false
	}

	return req, true
}

func bindUpdateCashTransferRequest(c *gin.Context) (UpdateCashTransferRequest, bool) {
	var req UpdateCashTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return req, false
	}
	if req.Amount == nil && req.BalanceBefore == nil && req.Direction == nil && req.Time == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one field must be provided"})
		return req, false
	}
	if req.Direction != nil && !domain.IsValidHistoryDirection(*req.Direction) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "direction must be OUTGOING or INCOMING"})
		return req, false
	}
	if req.Amount != nil && *req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be greater than 0"})
		return req, false
	}
	if req.Time != nil {
		if _, err := time.Parse(domain.HistoryTimeLayout, *req.Time); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "time must match format 2026-05-02T11:08:52+0700"})
			return req, false
		}
	}

	return req, true
}

func isValidRecipientCardNumber(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) != len("1234 5678 9100 0000") {
		return false
	}

	for index := 0; index < len(value); index++ {
		char := value[index]
		if index == 4 || index == 9 || index == 14 {
			if char != ' ' {
				return false
			}
			continue
		}
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

func bindSBPTransferRequest(c *gin.Context) (SBPTransferRequest, bool) {
	var req SBPTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return req, false
	}
	if !domain.IsValidHistoryDirection(req.Direction) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "direction must be OUTGOING or INCOMING"})
		return req, false
	}
	if req.Amount == nil || *req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be greater than 0"})
		return req, false
	}
	if req.BalanceBefore == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "balanceBefore is required"})
		return req, false
	}
	if !domain.IsValidCashTransferBalance(*req.Amount, *req.BalanceBefore, req.Direction) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "balanceBefore cannot be less than amount for OUTGOING direction"})
		return req, false
	}
	if _, err := time.Parse(domain.HistoryTimeLayout, req.Time); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "time must match format 2026-05-02T11:08:52+0700"})
		return req, false
	}
	if strings.TrimSpace(req.OperationFirstName) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "operationFirstName is required"})
		return req, false
	}
	if strings.TrimSpace(req.OperationLastName) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "operationLastName is required"})
		return req, false
	}
	if !domain.IsValidSBPTransferBankID(req.BankID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bankId is invalid"})
		return req, false
	}

	return req, true
}

func bindUpdateSBPTransferRequest(c *gin.Context) (UpdateSBPTransferRequest, bool) {
	var req UpdateSBPTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return req, false
	}
	if req.Amount == nil &&
		req.BalanceBefore == nil &&
		req.Direction == nil &&
		req.Time == nil &&
		req.OperationFirstName == nil &&
		req.OperationMiddleName == nil &&
		req.OperationLastName == nil &&
		req.BankID == nil &&
		req.PhoneNumber == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one field must be provided"})
		return req, false
	}
	if req.Direction != nil && !domain.IsValidHistoryDirection(*req.Direction) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "direction must be OUTGOING or INCOMING"})
		return req, false
	}
	if req.Amount != nil && *req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be greater than 0"})
		return req, false
	}
	if req.Time != nil {
		if _, err := time.Parse(domain.HistoryTimeLayout, *req.Time); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "time must match format 2026-05-02T11:08:52+0700"})
			return req, false
		}
	}
	if req.OperationFirstName != nil && strings.TrimSpace(*req.OperationFirstName) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "operationFirstName cannot be empty"})
		return req, false
	}
	if req.OperationLastName != nil && strings.TrimSpace(*req.OperationLastName) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "operationLastName cannot be empty"})
		return req, false
	}
	if req.BankID != nil && !domain.IsValidSBPTransferBankID(*req.BankID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bankId is invalid"})
		return req, false
	}

	return req, true
}
