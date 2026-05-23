CREATE TABLE IF NOT EXISTS beeline_sims (
	number TEXT PRIMARY KEY CHECK (number ~ '^\d{10}$'),
	balance DOUBLE PRECISION,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE beeline_payments
ADD COLUMN IF NOT EXISTS sim_number TEXT REFERENCES beeline_sims (number) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS beeline_payments_sim_number_idx
ON beeline_payments (sim_number, paid_at DESC);

DROP TABLE IF EXISTS beeline_configs;
