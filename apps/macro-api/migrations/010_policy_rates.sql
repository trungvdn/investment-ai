-- Existing production observation/raw/schedule tables; no new data tables.
INSERT INTO macro_indicator_definitions(code,unit,frequency,comparison_basis) VALUES
 ('VN_SBV_REFINANCING','%','EVENT','LEVEL'),
 ('VN_SBV_DISCOUNT','%','EVENT','LEVEL'),
 ('VN_SBV_OVERNIGHT','%','EVENT','LEVEL') ON CONFLICT DO NOTHING;
INSERT INTO macro_datasets(id,source_id,url,go_provider_id,interval_seconds,ingestion_owner) VALUES
 ('sbv_policy_rates','sbv','https://sbv.gov.vn/vi/lãi-suất1','sbv:policy',86400,'python') ON CONFLICT DO NOTHING;
INSERT INTO macro_indicator_bindings(id,dataset_id,indicator_code,series_id,extractor_key,transform_key,priority,validation) VALUES
 ('sbv_refinancing','sbv_policy_rates','VN_SBV_REFINANCING','SBV_POLICY_REFINANCING','sbv_policy_v1','effective_date_rate',100,'{"max_age_days":7}'),
 ('sbv_discount','sbv_policy_rates','VN_SBV_DISCOUNT','SBV_POLICY_DISCOUNT','sbv_policy_v1','effective_date_rate',100,'{"max_age_days":7}'),
 ('sbv_policy_overnight','sbv_policy_rates','VN_SBV_OVERNIGHT','SBV_POLICY_OVERNIGHT','sbv_policy_v1','effective_date_rate',100,'{"max_age_days":7}') ON CONFLICT DO NOTHING;
