package queue

import (
	"context"

	"project/internal/app/logger"
	smsdomain "project/internal/modules/sms/domain"
	agentdomain "project/internal/modules/smsagent/domain"
)

var queueLog = logger.New("sms-agent-queue")

type Sender struct {
	repo agentdomain.Repository
}

func NewSender(repo agentdomain.Repository) *Sender {
	return &Sender{repo: repo}
}

func (s *Sender) Send(ctx context.Context, message smsdomain.Message) error {
	if message.Address == "" || message.Body == "" {
		return smsdomain.ErrInvalidMessage
	}

	outbound, err := agentdomain.NewOutboundMessage(message.Address, message.Body, message.Bank)
	if err != nil {
		return err
	}

	created, err := s.repo.Enqueue(ctx, outbound)
	if err != nil {
		return err
	}

	queueLog.Successf("sms queued for agent: id=%s address=%s", created.ID, created.Address)
	return nil
}
