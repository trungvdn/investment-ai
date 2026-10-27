"""Source inventory, measured coverage, CPI backfill and canonical selection CLI."""
import argparse
import json
from datetime import datetime, timedelta
from pathlib import Path

from .core import ZONE
from .http import CollectionError, HTTPClient
from .pilot import connect, bindings, claim, persist, run
from .selection import activate, deactivate, periods_between, preview


def inventory(indicator="all"):
    data = json.loads(Path(__file__).with_name("source_catalog.json").read_text())
    if indicator != "all":
        data["indicators"] = [row for row in data["indicators"] if row["code"] == indicator]
        if not data["indicators"]:
            raise ValueError("Unknown indicator")
    return data


def coverage(db, indicator="all", first=None, last=None):
    catalog = {row["code"]: row for row in inventory()["indicators"]}
    if indicator != "all" and indicator not in catalog:
        raise ValueError("Unknown indicator")
    result = []
    for code in (sorted(catalog) if indicator == "all" else [indicator]):
        production = db.execute("""SELECT series_id,source,unit,min(period) AS first,max(period) AS last,
          count(DISTINCT period) AS periods,count(*) AS vintages FROM macro_observations
          WHERE indicator_code=%s GROUP BY series_id,source,unit ORDER BY series_id,source,unit""", (code,)).fetchall()
        configured = db.execute("""SELECT b.id,b.dataset_id,b.extractor_key,b.transform_key,b.enabled,d.enabled AS dataset_enabled,
          d.url,d.interval_seconds FROM macro_indicator_bindings b JOIN macro_datasets d ON d.id=b.dataset_id
          WHERE b.indicator_code=%s ORDER BY b.id""", (code,)).fetchall()
        shadow = []
        for binding in configured:
            rows = db.execute("""SELECT DISTINCT ON(period) period,revision,value FROM macro_source_values
              WHERE binding_id=%s ORDER BY period,revision DESC""", (binding["id"],)).fetchall()
            periods = [row["period"] for row in rows]
            monthly = catalog[code]["frequency"] == "MONTHLY"
            gaps = sorted(set(periods_between(periods[0], periods[-1])) - set(periods)) if periods and monthly else []
            requested = sorted(set(periods_between(first, last)) - set(periods)) if first and monthly else None
            state = db.execute("SELECT last_status,next_due,last_run_id FROM macro_pilot_state WHERE dataset_id=%s", (binding["dataset_id"],)).fetchone()
            shadow.append({**binding, "first": periods[0] if periods else None, "last": periods[-1] if periods else None,
                           "periods": len(periods), "revisions": sum(row["revision"] for row in rows),
                           "missingMonths": gaps, "missingRequestedMonths": requested, "state": state})
        policy = db.execute("SELECT * FROM macro_source_policies WHERE indicator_code=%s", (code,)).fetchone()
        result.append({"indicator": code, "definition": catalog[code], "productionSources": production,
                       "pythonSources": shadow, "selectionPolicy": policy})
    return result


def backfill(db, indicator, first, last, start_page, pages, client_factory=HTTPClient, source="sdmx"):
    from .cpi_reports import collect_archive
    from .cpi_shadow import bind_metrics
    if indicator != "VN_CPI_YOY":
        raise ValueError("Historical backfill is implemented for CPI first; other indicators remain inventoried")
    if source not in ("sdmx", "combined"):
        raise ValueError("Unknown CPI historical source")
    if source == "combined" and (not 1 <= pages <= 10 or not 1 <= start_page <= 100 or start_page + pages - 1 > 100):
        raise ValueError("Use 1-10 pages per bounded job, within archive pages 1-100")
    periods_between(first, last)
    results = run(db, "nso_cpi", True, client_factory)
    if results[0]["status"] in ("SOURCE_ERROR", "NOT_DUE_OR_BUSY"):
        return results
    if source == "sdmx":
        rows = db.execute("""SELECT DISTINCT period FROM macro_source_values WHERE binding_id='nso_cpi_yoy' ORDER BY period""").fetchall()
        available = [row["period"] for row in rows]
        requested = periods_between(first, last)
        missing = sorted(set(requested) - set(available))
        results.append({"dataset": "nso_cpi", "status": "BACKFILL_INCOMPLETE" if missing else "BACKFILL_COMPLETE",
                        "source": "sdmx", "series": "PCPI_IX", "requestedFrom": first, "requestedTo": last,
                        "availableFrom": available[0] if available else None, "availableTo": available[-1] if available else None,
                        "storedPeriods": len(available), "coveredRequestedPeriods": len(requested) - len(missing),
                        "missingRequestedMonths": missing,
                        "note": "All available transformed SDMX history is retained; requested dates bound the completeness check, not the upstream download. Missing prior-year levels cannot produce YoY."})
        return results
    dataset = db.execute("SELECT * FROM macro_datasets WHERE id='gso_cpi_reports' AND enabled").fetchone()
    if not dataset or dataset["ingestion_owner"] != "go":
        raise ValueError("CPI report shadow dataset must be enabled and Go-owned")
    config = bindings(db, dataset["id"])
    for page in range(start_page, start_page + pages):
        token = claim(db, dataset["id"], True)
        if not token:
            results.append({"page": page, "status": "NOT_DUE_OR_BUSY"})
            break
        body, rows, failure, warnings = None, [], None, []
        try:
            body, metrics, warnings = collect_archive(client_factory(), dataset["url"], page, datetime.now(ZONE).date(), first, last)
            rows = bind_metrics(metrics, config)
        except CollectionError as error:
            failure, body = error, getattr(error, "payload", body)
        status = persist(db, dataset, token, body, rows, config, failure,
                         notes={"backfill": {"page": page, "from": first, "to": last}, "warnings": warnings})
        item = {"dataset": dataset["id"], "page": page, "status": status, "observations": len(rows), "warnings": warnings}
        results.append(item)
        print(json.dumps(item, default=str), flush=True)
        activate(db, indicator, datetime.now(ZONE).date(), automatic=True)
    return results


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("action", choices=("inventory", "coverage", "backfill", "preview", "activate", "deactivate"))
    parser.add_argument("--indicator", default="all")
    parser.add_argument("--from", dest="first")
    parser.add_argument("--to", dest="last")
    parser.add_argument("--source", choices=("sdmx", "combined"), default="sdmx", help="Historical CPI source: SDMX only by default; combined also scans report archives")
    parser.add_argument("--start-page", type=int, default=1)
    parser.add_argument("--pages", type=int, default=3)
    args = parser.parse_args()
    if bool(args.first) != bool(args.last):
        parser.error("--from and --to must be provided together")
    if args.first:
        try:
            if len(args.first) != 7 or len(args.last) != 7:
                raise ValueError("Use YYYY-MM")
            periods_between(args.first, args.last)
        except ValueError:
            parser.error("Use a valid bounded YYYY-MM range")
    if args.action not in ("inventory", "coverage") and args.indicator == "all":
        parser.error("This action requires one explicit --indicator")
    if args.action == "backfill" and not args.first:
        parser.error("backfill requires --from YYYY-MM --to YYYY-MM")
    try:
        if args.action == "inventory":
            result = inventory(args.indicator)
        else:
            with connect() as db:
                today = datetime.now(ZONE).date()
                if args.action == "coverage":
                    result = coverage(db, args.indicator, args.first, args.last)
                elif args.action == "backfill":
                    result = backfill(db, args.indicator, args.first, args.last, args.start_page, args.pages, source=args.source)
                elif args.action == "preview":
                    with db.transaction():
                        db.execute("SET TRANSACTION ISOLATION LEVEL REPEATABLE READ READ ONLY")
                        result = preview(db, args.indicator, today)
                elif args.action == "activate":
                    result = activate(db, args.indicator, today)
                else:
                    result = deactivate(db, args.indicator)
        print(json.dumps(result, default=str, ensure_ascii=False), flush=True)
        if isinstance(result, list) and args.action == "backfill":
            return 1 if any(row["status"] in ("SOURCE_ERROR", "PARTIAL", "NOT_DUE_OR_BUSY", "BACKFILL_INCOMPLETE") for row in result) else 0
        if args.action == "activate" and result.get("mode") != "active":
            return 1
        return 0
    except Exception as error:
        # Do not leak connection strings, source tokens or raw document content.
        print(json.dumps({"status": "ERROR", "kind": type(error).__name__,
                          "sqlstate": getattr(error, "sqlstate", None),
                          "message": "Source command failed; check migrations, selector, permissions and source evidence"}), flush=True)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
