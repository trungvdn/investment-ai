import copy
import os
import unittest
from datetime import date
from unittest.mock import patch
from macro_collectors.sbv_deposit import collect, parse_report, INDEX, ARCHIVE, VERSION
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
        if url == ARCHIVE: return b'<html></html>'
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


class ArchiveTests(unittest.TestCase):
    def test_january_to_september_in_three_batches_and_paginated_discovery(self):
        class ArchiveClient:
            def get(self, url):
                if url == INDEX: return b'<html></html>'
                if url in (ARCHIVE, ARCHIVE + '?_publisher_cur=2'):
                    months = range(5, 10) if url == ARCHIVE else range(1, 5)
                    links = ''.join(f'<a href="https://sbv.gov.vn/report/{m}">Diễn biến lãi suất của tổ chức tín dụng đối với khách hàng tháng {m}/2026</a>' for m in months)
                    if url == ARCHIVE: links += '<a href="?_publisher_cur=2">2</a>'
                    return links.encode()
                month = int(url.rsplit('/', 1)[1])
                return ('<h1>Diễn biến lãi suất của tổ chức tín dụng đối với khách hàng tháng '+str(month)+'/2026</h1><p>'+REPORT.replace('THÁNG 9/2026', f'THÁNG {month}/2026')+'</p>').encode()
        output = []
        with patch.dict(os.environ, {'SBV_DEPOSIT_REPORT_URLS': ''}), patch('macro_collectors.sbv_deposit.load_registry', return_value=[]):
            for page in (1, 2, 3):
                _, rows, notes = collect(ArchiveClient(), {'url': INDEX}, [binding()], TODAY, page, '2026-01', '2026-09')
                self.assertFalse(notes['missingMonths'])
                self.assertFalse(notes['warnings'])
                output.extend(rows)
        self.assertEqual(sorted(r['period'] for r in output), [f'2026-{m:02d}' for m in range(1,10)])

    def test_missing_requested_months_are_explicit_not_a_successful_backfill(self):
        with patch.dict(os.environ, {'SBV_DEPOSIT_REPORT_URLS': ''}), patch('macro_collectors.sbv_deposit.load_registry', return_value=[]):
            _, rows, notes = collect(Client(), {'url': INDEX}, [binding()], TODAY, 1, '2026-08', '2026-09', text_extractor=lambda _: REPORT)
        self.assertEqual([r['period'] for r in rows], ['2026-09'])
        self.assertEqual(notes['missingMonths'], ['2026-08'])
        self.assertEqual(notes['warnings'][-1]['kind'], 'MISSING_PERIODS')

    def test_hidden_publisher_can_be_expanded_without_visible_pagination_links(self):
        from urllib.parse import parse_qs, urlsplit
        class HiddenClient:
            def get(self, url):
                if url == INDEX: return b'<html></html>'
                if url == ARCHIVE:
                    return b'<section id="portlet_com_liferay_asset_publisher_web_portlet_AssetPublisherPortlet_INSTANCE_old"><a href="https://sbv.gov.vn/report/9">Dien bien lai suat cua to chuc tin dung doi voi khach hang thang 9/2026</a></section>'
                if url.startswith(ARCHIVE+'?'):
                    query = parse_qs(urlsplit(url).query)
                    self_test.assertEqual(query['_com_liferay_asset_publisher_web_portlet_AssetPublisherPortlet_INSTANCE_old_delta'], ['100'])
                    return b'<a href="https://sbv.gov.vn/report/1">Dien bien lai suat cua to chuc tin dung doi voi khach hang thang 1/2026</a>'
                return ('<h1>Diễn biến lãi suất của tổ chức tín dụng đối với khách hàng tháng 1/2026</h1><p>'+REPORT.replace('THÁNG 9/2026','THÁNG 1/2026')+'</p>').encode()
        self_test = self
        with patch.dict(os.environ, {'SBV_DEPOSIT_REPORT_URLS': ''}), patch('macro_collectors.sbv_deposit.load_registry', return_value=[]):
            _, rows, notes = collect(HiddenClient(), {'url': INDEX}, [binding()], TODAY, 1, '2026-01', '2026-01')
        self.assertEqual([r['period'] for r in rows], ['2026-01'])
        self.assertFalse(notes['missingMonths'])

    def test_zero_matches_still_reports_each_missing_requested_month(self):
        with patch.dict(os.environ, {'SBV_DEPOSIT_REPORT_URLS': ''}), patch('macro_collectors.sbv_deposit.load_registry', return_value=[]):
            _, rows, notes = collect(Client(), {'url': INDEX}, [binding()], TODAY, 1, '2026-01', '2026-04')
        self.assertEqual(rows, [])
        self.assertEqual(notes['missingMonths'], ['2026-01','2026-02','2026-03','2026-04'])

    def test_registered_month_downloads_directly_and_skips_successful_alternate(self):
        registry = [{'period':'2026-01','url':PDF},{'period':'2026-01','url':ARTICLE}]
        calls = []
        class DirectClient:
            def get(self, url):
                calls.append(url)
                if url != PDF: raise AssertionError('Unexpected index/alternate request')
                return b'%PDF-fixture'
        with patch.dict(os.environ, {'SBV_DEPOSIT_REPORT_URLS': ''}), patch('macro_collectors.sbv_deposit.load_registry', return_value=registry):
            _, rows, notes = collect(DirectClient(), {'url': INDEX}, [binding()], TODAY, 1, '2026-01', '2026-01', text_extractor=lambda _: REPORT.replace('THÁNG 9/2026','THÁNG 1/2026'))
        self.assertEqual(calls, [PDF])
        self.assertEqual([r['period'] for r in rows], ['2026-01'])
        self.assertFalse(notes['warnings'])
