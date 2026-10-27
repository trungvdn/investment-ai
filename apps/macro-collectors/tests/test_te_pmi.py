import os
import unittest
from datetime import date

from macro_collectors.http import CollectionError
from macro_collectors.te_pmi import extract, URL
from test_pilot import PostgresPilotFixture

# Synthetic values, not a copy of the live source report.
SUMMARY = 'Manufacturing PMI in Vietnam decreased to 50.70 points in September from 51.20 points in August of 2026.'


def page(summary=SUMMARY):
    return ('<h1>Vietnam Manufacturing PMI</h1><h2>' + summary + '</h2>'
            '<h3>Manufacturing PMI in Vietnam is expected to be 60.00 points next quarter.</h3>'
            '<script>Manufacturing PMI in Vietnam decreased to 90 points in September from 91 points in August of 2026.</script>').encode()


class TradingEconomicsPMITests(unittest.TestCase):
    def test_actual_value_period_and_provenance(self):
        row = extract(page(), date(2026, 10, 26))[0]
        self.assertEqual((row['period'], row['value']), ('2026-09', 50.7))
        self.assertEqual(row['source_url'], URL)
        self.assertEqual(row['evidence']['previousValue'], 51.2)
        self.assertEqual(row['evidence']['originalProducer'], 'S&P Global')
        self.assertIsNone(row['published_at'])

    def test_year_rollover(self):
        row = extract(page(SUMMARY.replace('September', 'January').replace('August of 2026', 'December of 2025')), date(2026, 2, 5))[0]
        self.assertEqual(row['period'], '2026-01')

    def test_layout_drift_forecast_conflict_and_invalid_data_fail_closed(self):
        for body in (page().replace(b'<h2>', b'<h3>').replace(b'</h2>', b'</h3>'),
                     page().replace(b'Vietnam Manufacturing PMI', b'Thailand Manufacturing PMI'),
                     page(SUMMARY.replace('50.70', '101.0')),
                     page(SUMMARY.replace('August', 'July')),
                     page() + b'<h2>' + SUMMARY.replace('50.70', '52.0').encode() + b'</h2>',
                     page(SUMMARY.replace('decreased to', 'is expected to be'))):
            with self.subTest(body=body), self.assertRaises(CollectionError) as caught:
                extract(body, date(2026, 10, 26))
            self.assertEqual(caught.exception.payload, body)
        with self.assertRaises(CollectionError):
            extract(page(), date(2026, 9, 1))


@unittest.skipUnless(os.getenv('MACRO_TEST_DATABASE_URL'), 'requires CI PostgreSQL')
class TradingEconomicsPMIDatabaseTests(PostgresPilotFixture):
    def test_manual_duplicate_runs_keep_production_untouched(self):
        from macro_collectors.pilot import run, report, selected_datasets
        self.db.execute("UPDATE macro_datasets SET enabled=true WHERE id='te_vietnam_pmi'")

        class Client:
            def get(self, url):
                self.assert_url = url
                if url != URL:
                    raise AssertionError(url)
                return page()

        self.assertEqual(selected_datasets(self.db, 'all', 'VN_PMI'), ['te_vietnam_pmi'])
        for _ in range(2):
            result = run(self.db, 'te_vietnam_pmi', True, Client)
            self.assertIn(result[0]['status'], ('CROSS_SOURCE_REVIEW', 'STALE'))
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_pilot_raw").fetchone()['n'], 1)
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_pilot_observations").fetchone()['n'], 2)
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_observations").fetchone()['n'], 0)
        self.assertFalse(report(self.db, 'te_vietnam_pmi')[0]['eligibleForReview'])
