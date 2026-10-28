package scoring

import (
	"math"
	"testing"
	"time"

	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
)

var testNow = time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)

func observation(code, period string, value float64) macro.Observation {
	ind, _ := macro.Find(code)
	end, _ := macro.MonthEnd(period)
	if ind.Frequency == "DAILY" {
		end, _ = time.Parse("2006-01-02", period)
	}
	if ind.Frequency == "QUARTERLY" {
		end, _ = macro.QuarterEnd(period)
	}
	return macro.Observation{IndicatorCode: code, Period: period, PeriodEnd: end, Value: value, Unit: ind.Unit, Source: "test", SourceURL: "https://sbv.gov.vn/test", SeriesID: code, Method: "TEST", FirstSeenAt: testNow.Add(-time.Hour), CollectedAt: testNow, Revision: 1, RawID: 1}
}
func find(s Snapshot, code string) Component {
	for _, c := range s.Components {
		if c.Code == code {
			return c
		}
	}
	panic(code)
}
func monthly(code string, values []float64) []macro.Observation {
	out := []macro.Observation{}
	for i, v := range values {
		period := time.Date(2026, 4+time.Month(i), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")
		o := observation(code, period, v)
		if code == "VN_DEPOSIT_RATE" {
			upper := v + 1
			o.UpperValue = &upper
		}
		out = append(out, o)
	}
	return out
}
func creditFixture() []macro.Observation {
	out := []macro.Observation{}
	for i := 0; i < 7; i++ {
		o := observation("VN_CREDIT_GROWTH", time.Date(2026, 3+time.Month(i), 1, 0, 0, 0, 0, time.UTC).Format("2006-01"), float64(i+3))
		out = append(out, o)
	}
	return out
}
func full() map[string][]macro.Observation {
	return map[string][]macro.Observation{
		"VN_DEPOSIT_RATE":  monthly("VN_DEPOSIT_RATE", []float64{5, 5, 5, 4.5, 4.5, 4.5}),
		"VN_PMI":           monthly("VN_PMI", []float64{48, 49, 49, 50, 51, 52}),
		"VN_CPI_YOY":       monthly("VN_CPI_YOY", []float64{5, 5, 5, 4, 4, 4}),
		"VN_CREDIT_GROWTH": creditFixture(),
		"VN_IIP_YOY":       monthly("VN_IIP_YOY", []float64{3, 3, 3, 5, 5, 5}),
		"VN_PPI_YOY":       {observation("VN_PPI_YOY", "2026-Q2", 3), observation("VN_PPI_YOY", "2026-Q3", 1)},
		"VN_M2_GROWTH":     monthly("VN_M2_GROWTH", []float64{10, 10, 10, 12, 12, 12}),
		"US_EFFR":          {observation("US_EFFR", "2026-07-27", 4.5), observation("US_EFFR", "2026-10-27", 4)},
	}
}
func TestCoverageGatesAndContributions(t *testing.T) {
	m := DefaultModel()
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	data := full()
	s := Calculate(m, data, testNow)
	if s.Status != "READY" || s.Coverage != 100 || s.Score == nil {
		t.Fatalf("%+v", s)
	}
	sum := 0.0
	for _, c := range s.Components {
		sum += c.Contribution
	}
	if math.Abs(sum-*s.Score) > 1e-9 {
		t.Fatal("contributions do not sum")
	}
	delete(data, "VN_PMI")
	s = Calculate(m, data, testNow)
	if s.Status != "PROVISIONAL" || s.Coverage != 80 {
		t.Fatal("critical gate ignored")
	}
	s = Calculate(m, nil, testNow)
	if s.Score != nil || s.Coverage != 0 || s.Status != "INSUFFICIENT" {
		t.Fatal("missing data received score")
	}
}
func TestNoFakeTrendAndFreshness(t *testing.T) {
	m := DefaultModel()
	d := map[string][]macro.Observation{"VN_PMI": {observation("VN_PMI", "2026-09", 52)}}
	if find(Calculate(m, d, testNow), "VN_PMI").Score != nil {
		t.Fatal("fabricated trend")
	}
	d = full()
	s := Calculate(m, d, testNow.AddDate(0, 6, 0))
	if s.Score != nil {
		t.Fatal("stale data scored")
	}
	d = full()
	d["VN_PMI"][5].FirstSeenAt = testNow.Add(time.Hour)
	if find(Calculate(m, d, testNow), "VN_PMI").Inputs[0].Period == "2026-09" {
		t.Fatal("future vintage used")
	}
}
func TestFingerprintAndRangeBounds(t *testing.T) {
	d := full()
	m := DefaultModel()
	s := Calculate(m, d, testNow)
	d["VN_PMI"][5].CollectedAt = testNow.Add(time.Minute)
	if Calculate(m, d, testNow).InputHash != s.InputHash {
		t.Fatal("refresh created new vintage")
	}
	original := *find(s, "VN_DEPOSIT_RATE").Score
	upper := *d["VN_DEPOSIT_RATE"][5].UpperValue + 1
	d["VN_DEPOSIT_RATE"][5].UpperValue = &upper
	d["VN_DEPOSIT_RATE"][5].Revision++
	next := Calculate(m, d, testNow)
	if next.InputHash == s.InputHash || *find(next, "VN_DEPOSIT_RATE").Score >= original {
		t.Fatal("upper bound ignored")
	}
	if *find(s, "VN_DEPOSIT_RATE").Inputs[0].UpperValue == upper {
		t.Fatal("input snapshot changed")
	}
}
func TestCreditJanuaryReset(t *testing.T) {
	code := "VN_CREDIT_GROWTH"
	data := []macro.Observation{}
	// Constant 1% monthly credit growth across a year boundary. January YTD is
	// 1%, even though December YTD is over 12%; this is not a contraction.
	for i := 0; i < 7; i++ {
		date := time.Date(2025, 9+time.Month(i), 1, 0, 0, 0, 0, time.UTC)
		v := (math.Pow(1.01, float64(date.Month())) - 1) * 100
		o := observation(code, date.Format("2006-01"), v)
		data = append(data, o)
	}
	rule := DefaultModel().Rules[3]
	_, score, err := component(rule, data)
	if err != nil || math.Abs(score-50) > 0.00001 {
		t.Fatalf("January reset: %v %v", score, err)
	}
}
func TestDeflationNotAlwaysBetterAndPMITrend(t *testing.T) {
	m := DefaultModel()
	d := full()
	s := Calculate(m, d, testNow)
	before := *find(s, "VN_CPI_YOY").Score
	d["VN_CPI_YOY"][5].Value = -3
	if *find(Calculate(m, d, testNow), "VN_CPI_YOY").Score >= before {
		t.Fatal("deflation rewarded")
	}
	before = *find(s, "VN_PMI").Score
	d["VN_PMI"][5].Value = 48
	if *find(Calculate(m, d, testNow), "VN_PMI").Score >= before {
		t.Fatal("PMI deterioration ignored")
	}
}
func TestIncompatibleHistoryExcluded(t *testing.T) {
	d := full()
	d["VN_M2_GROWTH"][2].Method = "OLD METHODOLOGY"
	c := find(Calculate(DefaultModel(), d, testNow), "VN_M2_GROWTH")
	if c.Score != nil || c.Reason != "INCOMPARABLE_HISTORY" {
		t.Fatal("methodology break crossed")
	}
}

func TestZeroIsObservedValueAndManualProvenanceIsVisible(t *testing.T) {
	d := full()
	d["VN_PMI"][5].Value = 0
	d["VN_CPI_YOY"][5].Method = "USER_SUPPLIED_GAP_FILL; NOT_INDEPENDENTLY_VERIFIED"
	s := Calculate(DefaultModel(), d, testNow)
	pmi := find(s, "VN_PMI")
	if pmi.Score == nil || *pmi.Score != 0 || s.Coverage != 100 {
		t.Fatal("observed zero confused with missing data")
	}
	cpi := find(s, "VN_CPI_YOY")
	if len(cpi.Warnings) != 1 || cpi.Warnings[0] != "USER_SUPPLIED_UNVERIFIED" {
		t.Fatal("manual provenance hidden")
	}
}
func TestMethodProvenanceTagsAndFreshnessFingerprint(t *testing.T) {
	d := full()
	d["VN_M2_GROWTH"][2].Method = "TEST; ARTICLE_ID=old; PARSER=old"
	d["VN_M2_GROWTH"][5].Method = "TEST; ARTICLE_ID=new; DENOMINATOR_ARTICLE_ID=base; PARSER=new"
	m := DefaultModel()
	s := Calculate(m, d, testNow)
	if find(s, "VN_M2_GROWTH").Score == nil {
		t.Fatal("provenance confused with methodology")
	}
	expired := Calculate(m, d, testNow.AddDate(0, 3, 0))
	if s.InputHash == expired.InputHash {
		t.Fatal("expiry did not change fingerprint")
	}
}
func TestEFFRCalendarMonthAnchor(t *testing.T) {
	all := []macro.Observation{observation("US_EFFR", "2026-02-27", 4.5), observation("US_EFFR", "2026-05-31", 4)}
	r := DefaultModel().Rules[5]
	_, score, err := component(r, all)
	if err != nil || score != 62.5 {
		t.Fatalf("clamped calendar anchor: %v %v", score, err)
	}
	all[0] = observation("US_EFFR", "2026-02-20", 4.5)
	if _, _, err = component(r, all); err == nil {
		t.Fatal("anchor gap over seven days accepted")
	}
}

func TestReconstructionUsesPublicationWithoutChangingVintage(t *testing.T) {
	at := time.Date(2026, 10, 26, 11, 0, 0, 0, time.UTC)
	published := at.AddDate(0, 0, -2)
	data := map[string][]macro.Observation{"VN_PMI": {}}
	for _, month := range []time.Month{6, 9} {
		end := time.Date(2026, month+1, 0, 0, 0, 0, 0, time.UTC)
		data["VN_PMI"] = append(data["VN_PMI"], macro.Observation{IndicatorCode: "VN_PMI", Period: end.Format("2006-01"), PeriodEnd: end, Value: 51, Unit: "index", Source: "test", SourceURL: "https://example.com", SeriesID: "test", Method: "test", Revision: 1, RawID: 1, FirstSeenAt: at.AddDate(0, 0, 1), PublishedAt: &published})
	}
	model := DefaultModel()
	if Calculate(model, data, at).Score != nil {
		t.Fatal("live calculation used future first-seen data")
	}
	rebuilt := Reconstruct(model, data, at)
	if rebuilt.Score == nil {
		t.Fatal("published data missing from reconstruction")
	}
	if !rebuilt.Components[1].Inputs[0].FirstSeenAt.After(at) {
		t.Fatal("original vintage changed")
	}
	future := at.Add(time.Hour)
	for i := range data["VN_PMI"] {
		data["VN_PMI"][i].PublishedAt = &future
	}
	if Reconstruct(model, data, at).Score != nil {
		t.Fatal("used future release")
	}
	for i := range data["VN_PMI"] {
		data["VN_PMI"][i].PublishedAt = nil
	}
	if Reconstruct(model, data, at).Score != nil {
		t.Fatal("unknown publication bypassed first-seen cutoff")
	}
}
