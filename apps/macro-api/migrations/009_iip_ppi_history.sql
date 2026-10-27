INSERT INTO macro_indicator_definitions(code,unit,frequency,comparison_basis) VALUES
 ('VN_PPI_YOY','% YoY','QUARTERLY','YoY') ON CONFLICT DO NOTHING;
INSERT INTO macro_datasets(id,source_id,url,go_provider_id,interval_seconds) VALUES
 ('nso_ppi_history','nso','https://nsdp.nso.gov.vn/GSO-chung/SDMXFiles/GSO/PPIVNM.xml','nso:sdmx:VNM_PPPI_ISIC4_BTE_BY_PP_IX',21600) ON CONFLICT DO NOTHING;
INSERT INTO macro_indicator_bindings(id,dataset_id,indicator_code,series_id,extractor_key,transform_key,priority,validation) VALUES
 ('nso_ppi_yoy_history','nso_ppi_history','VN_PPI_YOY','VNM_PPPI_ISIC4_BTE_BY_PP_IX','nso_ppi_history_v1','same_quarter_index_minus_100',100,'{"tolerance":0.02,"max_age_days":120,"baseline_required":true}') ON CONFLICT DO NOTHING;
