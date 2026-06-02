package detalization

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"math/rand/v2"
	"sort"
	"time"

	"project/internal/modules/banks/beeline/domain"
)

const (
	dailyOperationTargetMin = 13
	dailyOperationTargetMax = 20
	smsPadMinGapMinutes     = 10
	smsPadMaxGapMinutes     = 50
	syntheticSMSIDPrefix    = "pad-sms-"
)

func dailyOperationTargetFor(simNumber string) int {
	rng := rand.New(rand.NewPCG(simNumberSeed(simNumber), 0))
	span := dailyOperationTargetMax - dailyOperationTargetMin + 1
	return dailyOperationTargetMin + rng.IntN(span)
}

func simNumberSeed(simNumber string) uint64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(simNumber))
	return hasher.Sum64()
}

func PadTodayIncomingSMS(data map[string]any, simNumber string, now time.Time) {
	transactions, ok := data["transactions"].([]any)
	if !ok {
		transactions = []any{}
	}

	loc := domain.BeelineLocation()
	today := now.In(loc)
	dayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, loc)
	dayEnd := dayStart.Add(24*time.Hour - time.Nanosecond)

	todayTransactions := make([]map[string]any, 0)
	otherTransactions := make([]any, 0, len(transactions))

	for _, item := range transactions {
		tx, ok := item.(map[string]any)
		if !ok {
			otherTransactions = append(otherTransactions, item)
			continue
		}
		if isSyntheticIncomingPaddingSMS(tx) {
			continue
		}

		txTime, ok := parseReportTransactionDateTime(transactionDateTime(tx))
		if !ok {
			otherTransactions = append(otherTransactions, item)
			continue
		}
		txTime = txTime.In(loc)
		if txTime.Before(dayStart) || txTime.After(dayEnd) {
			otherTransactions = append(otherTransactions, item)
			continue
		}

		todayTransactions = append(todayTransactions, tx)
	}

	target := dailyOperationTargetFor(simNumber)
	need := target - len(todayTransactions)
	if need <= 0 {
		merged := append(otherTransactions, transactionsFromMaps(todayTransactions)...)
		data["transactions"] = merged
		sortTransactionsDesc(data)
		return
	}

	anchor := dayStart.Add(12 * time.Hour)
	if len(todayTransactions) > 0 {
		anchor = earliestTransactionTime(todayTransactions, loc)
	} else if today.After(dayStart) {
		anchor = today
	}

	padded := append([]map[string]any(nil), todayTransactions...)
	for index, paidAt := range syntheticIncomingSMSTimes(simNumber, dayStart, anchor, need) {
		padded = append(padded, syntheticIncomingSMSTransaction(simNumber, dayStart, index, paidAt))
	}

	sort.SliceStable(padded, func(i, j int) bool {
		left, _ := parseReportTransactionDateTime(transactionDateTime(padded[i]))
		right, _ := parseReportTransactionDateTime(transactionDateTime(padded[j]))
		return left.After(right)
	})

	merged := append(otherTransactions, transactionsFromMaps(padded)...)
	data["transactions"] = merged
	sortTransactionsDesc(data)
	_, _ = recalculateBalances(data, nil)
}

func syntheticIncomingSMSTimes(simNumber string, dayStart, anchor time.Time, count int) []time.Time {
	rng := rand.New(rand.NewPCG(smsPadSeed(simNumber, dayStart), 0))
	times := make([]time.Time, count)
	current := anchor

	for index := count - 1; index >= 0; index-- {
		gapMinutes := smsPadMinGapMinutes
		if smsPadMaxGapMinutes > smsPadMinGapMinutes {
			gapMinutes += rng.IntN(smsPadMaxGapMinutes - smsPadMinGapMinutes + 1)
		}
		current = current.Add(-time.Duration(gapMinutes) * time.Minute)
		if current.Before(dayStart) {
			current = dayStart.Add(time.Duration(index+1) * time.Minute)
		}
		times[index] = current
	}

	return times
}

func smsPadSeed(simNumber string, dayStart time.Time) uint64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(simNumber))
	_, _ = hasher.Write([]byte(dayStart.Format("2006-01-02")))
	return hasher.Sum64()
}

func syntheticIncomingSMSTransaction(simNumber string, dayStart time.Time, index int, paidAt time.Time) map[string]any {
	id := syntheticIncomingSMSID(simNumber, dayStart, index)
	dateTime := domain.FormatBeelineDateTime(paidAt)
	return incomingSMSTransaction(id, dateTime, domain.DefaultIncomingSMSNumber)
}

func syntheticIncomingSMSID(simNumber string, dayStart time.Time, index int) string {
	seed := fmt.Sprintf("%s|%s|%d", simNumber, dayStart.Format("2006-01-02"), index)
	sum := sha256.Sum256([]byte(seed))
	return syntheticSMSIDPrefix + hex.EncodeToString(sum[:8])
}

func isSyntheticIncomingPaddingSMS(tx map[string]any) bool {
	id, _ := tx["id"].(string)
	return len(id) >= len(syntheticSMSIDPrefix) && id[:len(syntheticSMSIDPrefix)] == syntheticSMSIDPrefix
}

func earliestTransactionTime(transactions []map[string]any, loc *time.Location) time.Time {
	earliest := time.Time{}
	for _, tx := range transactions {
		parsed, ok := parseReportTransactionDateTime(transactionDateTime(tx))
		if !ok {
			continue
		}
		parsed = parsed.In(loc)
		if earliest.IsZero() || parsed.Before(earliest) {
			earliest = parsed
		}
	}
	return earliest
}

func transactionsFromMaps(items []map[string]any) []any {
	result := make([]any, len(items))
	for index, item := range items {
		result[index] = item
	}
	return result
}

func sortTransactionsDesc(data map[string]any) {
	transactions, ok := data["transactions"].([]any)
	if !ok || len(transactions) < 2 {
		return
	}

	sort.SliceStable(transactions, func(i, j int) bool {
		return transactionDateTime(transactions[i]) > transactionDateTime(transactions[j])
	})
	data["transactions"] = transactions
}

func incomingSMSTransaction(id, dateTime, number string) map[string]any {
	return map[string]any{
		"id": id,
		"balances": []any{
			map[string]any{
				"changeValue": 0,
				"code":        "coreBalance",
				"name":        "личный баланс",
				"unit":        "RUB",
			},
		},
		"category":        "SMS_MMS",
		"categoryName":    "сообщения",
		"dateTime":        dateTime,
		"formattedNumber": number,
		"icon":            "smsMms",
		"name":            "входящее SMS",
		"number":          number,
		"roaming":         false,
		"typeCall":        "incomingCall",
		"unit":            "PIECE",
		"volume":          1,
	}
}
