INSERT INTO macro_sources(id,name,homepage) VALUES('mbs','MBS Research (secondary publisher)','https://www.mbs.com.vn/') ON CONFLICT DO NOTHING;
INSERT INTO macro_datasets(id,source_id,url,go_provider_id,interval_seconds,baseline_policy) VALUES
 ('mbs_pmi_history','mbs','https://www.mbs.com.vn/trien-vong-vi-mo/','mbs:pmi:history',21600,'cross_source') ON CONFLICT DO NOTHING;
INSERT INTO macro_indicator_bindings(id,dataset_id,indicator_code,series_id,extractor_key,transform_key,priority,validation) VALUES
 ('mbs_pmi_headline','mbs_pmi_history','VN_PMI','MBS_VN_MANUFACTURING_PMI','mbs_pmi_history_v1','published_index',80,'{"tolerance":0.05,"max_age_days":60,"baseline_required":true}') ON CONFLICT DO NOTHING;
