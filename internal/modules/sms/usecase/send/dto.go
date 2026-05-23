package send

import "project/internal/modules/sms/domain"

type Input struct {
	Bank domain.BankCode
	Data any
}
