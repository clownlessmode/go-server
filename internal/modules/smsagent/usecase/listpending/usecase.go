package listpending

import (
	"context"

	"project/internal/modules/smsagent/domain"
)

type Input struct {
	Limit int
}

type Output struct {
	Messages []domain.OutboundMessage
}

type UseCase struct {
	repo domain.Repository
}

func New(repo domain.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Execute(ctx context.Context, input Input) (*Output, error) {
	messages, err := uc.repo.ListPending(ctx, input.Limit)
	if err != nil {
		return nil, err
	}

	return &Output{Messages: messages}, nil
}
