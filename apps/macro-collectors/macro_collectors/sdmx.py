"""Versioned monthly NSO extractor, configured by indicator-dataset bindings."""
import calendar
import math
import re
import xml.etree.ElementTree as ET
from datetime import date, timedelta

from .http import CollectionError

VERSION = "nso_sdmx_v1"


def month_end(period):
    if not re.fullmatch(r"\d{4}-\d{2}", period):
        raise ValueError("Invalid monthly period")
    year, month = map(int, period.split("-"))
    return date(year, month, calendar.monthrange(year, month)[1])


def parse(body, bindings, today):
    if b"<!DOCTYPE" in body.upper() or b"<!ENTITY" in body.upper():
        raise CollectionError("DATA_QUALITY", "XML declarations/entities are not supported")
    try:
        root = ET.fromstring(body)
        output = []
        for binding in bindings:
            if binding["extractor_key"] != VERSION or binding["transform_key"] != "yoy_12m":
                raise ValueError("Unrecognized binding parser/transform")
            rules = binding["validation"]
            candidates = [node for node in root.iter() if node.tag.split("}")[-1] == "Series" and node.get("INDICATOR") == binding["series_id"] and node.get("REF_AREA") == rules["ref_area"] and node.get("COUNTERPART_AREA") == rules["counterpart"] and node.get("FREQ") == rules["frequency"]]
            if len(candidates) != 1:
                raise ValueError("Missing/ambiguous NSO series")
            series = candidates[0]
            if series.get("UNIT_MULT") != rules["unit_mult"] or not series.get("BASE_PER"):
                raise ValueError("Unexpected NSO base or multiplier")
            levels = {}
            for node in series:
                if node.tag.split("}")[-1] != "Obs":
                    continue
                period = node.get("TIME_PERIOD", "")
                month_end(period)
                value = float(node.get("OBS_VALUE", ""))
                if period in levels or not math.isfinite(value):
                    raise ValueError("Duplicate/nonfinite NSO observation")
                levels[period] = value
            rows = []
            for period, value in sorted(levels.items()):
                previous = f"{int(period[:4])-1:04d}{period[4:]}"
                if previous not in levels or levels[previous] <= 0:
                    continue
                end = month_end(period)
                transformed = (value / levels[previous] - 1) * 100
                if end > today or not math.isfinite(transformed) or transformed < -100:
                    raise ValueError("Future/out-of-definition NSO observation")
                rows.append({"binding_id": binding["id"], "indicator_code": binding["indicator_code"], "period": period, "period_end": end, "value": transformed, "unit": binding["unit"], "series_id": binding["series_id"], "source_url": binding["source_url"], "priority": binding["priority"], "method": "YOY_FROM_SAME_MONTH_INDEX_OR_LEVEL; base=" + series.get("BASE_PER") + "; unit_mult=" + series.get("UNIT_MULT")})
            if not rows:
                raise ValueError("NSO series lacks comparable prior-year months")
            output.extend(rows)
        if not output:
            raise ValueError("No enabled bindings")
        return output
    except (ET.ParseError, ValueError, OverflowError, TypeError, KeyError):
        raise CollectionError("DATA_QUALITY", "NSO source does not satisfy the configured binding contract") from None


def compare(rows, reference, bindings, today):
    expected = {(row["indicator_code"], row["series_id"], row["period"]): row for row in reference}
    actual = {(row["indicator_code"], row["series_id"], row["period"]): row for row in rows}
    rules = {binding["indicator_code"]: binding["validation"] for binding in bindings}
    discrepancies = []
    matched = 0
    for key in sorted(actual.keys() | expected.keys()):
        python, go = actual.get(key), expected.get(key)
        reason = None
        if python is None:
            reason = "MISSING_PYTHON"
        elif go is None:
            reason = "MISSING_GO"
        elif any(python[field] != go[field] for field in ("unit", "period_end", "method", "source_url", "priority")):
            reason = "METADATA_MISMATCH"
        elif not math.isfinite(go["value"]) or abs(python["value"] - go["value"]) > rules[key[0]]["tolerance"]:
            reason = "VALUE_MISMATCH"
        if reason:
            discrepancies.append({"indicator": key[0], "series": key[1], "period": key[2], "reason": reason, "python": python["value"] if python else None, "go": go["value"] if go else None})
        else:
            matched += 1
    latest = {}
    for row in rows:
        latest[row["indicator_code"]] = max(latest.get(row["indicator_code"], row["period_end"]), row["period_end"])
    stale = [code for code, end in latest.items() if (today - end).days > rules[code]["max_age_days"]]
    expected_end = today.replace(day=1) - timedelta(days=1)
    behind = [code for code, end in latest.items() if end < expected_end]
    return {"expectedPeriod": expected_end.strftime("%Y-%m"), "behind": behind, "publicationWindow": today.day<=10, "matched": matched, "discrepancies": discrepancies, "stale": stale, "latest": {code: end.isoformat() for code, end in latest.items()}}
