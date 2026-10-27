package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/providers"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/testutil"
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
	if dashboard.Mode != "live" || dashboard.Score != nil || dashboard.Enabled != 19 || len(dashboard.Indicators) != 19 || dashboard.Available != 1 {
		t.Fatalf("invalid partial dashboard %+v", dashboard)
	}
	for _, c := range dashboard.Indicators {
		if c.Code == "US_EFFR" && (c.Status != "NOT_CONFIGURED" || c.Latest != nil) {
			t.Fatal("missing source was fabricated")
		}
	}
}

func TestPagedProvenanceAndExpectedMethodologyGap(t *testing.T) {
	repo := &testutil.Memory{}
	end, _ := macro.MonthEnd("2026-08")
	o := macro.Observation{IndicatorCode: "VN_M2_LEVEL", Period: "2026-08", PeriodEnd: end, Value: 20774696, Unit: "tỷ đồng", Source: "SBV", SourceURL: providers.SBVM2URL, SeriesID: "SBV_M2", Method: "SOURCE_LEVEL", RawIndex: 1}
	historical := o
	historical.IndicatorCode = "VN_M2_GROWTH"
	historical.Value = 8
	historical.Unit = "% YoY"
	first := macro.RawDocument{Source: "SBV", URL: "https://sbv.gov.vn/page1", Body: []byte("first"), FetchedAt: time.Now()}
	second := macro.RawDocument{Source: "SBV", URL: "https://sbv.gov.vn/page2", Body: []byte("second"), FetchedAt: time.Now()}
	s := Service{Repo: repo, Providers: []providers.Provider{fakeProvider{code: "VN_M2_GROWTH", batch: macro.Batch{Document: first, Documents: []macro.RawDocument{second}, Observations: []macro.Observation{o, historical}, Unavailable: map[string]string{"VN_M2_GROWTH": "NOT_COMPARABLE: methodology break"}}}}}
	result, err := s.Ingest(context.Background())
	if err != nil || result.Failed != 0 || len(repo.Raw) != 2 || repo.Rows[0].RawID != 2 {
		t.Fatalf("raw page binding or policy gap failed %+v %v", result, err)
	}
	dashboard, err := s.Dashboard(context.Background(), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for _, card := range dashboard.Indicators {
		if card.Code == "VN_M2_GROWTH" && card.Status != "NOT_COMPARABLE" {
			t.Fatal("gap status lost")
		}
	}
}

func TestBillsNoPublishedRateIsNotFailedIngestion(t *testing.T) {
	repo := &testutil.Memory{}
	s := Service{Repo: repo, Providers: []providers.Provider{fakeProvider{code: "VN_SBV_TBILL", batch: macro.Batch{Document: macro.RawDocument{Source: "SBV", URL: providers.SBVOMOURL, Body: []byte("official repo-only results"), FetchedAt: time.Now()}, Unavailable: map[string]string{"VN_SBV_TBILL": "NO_PUBLISHED_RATE: no bill rate in today's results"}}}}}
	result, err := s.Ingest(context.Background())
	if err != nil || result.Failed != 0 || result.Saved != 0 || len(repo.Raw) != 1 {
		t.Fatalf("normal absence failed or fabricated %+v %v", result, err)
	}
	dashboard, err := s.Dashboard(context.Background(), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range dashboard.Indicators {
		if c.Code == "VN_SBV_TBILL" && (c.Status != "NO_PUBLISHED_RATE" || c.Latest != nil) {
			t.Fatal("zero yield fabricated")
		}
	}
}

func TestExpectedPMIPublicationDelayIsNotFailedCollection(t *testing.T) {
	repo := &testutil.Memory{}
	s := Service{Repo: repo, Providers: []providers.Provider{fakeProvider{code: "VN_PMI", batch: macro.Batch{Document: macro.RawDocument{Source: "S&P Global", URL: "https://www.pmi.spglobal.com/", Body: []byte("pending manifest"), FetchedAt: time.Now()}, Unavailable: map[string]string{"VN_PMI": "AWAITING_RELEASE: expected month has not been published"}}}}}
	run, err := s.Ingest(context.Background())
	if err != nil || run.Failed != 0 || run.Saved != 0 {
		t.Fatalf("publication delay misclassified %+v %v", run, err)
	}
	d, err := s.Dashboard(context.Background(), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for _, card := range d.Indicators {
		if card.Code == "VN_PMI" && (card.Status != "AWAITING_RELEASE" || card.Latest != nil) {
			t.Fatal("pending PMI fabricated a value or outage")
		}
	}
}
