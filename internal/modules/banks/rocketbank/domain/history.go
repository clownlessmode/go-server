package domain

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	HistoryTimeLayout       = "2006-01-02T15:04:05-0700"
	HistoryTypeCashTransfer = "cash-transfer"
	HistoryTypeSBPTransfer  = "sbp-transfer"
	HistoryTypeCardTransfer = "card-transfer"
)

type HistoryItem struct {
	Type                string  `json:"type"`
	Amount              float64 `json:"amount"`
	BalanceBefore       float64 `json:"balanceBefore"`
	Direction           string  `json:"direction"`
	Time                string  `json:"time"`
	OperationFirstName  string  `json:"operationFirstName,omitempty"`
	OperationMiddleName string  `json:"operationMiddleName,omitempty"`
	OperationLastName   string  `json:"operationLastName,omitempty"`
	BankID              string  `json:"bankId,omitempty"`
	PhoneNumber         string  `json:"phoneNumber,omitempty"`
	RecipientCardNumber string  `json:"recipientCardNumber,omitempty"`
}

type CashTransferInput struct {
	Amount        float64
	BalanceBefore float64
	Direction     string
	Time          string
}

type SBPTransferInput struct {
	Amount              float64
	BalanceBefore       float64
	Direction           string
	Time                string
	OperationFirstName  string
	OperationMiddleName string
	OperationLastName   string
	BankID              string
	PhoneNumber         string
}

type CardTransferInput struct {
	Amount              float64
	BalanceBefore       float64
	Direction           string
	Time                string
	BankID              string
	RecipientCardNumber string
}

type SBPTransferBank struct {
	ID       string
	FullName string
	IconURL  string
	Name     string
}

var sbpTransferBanks = []SBPTransferBank{
	{ID: "vtb", FullName: "БАНК ВТБ", IconURL: "https://cdn.lifetechx.ru/icons/banks/icon_square/vtb_square.png", Name: "ВТБ"},
	{ID: "sberbank", FullName: "ПАО СБЕРБАНК", IconURL: "https://cdn.lifetechx.ru/icons/banks/icon_square/sberbank_square.png", Name: "Сбербанк"},
	{ID: "tbank", FullName: "АО «ТБАНК»", IconURL: "https://cdn.lifetechx.ru/icons/banks/icon_square/tbank_square.png", Name: "Т-Банк"},
	{ID: "ozon", FullName: "ООО «ОЗОН БАНК»", IconURL: "https://cdn.lifetechx.ru/icons/banks/icon_square/ozon_square.png", Name: "Озон Банк"},
	{ID: "psb", FullName: "ПАО БАНК ПСБ", IconURL: "https://cdn.lifetechx.ru/icons/banks/icon_square/unknown_square.png", Name: "ПСБ"},
	{ID: "wb", FullName: "ООО «ВАЙЛДБЕРРИЗ БАНК»", IconURL: "https://cdn.lifetechx.ru/icons/banks/icon_square/wb-bank_square.png", Name: "Вайлдберриз Банк"},
	{ID: "alfabank", FullName: "АО «АЛЬФА-БАНК»", IconURL: "https://cdn.lifetechx.ru/icons/banks/icon_square/alfabank_square.png", Name: "Альфа-Банк"},
	{ID: "sovcombank", FullName: "ПАО «СОВКОМБАНК»", IconURL: "https://cdn.lifetechx.ru/icons/banks/icon_square/sovcombank_square.png", Name: "Совкомбанк"},
	{ID: "dvbank", FullName: "АО «ДАЛЬНЕВОСТОЧНЫЙ БАНК»", IconURL: "https://cdn.lifetechx.ru/icons/banks/icon_square/dvbank_square.png", Name: "Дальневосточный Банк"},
	{ID: "raiffeisen", FullName: "РАЙФФАЙЗЕНБАНК", IconURL: "https://cdn.lifetechx.ru/icons/banks/icon_square/raiffeisen_square.png", Name: "Райффайзен"},
	{ID: "promsvyazbank", FullName: "ПРОМСВЯЗЬБАНК", IconURL: "https://cdn.lifetechx.ru/icons/banks/icon_square/promsvyazbank_square.png", Name: "Промсвязьбанк"},
	{ID: "gazprombank", FullName: "БАНК ГПБ (АО)", IconURL: "https://cdn.lifetechx.ru/icons/banks/icon_square/gazprombank_square.png", Name: "Газпромбанк"},
	{ID: "akbars", FullName: "ПАО \"АК БАРС\" БАНК", IconURL: "https://cdn.lifetechx.ru/icons/banks/icon_square/akbars_square.png", Name: "Ак Барс"},
	{ID: "other", FullName: "ДРУГОЙ БАНК", IconURL: "https://cdn.lifetechx.ru/icons/banks/icon_square/unknown_square.png", Name: "Другой"},
}

func NewCashTransferHistoryItem(input CashTransferInput) HistoryItem {
	return HistoryItem{
		Type:          HistoryTypeCashTransfer,
		Amount:        NormalizeHistoryAmount(input.Amount),
		BalanceBefore: NormalizeHistoryAmount(input.BalanceBefore),
		Direction:     NormalizeHistoryDirection(input.Direction),
		Time:          strings.TrimSpace(input.Time),
	}
}

func NewSBPTransferHistoryItem(input SBPTransferInput) HistoryItem {
	return HistoryItem{
		Type:                HistoryTypeSBPTransfer,
		Amount:              NormalizeHistoryAmount(input.Amount),
		BalanceBefore:       NormalizeHistoryAmount(input.BalanceBefore),
		Direction:           NormalizeHistoryDirection(input.Direction),
		Time:                strings.TrimSpace(input.Time),
		OperationFirstName:  strings.TrimSpace(input.OperationFirstName),
		OperationMiddleName: strings.TrimSpace(input.OperationMiddleName),
		OperationLastName:   strings.TrimSpace(input.OperationLastName),
		BankID:              strings.ToLower(strings.TrimSpace(input.BankID)),
		PhoneNumber:         strings.TrimSpace(input.PhoneNumber),
	}
}

func NewCardTransferHistoryItem(input CardTransferInput) HistoryItem {
	return HistoryItem{
		Type:                HistoryTypeCardTransfer,
		Amount:              NormalizeHistoryAmount(input.Amount),
		BalanceBefore:       NormalizeHistoryAmount(input.BalanceBefore),
		Direction:           NormalizeHistoryDirection(input.Direction),
		Time:                strings.TrimSpace(input.Time),
		BankID:              strings.ToLower(strings.TrimSpace(input.BankID)),
		RecipientCardNumber: strings.TrimSpace(input.RecipientCardNumber),
	}
}

func CashTransferTransactionID(input CashTransferInput) string {
	transactionTime, err := NormalizeHistoryTime(input.Time)
	if err != nil {
		transactionTime = strings.TrimSpace(input.Time)
	}

	seed := fmt.Sprintf("%.2f|%s|%s", NormalizeHistoryAmount(input.Amount), NormalizeHistoryDirection(input.Direction), transactionTime)
	sum := sha256.Sum256([]byte(seed))
	value := binary.BigEndian.Uint64(sum[:8]) % 100000000000

	return fmt.Sprintf("T%011d", value)
}

func SBPTransferTransactionID(input SBPTransferInput) string {
	transactionTime, err := NormalizeHistoryTime(input.Time)
	if err != nil {
		transactionTime = strings.TrimSpace(input.Time)
	}

	seed := fmt.Sprintf(
		"%.2f|%s|%s|%s|%s|%s|%s",
		NormalizeHistoryAmount(input.Amount),
		NormalizeHistoryDirection(input.Direction),
		transactionTime,
		strings.TrimSpace(input.OperationFirstName),
		strings.TrimSpace(input.OperationMiddleName),
		strings.TrimSpace(input.OperationLastName),
		strings.ToLower(strings.TrimSpace(input.BankID)),
	)
	sum := sha256.Sum256([]byte(seed))
	value := binary.BigEndian.Uint64(sum[:8]) % 100000000000

	if NormalizeHistoryDirection(input.Direction) == "INCOMING" {
		return fmt.Sprintf("T%011d", value)
	}

	return fmt.Sprintf("M%011d", value)
}

func CardTransferTransactionID(input CardTransferInput) string {
	transactionTime, err := NormalizeHistoryTime(input.Time)
	if err != nil {
		transactionTime = strings.TrimSpace(input.Time)
	}

	seed := fmt.Sprintf(
		"card-transfer|%.2f|%s|%s|%s|%s",
		NormalizeHistoryAmount(input.Amount),
		NormalizeHistoryDirection(input.Direction),
		transactionTime,
		strings.ToLower(strings.TrimSpace(input.BankID)),
		strings.TrimSpace(input.RecipientCardNumber),
	)
	sum := sha256.Sum256([]byte(seed))
	value := binary.BigEndian.Uint64(sum[:8]) % 100000000000

	return fmt.Sprintf("M%011d", value)
}

func HistoryItemID(item HistoryItem) string {
	switch item.Type {
	case HistoryTypeCashTransfer:
		return CashTransferTransactionID(CashTransferInput{
			Amount:    item.Amount,
			Direction: item.Direction,
			Time:      item.Time,
		})
	case HistoryTypeSBPTransfer:
		return SBPTransferTransactionID(SBPTransferInput{
			Amount:              item.Amount,
			Direction:           item.Direction,
			Time:                item.Time,
			OperationFirstName:  item.OperationFirstName,
			OperationMiddleName: item.OperationMiddleName,
			OperationLastName:   item.OperationLastName,
			BankID:              item.BankID,
		})
	case HistoryTypeCardTransfer:
		return CardTransferTransactionID(CardTransferInput{
			Amount:              item.Amount,
			BalanceBefore:       item.BalanceBefore,
			Direction:           item.Direction,
			Time:                item.Time,
			BankID:              item.BankID,
			RecipientCardNumber: item.RecipientCardNumber,
		})
	default:
		return ""
	}
}

func SBPTransferOperationID(item HistoryItem) string {
	prefix := "B"
	separator := "I0B"
	if NormalizeHistoryDirection(item.Direction) == "INCOMING" {
		prefix = "A"
		separator = "D0B"
	}

	sum := sha256.Sum256([]byte(SBPTransferOperationSeed(item)))
	left := binary.BigEndian.Uint64(sum[:8]) % 100000000000000
	right := binary.BigEndian.Uint64(sum[8:16]) % 100000000000000

	return fmt.Sprintf("%s613%014d%s%014d", prefix, left, separator, right)
}

func SBPTransferOperationSeed(item HistoryItem) string {
	transactionTime, err := NormalizeHistoryTime(item.Time)
	if err != nil {
		transactionTime = strings.TrimSpace(item.Time)
	}

	return fmt.Sprintf(
		"sbp-operation|%s|%.2f|%.2f|%s|%s|%s|%s|%s|%s|%s",
		HistoryItemID(item),
		NormalizeHistoryAmount(item.Amount),
		NormalizeHistoryAmount(item.BalanceBefore),
		NormalizeHistoryDirection(item.Direction),
		transactionTime,
		strings.TrimSpace(item.OperationFirstName),
		strings.TrimSpace(item.OperationMiddleName),
		strings.TrimSpace(item.OperationLastName),
		strings.ToLower(strings.TrimSpace(item.BankID)),
		strings.TrimSpace(item.PhoneNumber),
	)
}

func CashTransferInputFromHistoryItem(item HistoryItem) (CashTransferInput, bool) {
	if item.Type != HistoryTypeCashTransfer {
		return CashTransferInput{}, false
	}

	return CashTransferInput{
		Amount:        item.Amount,
		BalanceBefore: item.BalanceBefore,
		Direction:     NormalizeHistoryDirection(item.Direction),
		Time:          item.Time,
	}, true
}

func SBPTransferInputFromHistoryItem(item HistoryItem) (SBPTransferInput, bool) {
	if item.Type != HistoryTypeSBPTransfer {
		return SBPTransferInput{}, false
	}

	return SBPTransferInput{
		Amount:              item.Amount,
		BalanceBefore:       item.BalanceBefore,
		Direction:           NormalizeHistoryDirection(item.Direction),
		Time:                item.Time,
		OperationFirstName:  item.OperationFirstName,
		OperationMiddleName: item.OperationMiddleName,
		OperationLastName:   item.OperationLastName,
		BankID:              item.BankID,
		PhoneNumber:         item.PhoneNumber,
	}, true
}

func CardTransferInputFromHistoryItem(item HistoryItem) (CardTransferInput, bool) {
	if item.Type != HistoryTypeCardTransfer {
		return CardTransferInput{}, false
	}

	return CardTransferInput{
		Amount:              item.Amount,
		BalanceBefore:       item.BalanceBefore,
		Direction:           NormalizeHistoryDirection(item.Direction),
		Time:                item.Time,
		BankID:              item.BankID,
		RecipientCardNumber: item.RecipientCardNumber,
	}, true
}

func HistoryItemFromLegacyOperation(item map[string]any) (HistoryItem, bool) {
	input, ok := CashTransferInputFromLegacyOperation(item)
	if ok {
		return NewCashTransferHistoryItem(input), true
	}

	sbpInput, ok := SBPTransferInputFromLegacyOperation(item)
	if ok {
		return NewSBPTransferHistoryItem(sbpInput), true
	}

	return HistoryItem{}, false
}

func CashTransferInputFromLegacyOperation(item map[string]any) (CashTransferInput, bool) {
	operationType, ok := item["operationType"].(string)
	if !ok || operationType != "D_CASH_ATM" {
		return CashTransferInput{}, false
	}

	mainAmount, ok := item["mainAmount"].(map[string]any)
	if !ok {
		return CashTransferInput{}, false
	}

	amount, ok := numberFromHistoryValue(mainAmount["amount"])
	if !ok {
		return CashTransferInput{}, false
	}

	direction, ok := mainAmount["direction"].(string)
	if !ok {
		return CashTransferInput{}, false
	}

	transactionTime, ok := item["transactionDateTime"].(string)
	if !ok {
		return CashTransferInput{}, false
	}

	return CashTransferInput{
		Amount:        amount,
		BalanceBefore: 0,
		Direction:     NormalizeHistoryDirection(direction),
		Time:          transactionTime,
	}, true
}

func SBPTransferInputFromLegacyOperation(item map[string]any) (SBPTransferInput, bool) {
	operationType, ok := item["operationType"].(string)
	if !ok || (operationType != "D_BYPHONE_SBP_TRANSFER" && operationType != "C_BYPHONE_SBP_TRANSFER") {
		return SBPTransferInput{}, false
	}

	mainAmount, ok := item["mainAmount"].(map[string]any)
	if !ok {
		return SBPTransferInput{}, false
	}

	amount, ok := numberFromHistoryValue(mainAmount["amount"])
	if !ok {
		return SBPTransferInput{}, false
	}

	direction, ok := mainAmount["direction"].(string)
	if !ok {
		return SBPTransferInput{}, false
	}

	transactionTime, ok := item["transactionDateTime"].(string)
	if !ok {
		return SBPTransferInput{}, false
	}

	operationName, _ := item["operationName"].(string)

	return SBPTransferInput{
		Amount:             amount,
		BalanceBefore:      0,
		Direction:          NormalizeHistoryDirection(direction),
		Time:               transactionTime,
		OperationFirstName: strings.TrimSpace(operationName),
		BankID:             "other",
	}, true
}

func LegacyHistoryOperationID(item map[string]any) string {
	detailAction, ok := item["detailAction"].(map[string]any)
	if !ok {
		return ""
	}

	transactionID, ok := detailAction["transactionId"].(string)
	if !ok {
		return ""
	}

	return transactionID
}

func IsValidHistoryDirection(direction string) bool {
	normalized := strings.ToUpper(strings.TrimSpace(direction))
	return normalized == "OUTGOING" || normalized == "INCOMING"
}

func IsValidCashTransferBalance(amount float64, balanceBefore float64, direction string) bool {
	if NormalizeHistoryDirection(direction) != "OUTGOING" {
		return true
	}

	return NormalizeHistoryAmount(balanceBefore) >= NormalizeHistoryAmount(amount)
}

func IsValidSBPTransferBankID(bankID string) bool {
	_, ok := SBPTransferBankByID(bankID)
	return ok
}

func IsValidCardTransferBankID(bankID string) bool {
	return IsValidSBPTransferBankID(bankID)
}

func SBPTransferBankByID(bankID string) (SBPTransferBank, bool) {
	normalized := strings.ToLower(strings.TrimSpace(bankID))
	for _, bank := range sbpTransferBanks {
		if bank.ID == normalized {
			return bank, true
		}
	}

	return SBPTransferBank{}, false
}

func NormalizeHistoryDirection(direction string) string {
	return strings.ToUpper(strings.TrimSpace(direction))
}

func NormalizeHistoryAmount(amount float64) float64 {
	return math.Round(amount*100) / 100
}

func NormalizeHistoryTime(value string) (string, error) {
	parsedTime, err := time.Parse(HistoryTimeLayout, strings.TrimSpace(value))
	if err != nil {
		return "", err
	}

	return parsedTime.UTC().Format(HistoryTimeLayout), nil
}

func numberFromHistoryValue(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case json.Number:
		number, err := typed.Float64()
		return number, err == nil
	default:
		return 0, false
	}
}
