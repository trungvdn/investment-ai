"""Full available monthly IIP and quarterly industrial PPI SDMX history."""
import calendar
import math
import re
import xml.etree.ElementTree as ET
from datetime import date
from .http import CollectionError
from .sdmx import parse as parse_monthly

BASE="https://nsdp.nso.gov.vn/GSO-chung/SDMXFiles/GSO/"
PPI_SERIES="VNM_PPPI_ISIC4_BTE_BY_PP_IX"


def parse_ppi(body,config,today):
    try:
        if b"<!DOCTYPE" in body.upper() or b"<!ENTITY" in body.upper():raise ValueError("XML entity declarations")
        if len(config)!=1:raise ValueError("Expected one PPI binding")
        binding=config[0]
        if binding["extractor_key"]!="nso_ppi_history_v1" or binding["transform_key"]!="same_quarter_index_minus_100" or binding["indicator_code"]!="VN_PPI_YOY" or binding["unit"]!="% YoY" or binding["frequency"]!="QUARTERLY" or binding["comparison_basis"]!="YoY" or binding["series_id"]!=PPI_SERIES:
            raise ValueError("Unexpected industrial PPI definition")
        root=ET.fromstring(body)
        selected=[node for node in root.iter() if node.tag.split("}")[-1]=="Series" and
                  node.get("INDICATOR")==PPI_SERIES and node.get("REF_AREA")=="VN" and
                  node.get("COUNTERPART_AREA")=="_Z" and node.get("FREQ")=="Q"]
        if len(selected)!=1 or selected[0].get("UNIT_MULT")!="0" or selected[0].get("BASE_PER")!="_Z":
            raise ValueError("Unexpected PPI series/base/multiplier")
        rows=[];seen=set()
        for obs in selected[0]:
            if obs.tag.split("}")[-1]!="Obs":continue
            period=obs.get("TIME_PERIOD","")
            if not re.fullmatch(r"\d{4}-Q[1-4]",period):raise ValueError("Invalid quarter")
            year,quarter=int(period[:4]),int(period[-1]);month=quarter*3
            end=date(year,month,calendar.monthrange(year,month)[1])
            index=float(obs.get("OBS_VALUE",""))
            if period in seen or not math.isfinite(index) or index<0 or end>today:raise ValueError("Invalid PPI index")
            seen.add(period)
            rows.append({"binding_id":binding["id"],"indicator_code":"VN_PPI_YOY","period":period,"period_end":end,
                         "value":index-100,"unit":"% YoY","series_id":PPI_SERIES,"source_url":BASE+"PPIVNM.xml",
                         "priority":100,"published_at":None,
                         "method":"SAME_QUARTER_PREVIOUS_YEAR_100_INDEX_MINUS_100; INDUSTRY_B_TO_E",
                         "evidence":{"sourceIndex":index,"comparison":"same quarter previous year = 100"}})
        if not rows:raise ValueError("No PPI observations")
        return sorted(rows,key=lambda r:r["period"])
    except (ET.ParseError,ValueError,KeyError,TypeError,OverflowError):
        error=CollectionError("DATA_QUALITY","PPI does not satisfy the quarterly same-year comparison contract")
        error.payload=body;raise error from None


def collect(client,dataset,config,today,page=1,first=None,last=None):
    if page!=1:raise CollectionError("CONFIGURATION","Full SDMX history is one bounded document")
    if dataset["id"]=="nso_iip":
        url=BASE+"IIPVNM.xml"
        if dataset["url"]!=url or any(b["indicator_code"]!="VN_IIP_YOY" or b["series_id"]!="AIP_ISIC4_IX" for b in config):
            raise CollectionError("DATA_QUALITY","Unexpected IIP series/URL")
        body=client.get(url);rows=parse_monthly(body,config,today)
    elif dataset["id"]=="nso_ppi_history":
        url=BASE+"PPIVNM.xml"
        if dataset["url"]!=url:raise CollectionError("DATA_QUALITY","Unexpected PPI URL")
        body=client.get(url);rows=parse_ppi(body,config,today)
    else:raise CollectionError("CONFIGURATION","Unknown historical SDMX dataset")
    # Parse full input, including prior-year IIP levels. Requested bounds are
    # assessed by coverage; never discard denominators or claim older data exists.
    return body,rows,{"history":True,"warnings":[],"loadedAllAvailable":True,
                      "availableFrom":rows[0]["period"],"availableTo":rows[-1]["period"],
                      "requestedFrom":first,"requestedTo":last}
