import copy
import json
import os
import unittest
from datetime import date
from unittest.mock import patch
from macro_collectors.sbv_history import parse_credit,parse_credit_html,parse_interbank,collect,ARTICLES,CREDIT
from macro_collectors.fred_history import collect as fred_collect
from macro_collectors.http import CollectionError
from test_pilot import PostgresPilotFixture


def credit_article():
    return {"status":0,"ddmStructureId":10034332,"articleId":"fixture","title":"Dư nợ tín dụng tháng 01 năm 2025",
            "datePublished":"2025-02-15T00:00:00Z","version":1,"fields":{"Date62813077":"2025-01-31",
            "soDuTongCong":"1.234.567,89","tocDoTangGiamTongCong":"1,25"}}


def config():
    return [{"id":"sbv_credit_ytd_history","indicator_code":"VN_CREDIT_GROWTH","unit":"% YTD","series_id":"SBV_TOTAL_CREDIT_YTD",
             "priority":100,"extractor_key":"sbv_credit_history_v1"}]


def on_html(star=False):
    rate="3,25 (*)" if star else "3,25"
    note="Tham chiếu doanh số và lãi suất ngày 22/10/2026" if star else ""
    return ('<div class="bi01-wrap">Ngày áp dụng: 23/10/2026<table><tr><th>Thời hạn</th><th>Lãi suất BQ liên Ngân hàng (% năm)</th></tr>'
            '<tr><td>Qua đêm</td><td>'+rate+'</td></tr><tr><td>1 Tuần</td><td>7,5</td></tr></table>'+note+'</div>').encode()


class SBVHistoryTests(unittest.TestCase):
    def test_national_month_end_ytd_nested_fields_and_revision(self):
        article=credit_article()
        article["fields"]["tocDoTangGiamTongCong"]={"content":"1,25"}
        rows,warnings=parse_credit({"articles":[article]},config(),date(2026,10,27))
        self.assertEqual(rows[0]["value"],1.25)
        self.assertEqual(rows[0]["period"],"2025-01")
        self.assertFalse(warnings)
        bad=copy.deepcopy(article);bad["fields"]["Date62813077"]="2025-01-20"
        self.assertFalse(parse_credit({"articles":[bad]},config(),date(2026,10,27))[0])
        bad=copy.deepcopy(article);bad["title"]="Dư nợ tín dụng Hà Nội tháng 01 năm 2025"
        self.assertFalse(parse_credit({"articles":[bad]},config(),date(2026,10,27))[0])
        bad=copy.deepcopy(article);bad["fields"]["tocDoTangGiamTongCong"]="1,50"
        with self.assertRaises(CollectionError):parse_credit({"articles":[article,bad]},config(),date(2026,10,27))

    def test_on_uses_row_reference_date_and_rejects_future_or_wrong_currency(self):
        self.assertEqual(parse_interbank(on_html(),date(2026,10,27))[0]["period"],"2026-10-23")
        self.assertEqual(parse_interbank(on_html(True),date(2026,10,27))[0]["period"],"2026-10-22")
        self.assertFalse(parse_interbank(on_html().replace(b"Th",b"USD Th",1),date(2026,10,27)))
        with self.assertRaises(CollectionError):parse_interbank(on_html(),date(2026,10,1))

    def test_credit_page_parameters_and_pagination_validation(self):
        urls=[]
        class Client:
            def get(self,url):
                urls.append(url)
                return json.dumps({"page":2,"lastPage":3,"totalCount":450,"articles":[credit_article()]}).encode()
        _,rows,notes=collect(Client(),{"id":"sbv_credit_history","url":ARTICLES},config(),date(2026,10,27),page=2)
        self.assertEqual(rows[0]["value"],1.25)
        self.assertIn("page=2",urls[0]);self.assertIn("contentStructureId=10034332",urls[0])
        self.assertEqual(notes["lastPage"],3)

    def test_fred_complete_series_and_no_secret_in_stored_provenance(self):
        cfg=[{"id":"fred_effr_full","indicator_code":"US_EFFR","unit":"%","series_id":"DFF"}]
        class Client:
            def get(self,url):
                return json.dumps({"count":2,"observations":[{"date":"2020-01-01","value":"1.5"},{"date":"2020-01-02","value":"."}]}).encode()
        with patch.dict(os.environ,{"FRED_API_KEY":"fixture-secret"}):
            raw,rows,notes=fred_collect(Client(),{"id":"fred_effr_history","url":"https://fred.stlouisfed.org/series/DFF"},cfg,date(2026,10,27))
        self.assertEqual(len(rows),1)
        self.assertNotIn("fixture-secret",str(rows)+str(notes)+raw.decode())


@unittest.skipUnless(os.getenv("MACRO_TEST_DATABASE_URL"),"requires CI PostgreSQL")
class SBVHistoryDatabaseTests(PostgresPilotFixture):
    def test_credit_rerun_keeps_one_production_observation(self):
        from macro_collectors.history import ingest
        class Client:
            def get(self,url):
                return json.dumps({"page":1,"lastPage":1,"totalCount":1,"articles":[credit_article()]}).encode()
        for _ in range(2):
            result=ingest(self.db,"sbv_credit_history",True,client_factory=Client)
            self.assertEqual(result["importedObservations"],1)
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_source_values").fetchone()["n"],0)
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_observations").fetchone()["n"],1)
