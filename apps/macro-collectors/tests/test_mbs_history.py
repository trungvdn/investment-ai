import os
import unittest
from datetime import date
from macro_collectors.mbs_history import extract_text, collect, INDEX, VERSION
from macro_collectors.http import CollectionError
from test_pilot import PostgresPilotFixture

REPORT="https://www.mbs.com.vn/bao-cao-trien-vong-vi-mo-thang-2-2025/"
PDF="https://www.mbs.com.vn/media/fixture/report.pdf"


def binding():
    return {"id":"mbs_pmi_headline","indicator_code":"VN_PMI","unit":"index","series_id":"MBS_VN_MANUFACTURING_PMI",
            "priority":80,"extractor_key":VERSION,"transform_key":"published_index"}


class MBSHistoryTests(unittest.TestCase):
    def test_explicit_month_value_and_year_rollover(self):
        rows=extract_text("Trong tháng 1, chỉ số PMI giảm về mức 49,3.",2025,2,date(2026,10,27),REPORT)
        self.assertEqual((rows[0]["period"],rows[0]["value"]),("2025-01",49.3))
        rows=extract_text("PMI tháng 12 đạt 51.7 điểm.",2025,1,date(2026,10,27),REPORT)
        self.assertEqual(rows[0]["period"],"2024-12")

    def test_forecasts_chart_numbers_future_and_conflicts(self):
        for text in ("Dự báo PMI tháng 1 đạt 52 điểm.","PMI 40 50 60 2025","PMI tháng 2 đạt 52 điểm.",
                     "PMI tháng 1 năm 2027 đạt 52 điểm.","GDP tháng 1 đạt 8 điểm. PMI ổn định."):
            self.assertFalse(extract_text(text,2025,2,date(2026,10,27),REPORT))
        with self.assertRaises(CollectionError):
            extract_text("PMI tháng 1 đạt 49 điểm. PMI tháng 1 đạt 50 điểm.",2025,2,date(2026,10,27),REPORT)

    def test_discovered_pdf_is_downloaded_and_exact_bytes_are_archived(self):
        import base64,json
        calls=[]
        class Client:
            def get(self,url):
                calls.append(url)
                if url==INDEX:return ('<a href="'+REPORT+'">Báo cáo triển vọng vĩ mô tháng 2/2025</a>').encode()
                if url==REPORT:return ('<a href="'+PDF+'">Tải báo cáo chi tiết</a>').encode()
                if url==PDF:return b"%PDF-fixture"
                raise AssertionError(url)
        raw,rows,notes=collect(Client(),{"url":INDEX},[binding()],date(2026,10,27),
                               text_extractor=lambda b:"PMI tháng 1 năm 2025 đạt 49,3 điểm.")
        self.assertEqual(rows[0]["value"],49.3)
        self.assertEqual(rows[0]["source_url"],PDF)
        self.assertEqual(base64.b64decode(json.loads(raw)["documents"][-1]["pdfBase64"]),b"%PDF-fixture")
        self.assertFalse(notes["warnings"])
        self.assertEqual(calls,[INDEX,REPORT,PDF])

    def test_untrusted_download_host_is_not_requested(self):
        calls=[]
        class Client:
            def get(self,url):
                calls.append(url)
                if url==INDEX:return ('<a href="'+REPORT+'">Báo cáo triển vọng vĩ mô tháng 2/2025</a>').encode()
                return b'<a href="https://evil.example/report.pdf">Download</a>'
        _,rows,notes=collect(Client(),{"url":INDEX},[binding()],date(2026,10,27))
        self.assertFalse(rows)
        self.assertEqual(len(calls),2)
        self.assertTrue(notes["warnings"])


@unittest.skipUnless(os.getenv("MACRO_TEST_DATABASE_URL"),"requires CI PostgreSQL")
class MBSHistoryDatabaseTests(PostgresPilotFixture):
    def test_manual_rerun_stores_one_production_value_without_staging(self):
        from macro_collectors.history import ingest
        class Client:
            def get(self,url):
                if url==INDEX:return ('<a href="'+REPORT+'">Báo cáo triển vọng vĩ mô tháng 2/2025</a>').encode()
                return '<p>PMI tháng 1 năm 2025 đạt 49,3 điểm.</p>'.encode()
        for _ in range(2):
            result=ingest(self.db,"mbs_pmi_history",True,client_factory=Client)
            self.assertEqual(result["importedObservations"],1)
            self.assertNotEqual(result["status"],"SOURCE_ERROR")
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_source_values").fetchone()["n"],0)
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_observations").fetchone()["n"],1)
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_pilot_runs").fetchone()["n"],0)
