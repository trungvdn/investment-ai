import json
import os
import subprocess
import unittest
from datetime import date,datetime
from pathlib import Path
from macro_collectors.core import ZONE
from macro_collectors.nso_history import parse_ppi,PPI_SERIES
from macro_collectors.http import CollectionError
from macro_collectors.sources import expected_periods
from test_pilot import PostgresPilotFixture,fixture


def ppi():
    return ('<Message><DataSet><Series INDICATOR="'+PPI_SERIES+'" REF_AREA="VN" COUNTERPART_AREA="_Z" FREQ="Q" UNIT_MULT="0" BASE_PER="_Z">'
            '<Obs TIME_PERIOD="2025-Q1" OBS_VALUE="104.75"/><Obs TIME_PERIOD="2025-Q2" OBS_VALUE="103.25"/>'
            '</Series></DataSet></Message>').encode()


def binding():
    return {"id":"nso_ppi_yoy_history","indicator_code":"VN_PPI_YOY","unit":"% YoY","frequency":"QUARTERLY",
            "comparison_basis":"YoY","series_id":PPI_SERIES,"extractor_key":"nso_ppi_history_v1","transform_key":"same_quarter_index_minus_100"}


class NSOHistoryTests(unittest.TestCase):
    def test_ppi_is_quarterly_index_minus_100_not_yoy_again(self):
        rows=parse_ppi(ppi(),[binding()],date(2026,10,27))
        self.assertEqual([r["value"] for r in rows],[4.75,3.25])
        self.assertEqual(rows[0]["period_end"],date(2025,3,31))
        for bad in (ppi().replace(b'BASE_PER="_Z"',b'BASE_PER="2019"'),ppi().replace(b'FREQ="Q"',b'FREQ="M"'),
                    ppi().replace(b'104.75',b'NaN'),ppi().replace(b'2025-Q1',b'2027-Q1'),b'<!DOCTYPE x>'+ppi()):
            with self.assertRaises(CollectionError):parse_ppi(bad,[binding()],date(2026,10,27))

    def test_frequency_coverage_does_not_invent_monthly_ppi(self):
        self.assertEqual(expected_periods("QUARTERLY","2025-01","2025-09"),["2025-Q1","2025-Q2","2025-Q3"])
        self.assertEqual(expected_periods("MONTHLY","2025-01","2025-03"),["2025-01","2025-02","2025-03"])


@unittest.skipUnless(os.getenv("MACRO_TEST_DATABASE_URL"),"requires CI PostgreSQL")
class NSOHistoryDatabaseTests(PostgresPilotFixture):
    def test_ppi_matches_actual_go_parser_and_reruns_are_idempotent(self):
        from macro_collectors.history import ingest
        body=ppi();path=Path(self.tmp.name)/"ppi.xml";path.write_bytes(body)
        go=json.loads(subprocess.check_output([self.binary,"-dataset","nso_ppi_history","-file",str(path)]))
        python=parse_ppi(body,[binding()],datetime.now(ZONE).date())
        self.assertEqual([r["value"] for r in python],[r["value"] for r in go])
        self.assertEqual([r["period"] for r in python],[r["period"] for r in go])
        class Client:
            def get(self,url):return body
        for _ in range(2):
            self.assertEqual(ingest(self.db,"nso_ppi_history",True,client_factory=Client)["importedObservations"],2)
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_source_values").fetchone()["n"],2)
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_observations").fetchone()["n"],0)

    def test_iip_uses_full_prior_year_levels_and_existing_go_parity(self):
        from macro_collectors.history import ingest
        body=self.baseline("nso_iip")
        before=self.db.execute("SELECT count(*) AS n FROM macro_observations").fetchone()["n"]
        class Client:
            def get(self,url):return body
        result=ingest(self.db,"nso_iip",True,first="2026-01",last="2026-09",client_factory=Client)
        self.assertEqual(result["status"],"MATCH")
        self.assertEqual(result["importedObservations"],24)
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_observations").fetchone()["n"],before)
