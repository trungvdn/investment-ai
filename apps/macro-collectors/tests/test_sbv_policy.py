import unittest
from datetime import date
from macro_collectors.sbv_policy import collect, parse_table, parse_release, CODES, INTERBANK, RELEASE
from macro_collectors.http import CollectionError
from macro_collectors.production import validated
from test_pilot import PostgresPilotFixture
import os

TODAY = date(2026, 10, 27)
LABELS = ['Lãi suất tái cấp vốn', 'Lãi suất tái chiết khấu', 'Lãi suất cho vay qua đêm']

def table():
    return ('<table>' + ''.join(f'<tr><td>{label}</td><td>{value}</td><td>19/06/2023</td></tr>' for label, value in zip(LABELS, ['4,5', '3,0', '5,0'])) + '</table>').encode()

def release():
    return ('<article>Quyết định 1123/QĐ-NHNN có hiệu lực từ ngày 19/6/2023. ' + '; '.join(f'{label} giảm từ mức 6,0%/năm xuống {value}%/năm' for label, value in zip(LABELS, ['4,5', '3,0', '5,0'])) + '</article>').encode()

def config():
    return [dict(id=code, indicator_code=code, series_id=code, frequency='EVENT', unit='%', priority=100, extractor_key='sbv_policy_v1') for code in CODES.values()]

class PolicyTests(unittest.TestCase):
    def test_three_rates_and_effective_date_not_collection_date(self):
        rows = parse_table(table(), TODAY)
        self.assertEqual([r['value'] for r in rows], [4.5, 3, 5])
        self.assertTrue(all(r['period'] == '2023-06-19' for r in rows))
        self.assertEqual([r['value'] for r in parse_release(release(), TODAY)], [4.5, 3, 5])

    def test_fail_closed_for_interbank_future_ambiguous_and_incomplete(self):
        for body in [b'<table><tr><td>Qua dem</td><td>4,31</td></tr></table>', table().replace(b'19/06/2023', b'19/06/2027'), table().replace(b'19/06/2023', b'19/06/2023 20/06/2023'), table().replace(b'4,5', b'NaN')]:
            with self.assertRaises((CollectionError, ValueError)):
                parse_table(body, TODAY)
        with self.assertRaises(CollectionError):
            parse_release(release().replace(b'1123/', b'9999/'), TODAY)

    def test_current_success_and_403_fallback_is_partial_not_current(self):
        class Client:
            def __init__(self, denied): self.denied = denied
            def get(self, url):
                if url == RELEASE: return release()
                if self.denied: raise CollectionError('ACCESS_DENIED', '403')
                return table()
        dataset = {'id': 'sbv_policy_rates', 'url': INTERBANK}
        _, rows, notes = collect(Client(False), dataset, config(), TODAY)
        self.assertEqual(len(validated(rows, config(), TODAY)), 3)
        self.assertFalse(notes['warnings'])
        _, rows, notes = collect(Client(True), dataset, config(), TODAY)
        self.assertEqual(len(rows), 3)
        self.assertTrue(notes['warnings'])
        self.assertTrue(all(r['source_url'] == RELEASE and r['priority'] == 50 for r in rows))

@unittest.skipUnless(os.getenv('MACRO_TEST_DATABASE_URL'), 'requires CI PostgreSQL')
class PolicyPersistenceTests(PostgresPilotFixture):
    def test_repeat_ingestion_writes_three_events_without_duplicates(self):
        from macro_collectors.history import ingest
        class Client:
            def get(self, url): return table()
        for _ in range(2):
            result = ingest(self.db, 'sbv_policy_rates', force=True, client_factory=Client)
            self.assertEqual(result['status'], 'WRITTEN')
        rows = self.db.execute("SELECT * FROM macro_observations WHERE indicator_code LIKE 'VN_SBV_%'").fetchall()
        self.assertEqual(len(rows), 3)
        self.assertTrue(all(r['period'] == '2023-06-19' for r in rows))
        self.assertEqual(self.db.execute('SELECT count(*) AS n FROM macro_raw_documents').fetchone()['n'], 1)
