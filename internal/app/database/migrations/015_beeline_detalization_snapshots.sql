CREATE TABLE IF NOT EXISTS beeline_detalization_snapshots (
	sim_number TEXT PRIMARY KEY REFERENCES beeline_sims (number) ON DELETE CASCADE,
	period_start TIMESTAMPTZ NOT NULL,
	period_end TIMESTAMPTZ NOT NULL,
	snapshot JSONB NOT NULL,
	computed_balance DOUBLE PRECISION,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
