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
		"10650.00 руб.",
		"2202 20** **** 5206",
		"650.00 руб.",
		"ofertamc.beeline.ru",
	}
	for _, part := range expectedParts {
		if !strings.Contains(message.Body, part) {
			t.Fatalf("body missing %q: %s", part, message.Body)
		}
	}
}

func TestFormatBeelineSMSCard(t *testing.T) {
	got := formatBeelineSMSCard("220220**5206")
	want := "2202 20** **** 5206"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
