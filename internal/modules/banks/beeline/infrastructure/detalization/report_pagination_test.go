package detalization

import (
	"testing"

	detaildomain "project/internal/modules/banks/beeline/detalization"
)

func TestBuildTransactionPagePlan(t *testing.T) {
	tests := []struct {
		name  string
		total int
		want  []struct {
			kind       transactionPageKind
			offset     int
			limit      int
			page       int
			showBanner bool
			bannerOnly bool
		}
	}{
		{
			name:  "single second page plus banner page",
			total: 13,
			want: []struct {
				kind       transactionPageKind
				offset     int
				limit      int
				page       int
				showBanner bool
				bannerOnly bool
			}{
				{transactionPageSecond, 0, 13, 2, false, false},
				{transactionPageData, 13, 0, 3, true, true},
			},
		},
		{
			name:  "second page plus tail data page with banner",
			total: 14,
			want: []struct {
				kind       transactionPageKind
				offset     int
				limit      int
				page       int
				showBanner bool
				bannerOnly bool
			}{
				{transactionPageSecond, 0, 13, 2, false, false},
				{transactionPageData, 13, 1, 3, true, false},
			},
		},
		{
			name:  "multiple data pages with banner on last",
			total: 50,
			want: []struct {
				kind       transactionPageKind
				offset     int
				limit      int
				page       int
				showBanner bool
				bannerOnly bool
			}{
				{transactionPageSecond, 0, 13, 2, false, false},
				{transactionPageData, 13, 17, 3, false, false},
				{transactionPageData, 30, 17, 4, false, false},
				{transactionPageData, 47, 3, 5, true, false},
			},
		},
		{
			name:  "full last data page gets separate banner page",
			total: 47,
			want: []struct {
				kind       transactionPageKind
				offset     int
				limit      int
				page       int
				showBanner bool
				bannerOnly bool
			}{
				{transactionPageSecond, 0, 13, 2, false, false},
				{transactionPageData, 13, 17, 3, false, false},
				{transactionPageData, 30, 17, 4, false, false},
				{transactionPageData, 47, 0, 5, true, true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := makeTransactionsData(tt.total)
			pages := buildTransactionPagePlan(data)
			if len(pages) != len(tt.want) {
				t.Fatalf("pages = %d, want %d", len(pages), len(tt.want))
			}

			for index, page := range pages {
				want := tt.want[index]
				if page.Kind != want.kind {
					t.Fatalf("page %d kind = %v, want %v", index, page.Kind, want.kind)
				}
				if page.Offset != want.offset {
					t.Fatalf("page %d offset = %d, want %d", index, page.Offset, want.offset)
				}
				if page.Limit != want.limit {
					t.Fatalf("page %d limit = %d, want %d", index, page.Limit, want.limit)
				}
				if page.PageNumber != want.page {
					t.Fatalf("page %d number = %d, want %d", index, page.PageNumber, want.page)
				}
				if page.ShowBanner != want.showBanner {
					t.Fatalf("page %d showBanner = %v, want %v", index, page.ShowBanner, want.showBanner)
				}
				if page.BannerOnly != want.bannerOnly {
					t.Fatalf("page %d bannerOnly = %v, want %v", index, page.BannerOnly, want.bannerOnly)
				}
			}
		})
	}
}

func makeTransactionsData(total int) map[string]any {
	transactions := make([]any, 0, total)
	for index := 0; index < total; index++ {
		transactions = append(transactions, map[string]any{
			"dateTime": "2026-03-16T04:55:00",
			"name":     "операция",
			"balances": []any{map[string]any{"code": "coreBalance", "changeValue": -1.0}},
		})
	}

	return map[string]any{"transactions": transactions}
}

func TestReportTransactionsOffset(t *testing.T) {
	data := makeTransactionsData(20)
	rows := detaildomain.ReportTransactions(data, 13, 5)
	if len(rows) != 5 {
		t.Fatalf("rows = %d, want 5", len(rows))
	}
}
