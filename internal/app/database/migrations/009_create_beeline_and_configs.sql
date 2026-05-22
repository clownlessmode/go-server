INSERT INTO bank_catalog (id, code, name)
VALUES (2, 'beeline', 'Билайн')
ON CONFLICT (id) DO UPDATE
SET code = EXCLUDED.code,
	name = EXCLUDED.name,
	updated_at = NOW();

CREATE TABLE IF NOT EXISTS beeline_configs (
	id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO beeline_configs (id)
VALUES (1)
ON CONFLICT (id) DO NOTHING;
