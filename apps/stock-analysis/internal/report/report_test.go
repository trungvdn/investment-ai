package report

import (
	"strings"
	"testing"
	"time"

	"github.com/trungvdn/investment-ai/apps/stock-analysis/internal/wyckoff"
)

func TestBuildProducesStableReportContract(t *testing.T) {
	asOf := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	input := wyckoff.Analysis{
		Symbol:     "szc",
		AsOf:       asOf,
		Structure:  wyckoff.StructureAccumulation,
		Phase:      wyckoff.PhaseD,
		Confidence: 0.72,
		TradingRange: &wyckoff.TradingRange{
			Support: 35.5, Resistance: 40.2, WidthPct: 12.4,
		},
		Events: []wyckoff.Event{{
			Type: "SOS", Date: asOf, Confidence: 0.7, Evidence: "Breakout above resistance.",
		}},
		Evidence:   []string{"SOS (2026-10-20): Breakout above resistance."},
		DataPoints: 120,
		DataMode:   "kbsec",
	}

	got := Build(input)
	if got.SchemaVersion != SchemaVersion {
		t.Fatalf("schemaVersion = %q, want %q", got.SchemaVersion, SchemaVersion)
	}
	if got.Symbol != "SZC" || got.Structure != wyckoff.StructureAccumulation || got.Phase != wyckoff.PhaseD {
		t.Fatalf("unexpected report identity/classification: %+v", got)
	}
	if got.TradingRange == nil || got.TradingRange.Status != "IDENTIFIED" {
		t.Fatalf("expected identified trading range, got %+v", got.TradingRange)
	}
	if len(got.Events) != 1 || got.Events[0].Type != "SOS" {
		t.Fatalf("unexpected event sections: %+v", got.Events)
	}
	if len(got.Evidence) != 1 || len(got.Risks) == 0 || len(got.Watchlist) == 0 {
		t.Fatalf("report should include evidence, risks and watchlist: %+v", got)
	}
	if !strings.Contains(got.Disclaimer, "not investment advice") {
		t.Fatalf("missing research disclaimer: %q", got.Disclaimer)
	}
}

func TestBuildFlagsUnknownStructureAndDemoData(t *testing.T) {
	got := Build(wyckoff.Analysis{
		Symbol:     "TEST",
		Structure:  wyckoff.StructureUnclear,
		Phase:      wyckoff.PhaseUnknown,
		Confidence: 0.2,
		DataPoints: 90,
		DataMode:   "demo",
	})

	if got.Headline != "Market structure is not confirmed" {
		t.Fatalf("unexpected headline %q", got.Headline)
	}
	if got.DataQuality.Status != "DEMO" {
		t.Fatalf("data quality status = %q, want DEMO", got.DataQuality.Status)
	}
	if !containsText(got.Risks, "demo data") || !containsText(got.Risks, "UNKNOWN") {
		t.Fatalf("expected demo and unknown-structure risks, got %v", got.Risks)
	}
	if got.TradingRange != nil {
		t.Fatalf("did not expect range when source has no range: %+v", got.TradingRange)
	}
}

func TestBuildClampsConfidenceAndHandlesMissingData(t *testing.T) {
	got := Build(wyckoff.Analysis{
		Symbol: "TEST", Confidence: 1.5, DataPoints: 30, DataMode: "kbsec",
	})
	if got.Confidence != 1 {
		t.Fatalf("confidence = %v, want 1", got.Confidence)
	}
	if got.DataQuality.Status != "INSUFFICIENT" {
		t.Fatalf("data quality status = %q, want INSUFFICIENT", got.DataQuality.Status)
	}
	if len(got.Risks) == 0 || len(got.Watchlist) == 0 {
		t.Fatalf("expected risks and watchlist even for incomplete analysis: %+v", got)
	}
}

func containsText(values []string, fragment string) bool {
	for _, value := range values {
		if strings.Contains(value, fragment) {
			return true
		}
	}
	return false
}
