ALTER TABLE macro_datasets ADD COLUMN IF NOT EXISTS baseline_policy TEXT NOT NULL DEFAULT 'same_payload' CHECK(baseline_policy IN ('same_payload','cross_source'));
ALTER TABLE macro_pilot_observations ADD COLUMN IF NOT EXISTS source_url TEXT;
ALTER TABLE macro_pilot_observations ADD COLUMN IF NOT EXISTS published_at TIMESTAMPTZ;
ALTER TABLE macro_pilot_observations ADD COLUMN IF NOT EXISTS evidence JSONB NOT NULL DEFAULT '{}';
INSERT INTO macro_indicator_definitions(code,unit,frequency,comparison_basis) VALUES
 ('VN_CPI_MOM','% MoM','MONTHLY','MoM'),('VN_CPI_AVG_YOY','% average YoY','MONTHLY','AVG_YOY') ON CONFLICT DO NOTHING;
INSERT INTO macro_datasets(id,source_id,url,go_provider_id,interval_seconds,baseline_policy) VALUES
 ('gso_cpi_reports','nso','https://www.gso.gov.vn/cpi-vi/','nso:reports:economic',21600,'cross_source') ON CONFLICT DO NOTHING;
INSERT INTO macro_indicator_bindings(id,dataset_id,indicator_code,series_id,extractor_key,transform_key,priority,validation) VALUES
 ('gso_cpi_mom','gso_cpi_reports','VN_CPI_MOM','GSO_CPI_MONTHLY_MOM','gso_cpi_article_v1','published_percent',110,'{"basis":"MoM","tolerance":0.02,"max_age_days":60,"baseline_required":false}'),
 ('gso_cpi_yoy','gso_cpi_reports','VN_CPI_YOY','GSO_CPI_MONTHLY_YOY','gso_cpi_article_v1','published_percent',110,'{"basis":"YoY","tolerance":0.02,"max_age_days":60,"baseline_required":true}'),
 ('gso_cpi_avg_yoy','gso_cpi_reports','VN_CPI_AVG_YOY','GSO_CPI_YTD_AVERAGE_YOY','gso_cpi_article_v1','published_percent',110,'{"basis":"AVG_YOY","tolerance":0.02,"max_age_days":60,"baseline_required":false}') ON CONFLICT DO NOTHING;
