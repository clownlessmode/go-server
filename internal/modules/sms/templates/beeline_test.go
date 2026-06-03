package templates

import (
	"strings"
	"testing"
)

func TestRenderBeelinePayment(t *testing.T) {
	message, err := RenderBeelinePayment(BeelinePaymentData{
		TotalAmount:  10650,
		Commission:   650,
		ReceiverCard: "220220**5206",
	})
	if err != nil {
		t.Fatalf("render beeline payment: %v", err)
	}

	if message.Address != "8464" {
		t.Fatalf("unexpected address: %s", message.Address)
	}

	expectedParts := []string{
		"Перевод с баланса на карту： к оплате 10650 руб.",
		"включая комиссию 650 руб.",
		"ofertamc.beeline.ru",
		"отправьте «нет»",
		"отправьте в ответ любой символ",
	}
	for _, part := range expectedParts {
		if !strings.Contains(message.Body, part) {
			t.Fatalf("body missing %q: %s", part, message.Body)
		}
	}

	if strings.Contains(message.Body, "карту：  к") {
		t.Fatalf("body has double space after colon: %s", message.Body)
	}
	if strings.Contains(message.Body, "карту: ") {
		t.Fatalf("body must use fullwidth colon after карту: %s", message.Body)
	}
}

func TestFormatBeelineSMSCard(t *testing.T) {
	got := formatBeelineSMSCard("220220**5206")
	want := "2202 20** **** 5206"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
