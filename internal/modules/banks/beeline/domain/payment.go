package domain

import (
	"crypto/rand"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)

const (
	CommissionRate     = 0.065
	MinPaymentAmount   = 924
	PaymentTimeLayout  = time.RFC3339
)

type PaymentSource string

const (
	PaymentSourceManual         PaymentSource = "manual"
	PaymentSourcePaymentFlow    PaymentSource = "payment_flow"
	PaymentSourcePaymentFlowSMS PaymentSource = "payment_flow_sms"
)

const PaymentFlowSMSNumber = "free8464"

type PaymentDirection string

const (
	PaymentDirectionOutgoing PaymentDirection = "outgoing"
	PaymentDirectionIncoming PaymentDirection = "incoming"
)

var receiverCardPattern = regexp.MustCompile(`^\d{6}\*\*\d{4}$`)

type Payment struct {
	ID           string
	SimNumber    string
	Direction    PaymentDirection
	ReceiverCard string
	Amount       float64
	Commission   float64
	Total        float64
	Source       PaymentSource
	PaidAt       time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func CalculateCommission(amount float64) float64 {
	return RoundMoney(amount * CommissionRate)
}

func RoundMoney(value float64) float64 {
	return math.Round(value*100) / 100
}

func ValidateReceiverCard(card string) error {
	card = strings.TrimSpace(card)
	if !receiverCardPattern.MatchString(card) {
		return ErrInvalidReceiverCard
	}

	return nil
}

func ValidatePaymentAmount(amount float64) error {
	if amount < MinPaymentAmount {
		return ErrPaymentAmountTooLow
	}

	return nil
}

func ParsePaymentDirection(value string) (PaymentDirection, error) {
	switch PaymentDirection(strings.ToLower(strings.TrimSpace(value))) {
	case "", PaymentDirectionOutgoing:
		return PaymentDirectionOutgoing, nil
	case PaymentDirectionIncoming:
		return PaymentDirectionIncoming, nil
	default:
		return "", ErrInvalidPaymentDirection
	}
}

func EffectiveBalance(base *float64, outgoingTotal, incomingTotal float64) *float64 {
	if base == nil {
		return nil
	}

	effective := RoundMoney(*base + incomingTotal - outgoingTotal)
	return &effective
}

func NewManualPayment(direction PaymentDirection, receiverCard string, amount float64, paidAt time.Time) (Payment, error) {
	if direction == PaymentDirectionIncoming {
		return newIncomingManualPayment(amount, paidAt)
	}

	return newOutgoingManualPayment(receiverCard, amount, paidAt)
}

func newIncomingManualPayment(amount float64, paidAt time.Time) (Payment, error) {
	if amount <= 0 {
		return Payment{}, ErrInvalidPayment
	}

	amount = RoundMoney(amount)

	return Payment{
		ID:        newPaymentID(),
		Direction: PaymentDirectionIncoming,
		Amount:    amount,
		Commission: 0,
		Total:     amount,
		Source:    PaymentSourceManual,
		PaidAt:    paidAt.UTC(),
	}, nil
}

func newOutgoingManualPayment(receiverCard string, amount float64, paidAt time.Time) (Payment, error) {
	receiverCard = strings.TrimSpace(receiverCard)
	if err := ValidateReceiverCard(receiverCard); err != nil {
		return Payment{}, err
	}
	if err := ValidatePaymentAmount(amount); err != nil {
		return Payment{}, err
	}

	amount = RoundMoney(amount)
	commission := CalculateCommission(amount)

	return Payment{
		ID:           newPaymentID(),
		Direction:    PaymentDirectionOutgoing,
		ReceiverCard: receiverCard,
		Amount:       amount,
		Commission:   commission,
		Total:        RoundMoney(amount + commission),
		Source:       PaymentSourceManual,
		PaidAt:       paidAt.UTC(),
	}, nil
}

func NewPaymentFlowSMSPayment(paidAt time.Time) Payment {
	return Payment{
		ID:           newPaymentID(),
		Direction:    PaymentDirectionOutgoing,
		ReceiverCard: PaymentFlowSMSNumber,
		Source:       PaymentSourcePaymentFlowSMS,
		PaidAt:       paidAt.UTC(),
	}
}

func NewPaymentFlowPayment(receiverCard string, amount, commission float64, paidAt time.Time) (Payment, error) {
	receiverCard = strings.TrimSpace(receiverCard)
	if err := ValidateReceiverCard(receiverCard); err != nil {
		return Payment{}, err
	}
	if err := ValidatePaymentAmount(amount); err != nil {
		return Payment{}, err
	}
	if commission < 0 {
		return Payment{}, ErrInvalidPayment
	}

	amount = RoundMoney(amount)
	commission = RoundMoney(commission)

	return Payment{
		ID:           newPaymentID(),
		Direction:    PaymentDirectionOutgoing,
		ReceiverCard: receiverCard,
		Amount:       amount,
		Commission:   commission,
		Total:        RoundMoney(amount + commission),
		Source:       PaymentSourcePaymentFlow,
		PaidAt:       paidAt.UTC(),
	}, nil
}

func (p Payment) ApplyUpdate(direction *PaymentDirection, receiverCard *string, amount *float64, paidAt *time.Time) (Payment, error) {
	updated := p

	if direction != nil {
		if *direction != PaymentDirectionOutgoing && *direction != PaymentDirectionIncoming {
			return Payment{}, ErrInvalidPaymentDirection
		}
		updated.Direction = *direction
	}

	if receiverCard != nil {
		updated.ReceiverCard = strings.TrimSpace(*receiverCard)
	}
	if amount != nil {
		updated.Amount = RoundMoney(*amount)
	}
	if paidAt != nil {
		updated.PaidAt = paidAt.UTC()
	}

	switch updated.Direction {
	case PaymentDirectionIncoming:
		if updated.Amount <= 0 {
			return Payment{}, ErrInvalidPayment
		}
		updated.Commission = 0
		updated.Total = updated.Amount
		updated.ReceiverCard = ""
	default:
		if err := ValidateReceiverCard(updated.ReceiverCard); err != nil {
			return Payment{}, err
		}
		if err := ValidatePaymentAmount(updated.Amount); err != nil {
			return Payment{}, err
		}
		updated.Commission = CalculateCommission(updated.Amount)
		updated.Total = RoundMoney(updated.Amount + updated.Commission)
	}

	return updated, nil
}

func ParsePaymentTime(value string) (time.Time, error) {
	parsed, err := time.Parse(PaymentTimeLayout, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, ErrInvalidPaymentTime
	}

	return parsed, nil
}

func newPaymentID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])

	return fmt.Sprintf("%x", b[:])
}
