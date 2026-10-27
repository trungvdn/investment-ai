import copy
import os
import unittest
from datetime import date
from unittest.mock import patch
from macro_collectors.sbv_deposit import collect, parse_report, INDEX, VERSION
from macro_collectors.production import validated
from macro_collectors.http import CollectionError
from test_pilot import PostgresPilotFixture

TODAY = date(2026, 10, 27)
REPORT = '''DIỄN BIẾN LÃI SUẤT CỦA TỔ CHỨC TÍN DỤNG ĐỐI VỚI KHÁCH HÀNG THÁNG 9/2026
1. Lãi suất tiền gửi:
- Lãi suất VND: Lãi suất tiền gửi bằng VND bình quân của nhóm NHTM nhà nước và nhóm NHTM cổ phần đối với các khoản tiền gửi mới và cũ ở mức
0,1-0,2%/năm đối với tiền gửi không kỳ hạn; 4,1-4,6%/năm đối với tiền gửi có kỳ hạn từ 1 tháng đến dưới 6 tháng;
6,6-8,1%/năm đối với tiền gửi có kỳ hạn từ 6 tháng đến 12 tháng;
6,2-7,8%/năm đối với tiền gửi có kỳ hạn từ trên 12 tháng đến 24 tháng.
- Lãi suất USD: Lãi suất tiền gửi USD 0%/năm.
2. Lãi suất cho vay: 8,4-10,9%/năm.'''
ARTICLE = 'https://sbv.gov.vn/vi/w/report-9-2026'
PDF = 'https://sbv.gov.vn/documents/report.pdf/id'

def binding():
    return dict(id='sbv_deposit_6_12m', indicator_code='VN_DEPOSIT_RATE', unit='%', frequency='MONTHLY',
                comparison_basis='RANGE', series_id='SBV_VND_DEPOSIT_6_12M_RANGE', priority=100, extractor_key=VERSION)

class Client:
    def get(self, url):
        if url == INDEX:
            return f'<a href="{ARTICLE}">Diễn biến lãi suất của tổ chức tín dụng đối với khách hàng tháng 9/2026</a>'.encode()
        if url == ARTICLE:
            return f'<a href="{PDF}">Diễn biến lãi suất tháng 9/2026</a>'.encode()
        if url == PDF: return b'%PDF-fixture'
        raise AssertionError(url)

class DepositTests(unittest.TestCase):
    def test_range_month_tenor_population_and_currency(self):
        row = parse_report(REPORT, PDF, binding(), TODAY)
        self.assertEqual((row['period'], row['period_end'], row['value'], row['upper_value']), ('2026-09', date(2026, 9, 30), 6.6, 8.1))
        self.assertEqual(validated([row], [binding()], TODAY), [row])
        for text in [REPORT.replace('THÁNG 9/2026', 'THÁNG 11/2026'), REPORT.replace('6,6-8,1', '8,1-6,6'), REPORT.replace('6,6-8,1', 'NaN-8,1'), REPORT.replace('mới và cũ', 'mới'), REPORT.replace('bằng VND', 'bằng USD'), REPORT.replace('6 tháng đến 12 tháng', '6 tháng đến dưới 12 tháng')]:
            with self.assertRaises(CollectionError): parse_report(text, PDF, binding(), TODAY)

    def test_discovery_pdf_and_explicit_source_validation(self):
        with patch.dict(os.environ, {'SBV_DEPOSIT_REPORT_URLS': ''}):
            body, rows, notes = collect(Client(), {'url': INDEX}, [binding()], TODAY, text_extractor=lambda _: REPORT)
        self.assertEqual(len(rows), 1)
        self.assertEqual(rows[0]['source_url'], PDF)
        self.assertIn(b'bodyBase64', body)
        self.assertFalse(notes['warnings'])
        with patch.dict(os.environ, {'SBV_DEPOSIT_REPORT_URLS': 'https://evil.test/report.pdf'}):
            with self.assertRaises(CollectionError): collect(Client(), {'url': INDEX}, [binding()], TODAY)

    def test_validation_rejects_missing_upper_and_conflicting_upper(self):
        row = parse_report(REPORT, PDF, binding(), TODAY)
        for upper in (None, True, float('nan'), 6.5, 101):
            with self.assertRaises(CollectionError): validated([{**row, 'upper_value': upper}], [binding()], TODAY)
        with self.assertRaises(CollectionError): validated([row, {**row, 'upper_value': 8.2}], [binding()], TODAY)

@unittest.skipUnless(os.getenv('MACRO_TEST_DATABASE_URL'), 'requires CI PostgreSQL')
class DepositPersistenceTests(PostgresPilotFixture):
    def test_reruns_and_upper_only_revision(self):
        from macro_collectors.pilot import bindings
        from macro_collectors.production import claim, persist
        dataset = self.db.execute("SELECT * FROM macro_datasets WHERE id='sbv_deposit_rates'").fetchone()
        config = bindings(self.db, dataset['id'])
        row = parse_report(REPORT, PDF, config[0], TODAY)
        def save(value):
            return persist(self.db, dataset, claim(self.db, dataset['go_provider_id'], True), b'fixture', [value], config)
        self.assertEqual(save(row)['inserted'], 1)
        self.assertEqual(save(row)['unchanged'], 1)
        revised = {**row, 'upper_value': 8.2}
        self.assertEqual(save(revised)['revised'], 1)
        self.assertEqual(save(revised)['unchanged'], 1)
        stored = self.db.execute('SELECT value,upper_value,revision FROM macro_observations ORDER BY revision').fetchall()
        self.assertEqual([(r['value'], r['upper_value'], r['revision']) for r in stored], [(6.6, 8.1, 1), (6.6, 8.2, 2)])
        self.assertEqual(self.db.execute('SELECT count(*) AS n FROM macro_raw_documents').fetchone()['n'], 1)
