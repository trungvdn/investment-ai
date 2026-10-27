-- Stable source vintages, independent from per-run collection audit.
CREATE TABLE IF NOT EXISTS macro_source_values (
 id BIGSERIAL PRIMARY KEY,
 binding_id TEXT NOT NULL REFERENCES macro_indicator_bindings(id),
 period TEXT NOT NULL, period_end DATE NOT NULL,
 revision INTEGER NOT NULL CHECK(revision>0), fingerprint TEXT NOT NULL,
 value DOUBLE PRECISION NOT NULL, unit TEXT NOT NULL, method TEXT NOT NULL,
 source_url TEXT NOT NULL, published_at TIMESTAMPTZ, evidence JSONB NOT NULL,
 raw_id BIGINT NOT NULL REFERENCES macro_pilot_raw(id),
 first_run_id BIGINT NOT NULL REFERENCES macro_pilot_runs(id),
 first_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(), last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 UNIQUE(binding_id,period,revision)
);
CREATE INDEX IF NOT EXISTS macro_source_value_lookup ON macro_source_values(binding_id,period,revision DESC);
CREATE TABLE IF NOT EXISTS macro_source_policies (
 indicator_code TEXT PRIMARY KEY REFERENCES macro_indicator_definitions(code),
 history_binding TEXT NOT NULL REFERENCES macro_indicator_bindings(id),
 update_binding TEXT NOT NULL REFERENCES macro_indicator_bindings(id),
 tolerance DOUBLE PRECISION NOT NULL CHECK(tolerance>0 AND tolerance<=0.1),
 min_overlap INTEGER NOT NULL CHECK(min_overlap>=3),
 active_batch_id BIGINT,
 CHECK(history_binding<>update_binding)
);
CREATE TABLE IF NOT EXISTS macro_canonical_batches (
 id BIGSERIAL PRIMARY KEY, indicator_code TEXT NOT NULL REFERENCES macro_source_policies(indicator_code),
 fingerprint TEXT NOT NULL, report JSONB NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 UNIQUE(indicator_code,fingerprint), UNIQUE(indicator_code,id)
);
CREATE TABLE IF NOT EXISTS macro_canonical_links (
 batch_id BIGINT NOT NULL REFERENCES macro_canonical_batches(id), period TEXT NOT NULL,
 source_value_id BIGINT NOT NULL REFERENCES macro_source_values(id),
 observation_id BIGINT NOT NULL REFERENCES macro_observations(id),
 PRIMARY KEY(batch_id,period)
);
CREATE TABLE IF NOT EXISTS macro_selection_events (
 id BIGSERIAL PRIMARY KEY, indicator_code TEXT NOT NULL REFERENCES macro_source_policies(indicator_code),
 previous_batch_id BIGINT REFERENCES macro_canonical_batches(id),
 next_batch_id BIGINT REFERENCES macro_canonical_batches(id),
 occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(), reason TEXT NOT NULL
);
DO $$ BEGIN
 IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conrelid='macro_source_policies'::regclass AND conname='macro_source_policy_active_batch') THEN
  ALTER TABLE macro_source_policies ADD CONSTRAINT macro_source_policy_active_batch
   FOREIGN KEY(indicator_code,active_batch_id) REFERENCES macro_canonical_batches(indicator_code,id);
 END IF;
END $$;
INSERT INTO macro_source_policies(indicator_code,history_binding,update_binding,tolerance,min_overlap)
 VALUES('VN_CPI_YOY','nso_cpi_yoy','gso_cpi_yoy',0.02,3) ON CONFLICT DO NOTHING;
