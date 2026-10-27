INSERT INTO macro_sources(id,name,homepage) VALUES('mbs','MBS Research (secondary publisher)','https://www.mbs.com.vn/') ON CONFLICT DO NOTHING;
INSERT INTO macro_datasets(id,source_id,url,go_provider_id,interval_seconds) VALUES
 ('mbs_pmi_history','mbs','https://www.mbs.com.vn/trien-vong-vi-mo/','mbs:pmi:history',21600) ON CONFLICT DO NOTHING;
INSERT INTO macro_indicator_bindings(id,dataset_id,indicator_code,series_id,extractor_key,transform_key,priority,validation) VALUES
 ('mbs_pmi_headline','mbs_pmi_history','VN_PMI','MBS_VN_MANUFACTURING_PMI','mbs_pmi_history_v1','published_index',80,'{"tolerance":0.05,"max_age_days":60,"baseline_required":true}') ON CONFLICT DO NOTHING;

-- Shared production writer: history and latest jobs use the same value key.
-- No history/staging data table is introduced.
CREATE OR REPLACE FUNCTION macro_observation_method_key_v1(p_method TEXT)
RETURNS TEXT LANGUAGE sql IMMUTABLE STRICT AS $$
 SELECT regexp_replace(p_method, '; (ARTICLE_ID|DENOMINATOR_ARTICLE_ID|PARSER)=[^;]*', '', 'g')
$$;

CREATE OR REPLACE FUNCTION macro_save_observation_v1(
 p_code TEXT, p_period TEXT, p_end DATE, p_value DOUBLE PRECISION, p_unit TEXT,
 p_source TEXT, p_url TEXT, p_series TEXT, p_method TEXT, p_priority INTEGER,
 p_published TIMESTAMPTZ, p_collected TIMESTAMPTZ, p_raw BIGINT
) RETURNS TABLE(observation_id BIGINT, observation_revision INTEGER, inserted BOOLEAN)
LANGUAGE plpgsql AS $$
DECLARE
 previous macro_observations%ROWTYPE;
 saved_id BIGINT;
 saved_revision INTEGER;
BEGIN
 IF p_code IS NULL OR p_code='' OR p_period IS NULL OR p_period='' OR p_end IS NULL
    OR p_value IS NULL OR p_value IN ('NaN'::float8,'Infinity'::float8,'-Infinity'::float8)
    OR p_unit IS NULL OR p_unit='' OR p_source IS NULL OR p_source=''
    OR p_url IS NULL OR p_url='' OR p_series IS NULL OR p_series=''
    OR p_method IS NULL OR p_method='' OR p_priority IS NULL OR p_raw IS NULL
    OR p_raw<=0 OR p_collected IS NULL THEN
  RAISE EXCEPTION 'invalid observation/provenance' USING ERRCODE='22023';
 END IF;
 PERFORM pg_advisory_xact_lock(hashtextextended(p_code||'|'||p_period||'|'||p_series,0));
 SELECT o.* INTO previous FROM macro_observations o
  WHERE o.indicator_code=p_code AND o.period=p_period AND o.series_id=p_series
  ORDER BY o.revision DESC LIMIT 1;
 IF FOUND AND previous.value=p_value AND previous.unit=p_unit AND previous.period_end=p_end
    AND macro_observation_method_key_v1(previous.method)=macro_observation_method_key_v1(p_method) THEN
  UPDATE macro_observations o SET collected_at=GREATEST(o.collected_at,p_collected)
   WHERE o.id=previous.id;
  RETURN QUERY SELECT previous.id,previous.revision,false;
  RETURN;
 END IF;
 saved_revision=COALESCE(previous.revision,0)+1;
 INSERT INTO macro_observations(indicator_code,period,period_end,value,unit,source,source_url,
  series_id,method,priority,published_at,first_seen_at,collected_at,revision,raw_id)
 VALUES(p_code,p_period,p_end,p_value,p_unit,p_source,p_url,p_series,p_method,p_priority,
  p_published,p_collected,p_collected,saved_revision,p_raw)
 RETURNING id INTO saved_id;
 RETURN QUERY SELECT saved_id,saved_revision,true;
END $$;
