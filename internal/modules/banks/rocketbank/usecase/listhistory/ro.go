package listhistory

import "project/internal/modules/banks/rocketbank/domain"

type Output struct {
	History []domain.HistoryItem
}
