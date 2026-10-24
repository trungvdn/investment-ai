package macro

import (
	"math"
	"testing"
	"time"
)

func TestStatusRespectsPeriodAgeAndFailures(t *testing.T) {
	now := time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC)
	ind, _ := Find("VN_CPI_YOY")
	old := &Observation{PeriodEnd: now.AddDate(-1, 0, 0)}
	if Status(ind, old, nil, now) != "STALE" {
		t.Fatal("old period must not become fresh after collection")
	}
	key, _ := Find("US_EFFR")
	if Status(key, nil, &Attempt{Message: "FRED_API_KEY is not configured"}, now) != "NOT_CONFIGURED" {
		t.Fatal("missing configuration must be explicit")
	}
	recent := &Observation{PeriodEnd: now.AddDate(0, 0, -1)}
	if Status(key, recent, &Attempt{Success: false}, now) != "SOURCE_ERROR" {
		t.Fatal("keep latest good data on source failure")
	}
}
func TestValidationRejectsNonFiniteAndMismatchedPeriods(t *testing.T) {
	end, _ := MonthEnd("2026-01")
	o := Observation{IndicatorCode: "VN_CPI_YOY", Period: "2026-01", PeriodEnd: end, Value: 3, Unit: "% YoY", Source: "NSO", SourceURL: "https://nsdp.nso.gov.vn/cpi.xml", SeriesID: "PCPI_IX", Method: "YOY"}
	if err := Validate(o, time.Now()); err != nil {
		t.Fatal(err)
	}
	for _, v := range []float64{math.NaN(), math.Inf(1)} {
		o.Value = v
		if Validate(o, time.Now()) == nil {
			t.Fatal("nonfinite value accepted")
		}
	}
	o.Value = 3
	o.Period = "2026-02"
	if Validate(o, time.Now()) == nil {
		t.Fatal("period-end mismatch accepted")
	}
}
