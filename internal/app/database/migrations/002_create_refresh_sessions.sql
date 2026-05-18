ALTER TABLE users
	ADD CONSTRAINT users_login_unique UNIQUE (login);

CREATE TABLE refresh_sessions (
	id BIGSERIAL PRIMARY KEY,
	user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	token_hash TEXT NOT NULL UNIQUE,
	expires_at TIMESTAMPTZ NOT NULL,
	revoked_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX refresh_sessions_user_id_idx ON refresh_sessions (user_id);
CREATE INDEX refresh_sessions_expires_at_idx ON refresh_sessions (expires_at);
