CREATE TABLE IF NOT EXISTS sms_agent_messages (
	id TEXT PRIMARY KEY,
	address TEXT NOT NULL,
	body TEXT NOT NULL,
	bank TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'delivered', 'failed')),
	device_id TEXT,
	error_message TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	delivered_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS sms_agent_messages_status_created_idx
ON sms_agent_messages (status, created_at ASC);
