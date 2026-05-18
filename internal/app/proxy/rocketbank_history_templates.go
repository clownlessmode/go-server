package proxy

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"project/internal/modules/banks/rocketbank/domain"
)

func rocketbankHistoryOperation(item domain.HistoryItem, clientInfo domain.ClientInfo) (map[string]any, bool) {
	switch item.Type {
	case domain.HistoryTypeCashTransfer:
		return rocketbankCashTransferHistoryOperation(item)
	case domain.HistoryTypeSBPTransfer:
		return rocketbankSBPTransferHistoryOperation(item)
	case domain.HistoryTypeCardTransfer:
		return rocketbankCardTransferHistoryOperation(item, clientInfo)
	default:
		return nil, false
	}
}

func rocketbankHistoryTransactionDetails(item domain.HistoryItem, timezone string, clientInfo domain.ClientInfo) (map[string]any, bool) {
	switch item.Type {
	case domain.HistoryTypeCashTransfer:
		return rocketbankCashTransferTransactionDetails(item, timezone)
	case domain.HistoryTypeSBPTransfer:
		return rocketbankSBPTransferTransactionDetails(item, timezone)
	case domain.HistoryTypeCardTransfer:
		return rocketbankCardTransferTransactionDetails(item, timezone, clientInfo)
	default:
		return nil, false
	}
}

func rocketbankCashTransferHistoryOperation(item domain.HistoryItem) (map[string]any, bool) {
	if item.Type != domain.HistoryTypeCashTransfer {
		return nil, false
	}

	transactionTime, err := domain.NormalizeHistoryTime(item.Time)
	if err != nil {
		return nil, false
	}

	direction := domain.NormalizeHistoryDirection(item.Direction)
	operationName := "Снятие наличных"
	if strings.EqualFold(direction, "INCOMING") {
		operationName = "Внесение наличных"
	}

	return map[string]any{
		"mainIcon": map[string]any{
			"icon": "cash_ATM",
		},
		"mainAmount": map[string]any{
			"amount":    domain.NormalizeHistoryAmount(item.Amount),
			"currency":  "810",
			"direction": direction,
		},
		"detailAction": map[string]any{
			"source":        "RBS",
			"productId":     "13856641038",
			"productType":   "CARD",
			"transactionId": domain.HistoryItemID(item),
		},
		"operationName":       operationName,
		"operationType":       "D_CASH_ATM",
		"transactionDateTime": transactionTime,
	}, true
}

func rocketbankSBPTransferHistoryOperation(item domain.HistoryItem) (map[string]any, bool) {
	if item.Type != domain.HistoryTypeSBPTransfer {
		return nil, false
	}

	transactionTime, err := domain.NormalizeHistoryTime(item.Time)
	if err != nil {
		return nil, false
	}
	bank, ok := domain.SBPTransferBankByID(item.BankID)
	if !ok {
		return nil, false
	}

	direction := domain.NormalizeHistoryDirection(item.Direction)
	operationType := "D_BYPHONE_SBP_TRANSFER"
	iconCode := "mcc_rashod"
	if strings.EqualFold(direction, "INCOMING") {
		operationType = "C_BYPHONE_SBP_TRANSFER"
		iconCode = "mcc_prihod"
	}

	return map[string]any{
		"mainIcon": map[string]any{
			"icon":      iconCode,
			"iconLiter": rocketbankOperationInitial(item.OperationFirstName),
		},
		"mainAmount": map[string]any{
			"amount":    domain.NormalizeHistoryAmount(item.Amount),
			"currency":  "810",
			"direction": direction,
		},
		"detailAction": map[string]any{
			"source":        "RBS",
			"productId":     "14469802482",
			"productType":   "CARD",
			"transactionId": domain.HistoryItemID(item),
		},
		"operationName":       rocketbankSBPTransferListOperationName(item),
		"operationType":       operationType,
		"statusIcon":          map[string]any{"icon": "statusicon_unknownbank", "iconUrl": bank.IconURL},
		"transactionDateTime": transactionTime,
	}, true
}

func rocketbankCardTransferHistoryOperation(item domain.HistoryItem, clientInfo domain.ClientInfo) (map[string]any, bool) {
	if item.Type != domain.HistoryTypeCardTransfer {
		return nil, false
	}

	transactionTime := strings.TrimSpace(item.Time)
	if _, err := time.Parse(domain.HistoryTimeLayout, transactionTime); err != nil {
		return nil, false
	}

	direction := domain.NormalizeHistoryDirection(item.Direction)
	cardNumber := item.RecipientCardNumber
	statusIcon := map[string]any{
		"icon": "statusicon_unknownbank",
	}
	if strings.EqualFold(direction, "INCOMING") {
		if clientInfo.CardNumber != nil {
			cardNumber = *clientInfo.CardNumber
		}
	} else {
		bank, ok := domain.SBPTransferBankByID(item.BankID)
		if !ok {
			return nil, false
		}
		statusIcon["iconUrl"] = bank.IconURL
	}

	cardSuffix := rocketbankCardLastDigits(cardNumber)
	operationName := "На карту"
	if cardSuffix != "" {
		operationName += " " + cardSuffix
	}

	return map[string]any{
		"detailAction": map[string]any{
			"productId":     "14469802482",
			"productType":   "CARD",
			"source":        "RBS",
			"transactionId": domain.HistoryItemID(item),
		},
		"mainAmount": map[string]any{
			"amount":    domain.NormalizeHistoryAmount(item.Amount),
			"currency":  "810",
			"direction": direction,
		},
		"mainIcon": map[string]any{
			"icon": "cardpan_transfer",
		},
		"operationName":       operationName,
		"operationType":       "D_BYPAN_TRANSFER",
		"statusIcon":          statusIcon,
		"transactionDateTime": transactionTime,
	}, true
}

func rocketbankCashTransferTransactionDetails(item domain.HistoryItem, timezone string) (map[string]any, bool) {
	if item.Type != domain.HistoryTypeCashTransfer {
		return nil, false
	}

	transactionTime, err := domain.NormalizeHistoryTime(item.Time)
	if err != nil {
		return nil, false
	}

	parsedTime, err := time.Parse(domain.HistoryTimeLayout, transactionTime)
	if err != nil {
		return nil, false
	}

	direction := domain.NormalizeHistoryDirection(item.Direction)
	operationName := "Снятие наличных"
	balanceAfter := item.BalanceBefore - item.Amount
	if strings.EqualFold(direction, "INCOMING") {
		operationName = "Внесение наличных"
		balanceAfter = item.BalanceBefore + item.Amount
	}

	return map[string]any{
		"balance": map[string]any{
			"after": map[string]any{
				"amount":         domain.NormalizeHistoryAmount(balanceAfter),
				"currencySymbol": "₽",
			},
			"before": map[string]any{
				"amount":         domain.NormalizeHistoryAmount(item.BalanceBefore),
				"currencySymbol": "₽",
			},
			"separator": "→",
		},
		"cheque": map[string]any{
			"allowed": false,
		},
		"mainAmount": map[string]any{
			"amount":         domain.NormalizeHistoryAmount(item.Amount),
			"currencySymbol": "₽",
			"direction":      direction,
		},
		"mainIcon": map[string]any{
			"iconCode": "cash_ATM",
		},
		"operationFields": []any{
			map[string]any{
				"icon": map[string]any{
					"iconCode": "statusicon_done",
				},
				"key":   "operationStatus",
				"title": "Статус платежа",
				"value": "Выполнена",
			},
			map[string]any{
				"key":   "operationTime",
				"title": "Дата и время",
				"value": formatRocketbankOperationTime(parsedTime, timezone),
			},
			map[string]any{
				"isChangeAvailable": false,
				"key":               "category",
				"title":             "Категория",
				"value":             "Наличные",
			},
			map[string]any{
				"key":   "sourceProductName",
				"title": "Счёт списания",
				"value": "Мой счёт",
			},
		},
		"operationName": operationName,
		"operationStatus": map[string]any{
			"status": "ENTRIED",
		},
		"repeatableInfo": map[string]any{
			"repeatable": false,
		},
		"returnableInfo": map[string]any{
			"isReturnable": false,
		},
		"template": map[string]any{
			"isAllowed": false,
		},
	}, true
}

func rocketbankSBPTransferTransactionDetails(item domain.HistoryItem, timezone string) (map[string]any, bool) {
	if item.Type != domain.HistoryTypeSBPTransfer {
		return nil, false
	}

	transactionTime, err := domain.NormalizeHistoryTime(item.Time)
	if err != nil {
		return nil, false
	}

	parsedTime, err := time.Parse(domain.HistoryTimeLayout, transactionTime)
	if err != nil {
		return nil, false
	}

	bank, ok := domain.SBPTransferBankByID(item.BankID)
	if !ok {
		return nil, false
	}

	direction := domain.NormalizeHistoryDirection(item.Direction)
	iconCode := "mcc_rashod"
	balanceAfter := item.BalanceBefore - item.Amount
	accountField := map[string]any{"key": "sourceProductName", "title": "Счёт списания", "value": "Мой счёт"}
	phoneTitle := "Номер телефона получателя"
	bankTitle := "Банк получателя"
	if strings.EqualFold(direction, "INCOMING") {
		iconCode = "mcc_prihod"
		balanceAfter = item.BalanceBefore + item.Amount
		accountField = map[string]any{"key": "destinationProductName", "title": "Счёт зачисления", "value": "Мой счёт"}
		phoneTitle = "Номер телефона отправителя"
		bankTitle = "Банк отправителя"
	}

	operationFields := []any{
		map[string]any{
			"icon": map[string]any{
				"iconCode": "statusicon_done",
			},
			"key":   "operationStatus",
			"title": "Статус платежа",
			"value": "Выполнена",
		},
		map[string]any{
			"key":   "operationTime",
			"title": "Дата и время",
			"value": formatRocketbankOperationTime(parsedTime, timezone),
		},
		map[string]any{
			"isChangeAvailable": !strings.EqualFold(direction, "INCOMING"),
			"key":               "category",
			"title":             "Категория",
			"value":             "Переводы",
		},
		accountField,
		map[string]any{
			"icon": map[string]any{
				"iconCode": "sbp",
			},
			"key":   "phoneTransferMethod",
			"title": "Перевод по номеру телефона",
			"value": "Через СБП",
		},
	}
	if strings.EqualFold(direction, "OUTGOING") {
		operationFields = append([]any{
			map[string]any{
				"key":   "fee",
				"title": "Комиссия",
				"value": "Нет",
			},
		}, operationFields...)
	}
	if strings.TrimSpace(item.PhoneNumber) != "" {
		operationFields = append(operationFields, map[string]any{
			"key":   "phoneNumber",
			"title": phoneTitle,
			"value": strings.TrimSpace(item.PhoneNumber),
		})
	}
	operationFields = append(operationFields,
		map[string]any{
			"icon": map[string]any{
				"iconCode": "statusicon_unknownbank",
				"iconUrl":  bank.IconURL,
			},
			"key":   "bankName",
			"title": bankTitle,
			"value": bank.FullName,
		},
		map[string]any{
			"key":   "sbpOperationId",
			"title": "Номер операции в СБП",
			"value": rocketbankSBPOperationID(item),
		},
	)

	body := map[string]any{
		"balance": map[string]any{
			"after": map[string]any{
				"amount":         domain.NormalizeHistoryAmount(balanceAfter),
				"currencySymbol": "₽",
			},
			"before": map[string]any{
				"amount":         domain.NormalizeHistoryAmount(item.BalanceBefore),
				"currencySymbol": "₽",
			},
			"separator": "→",
		},
		"cheque": map[string]any{
			"allowed":       true,
			"productId":     "14469802482",
			"productType":   "CARD",
			"transactionId": domain.HistoryItemID(item),
		},
		"mainAmount": map[string]any{
			"amount":         domain.NormalizeHistoryAmount(item.Amount),
			"currencySymbol": "₽",
			"direction":      direction,
		},
		"mainIcon": map[string]any{
			"iconCode":  iconCode,
			"iconLiter": rocketbankOperationInitial(item.OperationFirstName),
		},
		"operationFields": operationFields,
		"operationName":   rocketbankSBPTransferDetailsOperationName(item),
		"operationStatus": map[string]any{
			"status": "ENTRIED",
		},
		"repeatableInfo": map[string]any{
			"repeatable": false,
		},
		"returnableInfo": map[string]any{
			"isReturnable": false,
		},
		"template": map[string]any{
			"isAllowed": false,
		},
	}

	if strings.EqualFold(direction, "OUTGOING") {
		operationID := strings.TrimPrefix(domain.HistoryItemID(item), "M")
		body["categoryChange"] = map[string]any{
			"categoryCode": "CAT_TRANSFERS",
			"operationId":  operationID,
		}
		body["repeatableInfo"] = map[string]any{
			"operationId":    operationID,
			"repeatable":     true,
			"repeatableType": "PHONE",
		}
		body["template"] = map[string]any{
			"isAllowed":             true,
			"operationType":         "PHONE",
			"originalOperationUuid": rocketbankDeterministicUUID("sbp-template|" + domain.HistoryItemID(item)),
		}
	} else if strings.TrimSpace(item.PhoneNumber) != "" {
		body["returnableInfo"] = map[string]any{
			"amount":         domain.NormalizeHistoryAmount(item.Amount),
			"bankType":       "SBP",
			"isReturnable":   true,
			"owner":          "ФИЛИАЛ \"КОРПОРАТИВНЫЙ\" ПАО \"СОВКОМБАНК\"",
			"phone":          strings.TrimSpace(item.PhoneNumber),
			"returnableType": "PHONE",
			"sbpCode":        rocketbankSBPCode(item),
			"source":         "CARD##14469802482",
		}
	}

	return body, true
}

func rocketbankCardTransferTransactionDetails(item domain.HistoryItem, timezone string, clientInfo domain.ClientInfo) (map[string]any, bool) {
	if item.Type != domain.HistoryTypeCardTransfer {
		return nil, false
	}

	transactionTime, err := domain.NormalizeHistoryTime(item.Time)
	if err != nil {
		return nil, false
	}

	parsedTime, err := time.Parse(domain.HistoryTimeLayout, transactionTime)
	if err != nil {
		return nil, false
	}

	bank, ok := domain.SBPTransferBankByID(item.BankID)
	if !ok {
		return nil, false
	}

	direction := domain.NormalizeHistoryDirection(item.Direction)
	balanceAfter := item.BalanceBefore - item.Amount
	cardNumber := item.RecipientCardNumber
	accountField := map[string]any{"key": "sourceProductName", "title": "Счёт списания", "value": "Мой счёт"}
	bankTitle := "Банк получателя"
	if strings.EqualFold(direction, "INCOMING") {
		balanceAfter = item.BalanceBefore + item.Amount
		if clientInfo.CardNumber != nil {
			cardNumber = *clientInfo.CardNumber
		}
		accountField = map[string]any{"key": "destinationProductName", "title": "Перевод на карту", "value": "светлая"}
		bankTitle = "Банк отправителя"
	}

	operationFields := []any{
		map[string]any{
			"icon": map[string]any{
				"iconCode": "statusicon_done",
			},
			"key":   "operationStatus",
			"title": "Статус платежа",
			"value": "Выполнена",
		},
		map[string]any{
			"key":   "operationTime",
			"title": "Дата и время",
			"value": formatRocketbankOperationTime(parsedTime, timezone),
		},
		map[string]any{
			"isChangeAvailable": !strings.EqualFold(direction, "INCOMING"),
			"key":               "category",
			"title":             "Категория",
			"value":             "Переводы",
		},
		accountField,
	}
	if strings.EqualFold(direction, "OUTGOING") {
		operationFields = append(operationFields, map[string]any{
			"icon": map[string]any{
				"iconCode": "statusicon_unknownbank",
				"iconUrl":  bank.IconURL,
			},
			"key":   "bankName",
			"title": bankTitle,
			"value": bank.Name,
		})
	}

	body := map[string]any{
		"balance": map[string]any{
			"after": map[string]any{
				"amount":         domain.NormalizeHistoryAmount(balanceAfter),
				"currencySymbol": "₽",
			},
			"before": map[string]any{
				"amount":         domain.NormalizeHistoryAmount(item.BalanceBefore),
				"currencySymbol": "₽",
			},
			"separator": "→",
		},
		"cheque": map[string]any{
			"allowed":       true,
			"productId":     "14469802482",
			"productType":   "CARD",
			"transactionId": domain.HistoryItemID(item),
		},
		"mainAmount": map[string]any{
			"amount":         domain.NormalizeHistoryAmount(item.Amount),
			"currencySymbol": "₽",
			"direction":      direction,
		},
		"mainIcon": map[string]any{
			"iconCode": "cardpan_transfer",
		},
		"operationFields": operationFields,
		"operationName":   "На карту " + rocketbankCardLastDigits(cardNumber),
		"operationStatus": map[string]any{
			"status": "ENTRIED",
		},
		"repeatableInfo": map[string]any{
			"repeatable": false,
		},
		"returnableInfo": map[string]any{
			"isReturnable": false,
		},
		"template": map[string]any{
			"isAllowed": false,
		},
	}

	if strings.EqualFold(direction, "OUTGOING") {
		operationID := strings.TrimPrefix(domain.HistoryItemID(item), "M")
		body["categoryChange"] = map[string]any{
			"categoryCode": "CAT_TRANSFERS",
			"operationId":  operationID,
		}
		body["repeatableInfo"] = map[string]any{
			"operationId":    operationID,
			"repeatable":     true,
			"repeatableType": "CARD",
		}
		body["template"] = map[string]any{
			"isAllowed":             true,
			"operationType":         "CARD",
			"originalOperationUuid": rocketbankDeterministicUUID("card-template|" + domain.HistoryItemID(item)),
		}
	}

	return body, true
}

func rocketbankSBPTransferListOperationName(item domain.HistoryItem) string {
	lastInitial := rocketbankOperationInitial(item.OperationLastName)
	if lastInitial == "" {
		return strings.TrimSpace(item.OperationFirstName)
	}

	return strings.TrimSpace(item.OperationFirstName + " " + lastInitial + ".")
}

func rocketbankSBPTransferDetailsOperationName(item domain.HistoryItem) string {
	lastInitial := rocketbankOperationInitial(item.OperationLastName)

	return strings.TrimSpace(item.OperationFirstName + " " + item.OperationMiddleName + " " + lastInitial)
}

func rocketbankOperationInitial(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	runes := []rune(value)
	return strings.ToUpper(string(runes[0]))
}

func rocketbankCardLastDigits(value string) string {
	digits := strings.Builder{}
	for _, char := range value {
		if char >= '0' && char <= '9' {
			digits.WriteRune(char)
		}
	}

	cardDigits := digits.String()
	if len(cardDigits) <= 4 {
		return cardDigits
	}

	return cardDigits[len(cardDigits)-4:]
}

func rocketbankSBPOperationID(item domain.HistoryItem) string {
	return domain.SBPTransferOperationID(item)
}

func rocketbankSBPCode(item domain.HistoryItem) string {
	sum := sha256.Sum256([]byte("sbp-code|" + strings.ToLower(strings.TrimSpace(item.BankID))))
	value := binary.BigEndian.Uint64(sum[:8]) % 1000

	return fmt.Sprintf("100000000%03d", value)
}

func rocketbankDeterministicUUID(seed string) string {
	sum := sha256.Sum256([]byte(seed))

	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		binary.BigEndian.Uint32(sum[0:4]),
		binary.BigEndian.Uint16(sum[4:6]),
		binary.BigEndian.Uint16(sum[6:8]),
		binary.BigEndian.Uint16(sum[8:10]),
		sum[10:16],
	)
}

func formatRocketbankOperationTime(value time.Time, timezone string) string {
	localTime := value.In(rocketbankTimezoneLocation(timezone))
	months := map[time.Month]string{
		time.January:   "января",
		time.February:  "февраля",
		time.March:     "марта",
		time.April:     "апреля",
		time.May:       "мая",
		time.June:      "июня",
		time.July:      "июля",
		time.August:    "августа",
		time.September: "сентября",
		time.October:   "октября",
		time.November:  "ноября",
		time.December:  "декабря",
	}

	return fmt.Sprintf("%02d %s, %02d:%02d", localTime.Day(), months[localTime.Month()], localTime.Hour(), localTime.Minute())
}

func rocketbankTimezoneLocation(timezone string) *time.Location {
	if len(timezone) != 5 {
		timezone = "+0700"
	}

	sign := 1
	if timezone[0] == '-' {
		sign = -1
	} else if timezone[0] != '+' {
		timezone = "+0700"
	}

	offsetTime, err := time.Parse("-0700", timezone)
	if err != nil {
		offsetTime, _ = time.Parse("-0700", "+0700")
	}

	_, offset := offsetTime.Zone()
	return time.FixedZone(timezone, sign*absInt(offset))
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}

	return value
}
