package detalization

import (
	"strings"
	"testing"
	"time"

	"project/internal/modules/banks/beeline/domain"
)

func TestPadTodayOutgoingSMSAddsMissingOperations(t *testing.T) {
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

	PadTodayOutgoingSMS(data, "9680659702", now)

	transactions, ok := data["transactions"].([]any)
	if !ok {
		t.Fatal("transactions missing")
	}

	todayCount := 0
	syntheticCount := 0
	todayPrefix := now.Format("2006-01-02")
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
		if isSyntheticPaddingSMS(tx) {
			syntheticCount++
			if !strings.HasPrefix(transactionDateTime(tx), todayPrefix) {
				t.Fatalf("synthetic sms date %q is not today %q", transactionDateTime(tx), todayPrefix)
			}
			if tx["name"] != "исходящее sms" {
				t.Fatalf("synthetic sms name = %q, want исходящее sms", tx["name"])
			}
			if tx["number"] == domain.PaymentFlowSMSNumber {
				t.Fatalf("synthetic sms must not use payment flow number %q", domain.PaymentFlowSMSNumber)
			}
			if tx["typeCall"] != "outgoingCall" {
				t.Fatalf("synthetic sms typeCall = %q, want outgoingCall", tx["typeCall"])
			}
		}
	}

	target := dailyOperationTargetFor("9680659702")
	if todayCount != target {
		t.Fatalf("today operations = %d, want %d", todayCount, target)
	}
	if syntheticCount != target-1 {
		t.Fatalf("synthetic sms = %d, want %d", syntheticCount, target-1)
	}
}

func TestSyntheticOutgoingSMSTimesStayOnSameDay(t *testing.T) {
	loc := domain.BeelineLocation()
	dayStart := time.Date(2026, 6, 2, 0, 0, 0, 0, loc)
	dayEnd := dayStart.Add(24*time.Hour - time.Nanosecond)
	anchor := time.Date(2026, 6, 2, 1, 36, 43, 0, loc)

	times := syntheticOutgoingSMSTimes("9680659702", dayStart, dayEnd, anchor, 12)
	if len(times) != 12 {
		t.Fatalf("times len = %d, want 12", len(times))
	}

	for index, paidAt := range times {
		paidAt = paidAt.In(loc)
		if paidAt.Year() != 2026 || paidAt.Month() != time.June || paidAt.Day() != 2 {
			t.Fatalf("time[%d] = %s, want date 2026-06-02", index, paidAt.Format(time.RFC3339))
		}
	}
}

func TestDailyOperationTargetIsSeededBySimNumber(t *testing.T) {
	first := dailyOperationTargetFor("9680659702")
	second := dailyOperationTargetFor("9680659702")
	if first != second {
		t.Fatalf("same sim target changed: %d vs %d", first, second)
	}
	if first < dailyOperationTargetMin || first > dailyOperationTargetMax {
		t.Fatalf("target = %d, want between %d and %d", first, dailyOperationTargetMin, dailyOperationTargetMax)
	}

	otherSim := dailyOperationTargetFor("9063747835")
	if first == otherSim {
		t.Fatalf("expected different sim numbers to vary target, got %d for both", first)
	}
}

func TestPadTodayOutgoingSMSDoesNotChangeBalance(t *testing.T) {
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
	PadTodayOutgoingSMS(data, "9680659702", now)
	after, _ := recalculateBalances(data, nil)

	if before != 61.92 || after != 61.92 {
		t.Fatalf("balance changed: before=%.2f after=%.2f", before, after)
	}
}
