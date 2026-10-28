-- Derived model output only: indicator history remains in macro_observations.
CREATE TABLE IF NOT EXISTS macro_score_snapshots (
 id BIGSERIAL PRIMARY KEY,
 model_version TEXT NOT NULL,
 model_hash TEXT NOT NULL,
 kind TEXT NOT NULL CHECK(kind IN ('EVENT','DAILY','WEEKLY')),
 bucket TEXT NOT NULL,
 as_of TIMESTAMPTZ NOT NULL,
 input_hash TEXT NOT NULL,
 payload JSONB NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 UNIQUE(model_hash,kind,bucket)
);
CREATE INDEX IF NOT EXISTS macro_score_snapshots_recent ON macro_score_snapshots(model_hash,as_of DESC,id DESC);
