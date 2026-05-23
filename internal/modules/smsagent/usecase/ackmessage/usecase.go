package ackmessage

import (
	"context"

	"project/internal/modules/smsagent/domain"
)

type Input struct {
	ID           string
	Status       domain.MessageStatus
	DeviceID     string
	ErrorMessage string
}

type UseCase struct {
	repo domain.Repository
}

func New(repo domain.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Execute(ctx context.Context, input Input) error {
	return uc.repo.Ack(ctx, input.ID, input.Status, input.DeviceID, input.ErrorMessage)
}
