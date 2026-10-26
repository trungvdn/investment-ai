import copy
import json
import unittest
from datetime import date, datetime
from macro_collectors.core import ZONE
from macro_collectors.cpi_reports import extract_article,collect,INDEX,ALTERNATE
from macro_collectors.http import CollectionError
from test_pilot import PostgresPilotFixture
import os

URL="https://www.nso.gov.vn/du-lieu-va-so-lieu-thong-ke/2026/10/chi-so-gia-tieu-dung-thang-chin-nam-2026/"
# Synthetic percentages intentionally differ from published statistics.
SUMMARY="Chỉ số giá tiêu dùng (CPI) tháng Chín tăng 0,25% so với tháng trước; tăng 2,90% so với tháng 12/2025 và tăng 3,40% so với cùng kỳ năm trước. CPI bình quân quý III/2026 tăng 3,60% so với quý III/2025. Bình quân chín tháng năm 2026, CPI tăng 3,10% so với cùng kỳ năm trước; lạm phát cơ bản tăng 2,80%."

def article(summary=SUMMARY,title="Chỉ số giá tiêu dùng, chỉ số giá vàng và chỉ số giá đô la Mỹ tháng Chín, quý III và 9 tháng năm 2026"):
 return '<html><h1>'+title+'</h1><p>'+summary+'</p><p>Chỉ số giá vàng tháng Chín tăng 12,30% so với cùng kỳ năm trước.</p><p>Ngày đăng: 03/10/2026</p></html>'

class CPIReportTests(unittest.TestCase):
 def test_bases_are_separate_and_average_is_not_monthly_yoy(self):
  rows=extract_article(article(),URL,date(2026,10,26));values={r['basis']:r['value'] for r in rows}
  self.assertEqual(values,{'MoM':0.25,'YoY':3.4,'AVG_YOY':3.1})
  self.assertEqual({r['period'] for r in rows},{'2026-09'})
  self.assertEqual(rows[-1]['evidence']['windowMonths'],9)
  self.assertIn('3,10',rows[-1]['evidence']['text'])
  self.assertEqual(rows[0]['evidence']['publicationPrecision'],'date')
 def test_missing_average_stays_missing_instead_of_using_quarter_average(self):
  text=SUMMARY.split("Bình quân chín tháng")[0]
  rows=extract_article(article(text),URL,date(2026,10,26))
  self.assertEqual({r["basis"] for r in rows},{"MoM","YoY"})
 def test_price_release_title_is_not_confused_with_province_word(self):
  rows=extract_article(article(title="Thông cáo báo chí về tình hình giá tháng Chín, quý III và 9 tháng năm 2026"),URL,date(2026,10,26))
  self.assertEqual(len(rows),3)
 def test_negative_unchanged_inverse_and_conflicting_values(self):
  text=SUMMARY.replace('tăng 0,25%','giảm 0,25%')
  self.assertEqual(extract_article(article(text),URL,date(2026,10,26))[0]['value'],-0.25)
  text=SUMMARY.replace('tăng 0,25%','không thay đổi')
  self.assertEqual(extract_article(article(text),URL,date(2026,10,26))[0]['value'],0)
  text='So với tháng trước, CPI tháng 9/2026 tăng 0,25%.</p><p>'+SUMMARY
  self.assertEqual(extract_article(article(text),URL,date(2026,10,26))[0]['value'],0.25)
  with self.assertRaises(CollectionError):extract_article(article(SUMMARY+'</p><p>'+SUMMARY.replace('0,25%','0,65%')),URL,date(2026,10,26))
 def test_rejects_regional_future_and_missing_comparators(self):
  for html,url,today in [(article(title='Chỉ số giá tiêu dùng Hà Nội tháng Chín năm 2026'),URL,date(2026,10,26)),(article(),URL,date(2026,9,1)),(article(SUMMARY.replace('so với cùng kỳ năm trước','so với tháng trước')),URL,date(2026,10,26)),(article(),'https://evil.test/report',date(2026,10,26))]:
   with self.assertRaises(CollectionError):extract_article(html,url,today)
 def test_rejected_article_is_retained_for_audit(self):
  class Client:
   def get(self,url):
    if url==INDEX:return ('<a href="'+URL+'">Chỉ số giá tiêu dùng tháng Chín năm 2026</a>').encode()
    return article(SUMMARY.replace('so với cùng kỳ năm trước','so với quý trước')).encode()
  with self.assertRaises(CollectionError) as caught:collect(Client(),INDEX,date(2026,10,26))
  self.assertEqual(json.loads(caught.exception.payload)['articles'][0]['url'],URL)
 def test_scrapes_topic_and_falls_back_to_same_publisher_new_domain(self):
  calls=[]
  class Client:
   def get(self,url):
    calls.append(url)
    if url==INDEX:raise CollectionError('HTTP','old domain unavailable')
    if url==ALTERNATE:return ('<a href="'+URL+'">Chỉ số giá tiêu dùng tháng Chín năm 2026</a>').encode()
    return article().encode()
  body,rows=collect(Client(),INDEX,date(2026,10,26))
  self.assertEqual(len(rows),3);self.assertEqual(json.loads(body)['indexUrl'],ALTERNATE)
  self.assertEqual(calls,[INDEX,ALTERNATE,URL])

@unittest.skipUnless(os.getenv("MACRO_TEST_DATABASE_URL"),"requires CI PostgreSQL")
class CPIReportDatabaseTests(PostgresPilotFixture):
 def test_cpi_report_shadow_has_evidence_and_cannot_claim_parser_parity(self):
  from macro_collectors.pilot import run,report,selected_datasets
  self.db.execute("UPDATE macro_datasets SET enabled=true WHERE id='gso_cpi_reports'")
  class Client:
   def get(self,url):
    if url in (INDEX,ALTERNATE):return ('<a href="'+URL+'">Chỉ số giá tiêu dùng tháng Chín năm 2026</a>').encode()
    return article().encode()
  result=run(self.db,'gso_cpi_reports',True,Client)
  # These fixtures deliberately have no current production CPI for this period.
  self.assertIn(result[0]['status'],['CROSS_SOURCE_REVIEW','STALE'])
  candidates=self.db.execute("SELECT value,unit,evidence,source_url FROM macro_pilot_observations ORDER BY unit").fetchall()
  self.assertEqual(len(candidates),3);self.assertTrue(all(row['source_url']==URL for row in candidates))
  self.assertEqual({row['evidence']['comparisonBasis'] for row in candidates},{'MoM','YoY','AVG_YOY'})
  r=report(self.db,'gso_cpi_reports')[0]
  self.assertFalse(r['eligibleForReview']);self.assertEqual(r['owner'],'go')
  self.assertEqual(self.db.execute('SELECT count(*) AS n FROM macro_observations').fetchone()['n'],0)
  self.assertIn('gso_cpi_reports',selected_datasets(self.db,indicator='VN_CPI_YOY'))
  raw=self.db.execute("INSERT INTO macro_raw_documents(source,source_url,content_type,checksum,payload,first_fetched_at,last_fetched_at) VALUES('NSO',%s,'text/html','fixture-report',%s,now(),now()) RETURNING id",(URL,b'Go fixture')).fetchone()['id']
  self.db.execute("INSERT INTO macro_observations(indicator_code,period,period_end,value,unit,source,source_url,series_id,method,priority,first_seen_at,collected_at,revision,raw_id) VALUES('VN_CPI_YOY','2026-09','2026-09-30',3.4,'%% YoY','NSO',%s,'GO_REPORT','SOURCE_MONTHLY_YOY',110,now(),now(),1,%s)",(URL,raw))
  run(self.db,'gso_cpi_reports',True,Client)
  r=report(self.db,'gso_cpi_reports')[0]
  self.assertFalse(r['eligibleForReview'])
  self.assertIn('CROSS_SOURCE_VALUE_MATCH',[check['status'] for check in r['latestRun']['report']['checks']])
  self.assertEqual(self.db.execute("SELECT value FROM macro_observations WHERE indicator_code='VN_CPI_YOY'").fetchone()['value'],3.4)
