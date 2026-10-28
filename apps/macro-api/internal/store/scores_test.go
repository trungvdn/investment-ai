package store

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
)

func scoreDatabase(t *testing.T) *Postgres {
	t.Helper()
	dsn := os.Getenv("MACRO_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires PostgreSQL; exercised in CI")
	}
	ctx := context.Background()
	admin, err := Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("score_test_%d", time.Now().UnixNano())
	quoted := pgx.Identifier{schema}.Sanitize()
	if _, err = admin.Pool.Exec(ctx, "CREATE SCHEMA "+quoted); err != nil {
		t.Fatal(err)
	}
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	p := &Postgres{Pool: pool}
	t.Cleanup(func() { p.Close(); admin.Pool.Exec(ctx, "DROP SCHEMA "+quoted+" CASCADE"); admin.Close() })
	if err = p.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	return p
}
func TestScoreSnapshotsPreserveVintagesAndSchedule(t *testing.T) {
	p := scoreDatabase(t)
	ctx := context.Background()
	now := time.Now().UTC()
	now = time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.UTC).AddDate(0, 0, 1)
	for now.Weekday() != time.Friday {
		now = now.AddDate(0, 0, 1)
	}
	raw, err := p.SaveDocument(ctx, macro.RawDocument{Source: "test", URL: "https://sbv.gov.vn/test", ContentType: "text/plain", Body: []byte("score fixture"), FetchedAt: time.Now().UTC()})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -1, 0)
	var latest macro.Observation
	for _, offset := range []int{-3, 0} {
		period := start.AddDate(0, offset, 0).Format("2006-01")
		end, _ := macro.MonthEnd(period)
		o := macro.Observation{IndicatorCode: "VN_PMI", Period: period, PeriodEnd: end, Value: 51, Unit: "index", Source: "test", SourceURL: "https://sbv.gov.vn/test", SeriesID: "SCORE_TEST", Method: "TEST", Priority: 100, RawID: raw, CollectedAt: time.Now().UTC()}
		if err = p.Save(ctx, o); err != nil {
			t.Fatal(err)
		}
		latest = o
	}
	refresh := func(at time.Time) {
		t.Helper()
		if e := p.RefreshMacroScore(ctx, at); e != nil {
			t.Fatal(e)
		}
	}
	refresh(now)
	refresh(now.Add(time.Minute))
	response, err := p.MacroScores(ctx, "daily")
	if err != nil {
		t.Fatal(err)
	}
	if response.Current == nil || response.Current.Score == nil || len(response.History) != 1 {
		t.Fatalf("snapshot missing: %+v", response)
	}
	original := response.History[0]
	if original.Coverage != 20 {
		t.Fatal("unexpected coverage")
	}
	var count int
	p.Pool.QueryRow(ctx, "SELECT count(*) FROM macro_score_snapshots").Scan(&count)
	if count != 3 {
		t.Fatalf("expected event,day,week once; got %d", count)
	}
	weekly, err := p.MacroScores(ctx, "weekly")
	if err != nil || len(weekly.History) != 1 {
		t.Fatal("weekly snapshot missing", err)
	}
	latest.Value = 54
	latest.CollectedAt = time.Now().UTC()
	if err = p.Save(ctx, latest); err != nil {
		t.Fatal(err)
	}
	refresh(now.Add(2 * time.Minute))
	response, err = p.MacroScores(ctx, "daily")
	if err != nil {
		t.Fatal(err)
	}
	if response.Current.InputHash == original.InputHash || *response.Current.Score == *original.Score {
		t.Fatal("new vintage not scored")
	}
	if response.History[0].InputHash != original.InputHash || *response.History[0].Score != *original.Score {
		t.Fatal("old snapshot rewritten")
	}
	// Competing replicas and identical input must not duplicate events or buckets.
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); errs <- p.RefreshMacroScore(ctx, now.Add(3*time.Minute)) }()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	p.Pool.QueryRow(ctx, "SELECT count(*) FROM macro_score_snapshots").Scan(&count)
	if count != 4 {
		t.Fatalf("duplicate snapshots: %d", count)
	}
	// A downtime gap creates only the actual new day, never missed historical days.
	refresh(now.AddDate(0, 0, 4))
	response, err = p.MacroScores(ctx, "daily")
	if err != nil || len(response.History) != 2 {
		t.Fatal("invented missing-day snapshots", err)
	}
}
