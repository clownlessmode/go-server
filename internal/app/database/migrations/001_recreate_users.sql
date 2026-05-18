DROP TABLE IF EXISTS users CASCADE;
DROP TYPE IF EXISTS user_role CASCADE;

CREATE TYPE user_role AS ENUM ('user', 'admin');

CREATE TABLE users (
	id BIGSERIAL PRIMARY KEY,
	login TEXT NOT NULL,
	password TEXT NOT NULL,
	role user_role NOT NULL,
	is_active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX users_role_idx ON users (role);
CREATE INDEX users_is_active_idx ON users (is_active);
