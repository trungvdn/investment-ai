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

func TestQuarterEndAndNotComparableStatus(t *testing.T) {
	end, err := QuarterEnd("2026-Q1")
	if err != nil || end.Format("2006-01-02") != "2026-03-31" {
		t.Fatal("quarter end incorrect")
	}
	for _, period := range []string{"2026-Q0", "2026-Q5", "2026-Q11", "bad"} {
		if _, err := QuarterEnd(period); err == nil {
			t.Fatal("invalid quarter accepted")
		}
	}
	ind, _ := Find("VN_M2_GROWTH")
	a := &Attempt{Success: false, Message: "NOT_COMPARABLE: methodology break"}
	if Status(ind, nil, a, time.Now()) != "NOT_COMPARABLE" {
		t.Fatal("methodology gap treated as normal missing data")
	}
	ppi, _ := Find("VN_PPI_YOY")
	end, _ = QuarterEnd("2026-Q3")
	obs := Observation{IndicatorCode: ppi.Code, Period: "2026-Q3", PeriodEnd: end, Value: 4.75, Unit: ppi.Unit, Source: "NSO", SourceURL: "https://nsdp.nso.gov.vn/ppi.xml", SeriesID: "PPI", Method: "INDEX_MINUS_100"}
	if err := Validate(obs, time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
}

func TestPMIConfigAndBillsNoPublishedRateRetainHistoricalMeaning(t *testing.T) {
	now := time.Now()
	pmi, _ := Find("VN_PMI")
	if Status(pmi, nil, &Attempt{Message: "PMI_RELEASE_URLS is not configured"}, now) != "NOT_CONFIGURED" {
		t.Fatal("PMI config gap hidden")
	}
	bill, _ := Find("VN_SBV_TBILL")
	old := &Observation{Value: 4, PeriodEnd: now.AddDate(0, 0, -30)}
	if Status(bill, old, &Attempt{Message: "NO_PUBLISHED_RATE: no published rate"}, now) != "NO_PUBLISHED_RATE" {
		t.Fatal("historical award masquerades as today's rate")
	}
}
