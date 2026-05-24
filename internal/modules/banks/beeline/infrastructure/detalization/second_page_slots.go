package detalization

import (
	detaildomain "project/internal/modules/banks/beeline/detalization"
)

const templateSecondPageSectionDate = "16 марта 2026 г."

var secondPageOperationSlots = []operationSlot{
	{titleY: "23", dateY: "24", descY: "25", amountClass: "xe", title: "плати с билайн: перевод на баланс билайн", date: "16 мар. 2026 04:55", description: "основной баланс", amount: "-400,93 ₽"},
	{titleY: "26", dateY: "27", descY: "28", amountClass: "xf", title: "sms free8464", date: "16 мар. 2026 04:55", description: "1 шт (основной баланс)", amount: "0,00 ₽"},
	{titleY: "29", dateY: "2a", descY: "2b", amountClass: "x10", title: "списание за мобильную коммерцию", date: "16 мар. 2026 02:28", description: "основной баланс", amount: "-12 780,00 ₽"},
	{titleY: "2c", dateY: "2d", descY: "2e", amountClass: "xf", title: "sms free8464", date: "16 мар. 2026 02:28", description: "1 шт (основной баланс)", amount: "0,00 ₽"},
	{titleY: "2f", dateY: "30", descY: "31", amountClass: "x11", title: "пополнение баланса", date: "16 мар. 2026 02:25", description: "основной баланс", amount: "+13 000,00 ₽"},
	{titleY: "32", dateY: "33", descY: "34", amountClass: "x10", title: "списание за мобильную коммерцию", date: "16 мар. 2026 02:20", description: "основной баланс", amount: "-12 780,00 ₽"},
	{titleY: "35", dateY: "36", descY: "37", amountClass: "xf", title: "sms free8464", date: "16 мар. 2026 02:20", description: "1 шт (основной баланс)", amount: "0,00 ₽"},
	{titleY: "38", dateY: "39", descY: "3a", amountClass: "x12", title: "пополнение баланса", date: "16 мар. 2026 02:19", description: "основной баланс", amount: "+12 500,00 ₽"},
	{titleY: "3b", dateY: "3c", descY: "3d", amountClass: "x10", title: "списание за мобильную коммерцию", date: "16 мар. 2026 02:16", description: "основной баланс", amount: "-12 780,00 ₽"},
	{titleY: "3e", dateY: "3f", descY: "40", amountClass: "xf", title: "sms free8464", date: "16 мар. 2026 02:16", description: "1 шт (основной баланс)", amount: "0,00 ₽"},
	{titleY: "41", dateY: "42", descY: "43", amountClass: "x13", title: "пополнение баланса", date: "16 мар. 2026 02:15", description: "основной баланс", amount: "+1 800,00 ₽"},
	{titleY: "44", dateY: "45", descY: "46", amountClass: "x14", title: "плати с билайн: перевод на баланс билайн", date: "16 мар. 2026 01:28", description: "основной баланс", amount: "-3 349,50 ₽"},
	{titleY: "47", dateY: "48", descY: "49", amountClass: "xf", title: "sms free8464", date: "16 мар. 2026 01:28", description: "1 шт (основной баланс)", amount: "0,00 ₽"},
}

func applySecondPageTransactions(html string, data map[string]any, offset, limit int) string {
	transactions := detaildomain.ReportTransactions(data, offset, limit)
	return applyOperationSlots(html, secondPageOperationSlots, transactions, "y22", templateSecondPageSectionDate)
}
