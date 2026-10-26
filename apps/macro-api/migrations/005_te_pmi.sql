INSERT INTO macro_sources(id,name,homepage) VALUES
 ('tradingeconomics','Trading Economics (secondary publisher)','https://tradingeconomics.com/') ON CONFLICT DO NOTHING;
INSERT INTO macro_indicator_definitions(code,unit,frequency,comparison_basis) VALUES
 ('VN_PMI','index','MONTHLY','INDEX') ON CONFLICT DO NOTHING;
INSERT INTO macro_datasets(id,source_id,url,go_provider_id,interval_seconds,baseline_policy) VALUES
 ('te_vietnam_pmi','tradingeconomics','https://tradingeconomics.com/vietnam/manufacturing-pmi','tradingeconomics:pmi:vn',21600,'cross_source') ON CONFLICT DO NOTHING;
INSERT INTO macro_indicator_bindings(id,dataset_id,indicator_code,series_id,extractor_key,transform_key,priority,validation) VALUES
 ('te_vietnam_pmi_headline','te_vietnam_pmi','VN_PMI','TE_VN_MANUFACTURING_PMI','te_vietnam_pmi_v1','published_index',90,'{"tolerance":0.01,"max_age_days":60,"baseline_required":true}') ON CONFLICT DO NOTHING;
