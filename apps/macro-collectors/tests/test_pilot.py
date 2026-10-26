import copy
import hashlib
import json
import os
import subprocess
import tempfile
import unittest
import uuid
from concurrent.futures import ThreadPoolExecutor
from datetime import date, datetime
from pathlib import Path

from macro_collectors.core import ZONE
from macro_collectors.http import CollectionError
from macro_collectors.sdmx import compare, parse, VERSION


def fixture(series="PCPI_IX", periods=36, lag=1):
    today = datetime.now(ZONE).date()
    end = today.year * 12 + today.month - 1 - lag
    rows = []
    for i in range(periods):
        year, month = divmod(end - periods + i + 1, 12)
        rows.append(f'<Obs TIME_PERIOD="{year:04d}-{month+1:02d}" OBS_VALUE="{100+i}"/>')
    return ('<Message><DataSet><Series REF_AREA="VN" INDICATOR="' + series + '" FREQ="M" COUNTERPART_AREA="_Z" UNIT_MULT="0" BASE_PER="2019">' + ''.join(rows) + '</Series></DataSet></Message>').encode()


def binding():
    return {"id":"nso_cpi_yoy","indicator_code":"VN_CPI_YOY","series_id":"PCPI_IX","unit":"% YoY","frequency":"MONTHLY","comparison_basis":"YoY","priority":50,"source_url":"https://nsdp.nso.gov.vn/GSO-chung/SDMXFiles/GSO/CPIVNM.xml","extractor_key":VERSION,"transform_key":"yoy_12m","validation":{"ref_area":"VN","counterpart":"_Z","frequency":"M","unit_mult":"0","tolerance":1e-8,"max_age_days":60}}


class ParserTests(unittest.TestCase):
    def test_same_series_yoy_and_contract_validation(self):
        config = [binding()]
        body = fixture()
        rows = parse(body, config, datetime.now(ZONE).date())
        self.assertEqual(len(rows),24)
        self.assertAlmostEqual(rows[0]["value"],12)
        for bad in (body.replace(b'FREQ="M"',b'FREQ="Q"'),body.replace(b'UNIT_MULT="0"',b'UNIT_MULT="6"'),body.replace(b'OBS_VALUE="100"',b'OBS_VALUE="NaN"'),body.replace(b'<Obs ',b'<Obs TIME_PERIOD="bad" ',1),b'<!DOCTYPE x>'+body):
            with self.assertRaises(CollectionError):
                parse(bad,config,datetime.now(ZONE).date())

    def test_comparison_exposes_values_gaps_metadata_and_freshness(self):
        config=[binding()];rows=parse(fixture(),config,datetime.now(ZONE).date());reference=copy.deepcopy(rows)
        result=compare(rows,reference,config,datetime.now(ZONE).date())
        self.assertEqual(result["matched"],24);self.assertFalse(result["behind"])
        reference[0]["value"]+=0.1;reference[1]["unit"]="wrong";reference.pop()
        result=compare(rows,reference,config,datetime.now(ZONE).date())
        self.assertEqual({d["reason"] for d in result["discrepancies"]},{"VALUE_MISMATCH","METADATA_MISMATCH","MISSING_GO"})
        old=parse(fixture(lag=4),config,datetime.now(ZONE).date());self.assertTrue(compare(old,old,config,datetime.now(ZONE).date())["stale"])


@unittest.skipUnless(os.getenv("MACRO_TEST_DATABASE_URL"),"requires CI PostgreSQL")
class PostgresPilotFixture(unittest.TestCase):
    def setUp(self):
        import psycopg
        from psycopg import sql
        from psycopg.rows import dict_row
        self.psycopg,self.sql,self.dict_row=psycopg,sql,dict_row
        self.dsn=os.environ["MACRO_TEST_DATABASE_URL"]
        self.schema="pilot_"+uuid.uuid4().hex
        self.admin=psycopg.connect(self.dsn,autocommit=True)
        self.admin.execute(sql.SQL("CREATE SCHEMA {} ").format(sql.Identifier(self.schema)))
        self.db=self.connection()
        migrations=Path(__file__).resolve().parents[2]/"macro-api"/"migrations"
        for name in ("001_macro.sql","002_operations.sql","003_ingestion_pilot.sql","004_cpi_reports.sql"):
            self.db.execute((migrations/name).read_text(),prepare=False)
        self.db.execute("UPDATE macro_datasets SET enabled=false WHERE id='gso_cpi_reports'")
        self.tmp=tempfile.TemporaryDirectory()
        self.binary=os.environ.get("SDMX_REPLAY_BIN")
        if not self.binary:
            self.fail("SDMX_REPLAY_BIN is required for real Go/Python parity integration tests")

    def connection(self):
        return self.psycopg.connect(self.dsn,autocommit=True,row_factory=self.dict_row,options=f"-c search_path={self.schema} -c statement_timeout=15000")

    def tearDown(self):
        self.db.close();self.admin.execute(self.sql.SQL("DROP SCHEMA {} CASCADE").format(self.sql.Identifier(self.schema)));self.admin.close();self.tmp.cleanup()

    def baseline(self,dataset="nso_cpi",lag=1):
        series="PCPI_IX" if dataset=="nso_cpi" else "AIP_ISIC4_IX"
        body=fixture(series,lag=lag);path=Path(self.tmp.name)/(dataset+".xml");path.write_bytes(body)
        values=json.loads(subprocess.check_output([self.binary,"-dataset",dataset,"-file",str(path)]))
        metadata=self.db.execute("SELECT * FROM macro_datasets WHERE id=%s",(dataset,)).fetchone()
        raw=self.db.execute("INSERT INTO macro_raw_documents(source,source_url,content_type,checksum,payload,first_fetched_at,last_fetched_at) VALUES('NSO',%s,'application/xml',%s,%s,now(),now()) RETURNING id",(metadata["url"],hashlib.sha256(body).hexdigest(),body)).fetchone()["id"]
        for row in values:
            self.db.execute("INSERT INTO macro_observations(indicator_code,period,period_end,value,unit,source,source_url,series_id,method,priority,first_seen_at,collected_at,revision,raw_id) VALUES(%s,%s,%s,%s,%s,'NSO',%s,%s,%s,50,now(),now(),1,%s)", (row["indicatorCode"],row["period"],row["periodEnd"][:10],row["value"],row["unit"],row["sourceUrl"],row["sourceSeriesId"],row["method"],raw))
        self.db.execute("INSERT INTO macro_attempts(indicator_code,provider_id,checked_at,success,message) VALUES(%s,%s,now(),true,'Go fixture replay')",(values[0]["indicatorCode"],metadata["go_provider_id"]))
        return body

    def execute(self,body,dataset="nso_cpi"):
        from macro_collectors.pilot import run
        class Client:
            def get(self,url):return body
        return run(self.db,dataset,True,Client)[0]["status"]



class DatabasePilotTests(PostgresPilotFixture):
    def test_real_go_parity_shadow_only_and_repeated_runs(self):
        from macro_collectors.pilot import report
        body=self.baseline();before=self.db.execute("SELECT count(*) AS n,max(collected_at) AS ts FROM macro_observations").fetchone()
        self.assertEqual(self.execute(body),"MATCH");self.assertEqual(self.execute(body),"MATCH")
        self.assertEqual(self.db.execute("SELECT count(*) AS n,max(collected_at) AS ts FROM macro_observations").fetchone(),before)
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_pilot_raw").fetchone()["n"],1)
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_pilot_runs").fetchone()["n"],2)
        self.assertFalse(report(self.db)[0]["eligibleForReview"])

    def test_vintages_and_incorrect_reference_are_not_false_passes(self):
        body=self.baseline();self.assertEqual(self.execute(body.replace(b'OBS_VALUE="135"',b'OBS_VALUE="136"')),"VINTAGE_MISMATCH")
        self.db.execute("UPDATE macro_observations SET value=value+1 WHERE period=(SELECT max(period) FROM macro_observations)")
        self.assertEqual(self.execute(body),"MISMATCH")
        self.assertEqual(self.db.execute("SELECT report FROM macro_pilot_runs ORDER BY id DESC LIMIT 1").fetchone()["report"]["discrepancies"][0]["reason"],"VALUE_MISMATCH")

    def test_eligibility_needs_distinct_days_and_never_changes_owner(self):
        from macro_collectors.pilot import report
        body=self.baseline()
        for _ in range(3):self.assertEqual(self.execute(body),"MATCH")
        self.assertFalse(report(self.db)[0]["eligibleForReview"])
        ids=[r["id"] for r in self.db.execute("SELECT id FROM macro_pilot_runs ORDER BY id").fetchall()]
        for days,run_id in zip((2,1,0),ids):self.db.execute("UPDATE macro_pilot_runs SET checked_at=now()-(%s*interval '1 day') WHERE id=%s",(days,run_id))
        self.assertTrue(report(self.db)[0]["eligibleForReview"])
        self.assertEqual(report(self.db)[0]["owner"],"go")
        self.db.execute("UPDATE macro_indicator_bindings SET validation=jsonb_set(validation,'{tolerance}','1e-9') WHERE dataset_id='nso_cpi'")
        self.assertEqual(self.execute(body),"MATCH");self.assertFalse(report(self.db)[0]["eligibleForReview"])

    def test_leases_fence_parallel_force_runs_and_owner_guard(self):
        from macro_collectors.pilot import claim,run,persist,bindings
        def attempt(_):
            with self.connection() as db:return claim(db,"nso_cpi",True)
        with ThreadPoolExecutor(max_workers=5) as pool:tokens=list(pool.map(attempt,range(5)))
        winner=[t for t in tokens if t];self.assertEqual(len(winner),1)
        self.db.execute("UPDATE macro_pilot_state SET lease_until=now()-interval '1 minute'")
        replacement=claim(self.db,"nso_cpi",True);self.assertTrue(replacement)
        dataset=self.db.execute("SELECT * FROM macro_datasets WHERE id='nso_cpi'").fetchone()
        with self.assertRaises(RuntimeError):persist(self.db,dataset,winner[0],None,[],bindings(self.db,"nso_cpi"))
        self.db.execute("UPDATE macro_datasets SET ingestion_owner='python' WHERE id='nso_cpi'")
        with self.assertRaises(ValueError):run(self.db,"nso_cpi",True)

    def test_freshness_and_all_dataset_selection(self):
        from macro_collectors.pilot import run,selected_datasets
        self.assertEqual(selected_datasets(self.db,indicator="VN_CPI_YOY"),["nso_cpi"])
        bodies={"CPIVNM.xml":self.baseline("nso_cpi"),"IIPVNM.xml":self.baseline("nso_iip")}
        class Client:
            def get(self,url):return bodies[url.rsplit('/',1)[1]]
        self.assertEqual([r["status"] for r in run(self.db,"all",True,Client)],["MATCH","MATCH"])


if __name__=="__main__":unittest.main()
