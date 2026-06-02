package detalization

import (
	"testing"
	"time"

	"project/internal/modules/banks/beeline/domain"
)

func TestPadTodayIncomingSMSAddsMissingOperations(t *testing.T) {
	loc := domain.BeelineLocation()
	now := time.Date(2026, 6, 2, 18, 0, 0, 0, loc)
	data := map[string]any{
		"balances": []any{
			map[string]any{
				"code":       "coreBalance",
				"name":       "личный баланс",
				"startValue": 61.92,
				"endValue":   61.92,
			},
		},
		"transactions": []any{
			map[string]any{
				"category": "SUBSCRIBER_FEE_AND_SERVICE_CHANGE_FEE",
				"name":     "Абонентская плата (в сутки)",
				"dateTime": "2026-06-02T01:36:43",
				"balances": []any{
					map[string]any{
						"code":        "coreBalance",
						"changeValue": -10.08,
						"startValue":  72,
						"endValue":    61.92,
					},
				},
			},
		},
	}

	PadTodayIncomingSMS(data, "9680659702", now)

	transactions, ok := data["transactions"].([]any)
	if !ok {
		t.Fatal("transactions missing")
	}

	todayCount := 0
	syntheticCount := 0
	for _, item := range transactions {
		tx, ok := item.(map[string]any)
		if !ok {
			continue
		}
		parsed, ok := parseReportTransactionDateTime(transactionDateTime(tx))
		if !ok {
			continue
		}
		parsed = parsed.In(loc)
		if parsed.Year() != now.Year() || parsed.Month() != now.Month() || parsed.Day() != now.Day() {
			continue
		}
		todayCount++
		if isSyntheticIncomingPaddingSMS(tx) {
			syntheticCount++
		}
	}

	if todayCount != dailyOperationTarget {
		t.Fatalf("today operations = %d, want %d", todayCount, dailyOperationTarget)
	}
	if syntheticCount != dailyOperationTarget-1 {
		t.Fatalf("synthetic sms = %d, want %d", syntheticCount, dailyOperationTarget-1)
	}
}

func TestPadTodayIncomingSMSDoesNotChangeBalance(t *testing.T) {
	loc := domain.BeelineLocation()
	now := time.Date(2026, 6, 2, 18, 0, 0, 0, loc)
	data := map[string]any{
		"balances": []any{
			map[string]any{
				"code":       "coreBalance",
				"name":       "личный баланс",
				"startValue": 61.92,
				"endValue":   61.92,
			},
		},
		"transactions": []any{
			map[string]any{
				"category": "OTHER",
				"name":     "period opening",
				"dateTime": "2026-06-02T00:00:00",
				"balances": []any{
					map[string]any{
						"code":        "coreBalance",
						"changeValue": 0,
						"startValue":  61.92,
						"endValue":    61.92,
					},
				},
			},
		},
	}

	before, _ := recalculateBalances(data, nil)
	PadTodayIncomingSMS(data, "9680659702", now)
	after, _ := recalculateBalances(data, nil)

	if before != 61.92 || after != 61.92 {
		t.Fatalf("balance changed: before=%.2f after=%.2f", before, after)
	}
}
