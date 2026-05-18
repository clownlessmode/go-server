ALTER TABLE rocketbank_configs
	ADD COLUMN IF NOT EXISTS phone_number TEXT,
	ADD COLUMN IF NOT EXISTS card_number TEXT;
