-- A reported range is one observation. value is the lower bound, upper_value
-- the upper bound; neither is a midpoint or a new standalone indicator.
ALTER TABLE macro_observations ADD COLUMN IF NOT EXISTS upper_value DOUBLE PRECISION;
INSERT INTO macro_indicator_definitions(code,unit,frequency,comparison_basis) VALUES
 ('VN_DEPOSIT_RATE','%','MONTHLY','RANGE') ON CONFLICT DO NOTHING;
INSERT INTO macro_datasets(id,source_id,url,go_provider_id,interval_seconds,ingestion_owner) VALUES
 ('sbv_deposit_rates','sbv','https://sbv.gov.vn/vi/trang-chu','sbv:deposit',86400,'python') ON CONFLICT DO NOTHING;
INSERT INTO macro_indicator_bindings(id,dataset_id,indicator_code,series_id,extractor_key,transform_key,priority,validation) VALUES
 ('sbv_deposit_6_12m','sbv_deposit_rates','VN_DEPOSIT_RATE','SBV_VND_DEPOSIT_6_12M_RANGE','sbv_deposit_v1','published_range',100,'{"max_age_days":75,"tenor":"6-12 months","currency":"VND"}') ON CONFLICT DO NOTHING;

CREATE OR REPLACE FUNCTION macro_save_range_observation_v1(
 p_code TEXT, p_period TEXT, p_end DATE, p_value DOUBLE PRECISION, p_unit TEXT,
 p_source TEXT, p_url TEXT, p_series TEXT, p_method TEXT, p_priority INTEGER,
 p_published TIMESTAMPTZ, p_collected TIMESTAMPTZ, p_raw BIGINT, p_upper DOUBLE PRECISION
) RETURNS TABLE(observation_id BIGINT, observation_revision INTEGER, inserted BOOLEAN)
LANGUAGE plpgsql AS $$
DECLARE saved RECORD;
BEGIN
 IF p_code IS DISTINCT FROM 'VN_DEPOSIT_RATE' OR p_upper IS NULL OR p_value IS NULL
    OR p_upper IN ('NaN'::float8,'Infinity'::float8,'-Infinity'::float8)
    OR p_value < 0 OR p_upper > 100 OR p_value > p_upper THEN
  RAISE EXCEPTION 'invalid deposit rate range' USING ERRCODE='22023';
 END IF;
 -- Include the upper bound in the shared revision fingerprint, so an upper-only
 -- revision is retained and retries of the same range remain idempotent.
 SELECT * INTO saved FROM macro_save_observation_v1(p_code,p_period,p_end,p_value,p_unit,
  p_source,p_url,p_series,regexp_replace(p_method,'; RANGE_UPPER=[^;]*','','g')||'; RANGE_UPPER='||p_upper::text,
  p_priority,p_published,p_collected,p_raw);
 UPDATE macro_observations SET upper_value=p_upper WHERE id=saved.observation_id;
 RETURN QUERY SELECT saved.observation_id,saved.observation_revision,saved.inserted;
END $$;

-- Retire the previous draft collector if it was tested before this revision.
-- Preserve previously collected observations/raw documents for audit.
UPDATE macro_datasets SET enabled=false WHERE id='sbv_policy_rates';
UPDATE macro_indicator_bindings SET enabled=false WHERE dataset_id='sbv_policy_rates';
