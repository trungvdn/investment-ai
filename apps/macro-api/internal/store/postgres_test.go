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
}
