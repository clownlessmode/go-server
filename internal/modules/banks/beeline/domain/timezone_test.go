package domain_test

import (
	"testing"
	"time"

	"project/internal/modules/banks/beeline/domain"
)

func TestFormatBeelineDateTime(t *testing.T) {
	value := time.Date(2026, 6, 1, 12, 21, 0, 0, time.UTC)
	got := domain.FormatBeelineDateTime(value)
	if got != "2026-06-01T15:21:00" {
		t.Fatalf("FormatBeelineDateTime = %q", got)
	}
}

func TestParsePaymentTimeMoscowNaive(t *testing.T) {
	got, err := domain.ParsePaymentTime("2026-06-01T15:21:00")
	if err != nil {
		t.Fatalf("ParsePaymentTime: %v", err)
	}

	want := time.Date(2026, 6, 1, 12, 21, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("parsed = %s want %s", got, want)
	}
}

func TestParsePaymentTimeRFC3339WithOffset(t *testing.T) {
	got, err := domain.ParsePaymentTime("2026-06-01T15:21:00+03:00")
	if err != nil {
		t.Fatalf("ParsePaymentTime: %v", err)
	}

	want := time.Date(2026, 6, 1, 12, 21, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("parsed = %s want %s", got, want)
	}
}
