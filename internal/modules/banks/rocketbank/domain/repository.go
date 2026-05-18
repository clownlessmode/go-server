package domain

import "context"

type Repository interface {
	GetConfig(ctx context.Context) (*Config, error)
	UpdateBalance(ctx context.Context, balance *float64) (*Config, error)
	UpdateClientInfo(ctx context.Context, clientInfo ClientInfo) (*Config, error)
	ListHistory(ctx context.Context) ([]HistoryItem, error)
	GetHistoryItem(ctx context.Context, id string) (HistoryItem, error)
	CreateHistoryItem(ctx context.Context, item HistoryItem) (HistoryItem, error)
	UpdateHistoryItem(ctx context.Context, id string, item HistoryItem) (HistoryItem, error)
	DeleteHistoryItem(ctx context.Context, id string) error
	ClearHistory(ctx context.Context) error
}

type ChequeGenerator interface {
	GenerateSBPTransferCheque(item HistoryItem, clientInfo ClientInfo) error
	GenerateCardTransferCheque(item HistoryItem, clientInfo ClientInfo) error
	GenerateMissingSBPTransferCheques(config *Config) error
}
