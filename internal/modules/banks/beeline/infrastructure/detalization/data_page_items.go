package detalization

import (
	"time"

	detaildomain "project/internal/modules/banks/beeline/detalization"
)

const (
	dataPageSlotUnit          = 10
	dataPageCapacitySlots     = 17 * dataPageSlotUnit
	dataPageOperationSlotCost = 1 * dataPageSlotUnit
	dataPageBadgeSlotCost     = 1 * dataPageSlotUnit
	dataPageBannerSlotCost    = 32 // 3.2 operations
)

type dataPageItemKind int

const (
	dataPageItemBadge dataPageItemKind = iota
	dataPageItemOperation
)

type dataPageItem struct {
	Kind     dataPageItemKind
	DateTime time.Time
	Tx       detaildomain.ReportSecondPageTransaction
}

func maxDataPageTransactionCount(data map[string]any, offset int) int {
	return countDataPageTransactions(data, offset, dataPageCapacitySlots)
}

func maxDataPageTransactionCountWithSlots(data map[string]any, offset int, maxSlots int) int {
	return countDataPageTransactions(data, offset, maxSlots)
}

func countDataPageTransactions(data map[string]any, offset int, maxSlots int) int {
	if maxSlots <= 0 {
		maxSlots = dataPageCapacitySlots
	}

	total := detaildomain.CountReportTransactions(data)
	if offset >= total {
		return 0
	}

	transactions := detaildomain.ReportTransactions(data, offset, total-offset)
	prevDate := previousTransactionDate(data, offset)

	used := 0
	count := 0

	for _, tx := range transactions {
		cost := dataPageOperationSlotCost
		if !prevDate.IsZero() && !detaildomain.SameReportCalendarDay(prevDate, tx.DateTime) {
			cost += dataPageBadgeSlotCost
		}
		if used+cost > maxSlots {
			break
		}

		if cost > dataPageOperationSlotCost {
			used += dataPageBadgeSlotCost
		}

		used += dataPageOperationSlotCost
		count++
		prevDate = tx.DateTime
	}

	return count
}

func dataPageSlotsUsed(data map[string]any, offset, limit int) int {
	if limit <= 0 {
		return 0
	}

	transactions := detaildomain.ReportTransactions(data, offset, limit)
	prevDate := previousTransactionDate(data, offset)

	used := 0
	for _, tx := range transactions {
		if !prevDate.IsZero() && !detaildomain.SameReportCalendarDay(prevDate, tx.DateTime) {
			used += dataPageBadgeSlotCost
		}
		used += dataPageOperationSlotCost
		prevDate = tx.DateTime
	}

	return used
}

func previousTransactionDate(data map[string]any, offset int) time.Time {
	if offset <= 0 {
		return time.Time{}
	}

	prev := detaildomain.ReportTransactions(data, offset-1, 1)
	if len(prev) == 0 {
		return time.Time{}
	}

	return prev[0].DateTime
}

func buildDataPageItems(data map[string]any, offset, limit int) []dataPageItem {
	if limit <= 0 {
		return nil
	}

	transactions := detaildomain.ReportTransactions(data, offset, limit)
	return buildDataPageItemsForTransactions(transactions, previousTransactionDate(data, offset))
}

func buildDataPageItemsForTransactions(
	transactions []detaildomain.ReportSecondPageTransaction,
	prevDate time.Time,
) []dataPageItem {
	items := make([]dataPageItem, 0, len(transactions)*2)

	for _, tx := range transactions {
		if !prevDate.IsZero() && !detaildomain.SameReportCalendarDay(prevDate, tx.DateTime) {
			items = append(items, dataPageItem{
				Kind:     dataPageItemBadge,
				DateTime: tx.DateTime,
			})
		}

		items = append(items, dataPageItem{
			Kind: dataPageItemOperation,
			Tx:   tx,
		})
		prevDate = tx.DateTime
	}

	return items
}
