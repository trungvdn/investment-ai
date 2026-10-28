-- User-supplied gap fill, not independently verified official releases.
-- Run against a migrated database with psql -v ON_ERROR_STOP=1.
BEGIN;
SELECT pg_advisory_xact_lock(614074120);
DO $$
DECLARE
 r RECORD;
 raw BIGINT;
 body TEXT;
 provenance TEXT := 'https://github.com/trungvdn/investment-ai/blob/main/apps/macro-api/data/manual/2026-10-27.sql';
BEGIN
 FOR r IN SELECT * FROM (VALUES
  ('VN_CPI_YOY','2026-03',DATE '2026-03-31',4.65::float8,'% YoY'),
  ('VN_CPI_YOY','2026-02',DATE '2026-02-28',2.94::float8,'% YoY'),
  ('VN_CPI_YOY','2026-01',DATE '2026-01-31',2.53::float8,'% YoY'),
  ('VN_CPI_YOY','2025-12',DATE '2025-12-31',3.27::float8,'% YoY'),
  ('VN_CREDIT_GROWTH','2026-09',DATE '2026-09-30',11.59::float8,'% YTD')
 ) AS supplied(code,period,period_end,value,unit) LOOP
  -- Fill missing visible periods only. Preserve any existing selected value.
  IF EXISTS (SELECT 1 FROM macro_observations o
   WHERE o.indicator_code=r.code AND o.period_end=r.period_end AND (
    o.series_id='MANUAL_USER_20261027'
    OR NOT EXISTS(SELECT 1 FROM macro_source_policies p
      WHERE p.indicator_code=r.code AND p.active_batch_id IS NOT NULL)
    OR o.id IN (SELECT l.observation_id FROM macro_canonical_links l
      JOIN macro_source_policies p ON p.active_batch_id=l.batch_id
      WHERE p.indicator_code=r.code))) THEN
   RAISE NOTICE 'Skipped existing % %',r.code,r.period;
   CONTINUE;
  END IF;
  body := jsonb_build_object('origin','user-provided manual entry',
    'independently_verified',false,'indicator',r.code,'period',r.period,
    'value',r.value,'unit',r.unit)::text;
  INSERT INTO macro_raw_documents(source,source_url,content_type,checksum,payload,first_fetched_at,last_fetched_at)
   VALUES('Manual (user supplied; unverified)',provenance,'application/json',
    encode(sha256(convert_to(body,'UTF8')),'hex'),convert_to(body,'UTF8'),now(),now())
   ON CONFLICT(source,source_url,checksum) DO UPDATE
    SET last_fetched_at=EXCLUDED.last_fetched_at RETURNING id INTO raw;
  PERFORM * FROM macro_save_observation_v1(r.code,r.period,r.period_end,r.value,r.unit,
   'Manual (user supplied; unverified)',provenance,'MANUAL_USER_20261027',
   'USER_SUPPLIED_GAP_FILL; NOT_INDEPENDENTLY_VERIFIED',-1000000,NULL,now(),raw);
  RAISE NOTICE 'Inserted % % = % %',r.code,r.period,r.value,r.unit;
 END LOOP;
END $$;
COMMIT;
