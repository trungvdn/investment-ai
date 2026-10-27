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
        self.assertEqual(result["selected"][3]["value"], 3.0)
        self.assertEqual(result["selected"][3]["binding"], "history")

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

    def test_normal_ingestion_refreshes_active_selection_and_source_failure_retains_it(self):
        from macro_collectors.selection import activate
        from macro_collectors.pilot import run
        from macro_collectors.http import CollectionError
        today = self.seed()
        first = activate(self.db, "VN_CPI_YOY", today)
        year, month = map(int, self.months[-1].split("-"))
        url = f"https://www.nso.gov.vn/du-lieu-va-so-lieu-thong-ke/{today.year}/fixture-cpi/"
        html = (f"<h1>Chỉ số giá tiêu dùng tháng {month} năm {year}</h1>"
                f"<p>CPI tháng {month} tăng 0,1% so với tháng trước; tăng 3,7% so với cùng kỳ năm trước.</p>"
                f"<p>Ngày đăng: {today.strftime('%d/%m/%Y')}</p>").encode()

        class Client:
            def get(self, requested):
                if requested.endswith("/cpi-vi/"):
                    return (f'<a href="{url}">Chỉ số giá tiêu dùng tháng {month} năm {year}</a>').encode()
                return html

        result = run(self.db, "gso_cpi_reports", True, Client)[0]
        self.assertEqual(result["canonical"], "active")
        current = self.db.execute("SELECT active_batch_id FROM macro_source_policies").fetchone()["active_batch_id"]
        self.assertNotEqual(first["activeBatchId"], current)
        selected = self.db.execute("""SELECT o.value FROM macro_canonical_links l JOIN macro_observations o ON o.id=l.observation_id
          WHERE l.batch_id=%s AND l.period=%s""", (current, self.months[-1])).fetchone()
        self.assertEqual(selected["value"], 3.7)

        class Denied:
            def get(self, requested):
                raise CollectionError("ACCESS_DENIED", "fixture HTTP 403")

        failed = run(self.db, "gso_cpi_reports", True, Denied)[0]
        self.assertEqual(failed["canonical"], "blocked_previous_selection_retained")
        self.assertEqual(self.db.execute("SELECT active_batch_id FROM macro_source_policies").fetchone()["active_batch_id"], current)

    def test_bounded_backfill_reruns_do_not_duplicate_source_vintages(self):
        from macro_collectors.sources import backfill
        from test_pilot import fixture
        self.db.execute("UPDATE macro_datasets SET enabled=true WHERE id='gso_cpi_reports'")
        body = fixture()

        class Client:
            def get(self, url):
                if url.endswith("CPIVNM.xml"):
                    return body
                if url.endswith("/cpi-vi/"):
                    return ('<a href="' + URL + '">Chỉ số giá tiêu dùng tháng Chín năm 2026</a>').encode()
                return article().encode()

        backfill(self.db, "VN_CPI_YOY", "2026-09", "2026-09", 1, 1, Client, source="combined")
        before = self.db.execute("SELECT count(*) AS n FROM macro_source_values").fetchone()["n"]
        backfill(self.db, "VN_CPI_YOY", "2026-09", "2026-09", 1, 1, Client)
        self.assertEqual(before, self.db.execute("SELECT count(*) AS n FROM macro_source_values").fetchone()["n"])
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_observations").fetchone()["n"], 0)

    def test_sdmx_only_backfill_downloads_full_history_without_report_requests(self):
        from macro_collectors.sources import backfill
        from test_pilot import fixture
        from macro_collectors.sdmx import parse
        from macro_collectors.pilot import bindings
        body = fixture()
        config = bindings(self.db, "nso_cpi")
        expected = parse(body, config, datetime.now(ZONE).date())
        first, last = expected[0]["period"], expected[-1]["period"]
        requested_urls = []

        class Client:
            def get(self, url):
                requested_urls.append(url)
                if not url.endswith("/CPIVNM.xml"):
                    raise AssertionError("SDMX-only backfill must not fetch report HTML")
                return body

        # Report dataset stays disabled: historical loading must be independent.
        all_history = backfill(self.db, "VN_CPI_YOY", None, None, 1, 3, Client)[-1]
        self.assertEqual(all_history["status"], "HISTORY_LOADED")
        self.assertEqual(all_history["coverageScope"], "available_sdmx_history")
        self.assertEqual(all_history["storedPeriods"], len(expected))
        requested_urls.clear()
        result = backfill(self.db, "VN_CPI_YOY", last, last, 1, 3, Client)
        summary = result[-1]
        self.assertEqual(summary["status"], "BACKFILL_COMPLETE")
        self.assertEqual((summary["availableFrom"], summary["availableTo"]), (first, last))
        self.assertEqual(summary["storedPeriods"], len(expected))
        self.assertEqual(summary["coveredRequestedPeriods"], 1)
        self.assertEqual(len(requested_urls), 1)
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_source_values").fetchone()["n"], len(expected))
        backfill(self.db, "VN_CPI_YOY", last, last, 1, 3, Client)
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_source_values").fetchone()["n"], len(expected))
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_observations").fetchone()["n"], 0)

    def test_sdmx_only_reports_requested_months_not_present_in_the_source(self):
        from macro_collectors.sources import backfill
        from test_pilot import fixture
        body = fixture(lag=4)

        class Client:
            def get(self, url):
                return body

        today = datetime.now(ZONE).date()
        end = today.year * 12 + today.month - 2
        requested = f"{end // 12:04d}-{end % 12 + 1:02d}"
        result = backfill(self.db, "VN_CPI_YOY", requested, requested, 1, 3, Client)
        self.assertEqual(result[-1]["status"], "BACKFILL_INCOMPLETE")
        self.assertEqual(result[-1]["missingRequestedMonths"], [requested])

    def test_evidence_url_and_publication_changes_do_not_create_value_revisions(self):
        from macro_collectors.selection import activate, fingerprint
        from datetime import timedelta
        today = self.seed()
        active = activate(self.db, "VN_CPI_YOY", today)["activeBatchId"]
        initial_source = self.db.execute("SELECT count(*) AS n FROM macro_source_values").fetchone()["n"]
        initial_production = self.db.execute("SELECT count(*) AS n FROM macro_observations").fetchone()["n"]
        initial_raw = self.db.execute("SELECT count(*) AS n FROM macro_pilot_raw").fetchone()["n"]
        values = copy.deepcopy(self.rows["gso_cpi_reports"])
        original = self.db.execute("SELECT * FROM macro_source_values WHERE binding_id='gso_cpi_yoy' ORDER BY period LIMIT 1").fetchone()
        # Exercise legacy hashes already stored by the previous implementation.
        self.db.execute("UPDATE macro_source_values SET fingerprint='legacy-metadata-hash' WHERE binding_id='gso_cpi_yoy'")
        for i in range(3):
            for value in values:
                value["source_url"] = f"https://www.nso.gov.vn/reissued-cpi-{i}/"
                value["published_at"] = datetime.now(ZONE) + timedelta(seconds=i)
                value["evidence"] = {"text": "same published value", "paragraphIndex": i}
            self.save("gso_cpi_reports", values)
            result = activate(self.db, "VN_CPI_YOY", today)
            self.assertEqual(result["activeBatchId"], active)
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_source_values").fetchone()["n"], initial_source)
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_observations").fetchone()["n"], initial_production)
        self.assertGreater(self.db.execute("SELECT count(*) AS n FROM macro_pilot_raw").fetchone()["n"], initial_raw)
        retained = self.db.execute("SELECT * FROM macro_source_values WHERE id=%s", (original["id"],)).fetchone()
        self.assertEqual(retained["raw_id"], original["raw_id"])
        self.assertEqual(retained["source_url"], original["source_url"])
        self.assertEqual(retained["fingerprint"], fingerprint({key: retained[key] for key in ("value", "unit", "period_end", "method")}))

    def test_real_value_changes_and_return_to_old_value_remain_distinct_revisions(self):
        self.seed()
        baseline = copy.deepcopy(self.rows["gso_cpi_reports"])
        revised = copy.deepcopy(baseline)
        revised[-1]["value"] += 0.2
        self.save("gso_cpi_reports", revised)
        self.save("gso_cpi_reports", revised)
        self.save("gso_cpi_reports", baseline)
        rows = self.db.execute("SELECT revision,value FROM macro_source_values WHERE binding_id='gso_cpi_yoy' AND period=%s ORDER BY revision", (self.months[-1],)).fetchall()
        self.assertEqual([r["revision"] for r in rows], [1, 2, 3])
        self.assertEqual(rows[0]["value"], rows[-1]["value"])
        self.assertNotEqual(rows[0]["value"], rows[1]["value"])

    def test_production_metadata_drift_does_not_create_duplicate_revision(self):
        from macro_collectors.selection import activate, publish_value, current_values
        today = self.seed()
        activate(self.db, "VN_CPI_YOY", today)
        current = current_values(self.db, ["gso_cpi_yoy"])[-1]
        initial = self.db.execute("SELECT count(*) AS n FROM macro_observations").fetchone()["n"]
        original_raw_count = self.db.execute("SELECT count(*) AS n FROM macro_raw_documents").fetchone()["n"]
        original_id = self.db.execute("SELECT id FROM macro_observations WHERE indicator_code='VN_CPI_YOY' AND period=%s AND series_id=%s ORDER BY revision DESC LIMIT 1", (current["period"], current["series_id"])).fetchone()["id"]
        current["source_url"] = "https://www.nso.gov.vn/reissued-but-unchanged/"
        current["published_at"] = datetime.now(ZONE)
        with self.db.transaction():
            repeated_id = publish_value(self.db, current)
        self.assertEqual(repeated_id, original_id)
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_observations").fetchone()["n"], initial)
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_raw_documents").fetchone()["n"], original_raw_count)

    def test_read_only_duplicate_diagnostics_separate_legacy_vintages_from_run_audit(self):
        from pathlib import Path
        self.seed()
        self.save("gso_cpi_reports", self.rows["gso_cpi_reports"])
        self.db.execute("""INSERT INTO macro_source_values(binding_id,period,period_end,revision,fingerprint,value,unit,method,source_url,published_at,evidence,raw_id,first_run_id)
          SELECT binding_id,period,period_end,revision+1,'legacy-duplicate',value,unit,method,source_url,published_at,evidence,raw_id,first_run_id
          FROM macro_source_values WHERE binding_id='gso_cpi_yoy' AND period=%s ORDER BY revision DESC LIMIT 1""", (self.months[-1],))
        queries = (Path(__file__).resolve().parents[1] / "sql" / "check_duplicates.sql").read_text()
        result = self.db.execute(queries, prepare=False)
        source_duplicates = result.fetchall()
        self.assertEqual(len(source_duplicates), 1)
        self.assertEqual(source_duplicates[0]["period"], self.months[-1])
        result.nextset()
        self.assertFalse(result.fetchall())
        result.nextset()
        self.assertTrue(result.fetchall())

    def test_concurrent_go_revision_retries_the_entire_canonical_transaction(self):
        from unittest.mock import patch
        from macro_collectors import selection
        today = self.seed()
        initial_batch = selection.activate(self.db, "VN_CPI_YOY", today)["activeBatchId"]
        values = copy.deepcopy(self.rows["gso_cpi_reports"])
        values[-1]["value"] = 3.7
        self.save("gso_cpi_reports", values)
        publish = selection.publish_value
        raced = []
        other = self.connection()

        def publish_after_go_commit(db, value):
            if value["period"] == self.months[-1] and not raced:
                raced.append(True)
                with other.transaction():
                    lock_key = value["indicator_code"] + "|" + value["period"] + "|" + value["series_id"]
                    other.execute("SELECT pg_advisory_xact_lock(hashtextextended(%s,0))", (lock_key,))
                    other.execute("""INSERT INTO macro_observations(indicator_code,period,period_end,value,unit,source,source_url,series_id,method,priority,published_at,first_seen_at,collected_at,revision,raw_id)
                      SELECT indicator_code,period,period_end,%s,unit,source,source_url,series_id,method,priority,published_at,now(),now(),revision+1,raw_id
                      FROM macro_observations WHERE indicator_code='VN_CPI_YOY' AND period=%s AND series_id=%s
                      ORDER BY revision DESC LIMIT 1""", (value["value"], value["period"], value["series_id"]))
            return publish(db, value)

        try:
            with patch.object(selection, "publish_value", side_effect=publish_after_go_commit):
                result = selection.activate(self.db, "VN_CPI_YOY", today)
        finally:
            other.close()
        self.assertTrue(raced)
        self.assertEqual(result["mode"], "active")
        self.assertNotEqual(initial_batch, result["activeBatchId"])
        revisions = self.db.execute("SELECT revision,value FROM macro_observations WHERE indicator_code='VN_CPI_YOY' AND period=%s AND series_id=%s ORDER BY revision", (self.months[-1], values[-1]["series_id"])).fetchall()
        self.assertEqual([row["revision"] for row in revisions], [1, 2])
        self.assertEqual(revisions[-1]["value"], 3.7)
