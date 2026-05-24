package http

import (
	"time"

	"project/internal/modules/banks/beeline/domain"
)

type SimResponse struct {
	Number    string    `json:"number"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ConfigResponse struct {
	Number        string    `json:"number"`
	Balance       *float64  `json:"balance"`
	PaymentsTotal float64   `json:"paymentsTotal"`
	IncomingTotal float64   `json:"incomingTotal"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type PaymentResponse struct {
	ID           string    `json:"id"`
	SimNumber    string    `json:"simNumber"`
	Direction    string    `json:"direction"`
	ReceiverCard string    `json:"receiverCard,omitempty"`
	Amount       float64   `json:"amount"`
	Commission   float64   `json:"commission"`
	Total        float64   `json:"total"`
	Source       string    `json:"source"`
	PaidAt       time.Time `json:"paidAt"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type DetalizationResponse struct {
	Number      string         `json:"number"`
	PeriodStart time.Time      `json:"periodStart"`
	PeriodEnd   time.Time      `json:"periodEnd"`
	Balance     *float64       `json:"balance"`
	Data        map[string]any `json:"data"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

type BeelineErrorResponse struct {
	Error string `json:"error"`
}

func simResponse(sim domain.Sim) SimResponse {
	return SimResponse{
		Number:    sim.Number,
		CreatedAt: sim.CreatedAt,
		UpdatedAt: sim.UpdatedAt,
	}
}

func simResponses(sims []domain.Sim) []SimResponse {
	result := make([]SimResponse, 0, len(sims))
	for _, sim := range sims {
		result = append(result, simResponse(sim))
	}

	return result
}

func configResponse(number string, balance *float64, paymentsTotal, incomingTotal float64, createdAt, updatedAt time.Time) ConfigResponse {
	return ConfigResponse{
		Number:        number,
		Balance:       balance,
		PaymentsTotal: paymentsTotal,
		IncomingTotal: incomingTotal,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}
}

func detalizationResponse(number string, periodStart, periodEnd time.Time, balance *float64, data map[string]any, createdAt, updatedAt time.Time) DetalizationResponse {
	return DetalizationResponse{
		Number:      number,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Balance:     balance,
		Data:        data,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

func paymentResponse(payment domain.Payment) PaymentResponse {
	return PaymentResponse{
		ID:           payment.ID,
		SimNumber:    payment.SimNumber,
		Direction:    string(payment.Direction),
		ReceiverCard: payment.ReceiverCard,
		Amount:       payment.Amount,
		Commission:   payment.Commission,
		Total:        payment.Total,
		Source:       string(payment.Source),
		PaidAt:       payment.PaidAt,
		CreatedAt:    payment.CreatedAt,
		UpdatedAt:    payment.UpdatedAt,
	}
}

func paymentResponses(payments []domain.Payment) []PaymentResponse {
	result := make([]PaymentResponse, 0, len(payments))
	for _, payment := range payments {
		result = append(result, paymentResponse(payment))
	}

	return result
}
