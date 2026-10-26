"""Direct PostgreSQL shadow ingestion. No production observation write path."""
import argparse
import hashlib
import json
import logging
import os
import signal
import threading
import uuid
from datetime import datetime, timezone
from urllib.parse import urlsplit

from .core import ZONE
from .http import CollectionError, HTTPClient
from .sdmx import VERSION, compare, parse

LOG = logging.getLogger("macro_pilot")
PILOT_IDS = {"nso_cpi", "nso_iip", "gso_cpi_reports"}


def connect():
    import psycopg
    from psycopg.rows import dict_row
    return psycopg.connect(os.getenv("DATABASE_URL", ""), autocommit=True, row_factory=dict_row, connect_timeout=10, options="-c statement_timeout=15000 -c lock_timeout=5000")


def bindings(db, dataset):
    return db.execute("""SELECT b.*,i.unit,i.frequency,i.comparison_basis,d.url AS source_url
      FROM macro_indicator_bindings b JOIN macro_indicator_definitions i ON i.code=b.indicator_code
      JOIN macro_datasets d ON d.id=b.dataset_id WHERE b.dataset_id=%s AND b.enabled ORDER BY b.id""", (dataset,)).fetchall()


def claim(db, dataset, force=False):
    token = uuid.uuid4().hex
    with db.transaction():
        db.execute("INSERT INTO macro_pilot_state(dataset_id) VALUES(%s) ON CONFLICT DO NOTHING", (dataset,))
        if force:
            db.execute("UPDATE macro_pilot_state SET next_due=now() WHERE dataset_id=%s AND (lease_until IS NULL OR lease_until<=now())", (dataset,))
        row = db.execute("""UPDATE macro_pilot_state SET lease_token=%s,lease_until=now()+interval '10 minutes'
          WHERE dataset_id=%s AND next_due<=now() AND (lease_until IS NULL OR lease_until<=now()) RETURNING lease_token""", (token, dataset)).fetchone()
    return token if row else None


def persist(db, dataset, token, body, rows, config, failure=None):
    from psycopg.types.json import Jsonb
    with db.transaction():
        db.execute("SET TRANSACTION ISOLATION LEVEL REPEATABLE READ")
        state = db.execute("SELECT lease_token,lease_until>now() AS active FROM macro_pilot_state WHERE dataset_id=%s FOR UPDATE", (dataset["id"],)).fetchone()
        if not state or state["lease_token"] != token or not state["active"]:
            raise RuntimeError("Pilot lease expired/replaced")
        raw_id = None
        digest = hashlib.sha256(body).hexdigest() if body is not None else None
        if body is not None:
            raw_id = db.execute("""INSERT INTO macro_pilot_raw(dataset_id,checksum,payload,fetched_at) VALUES(%s,%s,%s,now())
              ON CONFLICT(dataset_id,checksum) DO UPDATE SET checksum=EXCLUDED.checksum RETURNING id""", (dataset["id"], digest, body)).fetchone()["id"]
        report = {"configHash": hashlib.sha256(json.dumps({"url":dataset["url"],"bindings":config},sort_keys=True,default=str).encode()).hexdigest(), "mode": "shadow", "bindings": [{k: b[k] for k in ("id", "indicator_code", "series_id", "extractor_key", "transform_key", "validation")} for b in config]}
        baseline_id = None
        if failure:
            status = "SOURCE_ERROR"
            report.update(errorKind=failure.kind, message=str(failure))
        elif dataset.get("baseline_policy")=="cross_source":
            from .cpi_shadow import compare as compare_reports
            status,result=compare_reports(db,dataset,rows,config,datetime.now(ZONE).date())
            report.update(result)
        else:
            baseline = db.execute("SELECT id,checksum,last_fetched_at FROM macro_raw_documents WHERE source='NSO' AND source_url=%s ORDER BY last_fetched_at DESC,id DESC LIMIT 1", (dataset["url"],)).fetchone()
            attempts = {row["indicator_code"]:row for row in db.execute("SELECT indicator_code,success,checked_at FROM macro_attempts WHERE provider_id=%s", (dataset["go_provider_id"],)).fetchall()}
            go_state = db.execute("SELECT lease_until>now() AS active FROM macro_source_state WHERE provider=%s", (dataset["go_provider_id"],)).fetchone()
            baseline_id = baseline["id"] if baseline else None
            if not baseline or any(binding["indicator_code"] not in attempts for binding in config):
                status = "NO_BASELINE"
            elif go_state and go_state["active"]:
                status = "BASELINE_BUSY"
            elif any(not attempts[binding["indicator_code"]]["success"] or attempts[binding["indicator_code"]]["checked_at"] < baseline["last_fetched_at"] for binding in config):
                status = "BASELINE_INCOMPLETE"
            elif baseline["checksum"] != digest:
                status = "VINTAGE_MISMATCH"
                report.update(pythonChecksum=digest, goChecksum=baseline["checksum"])
            else:
                reference = db.execute("""SELECT DISTINCT ON(o.indicator_code,o.series_id,o.period)
                    o.indicator_code,o.series_id,o.period,o.period_end,o.value,o.unit,o.method,o.source_url,o.priority
                    FROM macro_observations o JOIN macro_indicator_bindings b
                      ON b.indicator_code=o.indicator_code AND b.series_id=o.series_id
                    WHERE b.dataset_id=%s AND b.enabled
                    ORDER BY o.indicator_code,o.series_id,o.period,o.revision DESC""", (dataset["id"],)).fetchall()
                report.update(compare(rows, reference, config, datetime.now(ZONE).date()))
                status = "MISMATCH" if report["discrepancies"] else "STALE" if report["stale"] else ("AWAITING_RELEASE" if report["publicationWindow"] else "RELEASE_LAG") if report["behind"] else "MATCH"
        run_id = db.execute("INSERT INTO macro_pilot_runs(dataset_id,raw_id,parser_version,status,baseline_raw_id,report) VALUES(%s,%s,%s,%s,%s,%s) RETURNING id", (dataset["id"], raw_id, config[0]["extractor_key"] if config else VERSION, status, baseline_id, Jsonb(report))).fetchone()["id"]
        for row in rows:
            db.execute("""INSERT INTO macro_pilot_observations(run_id,binding_id,period,period_end,value,unit,series_id,method,source_url,published_at,evidence)
              VALUES(%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)""", (run_id, row["binding_id"], row["period"], row["period_end"], row["value"], row["unit"], row["series_id"], row["method"],row.get("source_url"),row.get("published_at"),Jsonb(row.get("evidence",{}))))
        delay = 300 if status in ("NO_BASELINE", "BASELINE_BUSY", "BASELINE_INCOMPLETE", "VINTAGE_MISMATCH") else dataset["interval_seconds"]
        if failure:
            delay = max(delay, min(failure.retry_after, 365 * 86400))
        updated = db.execute("""UPDATE macro_pilot_state SET next_due=now()+(%s*interval '1 second'),lease_token=NULL,lease_until=NULL,last_status=%s,last_run_id=%s
          WHERE dataset_id=%s AND lease_token=%s""", (delay, status, run_id, dataset["id"], token))
        if updated.rowcount != 1:
            raise RuntimeError("Pilot lease changed")
    LOG.info(json.dumps({"dataset": dataset["id"], "mode": "shadow", "status": status, "runId": run_id}))
    return status


def run(db, dataset_id="all", force=False, client_factory=HTTPClient):
    selected = [r["id"] for r in db.execute("SELECT id FROM macro_datasets WHERE id=ANY(%s) AND enabled ORDER BY id",(sorted(PILOT_IDS),)).fetchall()] if dataset_id == "all" else [dataset_id]
    output = []
    for dataset_id in selected:
        dataset = db.execute("SELECT * FROM macro_datasets WHERE id=%s AND enabled", (dataset_id,)).fetchone()
        if not dataset or dataset["ingestion_owner"] != "go":
            if dataset_id in PILOT_IDS and len(selected)>1:
                output.append({"dataset":dataset_id,"status":"DISABLED_OR_OTHER_OWNER"})
                continue
            raise ValueError("Shadow pilot requires an enabled dataset still owned by Go")
        reports=dataset.get("baseline_policy")=="cross_source"
        parsed = urlsplit(dataset["url"])
        if not reports and (parsed.scheme != "https" or parsed.netloc != "nsdp.nso.gov.vn" or not parsed.path.startswith("/GSO-chung/SDMXFiles/GSO/") or parsed.query or parsed.fragment):
            raise ValueError("Pilot dataset must use the official NSO SDMX directory")
        token = claim(db, dataset_id, force)
        if not token:
            output.append({"dataset": dataset_id, "status": "NOT_DUE_OR_BUSY"})
            continue
        config, body, rows, failure = bindings(db, dataset_id), None, [], None
        try:
            if reports:
                from .cpi_shadow import normalize
                body,rows=normalize(client_factory(),dataset,config,datetime.now(ZONE).date())
            else:
                body = client_factory().get(dataset["url"])
                rows = parse(body, config, datetime.now(ZONE).date())
        except CollectionError as error:
            failure = error
            body=getattr(error,"payload",body)
        output.append({"dataset": dataset_id, "status": persist(db, dataset, token, body, rows, config, failure)})
    return output


def report(db, dataset_id="all"):
    output = []
    for dataset in db.execute("SELECT * FROM macro_datasets WHERE id=ANY(%s) AND enabled ORDER BY id", (sorted(PILOT_IDS) if dataset_id == "all" else [dataset_id],)).fetchall():
        latest = db.execute("SELECT id,checked_at,status,parser_version,report FROM macro_pilot_runs WHERE dataset_id=%s ORDER BY id DESC LIMIT 1", (dataset["id"],)).fetchone()
        passes = db.execute("""SELECT count(DISTINCT (checked_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::date) AS days
          FROM macro_pilot_runs WHERE dataset_id=%s AND parser_version=%s AND status='MATCH'
          AND checked_at>now()-interval '14 days' AND (report->>'matched')::integer>=12 AND report->>'configHash'=%s""", (dataset["id"], VERSION, latest["report"].get("configHash","") if latest else "")).fetchone()["days"]
        eligible = bool(latest and latest["status"] == "MATCH" and latest["parser_version"]==VERSION and latest["checked_at"].timestamp() >= datetime.now(timezone.utc).timestamp()-86400 and latest["report"].get("matched", 0)>=12 and passes>=3)
        output.append({"dataset": dataset["id"], "owner": dataset["ingestion_owner"], "mode": "shadow", "matchingDays": passes, "eligibleForReview": eligible, "latestRun": latest, "cutover": "NOT_PERFORMED"})
    return output


def selected_datasets(db, dataset="all", indicator=None):
    if indicator:
        ids=[row["dataset_id"] for row in db.execute("SELECT DISTINCT b.dataset_id FROM macro_indicator_bindings b JOIN macro_datasets d ON d.id=b.dataset_id WHERE b.indicator_code=%s AND b.enabled AND d.enabled AND b.dataset_id=ANY(%s) ORDER BY b.dataset_id",(indicator,sorted(PILOT_IDS))).fetchall()]
        if not ids:
            raise ValueError("No enabled pilot binding for indicator")
        return ids
    return sorted(PILOT_IDS) if dataset=="all" else [dataset]


def main():
    parser = argparse.ArgumentParser(description="NSO Python shadow pilot; production writes and cutover are intentionally unavailable")
    parser.add_argument("action", choices=["run", "report"])
    selector=parser.add_mutually_exclusive_group()
    selector.add_argument("--dataset", choices=["all", *sorted(PILOT_IDS)], default="all")
    selector.add_argument("--indicator", choices=["VN_CPI_YOY","VN_CPI_MOM","VN_CPI_AVG_YOY","VN_IIP_YOY"])
    parser.add_argument("--once", action="store_true")
    parser.add_argument("--force", action="store_true")
    args = parser.parse_args()
    logging.basicConfig(level=logging.INFO, format="%(message)s")
    if args.force and (args.action!="run" or not args.once):
        parser.error("--force requires run --once")
    stop = threading.Event()
    for sig in (signal.SIGTERM, signal.SIGINT):
        signal.signal(sig, lambda *_: stop.set())
    while not stop.is_set():
        try:
            with connect() as db:
                if args.indicator:
                    result=[]
                    for dataset in selected_datasets(db, indicator=args.indicator):
                        result.extend(report(db,dataset) if args.action=="report" else run(db,dataset,args.force))
                else:
                    result = report(db, args.dataset) if args.action == "report" else run(db, args.dataset, args.force)
            print(json.dumps(result, default=str), flush=True)
            if args.action == "report" or args.once:
                return 1 if args.action=="run" and any(row["status"]=="SOURCE_ERROR" for row in result) else 0
        except Exception as error:
            LOG.error(json.dumps({"kind":type(error).__name__,"sqlstate":getattr(error,"sqlstate",None),"message":"Pilot database/run unavailable; check migrations, permissions and source configuration"}))
            if args.once or args.action == "report":
                return 1
        stop.wait(60)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
