package templates

import (
	"fmt"
	"math"
	"strings"

	"project/internal/modules/sms/domain"
)

const beelineSMSAddress = "8464"

type BeelinePaymentData struct {
	TotalAmount  float64
	Commission   float64
	ReceiverCard string
}

func RenderBeelinePayment(data any) (domain.Message, error) {
	payment, ok := data.(BeelinePaymentData)
	if !ok {
		return domain.Message{}, domain.ErrInvalidMessage
	}
	if payment.TotalAmount <= 0 || payment.Commission < 0 || strings.TrimSpace(payment.ReceiverCard) == "" {
		return domain.Message{}, domain.ErrInvalidMessage
	}

	body := fmt.Sprintf(
		`Перевод с баланса на карту: к оплате %s руб., включая комиссию %s руб. Отменить покупку — отправьте «нет». Принять оферту (ofertamc.beeline.ru) и оплатить — отправьте в ответ любой символ`,
		formatSMSRubles(payment.TotalAmount),
		formatSMSRubles(payment.Commission),
	)

	return domain.Message{
		Address: beelineSMSAddress,
		Body:    body,
	}, nil
}

func formatSMSRubles(value float64) string {
	if math.Mod(value, 1) == 0 {
		return fmt.Sprintf("%.0f", value)
	}

	return fmt.Sprintf("%.2f", value)
}

func formatBeelineSMSCard(card string) string {
	card = strings.ReplaceAll(strings.TrimSpace(card), " ", "")
	if card == "" {
		return card
	}

	if idx := strings.Index(card, "**"); idx >= 0 {
		prefix := card[:idx]
		suffix := card[idx+2:]
		if len(prefix) >= 6 && len(suffix) >= 4 {
			return fmt.Sprintf("%s %s** **** %s", prefix[:4], prefix[4:6], suffix[len(suffix)-4:])
		}
	}

	return card
}
