ALTER TABLE beeline_payments
DROP CONSTRAINT IF EXISTS beeline_payments_direction_check;

ALTER TABLE beeline_payments
ADD CONSTRAINT beeline_payments_direction_check
CHECK (direction IN ('incoming', 'outgoing', 'balance_return'));
