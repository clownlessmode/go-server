ALTER TABLE beeline_sims
	ADD COLUMN IF NOT EXISTS hidden_transaction_ids JSONB NOT NULL DEFAULT '[]'::jsonb;
