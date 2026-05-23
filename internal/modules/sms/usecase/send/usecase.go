package send

import (
	"context"

	"project/internal/modules/sms/domain"
	"project/internal/modules/sms/templates"
)

type UseCase struct {
	sender   domain.Sender
	registry *templates.Registry
}

func New(sender domain.Sender, registry *templates.Registry) *UseCase {
	return &UseCase{
		sender:   sender,
		registry: registry,
	}
}

func (uc *UseCase) Execute(ctx context.Context, input Input) error {
	message, err := uc.registry.Render(input.Bank, input.Data)
	if err != nil {
		return err
	}
	message.Bank = string(input.Bank)

	return uc.sender.Send(ctx, message)
}
