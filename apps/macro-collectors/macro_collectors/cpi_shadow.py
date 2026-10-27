"""Cross-source review for the new national CPI report dataset.

A numerical match against a different article/series is NOT same-payload parser
parity and can never qualify this dataset for automatic migration review.
"""
import math

from . import cpi_reports
from .http import CollectionError


def normalize(client,dataset,config,today):
    body,metrics=cpi_reports.collect(client,dataset["url"],today)
    return body,bind_metrics(metrics,config)


def bind_metrics(metrics,config):
    result=[]
    for binding in config:
        if binding["extractor_key"]!=cpi_reports.VERSION or binding["transform_key"]!="published_percent":
            raise CollectionError("DATA_QUALITY","Unsupported CPI binding")
        basis=binding["validation"].get("basis")
        if basis not in {"MoM","YoY","AVG_YOY"}:raise CollectionError("DATA_QUALITY","Unknown CPI comparison basis")
        for row in metrics:
            if row["basis"]!=basis:continue
            result.append({**row,"binding_id":binding["id"],"indicator_code":binding["indicator_code"],"series_id":binding["series_id"],"unit":binding["unit"],"priority":binding["priority"],"method":"PUBLISHED_NATIONAL_CPI_PERCENT; BASIS="+basis+"; PARSER="+cpi_reports.VERSION})
    return result


def compare(db,dataset,rows,config,today):
    required={binding["indicator_code"]:binding for binding in config if binding["validation"].get("baseline_required",True)}
    checks=[]
    # Batch full daily histories instead of one query per observation.
    reference_rows=db.execute("""SELECT DISTINCT ON(indicator_code,period)
      indicator_code,value,unit,source,source_url,series_id,period
      FROM macro_observations WHERE indicator_code=ANY(%s) AND period=ANY(%s)
      ORDER BY indicator_code,period,priority DESC,revision DESC,collected_at DESC,id DESC""",
      (sorted(required),sorted({row["period"] for row in rows}))).fetchall() if required else []
    references={(ref["indicator_code"],ref["period"]):ref for ref in reference_rows}
    for row in rows:
        code=row["indicator_code"]
        if code not in required:
            checks.append({"indicator":code,"period":row["period"],"status":"NEW_METRIC_NO_BASELINE"});continue
        reference=references.get((code,row["period"]))
        if not reference:
            checks.append({"indicator":code,"period":row["period"],"status":"NEW_SOURCE_PERIOD"});continue
        tolerance=required[code]["validation"]["tolerance"]
        match=reference["unit"]==row["unit"] and math.isfinite(reference["value"]) and abs(row["value"]-reference["value"])<=tolerance
        checks.append({"indicator":code,"period":row["period"],"status":"CROSS_SOURCE_VALUE_MATCH" if match else "VALUE_MISMATCH","go":reference["value"],"python":row["value"],"goSource":reference["source_url"]})
    latest=max(row["period_end"] for row in rows)
    age_limits={binding["indicator_code"]:binding["validation"].get("max_age_days",60) for binding in config}
    latest_by_code={}
    for row in rows:
        code=row["indicator_code"]
        latest_by_code[code]=max(latest_by_code.get(code,row["period_end"]),row["period_end"])
    stale=any((today-end).days>age_limits[code] for code,end in latest_by_code.items())
    status="STALE" if stale else "CROSS_SOURCE_MISMATCH" if any(c["status"]=="VALUE_MISMATCH" for c in checks) else "CROSS_SOURCE_REVIEW"
    candidates=[{key:(value.isoformat() if hasattr(value,"isoformat") else value) for key,value in row.items() if key in ("indicator_code","period","value","unit","basis","source_url","published_at","evidence")} for row in rows]
    missing=[binding["indicator_code"] for binding in config if binding["indicator_code"] not in {row["indicator_code"] for row in rows}]
    return status,{"missingMetrics":missing,"baselinePolicy":"cross_source","cutoverEligible":False,"checks":checks,"candidateObservations":candidates,"latest":latest.isoformat(),"stale":stale}
