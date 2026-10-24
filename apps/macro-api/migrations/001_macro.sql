CREATE TABLE IF NOT EXISTS macro_raw_documents (
 id BIGSERIAL PRIMARY KEY,
 source TEXT NOT NULL,
 source_url TEXT NOT NULL,
 content_type TEXT NOT NULL,
 checksum TEXT NOT NULL,
 payload BYTEA NOT NULL,
 first_fetched_at TIMESTAMPTZ NOT NULL,
 last_fetched_at TIMESTAMPTZ NOT NULL,
 UNIQUE(source,source_url,checksum)
);
CREATE TABLE IF NOT EXISTS macro_observations (
 id BIGSERIAL PRIMARY KEY,
 indicator_code TEXT NOT NULL,
 period TEXT NOT NULL,
 period_end DATE NOT NULL,
 value DOUBLE PRECISION NOT NULL,
 unit TEXT NOT NULL,
 source TEXT NOT NULL,
 source_url TEXT NOT NULL,
 series_id TEXT NOT NULL,
 method TEXT NOT NULL,
 priority INTEGER NOT NULL,
 published_at TIMESTAMPTZ,
 first_seen_at TIMESTAMPTZ NOT NULL,
 collected_at TIMESTAMPTZ NOT NULL,
 revision INTEGER NOT NULL CHECK(revision>0),
 raw_id BIGINT NOT NULL REFERENCES macro_raw_documents(id),
 UNIQUE(indicator_code,period,series_id,revision)
);
CREATE INDEX IF NOT EXISTS macro_observations_lookup ON macro_observations(indicator_code,period_end DESC,priority DESC,revision DESC);
CREATE TABLE IF NOT EXISTS macro_attempts (
 indicator_code TEXT NOT NULL,
 provider_id TEXT NOT NULL,
 checked_at TIMESTAMPTZ NOT NULL,
 success BOOLEAN NOT NULL,
 message TEXT NOT NULL,
 PRIMARY KEY(indicator_code,provider_id)
);
