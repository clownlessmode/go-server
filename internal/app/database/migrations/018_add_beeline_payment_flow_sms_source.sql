ALTER TABLE beeline_payments DROP CONSTRAINT IF EXISTS beeline_payments_source_check;

ALTER TABLE beeline_payments
ADD CONSTRAINT beeline_payments_source_check
CHECK (source IN ('manual', 'payment_flow', 'payment_flow_sms'));
