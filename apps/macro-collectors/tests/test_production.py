import copy
import os
import unittest
from concurrent.futures import ThreadPoolExecutor
from datetime import date,datetime,timezone
from macro_collectors.http import CollectionError
from macro_collectors.production import claim,persist,validated,coverage
from test_pilot import PostgresPilotFixture


@unittest.skipUnless(os.getenv("MACRO_TEST_DATABASE_URL"),"requires CI PostgreSQL")
class ProductionWriterTests(PostgresPilotFixture):
    def setup_rows(self):
        from macro_collectors.pilot import bindings
        dataset=self.db.execute("SELECT * FROM macro_datasets WHERE id='mbs_pmi_history'").fetchone()
        config=bindings(self.db,dataset["id"])
        b=config[0]
        row={"binding_id":b["id"],"indicator_code":b["indicator_code"],"period":"2025-01","period_end":date(2025,1,31),
             "value":49.3,"unit":b["unit"],"series_id":b["series_id"],"method":"PUBLISHED_MANUFACTURING_PMI; SECONDARY=MBS",
             "priority":b["priority"],"source_url":"https://www.mbs.com.vn/fixture-report/","published_at":None}
        return dataset,config,[row]

    def save(self,dataset,config,rows,body=b"fixture",failure=None,notes=None):
        token=claim(self.db,dataset["go_provider_id"],True)
        return persist(self.db,dataset,token,body,rows,config,failure,notes)

    def test_backfill_and_latest_reuse_same_production_value_and_no_staging(self):
        dataset,config,rows=self.setup_rows()
        first=self.save(dataset,config,rows)
        second=self.save(dataset,config,rows)
        self.assertEqual((first["inserted"],second["inserted"],second["unchanged"]),(1,0,1))
        for table in ("macro_source_values","macro_pilot_raw","macro_pilot_runs","macro_pilot_observations","macro_canonical_batches"):
            self.assertEqual(self.db.execute("SELECT count(*) AS n FROM "+table).fetchone()["n"],0)
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_observations").fetchone()["n"],1)
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_raw_documents").fetchone()["n"],1)
        self.assertTrue(self.db.execute("SELECT success FROM macro_attempts").fetchone()["success"])
        report=coverage(self.db,[dataset["id"]],"2025-01","2025-03")[0]
        self.assertEqual(report["table"],"macro_observations")
        self.assertEqual(report["missingRequestedPeriods"],["2025-02","2025-03"])

    def test_metadata_changes_do_not_duplicate_but_real_revisions_are_kept(self):
        dataset,config,rows=self.setup_rows()
        self.save(dataset,config,rows)
        changed=copy.deepcopy(rows)
        changed[0].update(source_url="https://www.mbs.com.vn/reissued/",published_at=datetime.now(timezone.utc),
                          evidence={"paragraphIndex":99},method=rows[0]["method"]+"; ARTICLE_ID=new; PARSER=v2")
        self.assertEqual(self.save(dataset,config,changed,body=b"reissued")["unchanged"],1)
        changed[0]["value"]=50.1
        self.assertEqual(self.save(dataset,config,changed)["revised"],1)
        self.save(dataset,config,rows)
        result=self.db.execute("SELECT revision,value FROM macro_observations ORDER BY revision").fetchall()
        self.assertEqual([r["revision"] for r in result],[1,2,3])
        self.assertEqual([r["value"] for r in result],[49.3,50.1,49.3])

    def test_partial_and_failed_jobs_keep_valid_rows_and_retry_state(self):
        dataset,config,rows=self.setup_rows()
        good=self.save(dataset,config,rows)
        self.assertEqual(self.save(dataset,config,rows,notes={"warnings":[{"kind":"ARCHIVE_GAP"}]})["status"],"PARTIAL")
        failed=self.save(dataset,config,[],body=b"denied",failure=CollectionError("ACCESS_DENIED","fixture HTTP 403"))
        self.assertEqual(failed["status"],"SOURCE_ERROR")
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_observations").fetchone()["n"],1)
        state=self.db.execute("SELECT * FROM macro_source_state").fetchone()
        self.assertEqual(state["consecutive_failures"],2)
        self.assertIsNotNone(state["last_success"])
        self.assertEqual(state["error_kind"],"ACCESS_DENIED")

    def test_leases_fence_stale_owners(self):
        dataset,config,rows=self.setup_rows()
        provider=dataset["go_provider_id"]
        old=claim(self.db,provider,True)
        self.assertIsNone(claim(self.db,provider,True))
        self.db.execute("UPDATE macro_source_state SET lease_until=now()-interval '1 second' WHERE provider=%s",(provider,))
        new=claim(self.db,provider,True)
        with self.assertRaises(RuntimeError):persist(self.db,dataset,old,b"old",rows,config)
        self.assertEqual(persist(self.db,dataset,new,b"new",rows,config)["inserted"],1)

    def test_shared_sql_writer_deduplicates_concurrent_go_style_calls(self):
        dataset,config,rows=self.setup_rows()
        self.save(dataset,config,rows)
        row=rows[0]
        raw=self.db.execute("SELECT id FROM macro_raw_documents LIMIT 1").fetchone()["id"]
        def write(_):
            connection=self.connection()
            try:
                return connection.execute("SELECT * FROM macro_save_observation_v1(%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)",
                  (row["indicator_code"],row["period"],row["period_end"],50.1,row["unit"],"MBS Research",row["source_url"],
                   row["series_id"],row["method"],row["priority"],None,datetime.now(timezone.utc),raw)).fetchone()
            finally:connection.close()
        with ThreadPoolExecutor(max_workers=6) as pool:
            result=list(pool.map(write,range(12)))
        self.assertEqual(sum(r["inserted"] for r in result),1)
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_observations").fetchone()["n"],2)

    def test_batch_rollback_on_invalid_value_preserves_previous_data(self):
        dataset,config,rows=self.setup_rows()
        row=copy.deepcopy(rows[0]);row.update(period="2025-02",period_end=date(2025,2,28),priority=None)
        token=claim(self.db,dataset["go_provider_id"],True)
        with self.assertRaises(Exception):
            persist(self.db,dataset,token,b"batch",rows+[row],config)
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_observations").fetchone()["n"],0)
        self.assertEqual(self.db.execute("SELECT count(*) AS n FROM macro_raw_documents").fetchone()["n"],0)

    def test_validation_rejects_wrong_definition_and_conflicting_batch(self):
        _,config,rows=self.setup_rows()
        wrong=copy.deepcopy(rows);wrong[0]["unit"]="% YoY"
        with self.assertRaises(CollectionError):validated(wrong,config,date(2026,10,27))
        bad=copy.deepcopy(rows[0]);bad["value"]=1
        with self.assertRaises(CollectionError):validated(rows+[bad],config,date(2026,10,27))
