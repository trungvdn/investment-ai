"""Full available FRED daily history, bounded and with credential-free provenance."""
import json
import math
import os
from datetime import date
from urllib.parse import urlencode
from .http import CollectionError

SERIES={"fred_effr_history":("US_EFFR","DFF"),"fred_10y_history":("US_10Y","DGS10")}


def collect(client,dataset,config,today,page=1,first=None,last=None):
    code,series=SERIES[dataset["id"]]
    public="https://fred.stlouisfed.org/series/"+series
    if dataset["url"]!=public or len(config)!=1 or config[0]["indicator_code"]!=code or config[0]["unit"]!="%":
        raise CollectionError("DATA_QUALITY","Unexpected FRED history binding")
    key=os.getenv("FRED_API_KEY")
    if not key:raise CollectionError("CONFIGURATION","FRED_API_KEY is required")
    if page!=1:raise CollectionError("CONFIGURATION","FRED history is one bounded complete response")
    url="https://api.stlouisfed.org/fred/series/observations?"+urlencode({"api_key":key,"series_id":series,
        "file_type":"json","observation_start":"1900-01-01","observation_end":today.isoformat(),"sort_order":"asc","limit":100000})
    body=client.get(url)
    try:
        data=json.loads(body)
        if data.get("offset",0)!=0 or data["count"]>len(data["observations"]):raise ValueError("Truncated FRED history")
        rows=[];seen=set();binding=config[0]
        for item in data["observations"]:
            if item["value"] in (".",""):continue
            end=date.fromisoformat(item["date"]);value=float(item["value"])
            if end>today or not math.isfinite(value) or item["date"] in seen:raise ValueError("Invalid FRED history")
            seen.add(item["date"])
            if first and not first<=end.strftime("%Y-%m")<=last:continue
            rows.append({"binding_id":binding["id"],"indicator_code":code,"period":end.isoformat(),"period_end":end,
                         "value":value,"unit":"%","source_url":public,"series_id":series,"priority":100,
                         "published_at":None,"method":"SOURCE_LEVEL","evidence":{"series":series}})
        return body,rows,{"history":True,"warnings":[],"sourceUrl":public,"upstreamCount":data["count"]}
    except (ValueError,KeyError,TypeError,OverflowError):
        error=CollectionError("DATA_QUALITY","Invalid or truncated FRED history");error.payload=body;raise error from None
