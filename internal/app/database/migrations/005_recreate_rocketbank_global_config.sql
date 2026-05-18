DROP TABLE IF EXISTS rocketbank_configs;

CREATE TABLE rocketbank_configs (
	id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
	balance DOUBLE PRECISION,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO rocketbank_configs (id, balance)
VALUES (1, NULL);
