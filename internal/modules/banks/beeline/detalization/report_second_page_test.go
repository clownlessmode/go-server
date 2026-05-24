package detalization_test

import (
	"regexp"
	"testing"
	"time"

	"project/internal/modules/banks/beeline/detalization"
)

func TestSecondPageTransactionsNewestFirst(t *testing.T) {
	data := map[string]any{
		"transactions": []any{
			map[string]any{
				"dateTime": "2026-03-16T02:28:00",
				"name":     "старая операция",
				"balances": []any{map[string]any{"code": "coreBalance", "changeValue": -10.0}},
			},
			map[string]any{
				"dateTime": "2026-03-16T04:55:00",
				"name":     "новая операция",
				"balances": []any{map[string]any{"code": "coreBalance", "changeValue": -400.93}},
			},
		},
	}

	rows := detalization.SecondPageTransactions(data, 13)
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(rows))
	}
	if rows[0].Title != "новая операция" {
		t.Fatalf("first row = %q", rows[0].Title)
	}
	if rows[0].Amount != "-400,93 ₽" {
		t.Fatalf("first amount = %q", rows[0].Amount)
	}
}

func TestFormatReportSecondPageDates(t *testing.T) {
	loc := time.FixedZone("MSK", 3*60*60)
	value := time.Date(2026, 3, 16, 4, 55, 0, 0, loc)

	if got := detalization.FormatReportSecondPageSectionDate(value); got != "16 марта 2026 г." {
		t.Fatalf("section date = %q", got)
	}
	if got := detalization.FormatReportTransactionDateTime(value); got != "16 мар. 2026 04:55" {
		t.Fatalf("row date = %q", got)
	}
}

func TestFormatReportTransactionDescription(t *testing.T) {
	tests := []struct {
		name string
		tx   map[string]any
		want string
	}{
		{
			name: "refill",
			tx:   map[string]any{"name": "пополнение баланса"},
			want: "основной баланс",
		},
		{
			name: "compensation",
			tx:   map[string]any{"name": "компенсация затрат на пополнение баланса"},
			want: "основной баланс",
		},
		{
			name: "beeline pay",
			tx:   map[string]any{"name": "плати с билайн: перевод на баланс билайн"},
			want: "основной баланс",
		},
		{
			name: "sms",
			tx:   map[string]any{"name": "sms free8464"},
			want: "1 шт (основной баланс)",
		},
		{
			name: "mobile commerce",
			tx:   map[string]any{"name": "списание за мобильную коммерцию"},
			want: "основной баланс",
		},
		{
			name: "connection fee",
			tx:   map[string]any{"name": "плата за подключение"},
			want: "основной баланс",
		},
		{
			name: "minutes package",
			tx:   map[string]any{"name": "начисление пакета минут"},
			want: "пакет минут",
		},
		{
			name: "traffic package",
			tx:   map[string]any{"name": "начисление пакета трафика"},
			want: "пакет интернета",
		},
		{
			name: "unlimited internet",
			tx:   map[string]any{"name": "безлимитный интернет"},
			want: "",
		},
		{
			name: "outgoing call",
			tx:   map[string]any{"name": "исходящий звонок на Билайн (Кемеровская обл.)"},
			want: "",
		},
		{
			name: "incoming call",
			tx:   map[string]any{"name": "входящий звонок с МТС (Москва - МО)"},
			want: "",
		},
		{
			name: "tariff fee",
			tx:   map[string]any{"name": "абонентская плата за тариф"},
			want: "основной баланс",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detalization.FormatReportTransactionDescription(tt.tx)
			if tt.name == "unlimited internet" {
				pattern := regexp.MustCompile(`^\d{1,2},\d (мб|кб) \(основной баланс\)$`)
				if !pattern.MatchString(got) {
					t.Fatalf("description = %q, want format like 5,3 мб (основной баланс)", got)
				}
				return
			}
			if tt.name == "outgoing call" || tt.name == "incoming call" {
				pattern := regexp.MustCompile(`^00:(0[1-9]|1[0-5]):[0-5]\d \(основной баланс\)$`)
				if !pattern.MatchString(got) {
					t.Fatalf("description = %q, want format like 00:05:23 (основной баланс)", got)
				}
				return
			}
			if got != tt.want {
				t.Fatalf("description = %q, want %q", got, tt.want)
			}
		})
	}
}
