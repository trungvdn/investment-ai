import calendar
import copy
import json
import os
import unittest
from datetime import date, datetime

from macro_collectors.core import ZONE
from macro_collectors.selection import plan, periods_between
from macro_collectors.sources import inventory
from macro_collectors.cpi_reports import collect_archive, INDEX
from test_cpi_reports import article, SUMMARY, URL
from test_pilot import PostgresPilotFixture


POLICY = {"indicator_code": "VN_CPI_YOY", "history_binding": "history", "update_binding": "update", "tolerance": 0.02, "min_overlap": 3}


def row(binding, period, value=3.0, identifier=1):
    year, month = map(int, period.split("-"))
    return {"id": identifier, "binding_id": binding, "period": period,
            "period_end": date(year, month, calendar.monthrange(year, month)[1]),
            "value": value, "unit": "% YoY", "source_url": "https://example.test/fixture"}


class SelectionTests(unittest.TestCase):
    def fixture(self):
        return [row("history", f"2026-{m:02d}", identifier=m) for m in range(1, 7)] + [row("update", f"2026-{m:02d}", 3.01, 100+m) for m in range(4, 10)]

    def test_history_and_updates_stitch_only_after_consecutive_overlap(self):
        result = plan(self.fixture(), POLICY, date(2026, 10, 26))
        self.assertTrue(result["eligible"])
        self.assertEqual(len(result["overlap"]), 3)
        self.assertEqual(len(result["selected"]), 9)
        self.assertEqual(result["selected"][0]["binding"], "history")
        self.assertEqual(result["selected"][-1]["binding"], "update")
        self.assertEqual(result["selected"][3]["value"], 3.01)

    def test_conflicts_gaps_stale_wrong_unit_and_insufficient_overlap_block(self):
        cases = []
        values = self.fixture(); values[-4]["value"] = 4; cases.append((values, "SOURCE_CONFLICT"))
        values = self.fixture(); values = [v for v in values if v["period"] != "2026-03"]; cases.append((values, "MISSING_MONTHS"))
        values = self.fixture(); values[-1]["unit"] = "% MoM"; cases.append((values, "INVALID_DEFINITION"))
        values = self.fixture(); values = [v for v in values if v["period"] != "2026-09"]; cases.append((values, "LATEST_PERIOD_MISSING"))
        values = self.fixture(); values = [v for v in values if not (v["binding_id"] == "update" and v["period"] == "2026-05")]; cases.append((values, "INSUFFICIENT_OVERLAP"))
        for values, reason in cases:
            with self.subTest(reason=reason):
                result = plan(values, POLICY, date(2026, 10, 26))
                self.assertFalse(result["eligible"])
                self.assertIn(reason, {issue["kind"] for issue in result["issues"]})

    def test_month_range_bounds_and_inventory_has_no_fabricated_history(self):
        self.assertEqual(periods_between("2025-12", "2026-02"), ["2025-12", "2026-01", "2026-02"])
        with self.assertRaises(ValueError): periods_between("2026-12", "2026-01")
        data = inventory()["indicators"]
        self.assertEqual(len(data), 20)
        self.assertTrue(all(item["observedRange"] is None for item in data))
        pmi = inventory("VN_PMI")["indicators"][0]
        self.assertIsNone(pmi["historySource"])
        self.assertEqual(pmi["historyStatus"], "UNRESOLVED")

    def test_archive_page_query_and_card_publication_year(self):
        calls = []
        class Client:
            def get(self, url):
                calls.append(url)
                if url == INDEX + "?paged=2":
                    return ('<a href="' + URL + '">Chỉ số giá tiêu dùng tháng Chín năm 2026 Ngày đăng: 01/01/2027</a>').encode()
                return article().encode()
        raw, values, warnings = collect_archive(Client(), INDEX, 2, date(2026, 10, 26), "2026-01", "2026-09")
        self.assertEqual(len(values), 3)
        self.assertFalse(warnings)
        self.assertEqual(calls[0], INDEX + "?paged=2")
        self.assertIn("articles", json.loads(raw))

    def test_historical_release_before_month_end_is_valid(self):
        from macro_collectors.cpi_reports import extract_article
        html = article().replace("03/10/2026", "29/09/2026")
        self.assertEqual(len(extract_article(html, URL, date(2026, 10, 26))), 3)

    def test_archive_partial_failure_preserves_good_rows_and_warnings(self):
        class Client:
            def get(self, url):
                if url == INDEX:
                    return ('<a href="' + URL + '">Chỉ số giá tiêu dùng tháng Chín năm 2026</a><a href="' + URL + '?bad=1">Chỉ số giá tiêu dùng tháng Chín năm 2026</a>').encode()
                return b'<h1>broken report</h1>' if '?bad' in url else article().encode()
        raw, values, warnings = collect_archive(Client(), INDEX, 1, date(2026, 10, 26), "2026-09", "2026-09")
        self.assertEqual(len(values), 3)
        self.assertEqual(warnings[0]["kind"], "DATA_QUALITY")
        self.assertEqual(len(json.loads(raw)["articles"]), 2)


@unittest.skipUnless(os.getenv("MACRO_TEST_DATABASE_URL"), "requires CI PostgreSQL")
class SourceSelectionDatabaseTests(PostgresPilotFixture):
    def seed(self):
        from macro_collectors.pilot import bindings, claim, persist
        today = datetime.now(ZONE).date()
        end = today.year * 12 + today.month - 2
        self.months = [f"{(end - 5 + i) // 12:04d}-{(end - 5 + i) % 12 + 1:02d}" for i in range(6)]
        self.db.execute("UPDATE macro_datasets SET enabled=true WHERE id='gso_cpi_reports'")
        self.config = {}
        self.rows = {}
        for dataset, indexes in (("nso_cpi", range(4)), ("gso_cpi_reports", range(1, 6))):
            config = bindings(self.db, dataset)
            binding = next(b for b in config if b["indicator_code"] == "VN_CPI_YOY")
            values = []
            for i in indexes:
                value = row(binding["id"], self.months[i], 3 + i / 10)
                value.update(indicator_code="VN_CPI_YOY", series_id=binding["series_id"], priority=binding["priority"],
                             source_url=binding["source_url"], method="fixture_" + binding["extractor_key"],
                             published_at=None, evidence={})
                values.append(value)
            self.rows[dataset], self.config[dataset] = values, config
            self.save(dataset, values)
        return today

    def save(self, dataset, values):
        from macro_collectors.pilot import claim, persist
        metadata = self.db.execute("SELECT * FROM macro_datasets WHERE id=%s", (dataset,)).fetchone()
        return persist(self.db, metadata, claim(self.db, dataset, True),
                       json.dumps(values, default=str).encode(), values, self.config[dataset])

    def test_source_vintages_idempotency_coverage_and_revisions(self):
        from macro_collectors.sources import coverage
        self.seed()
        self.save("gso_cpi_reports", self.rows["gso_cpi_reports"])
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_source_values").fetchone()["n"], 9)
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_observations").fetchone()["n"], 0)
        values = copy.deepcopy(self.rows["gso_cpi_reports"]); values[-1]["value"] += 0.1
        self.save("gso_cpi_reports", values)
        self.assertEqual(self.db.execute("SELECT max(revision) AS n FROM macro_source_values").fetchone()["n"], 2)
        result = coverage(self.db, "VN_CPI_YOY", self.months[0], self.months[-1])[0]
        report = next(source for source in result["pythonSources"] if source["id"] == "gso_cpi_yoy")
        self.assertEqual(report["missingRequestedMonths"], [self.months[0]])

    def test_activation_is_gated_atomic_idempotent_and_reversible(self):
        from macro_collectors.selection import activate, deactivate, preview
        today = self.seed()
        self.assertTrue(preview(self.db, "VN_CPI_YOY", today)["eligible"])
        self.assertIsNone(self.db.execute("SELECT active_batch_id FROM macro_source_policies").fetchone()["active_batch_id"])
        first = activate(self.db, "VN_CPI_YOY", today)
        self.assertEqual(first["mode"], "active")
        self.assertEqual(len(first["selected"]), 6)
        second = activate(self.db, "VN_CPI_YOY", today)
        self.assertEqual(first["activeBatchId"], second["activeBatchId"])
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_observations").fetchone()["n"], 6)
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_selection_events").fetchone()["n"], 1)
        bad = copy.deepcopy(self.rows["gso_cpi_reports"]); bad[0]["value"] += 1
        self.save("gso_cpi_reports", bad)
        blocked = activate(self.db, "VN_CPI_YOY", today, automatic=True)
        self.assertEqual(blocked["mode"], "blocked_previous_selection_retained")
        self.assertEqual(self.db.execute("SELECT active_batch_id FROM macro_source_policies").fetchone()["active_batch_id"], first["activeBatchId"])
        health = self.db.execute("SELECT success FROM macro_attempts WHERE provider_id='python:canonical:VN_CPI_YOY'").fetchone()
        self.assertFalse(health["success"])
        deactivate(self.db, "VN_CPI_YOY")
        self.assertIsNone(self.db.execute("SELECT active_batch_id FROM macro_source_policies").fetchone()["active_batch_id"])
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_observations").fetchone()["n"], 6)

    def test_missing_overlap_never_publishes(self):
        from macro_collectors.selection import activate
        today = self.seed()
        self.db.execute("DELETE FROM macro_source_values WHERE binding_id='gso_cpi_yoy' AND period=%s", (self.months[1],))
        result = activate(self.db, "VN_CPI_YOY", today)
        self.assertFalse(result["eligible"])
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_observations").fetchone()["n"], 0)
