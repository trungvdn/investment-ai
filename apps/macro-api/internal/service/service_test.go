package service

import (
	"context"
	"errors"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/providers"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/testutil"
	"testing"
	"time"
)

type fakeProvider struct {
	code  string
	batch macro.Batch
	err   error
}

func (f fakeProvider) ID() string                                 { return f.code }
func (f fakeProvider) Codes() []string                            { return []string{f.code} }
func (f fakeProvider) Fetch(context.Context) (macro.Batch, error) { return f.batch, f.err }
func TestPartialIngestionPreservesRawAndDoesNotFabricateMissingValues(t *testing.T) {
	repo := &testutil.Memory{}
	end, _ := macro.MonthEnd("2026-01")
	o := macro.Observation{IndicatorCode: "VN_CPI_YOY", Period: "2026-01", PeriodEnd: end, Value: 3, Unit: "% YoY", Source: "NSO", SourceURL: "https://nsdp.nso.gov.vn/cpi.xml", SeriesID: "PCPI_IX", Method: "YOY"}
	doc := macro.RawDocument{Source: "NSO", URL: o.SourceURL, Body: []byte("raw"), FetchedAt: time.Now()}
	s := Service{Repo: repo, Providers: []providers.Provider{fakeProvider{code: "US_EFFR", err: providers.ErrNotConfigured}, fakeProvider{code: o.IndicatorCode, batch: macro.Batch{Document: doc, Observations: []macro.Observation{o}}}, fakeProvider{code: "VN_IIP_YOY", batch: macro.Batch{Document: doc}, err: errors.New("ambiguous series")}}}
	result, err := s.Ingest(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Saved != 1 || result.Failed != 2 || len(repo.Raw) != 2 || len(repo.Rows) != 1 {
		t.Fatalf("partial ingestion incorrect %+v", result)
	}
	dashboard, err := s.Dashboard(context.Background(), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if dashboard.Mode != "live" || dashboard.Score != nil || dashboard.Enabled != 8 || len(dashboard.Indicators) != 15 || dashboard.Available != 1 {
		t.Fatalf("invalid partial dashboard %+v", dashboard)
	}
	for _, c := range dashboard.Indicators {
		if c.Code == "US_EFFR" && (c.Status != "NOT_CONFIGURED" || c.Latest != nil) {
			t.Fatal("missing source was fabricated")
		}
	}
}
