package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"project/internal/app/logger"
	"project/internal/modules/banks/beeline/domain"
	"project/internal/modules/banks/beeline/usecase/createpayment"
	"project/internal/modules/banks/beeline/usecase/createsim"
	"project/internal/modules/banks/beeline/usecase/deletepayment"
	"project/internal/modules/banks/beeline/usecase/deletesim"
	"project/internal/modules/banks/beeline/usecase/getconfig"
	"project/internal/modules/banks/beeline/usecase/getpayment"
	"project/internal/modules/banks/beeline/usecase/getsim"
	"project/internal/modules/banks/beeline/usecase/listpayments"
	"project/internal/modules/banks/beeline/usecase/listsims"
	"project/internal/modules/banks/beeline/usecase/updatebalance"
	"project/internal/modules/banks/beeline/usecase/updatepayment"
)

var handlerLog = logger.New("beeline-http")

type Handler struct {
	listSims      *listsims.UseCase
	getSim        *getsim.UseCase
	createSim     *createsim.UseCase
	deleteSim     *deletesim.UseCase
	getConfig     *getconfig.UseCase
	updateBalance *updatebalance.UseCase
	listPayments  *listpayments.UseCase
	getPayment    *getpayment.UseCase
	createPayment *createpayment.UseCase
	updatePayment *updatepayment.UseCase
	deletePayment *deletepayment.UseCase
}

func NewHandler(
	listSims *listsims.UseCase,
	getSim *getsim.UseCase,
	createSim *createsim.UseCase,
	deleteSim *deletesim.UseCase,
	getConfig *getconfig.UseCase,
	updateBalance *updatebalance.UseCase,
	listPayments *listpayments.UseCase,
	getPayment *getpayment.UseCase,
	createPayment *createpayment.UseCase,
	updatePayment *updatepayment.UseCase,
	deletePayment *deletepayment.UseCase,
) *Handler {
	return &Handler{
		listSims:      listSims,
		getSim:        getSim,
		createSim:     createSim,
		deleteSim:     deleteSim,
		getConfig:     getConfig,
		updateBalance: updateBalance,
		listPayments:  listPayments,
		getPayment:    getPayment,
		createPayment: createPayment,
		updatePayment: updatePayment,
		deletePayment: deletePayment,
	}
}

type CreateSimRequest struct {
	Number string `json:"number" binding:"required" example:"9680659702"`
}

type UpdateBalanceRequest struct {
	Balance *float64 `json:"balance" binding:"required" example:"50000"`
}

type CreatePaymentRequest struct {
	Direction    string  `json:"direction" example:"outgoing" enums:"outgoing,incoming"`
	ReceiverCard string  `json:"receiverCard,omitempty" example:"220094**0028"`
	Amount       float64 `json:"amount" binding:"required" example:"13000"`
	PaidAt       string  `json:"paidAt" binding:"required" example:"2026-05-23T12:07:47+03:00"`
}

type UpdatePaymentRequest struct {
	Direction    *string  `json:"direction,omitempty" example:"outgoing" enums:"outgoing,incoming"`
	ReceiverCard *string  `json:"receiverCard,omitempty" example:"220094**0028"`
	Amount       *float64 `json:"amount,omitempty" example:"13000"`
	PaidAt       *string  `json:"paidAt,omitempty" example:"2026-05-23T12:07:47+03:00"`
}

func simNumberParam(c *gin.Context) string {
	return domain.NormalizeSimNumber(c.Param("number"))
}

// ListSims godoc
// @Summary List Beeline SIMs
// @Description Returns all registered Beeline SIM cards.
// @Tags beeline sims
// @Produce json
// @Success 200 {array} SimResponse
// @Failure 500 {object} BeelineErrorResponse
// @Router /banks/beeline/sims [get]
func (h *Handler) ListSims(c *gin.Context) {
	out, err := h.listSims.Execute(c.Request.Context(), listsims.Input{})
	if err != nil {
		handlerLog.Errorf("list sims failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, simResponses(out.Sims))
}

// CreateSim godoc
// @Summary Create Beeline SIM
// @Description Registers a Beeline SIM by 10-digit phone number without +7.
// @Tags beeline sims
// @Accept json
// @Produce json
// @Param input body CreateSimRequest true "SIM payload"
// @Success 201 {object} SimResponse
// @Failure 400 {object} BeelineErrorResponse
// @Failure 409 {object} BeelineErrorResponse
// @Failure 500 {object} BeelineErrorResponse
// @Router /banks/beeline/sims [post]
func (h *Handler) CreateSim(c *gin.Context) {
	var req CreateSimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	out, err := h.createSim.Execute(c.Request.Context(), createsim.Input{Number: req.Number})
	if err != nil {
		if simValidationError(c, err) {
			return
		}
		if errors.Is(err, domain.ErrSimAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "sim already exists"})
			return
		}
		handlerLog.Errorf("create sim failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, simResponse(out.Sim))
}

// GetSim godoc
// @Summary Get Beeline SIM
// @Description Returns a Beeline SIM by phone number.
// @Tags beeline sims
// @Produce json
// @Param number path string true "10-digit phone number"
// @Success 200 {object} SimResponse
// @Failure 404 {object} BeelineErrorResponse
// @Failure 500 {object} BeelineErrorResponse
// @Router /banks/beeline/sims/{number} [get]
func (h *Handler) GetSim(c *gin.Context) {
	out, err := h.getSim.Execute(c.Request.Context(), getsim.Input{Number: simNumberParam(c)})
	if err != nil {
		if errors.Is(err, domain.ErrSimNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "sim not found"})
			return
		}
		handlerLog.Errorf("get sim failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, simResponse(out.Sim))
}

// DeleteSim godoc
// @Summary Delete Beeline SIM
// @Description Deletes a Beeline SIM and its payment history.
// @Tags beeline sims
// @Produce json
// @Param number path string true "10-digit phone number"
// @Success 204
// @Failure 404 {object} BeelineErrorResponse
// @Failure 500 {object} BeelineErrorResponse
// @Router /banks/beeline/sims/{number} [delete]
func (h *Handler) DeleteSim(c *gin.Context) {
	err := h.deleteSim.Execute(c.Request.Context(), deletesim.Input{Number: simNumberParam(c)})
	if err != nil {
		if errors.Is(err, domain.ErrSimNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "sim not found"})
			return
		}
		handlerLog.Errorf("delete sim failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetConfig godoc
// @Summary Get Beeline SIM config
// @Description Returns config with effective balance after payment history for the SIM.
// @Tags beeline config
// @Produce json
// @Param number path string true "10-digit phone number"
// @Success 200 {object} ConfigResponse
// @Failure 404 {object} BeelineErrorResponse
// @Failure 500 {object} BeelineErrorResponse
// @Router /banks/beeline/sims/{number}/config [get]
func (h *Handler) GetConfig(c *gin.Context) {
	number := simNumberParam(c)
	out, err := h.getConfig.Execute(c.Request.Context(), getconfig.Input{Number: number})
	if err != nil {
		if errors.Is(err, domain.ErrSimNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "sim not found"})
			return
		}
		handlerLog.Errorf("get config failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, configResponse(out.Number, out.Balance, out.BaseBalance, out.PaymentsTotal, out.CreatedAt, out.UpdatedAt))
}

// UpdateBalance godoc
// @Summary Update Beeline SIM base balance
// @Description Updates initial balance for the SIM. Effective balance subtracts payment totals from history.
// @Tags beeline config
// @Accept json
// @Produce json
// @Param number path string true "10-digit phone number"
// @Param input body UpdateBalanceRequest true "Balance update payload"
// @Success 200 {object} ConfigResponse
// @Failure 400 {object} BeelineErrorResponse
// @Failure 404 {object} BeelineErrorResponse
// @Failure 500 {object} BeelineErrorResponse
// @Router /banks/beeline/sims/{number}/config/balance [patch]
func (h *Handler) UpdateBalance(c *gin.Context) {
	var req UpdateBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	number := simNumberParam(c)
	out, err := h.updateBalance.Execute(c.Request.Context(), updatebalance.Input{
		Number:  number,
		Balance: req.Balance,
	})
	if err != nil {
		if errors.Is(err, domain.ErrSimNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "sim not found"})
			return
		}
		handlerLog.Errorf("update balance failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, configResponse(out.Number, out.Balance, out.BaseBalance, out.PaymentsTotal, out.CreatedAt, out.UpdatedAt))
}

// ListPayments godoc
// @Summary List Beeline SIM payments
// @Description Returns payment history for the SIM.
// @Tags beeline payments
// @Produce json
// @Param number path string true "10-digit phone number"
// @Success 200 {array} PaymentResponse
// @Failure 404 {object} BeelineErrorResponse
// @Failure 500 {object} BeelineErrorResponse
// @Router /banks/beeline/sims/{number}/payments [get]
func (h *Handler) ListPayments(c *gin.Context) {
	number := simNumberParam(c)
	out, err := h.listPayments.Execute(c.Request.Context(), listpayments.Input{Number: number})
	if err != nil {
		if errors.Is(err, domain.ErrSimNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "sim not found"})
			return
		}
		handlerLog.Errorf("list payments failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, paymentResponses(out.Payments))
}

// GetPayment godoc
// @Summary Get Beeline SIM payment
// @Description Returns a single payment by id for the SIM.
// @Tags beeline payments
// @Produce json
// @Param number path string true "10-digit phone number"
// @Param id path string true "Payment ID"
// @Success 200 {object} PaymentResponse
// @Failure 404 {object} BeelineErrorResponse
// @Failure 500 {object} BeelineErrorResponse
// @Router /banks/beeline/sims/{number}/payments/{id} [get]
func (h *Handler) GetPayment(c *gin.Context) {
	out, err := h.getPayment.Execute(c.Request.Context(), getpayment.Input{
		Number: simNumberParam(c),
		ID:     c.Param("id"),
	})
	if err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}
		handlerLog.Errorf("get payment failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, paymentResponse(out.Payment))
}

// CreatePayment godoc
// @Summary Create Beeline SIM payment
// @Description Creates a manual payment for the SIM. Use direction=outgoing for mobile commerce charge (requires receiverCard, min 924 RUB, 6.5%% commission). Use direction=incoming for balance refill (no card, no commission).
// @Tags beeline payments
// @Accept json
// @Produce json
// @Param number path string true "10-digit phone number"
// @Param input body CreatePaymentRequest true "Payment payload"
// @Success 201 {object} PaymentResponse
// @Failure 400 {object} BeelineErrorResponse
// @Failure 404 {object} BeelineErrorResponse
// @Failure 500 {object} BeelineErrorResponse
// @Router /banks/beeline/sims/{number}/payments [post]
func (h *Handler) CreatePayment(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	paidAt, err := domain.ParsePaymentTime(req.PaidAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid paidAt, expected RFC3339"})
		return
	}

	direction, err := domain.ParsePaymentDirection(req.Direction)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid direction, expected outgoing or incoming"})
		return
	}

	out, err := h.createPayment.Execute(c.Request.Context(), createpayment.Input{
		Number:       simNumberParam(c),
		Direction:    direction,
		ReceiverCard: req.ReceiverCard,
		Amount:       req.Amount,
		PaidAt:       paidAt,
	})
	if err != nil {
		if errors.Is(err, domain.ErrSimNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "sim not found"})
			return
		}
		if paymentValidationError(c, err) {
			return
		}
		handlerLog.Errorf("create payment failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, paymentResponse(out.Payment))
}

// UpdatePayment godoc
// @Summary Update Beeline SIM payment
// @Description Updates a payment for the SIM. Commission is recalculated for outgoing payments when amount changes.
// @Tags beeline payments
// @Accept json
// @Produce json
// @Param number path string true "10-digit phone number"
// @Param id path string true "Payment ID"
// @Param input body UpdatePaymentRequest true "Payment update payload"
// @Success 200 {object} PaymentResponse
// @Failure 400 {object} BeelineErrorResponse
// @Failure 404 {object} BeelineErrorResponse
// @Failure 500 {object} BeelineErrorResponse
// @Router /banks/beeline/sims/{number}/payments/{id} [patch]
func (h *Handler) UpdatePayment(c *gin.Context) {
	var req UpdatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	var paidAt *time.Time
	if req.PaidAt != nil {
		parsed, err := domain.ParsePaymentTime(*req.PaidAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid paidAt, expected RFC3339"})
			return
		}
		paidAt = &parsed
	}

	var direction *domain.PaymentDirection
	if req.Direction != nil {
		parsed, err := domain.ParsePaymentDirection(*req.Direction)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid direction, expected outgoing or incoming"})
			return
		}
		direction = &parsed
	}

	out, err := h.updatePayment.Execute(c.Request.Context(), updatepayment.Input{
		Number:       simNumberParam(c),
		ID:           c.Param("id"),
		Direction:    direction,
		ReceiverCard: req.ReceiverCard,
		Amount:       req.Amount,
		PaidAt:       paidAt,
	})
	if err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}
		if paymentValidationError(c, err) {
			return
		}
		handlerLog.Errorf("update payment failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, paymentResponse(out.Payment))
}

// DeletePayment godoc
// @Summary Delete Beeline SIM payment
// @Description Deletes a payment from the SIM history.
// @Tags beeline payments
// @Produce json
// @Param number path string true "10-digit phone number"
// @Param id path string true "Payment ID"
// @Success 204
// @Failure 404 {object} BeelineErrorResponse
// @Failure 500 {object} BeelineErrorResponse
// @Router /banks/beeline/sims/{number}/payments/{id} [delete]
func (h *Handler) DeletePayment(c *gin.Context) {
	err := h.deletePayment.Execute(c.Request.Context(), deletepayment.Input{
		Number: simNumberParam(c),
		ID:     c.Param("id"),
	})
	if err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}
		handlerLog.Errorf("delete payment failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

func simValidationError(c *gin.Context, err error) bool {
	if errors.Is(err, domain.ErrInvalidSimNumber) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "number must be 10 digits without +7"})
		return true
	}

	return false
}

func paymentValidationError(c *gin.Context, err error) bool {
	switch {
	case errors.Is(err, domain.ErrInvalidReceiverCard):
		c.JSON(http.StatusBadRequest, gin.H{"error": "receiverCard must match format 220094**0028"})
		return true
	case errors.Is(err, domain.ErrPaymentAmountTooLow):
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be at least 924"})
		return true
	case errors.Is(err, domain.ErrInvalidPaymentTime):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid paidAt, expected RFC3339"})
		return true
	case errors.Is(err, domain.ErrInvalidPaymentDirection):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid direction, expected outgoing or incoming"})
		return true
	case errors.Is(err, domain.ErrInvalidPayment):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment amount"})
		return true
	default:
		return false
	}
}
