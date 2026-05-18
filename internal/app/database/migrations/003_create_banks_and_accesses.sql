CREATE TABLE IF NOT EXISTS bank_catalog (
	id BIGINT PRIMARY KEY,
	code TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO bank_catalog (id, code, name)
VALUES (1, 'rocketbank', 'Rocketbank')
ON CONFLICT (id) DO UPDATE
SET code = EXCLUDED.code,
	name = EXCLUDED.name,
	updated_at = NOW();

CREATE TABLE IF NOT EXISTS access_grants (
	id BIGSERIAL PRIMARY KEY,
	user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	bank_id BIGINT NOT NULL REFERENCES bank_catalog (id) ON DELETE CASCADE,
	granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	expires_at TIMESTAMPTZ NOT NULL,
	grant_reason TEXT NOT NULL,
	revoked_at TIMESTAMPTZ,
	revoke_reason TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS access_grants_user_id_idx ON access_grants (user_id);
CREATE INDEX IF NOT EXISTS access_grants_bank_id_idx ON access_grants (bank_id);
CREATE INDEX IF NOT EXISTS access_grants_expires_at_idx ON access_grants (expires_at);
CREATE INDEX IF NOT EXISTS access_grants_user_bank_idx ON access_grants (user_id, bank_id);
