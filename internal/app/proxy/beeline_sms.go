package proxy

import (
	"context"
	"time"

	smsdomain "project/internal/modules/sms/domain"
	"project/internal/modules/sms/templates"
	smssend "project/internal/modules/sms/usecase/send"
	"project/internal/modules/banks/beeline/usecase/recordpaymentflow"
)

const beelineMobileCommerceDelay = 30 * time.Second

func (s *Service) sendBeelinePaymentSMS() {
	s.beelinePaymentMu.Lock()
	snapshot := s.beelinePaymentContext.finalize()
	s.beelinePaymentMu.Unlock()

	total := snapshot.totalAmount()
	if total == nil || snapshot.Amount == nil || snapshot.Commission == nil || snapshot.ReceiverCard == "" {
		proxyLog.Warnf("beeline sms skipped: missing payment data")
		return
	}

	smsPaidAt := time.Now()
	s.recordBeelinePaymentFlowSMS(smsPaidAt)

	if s.smsSend != nil && s.smsCfg.Enabled {
		err := s.smsSend.Execute(context.Background(), smssend.Input{
			Bank: smsdomain.BankBeeline,
			Data: templates.BeelinePaymentData{
				TotalAmount:  *total,
				Commission:   *snapshot.Commission,
				ReceiverCard: snapshot.ReceiverCard,
			},
		})
		if err != nil {
			proxyLog.Warnf("beeline sms send failed: err=%v", err)
		} else {
			proxyLog.Successf("beeline sms sent")
		}
	}

	commercePaidAt := smsPaidAt.Add(beelineMobileCommerceDelay)
	go s.recordBeelinePaymentFlowDelayed(snapshot, commercePaidAt)
}

func (s *Service) recordBeelinePaymentFlowSMS(paidAt time.Time) {
	simNumber := s.beelineSimForProxy(context.Background())
	if s.recordPaymentFlow == nil || simNumber == "" {
		if simNumber == "" {
			proxyLog.Warnf("beeline payment sms skipped: active sim unknown")
		}
		return
	}

	out, err := s.recordPaymentFlow.ExecuteSMS(context.Background(), recordpaymentflow.SMSInput{
		SimNumber: simNumber,
		PaidAt:    paidAt,
	})
	if err != nil {
		proxyLog.Warnf("beeline payment sms save failed: err=%v", err)
		return
	}

	proxyLog.Infof("beeline payment sms saved: sim=%s id=%s number=%s", simNumber, out.Payment.ID, out.Payment.ReceiverCard)
}

func (s *Service) recordBeelinePaymentFlowDelayed(snapshot beelinePaymentSnapshot, paidAt time.Time) {
	delay := time.Until(paidAt)
	if delay > 0 {
		proxyLog.Infof("beeline payment flow: waiting %s before mobile commerce charge", delay.Round(time.Second))
		time.Sleep(delay)
	}

	s.recordBeelinePaymentFlow(snapshot, paidAt)
}

func (s *Service) recordBeelinePaymentFlow(snapshot beelinePaymentSnapshot, paidAt time.Time) {
	simNumber := s.beelineSimForProxy(context.Background())
	if s.recordPaymentFlow == nil || simNumber == "" || snapshot.Amount == nil || snapshot.Commission == nil || snapshot.ReceiverCard == "" {
		if simNumber == "" {
			proxyLog.Warnf("beeline payment history skipped: active sim unknown")
		}
		return
	}

	out, err := s.recordPaymentFlow.Execute(context.Background(), recordpaymentflow.Input{
		SimNumber:    simNumber,
		ReceiverCard: snapshot.ReceiverCard,
		Amount:       *snapshot.Amount,
		Commission:   *snapshot.Commission,
		PaidAt:       paidAt,
	})
	if err != nil {
		proxyLog.Warnf("beeline payment history save failed: err=%v", err)
		return
	}

	proxyLog.Infof("beeline payment saved: sim=%s id=%s total=%.2f", simNumber, out.Payment.ID, out.Payment.Total)
}
