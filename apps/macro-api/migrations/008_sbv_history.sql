INSERT INTO macro_sources(id,name,homepage) VALUES ('sbv','State Bank of Vietnam','https://sbv.gov.vn/'),('fred','FRED','https://fred.stlouisfed.org/') ON CONFLICT DO NOTHING;
INSERT INTO macro_indicator_definitions(code,unit,frequency,comparison_basis) VALUES
 ('VN_INTERBANK_ON','%','DAILY','LEVEL'),('VN_CREDIT_GROWTH','% YTD','MONTHLY','YTD'),
 ('US_EFFR','%','DAILY','LEVEL'),('US_10Y','%','DAILY','LEVEL') ON CONFLICT DO NOTHING;
INSERT INTO macro_datasets(id,source_id,url,go_provider_id,interval_seconds) VALUES
 ('sbv_credit_history','sbv','https://sbv.gov.vn/o/article/v1.0/articles','sbv:statistics:credit',21600),
 ('sbv_interbank_history','sbv','https://sbv.gov.vn/vi/lãi-suất1','sbv:interbank',21600),
 ('fred_effr_history','fred','https://fred.stlouisfed.org/series/DFF','fred:DFF',21600),
 ('fred_10y_history','fred','https://fred.stlouisfed.org/series/DGS10','fred:DGS10',21600) ON CONFLICT DO NOTHING;
INSERT INTO macro_indicator_bindings(id,dataset_id,indicator_code,series_id,extractor_key,transform_key,priority,validation) VALUES
 ('sbv_credit_ytd_history','sbv_credit_history','VN_CREDIT_GROWTH','SBV_TOTAL_CREDIT_YTD','sbv_credit_history_v1','published_ytd',100,'{"tolerance":0.02,"max_age_days":75,"baseline_required":true}'),
 ('sbv_on_history','sbv_interbank_history','VN_INTERBANK_ON','SBV_VND_INTERBANK_OVERNIGHT','sbv_interbank_history_v1','published_rate',100,'{"tolerance":0.001,"max_age_days":7,"baseline_required":true}'),
 ('fred_effr_full','fred_effr_history','US_EFFR','DFF','fred_history_v1','published_rate',100,'{"tolerance":0.000001,"max_age_days":7,"baseline_required":true,"calendar":"ALL_DAYS"}'),
 ('fred_10y_full','fred_10y_history','US_10Y','DGS10','fred_history_v1','published_rate',100,'{"tolerance":0.000001,"max_age_days":7,"baseline_required":true,"calendar":"WEEKDAYS_UNVERIFIED_HOLIDAYS"}') ON CONFLICT DO NOTHING;
