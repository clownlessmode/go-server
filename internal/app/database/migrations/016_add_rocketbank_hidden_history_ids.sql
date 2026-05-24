ALTER TABLE rocketbank_configs
	ADD COLUMN IF NOT EXISTS hidden_history_ids JSONB NOT NULL DEFAULT '[]'::jsonb;
