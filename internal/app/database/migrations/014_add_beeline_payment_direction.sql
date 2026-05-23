ALTER TABLE beeline_payments
ADD COLUMN IF NOT EXISTS direction TEXT NOT NULL DEFAULT 'outgoing'
CHECK (direction IN ('incoming', 'outgoing'));

UPDATE beeline_payments
SET direction = 'outgoing'
WHERE direction IS NULL OR direction = '';
