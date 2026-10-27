-- Read-only diagnostics. Consecutive identical vintages can be metadata-only
-- duplicates. A -> B -> A is a real revision sequence and is NOT flagged.
WITH ordered AS (
 SELECT id,binding_id,period,revision,value,unit,period_end,method,
  lag(id) OVER w AS previous_id,lag(value) OVER w AS previous_value,
  lag(unit) OVER w AS previous_unit,lag(period_end) OVER w AS previous_end,
  lag(method) OVER w AS previous_method
 FROM macro_source_values
 WINDOW w AS (PARTITION BY binding_id,period ORDER BY revision)
)
SELECT binding_id,period,previous_id,id,revision,value,unit
FROM ordered WHERE previous_id IS NOT NULL
 AND ROW(value,unit,period_end,method) IS NOT DISTINCT FROM ROW(previous_value,previous_unit,previous_end,previous_method)
ORDER BY binding_id,period,revision;

WITH ordered AS (
 SELECT id,indicator_code,period,series_id,revision,value,unit,period_end,method,
  lag(id) OVER w AS previous_id,lag(value) OVER w AS previous_value,
  lag(unit) OVER w AS previous_unit,lag(period_end) OVER w AS previous_end,
  lag(method) OVER w AS previous_method
 FROM macro_observations
 WINDOW w AS (PARTITION BY indicator_code,period,series_id ORDER BY revision)
)
SELECT indicator_code,period,series_id,previous_id,id,revision,value,unit
FROM ordered WHERE previous_id IS NOT NULL
 AND ROW(value,unit,period_end,method) IS NOT DISTINCT FROM ROW(previous_value,previous_unit,previous_end,previous_method)
ORDER BY indicator_code,period,series_id,revision;

-- These repeated rows are EXPECTED run snapshots, not current-value duplicates.
SELECT binding_id,period,count(*) AS collection_run_snapshots,count(DISTINCT run_id) AS runs
FROM macro_pilot_observations
GROUP BY binding_id,period HAVING count(*)>1
ORDER BY binding_id,period;
