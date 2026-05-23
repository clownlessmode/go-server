package domain

import (
	"context"
	"time"
)

type Repository interface {
	CreateSim(ctx context.Context, sim Sim) (Sim, error)
	EnsureSim(ctx context.Context, number string) (Sim, error)
	ListSims(ctx context.Context) ([]Sim, error)
	GetSim(ctx context.Context, number string) (Sim, error)
	DeleteSim(ctx context.Context, number string) error

	UpdateBalance(ctx context.Context, number string, balance *float64) (Sim, error)
	GetEffectiveBalance(ctx context.Context, number string) (*float64, error)
	FindConfiguredSimAmong(ctx context.Context, numbers []string) (string, bool)
	SumPaymentTotals(ctx context.Context, number string) (float64, error)
	SumIncomingTotals(ctx context.Context, number string) (float64, error)

	CreatePayment(ctx context.Context, number string, payment Payment) (Payment, error)
	ListPayments(ctx context.Context, number string) ([]Payment, error)
	ListPaymentsInPeriod(ctx context.Context, number string, start, end time.Time) ([]Payment, error)
	GetPayment(ctx context.Context, number, id string) (Payment, error)
	UpdatePayment(ctx context.Context, number string, payment Payment) (Payment, error)
	DeletePayment(ctx context.Context, number, id string) error

	SaveDetalizationSnapshot(ctx context.Context, snapshot DetalizationSnapshot) (DetalizationSnapshot, error)
	GetDetalizationSnapshot(ctx context.Context, number string) (DetalizationSnapshot, error)
	HasDetalizationSnapshot(ctx context.Context, number string) (bool, error)
	UpdateDetalizationComputedBalance(ctx context.Context, number string, balance float64) error
}
