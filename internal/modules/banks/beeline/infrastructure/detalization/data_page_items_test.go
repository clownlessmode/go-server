package detalization

import (
	"testing"
	"time"
)

func TestBuildDataPageItemsWithDateBadge(t *testing.T) {
	loc := time.FixedZone("MSK", 3*60*60)
	data := map[string]any{
		"transactions": []any{
			map[string]any{
				"dateTime": "2026-03-20T01:36:00",
				"name":     "пополнение баланса",
				"balances": []any{map[string]any{"code": "coreBalance", "changeValue": 300.0}},
			},
			map[string]any{
				"dateTime": "2026-03-19T19:35:00",
				"name":     "компенсация затрат на пополнение баланса",
				"balances": []any{map[string]any{"code": "coreBalance", "changeValue": -4.56}},
			},
		},
	}

	items := buildDataPageItems(data, 0, 2)
	if len(items) != 3 {
		t.Fatalf("items = %d, want 3", len(items))
	}
	if items[0].Kind != dataPageItemOperation {
		t.Fatalf("first item should be operation")
	}
	if items[1].Kind != dataPageItemBadge {
		t.Fatalf("second item should be badge")
	}
	if !items[1].DateTime.Equal(time.Date(2026, 3, 19, 19, 35, 0, 0, loc)) {
		t.Fatalf("badge date = %v", items[1].DateTime)
	}
}

func TestMaxDataPageTransactionCountWithBadges(t *testing.T) {
	transactions := make([]any, 0, 18)
	for day := 20; day >= 3; day-- {
		transactions = append(transactions, map[string]any{
			"dateTime": time.Date(2026, 3, day, 12, 0, 0, 0, time.UTC).Format(time.RFC3339),
			"name":     "операция",
			"balances": []any{map[string]any{"code": "coreBalance", "changeValue": -1.0}},
		})
	}

	data := map[string]any{"transactions": transactions}
	count := maxDataPageTransactionCount(data, 0)
	if count != 9 {
		t.Fatalf("count = %d, want 9", count)
	}
}

func TestMaxDataPageTransactionCountWithoutBadges(t *testing.T) {
	transactions := make([]any, 0, 17)
	for index := 0; index < 17; index++ {
		transactions = append(transactions, map[string]any{
			"dateTime": "2026-03-20T01:36:00",
			"name":     "операция",
			"balances": []any{map[string]any{"code": "coreBalance", "changeValue": -1.0}},
		})
	}

	data := map[string]any{"transactions": transactions}
	count := maxDataPageTransactionCount(data, 0)
	if count != 17 {
		t.Fatalf("count = %d, want 17", count)
	}
}

func TestMaxDataPageTransactionCountWithBannerReserve(t *testing.T) {
	transactions := make([]any, 0, 17)
	for index := 0; index < 17; index++ {
		transactions = append(transactions, map[string]any{
			"dateTime": "2026-03-20T01:36:00",
			"name":     "операция",
			"balances": []any{map[string]any{"code": "coreBalance", "changeValue": -1.0}},
		})
	}

	data := map[string]any{"transactions": transactions}
	count := maxDataPageTransactionCountWithSlots(data, 0, dataPageCapacitySlots-dataPageBannerSlotCost)
	if count != 13 {
		t.Fatalf("count = %d, want 13", count)
	}
}
