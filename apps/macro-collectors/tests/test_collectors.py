import fcntl
import io
import json
import tempfile
import unittest
import urllib.error
from datetime import datetime, timedelta, timezone
from pathlib import Path

from macro_collectors.core import IntervalSchedule, MonthlySchedule, Runner, ZONE, target_period
from macro_collectors.http import CollectionError, HTTPClient, NotPublished
from macro_collectors.pmi import PMICollector

NOW = datetime(2026, 10, 1, 1, tzinfo=timezone.utc)
RELEASE = "https://www.pmi.spglobal.com/Public/Home/PressRelease/dba8ca68ce13403591b68cb3e07175d9"
TEXT = "Embargoed until 1 October 2026 (0730 ICT / 0030 UTC) S&P Global Vietnam Manufacturing PMI September 2026"


class FakeCollector:
    schedule = MonthlySchedule()
    target = staticmethod(target_period)

    def __init__(self):
        self.calls = 0
        self.failure = None

    def collect(self, directory, target, now):
        self.calls += 1
        if self.failure:
            raise self.failure
        return {"releases": [{"path": str(directory / "raw.pdf"), "sourceUrl": RELEASE}]}


class CollectorsTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        self.collector = FakeCollector()
        self.runner = Runner(self.root, {"vn_pmi": self.collector})

    def state(self):
        return json.loads((self.root / "vn_pmi/manifest.json").read_text())

    def test_monthly_schedule_and_restart(self):
        self.assertEqual(target_period(datetime(2027, 1, 1, tzinfo=timezone.utc)), "2026-12")
        self.assertEqual(self.runner.run("vn_pmi", NOW), "READY")
        state = self.state()
        self.assertEqual(state["nextDue"], "2026-11-01T00:30:00Z")
        restarted = Runner(self.root, {"vn_pmi": self.collector})
        self.assertEqual(restarted.run("vn_pmi", NOW + timedelta(days=20)), "NOT_DUE")
        self.assertEqual(self.collector.calls, 1)
        # November 1, 2026 is Sunday: pending publication must be retried.
        self.collector.failure = NotPublished()
        sunday = datetime(2026, 11, 1, 1, tzinfo=timezone.utc)
        self.assertEqual(restarted.run("vn_pmi", sunday), "PENDING")
        self.assertEqual(self.state()["targetPeriod"], "2026-10")
        self.assertEqual(self.state()["nextDue"], "2026-11-01T07:00:00Z")
        self.assertEqual(restarted.run("vn_pmi", sunday + timedelta(hours=7)), "PENDING")
        self.assertEqual(restarted.run("vn_pmi", datetime(2026, 11, 11, 1, tzinfo=timezone.utc), force=True), "ERROR")
        self.assertEqual(self.state()["errorKind"], "RELEASE_OVERDUE")

    def test_error_preserves_previous_release_and_lock_blocks_force(self):
        self.runner.run("vn_pmi", NOW)
        releases = self.state()["releases"]
        self.collector.failure = CollectionError("ACCESS_DENIED", "Public source returned HTTP 403")
        self.assertEqual(self.runner.run("vn_pmi", NOW + timedelta(hours=1), force=True), "ERROR")
        self.assertEqual(self.state()["releases"], releases)
        self.assertEqual(self.state()["consecutiveFailures"], 1)
        with (self.root / "vn_pmi/.lock").open("a") as lock:
            fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
            self.assertEqual(self.runner.run("vn_pmi", NOW, force=True), "BUSY")
        self.assertFalse(list(self.root.rglob(".pending-*")))

    def test_future_collector_can_use_interval_and_own_target(self):
        class Daily(FakeCollector):
            schedule = IntervalSchedule(timedelta(hours=2))
            target = staticmethod(lambda now: now.astimezone(ZONE).strftime("%Y-%m-%d"))
        daily = Daily()
        runner = Runner(self.root, {"daily": daily})
        self.assertEqual(runner.run("daily", NOW), "READY")
        state = json.loads((self.root / "daily/manifest.json").read_text())
        self.assertEqual(state["targetPeriod"], "2026-10-01")
        self.assertEqual(state["nextDue"], "2026-10-01T03:00:00Z")

    def test_scrape_exports_immutable_official_pdf(self):
        class Client:
            def get(self, url):
                if "PressReleases?" in url:
                    return ('<a href="' + RELEASE + '">S&amp;P Global Vietnam Manufacturing PMI</a><a href="https://evil.test/report">Vietnam Manufacturing</a>').encode()
                if url != RELEASE:
                    raise AssertionError("unexpected document source")
                return b"%PDF-fixture"
        collector = PMICollector(client_factory=Client, extractor=lambda path: TEXT)
        result = collector.collect(self.root, "2026-09", NOW)
        path = Path(result["releases"][0]["path"])
        self.assertEqual(path.read_bytes(), b"%PDF-fixture")
        self.assertEqual(result["releases"][0]["sourceUrl"], RELEASE)
        self.assertTrue((self.root / "index.html").exists())
        with self.assertRaises(NotPublished):
            collector.collect(self.root, "2026-10", NOW)

    def test_http_403_not_retried_and_messages_omit_urls(self):
        calls, sleeps = [], []
        def deny(request, timeout):
            calls.append(request.full_url)
            raise urllib.error.HTTPError(request.full_url, 403, "Forbidden", {}, io.BytesIO(b"denied"))
        client = HTTPClient(opener=deny, sleep=sleeps.append)
        with self.assertRaises(CollectionError) as ctx:
            client.get("https://source.test/?key=private")
        self.assertEqual(ctx.exception.kind, "ACCESS_DENIED")
        self.assertNotIn("private", str(ctx.exception))
        self.assertEqual(len(calls), 1)
        self.assertEqual(sleeps, [])

    def test_retry_after_deferred_without_blocking_worker(self):
        sleeps = []
        def limited(request, timeout):
            raise urllib.error.HTTPError(request.full_url, 429, "Limited", {"Retry-After": "3600"}, io.BytesIO(b""))
        with self.assertRaises(CollectionError) as ctx:
            HTTPClient(opener=limited, sleep=sleeps.append).get("https://source.test/")
        self.assertEqual(ctx.exception.retry_after, 3600)
        self.assertEqual(sleeps, [])


if __name__ == "__main__":
    unittest.main()
