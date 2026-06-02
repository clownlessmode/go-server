CREATE TABLE IF NOT EXISTS beeline_operations (
	id TEXT PRIMARY KEY,
	sim_number TEXT NOT NULL REFERENCES beeline_sims (number) ON DELETE CASCADE,
	kind TEXT NOT NULL CHECK (kind IN ('incoming_sms', 'outgoing_sms')),
	number TEXT NOT NULL DEFAULT '',
	paid_at TIMESTAMPTZ NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS beeline_operations_sim_number_idx
ON beeline_operations (sim_number, paid_at DESC);
