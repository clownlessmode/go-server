package updateclientinfo

import (
	"time"

	"project/internal/modules/banks/rocketbank/domain"
)

type Output struct {
	Balance    *float64
	ClientInfo domain.ClientInfo
	History    []domain.HistoryItem
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
