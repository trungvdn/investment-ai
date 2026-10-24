CREATE TABLE IF NOT EXISTS macro_source_state (
 provider TEXT PRIMARY KEY,
 next_due TIMESTAMPTZ NOT NULL DEFAULT now(),
 lease_until TIMESTAMPTZ,
 lease_token TEXT,
 last_checked TIMESTAMPTZ,
 last_success TIMESTAMPTZ,
 consecutive_failures INTEGER NOT NULL DEFAULT 0 CHECK(consecutive_failures >= 0),
 error_kind TEXT NOT NULL DEFAULT '',
 message TEXT NOT NULL DEFAULT ''
);
