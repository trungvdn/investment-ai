"""Immutable source vintages and gated, reversible national CPI selection."""
import hashlib
import json
import math
from datetime import timedelta

from .core import ZONE


def fingerprint(value):
    return hashlib.sha256(json.dumps(value, sort_keys=True, default=str, ensure_ascii=False).encode()).hexdigest()


def periods_between(first, last):
    year, month = map(int, first.split("-"))
    end_year, end_month = map(int, last.split("-"))
    if not 1 <= month <= 12 or not 1 <= end_month <= 12:
        raise ValueError("Invalid monthly range")
    start, end = year * 12 + month - 1, end_year * 12 + end_month - 1
    if start > end or end - start > 1200:
        raise ValueError("Invalid/unbounded monthly range")
    return [f"{number // 12:04d}-{number % 12 + 1:02d}" for number in range(start, end + 1)]


def plan(rows, policy, today):
    """Never average conflicts or interpolate missing months."""
    issues = []
    sources = {policy["history_binding"]: {}, policy["update_binding"]: {}}
    for row in rows:
        if row["binding_id"] not in sources:
            continue
        if row["period"] in sources[row["binding_id"]]:
            raise ValueError("Expected one current source vintage per month")
        sources[row["binding_id"]][row["period"]] = row
        if row["unit"] != "% YoY" or not math.isfinite(row["value"]) or row["value"] < -100 or row["period_end"] > today:
            issues.append({"kind": "INVALID_DEFINITION", "period": row["period"]})
    history, update = sources[policy["history_binding"]], sources[policy["update_binding"]]
    overlap = sorted(history.keys() & update.keys())
    comparisons = [{"period": period, "history": history[period]["value"], "update": update[period]["value"],
                    "difference": abs(history[period]["value"] - update[period]["value"]),
                    "matches": history[period]["unit"] == update[period]["unit"] and
                    abs(history[period]["value"] - update[period]["value"]) <= policy["tolerance"] + 1e-12}
                   for period in overlap]
    for row in comparisons:
        if not row["matches"]:
            issues.append({"kind": "SOURCE_CONFLICT", **row})
    longest, streak, previous = 0, 0, None
    for period in overlap:
        number = int(period[:4]) * 12 + int(period[5:])
        streak = streak + 1 if previous == number - 1 else 1
        longest, previous = max(longest, streak), number
    if longest < policy["min_overlap"]:
        issues.append({"kind": "INSUFFICIENT_OVERLAP", "required": policy["min_overlap"], "consecutive": longest})
    # Keep the validated SDMX vintage wherever historical months overlap.
    # Reports supply new periods and holes, never silently replace SDMX history.
    chosen = {**update, **history}
    periods = sorted(chosen)
    gaps = sorted(set(periods_between(periods[0], periods[-1])) - chosen.keys()) if periods else []
    if gaps:
        issues.append({"kind": "MISSING_MONTHS", "periods": gaps})
    expected = (today.replace(day=1) - timedelta(days=1)).strftime("%Y-%m")
    if not periods or periods[-1] < expected:
        issues.append({"kind": "LATEST_PERIOD_MISSING", "expected": expected, "latest": periods[-1] if periods else None})
    if not history or not update:
        issues.append({"kind": "SOURCE_EMPTY"})
    return {"indicator": policy["indicator_code"], "eligible": not issues,
            "historyBinding": policy["history_binding"], "updateBinding": policy["update_binding"],
            "tolerance": policy["tolerance"], "overlap": comparisons, "issues": issues,
            "first": periods[0] if periods else None, "latest": periods[-1] if periods else None,
            "expected": expected, "missingMonths": gaps,
            "selected": [{"period": period, "sourceValueId": chosen[period]["id"],
                          "binding": chosen[period]["binding_id"], "value": chosen[period]["value"],
                          "unit": chosen[period]["unit"], "sourceUrl": chosen[period]["source_url"]}
                         for period in periods]}


VALUE_FIELDS = ("value", "unit", "period_end", "method")


def same_value(left, right):
    """A revision records a value/definition change, not a crawl-layout change."""
    return all(left.get(field) == right.get(field) for field in VALUE_FIELDS)


def remember(db, run_id, raw_id, rows):
    """Within the ingestion transaction: unchanged data only refreshes last_seen."""
    from psycopg.types.json import Jsonb
    for row in sorted(rows, key=lambda r: (r["binding_id"], r["period"])):
        key = f"source-value:{row['binding_id']}:{row['period']}"
        db.execute("SELECT pg_advisory_xact_lock(hashtextextended(%s,0))", (key,))
        payload = {key: row.get(key) for key in VALUE_FIELDS}
        digest = fingerprint(payload)
        previous = db.execute("SELECT * FROM macro_source_values WHERE binding_id=%s AND period=%s ORDER BY revision DESC LIMIT 1", (row["binding_id"], row["period"])).fetchone()
        if previous and same_value(previous, row):
            # Compare actual fields so pre-fix hashes do not create one more
            # duplicate. Retain the first raw/evidence and every run snapshot.
            db.execute("UPDATE macro_source_values SET fingerprint=%s,last_seen_at=now() WHERE id=%s", (digest, previous["id"]))
            continue
        db.execute("""INSERT INTO macro_source_values(binding_id,period,period_end,revision,fingerprint,value,unit,method,source_url,published_at,evidence,raw_id,first_run_id)
          VALUES(%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)""",
          (row["binding_id"], row["period"], row["period_end"], previous["revision"] + 1 if previous else 1, digest,
           row["value"], row["unit"], row["method"], row["source_url"], row.get("published_at"), Jsonb(row.get("evidence", {})), raw_id, run_id))


def current_values(db, binding_ids):
    return db.execute("""SELECT DISTINCT ON(v.binding_id,v.period) v.*,b.indicator_code,b.series_id,b.priority,d.source_id
      FROM macro_source_values v JOIN macro_indicator_bindings b ON b.id=v.binding_id
      JOIN macro_datasets d ON d.id=b.dataset_id
      WHERE v.binding_id=ANY(%s) AND b.enabled AND d.enabled
      ORDER BY v.binding_id,v.period,v.revision DESC""", (binding_ids,)).fetchall()


def preview(db, code, today):
    policy = db.execute("SELECT * FROM macro_source_policies WHERE indicator_code=%s", (code,)).fetchone()
    if not policy or code != "VN_CPI_YOY":
        raise ValueError("No implemented source-selection policy for this indicator")
    result = plan(current_values(db, [policy["history_binding"], policy["update_binding"]]), policy, today)
    for binding in (policy["history_binding"], policy["update_binding"]):
        state = db.execute("""SELECT s.last_status,r.checked_at,b.indicator_code,b.enabled,d.enabled AS dataset_enabled,i.unit,i.frequency,i.comparison_basis
          FROM macro_indicator_bindings b JOIN macro_datasets d ON d.id=b.dataset_id
          JOIN macro_indicator_definitions i ON i.code=b.indicator_code
          LEFT JOIN macro_pilot_state s ON s.dataset_id=b.dataset_id LEFT JOIN macro_pilot_runs r ON r.id=s.last_run_id
          WHERE b.id=%s""", (binding,)).fetchone()
        if not state or state["indicator_code"] != code or state["unit"] != "% YoY" or state["frequency"] != "MONTHLY" or state["comparison_basis"] != "YoY" or not state["enabled"] or not state["dataset_enabled"] or not state["checked_at"] or state["checked_at"].astimezone(ZONE).date() < today - timedelta(days=1) or state["last_status"] in ("SOURCE_ERROR", "PARTIAL"):
            result["issues"].append({"kind": "SOURCE_NOT_RECENTLY_VERIFIED", "binding": binding})
    result["eligible"] = not result["issues"]
    result["activeBatchId"] = policy["active_batch_id"]
    result["mode"] = "preview"
    return result


def publish_value(db, row):
    """Production revisions share Go's exact advisory-lock key."""
    key = row["indicator_code"] + "|" + row["period"] + "|" + row["series_id"]
    db.execute("SELECT pg_advisory_xact_lock(hashtextextended(%s,0))", (key,))
    old = db.execute("SELECT * FROM macro_observations WHERE indicator_code=%s AND period=%s AND series_id=%s ORDER BY revision DESC LIMIT 1", (row["indicator_code"], row["period"], row["series_id"])).fetchone()
    if old and same_value(old, row):
        db.execute("UPDATE macro_observations SET collected_at=GREATEST(collected_at,%s) WHERE id=%s", (row["last_seen_at"], old["id"]))
        return old["id"]
    raw = db.execute("SELECT payload,checksum,fetched_at FROM macro_pilot_raw WHERE id=%s", (row["raw_id"],)).fetchone()
    raw_id = db.execute("""INSERT INTO macro_raw_documents(source,source_url,content_type,checksum,payload,first_fetched_at,last_fetched_at)
      VALUES('NSO',%s,%s,%s,%s,%s,%s) ON CONFLICT(source,source_url,checksum) DO UPDATE
      SET last_fetched_at=GREATEST(macro_raw_documents.last_fetched_at,EXCLUDED.last_fetched_at) RETURNING id""",
      (row["source_url"], "application/xml" if row["binding_id"] == "nso_cpi_yoy" else "application/json", raw["checksum"], raw["payload"], raw["fetched_at"], raw["fetched_at"])).fetchone()["id"]
    return db.execute("""INSERT INTO macro_observations(indicator_code,period,period_end,value,unit,source,source_url,series_id,method,priority,published_at,first_seen_at,collected_at,revision,raw_id)
      VALUES(%s,%s,%s,%s,%s,'NSO',%s,%s,%s,%s,%s,now(),now(),%s,%s) RETURNING id""",
      (row["indicator_code"], row["period"], row["period_end"], row["value"], row["unit"], row["source_url"], row["series_id"], row["method"], row["priority"], row["published_at"], old["revision"] + 1 if old else 1, raw_id)).fetchone()["id"]


def activate(db, code, today, automatic=False):
    # Go may commit a new production revision after our repeatable-read snapshot
    # but before we acquire its value lock. Retry the entire rolled-back selection
    # with a fresh snapshot, never just the conflicting INSERT.
    for attempt in range(3):
        try:
            return _activate(db, code, today, automatic)
        except Exception as error:
            if getattr(error, "sqlstate", None) not in ("40001", "40P01", "23505") or attempt == 2:
                raise


def _activate(db, code, today, automatic=False):
    from psycopg.types.json import Jsonb
    with db.transaction():
        db.execute("SET TRANSACTION ISOLATION LEVEL REPEATABLE READ")
        db.execute("SELECT pg_advisory_xact_lock(hashtextextended(%s,0))", ("source-selection:" + code,))
        policy = db.execute("SELECT * FROM macro_source_policies WHERE indicator_code=%s", (code,)).fetchone()
        if not policy or (automatic and policy["active_batch_id"] is None):
            return {"indicator": code, "mode": "inactive"}
        policy = db.execute("SELECT * FROM macro_source_policies WHERE indicator_code=%s FOR UPDATE", (code,)).fetchone()
        result = preview(db, code, today)
        if not result["eligible"]:
            result["mode"] = "blocked_previous_selection_retained"
            if policy["active_batch_id"] is not None:
                record_health(db, code, False, "Source selection blocked; prior batch retained")
            return result
        digest = fingerprint({"selected": result["selected"], "history": policy["history_binding"], "update": policy["update_binding"],
                              "tolerance": policy["tolerance"], "minOverlap": policy["min_overlap"]})
        batch = db.execute("""INSERT INTO macro_canonical_batches(indicator_code,fingerprint,report) VALUES(%s,%s,%s)
          ON CONFLICT(indicator_code,fingerprint) DO UPDATE SET fingerprint=EXCLUDED.fingerprint RETURNING id""", (code, digest, Jsonb(result))).fetchone()["id"]
        rows = {row["id"]: row for row in current_values(db, [policy["history_binding"], policy["update_binding"]])}
        for selected in result["selected"]:
            value = rows[selected["sourceValueId"]]
            existing = db.execute("SELECT observation_id FROM macro_canonical_links WHERE batch_id=%s AND period=%s", (batch, value["period"])).fetchone()
            if not existing:
                observation_id = publish_value(db, value)
                db.execute("INSERT INTO macro_canonical_links(batch_id,period,source_value_id,observation_id) VALUES(%s,%s,%s,%s)", (batch, value["period"], value["id"], observation_id))
        db.execute("UPDATE macro_source_policies SET active_batch_id=%s WHERE indicator_code=%s", (batch, code))
        if policy["active_batch_id"] != batch:
            db.execute("INSERT INTO macro_selection_events(indicator_code,previous_batch_id,next_batch_id,reason) VALUES(%s,%s,%s,%s)", (code, policy["active_batch_id"], batch, "automatic refresh" if automatic else "explicit activation"))
        record_health(db, code, True, "Canonical CPI selection verified")
        result.update(mode="active", activeBatchId=batch)
        return result


def deactivate(db, code):
    with db.transaction():
        db.execute("SELECT pg_advisory_xact_lock(hashtextextended(%s,0))", ("source-selection:" + code,))
        policy = db.execute("SELECT * FROM macro_source_policies WHERE indicator_code=%s FOR UPDATE", (code,)).fetchone()
        if not policy:
            raise ValueError("Unknown source-selection policy")
        if policy["active_batch_id"] is not None:
            db.execute("INSERT INTO macro_selection_events(indicator_code,previous_batch_id,reason) VALUES(%s,%s,'return to Go priority selection')", (code, policy["active_batch_id"]))
            db.execute("UPDATE macro_source_policies SET active_batch_id=NULL WHERE indicator_code=%s", (code,))
    return {"indicator": code, "mode": "go_priority_selection"}


def record_health(db, code, success, message):
    db.execute("""INSERT INTO macro_attempts(indicator_code,provider_id,checked_at,success,message)
      VALUES(%s,%s,now(),%s,%s) ON CONFLICT(indicator_code,provider_id) DO UPDATE
      SET checked_at=EXCLUDED.checked_at,success=EXCLUDED.success,message=EXCLUDED.message""",
      (code, "python:canonical:" + code, success, message))
