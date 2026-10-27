"""Historical and scheduled ingestion share the existing production tables."""
import argparse
import json
import signal
import threading
from datetime import datetime

from .core import ZONE
from .http import CollectionError, HTTPClient
from .pilot import connect, bindings
from .production import claim, persist, validated, coverage, report
from .selection import periods_between

DATASETS = {"sbv_deposit_rates","mbs_pmi_history","sbv_credit_history","sbv_interbank_history","fred_effr_history","fred_10y_history","nso_iip","nso_ppi_history","te_vietnam_pmi"}


def collect(client, dataset, config, today, page, first, last):
    if dataset["id"] == "mbs_pmi_history":
        from .mbs_history import collect as adapter
        return adapter(client, dataset, config, today, page, first, last)
    if dataset["id"] == "sbv_deposit_rates":
        from .sbv_deposit import collect as adapter
        return adapter(client,dataset,config,today,page,first,last)
    if dataset["id"].startswith("sbv_"):
        from .sbv_history import collect as adapter
        return adapter(client,dataset,config,today,page,first,last)
    if dataset["id"].startswith("fred_"):
        from .fred_history import collect as adapter
        return adapter(client,dataset,config,today,page,first,last)
    if dataset["id"] in ("nso_iip","nso_ppi_history"):
        from .nso_history import collect as adapter
        return adapter(client,dataset,config,today,page,first,last)
    if dataset["id"]=="te_vietnam_pmi":
        from .te_pmi import normalize
        body,rows=normalize(client,dataset,config,today)
        return body,rows,{"warnings":[],"latestOnly":True}
    raise CollectionError("CONFIGURATION", "Unsupported historical dataset")


def ingest(db, dataset_id, force=False, page=1, first=None, last=None, client_factory=HTTPClient):
    if dataset_id not in DATASETS:
        raise ValueError("Unknown historical dataset")
    dataset = db.execute("SELECT * FROM macro_datasets WHERE id=%s AND enabled", (dataset_id,)).fetchone()
    if not dataset:
        return {"dataset": dataset_id, "status": "DISABLED_OR_OTHER_OWNER"}
    config = bindings(db, dataset_id)
    if not config:
        raise ValueError("Historical dataset has no enabled bindings")
    token = claim(db, dataset["go_provider_id"], force)
    if not token:
        return {"dataset": dataset_id, "status": "NOT_DUE_OR_BUSY"}
    body, rows, failure, notes = None, [], None, {}
    try:
        body, rows, notes = collect(client_factory(), dataset, config, datetime.now(ZONE).date(), page, first, last)
        rows=validated(rows,config,datetime.now(ZONE).date())
        if not rows:
            failure = CollectionError("NO_DATA", "No verified historical observations in this bounded job")
    except CollectionError as error:
        failure, body = error, getattr(error, "payload", body)
        rows=[]
    notes["page"]=page
    result = persist(db, dataset, token, body, rows, config, failure, notes)
    return {**result,"dataset":dataset_id,"importedObservations":len(rows),"page":page,"publication":"DIRECT"}



def selected_datasets(db,dataset="all",indicator="all"):
    if indicator!="all":
        ids=[row["dataset_id"] for row in db.execute("""SELECT DISTINCT b.dataset_id FROM macro_indicator_bindings b
          JOIN macro_datasets d ON d.id=b.dataset_id WHERE b.indicator_code=%s AND b.enabled AND d.enabled
          AND b.dataset_id=ANY(%s) ORDER BY b.dataset_id""",
          (indicator,sorted(DATASETS) if dataset=="all" else [dataset])).fetchall()]
        if not ids:raise ValueError("No enabled history binding for indicator")
        return ids
    if dataset!="all":return [dataset]
    return [row["id"] for row in db.execute("SELECT id FROM macro_datasets WHERE id=ANY(%s) AND enabled ORDER BY id",(sorted(DATASETS),)).fetchall()]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("action", choices=("run", "backfill", "coverage", "report"))
    parser.add_argument("--dataset", choices=("all", *sorted(DATASETS)), default="all")
    parser.add_argument("--indicator", default="all")
    parser.add_argument("--from", dest="first")
    parser.add_argument("--to", dest="last")
    parser.add_argument("--start-page", type=int, default=1)
    parser.add_argument("--pages", type=int, default=1)
    parser.add_argument("--once", action="store_true")
    parser.add_argument("--force", action="store_true")
    args = parser.parse_args()
    if bool(args.first) != bool(args.last):
        parser.error("--from and --to must be supplied together")
    if args.first:
        try: periods_between(args.first, args.last)
        except ValueError: parser.error("Invalid YYYY-MM interval")
    if not 1<=args.pages<=10 or not 1<=args.start_page<=100 or args.start_page+args.pages>101:
        parser.error("Use 1-10 pages within 1-100")
    if args.action=="run" and not args.once and (args.force or args.pages!=1 or args.start_page!=1):
        parser.error("Daemon uses due schedules and page one; manual options require --once")
    stop = threading.Event()
    for sig in (signal.SIGTERM, signal.SIGINT):
        signal.signal(sig, lambda *_: stop.set())
    while not stop.is_set():
        try:
            with connect() as db:
                if args.action=="coverage":
                    result = coverage(db, selected_datasets(db,args.dataset,args.indicator), args.first, args.last)
                elif args.action=="report":
                    result = report(db,selected_datasets(db,args.dataset,args.indicator))
                else:
                    result=[]
                    for dataset in selected_datasets(db,args.dataset,args.indicator):
                        for page in range(args.start_page,args.start_page+args.pages):
                            item=ingest(db,dataset,args.force or args.action=="backfill",page,args.first,args.last)
                            result.append(item)
                            print(json.dumps(item,default=str),flush=True)
                            if item["status"]=="NOT_DUE_OR_BUSY":break
            print(json.dumps(result,default=str,ensure_ascii=False),flush=True)
            if args.action!="run" or args.once:
                return int(any(row.get("status") in ("SOURCE_ERROR","PARTIAL","NOT_DUE_OR_BUSY","DISABLED_OR_OTHER_OWNER") for row in result))
        except Exception as error:
            print(json.dumps({"status":"ERROR","kind":type(error).__name__,"sqlstate":getattr(error,"sqlstate",None),
                              "message":"Historical ingestion unavailable; check migration, permissions and source report"}),flush=True)
            if args.once or args.action!="run":return 1
        stop.wait(60)
    return 0


if __name__=="__main__":
    raise SystemExit(main())
