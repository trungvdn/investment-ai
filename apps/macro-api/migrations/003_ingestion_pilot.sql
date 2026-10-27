CREATE TABLE IF NOT EXISTS macro_indicator_definitions (
 code TEXT PRIMARY KEY, unit TEXT NOT NULL, frequency TEXT NOT NULL, comparison_basis TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS macro_sources (
 id TEXT PRIMARY KEY, name TEXT NOT NULL, homepage TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS macro_datasets (
 id TEXT PRIMARY KEY, source_id TEXT NOT NULL REFERENCES macro_sources(id), url TEXT NOT NULL,
 go_provider_id TEXT NOT NULL UNIQUE, interval_seconds INTEGER NOT NULL CHECK(interval_seconds>=300),
 enabled BOOLEAN NOT NULL DEFAULT true, ingestion_owner TEXT NOT NULL DEFAULT 'go' CHECK(ingestion_owner IN ('go','python'))
);
CREATE TABLE IF NOT EXISTS macro_indicator_bindings (
 id TEXT PRIMARY KEY, dataset_id TEXT NOT NULL REFERENCES macro_datasets(id), indicator_code TEXT NOT NULL REFERENCES macro_indicator_definitions(code),
 series_id TEXT NOT NULL, extractor_key TEXT NOT NULL, transform_key TEXT NOT NULL, priority INTEGER NOT NULL,
 enabled BOOLEAN NOT NULL DEFAULT true, validation JSONB NOT NULL,
 UNIQUE(dataset_id,indicator_code,series_id)
);
CREATE TABLE IF NOT EXISTS macro_pilot_raw (
 id BIGSERIAL PRIMARY KEY, dataset_id TEXT NOT NULL REFERENCES macro_datasets(id), checksum TEXT NOT NULL,
 payload BYTEA NOT NULL, fetched_at TIMESTAMPTZ NOT NULL, UNIQUE(dataset_id,checksum)
);
CREATE TABLE IF NOT EXISTS macro_pilot_runs (
 id BIGSERIAL PRIMARY KEY, dataset_id TEXT NOT NULL REFERENCES macro_datasets(id), raw_id BIGINT REFERENCES macro_pilot_raw(id),
 checked_at TIMESTAMPTZ NOT NULL DEFAULT now(), parser_version TEXT NOT NULL, status TEXT NOT NULL,
 baseline_raw_id BIGINT REFERENCES macro_raw_documents(id), report JSONB NOT NULL
);
CREATE INDEX IF NOT EXISTS macro_pilot_run_dataset_date ON macro_pilot_runs(dataset_id,checked_at DESC);
CREATE TABLE IF NOT EXISTS macro_pilot_observations (
 run_id BIGINT NOT NULL REFERENCES macro_pilot_runs(id), binding_id TEXT NOT NULL REFERENCES macro_indicator_bindings(id),
 period TEXT NOT NULL, period_end DATE NOT NULL, value DOUBLE PRECISION NOT NULL, unit TEXT NOT NULL, series_id TEXT NOT NULL,
 method TEXT NOT NULL, PRIMARY KEY(run_id,binding_id,period)
);
INSERT INTO macro_sources(id,name,homepage) VALUES('nso','National Statistics Office of Viet Nam','https://nsdp.nso.gov.vn/') ON CONFLICT DO NOTHING;
INSERT INTO macro_indicator_definitions(code,unit,frequency,comparison_basis) VALUES
 ('VN_CPI_YOY','% YoY','MONTHLY','YoY'),('VN_IIP_YOY','% YoY','MONTHLY','YoY') ON CONFLICT DO NOTHING;
INSERT INTO macro_datasets(id,source_id,url,go_provider_id,interval_seconds) VALUES
 ('nso_cpi','nso','https://nsdp.nso.gov.vn/GSO-chung/SDMXFiles/GSO/CPIVNM.xml','nso:sdmx:PCPI_IX',21600),
 ('nso_iip','nso','https://nsdp.nso.gov.vn/GSO-chung/SDMXFiles/GSO/IIPVNM.xml','nso:sdmx:AIP_ISIC4_IX',21600) ON CONFLICT DO NOTHING;
INSERT INTO macro_indicator_bindings(id,dataset_id,indicator_code,series_id,extractor_key,transform_key,priority,validation) VALUES
 ('nso_cpi_yoy','nso_cpi','VN_CPI_YOY','PCPI_IX','nso_sdmx_v1','yoy_12m',50,'{"ref_area":"VN","counterpart":"_Z","frequency":"M","unit_mult":"0","tolerance":1e-8,"max_age_days":60}'),
 ('nso_iip_yoy','nso_iip','VN_IIP_YOY','AIP_ISIC4_IX','nso_sdmx_v1','yoy_12m',50,'{"ref_area":"VN","counterpart":"_Z","frequency":"M","unit_mult":"0","tolerance":1e-8,"max_age_days":60}') ON CONFLICT DO NOTHING;
CREATE TABLE IF NOT EXISTS macro_pilot_state (
 dataset_id TEXT PRIMARY KEY REFERENCES macro_datasets(id), next_due TIMESTAMPTZ NOT NULL DEFAULT now(),
 lease_until TIMESTAMPTZ, lease_token TEXT, last_status TEXT, last_run_id BIGINT REFERENCES macro_pilot_runs(id)
);
