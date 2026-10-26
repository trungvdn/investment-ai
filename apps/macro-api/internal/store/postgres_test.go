package store

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
	"os"
	"sync"
	"testing"
	"time"
)

func TestPostgresPersistenceRevisionsAndSourcePriority(t *testing.T) {
	dsn := os.Getenv("MACRO_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires MACRO_TEST_DATABASE_URL; CI runs with PostgreSQL service")
	}
	ctx := context.Background()
	admin, err := Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	name := fmt.Sprintf("macro_test_%d", time.Now().UnixNano())
	schemaName := pgx.Identifier{name}.Sanitize()
	if _, err = admin.Pool.Exec(ctx, "CREATE SCHEMA "+schemaName); err != nil {
		t.Fatal(err)
	}
	defer admin.Pool.Exec(ctx, "DROP SCHEMA "+schemaName+" CASCADE")
	reopen := func() *Postgres {
		config, err := pgxpool.ParseConfig(dsn)
		if err != nil {
			t.Fatal(err)
		}
		config.ConnConfig.RuntimeParams["search_path"] = name
		pool, err := pgxpool.NewWithConfig(ctx, config)
		if err != nil {
			t.Fatal(err)
		}
		if err = pool.Ping(ctx); err != nil {
			t.Fatal(err)
		}
		return &Postgres{Pool: pool}
	}
	p := reopen()
	defer func() {
		if p != nil {
			p.Close()
		}
	}()
	if err = p.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	if err = p.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	doc := macro.RawDocument{Source: "NSO", URL: "https://nsdp.nso.gov.vn/test.xml", ContentType: "text/xml", Body: []byte("fixture"), FetchedAt: time.Now().UTC()}
	id, err := p.SaveDocument(ctx, doc)
	if err != nil {
		t.Fatal(err)
	}
	same, err := p.SaveDocument(ctx, doc)
	if err != nil || same != id {
		t.Fatal("raw snapshot idempotency failed")
	}
	end, _ := macro.MonthEnd("2026-01")
	o := macro.Observation{IndicatorCode: "VN_CPI_YOY", Period: "2026-01", PeriodEnd: end, Value: 3, Unit: "% YoY", Source: "NSO", SourceURL: doc.URL, SeriesID: "PCPI_IX", Method: "YOY", Priority: 50, RawID: id, CollectedAt: time.Now().UTC()}
	if err = p.Save(ctx, o); err != nil {
		t.Fatal(err)
	}
	if err = p.Save(ctx, o); err != nil {
		t.Fatal(err)
	}
	first, _ := p.History(ctx, o.IndicatorCode, nil, nil, true)
	if len(first) != 1 {
		t.Fatal("same value duplicated")
	}
	seen := first[0].FirstSeenAt
	o.Value = 3.1
	o.CollectedAt = o.CollectedAt.Add(time.Second)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := p.Save(ctx, o); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	history, err := p.History(ctx, o.IndicatorCode, nil, nil, true)
	if err != nil || len(history) != 2 || history[0].Revision != 2 {
		t.Fatalf("revision/concurrency failed: %+v %v", history, err)
	}
	if !history[1].FirstSeenAt.Equal(seen) {
		t.Fatal("original vintage overwritten")
	}
	o.SeriesID = "NSO_REPORT"
	o.Priority = 110
	o.Value = 4
	if err = p.Save(ctx, o); err != nil {
		t.Fatal(err)
	}
	canonical, err := p.History(ctx, o.IndicatorCode, nil, nil, false)
	if err != nil || len(canonical) != 1 || canonical[0].Value != 4 {
		t.Fatal("canonical source priority failed")
	}

	// Active canonical batches pin explicit observation IDs, even if Go later
	// ingests a higher-priority source. Deactivation restores priority selection.
	_, err = p.Pool.Exec(ctx, `WITH raw AS (
	 INSERT INTO macro_pilot_raw(dataset_id,checksum,payload,fetched_at) VALUES('nso_cpi','selection-test','fixture',now()) RETURNING id
	), run AS (
	 INSERT INTO macro_pilot_runs(dataset_id,raw_id,parser_version,status,report)
	 SELECT 'nso_cpi',id,'test','MATCH','{}'::jsonb FROM raw RETURNING id,raw_id
	), vintage AS (
	 INSERT INTO macro_source_values(binding_id,period,period_end,revision,fingerprint,value,unit,method,source_url,evidence,raw_id,first_run_id)
	 SELECT 'nso_cpi_yoy','2026-01','2026-01-31',1,'test',3,'% YoY','YOY','https://example.test','{}'::jsonb,raw_id,id FROM run RETURNING id
	), batch AS (
	 INSERT INTO macro_canonical_batches(indicator_code,fingerprint,report) VALUES('VN_CPI_YOY','test','{}'::jsonb) RETURNING id
	), link AS (
	 INSERT INTO macro_canonical_links(batch_id,period,source_value_id,observation_id)
	 SELECT batch.id,'2026-01',vintage.id,(SELECT id FROM macro_observations WHERE series_id='PCPI_IX' AND revision=1)
	 FROM batch,vintage RETURNING batch_id
	) UPDATE macro_source_policies SET active_batch_id=(SELECT batch_id FROM link) WHERE indicator_code='VN_CPI_YOY'`)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err = p.History(ctx, o.IndicatorCode, nil, nil, false)
	if err != nil || len(canonical) != 1 || canonical[0].Value != 3 {
		t.Fatalf("canonical batch selection failed: %+v %v", canonical, err)
	}
	allRevisions, err := p.History(ctx, o.IndicatorCode, nil, nil, true)
	if err != nil || len(allRevisions) != 3 {
		t.Fatal("canonical selection discarded source revisions")
	}
	if err = p.RecordAttempt(ctx, "python:canonical:VN_CPI_YOY", macro.Attempt{Code: o.IndicatorCode, CheckedAt: time.Now().Add(-time.Hour), Success: true, Message: "selected"}); err != nil {
		t.Fatal(err)
	}
	if err = p.RecordAttempt(ctx, "test", macro.Attempt{Code: o.IndicatorCode, CheckedAt: time.Now().Add(-time.Minute), Success: false, Message: "unselected source failed"}); err != nil {
		t.Fatal(err)
	}
	selectedHealth, err := p.Attempts(ctx)
	if err != nil || !selectedHealth[o.IndicatorCode].Success {
		t.Fatal("canonical health was overridden by an unselected Go source")
	}
	if _, err = p.Pool.Exec(ctx, "UPDATE macro_source_policies SET active_batch_id=NULL WHERE indicator_code='VN_CPI_YOY'"); err != nil {
		t.Fatal(err)
	}
	canonical, err = p.History(ctx, o.IndicatorCode, nil, nil, false)
	if err != nil || len(canonical) != 1 || canonical[0].Value != 4 {
		t.Fatal("return to source priority failed")
	}
	if err = p.RecordAttempt(ctx, "test", macro.Attempt{Code: o.IndicatorCode, CheckedAt: time.Now(), Success: true, Message: "ok"}); err != nil {
		t.Fatal(err)
	}
	ppiEnd, _ := macro.QuarterEnd("2026-Q3")
	ppi := macro.Observation{IndicatorCode: "VN_PPI_YOY", Period: "2026-Q3", PeriodEnd: ppiEnd, Value: 4.75, Unit: "% YoY", Source: "NSO", SourceURL: doc.URL, SeriesID: "PPI_QUARTERLY", Method: "SAME_QUARTER_INDEX_MINUS_100", Priority: 100, RawID: id, CollectedAt: time.Now().UTC()}
	if err = p.Save(ctx, ppi); err != nil {
		t.Fatal(err)
	}
	p.Close()
	p = reopen()
	again, err := p.History(ctx, o.IndicatorCode, nil, nil, false)
	if err != nil || len(again) != 1 || again[0].Value != 4 {
		t.Fatal("data did not survive repository restart")
	}
	ppiHistory, err := p.History(ctx, "VN_PPI_YOY", nil, nil, false)
	if err != nil || len(ppiHistory) != 1 || ppiHistory[0].Period != "2026-Q3" || ppiHistory[0].PeriodEnd.Format("2006-01-02") != "2026-09-30" {
		t.Fatal("quarter period did not survive persistence")
	}
	attempts, err := p.Attempts(ctx)
	if err != nil || !attempts[o.IndicatorCode].Success {
		t.Fatal("attempt status not durable")
	}
	// Atomic leases, stale-owner fencing and persistent recovery schedule.
	now := time.Now().UTC()
	tokens := make(chan string, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			token, e := p.ClaimSource(ctx, "sbv:fx", now)
			if e != nil {
				t.Error(e)
			}
			tokens <- token
		}()
	}
	wg.Wait()
	close(tokens)
	winner := ""
	owners := 0
	for token := range tokens {
		if token != "" {
			owners++
			winner = token
		}
	}
	if owners != 1 {
		t.Fatalf("multiple source owners: %d", owners)
	}
	next := now.Add(6 * time.Hour)
	if err = p.FinishSource(ctx, "sbv:fx", "stale-owner", now, next, "ACCESS_DENIED", "403"); err == nil {
		t.Fatal("stale owner changed schedule")
	}
	if err = p.FinishSource(ctx, "sbv:fx", winner, now, next, "ACCESS_DENIED", "403"); err != nil {
		t.Fatal(err)
	}
	p.Close()
	p = reopen()
	states, e := p.SourceStates(ctx)
	if e != nil || len(states) != 1 || states[0].Failures != 1 || states[0].Kind != "ACCESS_DENIED" || states[0].NextDue.Sub(next).Abs() > time.Second {
		t.Fatalf("source schedule not durable %+v %v", states, e)
	}
	if token, e := p.ClaimSource(ctx, "sbv:fx", now); e != nil || token != "" {
		t.Fatal("future schedule ignored")
	}
	token, e := p.ClaimSource(ctx, "sbv:fx", next)
	if e != nil || token == "" {
		t.Fatal("due source not claimed")
	}
	if err = p.FinishSource(ctx, "sbv:fx", token, next, next.Add(time.Hour), "", "recovered"); err != nil {
		t.Fatal(err)
	}
	states, e = p.SourceStates(ctx)
	if e != nil || states[0].Failures != 0 || states[0].LastSuccess == nil {
		t.Fatal("source recovery not persisted")
	}
	// A crashed owner expires; the old token cannot finish after reassignment.
	token, e = p.ClaimSource(ctx, "sbv:bills", now)
	if e != nil || token == "" {
		t.Fatal("lease missing")
	}
	replacement, e := p.ClaimSource(ctx, "sbv:bills", now.Add(11*time.Minute))
	if e != nil || replacement == "" || replacement == token {
		t.Fatal("crashed source lease not reclaimed")
	}
	if e = p.FinishSource(ctx, "sbv:bills", token, now.Add(11*time.Minute), next, "", "late"); e == nil {
		t.Fatal("expired owner not fenced")
	}

	if e = p.MakeSourceDue(ctx, "sbv:bills"); e == nil {
		t.Fatal("operator override stole an active lease")
	}
	if e = p.FinishSource(ctx, "sbv:bills", replacement, now.Add(11*time.Minute), next, "", "done"); e != nil {
		t.Fatal(e)
	}
	if e = p.MakeSourceDue(ctx, "sbv:bills"); e != nil {
		t.Fatal(e)
	}
	token, e = p.ClaimSource(ctx, "sbv:bills", time.Now().Add(time.Second))
	if e != nil || token == "" {
		t.Fatal("operator requeue did not become due")
	}

}
