"""One production write path for historical and scheduled observations."""
import calendar
import hashlib
import json
import math
import re
import uuid
from datetime import date, datetime, timezone

from .http import CollectionError

SOURCES={"nso":"NSO","sbv":"SBV","fred":"FRED","mbs":"MBS Research","tradingeconomics":"Trading Economics"}


def claim(db,provider,force=False):
    token=uuid.uuid4().hex
    with db.transaction():
        db.execute("INSERT INTO macro_source_state(provider) VALUES(%s) ON CONFLICT DO NOTHING",(provider,))
        if force:
            db.execute("UPDATE macro_source_state SET next_due=now() WHERE provider=%s AND (lease_until IS NULL OR lease_until<=now())",(provider,))
        row=db.execute("""UPDATE macro_source_state SET lease_token=%s,lease_until=now()+interval '10 minutes'
          WHERE provider=%s AND next_due<=now() AND (lease_until IS NULL OR lease_until<=now()) RETURNING lease_token""",
          (token,provider)).fetchone()
    return token if row else None


def validated(rows,config,today):
    bindings={b["id"]:b for b in config}
    result={}
    try:
        for row in rows:
            b=bindings[row["binding_id"]]
            if any(row[field]!=b[field] for field in ("indicator_code","series_id","unit")):
                raise ValueError("Binding mismatch")
            if not math.isfinite(row["value"]) or isinstance(row["value"],bool) or not row.get("method") or not row.get("source_url","").startswith("https://"):
                raise ValueError("Invalid value or provenance")
            period=row["period"]
            if b["frequency"]=="MONTHLY":
                if not re.fullmatch(r"\d{4}-\d{2}",period):raise ValueError("Invalid month")
                year,month=map(int,period.split("-"));end=date(year,month,calendar.monthrange(year,month)[1])
            elif b["frequency"]=="QUARTERLY":
                if not re.fullmatch(r"\d{4}-Q[1-4]",period):raise ValueError("Invalid quarter")
                year,month=int(period[:4]),int(period[-1])*3
                end=date(year,month,calendar.monthrange(year,month)[1])
            elif b["frequency"]=="DAILY":
                end=date.fromisoformat(period)
            else:raise ValueError("Unsupported frequency")
            if row["period_end"]!=end or end>today:raise ValueError("Invalid reference date")
            if row.get("published_at") and (row["published_at"].tzinfo is None or row["published_at"].date()>today):
                raise ValueError("Invalid publication date")
            key=(row["indicator_code"],period,row["series_id"])
            if key in result and any(result[key][field]!=row[field] for field in ("value","unit","method","period_end")):
                raise ValueError("Conflicting period within job")
            result.setdefault(key,row)
        return [result[key] for key in sorted(result)]
    except (KeyError,ValueError,TypeError,OverflowError):
        raise CollectionError("DATA_QUALITY","Observation does not satisfy its configured production definition") from None


def content_type(body):
    if body.lstrip().startswith(b"<"):return "application/xml" if b"<Series" in body else "text/html"
    if body.startswith(b"%PDF-"):return "application/pdf"
    return "application/json"


def persist(db,dataset,token,body,rows,config,failure=None,notes=None):
    notes=notes or {}
    provider=dataset["go_provider_id"]
    now=datetime.now(timezone.utc)
    inserted=unchanged=revised=0
    raw_id=None
    with db.transaction():
        # Default READ COMMITTED is deliberate. Each locked value comparison must
        # see revisions committed by concurrent Go/Python writers.
        state=db.execute("""SELECT *,lease_until>clock_timestamp() AS active FROM macro_source_state
          WHERE provider=%s FOR UPDATE""",(provider,)).fetchone()
        if not state or state["lease_token"]!=token or not state["active"]:
            raise RuntimeError("Production source lease expired or replaced")
        if body is not None:
            raw_id=db.execute("""INSERT INTO macro_raw_documents(source,source_url,content_type,checksum,payload,first_fetched_at,last_fetched_at)
              VALUES(%s,%s,%s,%s,%s,%s,%s) ON CONFLICT(source,source_url,checksum) DO UPDATE
              SET last_fetched_at=GREATEST(macro_raw_documents.last_fetched_at,EXCLUDED.last_fetched_at) RETURNING id""",
              (SOURCES.get(dataset["source_id"],dataset["source_id"]),dataset["url"],content_type(body),
               hashlib.sha256(body).hexdigest(),body,now,now)).fetchone()["id"]
        if not failure:
            if raw_id is None:raise RuntimeError("Production values require raw provenance")
            for row in rows:
                saved=db.execute("""SELECT * FROM macro_save_observation_v1(%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)""",
                  (row["indicator_code"],row["period"],row["period_end"],row["value"],row["unit"],
                   SOURCES.get(dataset["source_id"],dataset["source_id"]),row["source_url"],row["series_id"],
                   row["method"],row["priority"],row.get("published_at"),now,raw_id)).fetchone()
                if saved["inserted"]:
                    inserted+=1
                    revised+=int(saved["observation_revision"]>1)
                else:unchanged+=1
        warnings=notes.get("warnings",[])
        status="SOURCE_ERROR" if failure else "PARTIAL" if warnings else "WRITTEN"
        success=not failure and not warnings
        summary={"status":status,"mode":"production","inserted":inserted,"revised":revised,
                 "unchanged":unchanged,"rawId":raw_id,"warnings":warnings[:20],"page":notes.get("page")}
        if failure:summary.update(errorKind=failure.kind,message=str(failure))
        for b in config:
            db.execute("""INSERT INTO macro_attempts(indicator_code,provider_id,checked_at,success,message)
              VALUES(%s,%s,%s,%s,%s) ON CONFLICT(indicator_code,provider_id) DO UPDATE
              SET checked_at=EXCLUDED.checked_at,success=EXCLUDED.success,message=EXCLUDED.message""",
              (b["indicator_code"],provider,now,success,json.dumps(summary,default=str)))
        failures=state["consecutive_failures"]+1 if not success else 0
        delay=dataset["interval_seconds"] if success else max(900,min(21600,900*2**min(failures-1,5)))
        if failure:delay=max(delay,min(failure.retry_after,365*86400))
        kind=failure.kind if failure else "PARTIAL" if warnings else ""
        updated=db.execute("""UPDATE macro_source_state SET next_due=now()+(%s*interval '1 second'),
          last_checked=%s,last_success=CASE WHEN %s THEN %s ELSE last_success END,
          consecutive_failures=%s,error_kind=%s,message=%s,lease_token=NULL,lease_until=NULL
          WHERE provider=%s AND lease_token=%s AND lease_until>clock_timestamp()""",
          (delay,now,success,now,failures,kind,json.dumps(summary,default=str),provider,token))
        if updated.rowcount!=1:raise RuntimeError("Production source lease expired or replaced")
    return summary


def report(db,datasets):
    return db.execute("""SELECT d.id AS dataset,d.go_provider_id AS provider,s.next_due,s.last_checked,
      s.last_success,s.consecutive_failures,s.error_kind,s.message,s.lease_until
      FROM macro_datasets d LEFT JOIN macro_source_state s ON s.provider=d.go_provider_id
      WHERE d.id=ANY(%s) ORDER BY d.id""",(datasets,)).fetchall()


def expected_periods(frequency,first,last,all_days=False):
    from .selection import periods_between
    from datetime import timedelta
    months=periods_between(first,last)
    if frequency=="MONTHLY":return months
    if frequency=="QUARTERLY":
        return [m[:4]+"-Q"+str(int(m[5:])//3) for m in months if int(m[5:])%3==0]
    if frequency=="DAILY":
        start=date.fromisoformat(first+"-01")
        year,month=map(int,last.split("-"))
        end=min(date(year,month,calendar.monthrange(year,month)[1]),datetime.now(timezone.utc).date())
        return [(start+timedelta(days=i)).isoformat() for i in range(max(0,(end-start).days+1))
                if all_days or (start+timedelta(days=i)).weekday()<5]
    return []


def coverage(db,datasets,first=None,last=None):
    definitions=db.execute("""SELECT b.*,i.unit,i.frequency,i.comparison_basis,d.go_provider_id
      FROM macro_indicator_bindings b JOIN macro_indicator_definitions i ON i.code=b.indicator_code
      JOIN macro_datasets d ON d.id=b.dataset_id WHERE b.dataset_id=ANY(%s) AND b.enabled ORDER BY b.id""",(datasets,)).fetchall()
    result=[]
    for b in definitions:
        rows=db.execute("""SELECT DISTINCT ON(period) period,period_end,revision FROM macro_observations
          WHERE indicator_code=%s AND series_id=%s ORDER BY period,revision DESC""",(b["indicator_code"],b["series_id"])).fetchall()
        periods=[r["period"] for r in rows]
        all_days=b["validation"].get("calendar")=="ALL_DAYS"
        left=rows[0]["period_end"].strftime("%Y-%m") if rows else None
        right=rows[-1]["period_end"].strftime("%Y-%m") if rows else None
        internal=[p for p in expected_periods(b["frequency"],left,right,all_days) if periods[0]<=p<=periods[-1] and p not in periods] if periods else []
        missing=sorted(set(expected_periods(b["frequency"],first,last,all_days))-set(periods)) if first else []
        result.append({"indicator":b["indicator_code"],"dataset":b["dataset_id"],"series":b["series_id"],
          "table":"macro_observations","first":periods[0] if periods else None,"last":periods[-1] if periods else None,
          "periods":len(periods),"revisions":sum(r["revision"] for r in rows),"missingPeriods":internal,"missingRequestedPeriods":missing,
          "dailyCalendarNote":"Unobserved weekdays may include holidays; no fill" if b["frequency"]=="DAILY" and not all_days else None})
    return result
