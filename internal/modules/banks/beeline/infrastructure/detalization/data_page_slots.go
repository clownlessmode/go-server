package detalization

import (
	detaildomain "project/internal/modules/banks/beeline/detalization"
)

var dataPageOperationSlots = []operationSlot{
	{titleY: "4a", dateY: "4b", descY: "4c", amountClass: "x15", title: "плати с билайн: перевод на баланс билайн", date: "16 мар. 2026 01:12", description: "основной баланс", amount: "-14 996,63 ₽"},
	{titleY: "4d", dateY: "4e", descY: "4f", amountClass: "xf", title: "sms free8464", date: "16 мар. 2026 01:12", description: "1 шт (основной баланс)", amount: "0,00 ₽"},
	{titleY: "50", dateY: "51", descY: "52", amountClass: "x15", title: "плати с билайн: перевод на баланс билайн", date: "16 мар. 2026 01:12", description: "основной баланс", amount: "-14 996,63 ₽"},
	{titleY: "53", dateY: "54", descY: "55", amountClass: "xf", title: "sms free8464", date: "16 мар. 2026 01:12", description: "1 шт (основной баланс)", amount: "0,00 ₽"},
	{titleY: "56", dateY: "57", descY: "58", amountClass: "x16", title: "пополнение баланса", date: "16 мар. 2026 01:06", description: "основной баланс", amount: "+160,00 ₽"},
	{titleY: "59", dateY: "5a", descY: "5b", amountClass: "xf", title: "компенсация затрат на пополнение баланса", date: "16 мар. 2026 01:06", description: "основной баланс", amount: "-1,52 ₽"},
	{titleY: "5c", dateY: "5d", descY: "5e", amountClass: "xe", title: "пополнение баланса", date: "16 мар. 2026 01:06", description: "основной баланс", amount: "+127,00 ₽"},
	{titleY: "5f", dateY: "60", descY: "61", amountClass: "x17", title: "компенсация затрат на пополнение баланса", date: "16 мар. 2026 01:06", description: "основной баланс", amount: "-2,40 ₽"},
	{titleY: "62", dateY: "63", descY: "64", amountClass: "x18", title: "пополнение баланса", date: "16 мар. 2026 01:06", description: "основной баланс", amount: "+200,00 ₽"},
	{titleY: "65", dateY: "66", descY: "67", amountClass: "x19", title: "компенсация затрат на пополнение баланса", date: "16 мар. 2026 01:05", description: "основной баланс", amount: "-12,64 ₽"},
	{titleY: "68", dateY: "69", descY: "6a", amountClass: "x1a", title: "пополнение баланса", date: "16 мар. 2026 01:05", description: "основной баланс", amount: "+1 053,00 ₽"},
	{titleY: "6b", dateY: "6c", descY: "6d", amountClass: "x16", title: "пополнение баланса", date: "16 мар. 2026 01:05", description: "основной баланс", amount: "+270,00 ₽"},
	{titleY: "6e", dateY: "6f", descY: "70", amountClass: "x15", title: "пополнение баланса", date: "16 мар. 2026 01:04", description: "основной баланс", amount: "+1 000,00 ₽"},
	{titleY: "71", dateY: "72", descY: "73", amountClass: "x1b", title: "компенсация затрат на пополнение баланса", date: "16 мар. 2026 01:04", description: "основной баланс", amount: "-18,00 ₽"},
	{titleY: "74", dateY: "75", descY: "76", amountClass: "x13", title: "пополнение баланса", date: "16 мар. 2026 01:04", description: "основной баланс", amount: "+1 500,00 ₽"},
	{titleY: "77", dateY: "78", descY: "79", amountClass: "x1a", title: "пополнение баланса", date: "16 мар. 2026 01:03", description: "основной баланс", amount: "+1 584,00 ₽"},
	{titleY: "7a", dateY: "7b", descY: "7c", amountClass: "xe", title: "пополнение баланса", date: "16 мар. 2026 01:03", description: "основной баланс", amount: "+131,00 ₽"},
}

func applyDataPageTransactions(html string, data map[string]any, offset, limit int) string {
	transactions := detaildomain.ReportTransactions(data, offset, limit)
	return applyOperationSlots(html, dataPageOperationSlots, transactions, "", "")
}
