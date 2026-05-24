package detalization

import (
	detaildomain "project/internal/modules/banks/beeline/detalization"
)

type transactionPageKind int

const (
	transactionPageSecond transactionPageKind = iota
	transactionPageData
)

type TransactionPageParams struct {
	Kind       transactionPageKind
	Offset     int
	Limit      int
	PageNumber int
	ShowBanner bool
	BannerOnly bool
}

func buildTransactionPagePlan(data map[string]any) []TransactionPageParams {
	total := detaildomain.CountReportTransactions(data)
	if total == 0 {
		return nil
	}

	var pages []TransactionPageParams
	pageNumber := 2
	offset := 0

	firstLimit := min(total, detaildomain.ReportSecondPageTransactionLimit)
	pages = append(pages, TransactionPageParams{
		Kind:       transactionPageSecond,
		Offset:     offset,
		Limit:      firstLimit,
		PageNumber: pageNumber,
	})
	offset += firstLimit
	pageNumber++

	bannerReserve := dataPageCapacitySlots - dataPageBannerSlotCost

	for offset < total {
		remaining := total - offset
		chunkWithBanner := maxDataPageTransactionCountWithSlots(data, offset, bannerReserve)
		chunkFull := maxDataPageTransactionCount(data, offset)

		if chunkFull <= 0 {
			break
		}

		if remaining <= chunkWithBanner {
			pages = append(pages, TransactionPageParams{
				Kind:       transactionPageData,
				Offset:     offset,
				Limit:      remaining,
				PageNumber: pageNumber,
				ShowBanner: true,
			})
			return pages
		}

		if remaining <= chunkFull {
			pages = append(pages, TransactionPageParams{
				Kind:       transactionPageData,
				Offset:     offset,
				Limit:      remaining,
				PageNumber: pageNumber,
			})
			pageNumber++
			pages = append(pages, bannerOnlyPage(pageNumber, total))
			return pages
		}

		pages = append(pages, TransactionPageParams{
			Kind:       transactionPageData,
			Offset:     offset,
			Limit:      chunkFull,
			PageNumber: pageNumber,
		})
		offset += chunkFull
		pageNumber++
	}

	if len(pages) == 1 {
		pages = append(pages, bannerOnlyPage(pageNumber, total))
	}

	return pages
}

func bannerOnlyPage(pageNumber, total int) TransactionPageParams {
	return TransactionPageParams{
		Kind:       transactionPageData,
		Offset:     total,
		Limit:      0,
		PageNumber: pageNumber,
		ShowBanner: true,
		BannerOnly: true,
	}
}

func ensureTransactionPages(pages []TransactionPageParams) []TransactionPageParams {
	if len(pages) > 0 {
		return pages
	}

	return []TransactionPageParams{{
		Kind:       transactionPageSecond,
		Offset:     0,
		Limit:      0,
		PageNumber: 2,
	}}
}
