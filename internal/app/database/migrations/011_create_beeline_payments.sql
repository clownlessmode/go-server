CREATE TABLE IF NOT EXISTS beeline_payments (
	id TEXT PRIMARY KEY,
	receiver_card TEXT NOT NULL,
	amount DOUBLE PRECISION NOT NULL,
	commission DOUBLE PRECISION NOT NULL,
	total DOUBLE PRECISION NOT NULL,
	source TEXT NOT NULL CHECK (source IN ('manual', 'payment_flow')),
	paid_at TIMESTAMPTZ NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS beeline_payments_paid_at_idx ON beeline_payments (paid_at DESC);
